// Code generated by protoc-gen-go. DO NOT EDIT.
// source: zregister.proto

package zmet

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type ZRegisterResult int32

const (
	ZRegisterResult_ZRegNone        ZRegisterResult = 0
	ZRegisterResult_ZRegSuccess     ZRegisterResult = 1
	ZRegisterResult_ZRegNotActive   ZRegisterResult = 2
	ZRegisterResult_ZRegAlreadyDone ZRegisterResult = 3
	ZRegisterResult_ZRegDeviceNA    ZRegisterResult = 4
	ZRegisterResult_ZRegFailed      ZRegisterResult = 5
)

var ZRegisterResult_name = map[int32]string{
	0: "ZRegNone",
	1: "ZRegSuccess",
	2: "ZRegNotActive",
	3: "ZRegAlreadyDone",
	4: "ZRegDeviceNA",
	5: "ZRegFailed",
}
var ZRegisterResult_value = map[string]int32{
	"ZRegNone":        0,
	"ZRegSuccess":     1,
	"ZRegNotActive":   2,
	"ZRegAlreadyDone": 3,
	"ZRegDeviceNA":    4,
	"ZRegFailed":      5,
}

func (x ZRegisterResult) String() string {
	return proto.EnumName(ZRegisterResult_name, int32(x))
}
func (ZRegisterResult) EnumDescriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

type ZRegisterResp struct {
	Result ZRegisterResult `protobuf:"varint,2,opt,name=result,enum=ZRegisterResult" json:"result,omitempty"`
}

func (m *ZRegisterResp) Reset()                    { *m = ZRegisterResp{} }
func (m *ZRegisterResp) String() string            { return proto.CompactTextString(m) }
func (*ZRegisterResp) ProtoMessage()               {}
func (*ZRegisterResp) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *ZRegisterResp) GetResult() ZRegisterResult {
	if m != nil {
		return m.Result
	}
	return ZRegisterResult_ZRegNone
}

type ZRegisterMsg struct {
	OnBoardKey string `protobuf:"bytes,1,opt,name=onBoardKey" json:"onBoardKey,omitempty"`
	PemCert    []byte `protobuf:"bytes,2,opt,name=pemCert,proto3" json:"pemCert,omitempty"`
}

func (m *ZRegisterMsg) Reset()                    { *m = ZRegisterMsg{} }
func (m *ZRegisterMsg) String() string            { return proto.CompactTextString(m) }
func (*ZRegisterMsg) ProtoMessage()               {}
func (*ZRegisterMsg) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *ZRegisterMsg) GetOnBoardKey() string {
	if m != nil {
		return m.OnBoardKey
	}
	return ""
}

func (m *ZRegisterMsg) GetPemCert() []byte {
	if m != nil {
		return m.PemCert
	}
	return nil
}

func init() {
	proto.RegisterType((*ZRegisterResp)(nil), "ZRegisterResp")
	proto.RegisterType((*ZRegisterMsg)(nil), "ZRegisterMsg")
	proto.RegisterEnum("ZRegisterResult", ZRegisterResult_name, ZRegisterResult_value)
}

func init() { proto.RegisterFile("zregister.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 246 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x5c, 0x90, 0xc1, 0x4b, 0xc3, 0x30,
	0x14, 0xc6, 0xed, 0xd4, 0xa9, 0xcf, 0xba, 0xc6, 0xe7, 0xa5, 0x88, 0xc8, 0xd8, 0xa9, 0x78, 0x68,
	0x41, 0x4f, 0x1e, 0x3b, 0x87, 0x08, 0x62, 0x0f, 0xf1, 0xb6, 0x5b, 0x96, 0x3c, 0x6a, 0xa0, 0x5d,
	0x4a, 0x92, 0x0e, 0xd6, 0xbf, 0x5e, 0x6a, 0xa7, 0x14, 0x8f, 0xdf, 0xef, 0x47, 0xbe, 0x7c, 0x3c,
	0x88, 0x3a, 0x4b, 0xa5, 0x76, 0x9e, 0x6c, 0xda, 0x58, 0xe3, 0xcd, 0xe2, 0x19, 0xae, 0xd6, 0xfc,
	0x80, 0x38, 0xb9, 0x06, 0x13, 0x98, 0x5a, 0x72, 0x6d, 0xe5, 0xe3, 0xc9, 0x3c, 0x48, 0x66, 0x8f,
	0x2c, 0x1d, 0xfb, 0xb6, 0xf2, 0xfc, 0xe0, 0x17, 0x6f, 0x10, 0xfe, 0xa9, 0x0f, 0x57, 0xe2, 0x3d,
	0x80, 0xd9, 0x2e, 0x8d, 0xb0, 0xea, 0x9d, 0xf6, 0x71, 0x30, 0x0f, 0x92, 0x0b, 0x3e, 0x22, 0x18,
	0xc3, 0x59, 0x43, 0xf5, 0x0b, 0xd9, 0xa1, 0x3a, 0xe4, 0xbf, 0xf1, 0xa1, 0x83, 0xe8, 0xdf, 0x27,
	0x18, 0xc2, 0x79, 0x8f, 0x0a, 0xb3, 0x25, 0x76, 0x84, 0x11, 0x5c, 0xf6, 0xe9, 0xb3, 0x95, 0x92,
	0x9c, 0x63, 0x01, 0x5e, 0x0f, 0xb3, 0x0b, 0xe3, 0x73, 0xe9, 0xf5, 0x8e, 0xd8, 0x04, 0x6f, 0x86,
	0x92, 0xbc, 0xb2, 0x24, 0xd4, 0x7e, 0xd5, 0x3f, 0x3c, 0x46, 0x36, 0x6c, 0x5c, 0xd1, 0x4e, 0x4b,
	0x2a, 0x72, 0x76, 0x82, 0x33, 0x80, 0x9e, 0xbc, 0x0a, 0x5d, 0x91, 0x62, 0xa7, 0xcb, 0xbb, 0xf5,
	0x6d, 0xa9, 0xfd, 0x57, 0xbb, 0x49, 0xa5, 0xa9, 0xb3, 0x8e, 0x14, 0x29, 0x91, 0x89, 0x46, 0x67,
	0x5d, 0x4d, 0x7e, 0x33, 0xfd, 0xb9, 0xd2, 0xd3, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0x9b, 0x35,
	0x8b, 0xa6, 0x38, 0x01, 0x00, 0x00,
}
