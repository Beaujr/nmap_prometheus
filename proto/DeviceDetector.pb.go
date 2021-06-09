// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.12.3
// source: DeviceDetector.proto

package proto

import (
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
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
type StringRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *StringRequest) Reset() {
	*x = StringRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StringRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StringRequest) ProtoMessage() {}

func (x *StringRequest) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use StringRequest.ProtoReflect.Descriptor instead.
func (*StringRequest) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{0}
}

func (x *StringRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

// The request message containing the user's name.
type TimedCommands struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Executeat int64  `protobuf:"varint,2,opt,name=executeat,proto3" json:"executeat,omitempty"`
	Owner     string `protobuf:"bytes,3,opt,name=owner,proto3" json:"owner,omitempty"`
	Command   string `protobuf:"bytes,4,opt,name=command,proto3" json:"command,omitempty"`
	Executed  bool   `protobuf:"varint,5,opt,name=executed,proto3" json:"executed,omitempty"`
}

func (x *TimedCommands) Reset() {
	*x = TimedCommands{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimedCommands) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimedCommands) ProtoMessage() {}

func (x *TimedCommands) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use TimedCommands.ProtoReflect.Descriptor instead.
func (*TimedCommands) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{1}
}

func (x *TimedCommands) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TimedCommands) GetExecuteat() int64 {
	if x != nil {
		return x.Executeat
	}
	return 0
}

func (x *TimedCommands) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *TimedCommands) GetCommand() string {
	if x != nil {
		return x.Command
	}
	return ""
}

func (x *TimedCommands) GetExecuted() bool {
	if x != nil {
		return x.Executed
	}
	return false
}

// The request message containing the user's name.
type CQsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cqs []*TimedCommands `protobuf:"bytes,1,rep,name=cqs,proto3" json:"cqs,omitempty"`
}

func (x *CQsResponse) Reset() {
	*x = CQsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CQsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CQsResponse) ProtoMessage() {}

func (x *CQsResponse) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use CQsResponse.ProtoReflect.Descriptor instead.
func (*CQsResponse) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{2}
}

func (x *CQsResponse) GetCqs() []*TimedCommands {
	if x != nil {
		return x.Cqs
	}
	return nil
}

type TCsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Bles []*BleDevices `protobuf:"bytes,1,rep,name=bles,proto3" json:"bles,omitempty"`
}

func (x *TCsResponse) Reset() {
	*x = TCsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TCsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TCsResponse) ProtoMessage() {}

func (x *TCsResponse) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use TCsResponse.ProtoReflect.Descriptor instead.
func (*TCsResponse) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{3}
}

func (x *TCsResponse) GetBles() []*BleDevices {
	if x != nil {
		return x.Bles
	}
	return nil
}

type DevicesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Devices []*Devices `protobuf:"bytes,1,rep,name=devices,proto3" json:"devices,omitempty"`
}

func (x *DevicesResponse) Reset() {
	*x = DevicesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DevicesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DevicesResponse) ProtoMessage() {}

func (x *DevicesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DevicesResponse.ProtoReflect.Descriptor instead.
func (*DevicesResponse) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{4}
}

func (x *DevicesResponse) GetDevices() []*Devices {
	if x != nil {
		return x.Devices
	}
	return nil
}

type BleDevices struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string      `protobuf:"bytes,1,opt,name=Id,proto3" json:"Id,omitempty"`
	LastSeen int64       `protobuf:"varint,2,opt,name=LastSeen,proto3" json:"LastSeen,omitempty"`
	Commands []*Commands `protobuf:"bytes,3,rep,name=commands,proto3" json:"commands,omitempty"`
	Name     string      `protobuf:"bytes,4,opt,name=Name,proto3" json:"Name,omitempty"`
	Home     string      `protobuf:"bytes,5,opt,name=Home,proto3" json:"Home,omitempty"`
}

func (x *BleDevices) Reset() {
	*x = BleDevices{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BleDevices) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BleDevices) ProtoMessage() {}

func (x *BleDevices) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BleDevices.ProtoReflect.Descriptor instead.
func (*BleDevices) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{5}
}

func (x *BleDevices) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *BleDevices) GetLastSeen() int64 {
	if x != nil {
		return x.LastSeen
	}
	return 0
}

func (x *BleDevices) GetCommands() []*Commands {
	if x != nil {
		return x.Commands
	}
	return nil
}

func (x *BleDevices) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *BleDevices) GetHome() string {
	if x != nil {
		return x.Home
	}
	return ""
}

type Commands struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timeout int64  `protobuf:"varint,1,opt,name=Timeout,proto3" json:"Timeout,omitempty"`
	Command string `protobuf:"bytes,2,opt,name=Command,proto3" json:"Command,omitempty"`
	Id      string `protobuf:"bytes,3,opt,name=Id,proto3" json:"Id,omitempty"`
}

func (x *Commands) Reset() {
	*x = Commands{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Commands) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Commands) ProtoMessage() {}

func (x *Commands) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Commands.ProtoReflect.Descriptor instead.
func (*Commands) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{6}
}

func (x *Commands) GetTimeout() int64 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

func (x *Commands) GetCommand() string {
	if x != nil {
		return x.Command
	}
	return ""
}

func (x *Commands) GetId() string {
	if x != nil {
		return x.Id
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
		mi := &file_DeviceDetector_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddressRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddressRequest) ProtoMessage() {}

func (x *AddressRequest) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[7]
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
	return file_DeviceDetector_proto_rawDescGZIP(), []int{7}
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
		mi := &file_DeviceDetector_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddressesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddressesRequest) ProtoMessage() {}

func (x *AddressesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[8]
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
	return file_DeviceDetector_proto_rawDescGZIP(), []int{8}
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
		mi := &file_DeviceDetector_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Reply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Reply) ProtoMessage() {}

func (x *Reply) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[9]
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
	return file_DeviceDetector_proto_rawDescGZIP(), []int{9}
}

func (x *Reply) GetAcknowledged() bool {
	if x != nil {
		return x.Acknowledged
	}
	return false
}

type Devices struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id            *NetworkId `protobuf:"bytes,1,opt,name=Id,proto3" json:"Id,omitempty"`
	Home          string     `protobuf:"bytes,2,opt,name=Home,proto3" json:"Home,omitempty"`
	LastSeen      int64      `protobuf:"varint,3,opt,name=LastSeen,proto3" json:"LastSeen,omitempty"`
	Away          bool       `protobuf:"varint,4,opt,name=Away,proto3" json:"Away,omitempty"`
	Name          string     `protobuf:"bytes,5,opt,name=Name,proto3" json:"Name,omitempty"`
	Person        bool       `protobuf:"varint,6,opt,name=Person,proto3" json:"Person,omitempty"`
	Command       string     `protobuf:"bytes,7,opt,name=Command,proto3" json:"Command,omitempty"`
	Smart         bool       `protobuf:"varint,8,opt,name=Smart,proto3" json:"Smart,omitempty"`
	Manufacturer  string     `protobuf:"bytes,9,opt,name=Manufacturer,proto3" json:"Manufacturer,omitempty"`
	PresenceAware bool       `protobuf:"varint,10,opt,name=PresenceAware,proto3" json:"PresenceAware,omitempty"`
}

func (x *Devices) Reset() {
	*x = Devices{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Devices) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Devices) ProtoMessage() {}

func (x *Devices) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Devices.ProtoReflect.Descriptor instead.
func (*Devices) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{10}
}

func (x *Devices) GetId() *NetworkId {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Devices) GetHome() string {
	if x != nil {
		return x.Home
	}
	return ""
}

func (x *Devices) GetLastSeen() int64 {
	if x != nil {
		return x.LastSeen
	}
	return 0
}

func (x *Devices) GetAway() bool {
	if x != nil {
		return x.Away
	}
	return false
}

func (x *Devices) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Devices) GetPerson() bool {
	if x != nil {
		return x.Person
	}
	return false
}

func (x *Devices) GetCommand() string {
	if x != nil {
		return x.Command
	}
	return ""
}

func (x *Devices) GetSmart() bool {
	if x != nil {
		return x.Smart
	}
	return false
}

func (x *Devices) GetManufacturer() string {
	if x != nil {
		return x.Manufacturer
	}
	return ""
}

func (x *Devices) GetPresenceAware() bool {
	if x != nil {
		return x.PresenceAware
	}
	return false
}

type NetworkId struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ip   string `protobuf:"bytes,1,opt,name=Ip,proto3" json:"Ip,omitempty"`
	Mac  string `protobuf:"bytes,2,opt,name=Mac,proto3" json:"Mac,omitempty"`
	UUID string `protobuf:"bytes,3,opt,name=UUID,proto3" json:"UUID,omitempty"`
}

func (x *NetworkId) Reset() {
	*x = NetworkId{}
	if protoimpl.UnsafeEnabled {
		mi := &file_DeviceDetector_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NetworkId) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkId) ProtoMessage() {}

func (x *NetworkId) ProtoReflect() protoreflect.Message {
	mi := &file_DeviceDetector_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetworkId.ProtoReflect.Descriptor instead.
func (*NetworkId) Descriptor() ([]byte, []int) {
	return file_DeviceDetector_proto_rawDescGZIP(), []int{11}
}

func (x *NetworkId) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *NetworkId) GetMac() string {
	if x != nil {
		return x.Mac
	}
	return ""
}

func (x *NetworkId) GetUUID() string {
	if x != nil {
		return x.UUID
	}
	return ""
}

var File_DeviceDetector_proto protoreflect.FileDescriptor

var file_DeviceDetector_proto_rawDesc = []byte{
	0x0a, 0x14, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65,
	0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x21, 0x0a, 0x0d, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22, 0x89, 0x01,
	0x0a, 0x0d, 0x54, 0x69, 0x6d, 0x65, 0x64, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x73, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x1c, 0x0a, 0x09, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x61, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x09, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x61, 0x74, 0x12, 0x14, 0x0a,
	0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6f, 0x77,
	0x6e, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x1a, 0x0a,
	0x08, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x08, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x64, 0x22, 0x35, 0x0a, 0x0b, 0x43, 0x51, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x26, 0x0a, 0x03, 0x63, 0x71, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x64, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x73, 0x52, 0x03, 0x63, 0x71, 0x73,
	0x22, 0x34, 0x0a, 0x0b, 0x54, 0x43, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x25, 0x0a, 0x04, 0x62, 0x6c, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x42, 0x6c, 0x65, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73,
	0x52, 0x04, 0x62, 0x6c, 0x65, 0x73, 0x22, 0x3b, 0x0a, 0x0f, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x07, 0x64, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x52, 0x07, 0x64, 0x65, 0x76, 0x69,
	0x63, 0x65, 0x73, 0x22, 0x8d, 0x01, 0x0a, 0x0a, 0x42, 0x6c, 0x65, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x4c, 0x61, 0x73, 0x74, 0x53, 0x65, 0x65, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x4c, 0x61, 0x73, 0x74, 0x53, 0x65, 0x65, 0x6e, 0x12, 0x2b,
	0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64,
	0x73, 0x52, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x48, 0x6f, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x48,
	0x6f, 0x6d, 0x65, 0x22, 0x4e, 0x0a, 0x08, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x73, 0x12,
	0x18, 0x0a, 0x07, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x07, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x43, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x43, 0x6f, 0x6d, 0x6d,
	0x61, 0x6e, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x49, 0x64, 0x22, 0x32, 0x0a, 0x0e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x02, 0x69, 0x70, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x63, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6d, 0x61, 0x63, 0x22, 0x47, 0x0a, 0x10, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x33, 0x0a, 0x09, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x09, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73,
	0x22, 0x2b, 0x0a, 0x05, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x22, 0x0a, 0x0c, 0x61, 0x63, 0x6b,
	0x6e, 0x6f, 0x77, 0x6c, 0x65, 0x64, 0x67, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x0c, 0x61, 0x63, 0x6b, 0x6e, 0x6f, 0x77, 0x6c, 0x65, 0x64, 0x67, 0x65, 0x64, 0x22, 0x95, 0x02,
	0x0a, 0x07, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x12, 0x20, 0x0a, 0x02, 0x49, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x6e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x49, 0x64, 0x52, 0x02, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x48,
	0x6f, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x48, 0x6f, 0x6d, 0x65, 0x12,
	0x1a, 0x0a, 0x08, 0x4c, 0x61, 0x73, 0x74, 0x53, 0x65, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x08, 0x4c, 0x61, 0x73, 0x74, 0x53, 0x65, 0x65, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x41,
	0x77, 0x61, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x41, 0x77, 0x61, 0x79, 0x12,
	0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x06, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x43,
	0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x43, 0x6f,
	0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x53, 0x6d, 0x61, 0x72, 0x74, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x53, 0x6d, 0x61, 0x72, 0x74, 0x12, 0x22, 0x0a, 0x0c, 0x4d,
	0x61, 0x6e, 0x75, 0x66, 0x61, 0x63, 0x74, 0x75, 0x72, 0x65, 0x72, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0c, 0x4d, 0x61, 0x6e, 0x75, 0x66, 0x61, 0x63, 0x74, 0x75, 0x72, 0x65, 0x72, 0x12,
	0x24, 0x0a, 0x0d, 0x50, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x63, 0x65, 0x41, 0x77, 0x61, 0x72, 0x65,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x50, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x63, 0x65,
	0x41, 0x77, 0x61, 0x72, 0x65, 0x22, 0x41, 0x0a, 0x09, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x49, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x49, 0x70, 0x12, 0x10, 0x0a, 0x03, 0x4d, 0x61, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x4d, 0x61, 0x63, 0x12, 0x12, 0x0a, 0x04, 0x55, 0x55, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x55, 0x55, 0x49, 0x44, 0x32, 0xde, 0x04, 0x0a, 0x0c, 0x48, 0x6f, 0x6d,
	0x65, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x2b, 0x0a, 0x03, 0x41, 0x63, 0x6b,
	0x12, 0x14, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x30, 0x0a, 0x07, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x09, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x65, 0x73, 0x12, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x41,
	0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x64, 0x43, 0x6f, 0x6d, 0x6d, 0x61,
	0x6e, 0x64, 0x73, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x12, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x54, 0x43, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x40, 0x0a, 0x10, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64,
	0x51, 0x75, 0x65, 0x75, 0x65, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x12, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x51, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x12, 0x3f, 0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x73, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x3a, 0x0a, 0x12, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x43, 0x6f,
	0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x51, 0x75, 0x65, 0x75, 0x65, 0x12, 0x14, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00,
	0x12, 0x3a, 0x0a, 0x12, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x64, 0x43,
	0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x14, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x3d, 0x0a, 0x15,
	0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x64, 0x43, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x73, 0x12, 0x14, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x3c, 0x0a, 0x14, 0x43,
	0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x64, 0x43, 0x6f, 0x6d, 0x6d,
	0x61, 0x6e, 0x64, 0x12, 0x14, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x07, 0x5a, 0x05, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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

var file_DeviceDetector_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_DeviceDetector_proto_goTypes = []interface{}{
	(*StringRequest)(nil),    // 0: proto.StringRequest
	(*TimedCommands)(nil),    // 1: proto.TimedCommands
	(*CQsResponse)(nil),      // 2: proto.CQsResponse
	(*TCsResponse)(nil),      // 3: proto.TCsResponse
	(*DevicesResponse)(nil),  // 4: proto.DevicesResponse
	(*BleDevices)(nil),       // 5: proto.BleDevices
	(*Commands)(nil),         // 6: proto.Commands
	(*AddressRequest)(nil),   // 7: proto.AddressRequest
	(*AddressesRequest)(nil), // 8: proto.AddressesRequest
	(*Reply)(nil),            // 9: proto.Reply
	(*Devices)(nil),          // 10: proto.Devices
	(*NetworkId)(nil),        // 11: proto.networkId
	(*empty.Empty)(nil),      // 12: google.protobuf.Empty
}
var file_DeviceDetector_proto_depIdxs = []int32{
	1,  // 0: proto.CQsResponse.cqs:type_name -> proto.TimedCommands
	5,  // 1: proto.TCsResponse.bles:type_name -> proto.BleDevices
	10, // 2: proto.DevicesResponse.devices:type_name -> proto.Devices
	6,  // 3: proto.BleDevices.commands:type_name -> proto.Commands
	7,  // 4: proto.AddressesRequest.addresses:type_name -> proto.AddressRequest
	11, // 5: proto.Devices.Id:type_name -> proto.networkId
	0,  // 6: proto.HomeDetector.Ack:input_type -> proto.StringRequest
	7,  // 7: proto.HomeDetector.Address:input_type -> proto.AddressRequest
	8,  // 8: proto.HomeDetector.Addresses:input_type -> proto.AddressesRequest
	12, // 9: proto.HomeDetector.ListTimedCommands:input_type -> google.protobuf.Empty
	12, // 10: proto.HomeDetector.ListCommandQueue:input_type -> google.protobuf.Empty
	12, // 11: proto.HomeDetector.ListDevices:input_type -> google.protobuf.Empty
	0,  // 12: proto.HomeDetector.DeleteCommandQueue:input_type -> proto.StringRequest
	0,  // 13: proto.HomeDetector.DeleteTimedCommand:input_type -> proto.StringRequest
	0,  // 14: proto.HomeDetector.CompleteTimedCommands:input_type -> proto.StringRequest
	0,  // 15: proto.HomeDetector.CompleteTimedCommand:input_type -> proto.StringRequest
	9,  // 16: proto.HomeDetector.Ack:output_type -> proto.Reply
	9,  // 17: proto.HomeDetector.Address:output_type -> proto.Reply
	9,  // 18: proto.HomeDetector.Addresses:output_type -> proto.Reply
	3,  // 19: proto.HomeDetector.ListTimedCommands:output_type -> proto.TCsResponse
	2,  // 20: proto.HomeDetector.ListCommandQueue:output_type -> proto.CQsResponse
	4,  // 21: proto.HomeDetector.ListDevices:output_type -> proto.DevicesResponse
	9,  // 22: proto.HomeDetector.DeleteCommandQueue:output_type -> proto.Reply
	9,  // 23: proto.HomeDetector.DeleteTimedCommand:output_type -> proto.Reply
	9,  // 24: proto.HomeDetector.CompleteTimedCommands:output_type -> proto.Reply
	9,  // 25: proto.HomeDetector.CompleteTimedCommand:output_type -> proto.Reply
	16, // [16:26] is the sub-list for method output_type
	6,  // [6:16] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_DeviceDetector_proto_init() }
func file_DeviceDetector_proto_init() {
	if File_DeviceDetector_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_DeviceDetector_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StringRequest); i {
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
			switch v := v.(*TimedCommands); i {
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
			switch v := v.(*CQsResponse); i {
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
			switch v := v.(*TCsResponse); i {
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
		file_DeviceDetector_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DevicesResponse); i {
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
		file_DeviceDetector_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BleDevices); i {
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
		file_DeviceDetector_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Commands); i {
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
		file_DeviceDetector_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
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
		file_DeviceDetector_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
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
		file_DeviceDetector_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
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
		file_DeviceDetector_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Devices); i {
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
		file_DeviceDetector_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NetworkId); i {
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
			NumMessages:   12,
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
