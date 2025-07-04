// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: rpc.proto

package evm

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
	RPCQueryService_ChainId_FullMethodName               = "/bds.evm.RPCQueryService/ChainId"
	RPCQueryService_GetBlockByNumber_FullMethodName      = "/bds.evm.RPCQueryService/GetBlockByNumber"
	RPCQueryService_GetBlockByHash_FullMethodName        = "/bds.evm.RPCQueryService/GetBlockByHash"
	RPCQueryService_GetLogs_FullMethodName               = "/bds.evm.RPCQueryService/GetLogs"
	RPCQueryService_GetTransactionByHash_FullMethodName  = "/bds.evm.RPCQueryService/GetTransactionByHash"
	RPCQueryService_GetTransactionReceipt_FullMethodName = "/bds.evm.RPCQueryService/GetTransactionReceipt"
	RPCQueryService_GetBlockReceipts_FullMethodName      = "/bds.evm.RPCQueryService/GetBlockReceipts"
)

// RPCQueryServiceClient is the client API for RPCQueryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Service for standard EVM RPC operations
// Equivalent to Ethereum JSON-RPC methods for node interactions
type RPCQueryServiceClient interface {
	ChainId(ctx context.Context, in *ChainIdRequest, opts ...grpc.CallOption) (*ChainIdResponse, error)
	// Get a block by its number (equivalent to eth_getBlockByNumber)
	GetBlockByNumber(ctx context.Context, in *GetBlockByNumberRequest, opts ...grpc.CallOption) (*GetBlockResponse, error)
	// Get a block by its hash (equivalent to eth_getBlockByHash)
	GetBlockByHash(ctx context.Context, in *GetBlockByHashRequest, opts ...grpc.CallOption) (*GetBlockResponse, error)
	// Get logs matching filter criteria (equivalent to eth_getLogs)
	GetLogs(ctx context.Context, in *GetLogsRequest, opts ...grpc.CallOption) (*GetLogsResponse, error)
	// Get a transaction by its hash (equivalent to eth_getTransactionByHash)
	GetTransactionByHash(ctx context.Context, in *GetTransactionByHashRequest, opts ...grpc.CallOption) (*GetTransactionByHashResponse, error)
	// Get a transaction receipt by its hash (equivalent to eth_getTransactionReceipt)
	GetTransactionReceipt(ctx context.Context, in *GetTransactionReceiptRequest, opts ...grpc.CallOption) (*GetTransactionReceiptResponse, error)
	// Get all transaction receipts for a block (equivalent to eth_getBlockReceipts)
	GetBlockReceipts(ctx context.Context, in *GetBlockReceiptsRequest, opts ...grpc.CallOption) (*GetBlockReceiptsResponse, error)
}

type rPCQueryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRPCQueryServiceClient(cc grpc.ClientConnInterface) RPCQueryServiceClient {
	return &rPCQueryServiceClient{cc}
}

func (c *rPCQueryServiceClient) ChainId(ctx context.Context, in *ChainIdRequest, opts ...grpc.CallOption) (*ChainIdResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ChainIdResponse)
	err := c.cc.Invoke(ctx, RPCQueryService_ChainId_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rPCQueryServiceClient) GetBlockByNumber(ctx context.Context, in *GetBlockByNumberRequest, opts ...grpc.CallOption) (*GetBlockResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetBlockResponse)
	err := c.cc.Invoke(ctx, RPCQueryService_GetBlockByNumber_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rPCQueryServiceClient) GetBlockByHash(ctx context.Context, in *GetBlockByHashRequest, opts ...grpc.CallOption) (*GetBlockResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetBlockResponse)
	err := c.cc.Invoke(ctx, RPCQueryService_GetBlockByHash_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rPCQueryServiceClient) GetLogs(ctx context.Context, in *GetLogsRequest, opts ...grpc.CallOption) (*GetLogsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetLogsResponse)
	err := c.cc.Invoke(ctx, RPCQueryService_GetLogs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rPCQueryServiceClient) GetTransactionByHash(ctx context.Context, in *GetTransactionByHashRequest, opts ...grpc.CallOption) (*GetTransactionByHashResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetTransactionByHashResponse)
	err := c.cc.Invoke(ctx, RPCQueryService_GetTransactionByHash_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rPCQueryServiceClient) GetTransactionReceipt(ctx context.Context, in *GetTransactionReceiptRequest, opts ...grpc.CallOption) (*GetTransactionReceiptResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetTransactionReceiptResponse)
	err := c.cc.Invoke(ctx, RPCQueryService_GetTransactionReceipt_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rPCQueryServiceClient) GetBlockReceipts(ctx context.Context, in *GetBlockReceiptsRequest, opts ...grpc.CallOption) (*GetBlockReceiptsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetBlockReceiptsResponse)
	err := c.cc.Invoke(ctx, RPCQueryService_GetBlockReceipts_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RPCQueryServiceServer is the server API for RPCQueryService service.
// All implementations must embed UnimplementedRPCQueryServiceServer
// for forward compatibility.
//
// Service for standard EVM RPC operations
// Equivalent to Ethereum JSON-RPC methods for node interactions
type RPCQueryServiceServer interface {
	ChainId(context.Context, *ChainIdRequest) (*ChainIdResponse, error)
	// Get a block by its number (equivalent to eth_getBlockByNumber)
	GetBlockByNumber(context.Context, *GetBlockByNumberRequest) (*GetBlockResponse, error)
	// Get a block by its hash (equivalent to eth_getBlockByHash)
	GetBlockByHash(context.Context, *GetBlockByHashRequest) (*GetBlockResponse, error)
	// Get logs matching filter criteria (equivalent to eth_getLogs)
	GetLogs(context.Context, *GetLogsRequest) (*GetLogsResponse, error)
	// Get a transaction by its hash (equivalent to eth_getTransactionByHash)
	GetTransactionByHash(context.Context, *GetTransactionByHashRequest) (*GetTransactionByHashResponse, error)
	// Get a transaction receipt by its hash (equivalent to eth_getTransactionReceipt)
	GetTransactionReceipt(context.Context, *GetTransactionReceiptRequest) (*GetTransactionReceiptResponse, error)
	// Get all transaction receipts for a block (equivalent to eth_getBlockReceipts)
	GetBlockReceipts(context.Context, *GetBlockReceiptsRequest) (*GetBlockReceiptsResponse, error)
	mustEmbedUnimplementedRPCQueryServiceServer()
}

// UnimplementedRPCQueryServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedRPCQueryServiceServer struct{}

func (UnimplementedRPCQueryServiceServer) ChainId(context.Context, *ChainIdRequest) (*ChainIdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChainId not implemented")
}
func (UnimplementedRPCQueryServiceServer) GetBlockByNumber(context.Context, *GetBlockByNumberRequest) (*GetBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlockByNumber not implemented")
}
func (UnimplementedRPCQueryServiceServer) GetBlockByHash(context.Context, *GetBlockByHashRequest) (*GetBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlockByHash not implemented")
}
func (UnimplementedRPCQueryServiceServer) GetLogs(context.Context, *GetLogsRequest) (*GetLogsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLogs not implemented")
}
func (UnimplementedRPCQueryServiceServer) GetTransactionByHash(context.Context, *GetTransactionByHashRequest) (*GetTransactionByHashResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransactionByHash not implemented")
}
func (UnimplementedRPCQueryServiceServer) GetTransactionReceipt(context.Context, *GetTransactionReceiptRequest) (*GetTransactionReceiptResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransactionReceipt not implemented")
}
func (UnimplementedRPCQueryServiceServer) GetBlockReceipts(context.Context, *GetBlockReceiptsRequest) (*GetBlockReceiptsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlockReceipts not implemented")
}
func (UnimplementedRPCQueryServiceServer) mustEmbedUnimplementedRPCQueryServiceServer() {}
func (UnimplementedRPCQueryServiceServer) testEmbeddedByValue()                         {}

// UnsafeRPCQueryServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RPCQueryServiceServer will
// result in compilation errors.
type UnsafeRPCQueryServiceServer interface {
	mustEmbedUnimplementedRPCQueryServiceServer()
}

func RegisterRPCQueryServiceServer(s grpc.ServiceRegistrar, srv RPCQueryServiceServer) {
	// If the following call pancis, it indicates UnimplementedRPCQueryServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&RPCQueryService_ServiceDesc, srv)
}

func _RPCQueryService_ChainId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChainIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCQueryServiceServer).ChainId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RPCQueryService_ChainId_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCQueryServiceServer).ChainId(ctx, req.(*ChainIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RPCQueryService_GetBlockByNumber_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockByNumberRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCQueryServiceServer).GetBlockByNumber(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RPCQueryService_GetBlockByNumber_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCQueryServiceServer).GetBlockByNumber(ctx, req.(*GetBlockByNumberRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RPCQueryService_GetBlockByHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockByHashRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCQueryServiceServer).GetBlockByHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RPCQueryService_GetBlockByHash_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCQueryServiceServer).GetBlockByHash(ctx, req.(*GetBlockByHashRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RPCQueryService_GetLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLogsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCQueryServiceServer).GetLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RPCQueryService_GetLogs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCQueryServiceServer).GetLogs(ctx, req.(*GetLogsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RPCQueryService_GetTransactionByHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTransactionByHashRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCQueryServiceServer).GetTransactionByHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RPCQueryService_GetTransactionByHash_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCQueryServiceServer).GetTransactionByHash(ctx, req.(*GetTransactionByHashRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RPCQueryService_GetTransactionReceipt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTransactionReceiptRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCQueryServiceServer).GetTransactionReceipt(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RPCQueryService_GetTransactionReceipt_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCQueryServiceServer).GetTransactionReceipt(ctx, req.(*GetTransactionReceiptRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RPCQueryService_GetBlockReceipts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockReceiptsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCQueryServiceServer).GetBlockReceipts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RPCQueryService_GetBlockReceipts_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCQueryServiceServer).GetBlockReceipts(ctx, req.(*GetBlockReceiptsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RPCQueryService_ServiceDesc is the grpc.ServiceDesc for RPCQueryService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RPCQueryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "bds.evm.RPCQueryService",
	HandlerType: (*RPCQueryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ChainId",
			Handler:    _RPCQueryService_ChainId_Handler,
		},
		{
			MethodName: "GetBlockByNumber",
			Handler:    _RPCQueryService_GetBlockByNumber_Handler,
		},
		{
			MethodName: "GetBlockByHash",
			Handler:    _RPCQueryService_GetBlockByHash_Handler,
		},
		{
			MethodName: "GetLogs",
			Handler:    _RPCQueryService_GetLogs_Handler,
		},
		{
			MethodName: "GetTransactionByHash",
			Handler:    _RPCQueryService_GetTransactionByHash_Handler,
		},
		{
			MethodName: "GetTransactionReceipt",
			Handler:    _RPCQueryService_GetTransactionReceipt_Handler,
		},
		{
			MethodName: "GetBlockReceipts",
			Handler:    _RPCQueryService_GetBlockReceipts_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc.proto",
}
