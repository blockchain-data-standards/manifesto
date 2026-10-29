package bds.evm;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@io.grpc.stub.annotations.GrpcGenerated
public final class QueryServiceGrpc {

  private QueryServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "bds.evm.QueryService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<bds.evm.Query.QueryBlocksRequest,
      bds.evm.Query.QueryBlocksResponse> getQueryBlocksMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "QueryBlocks",
      requestType = bds.evm.Query.QueryBlocksRequest.class,
      responseType = bds.evm.Query.QueryBlocksResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<bds.evm.Query.QueryBlocksRequest,
      bds.evm.Query.QueryBlocksResponse> getQueryBlocksMethod() {
    io.grpc.MethodDescriptor<bds.evm.Query.QueryBlocksRequest, bds.evm.Query.QueryBlocksResponse> getQueryBlocksMethod;
    if ((getQueryBlocksMethod = QueryServiceGrpc.getQueryBlocksMethod) == null) {
      synchronized (QueryServiceGrpc.class) {
        if ((getQueryBlocksMethod = QueryServiceGrpc.getQueryBlocksMethod) == null) {
          QueryServiceGrpc.getQueryBlocksMethod = getQueryBlocksMethod =
              io.grpc.MethodDescriptor.<bds.evm.Query.QueryBlocksRequest, bds.evm.Query.QueryBlocksResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "QueryBlocks"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Query.QueryBlocksRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Query.QueryBlocksResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryServiceMethodDescriptorSupplier("QueryBlocks"))
              .build();
        }
      }
    }
    return getQueryBlocksMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Query.QueryTransactionsRequest,
      bds.evm.Query.QueryTransactionsResponse> getQueryTransactionsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "QueryTransactions",
      requestType = bds.evm.Query.QueryTransactionsRequest.class,
      responseType = bds.evm.Query.QueryTransactionsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<bds.evm.Query.QueryTransactionsRequest,
      bds.evm.Query.QueryTransactionsResponse> getQueryTransactionsMethod() {
    io.grpc.MethodDescriptor<bds.evm.Query.QueryTransactionsRequest, bds.evm.Query.QueryTransactionsResponse> getQueryTransactionsMethod;
    if ((getQueryTransactionsMethod = QueryServiceGrpc.getQueryTransactionsMethod) == null) {
      synchronized (QueryServiceGrpc.class) {
        if ((getQueryTransactionsMethod = QueryServiceGrpc.getQueryTransactionsMethod) == null) {
          QueryServiceGrpc.getQueryTransactionsMethod = getQueryTransactionsMethod =
              io.grpc.MethodDescriptor.<bds.evm.Query.QueryTransactionsRequest, bds.evm.Query.QueryTransactionsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "QueryTransactions"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Query.QueryTransactionsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Query.QueryTransactionsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryServiceMethodDescriptorSupplier("QueryTransactions"))
              .build();
        }
      }
    }
    return getQueryTransactionsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Query.QueryLogsRequest,
      bds.evm.Query.QueryLogsResponse> getQueryLogsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "QueryLogs",
      requestType = bds.evm.Query.QueryLogsRequest.class,
      responseType = bds.evm.Query.QueryLogsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<bds.evm.Query.QueryLogsRequest,
      bds.evm.Query.QueryLogsResponse> getQueryLogsMethod() {
    io.grpc.MethodDescriptor<bds.evm.Query.QueryLogsRequest, bds.evm.Query.QueryLogsResponse> getQueryLogsMethod;
    if ((getQueryLogsMethod = QueryServiceGrpc.getQueryLogsMethod) == null) {
      synchronized (QueryServiceGrpc.class) {
        if ((getQueryLogsMethod = QueryServiceGrpc.getQueryLogsMethod) == null) {
          QueryServiceGrpc.getQueryLogsMethod = getQueryLogsMethod =
              io.grpc.MethodDescriptor.<bds.evm.Query.QueryLogsRequest, bds.evm.Query.QueryLogsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "QueryLogs"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Query.QueryLogsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Query.QueryLogsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryServiceMethodDescriptorSupplier("QueryLogs"))
              .build();
        }
      }
    }
    return getQueryLogsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Query.QueryTracesRequest,
      bds.evm.Query.QueryTracesResponse> getQueryTracesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "QueryTraces",
      requestType = bds.evm.Query.QueryTracesRequest.class,
      responseType = bds.evm.Query.QueryTracesResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<bds.evm.Query.QueryTracesRequest,
      bds.evm.Query.QueryTracesResponse> getQueryTracesMethod() {
    io.grpc.MethodDescriptor<bds.evm.Query.QueryTracesRequest, bds.evm.Query.QueryTracesResponse> getQueryTracesMethod;
    if ((getQueryTracesMethod = QueryServiceGrpc.getQueryTracesMethod) == null) {
      synchronized (QueryServiceGrpc.class) {
        if ((getQueryTracesMethod = QueryServiceGrpc.getQueryTracesMethod) == null) {
          QueryServiceGrpc.getQueryTracesMethod = getQueryTracesMethod =
              io.grpc.MethodDescriptor.<bds.evm.Query.QueryTracesRequest, bds.evm.Query.QueryTracesResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "QueryTraces"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Query.QueryTracesRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Query.QueryTracesResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryServiceMethodDescriptorSupplier("QueryTraces"))
              .build();
        }
      }
    }
    return getQueryTracesMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Query.QueryTransfersRequest,
      bds.evm.Query.QueryTransfersResponse> getQueryTransfersMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "QueryTransfers",
      requestType = bds.evm.Query.QueryTransfersRequest.class,
      responseType = bds.evm.Query.QueryTransfersResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<bds.evm.Query.QueryTransfersRequest,
      bds.evm.Query.QueryTransfersResponse> getQueryTransfersMethod() {
    io.grpc.MethodDescriptor<bds.evm.Query.QueryTransfersRequest, bds.evm.Query.QueryTransfersResponse> getQueryTransfersMethod;
    if ((getQueryTransfersMethod = QueryServiceGrpc.getQueryTransfersMethod) == null) {
      synchronized (QueryServiceGrpc.class) {
        if ((getQueryTransfersMethod = QueryServiceGrpc.getQueryTransfersMethod) == null) {
          QueryServiceGrpc.getQueryTransfersMethod = getQueryTransfersMethod =
              io.grpc.MethodDescriptor.<bds.evm.Query.QueryTransfersRequest, bds.evm.Query.QueryTransfersResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "QueryTransfers"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Query.QueryTransfersRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Query.QueryTransfersResponse.getDefaultInstance()))
              .setSchemaDescriptor(new QueryServiceMethodDescriptorSupplier("QueryTransfers"))
              .build();
        }
      }
    }
    return getQueryTransfersMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static QueryServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<QueryServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<QueryServiceStub>() {
        @java.lang.Override
        public QueryServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new QueryServiceStub(channel, callOptions);
        }
      };
    return QueryServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports all types of calls on the service
   */
  public static QueryServiceBlockingV2Stub newBlockingV2Stub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<QueryServiceBlockingV2Stub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<QueryServiceBlockingV2Stub>() {
        @java.lang.Override
        public QueryServiceBlockingV2Stub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new QueryServiceBlockingV2Stub(channel, callOptions);
        }
      };
    return QueryServiceBlockingV2Stub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static QueryServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<QueryServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<QueryServiceBlockingStub>() {
        @java.lang.Override
        public QueryServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new QueryServiceBlockingStub(channel, callOptions);
        }
      };
    return QueryServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static QueryServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<QueryServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<QueryServiceFutureStub>() {
        @java.lang.Override
        public QueryServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new QueryServiceFutureStub(channel, callOptions);
        }
      };
    return QueryServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void queryBlocks(bds.evm.Query.QueryBlocksRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Query.QueryBlocksResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getQueryBlocksMethod(), responseObserver);
    }

    /**
     */
    default void queryTransactions(bds.evm.Query.QueryTransactionsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Query.QueryTransactionsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getQueryTransactionsMethod(), responseObserver);
    }

    /**
     */
    default void queryLogs(bds.evm.Query.QueryLogsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Query.QueryLogsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getQueryLogsMethod(), responseObserver);
    }

    /**
     */
    default void queryTraces(bds.evm.Query.QueryTracesRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Query.QueryTracesResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getQueryTracesMethod(), responseObserver);
    }

    /**
     */
    default void queryTransfers(bds.evm.Query.QueryTransfersRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Query.QueryTransfersResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getQueryTransfersMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service QueryService.
   */
  public static abstract class QueryServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return QueryServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service QueryService.
   */
  public static final class QueryServiceStub
      extends io.grpc.stub.AbstractAsyncStub<QueryServiceStub> {
    private QueryServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected QueryServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new QueryServiceStub(channel, callOptions);
    }

    /**
     */
    public void queryBlocks(bds.evm.Query.QueryBlocksRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Query.QueryBlocksResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getQueryBlocksMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void queryTransactions(bds.evm.Query.QueryTransactionsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Query.QueryTransactionsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getQueryTransactionsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void queryLogs(bds.evm.Query.QueryLogsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Query.QueryLogsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getQueryLogsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void queryTraces(bds.evm.Query.QueryTracesRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Query.QueryTracesResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getQueryTracesMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void queryTransfers(bds.evm.Query.QueryTransfersRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Query.QueryTransfersResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getQueryTransfersMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service QueryService.
   */
  public static final class QueryServiceBlockingV2Stub
      extends io.grpc.stub.AbstractBlockingStub<QueryServiceBlockingV2Stub> {
    private QueryServiceBlockingV2Stub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected QueryServiceBlockingV2Stub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new QueryServiceBlockingV2Stub(channel, callOptions);
    }

    /**
     */
    @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/10918")
    public io.grpc.stub.BlockingClientCall<?, bds.evm.Query.QueryBlocksResponse>
        queryBlocks(bds.evm.Query.QueryBlocksRequest request) {
      return io.grpc.stub.ClientCalls.blockingV2ServerStreamingCall(
          getChannel(), getQueryBlocksMethod(), getCallOptions(), request);
    }

    /**
     */
    @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/10918")
    public io.grpc.stub.BlockingClientCall<?, bds.evm.Query.QueryTransactionsResponse>
        queryTransactions(bds.evm.Query.QueryTransactionsRequest request) {
      return io.grpc.stub.ClientCalls.blockingV2ServerStreamingCall(
          getChannel(), getQueryTransactionsMethod(), getCallOptions(), request);
    }

    /**
     */
    @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/10918")
    public io.grpc.stub.BlockingClientCall<?, bds.evm.Query.QueryLogsResponse>
        queryLogs(bds.evm.Query.QueryLogsRequest request) {
      return io.grpc.stub.ClientCalls.blockingV2ServerStreamingCall(
          getChannel(), getQueryLogsMethod(), getCallOptions(), request);
    }

    /**
     */
    @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/10918")
    public io.grpc.stub.BlockingClientCall<?, bds.evm.Query.QueryTracesResponse>
        queryTraces(bds.evm.Query.QueryTracesRequest request) {
      return io.grpc.stub.ClientCalls.blockingV2ServerStreamingCall(
          getChannel(), getQueryTracesMethod(), getCallOptions(), request);
    }

    /**
     */
    @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/10918")
    public io.grpc.stub.BlockingClientCall<?, bds.evm.Query.QueryTransfersResponse>
        queryTransfers(bds.evm.Query.QueryTransfersRequest request) {
      return io.grpc.stub.ClientCalls.blockingV2ServerStreamingCall(
          getChannel(), getQueryTransfersMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do limited synchronous rpc calls to service QueryService.
   */
  public static final class QueryServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<QueryServiceBlockingStub> {
    private QueryServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected QueryServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new QueryServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public java.util.Iterator<bds.evm.Query.QueryBlocksResponse> queryBlocks(
        bds.evm.Query.QueryBlocksRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getQueryBlocksMethod(), getCallOptions(), request);
    }

    /**
     */
    public java.util.Iterator<bds.evm.Query.QueryTransactionsResponse> queryTransactions(
        bds.evm.Query.QueryTransactionsRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getQueryTransactionsMethod(), getCallOptions(), request);
    }

    /**
     */
    public java.util.Iterator<bds.evm.Query.QueryLogsResponse> queryLogs(
        bds.evm.Query.QueryLogsRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getQueryLogsMethod(), getCallOptions(), request);
    }

    /**
     */
    public java.util.Iterator<bds.evm.Query.QueryTracesResponse> queryTraces(
        bds.evm.Query.QueryTracesRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getQueryTracesMethod(), getCallOptions(), request);
    }

    /**
     */
    public java.util.Iterator<bds.evm.Query.QueryTransfersResponse> queryTransfers(
        bds.evm.Query.QueryTransfersRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getQueryTransfersMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service QueryService.
   */
  public static final class QueryServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<QueryServiceFutureStub> {
    private QueryServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected QueryServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new QueryServiceFutureStub(channel, callOptions);
    }
  }

  private static final int METHODID_QUERY_BLOCKS = 0;
  private static final int METHODID_QUERY_TRANSACTIONS = 1;
  private static final int METHODID_QUERY_LOGS = 2;
  private static final int METHODID_QUERY_TRACES = 3;
  private static final int METHODID_QUERY_TRANSFERS = 4;

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
        case METHODID_QUERY_BLOCKS:
          serviceImpl.queryBlocks((bds.evm.Query.QueryBlocksRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Query.QueryBlocksResponse>) responseObserver);
          break;
        case METHODID_QUERY_TRANSACTIONS:
          serviceImpl.queryTransactions((bds.evm.Query.QueryTransactionsRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Query.QueryTransactionsResponse>) responseObserver);
          break;
        case METHODID_QUERY_LOGS:
          serviceImpl.queryLogs((bds.evm.Query.QueryLogsRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Query.QueryLogsResponse>) responseObserver);
          break;
        case METHODID_QUERY_TRACES:
          serviceImpl.queryTraces((bds.evm.Query.QueryTracesRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Query.QueryTracesResponse>) responseObserver);
          break;
        case METHODID_QUERY_TRANSFERS:
          serviceImpl.queryTransfers((bds.evm.Query.QueryTransfersRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Query.QueryTransfersResponse>) responseObserver);
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
          getQueryBlocksMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              bds.evm.Query.QueryBlocksRequest,
              bds.evm.Query.QueryBlocksResponse>(
                service, METHODID_QUERY_BLOCKS)))
        .addMethod(
          getQueryTransactionsMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              bds.evm.Query.QueryTransactionsRequest,
              bds.evm.Query.QueryTransactionsResponse>(
                service, METHODID_QUERY_TRANSACTIONS)))
        .addMethod(
          getQueryLogsMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              bds.evm.Query.QueryLogsRequest,
              bds.evm.Query.QueryLogsResponse>(
                service, METHODID_QUERY_LOGS)))
        .addMethod(
          getQueryTracesMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              bds.evm.Query.QueryTracesRequest,
              bds.evm.Query.QueryTracesResponse>(
                service, METHODID_QUERY_TRACES)))
        .addMethod(
          getQueryTransfersMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              bds.evm.Query.QueryTransfersRequest,
              bds.evm.Query.QueryTransfersResponse>(
                service, METHODID_QUERY_TRANSFERS)))
        .build();
  }

  private static abstract class QueryServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    QueryServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return bds.evm.Query.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("QueryService");
    }
  }

  private static final class QueryServiceFileDescriptorSupplier
      extends QueryServiceBaseDescriptorSupplier {
    QueryServiceFileDescriptorSupplier() {}
  }

  private static final class QueryServiceMethodDescriptorSupplier
      extends QueryServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    QueryServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (QueryServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new QueryServiceFileDescriptorSupplier())
              .addMethod(getQueryBlocksMethod())
              .addMethod(getQueryTransactionsMethod())
              .addMethod(getQueryLogsMethod())
              .addMethod(getQueryTracesMethod())
              .addMethod(getQueryTransfersMethod())
              .build();
        }
      }
    }
    return result;
  }
}
