from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TransactionType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    LEGACY: _ClassVar[TransactionType]
    ACCESS_LIST: _ClassVar[TransactionType]
    DYNAMIC_FEE: _ClassVar[TransactionType]
    BLOB: _ClassVar[TransactionType]
    SET_CODE: _ClassVar[TransactionType]

class TraceType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    TRACE_CALL: _ClassVar[TraceType]
    TRACE_CREATE: _ClassVar[TraceType]
    TRACE_SELFDESTRUCT: _ClassVar[TraceType]
    TRACE_REWARD: _ClassVar[TraceType]

class TraceCallType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    TRACE_CALL_CALL: _ClassVar[TraceCallType]
    TRACE_CALL_STATICCALL: _ClassVar[TraceCallType]
    TRACE_CALL_DELEGATECALL: _ClassVar[TraceCallType]
    TRACE_CALL_CALLCODE: _ClassVar[TraceCallType]
LEGACY: TransactionType
ACCESS_LIST: TransactionType
DYNAMIC_FEE: TransactionType
BLOB: TransactionType
SET_CODE: TransactionType
TRACE_CALL: TraceType
TRACE_CREATE: TraceType
TRACE_SELFDESTRUCT: TraceType
TRACE_REWARD: TraceType
TRACE_CALL_CALL: TraceCallType
TRACE_CALL_STATICCALL: TraceCallType
TRACE_CALL_DELEGATECALL: TraceCallType
TRACE_CALL_CALLCODE: TraceCallType

class BlockRef(_message.Message):
    __slots__ = ("number", "hash", "parentHash")
    NUMBER_FIELD_NUMBER: _ClassVar[int]
    HASH_FIELD_NUMBER: _ClassVar[int]
    PARENTHASH_FIELD_NUMBER: _ClassVar[int]
    number: int
    hash: bytes
    parentHash: bytes
    def __init__(self, number: _Optional[int] = ..., hash: _Optional[bytes] = ..., parentHash: _Optional[bytes] = ...) -> None: ...

class BlockHeader(_message.Message):
    __slots__ = ("number", "timestamp", "gasLimit", "gasUsed", "size", "hash", "parentHash", "stateRoot", "transactionsRoot", "receiptsRoot", "sha3Uncles", "miner", "logsBloom", "extraData", "nonce", "blobGasUsed", "excessBlobGas", "l1BlockNumber", "epoch", "slot", "proposerIndex", "sendCount", "transactionCount", "mixHash", "parentBeaconBlockRoot", "withdrawalsRoot", "sendRoot", "baseFeePerGas", "difficulty", "totalDifficulty", "proposerPublicKey", "withdrawals", "canonicalRlp", "uncles", "requestsHash")
    NUMBER_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    GASLIMIT_FIELD_NUMBER: _ClassVar[int]
    GASUSED_FIELD_NUMBER: _ClassVar[int]
    SIZE_FIELD_NUMBER: _ClassVar[int]
    HASH_FIELD_NUMBER: _ClassVar[int]
    PARENTHASH_FIELD_NUMBER: _ClassVar[int]
    STATEROOT_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONSROOT_FIELD_NUMBER: _ClassVar[int]
    RECEIPTSROOT_FIELD_NUMBER: _ClassVar[int]
    SHA3UNCLES_FIELD_NUMBER: _ClassVar[int]
    MINER_FIELD_NUMBER: _ClassVar[int]
    LOGSBLOOM_FIELD_NUMBER: _ClassVar[int]
    EXTRADATA_FIELD_NUMBER: _ClassVar[int]
    NONCE_FIELD_NUMBER: _ClassVar[int]
    BLOBGASUSED_FIELD_NUMBER: _ClassVar[int]
    EXCESSBLOBGAS_FIELD_NUMBER: _ClassVar[int]
    L1BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    EPOCH_FIELD_NUMBER: _ClassVar[int]
    SLOT_FIELD_NUMBER: _ClassVar[int]
    PROPOSERINDEX_FIELD_NUMBER: _ClassVar[int]
    SENDCOUNT_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONCOUNT_FIELD_NUMBER: _ClassVar[int]
    MIXHASH_FIELD_NUMBER: _ClassVar[int]
    PARENTBEACONBLOCKROOT_FIELD_NUMBER: _ClassVar[int]
    WITHDRAWALSROOT_FIELD_NUMBER: _ClassVar[int]
    SENDROOT_FIELD_NUMBER: _ClassVar[int]
    BASEFEEPERGAS_FIELD_NUMBER: _ClassVar[int]
    DIFFICULTY_FIELD_NUMBER: _ClassVar[int]
    TOTALDIFFICULTY_FIELD_NUMBER: _ClassVar[int]
    PROPOSERPUBLICKEY_FIELD_NUMBER: _ClassVar[int]
    WITHDRAWALS_FIELD_NUMBER: _ClassVar[int]
    CANONICALRLP_FIELD_NUMBER: _ClassVar[int]
    UNCLES_FIELD_NUMBER: _ClassVar[int]
    REQUESTSHASH_FIELD_NUMBER: _ClassVar[int]
    number: int
    timestamp: int
    gasLimit: int
    gasUsed: int
    size: int
    hash: bytes
    parentHash: bytes
    stateRoot: bytes
    transactionsRoot: bytes
    receiptsRoot: bytes
    sha3Uncles: bytes
    miner: bytes
    logsBloom: bytes
    extraData: bytes
    nonce: int
    blobGasUsed: int
    excessBlobGas: int
    l1BlockNumber: int
    epoch: int
    slot: int
    proposerIndex: int
    sendCount: int
    transactionCount: int
    mixHash: bytes
    parentBeaconBlockRoot: bytes
    withdrawalsRoot: bytes
    sendRoot: bytes
    baseFeePerGas: str
    difficulty: str
    totalDifficulty: str
    proposerPublicKey: str
    withdrawals: bytes
    canonicalRlp: bytes
    uncles: _containers.RepeatedScalarFieldContainer[bytes]
    requestsHash: bytes
    def __init__(self, number: _Optional[int] = ..., timestamp: _Optional[int] = ..., gasLimit: _Optional[int] = ..., gasUsed: _Optional[int] = ..., size: _Optional[int] = ..., hash: _Optional[bytes] = ..., parentHash: _Optional[bytes] = ..., stateRoot: _Optional[bytes] = ..., transactionsRoot: _Optional[bytes] = ..., receiptsRoot: _Optional[bytes] = ..., sha3Uncles: _Optional[bytes] = ..., miner: _Optional[bytes] = ..., logsBloom: _Optional[bytes] = ..., extraData: _Optional[bytes] = ..., nonce: _Optional[int] = ..., blobGasUsed: _Optional[int] = ..., excessBlobGas: _Optional[int] = ..., l1BlockNumber: _Optional[int] = ..., epoch: _Optional[int] = ..., slot: _Optional[int] = ..., proposerIndex: _Optional[int] = ..., sendCount: _Optional[int] = ..., transactionCount: _Optional[int] = ..., mixHash: _Optional[bytes] = ..., parentBeaconBlockRoot: _Optional[bytes] = ..., withdrawalsRoot: _Optional[bytes] = ..., sendRoot: _Optional[bytes] = ..., baseFeePerGas: _Optional[str] = ..., difficulty: _Optional[str] = ..., totalDifficulty: _Optional[str] = ..., proposerPublicKey: _Optional[str] = ..., withdrawals: _Optional[bytes] = ..., canonicalRlp: _Optional[bytes] = ..., uncles: _Optional[_Iterable[bytes]] = ..., requestsHash: _Optional[bytes] = ...) -> None: ...

class Block(_message.Message):
    __slots__ = ("header", "transactionHashes", "fullTransactions", "logs", "withdrawals")
    HEADER_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONHASHES_FIELD_NUMBER: _ClassVar[int]
    FULLTRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    WITHDRAWALS_FIELD_NUMBER: _ClassVar[int]
    header: BlockHeader
    transactionHashes: _containers.RepeatedScalarFieldContainer[bytes]
    fullTransactions: _containers.RepeatedCompositeFieldContainer[Transaction]
    logs: _containers.RepeatedCompositeFieldContainer[Log]
    withdrawals: _containers.RepeatedCompositeFieldContainer[Withdrawal]
    def __init__(self, header: _Optional[_Union[BlockHeader, _Mapping]] = ..., transactionHashes: _Optional[_Iterable[bytes]] = ..., fullTransactions: _Optional[_Iterable[_Union[Transaction, _Mapping]]] = ..., logs: _Optional[_Iterable[_Union[Log, _Mapping]]] = ..., withdrawals: _Optional[_Iterable[_Union[Withdrawal, _Mapping]]] = ...) -> None: ...

class TransactionRef(_message.Message):
    __slots__ = ("block", "transactionIndex", "transactionHash")
    BLOCK_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONINDEX_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONHASH_FIELD_NUMBER: _ClassVar[int]
    block: BlockRef
    transactionIndex: int
    transactionHash: bytes
    def __init__(self, block: _Optional[_Union[BlockRef, _Mapping]] = ..., transactionIndex: _Optional[int] = ..., transactionHash: _Optional[bytes] = ...) -> None: ...

class Transaction(_message.Message):
    __slots__ = ("hash", "nonce", "to", "value", "input", "type", "gasLimit", "gasPrice", "maxFeePerGas", "maxPriorityFeePerGas", "gasUsed", "effectiveGasPrice", "r", "s", "v", "yParity", "chainId", "blockNumber", "blockHash", "transactionIndex", "blockTimestamp", "accessList", "maxFeePerBlobGas", "blobVersionedHashes", "blobGasUsed", "blobGasPrice", "authorizationList", "l1Fee", "l1GasPrice", "l1GasUsed", "l1FeeScalar", "l1BlobBaseFee", "l1BlobBaseFeeScalar", "gatewayFee", "feeCurrency", "gatewayFeeRecipient", "beneficiary", "depositValue", "l1BaseFee", "maxSubmissionFee", "refundTo", "requestId", "retryData", "retryTo", "retryValue", "maxRefund", "submissionFeeRefund", "ticketId", "isSystemTx", "depositReceiptVersion", "sourceHash", "mint", "status", "cumulativeGasUsed", "contractAddress", "logsBloom")
    HASH_FIELD_NUMBER: _ClassVar[int]
    NONCE_FIELD_NUMBER: _ClassVar[int]
    FROM_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    GASLIMIT_FIELD_NUMBER: _ClassVar[int]
    GASPRICE_FIELD_NUMBER: _ClassVar[int]
    MAXFEEPERGAS_FIELD_NUMBER: _ClassVar[int]
    MAXPRIORITYFEEPERGAS_FIELD_NUMBER: _ClassVar[int]
    GASUSED_FIELD_NUMBER: _ClassVar[int]
    EFFECTIVEGASPRICE_FIELD_NUMBER: _ClassVar[int]
    R_FIELD_NUMBER: _ClassVar[int]
    S_FIELD_NUMBER: _ClassVar[int]
    V_FIELD_NUMBER: _ClassVar[int]
    YPARITY_FIELD_NUMBER: _ClassVar[int]
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONINDEX_FIELD_NUMBER: _ClassVar[int]
    BLOCKTIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    ACCESSLIST_FIELD_NUMBER: _ClassVar[int]
    MAXFEEPERBLOBGAS_FIELD_NUMBER: _ClassVar[int]
    BLOBVERSIONEDHASHES_FIELD_NUMBER: _ClassVar[int]
    BLOBGASUSED_FIELD_NUMBER: _ClassVar[int]
    BLOBGASPRICE_FIELD_NUMBER: _ClassVar[int]
    AUTHORIZATIONLIST_FIELD_NUMBER: _ClassVar[int]
    L1FEE_FIELD_NUMBER: _ClassVar[int]
    L1GASPRICE_FIELD_NUMBER: _ClassVar[int]
    L1GASUSED_FIELD_NUMBER: _ClassVar[int]
    L1FEESCALAR_FIELD_NUMBER: _ClassVar[int]
    L1BLOBBASEFEE_FIELD_NUMBER: _ClassVar[int]
    L1BLOBBASEFEESCALAR_FIELD_NUMBER: _ClassVar[int]
    GATEWAYFEE_FIELD_NUMBER: _ClassVar[int]
    FEECURRENCY_FIELD_NUMBER: _ClassVar[int]
    GATEWAYFEERECIPIENT_FIELD_NUMBER: _ClassVar[int]
    BENEFICIARY_FIELD_NUMBER: _ClassVar[int]
    DEPOSITVALUE_FIELD_NUMBER: _ClassVar[int]
    L1BASEFEE_FIELD_NUMBER: _ClassVar[int]
    MAXSUBMISSIONFEE_FIELD_NUMBER: _ClassVar[int]
    REFUNDTO_FIELD_NUMBER: _ClassVar[int]
    REQUESTID_FIELD_NUMBER: _ClassVar[int]
    RETRYDATA_FIELD_NUMBER: _ClassVar[int]
    RETRYTO_FIELD_NUMBER: _ClassVar[int]
    RETRYVALUE_FIELD_NUMBER: _ClassVar[int]
    MAXREFUND_FIELD_NUMBER: _ClassVar[int]
    SUBMISSIONFEEREFUND_FIELD_NUMBER: _ClassVar[int]
    TICKETID_FIELD_NUMBER: _ClassVar[int]
    ISSYSTEMTX_FIELD_NUMBER: _ClassVar[int]
    DEPOSITRECEIPTVERSION_FIELD_NUMBER: _ClassVar[int]
    SOURCEHASH_FIELD_NUMBER: _ClassVar[int]
    MINT_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    CUMULATIVEGASUSED_FIELD_NUMBER: _ClassVar[int]
    CONTRACTADDRESS_FIELD_NUMBER: _ClassVar[int]
    LOGSBLOOM_FIELD_NUMBER: _ClassVar[int]
    hash: bytes
    nonce: int
    to: bytes
    value: str
    input: bytes
    type: int
    gasLimit: int
    gasPrice: str
    maxFeePerGas: str
    maxPriorityFeePerGas: str
    gasUsed: int
    effectiveGasPrice: str
    r: bytes
    s: bytes
    v: bytes
    yParity: int
    chainId: int
    blockNumber: int
    blockHash: bytes
    transactionIndex: int
    blockTimestamp: int
    accessList: _containers.RepeatedCompositeFieldContainer[AccessListItem]
    maxFeePerBlobGas: str
    blobVersionedHashes: _containers.RepeatedScalarFieldContainer[bytes]
    blobGasUsed: int
    blobGasPrice: str
    authorizationList: _containers.RepeatedCompositeFieldContainer[AuthorizationListItem]
    l1Fee: str
    l1GasPrice: str
    l1GasUsed: str
    l1FeeScalar: float
    l1BlobBaseFee: str
    l1BlobBaseFeeScalar: int
    gatewayFee: str
    feeCurrency: bytes
    gatewayFeeRecipient: bytes
    beneficiary: bytes
    depositValue: str
    l1BaseFee: str
    maxSubmissionFee: str
    refundTo: bytes
    requestId: bytes
    retryData: bytes
    retryTo: bytes
    retryValue: str
    maxRefund: str
    submissionFeeRefund: str
    ticketId: bytes
    isSystemTx: bool
    depositReceiptVersion: str
    sourceHash: bytes
    mint: str
    status: int
    cumulativeGasUsed: int
    contractAddress: bytes
    logsBloom: bytes
    def __init__(self, hash: _Optional[bytes] = ..., nonce: _Optional[int] = ..., to: _Optional[bytes] = ..., value: _Optional[str] = ..., input: _Optional[bytes] = ..., type: _Optional[int] = ..., gasLimit: _Optional[int] = ..., gasPrice: _Optional[str] = ..., maxFeePerGas: _Optional[str] = ..., maxPriorityFeePerGas: _Optional[str] = ..., gasUsed: _Optional[int] = ..., effectiveGasPrice: _Optional[str] = ..., r: _Optional[bytes] = ..., s: _Optional[bytes] = ..., v: _Optional[bytes] = ..., yParity: _Optional[int] = ..., chainId: _Optional[int] = ..., blockNumber: _Optional[int] = ..., blockHash: _Optional[bytes] = ..., transactionIndex: _Optional[int] = ..., blockTimestamp: _Optional[int] = ..., accessList: _Optional[_Iterable[_Union[AccessListItem, _Mapping]]] = ..., maxFeePerBlobGas: _Optional[str] = ..., blobVersionedHashes: _Optional[_Iterable[bytes]] = ..., blobGasUsed: _Optional[int] = ..., blobGasPrice: _Optional[str] = ..., authorizationList: _Optional[_Iterable[_Union[AuthorizationListItem, _Mapping]]] = ..., l1Fee: _Optional[str] = ..., l1GasPrice: _Optional[str] = ..., l1GasUsed: _Optional[str] = ..., l1FeeScalar: _Optional[float] = ..., l1BlobBaseFee: _Optional[str] = ..., l1BlobBaseFeeScalar: _Optional[int] = ..., gatewayFee: _Optional[str] = ..., feeCurrency: _Optional[bytes] = ..., gatewayFeeRecipient: _Optional[bytes] = ..., beneficiary: _Optional[bytes] = ..., depositValue: _Optional[str] = ..., l1BaseFee: _Optional[str] = ..., maxSubmissionFee: _Optional[str] = ..., refundTo: _Optional[bytes] = ..., requestId: _Optional[bytes] = ..., retryData: _Optional[bytes] = ..., retryTo: _Optional[bytes] = ..., retryValue: _Optional[str] = ..., maxRefund: _Optional[str] = ..., submissionFeeRefund: _Optional[str] = ..., ticketId: _Optional[bytes] = ..., isSystemTx: _Optional[bool] = ..., depositReceiptVersion: _Optional[str] = ..., sourceHash: _Optional[bytes] = ..., mint: _Optional[str] = ..., status: _Optional[int] = ..., cumulativeGasUsed: _Optional[int] = ..., contractAddress: _Optional[bytes] = ..., logsBloom: _Optional[bytes] = ..., **kwargs) -> None: ...

class AccessListItem(_message.Message):
    __slots__ = ("address", "storageKeys")
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    STORAGEKEYS_FIELD_NUMBER: _ClassVar[int]
    address: bytes
    storageKeys: _containers.RepeatedScalarFieldContainer[bytes]
    def __init__(self, address: _Optional[bytes] = ..., storageKeys: _Optional[_Iterable[bytes]] = ...) -> None: ...

class Log(_message.Message):
    __slots__ = ("address", "topics", "data", "blockNumber", "blockHash", "transactionHash", "transactionIndex", "logIndex", "blockTimestamp")
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    TOPICS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONHASH_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONINDEX_FIELD_NUMBER: _ClassVar[int]
    LOGINDEX_FIELD_NUMBER: _ClassVar[int]
    BLOCKTIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    address: bytes
    topics: _containers.RepeatedScalarFieldContainer[bytes]
    data: bytes
    blockNumber: int
    blockHash: bytes
    transactionHash: bytes
    transactionIndex: int
    logIndex: int
    blockTimestamp: int
    def __init__(self, address: _Optional[bytes] = ..., topics: _Optional[_Iterable[bytes]] = ..., data: _Optional[bytes] = ..., blockNumber: _Optional[int] = ..., blockHash: _Optional[bytes] = ..., transactionHash: _Optional[bytes] = ..., transactionIndex: _Optional[int] = ..., logIndex: _Optional[int] = ..., blockTimestamp: _Optional[int] = ...) -> None: ...

class AuthorizationListItem(_message.Message):
    __slots__ = ("chainId", "address", "nonce", "r", "s", "yParity", "authority")
    CHAINID_FIELD_NUMBER: _ClassVar[int]
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    NONCE_FIELD_NUMBER: _ClassVar[int]
    R_FIELD_NUMBER: _ClassVar[int]
    S_FIELD_NUMBER: _ClassVar[int]
    YPARITY_FIELD_NUMBER: _ClassVar[int]
    AUTHORITY_FIELD_NUMBER: _ClassVar[int]
    chainId: int
    address: bytes
    nonce: int
    r: bytes
    s: bytes
    yParity: int
    authority: bytes
    def __init__(self, chainId: _Optional[int] = ..., address: _Optional[bytes] = ..., nonce: _Optional[int] = ..., r: _Optional[bytes] = ..., s: _Optional[bytes] = ..., yParity: _Optional[int] = ..., authority: _Optional[bytes] = ...) -> None: ...

class Withdrawal(_message.Message):
    __slots__ = ("index", "validatorIndex", "address", "amount")
    INDEX_FIELD_NUMBER: _ClassVar[int]
    VALIDATORINDEX_FIELD_NUMBER: _ClassVar[int]
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    AMOUNT_FIELD_NUMBER: _ClassVar[int]
    index: int
    validatorIndex: int
    address: bytes
    amount: int
    def __init__(self, index: _Optional[int] = ..., validatorIndex: _Optional[int] = ..., address: _Optional[bytes] = ..., amount: _Optional[int] = ...) -> None: ...

class Receipt(_message.Message):
    __slots__ = ("transactionHash", "blockNumber", "blockHash", "transactionIndex", "type", "to", "status", "gasUsed", "cumulativeGasUsed", "effectiveGasPrice", "logsBloom", "logs", "contractAddress", "root", "blockTimestamp", "blobGasUsed", "blobGasPrice", "timeboosted", "l1Fee", "l1GasUsed", "l1GasPrice", "l1FeeScalar", "l1BaseFeeScalar", "gasUsedForL1", "l1BlockNumber", "gatewayFee", "depositNonce", "depositReceiptVersion", "l1BlobBaseFee", "l1BlobBaseFeeScalar", "daFootprintGasScalar")
    TRANSACTIONHASH_FIELD_NUMBER: _ClassVar[int]
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONINDEX_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    FROM_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    GASUSED_FIELD_NUMBER: _ClassVar[int]
    CUMULATIVEGASUSED_FIELD_NUMBER: _ClassVar[int]
    EFFECTIVEGASPRICE_FIELD_NUMBER: _ClassVar[int]
    LOGSBLOOM_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    CONTRACTADDRESS_FIELD_NUMBER: _ClassVar[int]
    ROOT_FIELD_NUMBER: _ClassVar[int]
    BLOCKTIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    BLOBGASUSED_FIELD_NUMBER: _ClassVar[int]
    BLOBGASPRICE_FIELD_NUMBER: _ClassVar[int]
    TIMEBOOSTED_FIELD_NUMBER: _ClassVar[int]
    L1FEE_FIELD_NUMBER: _ClassVar[int]
    L1GASUSED_FIELD_NUMBER: _ClassVar[int]
    L1GASPRICE_FIELD_NUMBER: _ClassVar[int]
    L1FEESCALAR_FIELD_NUMBER: _ClassVar[int]
    L1BASEFEESCALAR_FIELD_NUMBER: _ClassVar[int]
    GASUSEDFORL1_FIELD_NUMBER: _ClassVar[int]
    L1BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    GATEWAYFEE_FIELD_NUMBER: _ClassVar[int]
    DEPOSITNONCE_FIELD_NUMBER: _ClassVar[int]
    DEPOSITRECEIPTVERSION_FIELD_NUMBER: _ClassVar[int]
    L1BLOBBASEFEE_FIELD_NUMBER: _ClassVar[int]
    L1BLOBBASEFEESCALAR_FIELD_NUMBER: _ClassVar[int]
    DAFOOTPRINTGASSCALAR_FIELD_NUMBER: _ClassVar[int]
    transactionHash: bytes
    blockNumber: int
    blockHash: bytes
    transactionIndex: int
    type: int
    to: bytes
    status: int
    gasUsed: int
    cumulativeGasUsed: int
    effectiveGasPrice: str
    logsBloom: bytes
    logs: _containers.RepeatedCompositeFieldContainer[Log]
    contractAddress: bytes
    root: bytes
    blockTimestamp: int
    blobGasUsed: int
    blobGasPrice: str
    timeboosted: bool
    l1Fee: str
    l1GasUsed: str
    l1GasPrice: str
    l1FeeScalar: float
    l1BaseFeeScalar: int
    gasUsedForL1: int
    l1BlockNumber: int
    gatewayFee: str
    depositNonce: str
    depositReceiptVersion: str
    l1BlobBaseFee: str
    l1BlobBaseFeeScalar: int
    daFootprintGasScalar: int
    def __init__(self, transactionHash: _Optional[bytes] = ..., blockNumber: _Optional[int] = ..., blockHash: _Optional[bytes] = ..., transactionIndex: _Optional[int] = ..., type: _Optional[int] = ..., to: _Optional[bytes] = ..., status: _Optional[int] = ..., gasUsed: _Optional[int] = ..., cumulativeGasUsed: _Optional[int] = ..., effectiveGasPrice: _Optional[str] = ..., logsBloom: _Optional[bytes] = ..., logs: _Optional[_Iterable[_Union[Log, _Mapping]]] = ..., contractAddress: _Optional[bytes] = ..., root: _Optional[bytes] = ..., blockTimestamp: _Optional[int] = ..., blobGasUsed: _Optional[int] = ..., blobGasPrice: _Optional[str] = ..., timeboosted: _Optional[bool] = ..., l1Fee: _Optional[str] = ..., l1GasUsed: _Optional[str] = ..., l1GasPrice: _Optional[str] = ..., l1FeeScalar: _Optional[float] = ..., l1BaseFeeScalar: _Optional[int] = ..., gasUsedForL1: _Optional[int] = ..., l1BlockNumber: _Optional[int] = ..., gatewayFee: _Optional[str] = ..., depositNonce: _Optional[str] = ..., depositReceiptVersion: _Optional[str] = ..., l1BlobBaseFee: _Optional[str] = ..., l1BlobBaseFeeScalar: _Optional[int] = ..., daFootprintGasScalar: _Optional[int] = ..., **kwargs) -> None: ...

class Trace(_message.Message):
    __slots__ = ("traceType", "callType", "to", "value", "input", "output", "gas", "gasUsed", "error", "subtraces", "traceAddress", "transactionHash", "transactionIndex", "blockNumber", "blockHash", "blockTimestamp", "type", "reverted")
    TRACETYPE_FIELD_NUMBER: _ClassVar[int]
    CALLTYPE_FIELD_NUMBER: _ClassVar[int]
    FROM_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    GAS_FIELD_NUMBER: _ClassVar[int]
    GASUSED_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    SUBTRACES_FIELD_NUMBER: _ClassVar[int]
    TRACEADDRESS_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONHASH_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONINDEX_FIELD_NUMBER: _ClassVar[int]
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    BLOCKTIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    REVERTED_FIELD_NUMBER: _ClassVar[int]
    traceType: TraceType
    callType: TraceCallType
    to: bytes
    value: str
    input: bytes
    output: bytes
    gas: int
    gasUsed: int
    error: str
    subtraces: int
    traceAddress: _containers.RepeatedScalarFieldContainer[int]
    transactionHash: bytes
    transactionIndex: int
    blockNumber: int
    blockHash: bytes
    blockTimestamp: int
    type: str
    reverted: bool
    def __init__(self, traceType: _Optional[_Union[TraceType, str]] = ..., callType: _Optional[_Union[TraceCallType, str]] = ..., to: _Optional[bytes] = ..., value: _Optional[str] = ..., input: _Optional[bytes] = ..., output: _Optional[bytes] = ..., gas: _Optional[int] = ..., gasUsed: _Optional[int] = ..., error: _Optional[str] = ..., subtraces: _Optional[int] = ..., traceAddress: _Optional[_Iterable[int]] = ..., transactionHash: _Optional[bytes] = ..., transactionIndex: _Optional[int] = ..., blockNumber: _Optional[int] = ..., blockHash: _Optional[bytes] = ..., blockTimestamp: _Optional[int] = ..., type: _Optional[str] = ..., reverted: _Optional[bool] = ..., **kwargs) -> None: ...

class NativeTransfer(_message.Message):
    __slots__ = ("to", "value", "transactionHash", "transactionIndex", "blockNumber", "blockHash", "traceAddress", "blockTimestamp", "type", "gas", "gasUsed", "input", "output", "error", "reverted")
    FROM_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONHASH_FIELD_NUMBER: _ClassVar[int]
    TRANSACTIONINDEX_FIELD_NUMBER: _ClassVar[int]
    BLOCKNUMBER_FIELD_NUMBER: _ClassVar[int]
    BLOCKHASH_FIELD_NUMBER: _ClassVar[int]
    TRACEADDRESS_FIELD_NUMBER: _ClassVar[int]
    BLOCKTIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    GAS_FIELD_NUMBER: _ClassVar[int]
    GASUSED_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    REVERTED_FIELD_NUMBER: _ClassVar[int]
    to: bytes
    value: str
    transactionHash: bytes
    transactionIndex: int
    blockNumber: int
    blockHash: bytes
    traceAddress: _containers.RepeatedScalarFieldContainer[int]
    blockTimestamp: int
    type: str
    gas: int
    gasUsed: int
    input: bytes
    output: bytes
    error: str
    reverted: bool
    def __init__(self, to: _Optional[bytes] = ..., value: _Optional[str] = ..., transactionHash: _Optional[bytes] = ..., transactionIndex: _Optional[int] = ..., blockNumber: _Optional[int] = ..., blockHash: _Optional[bytes] = ..., traceAddress: _Optional[_Iterable[int]] = ..., blockTimestamp: _Optional[int] = ..., type: _Optional[str] = ..., gas: _Optional[int] = ..., gasUsed: _Optional[int] = ..., input: _Optional[bytes] = ..., output: _Optional[bytes] = ..., error: _Optional[str] = ..., reverted: _Optional[bool] = ..., **kwargs) -> None: ...
