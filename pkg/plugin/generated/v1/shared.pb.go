// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.18.1
// source: shared.proto

package generated

import (
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

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_shared_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_shared_proto_rawDescGZIP(), []int{0}
}

type ProtectedEntityID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PeType     string                     `protobuf:"bytes,1,opt,name=peType,proto3" json:"peType,omitempty"`
	Id         string                     `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	SnapshotID *ProtectedEntitySnapshotID `protobuf:"bytes,3,opt,name=snapshotID,proto3" json:"snapshotID,omitempty"`
}

func (x *ProtectedEntityID) Reset() {
	*x = ProtectedEntityID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtectedEntityID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtectedEntityID) ProtoMessage() {}

func (x *ProtectedEntityID) ProtoReflect() protoreflect.Message {
	mi := &file_shared_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtectedEntityID.ProtoReflect.Descriptor instead.
func (*ProtectedEntityID) Descriptor() ([]byte, []int) {
	return file_shared_proto_rawDescGZIP(), []int{1}
}

func (x *ProtectedEntityID) GetPeType() string {
	if x != nil {
		return x.PeType
	}
	return ""
}

func (x *ProtectedEntityID) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ProtectedEntityID) GetSnapshotID() *ProtectedEntitySnapshotID {
	if x != nil {
		return x.SnapshotID
	}
	return nil
}

type ProtectedEntitySnapshotID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *ProtectedEntitySnapshotID) Reset() {
	*x = ProtectedEntitySnapshotID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtectedEntitySnapshotID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtectedEntitySnapshotID) ProtoMessage() {}

func (x *ProtectedEntitySnapshotID) ProtoReflect() protoreflect.Message {
	mi := &file_shared_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtectedEntitySnapshotID.ProtoReflect.Descriptor instead.
func (*ProtectedEntitySnapshotID) Descriptor() ([]byte, []int) {
	return file_shared_proto_rawDescGZIP(), []int{2}
}

func (x *ProtectedEntitySnapshotID) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type DataTransport struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransportType string            `protobuf:"bytes,1,opt,name=transportType,proto3" json:"transportType,omitempty"`
	Params        map[string]string `protobuf:"bytes,2,rep,name=params,proto3" json:"params,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *DataTransport) Reset() {
	*x = DataTransport{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataTransport) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataTransport) ProtoMessage() {}

func (x *DataTransport) ProtoReflect() protoreflect.Message {
	mi := &file_shared_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataTransport.ProtoReflect.Descriptor instead.
func (*DataTransport) Descriptor() ([]byte, []int) {
	return file_shared_proto_rawDescGZIP(), []int{3}
}

func (x *DataTransport) GetTransportType() string {
	if x != nil {
		return x.TransportType
	}
	return ""
}

func (x *DataTransport) GetParams() map[string]string {
	if x != nil {
		return x.Params
	}
	return nil
}

type ProtectedEntityInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id                 *ProtectedEntityID   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name               string               `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Size               int64                `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	DataTransports     []*DataTransport     `protobuf:"bytes,4,rep,name=dataTransports,proto3" json:"dataTransports,omitempty"`
	MetadataTransports []*DataTransport     `protobuf:"bytes,5,rep,name=metadataTransports,proto3" json:"metadataTransports,omitempty"`
	CombinedTransports []*DataTransport     `protobuf:"bytes,6,rep,name=combinedTransports,proto3" json:"combinedTransports,omitempty"`
	ComponentIDs       []*ProtectedEntityID `protobuf:"bytes,7,rep,name=componentIDs,proto3" json:"componentIDs,omitempty"`
}

func (x *ProtectedEntityInfo) Reset() {
	*x = ProtectedEntityInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtectedEntityInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtectedEntityInfo) ProtoMessage() {}

func (x *ProtectedEntityInfo) ProtoReflect() protoreflect.Message {
	mi := &file_shared_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtectedEntityInfo.ProtoReflect.Descriptor instead.
func (*ProtectedEntityInfo) Descriptor() ([]byte, []int) {
	return file_shared_proto_rawDescGZIP(), []int{4}
}

func (x *ProtectedEntityInfo) GetId() *ProtectedEntityID {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *ProtectedEntityInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ProtectedEntityInfo) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *ProtectedEntityInfo) GetDataTransports() []*DataTransport {
	if x != nil {
		return x.DataTransports
	}
	return nil
}

func (x *ProtectedEntityInfo) GetMetadataTransports() []*DataTransport {
	if x != nil {
		return x.MetadataTransports
	}
	return nil
}

func (x *ProtectedEntityInfo) GetCombinedTransports() []*DataTransport {
	if x != nil {
		return x.CombinedTransports
	}
	return nil
}

func (x *ProtectedEntityInfo) GetComponentIDs() []*ProtectedEntityID {
	if x != nil {
		return x.ComponentIDs
	}
	return nil
}

var File_shared_proto protoreflect.FileDescriptor

var file_shared_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09,
	0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x22, 0x81, 0x01, 0x0a, 0x11, 0x50, 0x72, 0x6f, 0x74, 0x65, 0x63, 0x74, 0x65, 0x64,
	0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x49, 0x44, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x65, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x44, 0x0a, 0x0a, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x49, 0x44, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64,
	0x2e, 0x50, 0x72, 0x6f, 0x74, 0x65, 0x63, 0x74, 0x65, 0x64, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x49, 0x44, 0x52, 0x0a, 0x73, 0x6e, 0x61, 0x70,
	0x73, 0x68, 0x6f, 0x74, 0x49, 0x44, 0x22, 0x2b, 0x0a, 0x19, 0x50, 0x72, 0x6f, 0x74, 0x65, 0x63,
	0x74, 0x65, 0x64, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f,
	0x74, 0x49, 0x44, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x22, 0xae, 0x01, 0x0a, 0x0d, 0x44, 0x61, 0x74, 0x61, 0x54, 0x72, 0x61, 0x6e,
	0x73, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x24, 0x0a, 0x0d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f,
	0x72, 0x74, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x3c, 0x0a, 0x06, 0x70,
	0x61, 0x72, 0x61, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x67, 0x65,
	0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x54, 0x72, 0x61, 0x6e,
	0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x39, 0x0a, 0x0b, 0x50, 0x61, 0x72,
	0x61, 0x6d, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x22, 0x83, 0x03, 0x0a, 0x13, 0x50, 0x72, 0x6f, 0x74, 0x65, 0x63, 0x74,
	0x65, 0x64, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x2c, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72,
	0x61, 0x74, 0x65, 0x64, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x65, 0x63, 0x74, 0x65, 0x64, 0x45, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x49, 0x44, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x69,
	0x7a, 0x65, 0x12, 0x40, 0x0a, 0x0e, 0x64, 0x61, 0x74, 0x61, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x67, 0x65, 0x6e,
	0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x54, 0x72, 0x61, 0x6e, 0x73,
	0x70, 0x6f, 0x72, 0x74, 0x52, 0x0e, 0x64, 0x61, 0x74, 0x61, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x73, 0x12, 0x48, 0x0a, 0x12, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x18, 0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x44, 0x61, 0x74,
	0x61, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x52, 0x12, 0x6d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x12, 0x48,
	0x0a, 0x12, 0x63, 0x6f, 0x6d, 0x62, 0x69, 0x6e, 0x65, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x67, 0x65, 0x6e,
	0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x54, 0x72, 0x61, 0x6e, 0x73,
	0x70, 0x6f, 0x72, 0x74, 0x52, 0x12, 0x63, 0x6f, 0x6d, 0x62, 0x69, 0x6e, 0x65, 0x64, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x12, 0x40, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x70,
	0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c,
	0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x65,
	0x63, 0x74, 0x65, 0x64, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x49, 0x44, 0x52, 0x0c, 0x63, 0x6f,
	0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x73, 0x42, 0x38, 0x5a, 0x36, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x6d, 0x77, 0x61, 0x72, 0x65, 0x2d,
	0x74, 0x61, 0x6e, 0x7a, 0x75, 0x2f, 0x61, 0x73, 0x74, 0x72, 0x6f, 0x6c, 0x61, 0x62, 0x65, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72,
	0x61, 0x74, 0x65, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_shared_proto_rawDescOnce sync.Once
	file_shared_proto_rawDescData = file_shared_proto_rawDesc
)

func file_shared_proto_rawDescGZIP() []byte {
	file_shared_proto_rawDescOnce.Do(func() {
		file_shared_proto_rawDescData = protoimpl.X.CompressGZIP(file_shared_proto_rawDescData)
	})
	return file_shared_proto_rawDescData
}

var file_shared_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_shared_proto_goTypes = []interface{}{
	(*Empty)(nil),                     // 0: generated.Empty
	(*ProtectedEntityID)(nil),         // 1: generated.ProtectedEntityID
	(*ProtectedEntitySnapshotID)(nil), // 2: generated.ProtectedEntitySnapshotID
	(*DataTransport)(nil),             // 3: generated.DataTransport
	(*ProtectedEntityInfo)(nil),       // 4: generated.ProtectedEntityInfo
	nil,                               // 5: generated.DataTransport.ParamsEntry
}
var file_shared_proto_depIdxs = []int32{
	2, // 0: generated.ProtectedEntityID.snapshotID:type_name -> generated.ProtectedEntitySnapshotID
	5, // 1: generated.DataTransport.params:type_name -> generated.DataTransport.ParamsEntry
	1, // 2: generated.ProtectedEntityInfo.id:type_name -> generated.ProtectedEntityID
	3, // 3: generated.ProtectedEntityInfo.dataTransports:type_name -> generated.DataTransport
	3, // 4: generated.ProtectedEntityInfo.metadataTransports:type_name -> generated.DataTransport
	3, // 5: generated.ProtectedEntityInfo.combinedTransports:type_name -> generated.DataTransport
	1, // 6: generated.ProtectedEntityInfo.componentIDs:type_name -> generated.ProtectedEntityID
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_shared_proto_init() }
func file_shared_proto_init() {
	if File_shared_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_shared_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
		file_shared_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtectedEntityID); i {
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
		file_shared_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtectedEntitySnapshotID); i {
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
		file_shared_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataTransport); i {
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
		file_shared_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtectedEntityInfo); i {
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
			RawDescriptor: file_shared_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_shared_proto_goTypes,
		DependencyIndexes: file_shared_proto_depIdxs,
		MessageInfos:      file_shared_proto_msgTypes,
	}.Build()
	File_shared_proto = out.File
	file_shared_proto_rawDesc = nil
	file_shared_proto_goTypes = nil
	file_shared_proto_depIdxs = nil
}
