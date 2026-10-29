from . import models_pb2 as _models_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ChainIdRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ChainIdResponse(_message.Message):
    __slots__ = ("chainId", "genesisHash")
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    GENESISHASH_FIELD_NUMBER: _ClassVar[int]
    chainId: int
    genesisHash: bytes
    def __init__(self, chainId: _Optional[int] = ..., genesisHash: _Optional[bytes] = ...) -> None: ...

class GetBlockByNumberRequest(_message.Message):
    __slots__ = ("blockNumber", "includeTransactions", "chainId", "chainGenesisHash")
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    INCLUDETRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    blockNumber: str
    includeTransactions: bool
    chainId: int
    chainGenesisHash: bytes
    def __init__(self, blockNumber: _Optional[str] = ..., includeTransactions: _Optional[bool] = ..., chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ...) -> None: ...

class GetBlockByHashRequest(_message.Message):
    __slots__ = ("blockHash", "includeTransactions", "chainId", "chainGenesisHash")
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    INCLUDETRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    blockHash: bytes
    includeTransactions: bool
    chainId: int
    chainGenesisHash: bytes
    def __init__(self, blockHash: _Optional[bytes] = ..., includeTransactions: _Optional[bool] = ..., chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ...) -> None: ...

class GetBlockResponse(_message.Message):
    __slots__ = ("block", "transactions", "chainId", "chainGenesisHash", "fullTransactions", "withdrawals")
    BLOCK_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    FULLTRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    WITHDRAWALS_FIELD_NUMBER: _ClassVar[int]
    block: _models_pb2.BlockHeader
    transactions: _containers.RepeatedScalarFieldContainer[bytes]
    chainId: int
    chainGenesisHash: bytes
    fullTransactions: _containers.RepeatedCompositeFieldContainer[_models_pb2.Transaction]
    withdrawals: _containers.RepeatedCompositeFieldContainer[_models_pb2.Withdrawal]
    def __init__(self, block: _Optional[_Union[_models_pb2.BlockHeader, _Mapping]] = ..., transactions: _Optional[_Iterable[bytes]] = ..., chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ..., fullTransactions: _Optional[_Iterable[_Union[_models_pb2.Transaction, _Mapping]]] = ..., withdrawals: _Optional[_Iterable[_Union[_models_pb2.Withdrawal, _Mapping]]] = ...) -> None: ...

class GetLogsRequest(_message.Message):
    __slots__ = ("fromBlock", "toBlock", "addresses", "topics", "blockHash", "chainId", "chainGenesisHash")
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    ADDRESSES_FIELD_NUMBER: _ClassVar[int]
    TOPICS_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    fromBlock: int
    toBlock: int
    addresses: _containers.RepeatedScalarFieldContainer[bytes]
    topics: _containers.RepeatedCompositeFieldContainer[TopicFilter]
    blockHash: bytes
    chainId: int
    chainGenesisHash: bytes
    def __init__(self, fromBlock: _Optional[int] = ..., toBlock: _Optional[int] = ..., addresses: _Optional[_Iterable[bytes]] = ..., topics: _Optional[_Iterable[_Union[TopicFilter, _Mapping]]] = ..., blockHash: _Optional[bytes] = ..., chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ...) -> None: ...

class TopicFilter(_message.Message):
    __slots__ = ("values",)
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.RepeatedScalarFieldContainer[bytes]
    def __init__(self, values: _Optional[_Iterable[bytes]] = ...) -> None: ...

class GetLogsResponse(_message.Message):
    __slots__ = ("logs",)
    LOGS_FIELD_NUMBER: _ClassVar[int]
    logs: _containers.RepeatedCompositeFieldContainer[_models_pb2.Log]
    def __init__(self, logs: _Optional[_Iterable[_Union[_models_pb2.Log, _Mapping]]] = ...) -> None: ...

class GetTransactionByHashRequest(_message.Message):
    __slots__ = ("transactionHash", "chainId", "chainGenesisHash")
    TRANSACTIONHASH_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    transactionHash: bytes
    chainId: int
    chainGenesisHash: bytes
    def __init__(self, transactionHash: _Optional[bytes] = ..., chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ...) -> None: ...

class GetTransactionByHashResponse(_message.Message):
    __slots__ = ("transaction",)
    TRANSACTION_FIELD_NUMBER: _ClassVar[int]
    transaction: _models_pb2.Transaction
    def __init__(self, transaction: _Optional[_Union[_models_pb2.Transaction, _Mapping]] = ...) -> None: ...

class GetTransactionReceiptRequest(_message.Message):
    __slots__ = ("transactionHash", "chainId", "chainGenesisHash")
    TRANSACTIONHASH_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    transactionHash: bytes
    chainId: int
    chainGenesisHash: bytes
    def __init__(self, transactionHash: _Optional[bytes] = ..., chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ...) -> None: ...

class GetTransactionReceiptResponse(_message.Message):
    __slots__ = ("receipt",)
    RECEIPT_FIELD_NUMBER: _ClassVar[int]
    receipt: _models_pb2.Receipt
    def __init__(self, receipt: _Optional[_Union[_models_pb2.Receipt, _Mapping]] = ...) -> None: ...

class GetBlockReceiptsRequest(_message.Message):
    __slots__ = ("blockNumber", "blockHash", "chainId", "chainGenesisHash")
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    blockNumber: str
    blockHash: bytes
    chainId: int
    chainGenesisHash: bytes
    def __init__(self, blockNumber: _Optional[str] = ..., blockHash: _Optional[bytes] = ..., chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ...) -> None: ...

class GetBlockReceiptsResponse(_message.Message):
    __slots__ = ("receipts",)
    RECEIPTS_FIELD_NUMBER: _ClassVar[int]
    receipts: _containers.RepeatedCompositeFieldContainer[_models_pb2.Receipt]
    def __init__(self, receipts: _Optional[_Iterable[_Union[_models_pb2.Receipt, _Mapping]]] = ...) -> None: ...

class TraceOptions(_message.Message):
    __slots__ = ("tracer", "tracerConfig", "options")
    TRACER_FIELD_NUMBER: _ClassVar[int]
    TRACERCONFIG_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    tracer: str
    tracerConfig: bytes
    options: bytes
    def __init__(self, tracer: _Optional[str] = ..., tracerConfig: _Optional[bytes] = ..., options: _Optional[bytes] = ...) -> None: ...

class TraceBlockRequest(_message.Message):
    __slots__ = ("blockNumber", "blockHash", "trace", "chainId", "chainGenesisHash")
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    TRACE_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    blockNumber: str
    blockHash: bytes
    trace: TraceOptions
    chainId: int
    chainGenesisHash: bytes
    def __init__(self, blockNumber: _Optional[str] = ..., blockHash: _Optional[bytes] = ..., trace: _Optional[_Union[TraceOptions, _Mapping]] = ..., chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ...) -> None: ...

class TraceTransactionRequest(_message.Message):
    __slots__ = ("transactionHash", "trace", "chainId", "chainGenesisHash")
    TRANSACTIONHASH_FIELD_NUMBER: _ClassVar[int]
    TRACE_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    transactionHash: bytes
    trace: TraceOptions
    chainId: int
    chainGenesisHash: bytes
    def __init__(self, transactionHash: _Optional[bytes] = ..., trace: _Optional[_Union[TraceOptions, _Mapping]] = ..., chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ...) -> None: ...

class ParityTraceBlockRequest(_message.Message):
    __slots__ = ("blockNumber", "blockHash", "chainId", "chainGenesisHash")
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    blockNumber: str
    blockHash: bytes
    chainId: int
    chainGenesisHash: bytes
    def __init__(self, blockNumber: _Optional[str] = ..., blockHash: _Optional[bytes] = ..., chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ...) -> None: ...

class TraceResponse(_message.Message):
    __slots__ = ("result",)
    RESULT_FIELD_NUMBER: _ClassVar[int]
    result: bytes
    def __init__(self, result: _Optional[bytes] = ...) -> None: ...

class GetBalanceRequest(_message.Message):
    __slots__ = ("address", "blockNumber", "blockHash", "chainId", "chainGenesisHash")
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    address: bytes
    blockNumber: str
    blockHash: bytes
    chainId: int
    chainGenesisHash: bytes
    def __init__(self, address: _Optional[bytes] = ..., blockNumber: _Optional[str] = ..., blockHash: _Optional[bytes] = ..., chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ...) -> None: ...

class GetBalanceResponse(_message.Message):
    __slots__ = ("balance",)
    BALANCE_FIELD_NUMBER: _ClassVar[int]
    balance: bytes
    def __init__(self, balance: _Optional[bytes] = ...) -> None: ...
