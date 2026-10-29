from . import models_pb2 as _models_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetBlocksByRangeRequest(_message.Message):
    __slots__ = ("fromBlock", "toBlock", "includeTransactions", "limit")
    FROMBLOCK_FIELD_NUMBER: _ClassVar[int]
    TOBLOCK_FIELD_NUMBER: _ClassVar[int]
    INCLUDETRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    fromBlock: int
    toBlock: int
    includeTransactions: bool
    limit: int
    def __init__(self, fromBlock: _Optional[int] = ..., toBlock: _Optional[int] = ..., includeTransactions: _Optional[bool] = ..., limit: _Optional[int] = ...) -> None: ...

class GetBlocksByRangeResponse(_message.Message):
    __slots__ = ("blocks", "nextCursor", "isPartial", "timestamp", "processingTimeMs", "metadata")
    class MetadataEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    BLOCKS_FIELD_NUMBER: _ClassVar[int]
    NEXTCURSOR_FIELD_NUMBER: _ClassVar[int]
    ISPARTIAL_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    PROCESSINGTIMEMS_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    blocks: _containers.RepeatedCompositeFieldContainer[_models_pb2.Block]
    nextCursor: str
    isPartial: bool
    timestamp: int
    processingTimeMs: int
    metadata: _containers.ScalarMap[str, str]
    def __init__(self, blocks: _Optional[_Iterable[_Union[_models_pb2.Block, _Mapping]]] = ..., nextCursor: _Optional[str] = ..., isPartial: _Optional[bool] = ..., timestamp: _Optional[int] = ..., processingTimeMs: _Optional[int] = ..., metadata: _Optional[_Mapping[str, str]] = ...) -> None: ...
