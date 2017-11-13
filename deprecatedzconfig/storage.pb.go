// Code generated by protoc-gen-go. DO NOT EDIT.
// source: storage.proto

package deprecatedzconfig

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type SignatureInfo struct {
	Intermediatecertspem []byte `protobuf:"bytes,1,opt,name=intermediatecertspem,proto3" json:"intermediatecertspem,omitempty"`
	Signercertpem        []byte `protobuf:"bytes,2,opt,name=signercertpem,proto3" json:"signercertpem,omitempty"`
	Signature            []byte `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (m *SignatureInfo) Reset()                    { *m = SignatureInfo{} }
func (m *SignatureInfo) String() string            { return proto.CompactTextString(m) }
func (*SignatureInfo) ProtoMessage()               {}
func (*SignatureInfo) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{0} }

func (m *SignatureInfo) GetIntermediatecertspem() []byte {
	if m != nil {
		return m.Intermediatecertspem
	}
	return nil
}

func (m *SignatureInfo) GetSignercertpem() []byte {
	if m != nil {
		return m.Signercertpem
	}
	return nil
}

func (m *SignatureInfo) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

type StorageConfig struct {
	Downloadurl string         `protobuf:"bytes,1,opt,name=downloadurl" json:"downloadurl,omitempty"`
	Maxsize     uint32         `protobuf:"varint,2,opt,name=maxsize" json:"maxsize,omitempty"`
	Siginfo     *SignatureInfo `protobuf:"bytes,3,opt,name=siginfo" json:"siginfo,omitempty"`
	Imagesha256 string         `protobuf:"bytes,4,opt,name=imagesha256" json:"imagesha256,omitempty"`
	Readonly    bool           `protobuf:"varint,5,opt,name=readonly" json:"readonly,omitempty"`
	Preserve    bool           `protobuf:"varint,6,opt,name=preserve" json:"preserve,omitempty"`
	Format      string         `protobuf:"bytes,7,opt,name=format" json:"format,omitempty"`
	Devtype     string         `protobuf:"bytes,8,opt,name=devtype" json:"devtype,omitempty"`
	Target      string         `protobuf:"bytes,9,opt,name=target" json:"target,omitempty"`
}

func (m *StorageConfig) Reset()                    { *m = StorageConfig{} }
func (m *StorageConfig) String() string            { return proto.CompactTextString(m) }
func (*StorageConfig) ProtoMessage()               {}
func (*StorageConfig) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{1} }

func (m *StorageConfig) GetDownloadurl() string {
	if m != nil {
		return m.Downloadurl
	}
	return ""
}

func (m *StorageConfig) GetMaxsize() uint32 {
	if m != nil {
		return m.Maxsize
	}
	return 0
}

func (m *StorageConfig) GetSiginfo() *SignatureInfo {
	if m != nil {
		return m.Siginfo
	}
	return nil
}

func (m *StorageConfig) GetImagesha256() string {
	if m != nil {
		return m.Imagesha256
	}
	return ""
}

func (m *StorageConfig) GetReadonly() bool {
	if m != nil {
		return m.Readonly
	}
	return false
}

func (m *StorageConfig) GetPreserve() bool {
	if m != nil {
		return m.Preserve
	}
	return false
}

func (m *StorageConfig) GetFormat() string {
	if m != nil {
		return m.Format
	}
	return ""
}

func (m *StorageConfig) GetDevtype() string {
	if m != nil {
		return m.Devtype
	}
	return ""
}

func (m *StorageConfig) GetTarget() string {
	if m != nil {
		return m.Target
	}
	return ""
}

func init() {
	proto.RegisterType((*SignatureInfo)(nil), "SignatureInfo")
	proto.RegisterType((*StorageConfig)(nil), "StorageConfig")
}

func init() { proto.RegisterFile("storage.proto", fileDescriptor4) }

var fileDescriptor4 = []byte{
	// 329 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x6c, 0x91, 0xdd, 0x4a, 0xc3, 0x30,
	0x14, 0xc7, 0xe9, 0xd4, 0x7d, 0x64, 0xd6, 0x8b, 0x20, 0x12, 0x44, 0x70, 0x0c, 0xc1, 0x5d, 0x75,
	0x30, 0xd1, 0x07, 0xd0, 0x2b, 0xbd, 0xec, 0xee, 0xbc, 0xcb, 0x9a, 0xd3, 0x2c, 0xd0, 0x24, 0xe5,
	0x24, 0x9d, 0x6e, 0x2f, 0xe0, 0x33, 0xf8, 0xb6, 0x92, 0x74, 0xd5, 0x0d, 0xbc, 0xfc, 0x7f, 0x9c,
	0xfc, 0x4e, 0x12, 0x92, 0x3a, 0x6f, 0x91, 0x4b, 0xc8, 0x6a, 0xb4, 0xde, 0x4e, 0xbf, 0x12, 0x92,
	0x2e, 0x95, 0x34, 0xdc, 0x37, 0x08, 0xaf, 0xa6, 0xb4, 0x74, 0x41, 0x2e, 0x95, 0xf1, 0x80, 0x1a,
	0x84, 0xe2, 0x1e, 0x0a, 0x40, 0xef, 0x6a, 0xd0, 0x2c, 0x99, 0x24, 0xb3, 0xf3, 0xfc, 0xdf, 0x8c,
	0xde, 0x91, 0xd4, 0x29, 0x69, 0x00, 0x83, 0x13, 0xca, 0xbd, 0x58, 0x3e, 0x36, 0xe9, 0x0d, 0x19,
	0xb9, 0x0e, 0xc5, 0x4e, 0x62, 0xe3, 0xcf, 0x98, 0x7e, 0xf7, 0x48, 0xba, 0x6c, 0x77, 0x7b, 0xb1,
	0xa6, 0x54, 0x92, 0x4e, 0xc8, 0x58, 0xd8, 0x0f, 0x53, 0x59, 0x2e, 0x1a, 0xac, 0xe2, 0x02, 0xa3,
	0xfc, 0xd0, 0xa2, 0x8c, 0x0c, 0x34, 0xff, 0x74, 0x6a, 0x07, 0x91, 0x98, 0xe6, 0x9d, 0xa4, 0x33,
	0x32, 0x70, 0x4a, 0x2a, 0x53, 0xda, 0x48, 0x1a, 0x2f, 0x2e, 0xb2, 0xa3, 0x6b, 0xe6, 0x5d, 0x1c,
	0x28, 0x4a, 0x73, 0x09, 0x6e, 0xcd, 0x17, 0x8f, 0x4f, 0xec, 0xb4, 0xa5, 0x1c, 0x58, 0xf4, 0x9a,
	0x0c, 0x11, 0xb8, 0xb0, 0xa6, 0xda, 0xb2, 0xb3, 0x49, 0x32, 0x1b, 0xe6, 0xbf, 0x3a, 0x64, 0x35,
	0x82, 0x03, 0xdc, 0x00, 0xeb, 0xb7, 0x59, 0xa7, 0xe9, 0x15, 0xe9, 0x97, 0x16, 0x35, 0xf7, 0x6c,
	0x10, 0x0f, 0xdd, 0xab, 0xb0, 0xb5, 0x80, 0x8d, 0xdf, 0xd6, 0xc0, 0x86, 0x31, 0xe8, 0x64, 0x98,
	0xf0, 0x1c, 0x25, 0x78, 0x36, 0x6a, 0x27, 0x5a, 0xf5, 0xfc, 0x46, 0x6e, 0x0b, 0xab, 0xb3, 0x1d,
	0x08, 0x10, 0x3c, 0x2b, 0x2a, 0xdb, 0x88, 0xac, 0x09, 0x10, 0x55, 0xec, 0x3f, 0xf2, 0xfd, 0x5e,
	0x2a, 0xbf, 0x6e, 0x56, 0x59, 0x61, 0xf5, 0xbc, 0xed, 0xcd, 0x79, 0xad, 0xe6, 0x02, 0x6a, 0x84,
	0x82, 0x7b, 0x10, 0xbb, 0x22, 0xbe, 0xea, 0xaa, 0x1f, 0xfb, 0x0f, 0x3f, 0x01, 0x00, 0x00, 0xff,
	0xff, 0xc6, 0xae, 0x32, 0xee, 0x09, 0x02, 0x00, 0x00,
}
