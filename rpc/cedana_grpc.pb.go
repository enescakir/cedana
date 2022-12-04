// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: rpc/cedana.proto

package cedana_orch

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// CedanaClient is the client API for Cedana service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CedanaClient interface {
	// unary call registering client with server
	RegisterClient(ctx context.Context, in *ClientData, opts ...grpc.CallOption) (*ConfigClient, error)
	// client-server streaming state
	RecordState(ctx context.Context, opts ...grpc.CallOption) (Cedana_RecordStateClient, error)
	// get commands from orchestration server
	PollForCommand(ctx context.Context, opts ...grpc.CallOption) (Cedana_PollForCommandClient, error)
}

type cedanaClient struct {
	cc grpc.ClientConnInterface
}

func NewCedanaClient(cc grpc.ClientConnInterface) CedanaClient {
	return &cedanaClient{cc}
}

func (c *cedanaClient) RegisterClient(ctx context.Context, in *ClientData, opts ...grpc.CallOption) (*ConfigClient, error) {
	out := new(ConfigClient)
	err := c.cc.Invoke(ctx, "/cedana.Cedana/RegisterClient", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cedanaClient) RecordState(ctx context.Context, opts ...grpc.CallOption) (Cedana_RecordStateClient, error) {
	stream, err := c.cc.NewStream(ctx, &Cedana_ServiceDesc.Streams[0], "/cedana.Cedana/RecordState", opts...)
	if err != nil {
		return nil, err
	}
	x := &cedanaRecordStateClient{stream}
	return x, nil
}

type Cedana_RecordStateClient interface {
	Send(*ClientData) error
	CloseAndRecv() (*ClientStateAck, error)
	grpc.ClientStream
}

type cedanaRecordStateClient struct {
	grpc.ClientStream
}

func (x *cedanaRecordStateClient) Send(m *ClientData) error {
	return x.ClientStream.SendMsg(m)
}

func (x *cedanaRecordStateClient) CloseAndRecv() (*ClientStateAck, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(ClientStateAck)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *cedanaClient) PollForCommand(ctx context.Context, opts ...grpc.CallOption) (Cedana_PollForCommandClient, error) {
	stream, err := c.cc.NewStream(ctx, &Cedana_ServiceDesc.Streams[1], "/cedana.Cedana/PollForCommand", opts...)
	if err != nil {
		return nil, err
	}
	x := &cedanaPollForCommandClient{stream}
	return x, nil
}

type Cedana_PollForCommandClient interface {
	Send(*ClientData) error
	Recv() (*ClientCommand, error)
	grpc.ClientStream
}

type cedanaPollForCommandClient struct {
	grpc.ClientStream
}

func (x *cedanaPollForCommandClient) Send(m *ClientData) error {
	return x.ClientStream.SendMsg(m)
}

func (x *cedanaPollForCommandClient) Recv() (*ClientCommand, error) {
	m := new(ClientCommand)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CedanaServer is the server API for Cedana service.
// All implementations must embed UnimplementedCedanaServer
// for forward compatibility
type CedanaServer interface {
	// unary call registering client with server
	RegisterClient(context.Context, *ClientData) (*ConfigClient, error)
	// client-server streaming state
	RecordState(Cedana_RecordStateServer) error
	// get commands from orchestration server
	PollForCommand(Cedana_PollForCommandServer) error
	mustEmbedUnimplementedCedanaServer()
}

// UnimplementedCedanaServer must be embedded to have forward compatible implementations.
type UnimplementedCedanaServer struct {
}

func (UnimplementedCedanaServer) RegisterClient(context.Context, *ClientData) (*ConfigClient, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterClient not implemented")
}
func (UnimplementedCedanaServer) RecordState(Cedana_RecordStateServer) error {
	return status.Errorf(codes.Unimplemented, "method RecordState not implemented")
}
func (UnimplementedCedanaServer) PollForCommand(Cedana_PollForCommandServer) error {
	return status.Errorf(codes.Unimplemented, "method PollForCommand not implemented")
}
func (UnimplementedCedanaServer) mustEmbedUnimplementedCedanaServer() {}

// UnsafeCedanaServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CedanaServer will
// result in compilation errors.
type UnsafeCedanaServer interface {
	mustEmbedUnimplementedCedanaServer()
}

func RegisterCedanaServer(s grpc.ServiceRegistrar, srv CedanaServer) {
	s.RegisterService(&Cedana_ServiceDesc, srv)
}

func _Cedana_RegisterClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CedanaServer).RegisterClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cedana.Cedana/RegisterClient",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CedanaServer).RegisterClient(ctx, req.(*ClientData))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cedana_RecordState_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CedanaServer).RecordState(&cedanaRecordStateServer{stream})
}

type Cedana_RecordStateServer interface {
	SendAndClose(*ClientStateAck) error
	Recv() (*ClientData, error)
	grpc.ServerStream
}

type cedanaRecordStateServer struct {
	grpc.ServerStream
}

func (x *cedanaRecordStateServer) SendAndClose(m *ClientStateAck) error {
	return x.ServerStream.SendMsg(m)
}

func (x *cedanaRecordStateServer) Recv() (*ClientData, error) {
	m := new(ClientData)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Cedana_PollForCommand_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CedanaServer).PollForCommand(&cedanaPollForCommandServer{stream})
}

type Cedana_PollForCommandServer interface {
	Send(*ClientCommand) error
	Recv() (*ClientData, error)
	grpc.ServerStream
}

type cedanaPollForCommandServer struct {
	grpc.ServerStream
}

func (x *cedanaPollForCommandServer) Send(m *ClientCommand) error {
	return x.ServerStream.SendMsg(m)
}

func (x *cedanaPollForCommandServer) Recv() (*ClientData, error) {
	m := new(ClientData)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Cedana_ServiceDesc is the grpc.ServiceDesc for Cedana service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Cedana_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "cedana.Cedana",
	HandlerType: (*CedanaServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterClient",
			Handler:    _Cedana_RegisterClient_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "RecordState",
			Handler:       _Cedana_RecordState_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "PollForCommand",
			Handler:       _Cedana_PollForCommand_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "rpc/cedana.proto",
}
