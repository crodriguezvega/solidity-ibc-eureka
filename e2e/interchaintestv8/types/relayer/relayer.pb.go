// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        (unknown)
// source: relayer/relayer.proto

package relayer

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// The request message
type RelayByTxRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The source chain identifier
	SrcChain string `protobuf:"bytes,1,opt,name=src_chain,json=srcChain,proto3" json:"src_chain,omitempty"`
	// The target chain identifier
	DstChain string `protobuf:"bytes,2,opt,name=dst_chain,json=dstChain,proto3" json:"dst_chain,omitempty"`
	// The identifiers for the IBC transactions to be relayed
	// This is usually the transaction hash
	SourceTxIds [][]byte `protobuf:"bytes,3,rep,name=source_tx_ids,json=sourceTxIds,proto3" json:"source_tx_ids,omitempty"`
	// The identifiers for the IBC transactions on the target chain to be timed out
	TimeoutTxIds [][]byte `protobuf:"bytes,4,rep,name=timeout_tx_ids,json=timeoutTxIds,proto3" json:"timeout_tx_ids,omitempty"`
	// The identifier for the target client
	// Used for event filtering
	TargetClientId string `protobuf:"bytes,5,opt,name=target_client_id,json=targetClientId,proto3" json:"target_client_id,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *RelayByTxRequest) Reset() {
	*x = RelayByTxRequest{}
	mi := &file_relayer_relayer_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RelayByTxRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RelayByTxRequest) ProtoMessage() {}

func (x *RelayByTxRequest) ProtoReflect() protoreflect.Message {
	mi := &file_relayer_relayer_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RelayByTxRequest.ProtoReflect.Descriptor instead.
func (*RelayByTxRequest) Descriptor() ([]byte, []int) {
	return file_relayer_relayer_proto_rawDescGZIP(), []int{0}
}

func (x *RelayByTxRequest) GetSrcChain() string {
	if x != nil {
		return x.SrcChain
	}
	return ""
}

func (x *RelayByTxRequest) GetDstChain() string {
	if x != nil {
		return x.DstChain
	}
	return ""
}

func (x *RelayByTxRequest) GetSourceTxIds() [][]byte {
	if x != nil {
		return x.SourceTxIds
	}
	return nil
}

func (x *RelayByTxRequest) GetTimeoutTxIds() [][]byte {
	if x != nil {
		return x.TimeoutTxIds
	}
	return nil
}

func (x *RelayByTxRequest) GetTargetClientId() string {
	if x != nil {
		return x.TargetClientId
	}
	return ""
}

// The response message
type RelayByTxResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The multicall transaction to be submitted by caller
	Tx []byte `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
	// The contract address to submit the transaction, if applicable
	Address       string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RelayByTxResponse) Reset() {
	*x = RelayByTxResponse{}
	mi := &file_relayer_relayer_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RelayByTxResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RelayByTxResponse) ProtoMessage() {}

func (x *RelayByTxResponse) ProtoReflect() protoreflect.Message {
	mi := &file_relayer_relayer_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RelayByTxResponse.ProtoReflect.Descriptor instead.
func (*RelayByTxResponse) Descriptor() ([]byte, []int) {
	return file_relayer_relayer_proto_rawDescGZIP(), []int{1}
}

func (x *RelayByTxResponse) GetTx() []byte {
	if x != nil {
		return x.Tx
	}
	return nil
}

func (x *RelayByTxResponse) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

// Information request message
type InfoRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The source chain identifier
	SrcChain string `protobuf:"bytes,1,opt,name=src_chain,json=srcChain,proto3" json:"src_chain,omitempty"`
	// The target chain identifier
	DstChain      string `protobuf:"bytes,2,opt,name=dst_chain,json=dstChain,proto3" json:"dst_chain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InfoRequest) Reset() {
	*x = InfoRequest{}
	mi := &file_relayer_relayer_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InfoRequest) ProtoMessage() {}

func (x *InfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_relayer_relayer_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InfoRequest.ProtoReflect.Descriptor instead.
func (*InfoRequest) Descriptor() ([]byte, []int) {
	return file_relayer_relayer_proto_rawDescGZIP(), []int{2}
}

func (x *InfoRequest) GetSrcChain() string {
	if x != nil {
		return x.SrcChain
	}
	return ""
}

func (x *InfoRequest) GetDstChain() string {
	if x != nil {
		return x.DstChain
	}
	return ""
}

// Information response message
type InfoResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The target chain information
	TargetChain *Chain `protobuf:"bytes,1,opt,name=target_chain,json=targetChain,proto3" json:"target_chain,omitempty"`
	// The source chain information
	SourceChain   *Chain `protobuf:"bytes,2,opt,name=source_chain,json=sourceChain,proto3" json:"source_chain,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InfoResponse) Reset() {
	*x = InfoResponse{}
	mi := &file_relayer_relayer_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InfoResponse) ProtoMessage() {}

func (x *InfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_relayer_relayer_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InfoResponse.ProtoReflect.Descriptor instead.
func (*InfoResponse) Descriptor() ([]byte, []int) {
	return file_relayer_relayer_proto_rawDescGZIP(), []int{3}
}

func (x *InfoResponse) GetTargetChain() *Chain {
	if x != nil {
		return x.TargetChain
	}
	return nil
}

func (x *InfoResponse) GetSourceChain() *Chain {
	if x != nil {
		return x.SourceChain
	}
	return nil
}

// The chain definition
type Chain struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The chain id
	ChainId string `protobuf:"bytes,1,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	// The ibc version
	IbcVersion string `protobuf:"bytes,2,opt,name=ibc_version,json=ibcVersion,proto3" json:"ibc_version,omitempty"`
	// The ibc contract address
	IbcContract   string `protobuf:"bytes,3,opt,name=ibc_contract,json=ibcContract,proto3" json:"ibc_contract,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Chain) Reset() {
	*x = Chain{}
	mi := &file_relayer_relayer_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Chain) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Chain) ProtoMessage() {}

func (x *Chain) ProtoReflect() protoreflect.Message {
	mi := &file_relayer_relayer_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Chain.ProtoReflect.Descriptor instead.
func (*Chain) Descriptor() ([]byte, []int) {
	return file_relayer_relayer_proto_rawDescGZIP(), []int{4}
}

func (x *Chain) GetChainId() string {
	if x != nil {
		return x.ChainId
	}
	return ""
}

func (x *Chain) GetIbcVersion() string {
	if x != nil {
		return x.IbcVersion
	}
	return ""
}

func (x *Chain) GetIbcContract() string {
	if x != nil {
		return x.IbcContract
	}
	return ""
}

var File_relayer_relayer_proto protoreflect.FileDescriptor

var file_relayer_relayer_proto_rawDesc = string([]byte{
	0x0a, 0x15, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x2f, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x65,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72,
	0x22, 0xc0, 0x01, 0x0a, 0x10, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x42, 0x79, 0x54, 0x78, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x72, 0x63, 0x5f, 0x63, 0x68, 0x61,
	0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x72, 0x63, 0x43, 0x68, 0x61,
	0x69, 0x6e, 0x12, 0x1b, 0x0a, 0x09, 0x64, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x64, 0x73, 0x74, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x12,
	0x22, 0x0a, 0x0d, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x74, 0x78, 0x5f, 0x69, 0x64, 0x73,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0b, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x78,
	0x49, 0x64, 0x73, 0x12, 0x24, 0x0a, 0x0e, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x74,
	0x78, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0c, 0x74, 0x69, 0x6d,
	0x65, 0x6f, 0x75, 0x74, 0x54, 0x78, 0x49, 0x64, 0x73, 0x12, 0x28, 0x0a, 0x10, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x5f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0e, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x49, 0x64, 0x22, 0x3d, 0x0a, 0x11, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x42, 0x79, 0x54, 0x78,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x78, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x74, 0x78, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x22, 0x47, 0x0a, 0x0b, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x72, 0x63, 0x5f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x72, 0x63, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x12, 0x1b,
	0x0a, 0x09, 0x64, 0x73, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x64, 0x73, 0x74, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x22, 0x74, 0x0a, 0x0c, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x31, 0x0a, 0x0c, 0x74,
	0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0e, 0x2e, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x61, 0x69,
	0x6e, 0x52, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x12, 0x31,
	0x0a, 0x0c, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x2e, 0x43,
	0x68, 0x61, 0x69, 0x6e, 0x52, 0x0b, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x43, 0x68, 0x61, 0x69,
	0x6e, 0x22, 0x66, 0x0a, 0x05, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x69, 0x62, 0x63, 0x5f, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x69, 0x62, 0x63, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x21, 0x0a, 0x0c, 0x69, 0x62, 0x63, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x69, 0x62,
	0x63, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x32, 0x89, 0x01, 0x0a, 0x0e, 0x52, 0x65,
	0x6c, 0x61, 0x79, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x42, 0x0a, 0x09,
	0x52, 0x65, 0x6c, 0x61, 0x79, 0x42, 0x79, 0x54, 0x78, 0x12, 0x19, 0x2e, 0x72, 0x65, 0x6c, 0x61,
	0x79, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x42, 0x79, 0x54, 0x78, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x2e, 0x52,
	0x65, 0x6c, 0x61, 0x79, 0x42, 0x79, 0x54, 0x78, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x33, 0x0a, 0x04, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x14, 0x2e, 0x72, 0x65, 0x6c, 0x61, 0x79,
	0x65, 0x72, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15,
	0x2e, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x66, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x2e, 0x72, 0x65, 0x6c,
	0x61, 0x79, 0x65, 0x72, 0x42, 0x0c, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x0d, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x72, 0x65, 0x6c, 0x61,
	0x79, 0x65, 0x72, 0xa2, 0x02, 0x03, 0x52, 0x58, 0x58, 0xaa, 0x02, 0x07, 0x52, 0x65, 0x6c, 0x61,
	0x79, 0x65, 0x72, 0xca, 0x02, 0x07, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0xe2, 0x02, 0x13,
	0x52, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0xea, 0x02, 0x07, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_relayer_relayer_proto_rawDescOnce sync.Once
	file_relayer_relayer_proto_rawDescData []byte
)

func file_relayer_relayer_proto_rawDescGZIP() []byte {
	file_relayer_relayer_proto_rawDescOnce.Do(func() {
		file_relayer_relayer_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_relayer_relayer_proto_rawDesc), len(file_relayer_relayer_proto_rawDesc)))
	})
	return file_relayer_relayer_proto_rawDescData
}

var file_relayer_relayer_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_relayer_relayer_proto_goTypes = []any{
	(*RelayByTxRequest)(nil),  // 0: relayer.RelayByTxRequest
	(*RelayByTxResponse)(nil), // 1: relayer.RelayByTxResponse
	(*InfoRequest)(nil),       // 2: relayer.InfoRequest
	(*InfoResponse)(nil),      // 3: relayer.InfoResponse
	(*Chain)(nil),             // 4: relayer.Chain
}
var file_relayer_relayer_proto_depIdxs = []int32{
	4, // 0: relayer.InfoResponse.target_chain:type_name -> relayer.Chain
	4, // 1: relayer.InfoResponse.source_chain:type_name -> relayer.Chain
	0, // 2: relayer.RelayerService.RelayByTx:input_type -> relayer.RelayByTxRequest
	2, // 3: relayer.RelayerService.Info:input_type -> relayer.InfoRequest
	1, // 4: relayer.RelayerService.RelayByTx:output_type -> relayer.RelayByTxResponse
	3, // 5: relayer.RelayerService.Info:output_type -> relayer.InfoResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_relayer_relayer_proto_init() }
func file_relayer_relayer_proto_init() {
	if File_relayer_relayer_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_relayer_relayer_proto_rawDesc), len(file_relayer_relayer_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_relayer_relayer_proto_goTypes,
		DependencyIndexes: file_relayer_relayer_proto_depIdxs,
		MessageInfos:      file_relayer_relayer_proto_msgTypes,
	}.Build()
	File_relayer_relayer_proto = out.File
	file_relayer_relayer_proto_goTypes = nil
	file_relayer_relayer_proto_depIdxs = nil
}
