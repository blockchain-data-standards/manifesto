package bds.evm;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Service for standard EVM RPC operations
 * Equivalent to Ethereum JSON-RPC methods for node interactions
 * </pre>
 */
@io.grpc.stub.annotations.GrpcGenerated
public final class RPCQueryServiceGrpc {

  private RPCQueryServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "bds.evm.RPCQueryService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<bds.evm.Rpc.ChainIdRequest,
      bds.evm.Rpc.ChainIdResponse> getChainIdMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ChainId",
      requestType = bds.evm.Rpc.ChainIdRequest.class,
      responseType = bds.evm.Rpc.ChainIdResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Rpc.ChainIdRequest,
      bds.evm.Rpc.ChainIdResponse> getChainIdMethod() {
    io.grpc.MethodDescriptor<bds.evm.Rpc.ChainIdRequest, bds.evm.Rpc.ChainIdResponse> getChainIdMethod;
    if ((getChainIdMethod = RPCQueryServiceGrpc.getChainIdMethod) == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        if ((getChainIdMethod = RPCQueryServiceGrpc.getChainIdMethod) == null) {
          RPCQueryServiceGrpc.getChainIdMethod = getChainIdMethod =
              io.grpc.MethodDescriptor.<bds.evm.Rpc.ChainIdRequest, bds.evm.Rpc.ChainIdResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ChainId"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.ChainIdRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.ChainIdResponse.getDefaultInstance()))
              .setSchemaDescriptor(new RPCQueryServiceMethodDescriptorSupplier("ChainId"))
              .build();
        }
      }
    }
    return getChainIdMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Rpc.GetBlockByNumberRequest,
      bds.evm.Rpc.GetBlockResponse> getGetBlockByNumberMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetBlockByNumber",
      requestType = bds.evm.Rpc.GetBlockByNumberRequest.class,
      responseType = bds.evm.Rpc.GetBlockResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Rpc.GetBlockByNumberRequest,
      bds.evm.Rpc.GetBlockResponse> getGetBlockByNumberMethod() {
    io.grpc.MethodDescriptor<bds.evm.Rpc.GetBlockByNumberRequest, bds.evm.Rpc.GetBlockResponse> getGetBlockByNumberMethod;
    if ((getGetBlockByNumberMethod = RPCQueryServiceGrpc.getGetBlockByNumberMethod) == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        if ((getGetBlockByNumberMethod = RPCQueryServiceGrpc.getGetBlockByNumberMethod) == null) {
          RPCQueryServiceGrpc.getGetBlockByNumberMethod = getGetBlockByNumberMethod =
              io.grpc.MethodDescriptor.<bds.evm.Rpc.GetBlockByNumberRequest, bds.evm.Rpc.GetBlockResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetBlockByNumber"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetBlockByNumberRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetBlockResponse.getDefaultInstance()))
              .setSchemaDescriptor(new RPCQueryServiceMethodDescriptorSupplier("GetBlockByNumber"))
              .build();
        }
      }
    }
    return getGetBlockByNumberMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Rpc.GetBlockByHashRequest,
      bds.evm.Rpc.GetBlockResponse> getGetBlockByHashMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetBlockByHash",
      requestType = bds.evm.Rpc.GetBlockByHashRequest.class,
      responseType = bds.evm.Rpc.GetBlockResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Rpc.GetBlockByHashRequest,
      bds.evm.Rpc.GetBlockResponse> getGetBlockByHashMethod() {
    io.grpc.MethodDescriptor<bds.evm.Rpc.GetBlockByHashRequest, bds.evm.Rpc.GetBlockResponse> getGetBlockByHashMethod;
    if ((getGetBlockByHashMethod = RPCQueryServiceGrpc.getGetBlockByHashMethod) == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        if ((getGetBlockByHashMethod = RPCQueryServiceGrpc.getGetBlockByHashMethod) == null) {
          RPCQueryServiceGrpc.getGetBlockByHashMethod = getGetBlockByHashMethod =
              io.grpc.MethodDescriptor.<bds.evm.Rpc.GetBlockByHashRequest, bds.evm.Rpc.GetBlockResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetBlockByHash"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetBlockByHashRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetBlockResponse.getDefaultInstance()))
              .setSchemaDescriptor(new RPCQueryServiceMethodDescriptorSupplier("GetBlockByHash"))
              .build();
        }
      }
    }
    return getGetBlockByHashMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Rpc.GetLogsRequest,
      bds.evm.Rpc.GetLogsResponse> getGetLogsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetLogs",
      requestType = bds.evm.Rpc.GetLogsRequest.class,
      responseType = bds.evm.Rpc.GetLogsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Rpc.GetLogsRequest,
      bds.evm.Rpc.GetLogsResponse> getGetLogsMethod() {
    io.grpc.MethodDescriptor<bds.evm.Rpc.GetLogsRequest, bds.evm.Rpc.GetLogsResponse> getGetLogsMethod;
    if ((getGetLogsMethod = RPCQueryServiceGrpc.getGetLogsMethod) == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        if ((getGetLogsMethod = RPCQueryServiceGrpc.getGetLogsMethod) == null) {
          RPCQueryServiceGrpc.getGetLogsMethod = getGetLogsMethod =
              io.grpc.MethodDescriptor.<bds.evm.Rpc.GetLogsRequest, bds.evm.Rpc.GetLogsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetLogs"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetLogsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetLogsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new RPCQueryServiceMethodDescriptorSupplier("GetLogs"))
              .build();
        }
      }
    }
    return getGetLogsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Rpc.GetTransactionByHashRequest,
      bds.evm.Rpc.GetTransactionByHashResponse> getGetTransactionByHashMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetTransactionByHash",
      requestType = bds.evm.Rpc.GetTransactionByHashRequest.class,
      responseType = bds.evm.Rpc.GetTransactionByHashResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Rpc.GetTransactionByHashRequest,
      bds.evm.Rpc.GetTransactionByHashResponse> getGetTransactionByHashMethod() {
    io.grpc.MethodDescriptor<bds.evm.Rpc.GetTransactionByHashRequest, bds.evm.Rpc.GetTransactionByHashResponse> getGetTransactionByHashMethod;
    if ((getGetTransactionByHashMethod = RPCQueryServiceGrpc.getGetTransactionByHashMethod) == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        if ((getGetTransactionByHashMethod = RPCQueryServiceGrpc.getGetTransactionByHashMethod) == null) {
          RPCQueryServiceGrpc.getGetTransactionByHashMethod = getGetTransactionByHashMethod =
              io.grpc.MethodDescriptor.<bds.evm.Rpc.GetTransactionByHashRequest, bds.evm.Rpc.GetTransactionByHashResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetTransactionByHash"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetTransactionByHashRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetTransactionByHashResponse.getDefaultInstance()))
              .setSchemaDescriptor(new RPCQueryServiceMethodDescriptorSupplier("GetTransactionByHash"))
              .build();
        }
      }
    }
    return getGetTransactionByHashMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Rpc.GetTransactionReceiptRequest,
      bds.evm.Rpc.GetTransactionReceiptResponse> getGetTransactionReceiptMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetTransactionReceipt",
      requestType = bds.evm.Rpc.GetTransactionReceiptRequest.class,
      responseType = bds.evm.Rpc.GetTransactionReceiptResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Rpc.GetTransactionReceiptRequest,
      bds.evm.Rpc.GetTransactionReceiptResponse> getGetTransactionReceiptMethod() {
    io.grpc.MethodDescriptor<bds.evm.Rpc.GetTransactionReceiptRequest, bds.evm.Rpc.GetTransactionReceiptResponse> getGetTransactionReceiptMethod;
    if ((getGetTransactionReceiptMethod = RPCQueryServiceGrpc.getGetTransactionReceiptMethod) == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        if ((getGetTransactionReceiptMethod = RPCQueryServiceGrpc.getGetTransactionReceiptMethod) == null) {
          RPCQueryServiceGrpc.getGetTransactionReceiptMethod = getGetTransactionReceiptMethod =
              io.grpc.MethodDescriptor.<bds.evm.Rpc.GetTransactionReceiptRequest, bds.evm.Rpc.GetTransactionReceiptResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetTransactionReceipt"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetTransactionReceiptRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetTransactionReceiptResponse.getDefaultInstance()))
              .setSchemaDescriptor(new RPCQueryServiceMethodDescriptorSupplier("GetTransactionReceipt"))
              .build();
        }
      }
    }
    return getGetTransactionReceiptMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Rpc.GetBlockReceiptsRequest,
      bds.evm.Rpc.GetBlockReceiptsResponse> getGetBlockReceiptsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetBlockReceipts",
      requestType = bds.evm.Rpc.GetBlockReceiptsRequest.class,
      responseType = bds.evm.Rpc.GetBlockReceiptsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Rpc.GetBlockReceiptsRequest,
      bds.evm.Rpc.GetBlockReceiptsResponse> getGetBlockReceiptsMethod() {
    io.grpc.MethodDescriptor<bds.evm.Rpc.GetBlockReceiptsRequest, bds.evm.Rpc.GetBlockReceiptsResponse> getGetBlockReceiptsMethod;
    if ((getGetBlockReceiptsMethod = RPCQueryServiceGrpc.getGetBlockReceiptsMethod) == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        if ((getGetBlockReceiptsMethod = RPCQueryServiceGrpc.getGetBlockReceiptsMethod) == null) {
          RPCQueryServiceGrpc.getGetBlockReceiptsMethod = getGetBlockReceiptsMethod =
              io.grpc.MethodDescriptor.<bds.evm.Rpc.GetBlockReceiptsRequest, bds.evm.Rpc.GetBlockReceiptsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetBlockReceipts"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetBlockReceiptsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetBlockReceiptsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new RPCQueryServiceMethodDescriptorSupplier("GetBlockReceipts"))
              .build();
        }
      }
    }
    return getGetBlockReceiptsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Rpc.TraceBlockRequest,
      bds.evm.Rpc.TraceResponse> getTraceBlockMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "TraceBlock",
      requestType = bds.evm.Rpc.TraceBlockRequest.class,
      responseType = bds.evm.Rpc.TraceResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Rpc.TraceBlockRequest,
      bds.evm.Rpc.TraceResponse> getTraceBlockMethod() {
    io.grpc.MethodDescriptor<bds.evm.Rpc.TraceBlockRequest, bds.evm.Rpc.TraceResponse> getTraceBlockMethod;
    if ((getTraceBlockMethod = RPCQueryServiceGrpc.getTraceBlockMethod) == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        if ((getTraceBlockMethod = RPCQueryServiceGrpc.getTraceBlockMethod) == null) {
          RPCQueryServiceGrpc.getTraceBlockMethod = getTraceBlockMethod =
              io.grpc.MethodDescriptor.<bds.evm.Rpc.TraceBlockRequest, bds.evm.Rpc.TraceResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "TraceBlock"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.TraceBlockRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.TraceResponse.getDefaultInstance()))
              .setSchemaDescriptor(new RPCQueryServiceMethodDescriptorSupplier("TraceBlock"))
              .build();
        }
      }
    }
    return getTraceBlockMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Rpc.TraceTransactionRequest,
      bds.evm.Rpc.TraceResponse> getTraceTransactionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "TraceTransaction",
      requestType = bds.evm.Rpc.TraceTransactionRequest.class,
      responseType = bds.evm.Rpc.TraceResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Rpc.TraceTransactionRequest,
      bds.evm.Rpc.TraceResponse> getTraceTransactionMethod() {
    io.grpc.MethodDescriptor<bds.evm.Rpc.TraceTransactionRequest, bds.evm.Rpc.TraceResponse> getTraceTransactionMethod;
    if ((getTraceTransactionMethod = RPCQueryServiceGrpc.getTraceTransactionMethod) == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        if ((getTraceTransactionMethod = RPCQueryServiceGrpc.getTraceTransactionMethod) == null) {
          RPCQueryServiceGrpc.getTraceTransactionMethod = getTraceTransactionMethod =
              io.grpc.MethodDescriptor.<bds.evm.Rpc.TraceTransactionRequest, bds.evm.Rpc.TraceResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "TraceTransaction"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.TraceTransactionRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.TraceResponse.getDefaultInstance()))
              .setSchemaDescriptor(new RPCQueryServiceMethodDescriptorSupplier("TraceTransaction"))
              .build();
        }
      }
    }
    return getTraceTransactionMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Rpc.ParityTraceBlockRequest,
      bds.evm.Rpc.TraceResponse> getParityTraceBlockMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ParityTraceBlock",
      requestType = bds.evm.Rpc.ParityTraceBlockRequest.class,
      responseType = bds.evm.Rpc.TraceResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Rpc.ParityTraceBlockRequest,
      bds.evm.Rpc.TraceResponse> getParityTraceBlockMethod() {
    io.grpc.MethodDescriptor<bds.evm.Rpc.ParityTraceBlockRequest, bds.evm.Rpc.TraceResponse> getParityTraceBlockMethod;
    if ((getParityTraceBlockMethod = RPCQueryServiceGrpc.getParityTraceBlockMethod) == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        if ((getParityTraceBlockMethod = RPCQueryServiceGrpc.getParityTraceBlockMethod) == null) {
          RPCQueryServiceGrpc.getParityTraceBlockMethod = getParityTraceBlockMethod =
              io.grpc.MethodDescriptor.<bds.evm.Rpc.ParityTraceBlockRequest, bds.evm.Rpc.TraceResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ParityTraceBlock"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.ParityTraceBlockRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.TraceResponse.getDefaultInstance()))
              .setSchemaDescriptor(new RPCQueryServiceMethodDescriptorSupplier("ParityTraceBlock"))
              .build();
        }
      }
    }
    return getParityTraceBlockMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Rpc.GetBalanceRequest,
      bds.evm.Rpc.GetBalanceResponse> getGetBalanceMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetBalance",
      requestType = bds.evm.Rpc.GetBalanceRequest.class,
      responseType = bds.evm.Rpc.GetBalanceResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Rpc.GetBalanceRequest,
      bds.evm.Rpc.GetBalanceResponse> getGetBalanceMethod() {
    io.grpc.MethodDescriptor<bds.evm.Rpc.GetBalanceRequest, bds.evm.Rpc.GetBalanceResponse> getGetBalanceMethod;
    if ((getGetBalanceMethod = RPCQueryServiceGrpc.getGetBalanceMethod) == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        if ((getGetBalanceMethod = RPCQueryServiceGrpc.getGetBalanceMethod) == null) {
          RPCQueryServiceGrpc.getGetBalanceMethod = getGetBalanceMethod =
              io.grpc.MethodDescriptor.<bds.evm.Rpc.GetBalanceRequest, bds.evm.Rpc.GetBalanceResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetBalance"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetBalanceRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Rpc.GetBalanceResponse.getDefaultInstance()))
              .setSchemaDescriptor(new RPCQueryServiceMethodDescriptorSupplier("GetBalance"))
              .build();
        }
      }
    }
    return getGetBalanceMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static RPCQueryServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<RPCQueryServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<RPCQueryServiceStub>() {
        @java.lang.Override
        public RPCQueryServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new RPCQueryServiceStub(channel, callOptions);
        }
      };
    return RPCQueryServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports all types of calls on the service
   */
  public static RPCQueryServiceBlockingV2Stub newBlockingV2Stub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<RPCQueryServiceBlockingV2Stub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<RPCQueryServiceBlockingV2Stub>() {
        @java.lang.Override
        public RPCQueryServiceBlockingV2Stub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new RPCQueryServiceBlockingV2Stub(channel, callOptions);
        }
      };
    return RPCQueryServiceBlockingV2Stub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static RPCQueryServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<RPCQueryServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<RPCQueryServiceBlockingStub>() {
        @java.lang.Override
        public RPCQueryServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new RPCQueryServiceBlockingStub(channel, callOptions);
        }
      };
    return RPCQueryServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static RPCQueryServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<RPCQueryServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<RPCQueryServiceFutureStub>() {
        @java.lang.Override
        public RPCQueryServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new RPCQueryServiceFutureStub(channel, callOptions);
        }
      };
    return RPCQueryServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * Service for standard EVM RPC operations
   * Equivalent to Ethereum JSON-RPC methods for node interactions
   * </pre>
   */
  public interface AsyncService {

    /**
     */
    default void chainId(bds.evm.Rpc.ChainIdRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.ChainIdResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getChainIdMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get a block by its number (equivalent to eth_getBlockByNumber)
     * </pre>
     */
    default void getBlockByNumber(bds.evm.Rpc.GetBlockByNumberRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBlockResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetBlockByNumberMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get a block by its hash (equivalent to eth_getBlockByHash)
     * </pre>
     */
    default void getBlockByHash(bds.evm.Rpc.GetBlockByHashRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBlockResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetBlockByHashMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get logs matching filter criteria (equivalent to eth_getLogs)
     * </pre>
     */
    default void getLogs(bds.evm.Rpc.GetLogsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetLogsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetLogsMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get a transaction by its hash (equivalent to eth_getTransactionByHash)
     * </pre>
     */
    default void getTransactionByHash(bds.evm.Rpc.GetTransactionByHashRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetTransactionByHashResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetTransactionByHashMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get a transaction receipt by its hash (equivalent to eth_getTransactionReceipt)
     * </pre>
     */
    default void getTransactionReceipt(bds.evm.Rpc.GetTransactionReceiptRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetTransactionReceiptResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetTransactionReceiptMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get all transaction receipts for a block (equivalent to eth_getBlockReceipts)
     * </pre>
     */
    default void getBlockReceipts(bds.evm.Rpc.GetBlockReceiptsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBlockReceiptsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetBlockReceiptsMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get the trace of every transaction in a block (equivalent to
     * debug_traceBlockByNumber and debug_traceBlockByHash).
     * </pre>
     */
    default void traceBlock(bds.evm.Rpc.TraceBlockRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.TraceResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getTraceBlockMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get the trace of one transaction (equivalent to debug_traceTransaction)
     * </pre>
     */
    default void traceTransaction(bds.evm.Rpc.TraceTransactionRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.TraceResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getTraceTransactionMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get the Parity-style traces of every transaction in a block (equivalent
     * to trace_block)
     * </pre>
     */
    default void parityTraceBlock(bds.evm.Rpc.ParityTraceBlockRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.TraceResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getParityTraceBlockMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get the native balance of an address at a block (equivalent to eth_getBalance)
     * </pre>
     */
    default void getBalance(bds.evm.Rpc.GetBalanceRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBalanceResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetBalanceMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service RPCQueryService.
   * <pre>
   * Service for standard EVM RPC operations
   * Equivalent to Ethereum JSON-RPC methods for node interactions
   * </pre>
   */
  public static abstract class RPCQueryServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return RPCQueryServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service RPCQueryService.
   * <pre>
   * Service for standard EVM RPC operations
   * Equivalent to Ethereum JSON-RPC methods for node interactions
   * </pre>
   */
  public static final class RPCQueryServiceStub
      extends io.grpc.stub.AbstractAsyncStub<RPCQueryServiceStub> {
    private RPCQueryServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected RPCQueryServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new RPCQueryServiceStub(channel, callOptions);
    }

    /**
     */
    public void chainId(bds.evm.Rpc.ChainIdRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.ChainIdResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getChainIdMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get a block by its number (equivalent to eth_getBlockByNumber)
     * </pre>
     */
    public void getBlockByNumber(bds.evm.Rpc.GetBlockByNumberRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBlockResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetBlockByNumberMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get a block by its hash (equivalent to eth_getBlockByHash)
     * </pre>
     */
    public void getBlockByHash(bds.evm.Rpc.GetBlockByHashRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBlockResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetBlockByHashMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get logs matching filter criteria (equivalent to eth_getLogs)
     * </pre>
     */
    public void getLogs(bds.evm.Rpc.GetLogsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetLogsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetLogsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get a transaction by its hash (equivalent to eth_getTransactionByHash)
     * </pre>
     */
    public void getTransactionByHash(bds.evm.Rpc.GetTransactionByHashRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetTransactionByHashResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetTransactionByHashMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get a transaction receipt by its hash (equivalent to eth_getTransactionReceipt)
     * </pre>
     */
    public void getTransactionReceipt(bds.evm.Rpc.GetTransactionReceiptRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetTransactionReceiptResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetTransactionReceiptMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get all transaction receipts for a block (equivalent to eth_getBlockReceipts)
     * </pre>
     */
    public void getBlockReceipts(bds.evm.Rpc.GetBlockReceiptsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBlockReceiptsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetBlockReceiptsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get the trace of every transaction in a block (equivalent to
     * debug_traceBlockByNumber and debug_traceBlockByHash).
     * </pre>
     */
    public void traceBlock(bds.evm.Rpc.TraceBlockRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.TraceResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getTraceBlockMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get the trace of one transaction (equivalent to debug_traceTransaction)
     * </pre>
     */
    public void traceTransaction(bds.evm.Rpc.TraceTransactionRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.TraceResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getTraceTransactionMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get the Parity-style traces of every transaction in a block (equivalent
     * to trace_block)
     * </pre>
     */
    public void parityTraceBlock(bds.evm.Rpc.ParityTraceBlockRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.TraceResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getParityTraceBlockMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get the native balance of an address at a block (equivalent to eth_getBalance)
     * </pre>
     */
    public void getBalance(bds.evm.Rpc.GetBalanceRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBalanceResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetBalanceMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service RPCQueryService.
   * <pre>
   * Service for standard EVM RPC operations
   * Equivalent to Ethereum JSON-RPC methods for node interactions
   * </pre>
   */
  public static final class RPCQueryServiceBlockingV2Stub
      extends io.grpc.stub.AbstractBlockingStub<RPCQueryServiceBlockingV2Stub> {
    private RPCQueryServiceBlockingV2Stub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected RPCQueryServiceBlockingV2Stub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new RPCQueryServiceBlockingV2Stub(channel, callOptions);
    }

    /**
     */
    public bds.evm.Rpc.ChainIdResponse chainId(bds.evm.Rpc.ChainIdRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getChainIdMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get a block by its number (equivalent to eth_getBlockByNumber)
     * </pre>
     */
    public bds.evm.Rpc.GetBlockResponse getBlockByNumber(bds.evm.Rpc.GetBlockByNumberRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getGetBlockByNumberMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get a block by its hash (equivalent to eth_getBlockByHash)
     * </pre>
     */
    public bds.evm.Rpc.GetBlockResponse getBlockByHash(bds.evm.Rpc.GetBlockByHashRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getGetBlockByHashMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get logs matching filter criteria (equivalent to eth_getLogs)
     * </pre>
     */
    public bds.evm.Rpc.GetLogsResponse getLogs(bds.evm.Rpc.GetLogsRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getGetLogsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get a transaction by its hash (equivalent to eth_getTransactionByHash)
     * </pre>
     */
    public bds.evm.Rpc.GetTransactionByHashResponse getTransactionByHash(bds.evm.Rpc.GetTransactionByHashRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getGetTransactionByHashMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get a transaction receipt by its hash (equivalent to eth_getTransactionReceipt)
     * </pre>
     */
    public bds.evm.Rpc.GetTransactionReceiptResponse getTransactionReceipt(bds.evm.Rpc.GetTransactionReceiptRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getGetTransactionReceiptMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get all transaction receipts for a block (equivalent to eth_getBlockReceipts)
     * </pre>
     */
    public bds.evm.Rpc.GetBlockReceiptsResponse getBlockReceipts(bds.evm.Rpc.GetBlockReceiptsRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getGetBlockReceiptsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get the trace of every transaction in a block (equivalent to
     * debug_traceBlockByNumber and debug_traceBlockByHash).
     * </pre>
     */
    public bds.evm.Rpc.TraceResponse traceBlock(bds.evm.Rpc.TraceBlockRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getTraceBlockMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get the trace of one transaction (equivalent to debug_traceTransaction)
     * </pre>
     */
    public bds.evm.Rpc.TraceResponse traceTransaction(bds.evm.Rpc.TraceTransactionRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getTraceTransactionMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get the Parity-style traces of every transaction in a block (equivalent
     * to trace_block)
     * </pre>
     */
    public bds.evm.Rpc.TraceResponse parityTraceBlock(bds.evm.Rpc.ParityTraceBlockRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getParityTraceBlockMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get the native balance of an address at a block (equivalent to eth_getBalance)
     * </pre>
     */
    public bds.evm.Rpc.GetBalanceResponse getBalance(bds.evm.Rpc.GetBalanceRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getGetBalanceMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do limited synchronous rpc calls to service RPCQueryService.
   * <pre>
   * Service for standard EVM RPC operations
   * Equivalent to Ethereum JSON-RPC methods for node interactions
   * </pre>
   */
  public static final class RPCQueryServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<RPCQueryServiceBlockingStub> {
    private RPCQueryServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected RPCQueryServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new RPCQueryServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public bds.evm.Rpc.ChainIdResponse chainId(bds.evm.Rpc.ChainIdRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getChainIdMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get a block by its number (equivalent to eth_getBlockByNumber)
     * </pre>
     */
    public bds.evm.Rpc.GetBlockResponse getBlockByNumber(bds.evm.Rpc.GetBlockByNumberRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetBlockByNumberMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get a block by its hash (equivalent to eth_getBlockByHash)
     * </pre>
     */
    public bds.evm.Rpc.GetBlockResponse getBlockByHash(bds.evm.Rpc.GetBlockByHashRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetBlockByHashMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get logs matching filter criteria (equivalent to eth_getLogs)
     * </pre>
     */
    public bds.evm.Rpc.GetLogsResponse getLogs(bds.evm.Rpc.GetLogsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetLogsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get a transaction by its hash (equivalent to eth_getTransactionByHash)
     * </pre>
     */
    public bds.evm.Rpc.GetTransactionByHashResponse getTransactionByHash(bds.evm.Rpc.GetTransactionByHashRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetTransactionByHashMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get a transaction receipt by its hash (equivalent to eth_getTransactionReceipt)
     * </pre>
     */
    public bds.evm.Rpc.GetTransactionReceiptResponse getTransactionReceipt(bds.evm.Rpc.GetTransactionReceiptRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetTransactionReceiptMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get all transaction receipts for a block (equivalent to eth_getBlockReceipts)
     * </pre>
     */
    public bds.evm.Rpc.GetBlockReceiptsResponse getBlockReceipts(bds.evm.Rpc.GetBlockReceiptsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetBlockReceiptsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get the trace of every transaction in a block (equivalent to
     * debug_traceBlockByNumber and debug_traceBlockByHash).
     * </pre>
     */
    public bds.evm.Rpc.TraceResponse traceBlock(bds.evm.Rpc.TraceBlockRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getTraceBlockMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get the trace of one transaction (equivalent to debug_traceTransaction)
     * </pre>
     */
    public bds.evm.Rpc.TraceResponse traceTransaction(bds.evm.Rpc.TraceTransactionRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getTraceTransactionMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get the Parity-style traces of every transaction in a block (equivalent
     * to trace_block)
     * </pre>
     */
    public bds.evm.Rpc.TraceResponse parityTraceBlock(bds.evm.Rpc.ParityTraceBlockRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getParityTraceBlockMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get the native balance of an address at a block (equivalent to eth_getBalance)
     * </pre>
     */
    public bds.evm.Rpc.GetBalanceResponse getBalance(bds.evm.Rpc.GetBalanceRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetBalanceMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service RPCQueryService.
   * <pre>
   * Service for standard EVM RPC operations
   * Equivalent to Ethereum JSON-RPC methods for node interactions
   * </pre>
   */
  public static final class RPCQueryServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<RPCQueryServiceFutureStub> {
    private RPCQueryServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected RPCQueryServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new RPCQueryServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Rpc.ChainIdResponse> chainId(
        bds.evm.Rpc.ChainIdRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getChainIdMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get a block by its number (equivalent to eth_getBlockByNumber)
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Rpc.GetBlockResponse> getBlockByNumber(
        bds.evm.Rpc.GetBlockByNumberRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetBlockByNumberMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get a block by its hash (equivalent to eth_getBlockByHash)
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Rpc.GetBlockResponse> getBlockByHash(
        bds.evm.Rpc.GetBlockByHashRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetBlockByHashMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get logs matching filter criteria (equivalent to eth_getLogs)
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Rpc.GetLogsResponse> getLogs(
        bds.evm.Rpc.GetLogsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetLogsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get a transaction by its hash (equivalent to eth_getTransactionByHash)
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Rpc.GetTransactionByHashResponse> getTransactionByHash(
        bds.evm.Rpc.GetTransactionByHashRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetTransactionByHashMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get a transaction receipt by its hash (equivalent to eth_getTransactionReceipt)
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Rpc.GetTransactionReceiptResponse> getTransactionReceipt(
        bds.evm.Rpc.GetTransactionReceiptRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetTransactionReceiptMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get all transaction receipts for a block (equivalent to eth_getBlockReceipts)
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Rpc.GetBlockReceiptsResponse> getBlockReceipts(
        bds.evm.Rpc.GetBlockReceiptsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetBlockReceiptsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get the trace of every transaction in a block (equivalent to
     * debug_traceBlockByNumber and debug_traceBlockByHash).
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Rpc.TraceResponse> traceBlock(
        bds.evm.Rpc.TraceBlockRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getTraceBlockMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get the trace of one transaction (equivalent to debug_traceTransaction)
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Rpc.TraceResponse> traceTransaction(
        bds.evm.Rpc.TraceTransactionRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getTraceTransactionMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get the Parity-style traces of every transaction in a block (equivalent
     * to trace_block)
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Rpc.TraceResponse> parityTraceBlock(
        bds.evm.Rpc.ParityTraceBlockRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getParityTraceBlockMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get the native balance of an address at a block (equivalent to eth_getBalance)
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Rpc.GetBalanceResponse> getBalance(
        bds.evm.Rpc.GetBalanceRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetBalanceMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_CHAIN_ID = 0;
  private static final int METHODID_GET_BLOCK_BY_NUMBER = 1;
  private static final int METHODID_GET_BLOCK_BY_HASH = 2;
  private static final int METHODID_GET_LOGS = 3;
  private static final int METHODID_GET_TRANSACTION_BY_HASH = 4;
  private static final int METHODID_GET_TRANSACTION_RECEIPT = 5;
  private static final int METHODID_GET_BLOCK_RECEIPTS = 6;
  private static final int METHODID_TRACE_BLOCK = 7;
  private static final int METHODID_TRACE_TRANSACTION = 8;
  private static final int METHODID_PARITY_TRACE_BLOCK = 9;
  private static final int METHODID_GET_BALANCE = 10;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final AsyncService serviceImpl;
    private final int methodId;

    MethodHandlers(AsyncService serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_CHAIN_ID:
          serviceImpl.chainId((bds.evm.Rpc.ChainIdRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Rpc.ChainIdResponse>) responseObserver);
          break;
        case METHODID_GET_BLOCK_BY_NUMBER:
          serviceImpl.getBlockByNumber((bds.evm.Rpc.GetBlockByNumberRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBlockResponse>) responseObserver);
          break;
        case METHODID_GET_BLOCK_BY_HASH:
          serviceImpl.getBlockByHash((bds.evm.Rpc.GetBlockByHashRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBlockResponse>) responseObserver);
          break;
        case METHODID_GET_LOGS:
          serviceImpl.getLogs((bds.evm.Rpc.GetLogsRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Rpc.GetLogsResponse>) responseObserver);
          break;
        case METHODID_GET_TRANSACTION_BY_HASH:
          serviceImpl.getTransactionByHash((bds.evm.Rpc.GetTransactionByHashRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Rpc.GetTransactionByHashResponse>) responseObserver);
          break;
        case METHODID_GET_TRANSACTION_RECEIPT:
          serviceImpl.getTransactionReceipt((bds.evm.Rpc.GetTransactionReceiptRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Rpc.GetTransactionReceiptResponse>) responseObserver);
          break;
        case METHODID_GET_BLOCK_RECEIPTS:
          serviceImpl.getBlockReceipts((bds.evm.Rpc.GetBlockReceiptsRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBlockReceiptsResponse>) responseObserver);
          break;
        case METHODID_TRACE_BLOCK:
          serviceImpl.traceBlock((bds.evm.Rpc.TraceBlockRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Rpc.TraceResponse>) responseObserver);
          break;
        case METHODID_TRACE_TRANSACTION:
          serviceImpl.traceTransaction((bds.evm.Rpc.TraceTransactionRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Rpc.TraceResponse>) responseObserver);
          break;
        case METHODID_PARITY_TRACE_BLOCK:
          serviceImpl.parityTraceBlock((bds.evm.Rpc.ParityTraceBlockRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Rpc.TraceResponse>) responseObserver);
          break;
        case METHODID_GET_BALANCE:
          serviceImpl.getBalance((bds.evm.Rpc.GetBalanceRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Rpc.GetBalanceResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getChainIdMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Rpc.ChainIdRequest,
              bds.evm.Rpc.ChainIdResponse>(
                service, METHODID_CHAIN_ID)))
        .addMethod(
          getGetBlockByNumberMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Rpc.GetBlockByNumberRequest,
              bds.evm.Rpc.GetBlockResponse>(
                service, METHODID_GET_BLOCK_BY_NUMBER)))
        .addMethod(
          getGetBlockByHashMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Rpc.GetBlockByHashRequest,
              bds.evm.Rpc.GetBlockResponse>(
                service, METHODID_GET_BLOCK_BY_HASH)))
        .addMethod(
          getGetLogsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Rpc.GetLogsRequest,
              bds.evm.Rpc.GetLogsResponse>(
                service, METHODID_GET_LOGS)))
        .addMethod(
          getGetTransactionByHashMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Rpc.GetTransactionByHashRequest,
              bds.evm.Rpc.GetTransactionByHashResponse>(
                service, METHODID_GET_TRANSACTION_BY_HASH)))
        .addMethod(
          getGetTransactionReceiptMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Rpc.GetTransactionReceiptRequest,
              bds.evm.Rpc.GetTransactionReceiptResponse>(
                service, METHODID_GET_TRANSACTION_RECEIPT)))
        .addMethod(
          getGetBlockReceiptsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Rpc.GetBlockReceiptsRequest,
              bds.evm.Rpc.GetBlockReceiptsResponse>(
                service, METHODID_GET_BLOCK_RECEIPTS)))
        .addMethod(
          getTraceBlockMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Rpc.TraceBlockRequest,
              bds.evm.Rpc.TraceResponse>(
                service, METHODID_TRACE_BLOCK)))
        .addMethod(
          getTraceTransactionMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Rpc.TraceTransactionRequest,
              bds.evm.Rpc.TraceResponse>(
                service, METHODID_TRACE_TRANSACTION)))
        .addMethod(
          getParityTraceBlockMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Rpc.ParityTraceBlockRequest,
              bds.evm.Rpc.TraceResponse>(
                service, METHODID_PARITY_TRACE_BLOCK)))
        .addMethod(
          getGetBalanceMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Rpc.GetBalanceRequest,
              bds.evm.Rpc.GetBalanceResponse>(
                service, METHODID_GET_BALANCE)))
        .build();
  }

  private static abstract class RPCQueryServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    RPCQueryServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return bds.evm.Rpc.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("RPCQueryService");
    }
  }

  private static final class RPCQueryServiceFileDescriptorSupplier
      extends RPCQueryServiceBaseDescriptorSupplier {
    RPCQueryServiceFileDescriptorSupplier() {}
  }

  private static final class RPCQueryServiceMethodDescriptorSupplier
      extends RPCQueryServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    RPCQueryServiceMethodDescriptorSupplier(java.lang.String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (RPCQueryServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new RPCQueryServiceFileDescriptorSupplier())
              .addMethod(getChainIdMethod())
              .addMethod(getGetBlockByNumberMethod())
              .addMethod(getGetBlockByHashMethod())
              .addMethod(getGetLogsMethod())
              .addMethod(getGetTransactionByHashMethod())
              .addMethod(getGetTransactionReceiptMethod())
              .addMethod(getGetBlockReceiptsMethod())
              .addMethod(getTraceBlockMethod())
              .addMethod(getTraceTransactionMethod())
              .addMethod(getParityTraceBlockMethod())
              .addMethod(getGetBalanceMethod())
              .build();
        }
      }
    }
    return result;
  }
}
