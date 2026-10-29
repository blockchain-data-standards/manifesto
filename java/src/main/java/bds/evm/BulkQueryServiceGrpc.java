package bds.evm;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Service for bulk EVM blockchain data operations
 * Designed for high-throughput data extraction and analytics use cases
 * </pre>
 */
@io.grpc.stub.annotations.GrpcGenerated
public final class BulkQueryServiceGrpc {

  private BulkQueryServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "bds.evm.BulkQueryService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<bds.evm.Bulk.GetBlocksByRangeRequest,
      bds.evm.Bulk.GetBlocksByRangeResponse> getGetBlocksByRangeMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetBlocksByRange",
      requestType = bds.evm.Bulk.GetBlocksByRangeRequest.class,
      responseType = bds.evm.Bulk.GetBlocksByRangeResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<bds.evm.Bulk.GetBlocksByRangeRequest,
      bds.evm.Bulk.GetBlocksByRangeResponse> getGetBlocksByRangeMethod() {
    io.grpc.MethodDescriptor<bds.evm.Bulk.GetBlocksByRangeRequest, bds.evm.Bulk.GetBlocksByRangeResponse> getGetBlocksByRangeMethod;
    if ((getGetBlocksByRangeMethod = BulkQueryServiceGrpc.getGetBlocksByRangeMethod) == null) {
      synchronized (BulkQueryServiceGrpc.class) {
        if ((getGetBlocksByRangeMethod = BulkQueryServiceGrpc.getGetBlocksByRangeMethod) == null) {
          BulkQueryServiceGrpc.getGetBlocksByRangeMethod = getGetBlocksByRangeMethod =
              io.grpc.MethodDescriptor.<bds.evm.Bulk.GetBlocksByRangeRequest, bds.evm.Bulk.GetBlocksByRangeResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetBlocksByRange"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Bulk.GetBlocksByRangeRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Bulk.GetBlocksByRangeResponse.getDefaultInstance()))
              .setSchemaDescriptor(new BulkQueryServiceMethodDescriptorSupplier("GetBlocksByRange"))
              .build();
        }
      }
    }
    return getGetBlocksByRangeMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static BulkQueryServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<BulkQueryServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<BulkQueryServiceStub>() {
        @java.lang.Override
        public BulkQueryServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new BulkQueryServiceStub(channel, callOptions);
        }
      };
    return BulkQueryServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports all types of calls on the service
   */
  public static BulkQueryServiceBlockingV2Stub newBlockingV2Stub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<BulkQueryServiceBlockingV2Stub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<BulkQueryServiceBlockingV2Stub>() {
        @java.lang.Override
        public BulkQueryServiceBlockingV2Stub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new BulkQueryServiceBlockingV2Stub(channel, callOptions);
        }
      };
    return BulkQueryServiceBlockingV2Stub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static BulkQueryServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<BulkQueryServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<BulkQueryServiceBlockingStub>() {
        @java.lang.Override
        public BulkQueryServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new BulkQueryServiceBlockingStub(channel, callOptions);
        }
      };
    return BulkQueryServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static BulkQueryServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<BulkQueryServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<BulkQueryServiceFutureStub>() {
        @java.lang.Override
        public BulkQueryServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new BulkQueryServiceFutureStub(channel, callOptions);
        }
      };
    return BulkQueryServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * Service for bulk EVM blockchain data operations
   * Designed for high-throughput data extraction and analytics use cases
   * </pre>
   */
  public interface AsyncService {

    /**
     * <pre>
     * Get multiple blocks within a range in a single request
     * </pre>
     */
    default void getBlocksByRange(bds.evm.Bulk.GetBlocksByRangeRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Bulk.GetBlocksByRangeResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetBlocksByRangeMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service BulkQueryService.
   * <pre>
   * Service for bulk EVM blockchain data operations
   * Designed for high-throughput data extraction and analytics use cases
   * </pre>
   */
  public static abstract class BulkQueryServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return BulkQueryServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service BulkQueryService.
   * <pre>
   * Service for bulk EVM blockchain data operations
   * Designed for high-throughput data extraction and analytics use cases
   * </pre>
   */
  public static final class BulkQueryServiceStub
      extends io.grpc.stub.AbstractAsyncStub<BulkQueryServiceStub> {
    private BulkQueryServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected BulkQueryServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new BulkQueryServiceStub(channel, callOptions);
    }

    /**
     * <pre>
     * Get multiple blocks within a range in a single request
     * </pre>
     */
    public void getBlocksByRange(bds.evm.Bulk.GetBlocksByRangeRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Bulk.GetBlocksByRangeResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetBlocksByRangeMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service BulkQueryService.
   * <pre>
   * Service for bulk EVM blockchain data operations
   * Designed for high-throughput data extraction and analytics use cases
   * </pre>
   */
  public static final class BulkQueryServiceBlockingV2Stub
      extends io.grpc.stub.AbstractBlockingStub<BulkQueryServiceBlockingV2Stub> {
    private BulkQueryServiceBlockingV2Stub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected BulkQueryServiceBlockingV2Stub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new BulkQueryServiceBlockingV2Stub(channel, callOptions);
    }

    /**
     * <pre>
     * Get multiple blocks within a range in a single request
     * </pre>
     */
    public bds.evm.Bulk.GetBlocksByRangeResponse getBlocksByRange(bds.evm.Bulk.GetBlocksByRangeRequest request) throws io.grpc.StatusException {
      return io.grpc.stub.ClientCalls.blockingV2UnaryCall(
          getChannel(), getGetBlocksByRangeMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do limited synchronous rpc calls to service BulkQueryService.
   * <pre>
   * Service for bulk EVM blockchain data operations
   * Designed for high-throughput data extraction and analytics use cases
   * </pre>
   */
  public static final class BulkQueryServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<BulkQueryServiceBlockingStub> {
    private BulkQueryServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected BulkQueryServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new BulkQueryServiceBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * Get multiple blocks within a range in a single request
     * </pre>
     */
    public bds.evm.Bulk.GetBlocksByRangeResponse getBlocksByRange(bds.evm.Bulk.GetBlocksByRangeRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetBlocksByRangeMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service BulkQueryService.
   * <pre>
   * Service for bulk EVM blockchain data operations
   * Designed for high-throughput data extraction and analytics use cases
   * </pre>
   */
  public static final class BulkQueryServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<BulkQueryServiceFutureStub> {
    private BulkQueryServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected BulkQueryServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new BulkQueryServiceFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * Get multiple blocks within a range in a single request
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<bds.evm.Bulk.GetBlocksByRangeResponse> getBlocksByRange(
        bds.evm.Bulk.GetBlocksByRangeRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetBlocksByRangeMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_GET_BLOCKS_BY_RANGE = 0;

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
        case METHODID_GET_BLOCKS_BY_RANGE:
          serviceImpl.getBlocksByRange((bds.evm.Bulk.GetBlocksByRangeRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Bulk.GetBlocksByRangeResponse>) responseObserver);
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
          getGetBlocksByRangeMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              bds.evm.Bulk.GetBlocksByRangeRequest,
              bds.evm.Bulk.GetBlocksByRangeResponse>(
                service, METHODID_GET_BLOCKS_BY_RANGE)))
        .build();
  }

  private static abstract class BulkQueryServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    BulkQueryServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return bds.evm.Bulk.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("BulkQueryService");
    }
  }

  private static final class BulkQueryServiceFileDescriptorSupplier
      extends BulkQueryServiceBaseDescriptorSupplier {
    BulkQueryServiceFileDescriptorSupplier() {}
  }

  private static final class BulkQueryServiceMethodDescriptorSupplier
      extends BulkQueryServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    BulkQueryServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (BulkQueryServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new BulkQueryServiceFileDescriptorSupplier())
              .addMethod(getGetBlocksByRangeMethod())
              .build();
        }
      }
    }
    return result;
  }
}
