// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: server/server.proto

package checkpoint

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

// DumpServiceClient is the client API for DumpService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DumpServiceClient interface {
	Dump(ctx context.Context, in *DumpArgs, opts ...grpc.CallOption) (*DumpResp, error)
}

type dumpServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDumpServiceClient(cc grpc.ClientConnInterface) DumpServiceClient {
	return &dumpServiceClient{cc}
}

func (c *dumpServiceClient) Dump(ctx context.Context, in *DumpArgs, opts ...grpc.CallOption) (*DumpResp, error) {
	out := new(DumpResp)
	err := c.cc.Invoke(ctx, "/DumpService/Dump", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DumpServiceServer is the server API for DumpService service.
// All implementations must embed UnimplementedDumpServiceServer
// for forward compatibility
type DumpServiceServer interface {
	Dump(context.Context, *DumpArgs) (*DumpResp, error)
	mustEmbedUnimplementedDumpServiceServer()
}

// UnimplementedDumpServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDumpServiceServer struct {
}

func (UnimplementedDumpServiceServer) Dump(context.Context, *DumpArgs) (*DumpResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Dump not implemented")
}
func (UnimplementedDumpServiceServer) mustEmbedUnimplementedDumpServiceServer() {}

// UnsafeDumpServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DumpServiceServer will
// result in compilation errors.
type UnsafeDumpServiceServer interface {
	mustEmbedUnimplementedDumpServiceServer()
}

func RegisterDumpServiceServer(s grpc.ServiceRegistrar, srv DumpServiceServer) {
	s.RegisterService(&DumpService_ServiceDesc, srv)
}

func _DumpService_Dump_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DumpArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DumpServiceServer).Dump(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DumpService/Dump",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DumpServiceServer).Dump(ctx, req.(*DumpArgs))
	}
	return interceptor(ctx, in, info, handler)
}

// DumpService_ServiceDesc is the grpc.ServiceDesc for DumpService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DumpService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DumpService",
	HandlerType: (*DumpServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Dump",
			Handler:    _DumpService_Dump_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server/server.proto",
}
