// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.6.1
// source: grpc/raft_service.proto

package service

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

type LogEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EntryTerm int64  `protobuf:"varint,1,opt,name=entryTerm,proto3" json:"entryTerm,omitempty"`
	Payload   string `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *LogEntry) Reset() {
	*x = LogEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_raft_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogEntry) ProtoMessage() {}

func (x *LogEntry) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_raft_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogEntry.ProtoReflect.Descriptor instead.
func (*LogEntry) Descriptor() ([]byte, []int) {
	return file_grpc_raft_service_proto_rawDescGZIP(), []int{0}
}

func (x *LogEntry) GetEntryTerm() int64 {
	if x != nil {
		return x.EntryTerm
	}
	return 0
}

func (x *LogEntry) GetPayload() string {
	if x != nil {
		return x.Payload
	}
	return ""
}

type AppendEntryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServerTerm     int64       `protobuf:"varint,1,opt,name=serverTerm,proto3" json:"serverTerm,omitempty"`
	ServerID       int64       `protobuf:"varint,2,opt,name=serverID,proto3" json:"serverID,omitempty"`
	PrevEntryTerm  int64       `protobuf:"varint,3,opt,name=prevEntryTerm,proto3" json:"prevEntryTerm,omitempty"`
	PrevEntryIndex int64       `protobuf:"varint,4,opt,name=prevEntryIndex,proto3" json:"prevEntryIndex,omitempty"`
	Entries        []*LogEntry `protobuf:"bytes,5,rep,name=entries,proto3" json:"entries,omitempty"`
}

func (x *AppendEntryRequest) Reset() {
	*x = AppendEntryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_raft_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AppendEntryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppendEntryRequest) ProtoMessage() {}

func (x *AppendEntryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_raft_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppendEntryRequest.ProtoReflect.Descriptor instead.
func (*AppendEntryRequest) Descriptor() ([]byte, []int) {
	return file_grpc_raft_service_proto_rawDescGZIP(), []int{1}
}

func (x *AppendEntryRequest) GetServerTerm() int64 {
	if x != nil {
		return x.ServerTerm
	}
	return 0
}

func (x *AppendEntryRequest) GetServerID() int64 {
	if x != nil {
		return x.ServerID
	}
	return 0
}

func (x *AppendEntryRequest) GetPrevEntryTerm() int64 {
	if x != nil {
		return x.PrevEntryTerm
	}
	return 0
}

func (x *AppendEntryRequest) GetPrevEntryIndex() int64 {
	if x != nil {
		return x.PrevEntryIndex
	}
	return 0
}

func (x *AppendEntryRequest) GetEntries() []*LogEntry {
	if x != nil {
		return x.Entries
	}
	return nil
}

type AppendEntryReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CurrentTerm int64 `protobuf:"varint,1,opt,name=currentTerm,proto3" json:"currentTerm,omitempty"`
	Success     bool  `protobuf:"varint,2,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *AppendEntryReply) Reset() {
	*x = AppendEntryReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_raft_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AppendEntryReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AppendEntryReply) ProtoMessage() {}

func (x *AppendEntryReply) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_raft_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AppendEntryReply.ProtoReflect.Descriptor instead.
func (*AppendEntryReply) Descriptor() ([]byte, []int) {
	return file_grpc_raft_service_proto_rawDescGZIP(), []int{2}
}

func (x *AppendEntryReply) GetCurrentTerm() int64 {
	if x != nil {
		return x.CurrentTerm
	}
	return 0
}

func (x *AppendEntryReply) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type RequestVoteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServerTerm int64 `protobuf:"varint,1,opt,name=serverTerm,proto3" json:"serverTerm,omitempty"`
	ServerID   int64 `protobuf:"varint,2,opt,name=serverID,proto3" json:"serverID,omitempty"`
}

func (x *RequestVoteRequest) Reset() {
	*x = RequestVoteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_raft_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestVoteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestVoteRequest) ProtoMessage() {}

func (x *RequestVoteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_raft_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestVoteRequest.ProtoReflect.Descriptor instead.
func (*RequestVoteRequest) Descriptor() ([]byte, []int) {
	return file_grpc_raft_service_proto_rawDescGZIP(), []int{3}
}

func (x *RequestVoteRequest) GetServerTerm() int64 {
	if x != nil {
		return x.ServerTerm
	}
	return 0
}

func (x *RequestVoteRequest) GetServerID() int64 {
	if x != nil {
		return x.ServerID
	}
	return 0
}

type RequestVoteReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success    bool  `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	ServerTerm int64 `protobuf:"varint,2,opt,name=serverTerm,proto3" json:"serverTerm,omitempty"`
}

func (x *RequestVoteReply) Reset() {
	*x = RequestVoteReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_raft_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestVoteReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestVoteReply) ProtoMessage() {}

func (x *RequestVoteReply) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_raft_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestVoteReply.ProtoReflect.Descriptor instead.
func (*RequestVoteReply) Descriptor() ([]byte, []int) {
	return file_grpc_raft_service_proto_rawDescGZIP(), []int{4}
}

func (x *RequestVoteReply) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *RequestVoteReply) GetServerTerm() int64 {
	if x != nil {
		return x.ServerTerm
	}
	return 0
}

var File_grpc_raft_service_proto protoreflect.FileDescriptor

var file_grpc_raft_service_proto_rawDesc = []byte{
	0x0a, 0x17, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x72, 0x61, 0x66, 0x74, 0x5f, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x42, 0x0a, 0x08, 0x4c, 0x6f, 0x67,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x1c, 0x0a, 0x09, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x54, 0x65,
	0x72, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x54,
	0x65, 0x72, 0x6d, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0xc3, 0x01,
	0x0a, 0x12, 0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x54, 0x65,
	0x72, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x54, 0x65, 0x72, 0x6d, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x44,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x44,
	0x12, 0x24, 0x0a, 0x0d, 0x70, 0x72, 0x65, 0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x54, 0x65, 0x72,
	0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x70, 0x72, 0x65, 0x76, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x54, 0x65, 0x72, 0x6d, 0x12, 0x26, 0x0a, 0x0e, 0x70, 0x72, 0x65, 0x76, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0e,
	0x70, 0x72, 0x65, 0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x23,
	0x0a, 0x07, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x09, 0x2e, 0x4c, 0x6f, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x65, 0x6e, 0x74, 0x72,
	0x69, 0x65, 0x73, 0x22, 0x4e, 0x0a, 0x10, 0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x20, 0x0a, 0x0b, 0x63, 0x75, 0x72, 0x72, 0x65,
	0x6e, 0x74, 0x54, 0x65, 0x72, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x63, 0x75,
	0x72, 0x72, 0x65, 0x6e, 0x74, 0x54, 0x65, 0x72, 0x6d, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x22, 0x50, 0x0a, 0x12, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x56, 0x6f,
	0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x54, 0x65, 0x72, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x54, 0x65, 0x72, 0x6d, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x49, 0x44, 0x22, 0x4c, 0x0a, 0x10, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x56, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x54, 0x65, 0x72,
	0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x54,
	0x65, 0x72, 0x6d, 0x32, 0x78, 0x0a, 0x04, 0x52, 0x61, 0x66, 0x74, 0x12, 0x37, 0x0a, 0x0b, 0x41,
	0x70, 0x70, 0x65, 0x6e, 0x64, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x13, 0x2e, 0x41, 0x70, 0x70,
	0x65, 0x6e, 0x64, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x11, 0x2e, 0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x22, 0x00, 0x12, 0x37, 0x0a, 0x0b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x56,
	0x6f, 0x74, 0x65, 0x12, 0x13, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x56, 0x6f, 0x74,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x11, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x56, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x3b, 0x5a,
	0x39, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x69, 0x75, 0x6c,
	0x69, 0x6f, 0x62, 0x6f, 0x72, 0x67, 0x68, 0x65, 0x73, 0x69, 0x2f, 0x72, 0x61, 0x66, 0x74, 0x2d,
	0x69, 0x6d, 0x70, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x73,
	0x72, 0x63, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_grpc_raft_service_proto_rawDescOnce sync.Once
	file_grpc_raft_service_proto_rawDescData = file_grpc_raft_service_proto_rawDesc
)

func file_grpc_raft_service_proto_rawDescGZIP() []byte {
	file_grpc_raft_service_proto_rawDescOnce.Do(func() {
		file_grpc_raft_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_grpc_raft_service_proto_rawDescData)
	})
	return file_grpc_raft_service_proto_rawDescData
}

var file_grpc_raft_service_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_grpc_raft_service_proto_goTypes = []interface{}{
	(*LogEntry)(nil),           // 0: LogEntry
	(*AppendEntryRequest)(nil), // 1: AppendEntryRequest
	(*AppendEntryReply)(nil),   // 2: AppendEntryReply
	(*RequestVoteRequest)(nil), // 3: RequestVoteRequest
	(*RequestVoteReply)(nil),   // 4: RequestVoteReply
}
var file_grpc_raft_service_proto_depIdxs = []int32{
	0, // 0: AppendEntryRequest.entries:type_name -> LogEntry
	1, // 1: Raft.AppendEntry:input_type -> AppendEntryRequest
	3, // 2: Raft.RequestVote:input_type -> RequestVoteRequest
	2, // 3: Raft.AppendEntry:output_type -> AppendEntryReply
	4, // 4: Raft.RequestVote:output_type -> RequestVoteReply
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_grpc_raft_service_proto_init() }
func file_grpc_raft_service_proto_init() {
	if File_grpc_raft_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_grpc_raft_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogEntry); i {
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
		file_grpc_raft_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AppendEntryRequest); i {
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
		file_grpc_raft_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AppendEntryReply); i {
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
		file_grpc_raft_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequestVoteRequest); i {
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
		file_grpc_raft_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequestVoteReply); i {
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
			RawDescriptor: file_grpc_raft_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_grpc_raft_service_proto_goTypes,
		DependencyIndexes: file_grpc_raft_service_proto_depIdxs,
		MessageInfos:      file_grpc_raft_service_proto_msgTypes,
	}.Build()
	File_grpc_raft_service_proto = out.File
	file_grpc_raft_service_proto_rawDesc = nil
	file_grpc_raft_service_proto_goTypes = nil
	file_grpc_raft_service_proto_depIdxs = nil
}
