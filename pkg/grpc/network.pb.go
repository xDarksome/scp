// Code generated by protoc-gen-go. DO NOT EDIT.
// source: network.proto

package grpc

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Message_Type int32

const (
	Message_UNKNOWN         Message_Type = 0
	Message_VOTE_NOMINATE   Message_Type = 1
	Message_ACCEPT_NOMINATE Message_Type = 2
	Message_VOTE_PREPARE    Message_Type = 3
	Message_ACCEPT_PREPARE  Message_Type = 4
	Message_VOTE_COMMIT     Message_Type = 5
	Message_ACCEPT_COMMIT   Message_Type = 6
)

var Message_Type_name = map[int32]string{
	0: "UNKNOWN",
	1: "VOTE_NOMINATE",
	2: "ACCEPT_NOMINATE",
	3: "VOTE_PREPARE",
	4: "ACCEPT_PREPARE",
	5: "VOTE_COMMIT",
	6: "ACCEPT_COMMIT",
}

var Message_Type_value = map[string]int32{
	"UNKNOWN":         0,
	"VOTE_NOMINATE":   1,
	"ACCEPT_NOMINATE": 2,
	"VOTE_PREPARE":    3,
	"ACCEPT_PREPARE":  4,
	"VOTE_COMMIT":     5,
	"ACCEPT_COMMIT":   6,
}

func (x Message_Type) String() string {
	return proto.EnumName(Message_Type_name, int32(x))
}

func (Message_Type) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_8571034d60397816, []int{2, 0}
}

type StreamMessagesRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StreamMessagesRequest) Reset()         { *m = StreamMessagesRequest{} }
func (m *StreamMessagesRequest) String() string { return proto.CompactTextString(m) }
func (*StreamMessagesRequest) ProtoMessage()    {}
func (*StreamMessagesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8571034d60397816, []int{0}
}

func (m *StreamMessagesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StreamMessagesRequest.Unmarshal(m, b)
}
func (m *StreamMessagesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StreamMessagesRequest.Marshal(b, m, deterministic)
}
func (m *StreamMessagesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StreamMessagesRequest.Merge(m, src)
}
func (m *StreamMessagesRequest) XXX_Size() int {
	return xxx_messageInfo_StreamMessagesRequest.Size(m)
}
func (m *StreamMessagesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StreamMessagesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StreamMessagesRequest proto.InternalMessageInfo

type GetSlotRequest struct {
	Index                uint64   `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetSlotRequest) Reset()         { *m = GetSlotRequest{} }
func (m *GetSlotRequest) String() string { return proto.CompactTextString(m) }
func (*GetSlotRequest) ProtoMessage()    {}
func (*GetSlotRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8571034d60397816, []int{1}
}

func (m *GetSlotRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetSlotRequest.Unmarshal(m, b)
}
func (m *GetSlotRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetSlotRequest.Marshal(b, m, deterministic)
}
func (m *GetSlotRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetSlotRequest.Merge(m, src)
}
func (m *GetSlotRequest) XXX_Size() int {
	return xxx_messageInfo_GetSlotRequest.Size(m)
}
func (m *GetSlotRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetSlotRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetSlotRequest proto.InternalMessageInfo

func (m *GetSlotRequest) GetIndex() uint64 {
	if m != nil {
		return m.Index
	}
	return 0
}

type Message struct {
	Type                 Message_Type `protobuf:"varint,1,opt,name=type,proto3,enum=grpc.Message_Type" json:"type,omitempty"`
	SlotIndex            uint64       `protobuf:"varint,2,opt,name=slotIndex,proto3" json:"slotIndex,omitempty"`
	Counter              uint32       `protobuf:"varint,3,opt,name=counter,proto3" json:"counter,omitempty"`
	Value                []byte       `protobuf:"bytes,4,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return fileDescriptor_8571034d60397816, []int{2}
}

func (m *Message) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Message.Unmarshal(m, b)
}
func (m *Message) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Message.Marshal(b, m, deterministic)
}
func (m *Message) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Message.Merge(m, src)
}
func (m *Message) XXX_Size() int {
	return xxx_messageInfo_Message.Size(m)
}
func (m *Message) XXX_DiscardUnknown() {
	xxx_messageInfo_Message.DiscardUnknown(m)
}

var xxx_messageInfo_Message proto.InternalMessageInfo

func (m *Message) GetType() Message_Type {
	if m != nil {
		return m.Type
	}
	return Message_UNKNOWN
}

func (m *Message) GetSlotIndex() uint64 {
	if m != nil {
		return m.SlotIndex
	}
	return 0
}

func (m *Message) GetCounter() uint32 {
	if m != nil {
		return m.Counter
	}
	return 0
}

func (m *Message) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type Ballot struct {
	Counter              uint32   `protobuf:"varint,1,opt,name=counter,proto3" json:"counter,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Ballot) Reset()         { *m = Ballot{} }
func (m *Ballot) String() string { return proto.CompactTextString(m) }
func (*Ballot) ProtoMessage()    {}
func (*Ballot) Descriptor() ([]byte, []int) {
	return fileDescriptor_8571034d60397816, []int{3}
}

func (m *Ballot) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Ballot.Unmarshal(m, b)
}
func (m *Ballot) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Ballot.Marshal(b, m, deterministic)
}
func (m *Ballot) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Ballot.Merge(m, src)
}
func (m *Ballot) XXX_Size() int {
	return xxx_messageInfo_Ballot.Size(m)
}
func (m *Ballot) XXX_DiscardUnknown() {
	xxx_messageInfo_Ballot.DiscardUnknown(m)
}

var xxx_messageInfo_Ballot proto.InternalMessageInfo

func (m *Ballot) GetCounter() uint32 {
	if m != nil {
		return m.Counter
	}
	return 0
}

func (m *Ballot) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type Slot struct {
	Index                uint64   `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Slot) Reset()         { *m = Slot{} }
func (m *Slot) String() string { return proto.CompactTextString(m) }
func (*Slot) ProtoMessage()    {}
func (*Slot) Descriptor() ([]byte, []int) {
	return fileDescriptor_8571034d60397816, []int{4}
}

func (m *Slot) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Slot.Unmarshal(m, b)
}
func (m *Slot) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Slot.Marshal(b, m, deterministic)
}
func (m *Slot) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Slot.Merge(m, src)
}
func (m *Slot) XXX_Size() int {
	return xxx_messageInfo_Slot.Size(m)
}
func (m *Slot) XXX_DiscardUnknown() {
	xxx_messageInfo_Slot.DiscardUnknown(m)
}

var xxx_messageInfo_Slot proto.InternalMessageInfo

func (m *Slot) GetIndex() uint64 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *Slot) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func init() {
	proto.RegisterEnum("grpc.Message_Type", Message_Type_name, Message_Type_value)
	proto.RegisterType((*StreamMessagesRequest)(nil), "grpc.StreamMessagesRequest")
	proto.RegisterType((*GetSlotRequest)(nil), "grpc.GetSlotRequest")
	proto.RegisterType((*Message)(nil), "grpc.Message")
	proto.RegisterType((*Ballot)(nil), "grpc.Ballot")
	proto.RegisterType((*Slot)(nil), "grpc.Slot")
}

func init() { proto.RegisterFile("network.proto", fileDescriptor_8571034d60397816) }

var fileDescriptor_8571034d60397816 = []byte{
	// 347 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x52, 0x51, 0x4b, 0xc2, 0x50,
	0x18, 0xed, 0xce, 0xab, 0xa3, 0x4f, 0x37, 0xd7, 0x97, 0xd1, 0xb0, 0x1e, 0x64, 0x0f, 0x22, 0x04,
	0x23, 0xec, 0xa5, 0xa7, 0xc0, 0x64, 0x84, 0xc4, 0x36, 0x99, 0xab, 0x1e, 0x63, 0xe9, 0x45, 0xa2,
	0xe5, 0x5d, 0xdb, 0xb5, 0xf2, 0x07, 0x44, 0x7f, 0x3b, 0xb6, 0x5d, 0x51, 0xc1, 0x1e, 0xbf, 0xf3,
	0x9d, 0xef, 0x6c, 0xe7, 0x9c, 0x0b, 0xda, 0x82, 0x89, 0x2f, 0x9e, 0xbe, 0xd9, 0x49, 0xca, 0x05,
	0x47, 0x3a, 0x4f, 0x93, 0xa9, 0x75, 0x0a, 0x27, 0x13, 0x91, 0xb2, 0xe8, 0xdd, 0x65, 0x59, 0x16,
	0xcd, 0x59, 0x16, 0xb0, 0x8f, 0x25, 0xcb, 0x84, 0xd5, 0x05, 0xfd, 0x8e, 0x89, 0x49, 0xcc, 0x85,
	0x44, 0xb0, 0x05, 0xd5, 0xd7, 0xc5, 0x8c, 0x7d, 0x9b, 0xa4, 0x43, 0x7a, 0x34, 0x28, 0x07, 0xeb,
	0x57, 0x01, 0x55, 0xde, 0x62, 0x17, 0xa8, 0x58, 0x25, 0xac, 0x20, 0xe8, 0x7d, 0xb4, 0xf3, 0x2f,
	0xd8, 0x72, 0x69, 0x87, 0xab, 0x84, 0x05, 0xc5, 0x1e, 0xcf, 0xe1, 0x30, 0x8b, 0xb9, 0x18, 0x15,
	0x6a, 0x4a, 0xa1, 0xb6, 0x01, 0xd0, 0x04, 0x75, 0xca, 0x97, 0x0b, 0xc1, 0x52, 0xb3, 0xd2, 0x21,
	0x3d, 0x2d, 0x58, 0x8f, 0xf9, 0x1f, 0x7c, 0x46, 0xf1, 0x92, 0x99, 0xb4, 0x43, 0x7a, 0x8d, 0xa0,
	0x1c, 0xac, 0x1f, 0x02, 0x34, 0x17, 0xc7, 0x3a, 0xa8, 0x0f, 0xde, 0xbd, 0xe7, 0x3f, 0x79, 0xc6,
	0x01, 0x1e, 0x81, 0xf6, 0xe8, 0x87, 0xce, 0xb3, 0xe7, 0xbb, 0x23, 0x6f, 0x10, 0x3a, 0x06, 0xc1,
	0x63, 0x68, 0x0e, 0x86, 0x43, 0x67, 0x1c, 0x6e, 0x40, 0x05, 0x0d, 0x68, 0x14, 0xbc, 0x71, 0xe0,
	0x8c, 0x07, 0x81, 0x63, 0x54, 0x10, 0x41, 0x97, 0xb4, 0x35, 0x46, 0xb1, 0x09, 0xf5, 0x82, 0x35,
	0xf4, 0x5d, 0x77, 0x14, 0x1a, 0xd5, 0x5c, 0x5e, 0x92, 0x24, 0x54, 0xb3, 0xae, 0xa1, 0x76, 0x1b,
	0xc5, 0x31, 0x17, 0xdb, 0x0e, 0xc8, 0x3f, 0x0e, 0x94, 0x6d, 0x07, 0x7d, 0xa0, 0x79, 0xd0, 0xfb,
	0x13, 0xde, 0x7f, 0xd3, 0xcf, 0x80, 0x7a, 0x7c, 0xc6, 0xf0, 0x06, 0xf4, 0xdd, 0x02, 0xf1, 0xac,
	0xcc, 0x7d, 0x6f, 0xad, 0x6d, 0x6d, 0xa7, 0x94, 0x4b, 0x82, 0x17, 0xa0, 0xca, 0x9e, 0xb1, 0x55,
	0xee, 0x76, 0x6b, 0x6f, 0x83, 0x94, 0x8b, 0xb9, 0x78, 0xa9, 0x15, 0x4f, 0xe7, 0xea, 0x2f, 0x00,
	0x00, 0xff, 0xff, 0x70, 0x40, 0x45, 0xac, 0x4b, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// NodeClient is the client API for Node service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type NodeClient interface {
	StreamMessages(ctx context.Context, in *StreamMessagesRequest, opts ...grpc.CallOption) (Node_StreamMessagesClient, error)
	GetSlot(ctx context.Context, in *GetSlotRequest, opts ...grpc.CallOption) (*Slot, error)
}

type nodeClient struct {
	cc *grpc.ClientConn
}

func NewNodeClient(cc *grpc.ClientConn) NodeClient {
	return &nodeClient{cc}
}

func (c *nodeClient) StreamMessages(ctx context.Context, in *StreamMessagesRequest, opts ...grpc.CallOption) (Node_StreamMessagesClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Node_serviceDesc.Streams[0], "/grpc.Node/StreamMessages", opts...)
	if err != nil {
		return nil, err
	}
	x := &nodeStreamMessagesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Node_StreamMessagesClient interface {
	Recv() (*Message, error)
	grpc.ClientStream
}

type nodeStreamMessagesClient struct {
	grpc.ClientStream
}

func (x *nodeStreamMessagesClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *nodeClient) GetSlot(ctx context.Context, in *GetSlotRequest, opts ...grpc.CallOption) (*Slot, error) {
	out := new(Slot)
	err := c.cc.Invoke(ctx, "/grpc.Node/GetSlot", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NodeServer is the server API for Node service.
type NodeServer interface {
	StreamMessages(*StreamMessagesRequest, Node_StreamMessagesServer) error
	GetSlot(context.Context, *GetSlotRequest) (*Slot, error)
}

func RegisterNodeServer(s *grpc.Server, srv NodeServer) {
	s.RegisterService(&_Node_serviceDesc, srv)
}

func _Node_StreamMessages_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamMessagesRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(NodeServer).StreamMessages(m, &nodeStreamMessagesServer{stream})
}

type Node_StreamMessagesServer interface {
	Send(*Message) error
	grpc.ServerStream
}

type nodeStreamMessagesServer struct {
	grpc.ServerStream
}

func (x *nodeStreamMessagesServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func _Node_GetSlot_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSlotRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeServer).GetSlot(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.Node/GetSlot",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeServer).GetSlot(ctx, req.(*GetSlotRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Node_serviceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.Node",
	HandlerType: (*NodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSlot",
			Handler:    _Node_GetSlot_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamMessages",
			Handler:       _Node_StreamMessages_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "network.proto",
}
