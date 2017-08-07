// Copyright (c) 2017 Zededa, Inc.
// All rights reserved.

// Process input changes from a config directory containing json encoded files
// with AppNetworkConfig and compare against AppNetworkStatus in the status
// dir.
// Produce the updated configlets (for radvd, dnsmasq, ip*tables, lisp.config,
// ipset, ip link/addr/route configuration) based on that and apply those
// configlets.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/zededa/go-provision/types"
	"github.com/zededa/go-provision/watch"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	// Keeping status in /var/run to be clean after a crash/reboot
	baseDirname := "/var/tmp/zedrouter"
	runDirname := "/var/run/zedrouter"
	configDirname := baseDirname + "/config"
	statusDirname := runDirname + "/status"

	if _, err := os.Stat(baseDirname); err != nil {
		if err := os.Mkdir(baseDirname, 0755); err != nil {
			log.Fatal(err)
		}
	}
	if _, err := os.Stat(configDirname); err != nil {
		if err := os.Mkdir(configDirname, 0755); err != nil {
			log.Fatal(err)
		}
	}
	if _, err := os.Stat(runDirname); err != nil {
		if err := os.Mkdir(runDirname, 0755); err != nil {
			log.Fatal(err)
		}
	}
	if _, err := os.Stat(statusDirname); err != nil {
		if err := os.Mkdir(statusDirname, 0755); err != nil {
			log.Fatal(err)
		}
	}
	appNumAllocatorInit(statusDirname, configDirname)

	handleInit(configDirname+"/global", statusDirname+"/global", runDirname)

	fileChanges := make(chan string)
	go watch.WatchConfigStatus(configDirname, statusDirname, fileChanges)
	// XXX can we feed in a "L" change when LISP needs to be restarted
	// to avoid multiple restarts when we do the initial ReadDir of
	// of the application configs? Better to remove the raw lisp iptable
	for {
		change := <-fileChanges
		parts := strings.Split(change, " ")
		operation := parts[0]
		fileName := parts[1]
		if !strings.HasSuffix(fileName, ".json") {
			log.Printf("Ignoring file <%s>\n", fileName)
			continue
		}
		if operation == "D" {
			statusFile := statusDirname + "/" + fileName
			if _, err := os.Stat(statusFile); err != nil {
				// File just vanished!
				log.Printf("File disappeared <%s>\n", fileName)
				continue
			}
			sb, err := ioutil.ReadFile(statusFile)
			if err != nil {
				log.Printf("%s for %s\n", err, statusFile)
				continue
			}
			status := types.AppNetworkStatus{}
			if err := json.Unmarshal(sb, &status); err != nil {
				log.Printf("%s AppNetworkStatus file: %s\n",
					err, statusFile)
				continue
			}
			uuid := status.UUIDandVersion.UUID
			if uuid.String()+".json" != fileName {
				log.Printf("Mismatch between filename and contained uuid: %s vs. %s\n",
					fileName, uuid.String())
				continue
			}
			statusName := statusDirname + "/" + fileName
			handleDelete(statusName, status)
			continue
		}
		if operation != "M" {
			log.Fatal("Unknown operation from Watcher: ", operation)
		}
		configFile := configDirname + "/" + fileName
		cb, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Printf("%s for %s\n", err, configFile)
			continue
		}
		config := types.AppNetworkConfig{}
		if err := json.Unmarshal(cb, &config); err != nil {
			log.Printf("%s AppNetworkConfig file: %s\n",
				err, configFile)
			continue
		}
		uuid := config.UUIDandVersion.UUID
		if uuid.String()+".json" != fileName {
			log.Printf("Mismatch between filename and contained uuid: %s vs. %s\n",
				fileName, uuid.String())
			continue
		}
		statusFile := statusDirname + "/" + fileName
		if _, err := os.Stat(statusFile); err != nil {
			// File does not exist in status hence new
			statusName := statusDirname + "/" + fileName
			handleCreate(statusName, config)
			continue
		}
		// Read and check status
		sb, err := ioutil.ReadFile(statusFile)
		if err != nil {
			log.Printf("%s for %s\n", err, statusFile)
			continue
		}
		status := types.AppNetworkStatus{}
		if err := json.Unmarshal(sb, &status); err != nil {
			log.Printf("%s AppNetworkStatus file: %s\n",
				err, statusFile)
			continue
		}
		uuid = status.UUIDandVersion.UUID
		if uuid.String()+".json" != fileName {
			log.Printf("Mismatch between filename and contained uuid: %s vs. %s\n",
				fileName, uuid.String())
			continue
		}
		// Look for pending* in status and repeat that operation.
		// XXX After that do a full ReadDir to restart ...
		if status.PendingAdd {
			statusName := statusDirname + "/" + fileName
			handleCreate(statusName, config)
			// XXX set something to rescan?
			continue
		}
		if status.PendingDelete {
			statusName := statusDirname + "/" + fileName
			handleDelete(statusName, status)
			// XXX set something to rescan?
			continue
		}
		if status.PendingModify {
			statusName := statusDirname + "/" + fileName
			handleModify(statusName, config, status)
			// XXX set something to rescan?
			continue
		}
		statusName := statusDirname + "/" + fileName
		handleModify(statusName, config, status)
	}
}

var globalConfig types.DeviceNetworkConfig
var globalStatus types.DeviceNetworkStatus
var globalStatusFilename string
var globalRunDirname string
var lispRunDirname string

func handleInit(configFilename string, statusFilename string,
     runDirname string) {
	globalStatusFilename = statusFilename
	globalRunDirname = runDirname

	// XXX should this be in the lisp code?
	lispRunDirname = runDirname + "/lisp"
	if _, err := os.Stat(lispRunDirname); err != nil {
		if err := os.Mkdir(lispRunDirname, 0755); err != nil {
			log.Fatal(err)
		}
	}

	cb, err := ioutil.ReadFile(configFilename)
	if err != nil {
		log.Printf("%s for %s\n", err, configFilename)
		log.Fatal(err)
	}
	if err := json.Unmarshal(cb, &globalConfig); err != nil {
		log.Printf("%s DeviceNetworkConfig file: %s\n",
			err, configFilename)
		log.Fatal(err)
	}
	_, err = netlink.LinkByName(globalConfig.Uplink)
	if err != nil {
		log.Fatal("Uplink in config/global does not exist: ",
			globalConfig.Uplink)
	}
	if _, err := os.Stat(statusFilename); err != nil {
		// Create and write with initial values
		globalStatus.Uplink = globalConfig.Uplink
		writeGlobalStatus()
	}
	sb, err := ioutil.ReadFile(statusFilename)
	if err != nil {
		log.Printf("%s for %s\n", err, statusFilename)
		log.Fatal(err)
	}
	if err := json.Unmarshal(sb, &globalStatus); err != nil {
		log.Printf("%s DeviceNetworkStatus file: %s\n",
			err, statusFilename)
		log.Fatal(err)
	}

	// Setup initial iptables rules
	iptablesInit()

	// ipsets which are independent of config
	createDefaultIpset()

	_, err = exec.Command("sysctl", "-w",
		"net.ipv4.ip_forward=1").Output()
	if err != nil {
		log.Fatal("Failed setting ip_forward ", err)
	}
	_, err = exec.Command("sysctl", "-w",
		"net.ipv6.conf.all.forwarding=1").Output()
	if err != nil {
		log.Fatal("Failed setting ipv6.conf.all.forwarding ", err)
	}
	// We use ip6tables for the bridge
	_, err = exec.Command("sysctl", "-w",
		"net.bridge.bridge-nf-call-ip6tables=1").Output()
	if err != nil {
		// XXX not in bobo's kernel
		log.Println("Failed setting net.bridge-nf-call-ip6tables ", err)
	}
	_, err = exec.Command("sysctl", "-w",
		"net.bridge.bridge-nf-call-iptables=1").Output()
	if err != nil {
		// XXX not in bobo's kernel
		log.Println("Failed setting net.bridge-nf-call-iptables ", err)
	}
	_, err = exec.Command("sysctl", "-w",
		"net.bridge.bridge-nf-call-arptables=1").Output()
	if err != nil {
		// XXX not in bobo's kernel
		log.Println("Failed setting net.bridge-nf-call-arptables ", err)
	}
}

func writeGlobalStatus() {
	b, err := json.Marshal(globalStatus)
	if err != nil {
		log.Fatal(err, "json Marshal DeviceNetworkStatus")
	}
	// We assume a /var/run path hence we don't need to worry about
	// partial writes/empty files due to a kernel crash.
	err = ioutil.WriteFile(globalStatusFilename, b, 0644)
	if err != nil {
		log.Fatal(err, globalStatusFilename)
	}
}

func writeAppNetworkStatus(status *types.AppNetworkStatus,
	statusFilename string) {
	b, err := json.Marshal(status)
	if err != nil {
		log.Fatal(err, "json Marshal AppNetworkStatus")
	}
	// We assume a /var/run path hence we don't need to worry about
	// partial writes/empty files due to a kernel crash.
	err = ioutil.WriteFile(statusFilename, b, 0644)
	if err != nil {
		log.Fatal(err, statusFilename)
	}
}

func handleCreate(statusFilename string, config types.AppNetworkConfig) {
	log.Printf("handleCreate(%v) for %s\n",
		config.UUIDandVersion, config.DisplayName)

	// Pick a local number to identify the application instance
	// Used for IP addresses as well bridge and file names.
	appNum := appNumAllocate(config.UUIDandVersion.UUID,
		config.IsZedmanager)

	// Start by marking with PendingAdd
	status := types.AppNetworkStatus{
		UUIDandVersion: config.UUIDandVersion,
		AppNum:         appNum,
		PendingAdd:     true,
		OlNum:          len(config.OverlayNetworkList),
		UlNum:          len(config.UnderlayNetworkList),
		DisplayName:    config.DisplayName,
		IsZedmanager:   config.IsZedmanager,
	}
	writeAppNetworkStatus(&status, statusFilename)

	if config.IsZedmanager {
		fmt.Printf("handleCreate: for %s IsZedmanager\n",
			config.DisplayName)
		if len(config.OverlayNetworkList) != 1 ||
		   len(config.UnderlayNetworkList) != 0 {
			log.Println("Malformed IsZedmanager config; ignored")
			return
		}
		
		// Use this olIfname to name files
		// XXX some files might not be used until Zedmanager becomes
		// a domU at which point IsZedMansger boolean won't be needed
		olConfig := config.OverlayNetworkList[0]
		olNum := 1
		olIfname := "dbo" + strconv.Itoa(olNum) + "x" +
			strconv.Itoa(appNum)

		// Create olIfname dummy interface with EID and fd00::/8 route
		// pointing at it.
		// XXX also a separate route for eidAllocationPrefix if global
		
		// Start clean
		attrs := netlink.NewLinkAttrs()
		attrs.Name = olIfname
		oLink := &netlink.Dummy{LinkAttrs: attrs}
		netlink.LinkDel(oLink)

		//    ip link add ${olIfname} type dummy
		attrs = netlink.NewLinkAttrs()
		attrs.Name = olIfname
		olIfMac := fmt.Sprintf("00:16:3e:02:%02x:%02x", olNum, appNum)
		hw, err := net.ParseMAC(olIfMac)
		if err != nil {
			log.Fatal("ParseMAC failed: ", olIfMac, err)
		}
		attrs.HardwareAddr = hw
		oLink = &netlink.Dummy{LinkAttrs: attrs}
		if err := netlink.LinkAdd(oLink); err != nil {
			fmt.Printf("LinkAdd on %s failed: %s\n", olIfname, err)
		}

		//    ip link set ${olIfname} up
		if err := netlink.LinkSetUp(oLink); err != nil {
			fmt.Printf("LinkSetUp on %s failed: %s\n",
				olIfname, err)
		}

		//    ip link set ${olIfname} arp on
		if err := netlink.LinkSetARPOn(oLink); err != nil {
			fmt.Printf("LinkSetARPOn on %s failed: %s\n", olIfname,
			err)
		}

		// Configure the EID on olIfname and set up a default route
		// for all fd00 EIDs
		//    ip addr add ${EID}/128 dev ${olIfname}
		EID := config.OverlayNetworkList[0].EID
		addr, err := netlink.ParseAddr(EID.String() + "/128")
		if err != nil {
			log.Printf("ParseAddr %s failed: %s\n", EID, err)
			return
		}
		if err := netlink.AddrAdd(oLink, addr); err != nil {
			log.Printf("AddrAdd %s failed: %s\n", EID, err)
		}

		//    ip route add fd00::/8 via fe80::1 src $eid dev $intf
		index := oLink.Attrs().Index
		_, ipnet, err := net.ParseCIDR("fd00::/8")
		if err != nil {
			log.Fatal("ParseCIDR fd00::/8 failed:\n", err)
		}
		via := net.ParseIP("fe80::1")

		if via == nil {
			log.Fatal("ParseIP fe80::1 failed: ", err)
		}
		// Need to do both an add and a change since we could have
		// a FAILED neighbor entry from a previous run and a down
		// uplink interface.
		//    ip nei add fe80::1 lladdr 0:0:0:0:0:1 dev $intf
		//    ip nei change fe80::1 lladdr 0:0:0:0:0:1 dev $intf
		hw, err = net.ParseMAC("00:00:00:00:00:01")
		if err != nil {
			log.Fatal("ParseMAC failed: ", err)
		}
		neigh := netlink.Neigh{LinkIndex: index, IP: via,
			HardwareAddr: hw, State: netlink.NUD_PERMANENT}
		if err := netlink.NeighAdd(&neigh); err != nil {
			fmt.Printf("NeighAdd fe80::1 failed: %s\n", err)
		}		
		if err := netlink.NeighSet(&neigh); err != nil {
			fmt.Printf("NeighSet fe80::1 failed: %s\n", err)
		}		

		// XXX needed fix in library for Src to work
		// XXX do we need src when using dbo1x0? Route is out that
		// interface. CHECK
		// /home/nordmark/gocode/src/github.com/vishvananda/netlink/route_linux.go
		// Replaced RTA_PREFSRC with RTA_SRC
		// XXX is this working? Don't see SRC in the added route on bobo
		// nor hikey
		rt := netlink.Route{Dst: ipnet, LinkIndex: index,
			Gw: via, Src: EID}
		// XXX hikey ended up with a route without the src
		// Could we have an issue with DAD delay?
		if err := netlink.RouteAdd(&rt); err != nil {
			fmt.Printf("RouteAdd fd00::/8 failed: %s\n", err)
		}

		// XXX NOTE: this hosts file is not read!
		// XXX easier when Zedmanager is in separate domU!
		// Create a hosts file for the overlay based on NameToEidList
		// Directory is /var/run/zedrouter/hosts.${OLIFNAME}
		// Each hostname in a separate file in directory to facilitate
		// adds and deletes
		hostsDirpath := globalRunDirname + "/hosts." + olIfname
		deleteHostsConfiglet(hostsDirpath, false)
		createHostsConfiglet(hostsDirpath, olConfig.NameToEidList)

		// Default EID ipset
		deleteEidIpsetConfiglet(olIfname, false)
		createEidIpsetConfiglet(olIfname, olConfig.NameToEidList)

		// Set up ACLs
		createACLConfiglet(olIfname, true, olConfig.ACLs, 6, "")

		// Create LISP configlets for IID and EID/signature		
		createLispConfiglet(lispRunDirname, true, olConfig.IID,
			olConfig.EID, olConfig.LispSignature,
			globalConfig.Uplink, olIfname, olIfname)
		status.OverlayNetworkList = make([]types.OverlayNetworkStatus,
			len(config.OverlayNetworkList))
		for i, _ := range config.OverlayNetworkList {
			status.OverlayNetworkList[i].OverlayNetworkConfig =
			config.OverlayNetworkList[i]
		}
		status.PendingAdd = false
		writeAppNetworkStatus(&status, statusFilename)
		log.Printf("handleCreate done for %s\n",
			config.DisplayName)
		return
	}
	
	// XXX introduce init function?
	status.OverlayNetworkList = make([]types.OverlayNetworkStatus,
		len(config.OverlayNetworkList))
	for i, _ := range config.OverlayNetworkList {
		status.OverlayNetworkList[i].OverlayNetworkConfig =
		config.OverlayNetworkList[i]
	}
	status.UnderlayNetworkList = make([]types.UnderlayNetworkStatus,
		len(config.UnderlayNetworkList))
	for i, _ := range config.UnderlayNetworkList {
		status.UnderlayNetworkList[i].UnderlayNetworkConfig =
		config.UnderlayNetworkList[i]
	}

	for i, olConfig := range config.OverlayNetworkList {
		olNum := i + 1
		fmt.Printf("olNum %d ACLs %v\n", olNum, olConfig.ACLs)

		EID := olConfig.EID
		olIfname := "bo" + strconv.Itoa(olNum) + "x" +
			strconv.Itoa(appNum)
		fmt.Printf("olIfname %s\n", olIfname)
		olAddr1 := "fd00::" + strconv.FormatInt(int64(olNum), 16) +
			":" + strconv.FormatInt(int64(appNum), 16)
		fmt.Printf("olAddr1 %s EID %s\n", olAddr1, EID)
		olMac := "00:16:3e:1:" + strconv.FormatInt(int64(olNum), 16) +
			":" + strconv.FormatInt(int64(appNum), 16)
		fmt.Printf("olMac %s\n", olMac)

		// Start clean
		attrs := netlink.NewLinkAttrs()
		attrs.Name = olIfname
		oLink := &netlink.Bridge{LinkAttrs: attrs}
		netlink.LinkDel(oLink)

		//    ip link add ${olIfname} type bridge
		attrs = netlink.NewLinkAttrs()
		attrs.Name = olIfname
		bridgeMac := fmt.Sprintf("00:16:3e:02:%02x:%02x", olNum, appNum)
		hw, err := net.ParseMAC(bridgeMac)
		if err != nil {
			log.Fatal("ParseMAC failed: ", bridgeMac, err)
		}
		attrs.HardwareAddr = hw
		oLink = &netlink.Bridge{LinkAttrs: attrs}
		if err := netlink.LinkAdd(oLink); err != nil {
			fmt.Printf("LinkAdd on %s failed: %s\n", olIfname, err)
		}

		//    ip link set ${olIfname} up
		if err := netlink.LinkSetUp(oLink); err != nil {
			fmt.Printf("LinkSetUp on %s failed: %s\n", olIfname, err)
		}

		//    ip addr add ${olAddr1}/128 dev ${olIfname}
		addr, err := netlink.ParseAddr(olAddr1 + "/128")
		if err != nil {
			fmt.Printf("ParseAddr %s failed: %s\n", olAddr1, err)
		}
		if err := netlink.AddrAdd(oLink, addr); err != nil {
			fmt.Printf("AddrAdd %s failed: %s\n", olAddr1, err)
		}

		//    ip -6 route add ${EID}/128 dev ${olIfname}
		_, ipnet, err := net.ParseCIDR(EID.String() + "/128")
		if err != nil {
			fmt.Printf("ParseCIDR %s failed: %v\n", EID, err)
		}
		fmt.Printf("oLink.Index %d\n", oLink.Index)
		rt := netlink.Route{Dst: ipnet, LinkIndex: oLink.Index}
		if err := netlink.RouteAdd(&rt); err != nil {
			fmt.Printf("RouteAdd %s failed: %s\n", EID, err)
		}

		// Write radvd configlet; start radvd
		cfgFilename := "radvd." + olIfname + ".conf"
		cfgPathname := "/etc/" + cfgFilename

		//    Start clean; kill just in case
		//    pkill -u radvd -f radvd.${OLIFNAME}.conf
		stopRadvd(cfgFilename, false)
		createRadvdConfiglet(cfgPathname, olIfname)
		startRadvd(cfgPathname, olIfname)

		// Create a hosts file for the overlay based on NameToEidList
		// Directory is /var/run/zedrouter/hosts.${OLIFNAME}
		// Each hostname in a separate file in directory to facilitate
		// adds and deletes
		hostsDirpath := globalRunDirname + "/hosts." + olIfname
		deleteHostsConfiglet(hostsDirpath, false)
		createHostsConfiglet(hostsDirpath, olConfig.NameToEidList)

		// Create default ipset with all the EIDs in NameToEidList
		// Can be used in ACLs by specifying "alleids" as match.
		deleteEidIpsetConfiglet(olIfname, false)
		createEidIpsetConfiglet(olIfname, olConfig.NameToEidList)
		
		// Set up ACLs before we setup dnsmasq
		createACLConfiglet(olIfname, false, olConfig.ACLs, 6, olAddr1)
		
		// Start clean
		cfgFilename = "dnsmasq." + olIfname + ".conf"
		cfgPathname = "/etc/" + cfgFilename
		stopDnsmasq(cfgFilename, false)
		createDnsmasqOverlayConfiglet(cfgPathname, olIfname, olAddr1,
			EID.String(), olMac, hostsDirpath, config.DisplayName)
		startDnsmasq(cfgPathname)

		// Create LISP configlets for IID and EID/signature		
		createLispConfiglet(lispRunDirname, false, olConfig.IID,
			olConfig.EID, olConfig.LispSignature,
			globalConfig.Uplink, olIfname, olIfname)

		// Add bridge parameters for Xen to Status
		olStatus := &status.OverlayNetworkList[olNum-1]
		olStatus.Bridge = olIfname
		olStatus.Vif = "n" + olIfname
		olStatus.Mac = olMac
	}

	for i, ulConfig := range config.UnderlayNetworkList {
		ulNum := i + 1
		if ulNum != 1 {
			// For now we only support one underlay interface
			// in app
			fmt.Printf("Ignoring multiple UnderlayNetwork\n")
			continue
		}
		fmt.Printf("ulNum %d ACLs %v\n", ulNum, ulConfig.ACLs)
		ulIfname := "bu" + strconv.Itoa(appNum)
		fmt.Printf("ulIfname %s\n", ulIfname)
		// Not clear how to handle multiple ul; use /30 prefix?
		ulAddr1 := "172.27." + strconv.Itoa(appNum) + ".1"
		ulAddr2 := "172.27." + strconv.Itoa(appNum) + ".2"
		fmt.Printf("ulAddr1 %s ulAddr2 %s\n", ulAddr1, ulAddr2)
		// Room to handle multiple underlays in 5th byte
		ulMac := "00:16:3e:0:0:" + strconv.FormatInt(int64(appNum), 16)	
		fmt.Printf("ulMac %s\n", ulMac)

		// Start clean
		attrs := netlink.NewLinkAttrs()
		attrs.Name = ulIfname
		uLink := &netlink.Bridge{LinkAttrs: attrs}
		netlink.LinkDel(uLink)

		//    ip link add ${ulIfname} type bridge
		attrs = netlink.NewLinkAttrs()
		attrs.Name = ulIfname
		bridgeMac := fmt.Sprintf("00:16:3e:04:00:%02x", appNum)
		hw, err := net.ParseMAC(bridgeMac)
		if err != nil {
			log.Fatal("ParseMAC failed: ", bridgeMac, err)
		}
		attrs.HardwareAddr = hw
		uLink = &netlink.Bridge{LinkAttrs: attrs}
		if err := netlink.LinkAdd(uLink); err != nil {
			fmt.Printf("LinkAdd on %s failed: %s\n", ulIfname, err)
		}
		//    ip link set ${ulIfname} up
		if err := netlink.LinkSetUp(uLink); err != nil {
			fmt.Printf("LinkSetUp on %s failed: %s\n", ulIfname, err)
		}
		//    ip addr add ${ulAddr1}/24 dev ${ulIfname}
		addr, err := netlink.ParseAddr(ulAddr1 + "/24")
		if err != nil {
			fmt.Printf("ParseAddr %s failed: %s\n", ulAddr1, err)
		}
		if err := netlink.AddrAdd(uLink, addr); err != nil {
			fmt.Printf("AddrAdd %s failed: %s\n", ulAddr1, err)
		}

		// Create iptables with optional ipset's based ACL
		createACLConfiglet(ulIfname, false, ulConfig.ACLs, 4, "")

		// Start clean
		cfgFilename := "dnsmasq." + ulIfname + ".conf"
		cfgPathname := "/etc/" + cfgFilename
		stopDnsmasq(cfgFilename, false)

		createDnsmasqUnderlayConfiglet(cfgPathname, ulIfname, ulAddr1,
			ulAddr2, ulMac, config.DisplayName)
		startDnsmasq(cfgPathname)

		// Add bridge parameters for Xen to Status
		ulStatus := &status.UnderlayNetworkList[ulNum-1]
		ulStatus.Bridge = ulIfname
		ulStatus.Vif = "n" + ulIfname
		ulStatus.Mac = ulMac
	}
	// Write out what we created to AppNetworkStatus
	status.PendingAdd = false
	writeAppNetworkStatus(&status, statusFilename)
	log.Printf("handleCreate done for %s\n", config.DisplayName)
}

// Note that modify will not touch the EID; just ACLs and NameToEidList
func handleModify(statusFilename string, config types.AppNetworkConfig,
	status types.AppNetworkStatus) {
	log.Printf("handleModify(%v) for %s\n",
		config.UUIDandVersion, config.DisplayName)

	if config.UUIDandVersion.Version == status.UUIDandVersion.Version {
		fmt.Printf("Same version %s for %s\n",
			config.UUIDandVersion.Version, statusFilename)
		return
	}

	appNum := status.AppNum
	fmt.Printf("handleModify appNum %d\n", appNum)

	// Check for unsupported changes
	if config.IsZedmanager != status.IsZedmanager {
		log.Println("Unsupported: IsZedmanager changed for ",
			config.UUIDandVersion)
		return
	}
	// XXX should we allow the addition of interfaces?
	// XXX can we allow the deletion (keep bridge around but disable intf?)
	// Inifinite lease time means painful unless domU sees down...
	if len(config.OverlayNetworkList) != status.OlNum {
		log.Println("Unsupported: Changed number of overlays for ",
			config.UUIDandVersion)
		return
	}
	if len(config.UnderlayNetworkList) != status.UlNum {
		log.Println("Unsupported: Changed number of underlays for ",
			config.UUIDandVersion)
		return
	}

	status.PendingModify = true
	status.UUIDandVersion = config.UUIDandVersion
	writeAppNetworkStatus(&status, statusFilename)

	if config.IsZedmanager {
		// XXX Which things can we update? NameToEid? ACLs?
		olConfig := config.OverlayNetworkList[0]
		olStatus := status.OverlayNetworkList[0]
		olNum := 1
		olIfname := "dbo" + strconv.Itoa(olNum) + "x" +
			strconv.Itoa(appNum)
		// Update hosts
		hostsDirpath := globalRunDirname + "/hosts." + olIfname
		updateHostsConfiglet(hostsDirpath, olStatus.NameToEidList,
			olConfig.NameToEidList)

		// Default EID ipset
		updateEidIpsetConfiglet(olIfname, olStatus.NameToEidList,
			olConfig.NameToEidList)

		// Update ACLs
		updateACLConfiglet(olIfname, true, olStatus.ACLs,
			olConfig.ACLs, 6, "")
		fmt.Printf("handleUpdate done for %s\n", config.DisplayName)
		return
	}
	
	// Look for ACL and NametoEidList changes in overlay
	// XXX flag others as errors; need lastError in status?
	for i, olConfig := range config.OverlayNetworkList {
		olNum := i + 1
		fmt.Printf("handleModify olNum %d\n", olNum)
		olIfname := "bo" + strconv.Itoa(olNum) + "x" +
			strconv.Itoa(appNum)
		olStatus := status.OverlayNetworkList[olNum-1]
		olAddr1 := "fd00::" + strconv.FormatInt(int64(olNum), 16) +
			":" + strconv.FormatInt(int64(appNum), 16)

		// Update hosts
		hostsDirpath := globalRunDirname + "/hosts." + olIfname
		updateHostsConfiglet(hostsDirpath, olStatus.NameToEidList,
			olConfig.NameToEidList)

		// Default EID ipset
		updateEidIpsetConfiglet(olIfname, olStatus.NameToEidList,
			olConfig.NameToEidList)

		// Update ACLs
		updateACLConfiglet(olIfname, false, olStatus.ACLs,
			olConfig.ACLs, 6, olAddr1)

		// XXX if the ACL update resulted in new eid sets, then
		// we need to restart dnsmasq (and update its ipset configs?
		// XXX get a return value from updateAclConfiglet to indicate
		// whether there were such changes?
		
		// Update any signature changes
		// XXX should we check that EID didn't change?
		// Create LISP configlets for IID and EID/signature		
		updateLispConfiglet(lispRunDirname, false, olConfig.IID,
			olConfig.EID, olConfig.LispSignature,
			globalConfig.Uplink, olIfname, olIfname)

	}
	// Look for ACL changes in underlay
	for i, ulConfig := range config.UnderlayNetworkList {
		ulNum := i + 1
		fmt.Printf("handleModify ulNum %d\n", ulNum)
		ulIfname := "bu" + strconv.Itoa(appNum)
		ulStatus := status.UnderlayNetworkList[ulNum-1]

		// Update ACLs
		updateACLConfiglet(ulIfname, false, ulStatus.ACLs,
			ulConfig.ACLs, 4, "")
	}
	// Write out what we modified to AppNetworkStatus
	status.OverlayNetworkList = make([]types.OverlayNetworkStatus,
		len(config.OverlayNetworkList))
	for i, _ := range config.OverlayNetworkList {
		status.OverlayNetworkList[i].OverlayNetworkConfig =
		config.OverlayNetworkList[i]
	}
	status.UnderlayNetworkList = make([]types.UnderlayNetworkStatus,
		len(config.UnderlayNetworkList))
	for i, _ := range config.UnderlayNetworkList {
		status.UnderlayNetworkList[i].UnderlayNetworkConfig =
		config.UnderlayNetworkList[i]
	}
	status.PendingModify = false
	writeAppNetworkStatus(&status, statusFilename)
	log.Printf("handleUpdate done for %s\n", config.DisplayName)
}

// Need the olNum and ulNum to delete and EID route to delete
func handleDelete(statusFilename string, status types.AppNetworkStatus) {
	log.Printf("handleDelete(%v) for %s\n",
		status.UUIDandVersion, status.DisplayName)

	appNum := status.AppNum
	maxOlNum := status.OlNum
	maxUlNum := status.UlNum
	fmt.Printf("handleDelete appNum %d maxOlNum %d maxUlNum %d\n",
		appNum, maxOlNum, maxUlNum)

	status.PendingDelete = true
	writeAppNetworkStatus(&status, statusFilename)

	if status.IsZedmanager {
		if len(status.OverlayNetworkList) != 1 ||
		   len(status.UnderlayNetworkList) != 0 {
			log.Println("Malformed IsZedmanager status; ignored")
			return
		}
		olNum := 1
		olStatus := &status.OverlayNetworkList[0]
		olIfname := "dbo" + strconv.Itoa(olNum) + "x" +
			strconv.Itoa(appNum)

		// Delete the address from loopback
		// Delete fd00::/8 route
		// Delete fe80::1 neighbor

		//    ip addr del ${EID}/128 dev ${olIfname}
		EID := status.OverlayNetworkList[0].EID
		addr, err := netlink.ParseAddr(EID.String() + "/128")
		if err != nil {
			fmt.Printf("ParseAddr %s failed: %s\n", EID, err)
			return
		}
		attrs := netlink.NewLinkAttrs()
		attrs.Name = olIfname
		oLink := &netlink.Dummy{LinkAttrs: attrs}
		// XXX can we skip explicit deletes and just remove the oLink?
		if err := netlink.AddrDel(oLink, addr); err != nil {
			fmt.Printf("AddrDel %s failed: %s\n", EID, err)
		}

		//    ip route del fd00::/8 via fe80::1 src $eid dev $intf
		index := oLink.Attrs().Index
		_, ipnet, err := net.ParseCIDR("fd00::/8")
		if err != nil {
			log.Fatal("ParseCIDR fd00::/8 failed:\n", err)
		}
		via := net.ParseIP("fe80::1")
		if via == nil {
			log.Fatal("ParseIP fe80::1 failed: ", err)
		}
		rt := netlink.Route{Dst: ipnet, LinkIndex: index,
			Gw: via, Src: EID}
		if err := netlink.RouteDel(&rt); err != nil {
			fmt.Printf("RouteDel fd00::/8 failed: %s\n", err)
		}
		//    ip nei del fe80::1 lladdr 0:0:0:0:0:1 dev $intf
		neigh := netlink.Neigh{LinkIndex: index, IP: via}
		if err := netlink.NeighDel(&neigh); err != nil {
			fmt.Printf("NeighDel fe80::1 failed: %s\n", err)
		}		

		// Remove link and associated addresses
		netlink.LinkDel(oLink)

		// Delete overlay hosts file
		hostsDirpath := globalRunDirname + "/hosts." + olIfname
		deleteHostsConfiglet(hostsDirpath, true)

		// Default EID ipset
		deleteEidIpsetConfiglet(olIfname, true)

		// Delete ACLs
		deleteACLConfiglet(olIfname, true, olStatus.ACLs, 6, "")

		// Delete LISP configlets
		deleteLispConfiglet(lispRunDirname, true, olStatus.IID,
			olStatus.EID, globalConfig.Uplink)
	} else {
		// Delete everything for overlay
		for olNum := 1; olNum <= maxOlNum; olNum++ {
			fmt.Printf("handleDelete olNum %d\n", olNum)
			olIfname := "bo" + strconv.Itoa(olNum) + "x" +
				strconv.Itoa(appNum)
			fmt.Printf("Deleting olIfname %s\n", olIfname)
			olAddr1 := "fd00::" + strconv.FormatInt(int64(olNum), 16) +
				":" + strconv.FormatInt(int64(appNum), 16)

			attrs := netlink.NewLinkAttrs()
			attrs.Name = olIfname
			oLink := &netlink.Bridge{LinkAttrs: attrs}
			// Remove link and associated addresses
			netlink.LinkDel(oLink)

			// radvd cleanup
			cfgFilename := "radvd." + olIfname + ".conf"
			cfgPathname := "/etc/" + cfgFilename
			stopRadvd(cfgFilename, true)
			deleteRadvdConfiglet(cfgPathname)

			// dnsmasgq cleanup
			cfgFilename = "dnsmasq." + olIfname + ".conf"
			cfgPathname = "/etc/" + cfgFilename
			stopDnsmasq(cfgFilename, true)
			deleteDnsmasqConfiglet(cfgPathname)
			
			// Need to check that index exists
			if len(status.OverlayNetworkList) >= olNum {
				olStatus := status.OverlayNetworkList[olNum-1]
				// Delete ACLs
				deleteACLConfiglet(olIfname, false,
					olStatus.ACLs, 6, olAddr1)

				// Delete LISP configlets
				deleteLispConfiglet(lispRunDirname, false,
					olStatus.IID, olStatus.EID,
					globalConfig.Uplink)
			} else {
				log.Println("Missing status for overlay %d; can not clean up ACLs and LISP\n",
					olNum)
			}
			

			// Delete overlay hosts file
			hostsDirpath := globalRunDirname + "/hosts." + olIfname
			deleteHostsConfiglet(hostsDirpath, true)

			// Default EID ipset
			deleteEidIpsetConfiglet(olIfname, true)
		}

		// XXX check if any IIDs are now unreferenced and delete them
		// XXX requires looking at all of configDir and statusDir

		// Delete everything in underlay
		for ulNum := 1; ulNum <= maxUlNum; ulNum++ {
			fmt.Printf("handleDelete ulNum %d\n", ulNum)
			ulIfname := "bu" + strconv.Itoa(appNum)
			fmt.Printf("Deleting ulIfname %s\n", ulIfname)
	
			attrs := netlink.NewLinkAttrs()
			attrs.Name = ulIfname
			uLink := &netlink.Bridge{LinkAttrs: attrs}
			// Remove link and associated addresses
			netlink.LinkDel(uLink)

			// dnsmasgq cleanup
			cfgFilename := "dnsmasq." + ulIfname + ".conf"
			cfgPathname := "/etc/" + cfgFilename
			stopDnsmasq(cfgFilename, true)
			deleteDnsmasqConfiglet(cfgPathname)

			// Delete ACLs
			// Need to check that index exists
			if len(status.UnderlayNetworkList) >= ulNum {
				ulStatus := status.UnderlayNetworkList[ulNum-1]
				deleteACLConfiglet(ulIfname, false,
					ulStatus.ACLs, 4, "")
			} else {
				log.Println("Missing status for underlay %d; can not clean up ACLs\n",
					ulNum)
			}
		}
	}
	// Write out what we modified to AppNetworkStatus aka delete
	if err := os.Remove(statusFilename); err != nil {
		log.Println("Failed to remove", statusFilename, err)
	}
	appNumFree(status.UUIDandVersion.UUID)
	log.Printf("handleDelete done for %s\n", status.DisplayName)
}

func pkillUserArgs(userName string, match string, printOnError bool) {
	cmd := "pkill"
	args := []string{
		"-u",
		userName,
		"-f",
		match,
	}
	_, err := exec.Command(cmd, args...).Output()
	if err != nil && printOnError {
		fmt.Printf("Command %v %v failed: %s\n", cmd, args, err)
	}
}
