// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.1
// source: tx.proto

package rtx

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
	Tx_TxSign_FullMethodName    = "/Tx/TxSign"
	Tx_TxSend_FullMethodName    = "/Tx/TxSend"
	Tx_TxReceive_FullMethodName = "/Tx/TxReceive"
)

// TxClient is the client API for Tx service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TxClient interface {
	TxSign(ctx context.Context, in *TxSignReq, opts ...grpc.CallOption) (*TxSignRes, error)
	TxSend(ctx context.Context, in *TxSendReq, opts ...grpc.CallOption) (*TxSendRes, error)
	TxReceive(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[TxReceiveReq, TxReceiveRes], error)
}

type txClient struct {
	cc grpc.ClientConnInterface
}

func NewTxClient(cc grpc.ClientConnInterface) TxClient {
	return &txClient{cc}
}

func (c *txClient) TxSign(ctx context.Context, in *TxSignReq, opts ...grpc.CallOption) (*TxSignRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TxSignRes)
	err := c.cc.Invoke(ctx, Tx_TxSign_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txClient) TxSend(ctx context.Context, in *TxSendReq, opts ...grpc.CallOption) (*TxSendRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TxSendRes)
	err := c.cc.Invoke(ctx, Tx_TxSend_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txClient) TxReceive(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[TxReceiveReq, TxReceiveRes], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &Tx_ServiceDesc.Streams[0], Tx_TxReceive_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[TxReceiveReq, TxReceiveRes]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Tx_TxReceiveClient = grpc.ClientStreamingClient[TxReceiveReq, TxReceiveRes]

// TxServer is the server API for Tx service.
// All implementations must embed UnimplementedTxServer
// for forward compatibility.
type TxServer interface {
	TxSign(context.Context, *TxSignReq) (*TxSignRes, error)
	TxSend(context.Context, *TxSendReq) (*TxSendRes, error)
	TxReceive(grpc.ClientStreamingServer[TxReceiveReq, TxReceiveRes]) error
	mustEmbedUnimplementedTxServer()
}

// UnimplementedTxServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTxServer struct{}

func (UnimplementedTxServer) TxSign(context.Context, *TxSignReq) (*TxSignRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TxSign not implemented")
}
func (UnimplementedTxServer) TxSend(context.Context, *TxSendReq) (*TxSendRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TxSend not implemented")
}
func (UnimplementedTxServer) TxReceive(grpc.ClientStreamingServer[TxReceiveReq, TxReceiveRes]) error {
	return status.Errorf(codes.Unimplemented, "method TxReceive not implemented")
}
func (UnimplementedTxServer) mustEmbedUnimplementedTxServer() {}
func (UnimplementedTxServer) testEmbeddedByValue()            {}

// UnsafeTxServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TxServer will
// result in compilation errors.
type UnsafeTxServer interface {
	mustEmbedUnimplementedTxServer()
}

func RegisterTxServer(s grpc.ServiceRegistrar, srv TxServer) {
	// If the following call pancis, it indicates UnimplementedTxServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Tx_ServiceDesc, srv)
}

func _Tx_TxSign_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxSignReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServer).TxSign(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Tx_TxSign_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServer).TxSign(ctx, req.(*TxSignReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tx_TxSend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxSendReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServer).TxSend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Tx_TxSend_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServer).TxSend(ctx, req.(*TxSendReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tx_TxReceive_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TxServer).TxReceive(&grpc.GenericServerStream[TxReceiveReq, TxReceiveRes]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type Tx_TxReceiveServer = grpc.ClientStreamingServer[TxReceiveReq, TxReceiveRes]

// Tx_ServiceDesc is the grpc.ServiceDesc for Tx service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Tx_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Tx",
	HandlerType: (*TxServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "TxSign",
			Handler:    _Tx_TxSign_Handler,
		},
		{
			MethodName: "TxSend",
			Handler:    _Tx_TxSend_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "TxReceive",
			Handler:       _Tx_TxReceive_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "tx.proto",
}
