// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.1
// source: block.proto

package rpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Block_GenesisSync_FullMethodName  = "/Block/GenesisSync"
	Block_BlockSync_FullMethodName    = "/Block/BlockSync"
	Block_BlockReceive_FullMethodName = "/Block/BlockReceive"
	Block_BlockSearch_FullMethodName  = "/Block/BlockSearch"
)

// BlockClient is the client API for Block service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BlockClient interface {
	GenesisSync(ctx context.Context, in *GenesisSyncReq, opts ...grpc.CallOption) (*GenesisSyncRes, error)
	BlockSync(ctx context.Context, in *BlockSyncReq, opts ...grpc.CallOption) (grpc.ServerStreamingClient[BlockSyncRes], error)
	BlockReceive(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[BlockReceiveReq, BlockReceiveRes], error)
	BlockSearch(ctx context.Context, in *BlockSearchReq, opts ...grpc.CallOption) (grpc.ServerStreamingClient[BlockSearchRes], error)
}

type blockClient struct {
	cc grpc.ClientConnInterface
}

func NewBlockClient(cc grpc.ClientConnInterface) BlockClient {
	return &blockClient{cc}
}

func (c *blockClient) GenesisSync(ctx context.Context, in *GenesisSyncReq, opts ...grpc.CallOption) (*GenesisSyncRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GenesisSyncRes)
	err := c.cc.Invoke(ctx, Block_GenesisSync_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockClient) BlockSync(ctx context.Context, in *BlockSyncReq, opts ...grpc.CallOption) (grpc.ServerStreamingClient[BlockSyncRes], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Block_ServiceDesc.Streams[0], Block_BlockSync_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[BlockSyncReq, BlockSyncRes]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Block_BlockSyncClient = grpc.ServerStreamingClient[BlockSyncRes]

func (c *blockClient) BlockReceive(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[BlockReceiveReq, BlockReceiveRes], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Block_ServiceDesc.Streams[1], Block_BlockReceive_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[BlockReceiveReq, BlockReceiveRes]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Block_BlockReceiveClient = grpc.ClientStreamingClient[BlockReceiveReq, BlockReceiveRes]

func (c *blockClient) BlockSearch(ctx context.Context, in *BlockSearchReq, opts ...grpc.CallOption) (grpc.ServerStreamingClient[BlockSearchRes], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Block_ServiceDesc.Streams[2], Block_BlockSearch_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[BlockSearchReq, BlockSearchRes]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Block_BlockSearchClient = grpc.ServerStreamingClient[BlockSearchRes]

// BlockServer is the server API for Block service.
// All implementations must embed UnimplementedBlockServer
// for forward compatibility.
type BlockServer interface {
	GenesisSync(context.Context, *GenesisSyncReq) (*GenesisSyncRes, error)
	BlockSync(*BlockSyncReq, grpc.ServerStreamingServer[BlockSyncRes]) error
	BlockReceive(grpc.ClientStreamingServer[BlockReceiveReq, BlockReceiveRes]) error
	BlockSearch(*BlockSearchReq, grpc.ServerStreamingServer[BlockSearchRes]) error
	mustEmbedUnimplementedBlockServer()
}

// UnimplementedBlockServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedBlockServer struct{}

func (UnimplementedBlockServer) GenesisSync(context.Context, *GenesisSyncReq) (*GenesisSyncRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenesisSync not implemented")
}
func (UnimplementedBlockServer) BlockSync(*BlockSyncReq, grpc.ServerStreamingServer[BlockSyncRes]) error {
	return status.Errorf(codes.Unimplemented, "method BlockSync not implemented")
}
func (UnimplementedBlockServer) BlockReceive(grpc.ClientStreamingServer[BlockReceiveReq, BlockReceiveRes]) error {
	return status.Errorf(codes.Unimplemented, "method BlockReceive not implemented")
}
func (UnimplementedBlockServer) BlockSearch(*BlockSearchReq, grpc.ServerStreamingServer[BlockSearchRes]) error {
	return status.Errorf(codes.Unimplemented, "method BlockSearch not implemented")
}
func (UnimplementedBlockServer) mustEmbedUnimplementedBlockServer() {}
func (UnimplementedBlockServer) testEmbeddedByValue()               {}

// UnsafeBlockServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BlockServer will
// result in compilation errors.
type UnsafeBlockServer interface {
	mustEmbedUnimplementedBlockServer()
}

func RegisterBlockServer(s grpc.ServiceRegistrar, srv BlockServer) {
	// If the following call pancis, it indicates UnimplementedBlockServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Block_ServiceDesc, srv)
}

func _Block_GenesisSync_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenesisSyncReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockServer).GenesisSync(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Block_GenesisSync_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockServer).GenesisSync(ctx, req.(*GenesisSyncReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Block_BlockSync_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BlockSyncReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BlockServer).BlockSync(m, &grpc.GenericServerStream[BlockSyncReq, BlockSyncRes]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Block_BlockSyncServer = grpc.ServerStreamingServer[BlockSyncRes]

func _Block_BlockReceive_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BlockServer).BlockReceive(&grpc.GenericServerStream[BlockReceiveReq, BlockReceiveRes]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Block_BlockReceiveServer = grpc.ClientStreamingServer[BlockReceiveReq, BlockReceiveRes]

func _Block_BlockSearch_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BlockSearchReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BlockServer).BlockSearch(m, &grpc.GenericServerStream[BlockSearchReq, BlockSearchRes]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Block_BlockSearchServer = grpc.ServerStreamingServer[BlockSearchRes]

// Block_ServiceDesc is the grpc.ServiceDesc for Block service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Block_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Block",
	HandlerType: (*BlockServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GenesisSync",
			Handler:    _Block_GenesisSync_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "BlockSync",
			Handler:       _Block_BlockSync_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "BlockReceive",
			Handler:       _Block_BlockReceive_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "BlockSearch",
			Handler:       _Block_BlockSearch_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "block.proto",
}