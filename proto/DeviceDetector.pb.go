// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.12.3
// source: DeviceDetector.proto

package proto

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// The request message containing the user's name.
type BleRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Mac string `protobuf:"bytes,1,opt,name=mac,proto3" json:"mac,omitempty"`
}

func (x *BleRequest) Reset() {
	*x = BleRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BleRequest) ProtoMessage() {}

func (x *BleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BleRequest.ProtoReflect.Descriptor instead.
func (*BleRequest) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{0}
}

func (x *BleRequest) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

// The request message containing the user's name.
type AddressRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ip  string `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Mac string `protobuf:"bytes,2,opt,name=mac,proto3" json:"mac,omitempty"`
}

func (x *AddressRequest) Reset() {
	*x = AddressRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddressRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddressRequest) ProtoMessage() {}

func (x *AddressRequest) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddressRequest.ProtoReflect.Descriptor instead.
func (*AddressRequest) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{1}
}

func (x *AddressRequest) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *AddressRequest) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

type AddressesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Addresses []*AddressRequest `protobuf:"bytes,1,rep,name=addresses,proto3" json:"addresses,omitempty"`
}

func (x *AddressesRequest) Reset() {
	*x = AddressesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddressesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddressesRequest) ProtoMessage() {}

func (x *AddressesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddressesRequest.ProtoReflect.Descriptor instead.
func (*AddressesRequest) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{2}
}

func (x *AddressesRequest) GetAddresses() []*AddressRequest {
	if x != nil {
		return x.Addresses
	}
	return nil
}

// The response message containing the greetings
type Reply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Acknowledged bool `protobuf:"varint,1,opt,name=acknowledged,proto3" json:"acknowledged,omitempty"`
}

func (x *Reply) Reset() {
	*x = Reply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Reply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Reply) ProtoMessage() {}

func (x *Reply) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Reply.ProtoReflect.Descriptor instead.
func (*Reply) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{3}
}

func (x *Reply) GetAcknowledged() bool {
	if x != nil {
		return x.Acknowledged
	}
	return false
}

var File_DeviceDetector_proto protoreflect.FileDescriptor

var file_DeviceDetector_proto_rawDesc = []byte{
	0x0a, 0x14, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x1e, 0x0a,
	0x0a, 0x42, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6d,
	0x61, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x61, 0x63, 0x22, 0x32, 0x0a,
	0x0e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x12,
	0x10, 0x0a, 0x03, 0x6d, 0x61, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x61,
	0x63, 0x22, 0x47, 0x0a, 0x10, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x33, 0x0a, 0x09, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52,
	0x09, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x22, 0x2b, 0x0a, 0x05, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x12, 0x22, 0x0a, 0x0c, 0x61, 0x63, 0x6b, 0x6e, 0x6f, 0x77, 0x6c, 0x65, 0x64,
	0x67, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x61, 0x63, 0x6b, 0x6e, 0x6f,
	0x77, 0x6c, 0x65, 0x64, 0x67, 0x65, 0x64, 0x32, 0xa0, 0x01, 0x0a, 0x0c, 0x48, 0x6f, 0x6d, 0x65,
	0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x28, 0x0a, 0x03, 0x41, 0x63, 0x6b, 0x12,
	0x11, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x42, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79,
	0x22, 0x00, 0x12, 0x30, 0x0a, 0x07, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x15, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x09, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65,
	0x73, 0x12, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x07, 0x5a, 0x05, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_DeviceDetector_proto_rawDescOnce sync.Once
	file_DeviceDetector_proto_rawDescData = file_DeviceDetector_proto_rawDesc
)

func file_DeviceDetector_proto_rawDescGZIP() []byte {
	file_DeviceDetector_proto_rawDescOnce.Do(func() {
		file_DeviceDetector_proto_rawDescData = protoimpl.X.CompressGZIP(file_DeviceDetector_proto_rawDescData)
	})
	return file_DeviceDetector_proto_rawDescData
}

var file_DeviceDetector_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_DeviceDetector_proto_goTypes = []interface{}{
	(*BleRequest)(nil),       // 0: proto.BleRequest
	(*AddressRequest)(nil),   // 1: proto.AddressRequest
	(*AddressesRequest)(nil), // 2: proto.AddressesRequest
	(*Reply)(nil),            // 3: proto.Reply
}
var file_DeviceDetector_proto_depIdxs = []int32{
	1, // 0: proto.AddressesRequest.addresses:type_name -> proto.AddressRequest
	0, // 1: proto.HomeDetector.Ack:input_type -> proto.BleRequest
	1, // 2: proto.HomeDetector.Address:input_type -> proto.AddressRequest
	2, // 3: proto.HomeDetector.Addresses:input_type -> proto.AddressesRequest
	3, // 4: proto.HomeDetector.Ack:output_type -> proto.Reply
	3, // 5: proto.HomeDetector.Address:output_type -> proto.Reply
	3, // 6: proto.HomeDetector.Addresses:output_type -> proto.Reply
	4, // [4:7] is the sub-list for method output_type
	1, // [1:4] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_DeviceDetector_proto_init() }
func file_DeviceDetector_proto_init() {
	if File_DeviceDetector_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_DeviceDetector_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BleRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_DeviceDetector_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddressRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_DeviceDetector_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddressesRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_DeviceDetector_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Reply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_DeviceDetector_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_DeviceDetector_proto_goTypes,
		DependencyIndexes: file_DeviceDetector_proto_depIdxs,
		MessageInfos:      file_DeviceDetector_proto_msgTypes,
	}.Build()
	File_DeviceDetector_proto = out.File
	file_DeviceDetector_proto_rawDesc = nil
	file_DeviceDetector_proto_goTypes = nil
	file_DeviceDetector_proto_depIdxs = nil
}
