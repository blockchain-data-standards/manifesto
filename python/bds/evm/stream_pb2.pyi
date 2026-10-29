from . import models_pb2 as _models_pb2
from . import query_pb2 as _query_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Finality(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    FINALITY_LATEST: _ClassVar[Finality]
    FINALITY_SAFE: _ClassVar[Finality]
    FINALITY_FINALIZED: _ClassVar[Finality]
    FINALITY_DEPTH: _ClassVar[Finality]

class Reorg(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    REORG_UNDO: _ClassVar[Reorg]
    REORG_NONE: _ClassVar[Reorg]
FINALITY_LATEST: Finality
FINALITY_SAFE: Finality
FINALITY_FINALIZED: Finality
FINALITY_DEPTH: Finality
REORG_UNDO: Reorg
REORG_NONE: Reorg

class StreamBlocksRequest(_message.Message):
    __slots__ = ("chainId", "chainGenesisHash", "fromBlock", "fromHash", "finality", "depth", "heartbeatSeconds")
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    FROMHASH_FIELD_NUMBER: _ClassVar[int]
    FINALITY_FIELD_NUMBER: _ClassVar[int]
    DEPTH_FIELD_NUMBER: _ClassVar[int]
    HEARTBEATSECONDS_FIELD_NUMBER: _ClassVar[int]
    chainId: int
    chainGenesisHash: bytes
    fromBlock: int
    fromHash: bytes
    finality: Finality
    depth: int
    heartbeatSeconds: int
    def __init__(self, chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ..., fromBlock: _Optional[int] = ..., fromHash: _Optional[bytes] = ..., finality: _Optional[_Union[Finality, str]] = ..., depth: _Optional[int] = ..., heartbeatSeconds: _Optional[int] = ...) -> None: ...

class StreamBlocksResponse(_message.Message):
    __slots__ = ("header", "heartbeat", "status")
    HEADER_FIELD_NUMBER: _ClassVar[int]
    HEARTBEAT_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    header: _models_pb2.BlockHeader
    heartbeat: Heartbeat
    status: StreamStatus
    def __init__(self, header: _Optional[_Union[_models_pb2.BlockHeader, _Mapping]] = ..., heartbeat: _Optional[_Union[Heartbeat, _Mapping]] = ..., status: _Optional[_Union[StreamStatus, _Mapping]] = ...) -> None: ...

class StreamLogsRequest(_message.Message):
    __slots__ = ("chainId", "chainGenesisHash", "fromBlock", "fromHash", "filter", "finality", "depth", "reorg", "heartbeatSeconds")
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    FROMHASH_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    FINALITY_FIELD_NUMBER: _ClassVar[int]
    DEPTH_FIELD_NUMBER: _ClassVar[int]
    REORG_FIELD_NUMBER: _ClassVar[int]
    HEARTBEATSECONDS_FIELD_NUMBER: _ClassVar[int]
    chainId: int
    chainGenesisHash: bytes
    fromBlock: int
    fromHash: bytes
    filter: _query_pb2.LogFilter
    finality: Finality
    depth: int
    reorg: Reorg
    heartbeatSeconds: int
    def __init__(self, chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ..., fromBlock: _Optional[int] = ..., fromHash: _Optional[bytes] = ..., filter: _Optional[_Union[_query_pb2.LogFilter, _Mapping]] = ..., finality: _Optional[_Union[Finality, str]] = ..., depth: _Optional[int] = ..., reorg: _Optional[_Union[Reorg, str]] = ..., heartbeatSeconds: _Optional[int] = ...) -> None: ...

class StreamLogsResponse(_message.Message):
    __slots__ = ("block", "undo", "heartbeat", "status")
    BLOCK_FIELD_NUMBER: _ClassVar[int]
    UNDO_FIELD_NUMBER: _ClassVar[int]
    HEARTBEAT_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    block: BlockLogs
    undo: Undo
    heartbeat: Heartbeat
    status: StreamStatus
    def __init__(self, block: _Optional[_Union[BlockLogs, _Mapping]] = ..., undo: _Optional[_Union[Undo, _Mapping]] = ..., heartbeat: _Optional[_Union[Heartbeat, _Mapping]] = ..., status: _Optional[_Union[StreamStatus, _Mapping]] = ...) -> None: ...

class BlockLogs(_message.Message):
    __slots__ = ("block", "logs", "complete")
    BLOCK_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    COMPLETE_FIELD_NUMBER: _ClassVar[int]
    block: _models_pb2.BlockRef
    logs: _containers.RepeatedCompositeFieldContainer[_models_pb2.Log]
    complete: bool
    def __init__(self, block: _Optional[_Union[_models_pb2.BlockRef, _Mapping]] = ..., logs: _Optional[_Iterable[_Union[_models_pb2.Log, _Mapping]]] = ..., complete: _Optional[bool] = ...) -> None: ...

class StreamPendingTransactionsRequest(_message.Message):
    __slots__ = ("chainId", "chainGenesisHash", "fullTransactions", "heartbeatSeconds")
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    CHAINGENESISHASH_FIELD_NUMBER: _ClassVar[int]
    FULLTRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    HEARTBEATSECONDS_FIELD_NUMBER: _ClassVar[int]
    chainId: int
    chainGenesisHash: bytes
    fullTransactions: bool
    heartbeatSeconds: int
    def __init__(self, chainId: _Optional[int] = ..., chainGenesisHash: _Optional[bytes] = ..., fullTransactions: _Optional[bool] = ..., heartbeatSeconds: _Optional[int] = ...) -> None: ...

class StreamPendingTransactionsResponse(_message.Message):
    __slots__ = ("transactions", "heartbeat", "status")
    TRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    HEARTBEAT_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    transactions: PendingTransactions
    heartbeat: Heartbeat
    status: StreamStatus
    def __init__(self, transactions: _Optional[_Union[PendingTransactions, _Mapping]] = ..., heartbeat: _Optional[_Union[Heartbeat, _Mapping]] = ..., status: _Optional[_Union[StreamStatus, _Mapping]] = ...) -> None: ...

class PendingTransactions(_message.Message):
    __slots__ = ("hashes", "transactions")
    HASHES_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    hashes: _containers.RepeatedScalarFieldContainer[bytes]
    transactions: _containers.RepeatedCompositeFieldContainer[_models_pb2.Transaction]
    def __init__(self, hashes: _Optional[_Iterable[bytes]] = ..., transactions: _Optional[_Iterable[_Union[_models_pb2.Transaction, _Mapping]]] = ...) -> None: ...

class Undo(_message.Message):
    __slots__ = ("forkPoint",)
    FORKPOINT_FIELD_NUMBER: _ClassVar[int]
    forkPoint: _models_pb2.BlockRef
    def __init__(self, forkPoint: _Optional[_Union[_models_pb2.BlockRef, _Mapping]] = ...) -> None: ...

class Heartbeat(_message.Message):
    __slots__ = ("head", "finalized", "safe")
    HEAD_FIELD_NUMBER: _ClassVar[int]
    FINALIZED_FIELD_NUMBER: _ClassVar[int]
    SAFE_FIELD_NUMBER: _ClassVar[int]
    head: _models_pb2.BlockRef
    finalized: int
    safe: int
    def __init__(self, head: _Optional[_Union[_models_pb2.BlockRef, _Mapping]] = ..., finalized: _Optional[int] = ..., safe: _Optional[int] = ...) -> None: ...

class StreamStatus(_message.Message):
    __slots__ = ("state", "reason")
    class State(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        STATE_OK: _ClassVar[StreamStatus.State]
        STATE_DEGRADED: _ClassVar[StreamStatus.State]
        STATE_LAGGING: _ClassVar[StreamStatus.State]
    STATE_OK: StreamStatus.State
    STATE_DEGRADED: StreamStatus.State
    STATE_LAGGING: StreamStatus.State
    STATE_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    state: StreamStatus.State
    reason: str
    def __init__(self, state: _Optional[_Union[StreamStatus.State, str]] = ..., reason: _Optional[str] = ...) -> None: ...
