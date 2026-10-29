package bds.evm;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Service for streaming subscriptions over chain data.
 * Unlike the request/response RPCQueryService and the paginated QueryService,
 * StreamService opens a long-lived server stream that pushes one message per new
 * item as the chain advances. A provider feeds it from whatever head source it
 * has (a state poller, a gap-free chain view, …); the wire contract is the same
 * everywhere.
 * Status codes that end a stream:
 *   OUT_OF_RANGE  the server cannot continue from the resume position (too
 *                 old, or a block it cannot place); backfill with unary
 *                 calls, then reopen.
 *   UNAVAILABLE   the server stopped making progress, or the client read too
 *                 slowly; reopen with a resume position.
 * </pre>
 */
@io.grpc.stub.annotations.GrpcGenerated
public final class StreamServiceGrpc {

  private StreamServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "bds.evm.StreamService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<bds.evm.Stream.StreamBlocksRequest,
      bds.evm.Stream.StreamBlocksResponse> getStreamBlocksMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "StreamBlocks",
      requestType = bds.evm.Stream.StreamBlocksRequest.class,
      responseType = bds.evm.Stream.StreamBlocksResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<bds.evm.Stream.StreamBlocksRequest,
      bds.evm.Stream.StreamBlocksResponse> getStreamBlocksMethod() {
    io.grpc.MethodDescriptor<bds.evm.Stream.StreamBlocksRequest, bds.evm.Stream.StreamBlocksResponse> getStreamBlocksMethod;
    if ((getStreamBlocksMethod = StreamServiceGrpc.getStreamBlocksMethod) == null) {
      synchronized (StreamServiceGrpc.class) {
        if ((getStreamBlocksMethod = StreamServiceGrpc.getStreamBlocksMethod) == null) {
          StreamServiceGrpc.getStreamBlocksMethod = getStreamBlocksMethod =
              io.grpc.MethodDescriptor.<bds.evm.Stream.StreamBlocksRequest, bds.evm.Stream.StreamBlocksResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "StreamBlocks"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Stream.StreamBlocksRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Stream.StreamBlocksResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StreamServiceMethodDescriptorSupplier("StreamBlocks"))
              .build();
        }
      }
    }
    return getStreamBlocksMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Stream.StreamLogsRequest,
      bds.evm.Stream.StreamLogsResponse> getStreamLogsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "StreamLogs",
      requestType = bds.evm.Stream.StreamLogsRequest.class,
      responseType = bds.evm.Stream.StreamLogsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<bds.evm.Stream.StreamLogsRequest,
      bds.evm.Stream.StreamLogsResponse> getStreamLogsMethod() {
    io.grpc.MethodDescriptor<bds.evm.Stream.StreamLogsRequest, bds.evm.Stream.StreamLogsResponse> getStreamLogsMethod;
    if ((getStreamLogsMethod = StreamServiceGrpc.getStreamLogsMethod) == null) {
      synchronized (StreamServiceGrpc.class) {
        if ((getStreamLogsMethod = StreamServiceGrpc.getStreamLogsMethod) == null) {
          StreamServiceGrpc.getStreamLogsMethod = getStreamLogsMethod =
              io.grpc.MethodDescriptor.<bds.evm.Stream.StreamLogsRequest, bds.evm.Stream.StreamLogsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "StreamLogs"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Stream.StreamLogsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Stream.StreamLogsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StreamServiceMethodDescriptorSupplier("StreamLogs"))
              .build();
        }
      }
    }
    return getStreamLogsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<bds.evm.Stream.StreamPendingTransactionsRequest,
      bds.evm.Stream.StreamPendingTransactionsResponse> getStreamPendingTransactionsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "StreamPendingTransactions",
      requestType = bds.evm.Stream.StreamPendingTransactionsRequest.class,
      responseType = bds.evm.Stream.StreamPendingTransactionsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<bds.evm.Stream.StreamPendingTransactionsRequest,
      bds.evm.Stream.StreamPendingTransactionsResponse> getStreamPendingTransactionsMethod() {
    io.grpc.MethodDescriptor<bds.evm.Stream.StreamPendingTransactionsRequest, bds.evm.Stream.StreamPendingTransactionsResponse> getStreamPendingTransactionsMethod;
    if ((getStreamPendingTransactionsMethod = StreamServiceGrpc.getStreamPendingTransactionsMethod) == null) {
      synchronized (StreamServiceGrpc.class) {
        if ((getStreamPendingTransactionsMethod = StreamServiceGrpc.getStreamPendingTransactionsMethod) == null) {
          StreamServiceGrpc.getStreamPendingTransactionsMethod = getStreamPendingTransactionsMethod =
              io.grpc.MethodDescriptor.<bds.evm.Stream.StreamPendingTransactionsRequest, bds.evm.Stream.StreamPendingTransactionsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "StreamPendingTransactions"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Stream.StreamPendingTransactionsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  bds.evm.Stream.StreamPendingTransactionsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StreamServiceMethodDescriptorSupplier("StreamPendingTransactions"))
              .build();
        }
      }
    }
    return getStreamPendingTransactionsMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static StreamServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<StreamServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<StreamServiceStub>() {
        @java.lang.Override
        public StreamServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new StreamServiceStub(channel, callOptions);
        }
      };
    return StreamServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports all types of calls on the service
   */
  public static StreamServiceBlockingV2Stub newBlockingV2Stub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<StreamServiceBlockingV2Stub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<StreamServiceBlockingV2Stub>() {
        @java.lang.Override
        public StreamServiceBlockingV2Stub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new StreamServiceBlockingV2Stub(channel, callOptions);
        }
      };
    return StreamServiceBlockingV2Stub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static StreamServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<StreamServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<StreamServiceBlockingStub>() {
        @java.lang.Override
        public StreamServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new StreamServiceBlockingStub(channel, callOptions);
        }
      };
    return StreamServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static StreamServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<StreamServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<StreamServiceFutureStub>() {
        @java.lang.Override
        public StreamServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new StreamServiceFutureStub(channel, callOptions);
        }
      };
    return StreamServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * Service for streaming subscriptions over chain data.
   * Unlike the request/response RPCQueryService and the paginated QueryService,
   * StreamService opens a long-lived server stream that pushes one message per new
   * item as the chain advances. A provider feeds it from whatever head source it
   * has (a state poller, a gap-free chain view, …); the wire contract is the same
   * everywhere.
   * Status codes that end a stream:
   *   OUT_OF_RANGE  the server cannot continue from the resume position (too
   *                 old, or a block it cannot place); backfill with unary
   *                 calls, then reopen.
   *   UNAVAILABLE   the server stopped making progress, or the client read too
   *                 slowly; reopen with a resume position.
   * </pre>
   */
  public interface AsyncService {

    /**
     * <pre>
     * Stream block headers as the chain head advances: one header message per
     * canonical block, ascending, with no number skipped, until the client
     * cancels the stream. After a reorg the server restarts at the fork point: it
     * resends from the first block of the new branch. Client rule: a header whose
     * number is at or below the last one received replaces that number and every
     * number above it. Each header carries hash and parentHash, so a client can
     * check linkage itself.
     * </pre>
     */
    default void streamBlocks(bds.evm.Stream.StreamBlocksRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Stream.StreamBlocksResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getStreamBlocksMethod(), responseObserver);
    }

    /**
     * <pre>
     * Stream the logs matching a filter, one block at a time, ascending. A
     * block's logs arrive in one or more BlockLogs messages; the last sets
     * complete. A reorg sends Undo with the fork point, then the new branch
     * from the fork point + 1 (REORG_UNDO).
     * </pre>
     */
    default void streamLogs(bds.evm.Stream.StreamLogsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Stream.StreamLogsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getStreamLogsMethod(), responseObserver);
    }

    /**
     * <pre>
     * Stream transactions as they enter the provider's mempool view.
     * </pre>
     */
    default void streamPendingTransactions(bds.evm.Stream.StreamPendingTransactionsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Stream.StreamPendingTransactionsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getStreamPendingTransactionsMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service StreamService.
   * <pre>
   * Service for streaming subscriptions over chain data.
   * Unlike the request/response RPCQueryService and the paginated QueryService,
   * StreamService opens a long-lived server stream that pushes one message per new
   * item as the chain advances. A provider feeds it from whatever head source it
   * has (a state poller, a gap-free chain view, …); the wire contract is the same
   * everywhere.
   * Status codes that end a stream:
   *   OUT_OF_RANGE  the server cannot continue from the resume position (too
   *                 old, or a block it cannot place); backfill with unary
   *                 calls, then reopen.
   *   UNAVAILABLE   the server stopped making progress, or the client read too
   *                 slowly; reopen with a resume position.
   * </pre>
   */
  public static abstract class StreamServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return StreamServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service StreamService.
   * <pre>
   * Service for streaming subscriptions over chain data.
   * Unlike the request/response RPCQueryService and the paginated QueryService,
   * StreamService opens a long-lived server stream that pushes one message per new
   * item as the chain advances. A provider feeds it from whatever head source it
   * has (a state poller, a gap-free chain view, …); the wire contract is the same
   * everywhere.
   * Status codes that end a stream:
   *   OUT_OF_RANGE  the server cannot continue from the resume position (too
   *                 old, or a block it cannot place); backfill with unary
   *                 calls, then reopen.
   *   UNAVAILABLE   the server stopped making progress, or the client read too
   *                 slowly; reopen with a resume position.
   * </pre>
   */
  public static final class StreamServiceStub
      extends io.grpc.stub.AbstractAsyncStub<StreamServiceStub> {
    private StreamServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected StreamServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new StreamServiceStub(channel, callOptions);
    }

    /**
     * <pre>
     * Stream block headers as the chain head advances: one header message per
     * canonical block, ascending, with no number skipped, until the client
     * cancels the stream. After a reorg the server restarts at the fork point: it
     * resends from the first block of the new branch. Client rule: a header whose
     * number is at or below the last one received replaces that number and every
     * number above it. Each header carries hash and parentHash, so a client can
     * check linkage itself.
     * </pre>
     */
    public void streamBlocks(bds.evm.Stream.StreamBlocksRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Stream.StreamBlocksResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getStreamBlocksMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Stream the logs matching a filter, one block at a time, ascending. A
     * block's logs arrive in one or more BlockLogs messages; the last sets
     * complete. A reorg sends Undo with the fork point, then the new branch
     * from the fork point + 1 (REORG_UNDO).
     * </pre>
     */
    public void streamLogs(bds.evm.Stream.StreamLogsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Stream.StreamLogsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getStreamLogsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Stream transactions as they enter the provider's mempool view.
     * </pre>
     */
    public void streamPendingTransactions(bds.evm.Stream.StreamPendingTransactionsRequest request,
        io.grpc.stub.StreamObserver<bds.evm.Stream.StreamPendingTransactionsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getStreamPendingTransactionsMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service StreamService.
   * <pre>
   * Service for streaming subscriptions over chain data.
   * Unlike the request/response RPCQueryService and the paginated QueryService,
   * StreamService opens a long-lived server stream that pushes one message per new
   * item as the chain advances. A provider feeds it from whatever head source it
   * has (a state poller, a gap-free chain view, …); the wire contract is the same
   * everywhere.
   * Status codes that end a stream:
   *   OUT_OF_RANGE  the server cannot continue from the resume position (too
   *                 old, or a block it cannot place); backfill with unary
   *                 calls, then reopen.
   *   UNAVAILABLE   the server stopped making progress, or the client read too
   *                 slowly; reopen with a resume position.
   * </pre>
   */
  public static final class StreamServiceBlockingV2Stub
      extends io.grpc.stub.AbstractBlockingStub<StreamServiceBlockingV2Stub> {
    private StreamServiceBlockingV2Stub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected StreamServiceBlockingV2Stub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new StreamServiceBlockingV2Stub(channel, callOptions);
    }

    /**
     * <pre>
     * Stream block headers as the chain head advances: one header message per
     * canonical block, ascending, with no number skipped, until the client
     * cancels the stream. After a reorg the server restarts at the fork point: it
     * resends from the first block of the new branch. Client rule: a header whose
     * number is at or below the last one received replaces that number and every
     * number above it. Each header carries hash and parentHash, so a client can
     * check linkage itself.
     * </pre>
     */
    @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/10918")
    public io.grpc.stub.BlockingClientCall<?, bds.evm.Stream.StreamBlocksResponse>
        streamBlocks(bds.evm.Stream.StreamBlocksRequest request) {
      return io.grpc.stub.ClientCalls.blockingV2ServerStreamingCall(
          getChannel(), getStreamBlocksMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Stream the logs matching a filter, one block at a time, ascending. A
     * block's logs arrive in one or more BlockLogs messages; the last sets
     * complete. A reorg sends Undo with the fork point, then the new branch
     * from the fork point + 1 (REORG_UNDO).
     * </pre>
     */
    @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/10918")
    public io.grpc.stub.BlockingClientCall<?, bds.evm.Stream.StreamLogsResponse>
        streamLogs(bds.evm.Stream.StreamLogsRequest request) {
      return io.grpc.stub.ClientCalls.blockingV2ServerStreamingCall(
          getChannel(), getStreamLogsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Stream transactions as they enter the provider's mempool view.
     * </pre>
     */
    @io.grpc.ExperimentalApi("https://github.com/grpc/grpc-java/issues/10918")
    public io.grpc.stub.BlockingClientCall<?, bds.evm.Stream.StreamPendingTransactionsResponse>
        streamPendingTransactions(bds.evm.Stream.StreamPendingTransactionsRequest request) {
      return io.grpc.stub.ClientCalls.blockingV2ServerStreamingCall(
          getChannel(), getStreamPendingTransactionsMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do limited synchronous rpc calls to service StreamService.
   * <pre>
   * Service for streaming subscriptions over chain data.
   * Unlike the request/response RPCQueryService and the paginated QueryService,
   * StreamService opens a long-lived server stream that pushes one message per new
   * item as the chain advances. A provider feeds it from whatever head source it
   * has (a state poller, a gap-free chain view, …); the wire contract is the same
   * everywhere.
   * Status codes that end a stream:
   *   OUT_OF_RANGE  the server cannot continue from the resume position (too
   *                 old, or a block it cannot place); backfill with unary
   *                 calls, then reopen.
   *   UNAVAILABLE   the server stopped making progress, or the client read too
   *                 slowly; reopen with a resume position.
   * </pre>
   */
  public static final class StreamServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<StreamServiceBlockingStub> {
    private StreamServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected StreamServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new StreamServiceBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * Stream block headers as the chain head advances: one header message per
     * canonical block, ascending, with no number skipped, until the client
     * cancels the stream. After a reorg the server restarts at the fork point: it
     * resends from the first block of the new branch. Client rule: a header whose
     * number is at or below the last one received replaces that number and every
     * number above it. Each header carries hash and parentHash, so a client can
     * check linkage itself.
     * </pre>
     */
    public java.util.Iterator<bds.evm.Stream.StreamBlocksResponse> streamBlocks(
        bds.evm.Stream.StreamBlocksRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getStreamBlocksMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Stream the logs matching a filter, one block at a time, ascending. A
     * block's logs arrive in one or more BlockLogs messages; the last sets
     * complete. A reorg sends Undo with the fork point, then the new branch
     * from the fork point + 1 (REORG_UNDO).
     * </pre>
     */
    public java.util.Iterator<bds.evm.Stream.StreamLogsResponse> streamLogs(
        bds.evm.Stream.StreamLogsRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getStreamLogsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Stream transactions as they enter the provider's mempool view.
     * </pre>
     */
    public java.util.Iterator<bds.evm.Stream.StreamPendingTransactionsResponse> streamPendingTransactions(
        bds.evm.Stream.StreamPendingTransactionsRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getStreamPendingTransactionsMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service StreamService.
   * <pre>
   * Service for streaming subscriptions over chain data.
   * Unlike the request/response RPCQueryService and the paginated QueryService,
   * StreamService opens a long-lived server stream that pushes one message per new
   * item as the chain advances. A provider feeds it from whatever head source it
   * has (a state poller, a gap-free chain view, …); the wire contract is the same
   * everywhere.
   * Status codes that end a stream:
   *   OUT_OF_RANGE  the server cannot continue from the resume position (too
   *                 old, or a block it cannot place); backfill with unary
   *                 calls, then reopen.
   *   UNAVAILABLE   the server stopped making progress, or the client read too
   *                 slowly; reopen with a resume position.
   * </pre>
   */
  public static final class StreamServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<StreamServiceFutureStub> {
    private StreamServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected StreamServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new StreamServiceFutureStub(channel, callOptions);
    }
  }

  private static final int METHODID_STREAM_BLOCKS = 0;
  private static final int METHODID_STREAM_LOGS = 1;
  private static final int METHODID_STREAM_PENDING_TRANSACTIONS = 2;

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
        case METHODID_STREAM_BLOCKS:
          serviceImpl.streamBlocks((bds.evm.Stream.StreamBlocksRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Stream.StreamBlocksResponse>) responseObserver);
          break;
        case METHODID_STREAM_LOGS:
          serviceImpl.streamLogs((bds.evm.Stream.StreamLogsRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Stream.StreamLogsResponse>) responseObserver);
          break;
        case METHODID_STREAM_PENDING_TRANSACTIONS:
          serviceImpl.streamPendingTransactions((bds.evm.Stream.StreamPendingTransactionsRequest) request,
              (io.grpc.stub.StreamObserver<bds.evm.Stream.StreamPendingTransactionsResponse>) responseObserver);
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
          getStreamBlocksMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              bds.evm.Stream.StreamBlocksRequest,
              bds.evm.Stream.StreamBlocksResponse>(
                service, METHODID_STREAM_BLOCKS)))
        .addMethod(
          getStreamLogsMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              bds.evm.Stream.StreamLogsRequest,
              bds.evm.Stream.StreamLogsResponse>(
                service, METHODID_STREAM_LOGS)))
        .addMethod(
          getStreamPendingTransactionsMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              bds.evm.Stream.StreamPendingTransactionsRequest,
              bds.evm.Stream.StreamPendingTransactionsResponse>(
                service, METHODID_STREAM_PENDING_TRANSACTIONS)))
        .build();
  }

  private static abstract class StreamServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    StreamServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return bds.evm.Stream.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("StreamService");
    }
  }

  private static final class StreamServiceFileDescriptorSupplier
      extends StreamServiceBaseDescriptorSupplier {
    StreamServiceFileDescriptorSupplier() {}
  }

  private static final class StreamServiceMethodDescriptorSupplier
      extends StreamServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    StreamServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (StreamServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new StreamServiceFileDescriptorSupplier())
              .addMethod(getStreamBlocksMethod())
              .addMethod(getStreamLogsMethod())
              .addMethod(getStreamPendingTransactionsMethod())
              .build();
        }
      }
    }
    return result;
  }
}
