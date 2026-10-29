from . import models_pb2 as _models_pb2
from . import rpc_pb2 as _rpc_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SortOrder(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ASC: _ClassVar[SortOrder]
    DESC: _ClassVar[SortOrder]
ASC: SortOrder
DESC: SortOrder

class CursorBlock(_message.Message):
    __slots__ = ("number", "hash", "parentHash")
    NUMBER_FIELD_NUMBER: _ClassVar[int]
    HASH_FIELD_NUMBER: _ClassVar[int]
    PARENTHASH_FIELD_NUMBER: _ClassVar[int]
    number: int
    hash: bytes
    parentHash: bytes
    def __init__(self, number: _Optional[int] = ..., hash: _Optional[bytes] = ..., parentHash: _Optional[bytes] = ...) -> None: ...

class BlockFieldSelection(_message.Message):
    __slots__ = ("number", "timestamp", "hash", "parentHash", "stateRoot", "transactionsRoot", "receiptsRoot", "logsBloom", "gasLimit", "gasUsed", "miner", "extraData", "size", "sha3Uncles", "nonce", "baseFeePerGas", "difficulty", "totalDifficulty", "mixHash", "blobGasUsed", "excessBlobGas", "withdrawalsRoot", "parentBeaconBlockRoot", "requestsHash")
    NUMBER_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    HASH_FIELD_NUMBER: _ClassVar[int]
    PARENTHASH_FIELD_NUMBER: _ClassVar[int]
    STATEROOT_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONSROOT_FIELD_NUMBER: _ClassVar[int]
    RECEIPTSROOT_FIELD_NUMBER: _ClassVar[int]
    LOGSBLOOM_FIELD_NUMBER: _ClassVar[int]
    GASLIMIT_FIELD_NUMBER: _ClassVar[int]
    GASUSED_FIELD_NUMBER: _ClassVar[int]
    MINER_FIELD_NUMBER: _ClassVar[int]
    EXTRADATA_FIELD_NUMBER: _ClassVar[int]
    SIZE_FIELD_NUMBER: _ClassVar[int]
    SHA3UNCLES_FIELD_NUMBER: _ClassVar[int]
    NONCE_FIELD_NUMBER: _ClassVar[int]
    BASEFEEPERGAS_FIELD_NUMBER: _ClassVar[int]
    DIFFICULTY_FIELD_NUMBER: _ClassVar[int]
    TOTALDIFFICULTY_FIELD_NUMBER: _ClassVar[int]
    MIXHASH_FIELD_NUMBER: _ClassVar[int]
    BLOBGASUSED_FIELD_NUMBER: _ClassVar[int]
    EXCESSBLOBGAS_FIELD_NUMBER: _ClassVar[int]
    WITHDRAWALSROOT_FIELD_NUMBER: _ClassVar[int]
    PARENTBEACONBLOCKROOT_FIELD_NUMBER: _ClassVar[int]
    REQUESTSHASH_FIELD_NUMBER: _ClassVar[int]
    number: bool
    timestamp: bool
    hash: bool
    parentHash: bool
    stateRoot: bool
    transactionsRoot: bool
    receiptsRoot: bool
    logsBloom: bool
    gasLimit: bool
    gasUsed: bool
    miner: bool
    extraData: bool
    size: bool
    sha3Uncles: bool
    nonce: bool
    baseFeePerGas: bool
    difficulty: bool
    totalDifficulty: bool
    mixHash: bool
    blobGasUsed: bool
    excessBlobGas: bool
    withdrawalsRoot: bool
    parentBeaconBlockRoot: bool
    requestsHash: bool
    def __init__(self, number: _Optional[bool] = ..., timestamp: _Optional[bool] = ..., hash: _Optional[bool] = ..., parentHash: _Optional[bool] = ..., stateRoot: _Optional[bool] = ..., transactionsRoot: _Optional[bool] = ..., receiptsRoot: _Optional[bool] = ..., logsBloom: _Optional[bool] = ..., gasLimit: _Optional[bool] = ..., gasUsed: _Optional[bool] = ..., miner: _Optional[bool] = ..., extraData: _Optional[bool] = ..., size: _Optional[bool] = ..., sha3Uncles: _Optional[bool] = ..., nonce: _Optional[bool] = ..., baseFeePerGas: _Optional[bool] = ..., difficulty: _Optional[bool] = ..., totalDifficulty: _Optional[bool] = ..., mixHash: _Optional[bool] = ..., blobGasUsed: _Optional[bool] = ..., excessBlobGas: _Optional[bool] = ..., withdrawalsRoot: _Optional[bool] = ..., parentBeaconBlockRoot: _Optional[bool] = ..., requestsHash: _Optional[bool] = ...) -> None: ...

class TransactionFieldSelection(_message.Message):
    __slots__ = ("hash", "nonce", "to", "value", "input", "type", "gas", "gasPrice", "maxFeePerGas", "maxPriorityFeePerGas", "gasUsed", "effectiveGasPrice", "blockNumber", "blockHash", "transactionIndex", "blockTimestamp", "chainId", "accessList", "maxFeePerBlobGas", "blobVersionedHashes", "r", "s", "v", "yParity", "status", "cumulativeGasUsed", "contractAddress", "logsBloom", "authorizationList", "sourceHash", "mint")
    HASH_FIELD_NUMBER: _ClassVar[int]
    NONCE_FIELD_NUMBER: _ClassVar[int]
    FROM_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    GAS_FIELD_NUMBER: _ClassVar[int]
    GASPRICE_FIELD_NUMBER: _ClassVar[int]
    MAXFEEPERGAS_FIELD_NUMBER: _ClassVar[int]
    MAXPRIORITYFEEPERGAS_FIELD_NUMBER: _ClassVar[int]
    GASUSED_FIELD_NUMBER: _ClassVar[int]
    EFFECTIVEGASPRICE_FIELD_NUMBER: _ClassVar[int]
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONINDEX_FIELD_NUMBER: _ClassVar[int]
    BLOCKTIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    ACCESSLIST_FIELD_NUMBER: _ClassVar[int]
    MAXFEEPERBLOBGAS_FIELD_NUMBER: _ClassVar[int]
    BLOBVERSIONEDHASHES_FIELD_NUMBER: _ClassVar[int]
    R_FIELD_NUMBER: _ClassVar[int]
    S_FIELD_NUMBER: _ClassVar[int]
    V_FIELD_NUMBER: _ClassVar[int]
    YPARITY_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    CUMULATIVEGASUSED_FIELD_NUMBER: _ClassVar[int]
    CONTRACTADDRESS_FIELD_NUMBER: _ClassVar[int]
    LOGSBLOOM_FIELD_NUMBER: _ClassVar[int]
    AUTHORIZATIONLIST_FIELD_NUMBER: _ClassVar[int]
    SOURCEHASH_FIELD_NUMBER: _ClassVar[int]
    MINT_FIELD_NUMBER: _ClassVar[int]
    hash: bool
    nonce: bool
    to: bool
    value: bool
    input: bool
    type: bool
    gas: bool
    gasPrice: bool
    maxFeePerGas: bool
    maxPriorityFeePerGas: bool
    gasUsed: bool
    effectiveGasPrice: bool
    blockNumber: bool
    blockHash: bool
    transactionIndex: bool
    blockTimestamp: bool
    chainId: bool
    accessList: bool
    maxFeePerBlobGas: bool
    blobVersionedHashes: bool
    r: bool
    s: bool
    v: bool
    yParity: bool
    status: bool
    cumulativeGasUsed: bool
    contractAddress: bool
    logsBloom: bool
    authorizationList: bool
    sourceHash: bool
    mint: bool
    def __init__(self, hash: _Optional[bool] = ..., nonce: _Optional[bool] = ..., to: _Optional[bool] = ..., value: _Optional[bool] = ..., input: _Optional[bool] = ..., type: _Optional[bool] = ..., gas: _Optional[bool] = ..., gasPrice: _Optional[bool] = ..., maxFeePerGas: _Optional[bool] = ..., maxPriorityFeePerGas: _Optional[bool] = ..., gasUsed: _Optional[bool] = ..., effectiveGasPrice: _Optional[bool] = ..., blockNumber: _Optional[bool] = ..., blockHash: _Optional[bool] = ..., transactionIndex: _Optional[bool] = ..., blockTimestamp: _Optional[bool] = ..., chainId: _Optional[bool] = ..., accessList: _Optional[bool] = ..., maxFeePerBlobGas: _Optional[bool] = ..., blobVersionedHashes: _Optional[bool] = ..., r: _Optional[bool] = ..., s: _Optional[bool] = ..., v: _Optional[bool] = ..., yParity: _Optional[bool] = ..., status: _Optional[bool] = ..., cumulativeGasUsed: _Optional[bool] = ..., contractAddress: _Optional[bool] = ..., logsBloom: _Optional[bool] = ..., authorizationList: _Optional[bool] = ..., sourceHash: _Optional[bool] = ..., mint: _Optional[bool] = ..., **kwargs) -> None: ...

class LogFieldSelection(_message.Message):
    __slots__ = ("address", "topics", "data", "blockNumber", "blockHash", "transactionHash", "transactionIndex", "logIndex", "blockTimestamp", "removed")
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    TOPICS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONHASH_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONINDEX_FIELD_NUMBER: _ClassVar[int]
    LOGINDEX_FIELD_NUMBER: _ClassVar[int]
    BLOCKTIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    REMOVED_FIELD_NUMBER: _ClassVar[int]
    address: bool
    topics: bool
    data: bool
    blockNumber: bool
    blockHash: bool
    transactionHash: bool
    transactionIndex: bool
    logIndex: bool
    blockTimestamp: bool
    removed: bool
    def __init__(self, address: _Optional[bool] = ..., topics: _Optional[bool] = ..., data: _Optional[bool] = ..., blockNumber: _Optional[bool] = ..., blockHash: _Optional[bool] = ..., transactionHash: _Optional[bool] = ..., transactionIndex: _Optional[bool] = ..., logIndex: _Optional[bool] = ..., blockTimestamp: _Optional[bool] = ..., removed: _Optional[bool] = ...) -> None: ...

class TraceFieldSelection(_message.Message):
    __slots__ = ("type", "to", "value", "gas", "gasUsed", "input", "output", "error", "reverted", "blockHash", "blockNumber", "transactionHash", "transactionIndex", "traceAddress")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    FROM_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    GAS_FIELD_NUMBER: _ClassVar[int]
    GASUSED_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    REVERTED_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONHASH_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONINDEX_FIELD_NUMBER: _ClassVar[int]
    TRACEADDRESS_FIELD_NUMBER: _ClassVar[int]
    type: bool
    to: bool
    value: bool
    gas: bool
    gasUsed: bool
    input: bool
    output: bool
    error: bool
    reverted: bool
    blockHash: bool
    blockNumber: bool
    transactionHash: bool
    transactionIndex: bool
    traceAddress: bool
    def __init__(self, type: _Optional[bool] = ..., to: _Optional[bool] = ..., value: _Optional[bool] = ..., gas: _Optional[bool] = ..., gasUsed: _Optional[bool] = ..., input: _Optional[bool] = ..., output: _Optional[bool] = ..., error: _Optional[bool] = ..., reverted: _Optional[bool] = ..., blockHash: _Optional[bool] = ..., blockNumber: _Optional[bool] = ..., transactionHash: _Optional[bool] = ..., transactionIndex: _Optional[bool] = ..., traceAddress: _Optional[bool] = ..., **kwargs) -> None: ...

class TransferFieldSelection(_message.Message):
    __slots__ = ("type", "to", "value", "gas", "gasUsed", "input", "output", "error", "reverted", "blockHash", "blockNumber", "transactionHash", "transactionIndex", "traceAddress")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    FROM_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    GAS_FIELD_NUMBER: _ClassVar[int]
    GASUSED_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    REVERTED_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONHASH_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONINDEX_FIELD_NUMBER: _ClassVar[int]
    TRACEADDRESS_FIELD_NUMBER: _ClassVar[int]
    type: bool
    to: bool
    value: bool
    gas: bool
    gasUsed: bool
    input: bool
    output: bool
    error: bool
    reverted: bool
    blockHash: bool
    blockNumber: bool
    transactionHash: bool
    transactionIndex: bool
    traceAddress: bool
    def __init__(self, type: _Optional[bool] = ..., to: _Optional[bool] = ..., value: _Optional[bool] = ..., gas: _Optional[bool] = ..., gasUsed: _Optional[bool] = ..., input: _Optional[bool] = ..., output: _Optional[bool] = ..., error: _Optional[bool] = ..., reverted: _Optional[bool] = ..., blockHash: _Optional[bool] = ..., blockNumber: _Optional[bool] = ..., transactionHash: _Optional[bool] = ..., transactionIndex: _Optional[bool] = ..., traceAddress: _Optional[bool] = ..., **kwargs) -> None: ...

class BlockFilter(_message.Message):
    __slots__ = ("miner",)
    MINER_FIELD_NUMBER: _ClassVar[int]
    miner: _containers.RepeatedScalarFieldContainer[bytes]
    def __init__(self, miner: _Optional[_Iterable[bytes]] = ...) -> None: ...

class TransactionFilter(_message.Message):
    __slots__ = ("to", "selector")
    FROM_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    SELECTOR_FIELD_NUMBER: _ClassVar[int]
    to: _containers.RepeatedScalarFieldContainer[bytes]
    selector: _containers.RepeatedScalarFieldContainer[bytes]
    def __init__(self, to: _Optional[_Iterable[bytes]] = ..., selector: _Optional[_Iterable[bytes]] = ..., **kwargs) -> None: ...

class LogFilter(_message.Message):
    __slots__ = ("address", "topics")
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    TOPICS_FIELD_NUMBER: _ClassVar[int]
    address: _containers.RepeatedScalarFieldContainer[bytes]
    topics: _containers.RepeatedCompositeFieldContainer[_rpc_pb2.TopicFilter]
    def __init__(self, address: _Optional[_Iterable[bytes]] = ..., topics: _Optional[_Iterable[_Union[_rpc_pb2.TopicFilter, _Mapping]]] = ...) -> None: ...

class TraceFilter(_message.Message):
    __slots__ = ("to", "selector", "includeReverted")
    FROM_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    SELECTOR_FIELD_NUMBER: _ClassVar[int]
    INCLUDEREVERTED_FIELD_NUMBER: _ClassVar[int]
    to: _containers.RepeatedScalarFieldContainer[bytes]
    selector: _containers.RepeatedScalarFieldContainer[bytes]
    includeReverted: bool
    def __init__(self, to: _Optional[_Iterable[bytes]] = ..., selector: _Optional[_Iterable[bytes]] = ..., includeReverted: _Optional[bool] = ..., **kwargs) -> None: ...

class TransferFilter(_message.Message):
    __slots__ = ("to", "includeReverted")
    FROM_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    INCLUDEREVERTED_FIELD_NUMBER: _ClassVar[int]
    to: _containers.RepeatedScalarFieldContainer[bytes]
    includeReverted: bool
    def __init__(self, to: _Optional[_Iterable[bytes]] = ..., includeReverted: _Optional[bool] = ..., **kwargs) -> None: ...

class QueryBlocksRequest(_message.Message):
    __slots__ = ("fromBlock", "toBlock", "order", "target", "chainId", "blockFields", "filter")
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    TARGET_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    BLOCKFIELDS_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    fromBlock: str
    toBlock: str
    order: SortOrder
    target: int
    chainId: int
    blockFields: BlockFieldSelection
    filter: BlockFilter
    def __init__(self, fromBlock: _Optional[str] = ..., toBlock: _Optional[str] = ..., order: _Optional[_Union[SortOrder, str]] = ..., target: _Optional[int] = ..., chainId: _Optional[int] = ..., blockFields: _Optional[_Union[BlockFieldSelection, _Mapping]] = ..., filter: _Optional[_Union[BlockFilter, _Mapping]] = ...) -> None: ...

class QueryTransactionsRequest(_message.Message):
    __slots__ = ("fromBlock", "toBlock", "order", "target", "chainId", "filter", "transactionFields", "blockFields")
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    TARGET_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONFIELDS_FIELD_NUMBER: _ClassVar[int]
    BLOCKFIELDS_FIELD_NUMBER: _ClassVar[int]
    fromBlock: str
    toBlock: str
    order: SortOrder
    target: int
    chainId: int
    filter: TransactionFilter
    transactionFields: TransactionFieldSelection
    blockFields: BlockFieldSelection
    def __init__(self, fromBlock: _Optional[str] = ..., toBlock: _Optional[str] = ..., order: _Optional[_Union[SortOrder, str]] = ..., target: _Optional[int] = ..., chainId: _Optional[int] = ..., filter: _Optional[_Union[TransactionFilter, _Mapping]] = ..., transactionFields: _Optional[_Union[TransactionFieldSelection, _Mapping]] = ..., blockFields: _Optional[_Union[BlockFieldSelection, _Mapping]] = ...) -> None: ...

class QueryLogsRequest(_message.Message):
    __slots__ = ("fromBlock", "toBlock", "order", "target", "chainId", "filter", "logFields", "transactionFields", "blockFields")
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    TARGET_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    LOGFIELDS_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONFIELDS_FIELD_NUMBER: _ClassVar[int]
    BLOCKFIELDS_FIELD_NUMBER: _ClassVar[int]
    fromBlock: str
    toBlock: str
    order: SortOrder
    target: int
    chainId: int
    filter: LogFilter
    logFields: LogFieldSelection
    transactionFields: TransactionFieldSelection
    blockFields: BlockFieldSelection
    def __init__(self, fromBlock: _Optional[str] = ..., toBlock: _Optional[str] = ..., order: _Optional[_Union[SortOrder, str]] = ..., target: _Optional[int] = ..., chainId: _Optional[int] = ..., filter: _Optional[_Union[LogFilter, _Mapping]] = ..., logFields: _Optional[_Union[LogFieldSelection, _Mapping]] = ..., transactionFields: _Optional[_Union[TransactionFieldSelection, _Mapping]] = ..., blockFields: _Optional[_Union[BlockFieldSelection, _Mapping]] = ...) -> None: ...

class QueryTracesRequest(_message.Message):
    __slots__ = ("fromBlock", "toBlock", "order", "target", "chainId", "filter", "traceFields", "transactionFields", "blockFields")
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    TARGET_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    TRACEFIELDS_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONFIELDS_FIELD_NUMBER: _ClassVar[int]
    BLOCKFIELDS_FIELD_NUMBER: _ClassVar[int]
    fromBlock: str
    toBlock: str
    order: SortOrder
    target: int
    chainId: int
    filter: TraceFilter
    traceFields: TraceFieldSelection
    transactionFields: TransactionFieldSelection
    blockFields: BlockFieldSelection
    def __init__(self, fromBlock: _Optional[str] = ..., toBlock: _Optional[str] = ..., order: _Optional[_Union[SortOrder, str]] = ..., target: _Optional[int] = ..., chainId: _Optional[int] = ..., filter: _Optional[_Union[TraceFilter, _Mapping]] = ..., traceFields: _Optional[_Union[TraceFieldSelection, _Mapping]] = ..., transactionFields: _Optional[_Union[TransactionFieldSelection, _Mapping]] = ..., blockFields: _Optional[_Union[BlockFieldSelection, _Mapping]] = ...) -> None: ...

class QueryTransfersRequest(_message.Message):
    __slots__ = ("fromBlock", "toBlock", "order", "target", "chainId", "filter", "transferFields", "transactionFields", "blockFields")
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    TARGET_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    TRANSFERFIELDS_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONFIELDS_FIELD_NUMBER: _ClassVar[int]
    BLOCKFIELDS_FIELD_NUMBER: _ClassVar[int]
    fromBlock: str
    toBlock: str
    order: SortOrder
    target: int
    chainId: int
    filter: TransferFilter
    transferFields: TransferFieldSelection
    transactionFields: TransactionFieldSelection
    blockFields: BlockFieldSelection
    def __init__(self, fromBlock: _Optional[str] = ..., toBlock: _Optional[str] = ..., order: _Optional[_Union[SortOrder, str]] = ..., target: _Optional[int] = ..., chainId: _Optional[int] = ..., filter: _Optional[_Union[TransferFilter, _Mapping]] = ..., transferFields: _Optional[_Union[TransferFieldSelection, _Mapping]] = ..., transactionFields: _Optional[_Union[TransactionFieldSelection, _Mapping]] = ..., blockFields: _Optional[_Union[BlockFieldSelection, _Mapping]] = ...) -> None: ...

class QueryBlocksResponse(_message.Message):
    __slots__ = ("blocks", "fromBlock", "toBlock", "cursorBlock")
    BLOCKS_FIELD_NUMBER: _ClassVar[int]
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    CURSORBLOCK_FIELD_NUMBER: _ClassVar[int]
    blocks: _containers.RepeatedCompositeFieldContainer[_models_pb2.BlockHeader]
    fromBlock: CursorBlock
    toBlock: CursorBlock
    cursorBlock: CursorBlock
    def __init__(self, blocks: _Optional[_Iterable[_Union[_models_pb2.BlockHeader, _Mapping]]] = ..., fromBlock: _Optional[_Union[CursorBlock, _Mapping]] = ..., toBlock: _Optional[_Union[CursorBlock, _Mapping]] = ..., cursorBlock: _Optional[_Union[CursorBlock, _Mapping]] = ...) -> None: ...

class QueryTransactionsResponse(_message.Message):
    __slots__ = ("transactions", "blocks", "fromBlock", "toBlock", "cursorBlock")
    TRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    BLOCKS_FIELD_NUMBER: _ClassVar[int]
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    CURSORBLOCK_FIELD_NUMBER: _ClassVar[int]
    transactions: _containers.RepeatedCompositeFieldContainer[_models_pb2.Transaction]
    blocks: _containers.RepeatedCompositeFieldContainer[_models_pb2.BlockHeader]
    fromBlock: CursorBlock
    toBlock: CursorBlock
    cursorBlock: CursorBlock
    def __init__(self, transactions: _Optional[_Iterable[_Union[_models_pb2.Transaction, _Mapping]]] = ..., blocks: _Optional[_Iterable[_Union[_models_pb2.BlockHeader, _Mapping]]] = ..., fromBlock: _Optional[_Union[CursorBlock, _Mapping]] = ..., toBlock: _Optional[_Union[CursorBlock, _Mapping]] = ..., cursorBlock: _Optional[_Union[CursorBlock, _Mapping]] = ...) -> None: ...

class QueryLogsResponse(_message.Message):
    __slots__ = ("logs", "transactions", "blocks", "fromBlock", "toBlock", "cursorBlock")
    LOGS_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    BLOCKS_FIELD_NUMBER: _ClassVar[int]
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    CURSORBLOCK_FIELD_NUMBER: _ClassVar[int]
    logs: _containers.RepeatedCompositeFieldContainer[_models_pb2.Log]
    transactions: _containers.RepeatedCompositeFieldContainer[_models_pb2.Transaction]
    blocks: _containers.RepeatedCompositeFieldContainer[_models_pb2.BlockHeader]
    fromBlock: CursorBlock
    toBlock: CursorBlock
    cursorBlock: CursorBlock
    def __init__(self, logs: _Optional[_Iterable[_Union[_models_pb2.Log, _Mapping]]] = ..., transactions: _Optional[_Iterable[_Union[_models_pb2.Transaction, _Mapping]]] = ..., blocks: _Optional[_Iterable[_Union[_models_pb2.BlockHeader, _Mapping]]] = ..., fromBlock: _Optional[_Union[CursorBlock, _Mapping]] = ..., toBlock: _Optional[_Union[CursorBlock, _Mapping]] = ..., cursorBlock: _Optional[_Union[CursorBlock, _Mapping]] = ...) -> None: ...

class QueryTracesResponse(_message.Message):
    __slots__ = ("traces", "transactions", "blocks", "fromBlock", "toBlock", "cursorBlock")
    TRACES_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    BLOCKS_FIELD_NUMBER: _ClassVar[int]
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    CURSORBLOCK_FIELD_NUMBER: _ClassVar[int]
    traces: _containers.RepeatedCompositeFieldContainer[_models_pb2.Trace]
    transactions: _containers.RepeatedCompositeFieldContainer[_models_pb2.Transaction]
    blocks: _containers.RepeatedCompositeFieldContainer[_models_pb2.BlockHeader]
    fromBlock: CursorBlock
    toBlock: CursorBlock
    cursorBlock: CursorBlock
    def __init__(self, traces: _Optional[_Iterable[_Union[_models_pb2.Trace, _Mapping]]] = ..., transactions: _Optional[_Iterable[_Union[_models_pb2.Transaction, _Mapping]]] = ..., blocks: _Optional[_Iterable[_Union[_models_pb2.BlockHeader, _Mapping]]] = ..., fromBlock: _Optional[_Union[CursorBlock, _Mapping]] = ..., toBlock: _Optional[_Union[CursorBlock, _Mapping]] = ..., cursorBlock: _Optional[_Union[CursorBlock, _Mapping]] = ...) -> None: ...

class QueryTransfersResponse(_message.Message):
    __slots__ = ("transfers", "transactions", "blocks", "fromBlock", "toBlock", "cursorBlock")
    TRANSFERS_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    BLOCKS_FIELD_NUMBER: _ClassVar[int]
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    CURSORBLOCK_FIELD_NUMBER: _ClassVar[int]
    transfers: _containers.RepeatedCompositeFieldContainer[_models_pb2.NativeTransfer]
    transactions: _containers.RepeatedCompositeFieldContainer[_models_pb2.Transaction]
    blocks: _containers.RepeatedCompositeFieldContainer[_models_pb2.BlockHeader]
    fromBlock: CursorBlock
    toBlock: CursorBlock
    cursorBlock: CursorBlock
    def __init__(self, transfers: _Optional[_Iterable[_Union[_models_pb2.NativeTransfer, _Mapping]]] = ..., transactions: _Optional[_Iterable[_Union[_models_pb2.Transaction, _Mapping]]] = ..., blocks: _Optional[_Iterable[_Union[_models_pb2.BlockHeader, _Mapping]]] = ..., fromBlock: _Optional[_Union[CursorBlock, _Mapping]] = ..., toBlock: _Optional[_Union[CursorBlock, _Mapping]] = ..., cursorBlock: _Optional[_Union[CursorBlock, _Mapping]] = ...) -> None: ...
