// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: signer/signer.proto

package pb_signer

import (
	networks "github.com/BoostyLabs/casper-eth-bridge/boosty-communication/go-gen/networks"
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

type DataType int32

const (
	DataType_DT_TRANSACTION DataType = 0
	DataType_DT_SIGNATURE   DataType = 1
)

// Enum value maps for DataType.
var (
	DataType_name = map[int32]string{
		0: "DT_TRANSACTION",
		1: "DT_SIGNATURE",
	}
	DataType_value = map[string]int32{
		"DT_TRANSACTION": 0,
		"DT_SIGNATURE":   1,
	}
)

func (x DataType) Enum() *DataType {
	p := new(DataType)
	*p = x
	return p
}

func (x DataType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DataType) Descriptor() protoreflect.EnumDescriptor {
	return file_signer_signer_proto_enumTypes[0].Descriptor()
}

func (DataType) Type() protoreflect.EnumType {
	return &file_signer_signer_proto_enumTypes[0]
}

func (x DataType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DataType.Descriptor instead.
func (DataType) EnumDescriptor() ([]byte, []int) {
	return file_signer_signer_proto_rawDescGZIP(), []int{0}
}

type SignRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NetworkId networks.NetworkType `protobuf:"varint,1,opt,name=network_id,json=networkId,proto3,enum=tricorn.NetworkType" json:"network_id,omitempty"`
	DataType  DataType             `protobuf:"varint,2,opt,name=data_type,json=dataType,proto3,enum=tricorn.DataType" json:"data_type,omitempty"`
	Data      []byte               `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *SignRequest) Reset() {
	*x = SignRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signer_signer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignRequest) ProtoMessage() {}

func (x *SignRequest) ProtoReflect() protoreflect.Message {
	mi := &file_signer_signer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignRequest.ProtoReflect.Descriptor instead.
func (*SignRequest) Descriptor() ([]byte, []int) {
	return file_signer_signer_proto_rawDescGZIP(), []int{0}
}

func (x *SignRequest) GetNetworkId() networks.NetworkType {
	if x != nil {
		return x.NetworkId
	}
	return networks.NetworkType(0)
}

func (x *SignRequest) GetDataType() DataType {
	if x != nil {
		return x.DataType
	}
	return DataType_DT_TRANSACTION
}

func (x *SignRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type Signature struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NetworkId networks.NetworkType `protobuf:"varint,1,opt,name=network_id,json=networkId,proto3,enum=tricorn.NetworkType" json:"network_id,omitempty"`
	Signature []byte               `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *Signature) Reset() {
	*x = Signature{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signer_signer_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Signature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Signature) ProtoMessage() {}

func (x *Signature) ProtoReflect() protoreflect.Message {
	mi := &file_signer_signer_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Signature.ProtoReflect.Descriptor instead.
func (*Signature) Descriptor() ([]byte, []int) {
	return file_signer_signer_proto_rawDescGZIP(), []int{1}
}

func (x *Signature) GetNetworkId() networks.NetworkType {
	if x != nil {
		return x.NetworkId
	}
	return networks.NetworkType(0)
}

func (x *Signature) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type PublicKeyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NetworkId networks.NetworkType `protobuf:"varint,1,opt,name=network_id,json=networkId,proto3,enum=tricorn.NetworkType" json:"network_id,omitempty"`
}

func (x *PublicKeyRequest) Reset() {
	*x = PublicKeyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signer_signer_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PublicKeyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublicKeyRequest) ProtoMessage() {}

func (x *PublicKeyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_signer_signer_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublicKeyRequest.ProtoReflect.Descriptor instead.
func (*PublicKeyRequest) Descriptor() ([]byte, []int) {
	return file_signer_signer_proto_rawDescGZIP(), []int{2}
}

func (x *PublicKeyRequest) GetNetworkId() networks.NetworkType {
	if x != nil {
		return x.NetworkId
	}
	return networks.NetworkType(0)
}

type PublicKeyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PublicKey []byte `protobuf:"bytes,1,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
}

func (x *PublicKeyResponse) Reset() {
	*x = PublicKeyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signer_signer_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PublicKeyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublicKeyResponse) ProtoMessage() {}

func (x *PublicKeyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_signer_signer_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublicKeyResponse.ProtoReflect.Descriptor instead.
func (*PublicKeyResponse) Descriptor() ([]byte, []int) {
	return file_signer_signer_proto_rawDescGZIP(), []int{3}
}

func (x *PublicKeyResponse) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

var File_signer_signer_proto protoreflect.FileDescriptor

var file_signer_signer_proto_rawDesc = []byte{
	0x0a, 0x13, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x74, 0x72, 0x69, 0x63, 0x6f, 0x72, 0x6e, 0x1a, 0x17,
	0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x2f, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x86, 0x01, 0x0a, 0x0b, 0x53, 0x69, 0x67, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x33, 0x0a, 0x0a, 0x6e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x74, 0x72,
	0x69, 0x63, 0x6f, 0x72, 0x6e, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x54, 0x79, 0x70,
	0x65, 0x52, 0x09, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x49, 0x64, 0x12, 0x2e, 0x0a, 0x09,
	0x64, 0x61, 0x74, 0x61, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x11, 0x2e, 0x74, 0x72, 0x69, 0x63, 0x6f, 0x72, 0x6e, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x08, 0x64, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x22, 0x5e, 0x0a, 0x09, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x33, 0x0a,
	0x0a, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x14, 0x2e, 0x74, 0x72, 0x69, 0x63, 0x6f, 0x72, 0x6e, 0x2e, 0x4e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x52, 0x09, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x49, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x22, 0x47, 0x0a, 0x10, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x33, 0x0a, 0x0a, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x74, 0x72, 0x69, 0x63, 0x6f,
	0x72, 0x6e, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x52, 0x09,
	0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x49, 0x64, 0x22, 0x32, 0x0a, 0x11, 0x50, 0x75, 0x62,
	0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1d,
	0x0a, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x2a, 0x30, 0x0a,
	0x08, 0x44, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x0e, 0x44, 0x54, 0x5f,
	0x54, 0x52, 0x41, 0x4e, 0x53, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x00, 0x12, 0x10, 0x0a,
	0x0c, 0x44, 0x54, 0x5f, 0x53, 0x49, 0x47, 0x4e, 0x41, 0x54, 0x55, 0x52, 0x45, 0x10, 0x01, 0x42,
	0x56, 0x5a, 0x54, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x42, 0x6f,
	0x6f, 0x73, 0x74, 0x79, 0x4c, 0x61, 0x62, 0x73, 0x2f, 0x63, 0x61, 0x73, 0x70, 0x65, 0x72, 0x2d,
	0x65, 0x74, 0x68, 0x2d, 0x62, 0x72, 0x69, 0x64, 0x67, 0x65, 0x2f, 0x62, 0x6f, 0x6f, 0x73, 0x74,
	0x79, 0x2d, 0x63, 0x6f, 0x6d, 0x6d, 0x75, 0x6e, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f,
	0x67, 0x6f, 0x2d, 0x67, 0x65, 0x6e, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x3b, 0x70, 0x62,
	0x5f, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_signer_signer_proto_rawDescOnce sync.Once
	file_signer_signer_proto_rawDescData = file_signer_signer_proto_rawDesc
)

func file_signer_signer_proto_rawDescGZIP() []byte {
	file_signer_signer_proto_rawDescOnce.Do(func() {
		file_signer_signer_proto_rawDescData = protoimpl.X.CompressGZIP(file_signer_signer_proto_rawDescData)
	})
	return file_signer_signer_proto_rawDescData
}

var file_signer_signer_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_signer_signer_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_signer_signer_proto_goTypes = []interface{}{
	(DataType)(0),             // 0: tricorn.DataType
	(*SignRequest)(nil),       // 1: tricorn.SignRequest
	(*Signature)(nil),         // 2: tricorn.Signature
	(*PublicKeyRequest)(nil),  // 3: tricorn.PublicKeyRequest
	(*PublicKeyResponse)(nil), // 4: tricorn.PublicKeyResponse
	(networks.NetworkType)(0), // 5: tricorn.NetworkType
}
var file_signer_signer_proto_depIdxs = []int32{
	5, // 0: tricorn.SignRequest.network_id:type_name -> tricorn.NetworkType
	0, // 1: tricorn.SignRequest.data_type:type_name -> tricorn.DataType
	5, // 2: tricorn.Signature.network_id:type_name -> tricorn.NetworkType
	5, // 3: tricorn.PublicKeyRequest.network_id:type_name -> tricorn.NetworkType
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_signer_signer_proto_init() }
func file_signer_signer_proto_init() {
	if File_signer_signer_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_signer_signer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignRequest); i {
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
		file_signer_signer_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Signature); i {
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
		file_signer_signer_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PublicKeyRequest); i {
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
		file_signer_signer_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PublicKeyResponse); i {
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
			RawDescriptor: file_signer_signer_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_signer_signer_proto_goTypes,
		DependencyIndexes: file_signer_signer_proto_depIdxs,
		EnumInfos:         file_signer_signer_proto_enumTypes,
		MessageInfos:      file_signer_signer_proto_msgTypes,
	}.Build()
	File_signer_signer_proto = out.File
	file_signer_signer_proto_rawDesc = nil
	file_signer_signer_proto_goTypes = nil
	file_signer_signer_proto_depIdxs = nil
}
