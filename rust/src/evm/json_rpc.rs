//! Ethereum JSON-RPC ⇄ BDS mapping.
//!
//! This module maps Ethereum JSON-RPC requests onto `bds.evm.RPCQueryService`
//! calls ([`map_request`]) and maps the BDS protobuf responses back into the
//! exact JSON-RPC result shapes a client expects ([`RpcQueryCall::execute`],
//! plus the standalone `*_to_json` helpers). It is the Rust counterpart of
//! this repository's Go `evm/json_rpc.go`, and every field-presence decision
//! below (omitted vs. `null` vs. always-present) is ported from that file —
//! see the inline comments on the handful of fields where the three
//! behaviors diverge.

use std::time::Duration;

use bytes::Bytes;
use serde_json::{Map, Value};

use super::rpc_query_service_client::RpcQueryServiceClient;
use super::{
    AccessListItem, AuthorizationListItem, BlockHeader, ChainIdRequest, GetBlockByHashRequest,
    GetBlockByNumberRequest, GetBlockReceiptsRequest, GetBlockReceiptsResponse, GetBlockResponse,
    GetLogsRequest, GetLogsResponse, GetTransactionByHashRequest, GetTransactionByHashResponse,
    GetTransactionReceiptRequest, GetTransactionReceiptResponse, Log, Receipt, TopicFilter,
    Transaction, Withdrawal,
};

/// Whether `method` has a `bds.evm.RPCQueryService` equivalent.
#[must_use]
pub fn method_supported(method: &str) -> bool {
    matches!(
        method,
        "eth_chainId"
            | "eth_getBlockByNumber"
            | "eth_getBlockByHash"
            | "eth_getTransactionByHash"
            | "eth_getTransactionReceipt"
            | "eth_getLogs"
            | "eth_getBlockReceipts"
    )
}

/// An error mapping a JSON-RPC request onto a BDS call.
#[derive(Debug, thiserror::Error)]
pub enum MapError {
    /// `method` has no `RPCQueryService` equivalent (see [`method_supported`]).
    #[error("method has no BDS RPC equivalent")]
    UnsupportedMethod,
    /// `params` could not be mapped onto the target BDS request message.
    #[error("params not expressible as a BDS request: {0}")]
    Unmappable(&'static str),
}

/// An error converting a JSON-RPC response value (a block/transaction/
/// receipt/log/withdrawal, or a `result` wrapping one of them) into its BDS
/// protobuf equivalent.
#[derive(Debug, thiserror::Error)]
pub enum FromJsonError {
    /// The value (or a nested field) wasn't the JSON shape a converter
    /// expected — e.g. an object where an array was required.
    #[error("expected {expected}, found {found}")]
    Shape {
        /// What the converter needed at this position.
        expected: &'static str,
        /// A human-readable description of what was actually there.
        found: &'static str,
    },
    /// A field required by the target message was absent or `null`.
    #[error("missing required field `{0}`")]
    MissingField(&'static str),
    /// A field held a JSON value of the wrong type (e.g. a bool where a
    /// string or array was required).
    #[error("field `{field}` has the wrong JSON type: {reason}")]
    WrongType {
        /// The JSON-RPC field name.
        field: &'static str,
        /// What was expected instead.
        reason: &'static str,
    },
    /// A field held a string that wasn't valid (optionally `0x`-prefixed) hex.
    #[error("field `{field}` is not valid hex: {value:?}")]
    InvalidHex {
        /// The JSON-RPC field name.
        field: &'static str,
        /// The offending value, for diagnostics.
        value: String,
    },
    /// A field held a value that wasn't a valid numeric quantity (hex or
    /// decimal, depending on the field).
    #[error("field `{field}` is not a valid numeric quantity: {value:?}")]
    InvalidNumber {
        /// The JSON-RPC field name.
        field: &'static str,
        /// The offending value, for diagnostics.
        value: String,
    },
}

/// A typed, ready-to-send `RPCQueryService` call.
#[derive(Debug)]
pub enum RpcQueryCall {
    /// `eth_chainId`.
    ChainId(ChainIdRequest),
    /// `eth_getBlockByNumber`.
    GetBlockByNumber(GetBlockByNumberRequest),
    /// `eth_getBlockByHash`.
    GetBlockByHash(GetBlockByHashRequest),
    /// `eth_getTransactionByHash`.
    GetTransactionByHash(GetTransactionByHashRequest),
    /// `eth_getTransactionReceipt`.
    GetTransactionReceipt(GetTransactionReceiptRequest),
    /// `eth_getLogs`.
    GetLogs(GetLogsRequest),
    /// `eth_getBlockReceipts`.
    GetBlockReceipts(GetBlockReceiptsRequest),
}

/// Maps a JSON-RPC `method` + `params` pair onto a [`RpcQueryCall`].
///
/// `chainId` / `chainGenesisHash` are deliberately never set on the built
/// request — the caller's connection already pins the chain.
///
/// # Errors
///
/// Returns [`MapError::UnsupportedMethod`] when `method` has no BDS
/// equivalent, or [`MapError::Unmappable`] when `params` doesn't have the
/// expected JSON-RPC shape for `method` (missing/mistyped fields, malformed
/// hex, wrong byte length, or — for `eth_getLogs` — a block *tag* / missing
/// bound with no `blockHash`, which this mapping deliberately refuses since
/// those queries are served elsewhere, never by the BDS server).
pub fn map_request(method: &str, params: &Value) -> Result<RpcQueryCall, MapError> {
    match method {
        "eth_chainId" => Ok(RpcQueryCall::ChainId(ChainIdRequest {})),
        "eth_getBlockByNumber" => map_get_block_by_number(params),
        "eth_getBlockByHash" => map_get_block_by_hash(params),
        "eth_getTransactionByHash" => map_get_transaction_by_hash(params),
        "eth_getTransactionReceipt" => map_get_transaction_receipt(params),
        "eth_getLogs" => map_get_logs(params),
        "eth_getBlockReceipts" => map_get_block_receipts(params),
        _ => Err(MapError::UnsupportedMethod),
    }
}

fn param(params: &Value, idx: usize) -> Option<&Value> {
    params.as_array().and_then(|a| a.get(idx))
}

fn map_get_block_by_number(params: &Value) -> Result<RpcQueryCall, MapError> {
    // Hex numbers and tags ("latest", "earliest", "pending", "finalized",
    // "safe") both pass through verbatim into the proto's string field — the
    // BDS server, not this mapper, is responsible for interpreting tags.
    let block_number = param(params, 0)
        .and_then(Value::as_str)
        .ok_or(MapError::Unmappable(
            "eth_getBlockByNumber: params[0] must be a block number or tag string",
        ))?
        .to_string();
    let include_transactions = param(params, 1).and_then(Value::as_bool).unwrap_or(false);
    Ok(RpcQueryCall::GetBlockByNumber(GetBlockByNumberRequest {
        block_number,
        include_transactions,
        chain_id: None,
        chain_genesis_hash: None,
    }))
}

fn map_get_block_by_hash(params: &Value) -> Result<RpcQueryCall, MapError> {
    let hash_str = param(params, 0)
        .and_then(Value::as_str)
        .ok_or(MapError::Unmappable(
            "eth_getBlockByHash: params[0] must be a 0x-hex block hash",
        ))?;
    let block_hash = parse_hex_bytes_exact(hash_str, 32)?;
    let include_transactions = param(params, 1).and_then(Value::as_bool).unwrap_or(false);
    Ok(RpcQueryCall::GetBlockByHash(GetBlockByHashRequest {
        block_hash: block_hash.into(),
        include_transactions,
        chain_id: None,
        chain_genesis_hash: None,
    }))
}

fn map_get_transaction_by_hash(params: &Value) -> Result<RpcQueryCall, MapError> {
    let hash_str = param(params, 0)
        .and_then(Value::as_str)
        .ok_or(MapError::Unmappable(
            "eth_getTransactionByHash: params[0] must be a 0x-hex transaction hash",
        ))?;
    let transaction_hash = parse_hex_bytes_exact(hash_str, 32)?;
    Ok(RpcQueryCall::GetTransactionByHash(
        GetTransactionByHashRequest {
            transaction_hash: transaction_hash.into(),
            chain_id: None,
            chain_genesis_hash: None,
        },
    ))
}

fn map_get_transaction_receipt(params: &Value) -> Result<RpcQueryCall, MapError> {
    let hash_str = param(params, 0)
        .and_then(Value::as_str)
        .ok_or(MapError::Unmappable(
            "eth_getTransactionReceipt: params[0] must be a 0x-hex transaction hash",
        ))?;
    let transaction_hash = parse_hex_bytes_exact(hash_str, 32)?;
    Ok(RpcQueryCall::GetTransactionReceipt(
        GetTransactionReceiptRequest {
            transaction_hash: transaction_hash.into(),
            chain_id: None,
            chain_genesis_hash: None,
        },
    ))
}

fn map_get_block_receipts(params: &Value) -> Result<RpcQueryCall, MapError> {
    let p0 = param(params, 0).ok_or(MapError::Unmappable(
        "eth_getBlockReceipts: params[0] is required",
    ))?;
    let mut req = GetBlockReceiptsRequest {
        block_number: None,
        block_hash: None,
        chain_id: None,
        chain_genesis_hash: None,
    };
    match p0 {
        Value::String(s) => req.block_number = Some(s.clone()),
        Value::Object(obj) => {
            if let Some(h) = obj.get("blockHash").and_then(Value::as_str) {
                req.block_hash = Some(parse_hex_bytes_exact(h, 32)?.into());
            } else if let Some(n) = obj.get("blockNumber").and_then(Value::as_str) {
                req.block_number = Some(n.to_string());
            } else {
                return Err(MapError::Unmappable(
                    "eth_getBlockReceipts: object param needs a string blockHash or blockNumber",
                ));
            }
        }
        _ => {
            return Err(MapError::Unmappable(
                "eth_getBlockReceipts: params[0] must be a block number/tag string or an object",
            ));
        }
    }
    Ok(RpcQueryCall::GetBlockReceipts(req))
}

fn map_get_logs(params: &Value) -> Result<RpcQueryCall, MapError> {
    let filter = param(params, 0)
        .and_then(Value::as_object)
        .ok_or(MapError::Unmappable(
            "eth_getLogs: params[0] must be a filter object",
        ))?;

    let block_hash = filter
        .get("blockHash")
        .and_then(Value::as_str)
        .map(|s| parse_hex_bytes_exact(s, 32))
        .transpose()?;

    // fromBlock/toBlock and blockHash are mutually exclusive; when a
    // blockHash is given the range bounds are simply not read.
    let (from_block, to_block) = if block_hash.is_some() {
        (None, None)
    } else {
        (
            Some(parse_block_bound(filter.get("fromBlock"))?),
            Some(parse_block_bound(filter.get("toBlock"))?),
        )
    };

    let addresses = match filter.get("address") {
        None | Some(Value::Null) => Vec::new(),
        Some(Value::String(s)) => vec![Bytes::from(parse_hex_bytes_exact(s, 20)?)],
        Some(Value::Array(arr)) => arr
            .iter()
            .map(|v| {
                v.as_str()
                    .ok_or(MapError::Unmappable(
                        "eth_getLogs: address array entries must be strings",
                    ))
                    .and_then(|s| parse_hex_bytes_exact(s, 20))
                    .map(Bytes::from)
            })
            .collect::<Result<Vec<_>, _>>()?,
        Some(_) => {
            return Err(MapError::Unmappable(
                "eth_getLogs: address must be a string or an array of strings",
            ));
        }
    };

    let topics = build_topic_filters(filter.get("topics"))?;

    Ok(RpcQueryCall::GetLogs(GetLogsRequest {
        from_block,
        to_block,
        addresses,
        topics,
        block_hash: block_hash.map(Bytes::from),
        chain_id: None,
        chain_genesis_hash: None,
    }))
}

fn parse_block_bound(v: Option<&Value>) -> Result<u64, MapError> {
    // A tag ("latest"/"earliest"/...) or a missing bound is refused here —
    // those queries are forwarded to a different backend, never to the BDS
    // server — so only a 0x-hex quantity is accepted.
    let s = v.and_then(Value::as_str).ok_or(MapError::Unmappable(
        "eth_getLogs: fromBlock/toBlock must be a 0x-hex quantity unless blockHash is set",
    ))?;
    parse_hex_u64(s)
}

/// Converts the JSON-RPC `topics` array (each entry `null` | string | array
/// of strings) into BDS `TopicFilter`s. A `null` entry is a wildcard at that
/// position and MUST still emit an empty `TopicFilter` — dropping it would
/// shift every later position left, silently changing which topic index
/// each filter matches. Mirrors `buildTopicFilters` in erpc's
/// `grpc_bds_client.go`.
fn build_topic_filters(topics: Option<&Value>) -> Result<Vec<TopicFilter>, MapError> {
    let Some(topics) = topics else {
        return Ok(Vec::new());
    };
    if topics.is_null() {
        return Ok(Vec::new());
    }
    let arr = topics
        .as_array()
        .ok_or(MapError::Unmappable("eth_getLogs: topics must be an array"))?;

    let mut out = Vec::with_capacity(arr.len());
    for entry in arr {
        let values = match entry {
            Value::Null => Vec::new(),
            Value::String(s) => vec![Bytes::from(parse_hex_bytes(s)?)],
            Value::Array(inner) => inner
                .iter()
                .map(|v| {
                    v.as_str()
                        .ok_or(MapError::Unmappable(
                            "eth_getLogs: topic array entries must be strings",
                        ))
                        .and_then(|s| parse_hex_bytes(s).map(Bytes::from))
                })
                .collect::<Result<Vec<_>, _>>()?,
            _ => {
                return Err(MapError::Unmappable(
                    "eth_getLogs: each topics entry must be null, a string, or an array of strings",
                ));
            }
        };
        out.push(TopicFilter { values });
    }
    Ok(out)
}

// --- inverse: proto request → JSON-RPC --------------------------------

/// The JSON-RPC `(method, positional params)` equivalent of a typed call —
/// what a gateway fronting a JSON-RPC node sends upstream. Exact inverse of
/// [`map_request`]: for any `call` built by `map_request(method, params)`,
/// re-running `map_request` on the method/params pair returned here
/// reproduces an equivalent call (see the round-trip tests below).
#[must_use]
pub fn call_to_json_rpc(call: &RpcQueryCall) -> (&'static str, Value) {
    match call {
        RpcQueryCall::ChainId(_) => ("eth_chainId", Value::Array(Vec::new())),
        RpcQueryCall::GetBlockByNumber(req) => (
            "eth_getBlockByNumber",
            Value::Array(vec![
                Value::String(req.block_number.clone()),
                Value::Bool(req.include_transactions),
            ]),
        ),
        RpcQueryCall::GetBlockByHash(req) => (
            "eth_getBlockByHash",
            Value::Array(vec![
                Value::String(bytes_to_hex(&req.block_hash)),
                Value::Bool(req.include_transactions),
            ]),
        ),
        RpcQueryCall::GetTransactionByHash(req) => (
            "eth_getTransactionByHash",
            Value::Array(vec![Value::String(bytes_to_hex(&req.transaction_hash))]),
        ),
        RpcQueryCall::GetTransactionReceipt(req) => (
            "eth_getTransactionReceipt",
            Value::Array(vec![Value::String(bytes_to_hex(&req.transaction_hash))]),
        ),
        RpcQueryCall::GetLogs(req) => (
            "eth_getLogs",
            Value::Array(vec![get_logs_filter_to_json(req)]),
        ),
        RpcQueryCall::GetBlockReceipts(req) => (
            "eth_getBlockReceipts",
            Value::Array(vec![get_block_receipts_param_to_json(req)]),
        ),
    }
}

fn get_logs_filter_to_json(req: &GetLogsRequest) -> Value {
    let mut o = Map::new();
    if let Some(h) = &req.block_hash {
        // Mutually exclusive with fromBlock/toBlock — map_get_logs never
        // reads the range bounds once blockHash is present, so they're
        // simply not emitted here either.
        o.insert("blockHash".into(), Value::String(bytes_to_hex(h)));
    } else {
        if let Some(v) = req.from_block {
            o.insert("fromBlock".into(), Value::String(quantity_hex(v)));
        }
        if let Some(v) = req.to_block {
            o.insert("toBlock".into(), Value::String(quantity_hex(v)));
        }
    }
    if !req.addresses.is_empty() {
        o.insert(
            "address".into(),
            Value::Array(
                req.addresses
                    .iter()
                    .map(|a| Value::String(bytes_to_hex(a)))
                    .collect(),
            ),
        );
    }
    if !req.topics.is_empty() {
        o.insert(
            "topics".into(),
            Value::Array(
                req.topics
                    .iter()
                    .map(|t| {
                        if t.values.is_empty() {
                            // A wildcard position — mirrors build_topic_filters's
                            // null handling, which must not shift later indices.
                            Value::Null
                        } else {
                            Value::Array(
                                t.values
                                    .iter()
                                    .map(|v| Value::String(bytes_to_hex(v)))
                                    .collect(),
                            )
                        }
                    })
                    .collect(),
            ),
        );
    }
    Value::Object(o)
}

fn get_block_receipts_param_to_json(req: &GetBlockReceiptsRequest) -> Value {
    if let Some(h) = &req.block_hash {
        // Wrapped in an object (rather than a bare hex string) so that
        // map_get_block_receipts routes it back into blockHash instead of
        // misreading it as a blockNumber tag/hex string.
        let mut o = Map::new();
        o.insert("blockHash".into(), Value::String(bytes_to_hex(h)));
        Value::Object(o)
    } else {
        Value::String(req.block_number.clone().unwrap_or_default())
    }
}

// --- hex parsing (request side) --------------------------------------------

/// Decodes already-prefix-stripped hex digits into bytes, padding an
/// odd-length input with a leading zero nibble — mirrors Go's `HexToBytes`
/// (some RPC nodes emit values like "0x1" instead of "0x01"). Shared by both
/// directions of this module: the request-side [`parse_hex_bytes`] wraps the
/// error as [`MapError`], the response-side [`hex_to_bytes`] wraps it as
/// [`FromJsonError`].
fn hex_digit_pairs_to_bytes(hex_digits: &str) -> Result<Vec<u8>, &'static str> {
    if !hex_digits.bytes().all(|b| b.is_ascii_hexdigit()) {
        return Err("invalid hex digits");
    }
    let owned;
    let digits: &str = if hex_digits.len() % 2 == 0 {
        hex_digits
    } else {
        owned = format!("0{hex_digits}");
        &owned
    };
    digits
        .as_bytes()
        .chunks_exact(2)
        .map(|chunk| {
            let s = std::str::from_utf8(chunk).unwrap_or_default();
            u8::from_str_radix(s, 16).map_err(|_| "invalid hex digits")
        })
        .collect()
}

fn parse_hex_bytes(s: &str) -> Result<Vec<u8>, MapError> {
    let stripped = s
        .strip_prefix("0x")
        .or_else(|| s.strip_prefix("0X"))
        .unwrap_or(s);
    hex_digit_pairs_to_bytes(stripped).map_err(MapError::Unmappable)
}

fn parse_hex_bytes_exact(s: &str, len: usize) -> Result<Vec<u8>, MapError> {
    let bytes = parse_hex_bytes(s)?;
    if bytes.len() == len {
        Ok(bytes)
    } else {
        Err(MapError::Unmappable("hex value has the wrong byte length"))
    }
}

fn parse_hex_u64(s: &str) -> Result<u64, MapError> {
    let stripped = s
        .strip_prefix("0x")
        .or_else(|| s.strip_prefix("0X"))
        .ok_or(MapError::Unmappable("expected a 0x-prefixed hex quantity"))?;
    if stripped.is_empty() {
        return Ok(0);
    }
    u64::from_str_radix(stripped, 16).map_err(|_| MapError::Unmappable("invalid hex quantity"))
}

// --- execute -----------------------------------------------------------

impl RpcQueryCall {
    /// Sends this call over `client`.
    ///
    /// `client` is generic over the underlying tonic service — the same
    /// bounds the generated `RpcQueryServiceClient` impl uses — so a caller
    /// may pass a client wrapped in tower middleware (most usefully
    /// [`RpcQueryServiceClient::with_interceptor`], to inject per-call
    /// metadata such as a W3C `traceparent`). A plain
    /// `RpcQueryServiceClient<Channel>` still works unchanged.
    ///
    /// `timeout`, when set, becomes the gRPC `grpc-timeout` deadline for the
    /// call (via [`tonic::Request::set_timeout`]).
    ///
    /// Two [`tracing`] spans break the call down for a traced caller —
    /// inert (near-zero cost) without an interested subscriber:
    ///
    /// - **`bds.rpc`** — the gRPC await: request send, server time, response
    ///   transfer, and prost decode. A server that extracts `traceparent`
    ///   nests its server span under this one, which splits wire from server
    ///   time in the trace.
    /// - **`bds.to_json`** — BDS proto structs → JSON-RPC [`Value`] mapping:
    ///   pure CPU in the caller's process, and the only other place a
    ///   multi-MB block response spends time on this path.
    ///
    /// `Ok(Some(_))` carries the JSON-RPC-shaped result. `Ok(None)` means the
    /// server answered but the block/transaction/receipt was absent — the
    /// JSON-RPC `null` "not found" case. `eth_getLogs` and
    /// `eth_getBlockReceipts` always return `Ok(Some(_))`, using an empty
    /// array for a zero-hit query, since an empty result is a valid answer
    /// rather than a miss. Any `tonic::Status` from the transport is
    /// propagated untouched.
    ///
    /// # Errors
    ///
    /// Returns the `tonic::Status` produced by the underlying gRPC call.
    pub async fn execute<T>(
        self,
        client: &mut RpcQueryServiceClient<T>,
        timeout: Option<Duration>,
    ) -> Result<Option<Value>, tonic::Status>
    where
        T: tonic::client::GrpcService<tonic::body::Body>,
        T::Error: Into<tonic::codegen::StdError>,
        T::ResponseBody: tonic::codegen::Body<Data = Bytes> + Send + 'static,
        <T::ResponseBody as tonic::codegen::Body>::Error: Into<tonic::codegen::StdError> + Send,
    {
        use tracing::Instrument as _;
        match self {
            Self::ChainId(req) => {
                let resp = client
                    .chain_id(with_timeout(req, timeout))
                    .instrument(rpc_span())
                    .await?
                    .into_inner();
                Ok(to_json_span().in_scope(|| Some(Value::String(quantity_hex(resp.chain_id)))))
            }
            Self::GetBlockByNumber(req) => {
                let resp = client
                    .get_block_by_number(with_timeout(req, timeout))
                    .instrument(rpc_span())
                    .await?
                    .into_inner();
                Ok(to_json_span().in_scope(|| get_block_response_to_json(&resp)))
            }
            Self::GetBlockByHash(req) => {
                let resp = client
                    .get_block_by_hash(with_timeout(req, timeout))
                    .instrument(rpc_span())
                    .await?
                    .into_inner();
                Ok(to_json_span().in_scope(|| get_block_response_to_json(&resp)))
            }
            Self::GetTransactionByHash(req) => {
                let resp = client
                    .get_transaction_by_hash(with_timeout(req, timeout))
                    .instrument(rpc_span())
                    .await?
                    .into_inner();
                Ok(to_json_span().in_scope(|| resp.transaction.as_ref().map(transaction_to_json)))
            }
            Self::GetTransactionReceipt(req) => {
                let resp = client
                    .get_transaction_receipt(with_timeout(req, timeout))
                    .instrument(rpc_span())
                    .await?
                    .into_inner();
                Ok(to_json_span().in_scope(|| resp.receipt.as_ref().map(receipt_to_json)))
            }
            Self::GetLogs(req) => {
                let resp = client
                    .get_logs(with_timeout(req, timeout))
                    .instrument(rpc_span())
                    .await?
                    .into_inner();
                Ok(to_json_span()
                    .in_scope(|| Some(Value::Array(resp.logs.iter().map(log_to_json).collect()))))
            }
            Self::GetBlockReceipts(req) => {
                let resp = client
                    .get_block_receipts(with_timeout(req, timeout))
                    .instrument(rpc_span())
                    .await?
                    .into_inner();
                Ok(to_json_span().in_scope(|| {
                    Some(Value::Array(
                        resp.receipts.iter().map(receipt_to_json).collect(),
                    ))
                }))
            }
        }
    }
}

/// Span over the raw gRPC await inside [`RpcQueryCall::execute`] — see its
/// doc for what the span covers.
fn rpc_span() -> tracing::Span {
    tracing::info_span!("bds.rpc")
}

/// Span over the proto → JSON-RPC [`Value`] mapping inside
/// [`RpcQueryCall::execute`].
fn to_json_span() -> tracing::Span {
    tracing::info_span!("bds.to_json")
}

fn with_timeout<T>(msg: T, timeout: Option<Duration>) -> tonic::Request<T> {
    let mut req = tonic::Request::new(msg);
    if let Some(t) = timeout {
        req.set_timeout(t);
    }
    req
}

fn get_block_response_to_json(resp: &GetBlockResponse) -> Option<Value> {
    let header = resp.block.as_ref()?;
    Some(block_to_json(
        header,
        &resp.transactions,
        &resp.full_transactions,
        &resp.withdrawals,
    ))
}

// --- response mapping ----------------------------------------------------

/// Converts a BDS `BlockHeader` — plus its transactions (either hashes or
/// full objects, whichever the server populated) and withdrawals — into the
/// exact JSON-RPC block object shape. Mirrors `evm.BlockToJsonRpc` in
/// `evm/json_rpc.go`.
#[must_use]
pub fn block_to_json(
    header: &BlockHeader,
    transaction_hashes: &[Bytes],
    full_transactions: &[Transaction],
    withdrawals: &[Withdrawal],
) -> Value {
    let mut o = Map::new();
    o.insert("number".into(), Value::String(quantity_hex(header.number)));
    o.insert("hash".into(), Value::String(bytes_to_hex(&header.hash)));
    o.insert(
        "parentHash".into(),
        Value::String(bytes_to_hex(&header.parent_hash)),
    );
    o.insert(
        "sha3Uncles".into(),
        Value::String(bytes_to_hex(&header.sha3_uncles)),
    );
    o.insert(
        "logsBloom".into(),
        Value::String(bytes_to_hex(&header.logs_bloom)),
    );
    o.insert(
        "transactionsRoot".into(),
        Value::String(bytes_to_hex(&header.transactions_root)),
    );
    o.insert(
        "stateRoot".into(),
        Value::String(bytes_to_hex(&header.state_root)),
    );
    o.insert(
        "receiptsRoot".into(),
        Value::String(bytes_to_hex(&header.receipts_root)),
    );
    o.insert("miner".into(), Value::String(bytes_to_hex(&header.miner)));
    o.insert(
        "extraData".into(),
        Value::String(bytes_to_hex(&header.extra_data)),
    );
    o.insert("size".into(), Value::String(quantity_hex(header.size)));
    o.insert(
        "gasLimit".into(),
        Value::String(quantity_hex(header.gas_limit)),
    );
    o.insert(
        "gasUsed".into(),
        Value::String(quantity_hex(header.gas_used)),
    );
    o.insert(
        "timestamp".into(),
        Value::String(quantity_hex(header.timestamp)),
    );

    if let Some(nonce) = header.nonce {
        // Unlike every other quantity field, the nonce is zero-padded to a
        // fixed 16 hex digits (8 bytes) — `fmt.Sprintf("0x%016x", ...)`.
        o.insert("nonce".into(), Value::String(format!("0x{nonce:016x}")));
    }
    insert_decimal_omit(&mut o, "baseFeePerGas", header.base_fee_per_gas.as_deref());
    insert_decimal_omit(&mut o, "difficulty", header.difficulty.as_deref());
    insert_decimal_omit(
        &mut o,
        "totalDifficulty",
        header.total_difficulty.as_deref(),
    );
    if let Some(v) = header.mix_hash.as_deref() {
        o.insert("mixHash".into(), Value::String(bytes_to_hex(v)));
    }
    if let Some(v) = header.withdrawals_root.as_deref() {
        o.insert("withdrawalsRoot".into(), Value::String(bytes_to_hex(v)));
    }
    if let Some(v) = header.requests_hash.as_deref() {
        o.insert("requestsHash".into(), Value::String(bytes_to_hex(v)));
    }
    if let Some(v) = header.blob_gas_used {
        o.insert("blobGasUsed".into(), Value::String(quantity_hex(v)));
    }
    if let Some(v) = header.excess_blob_gas {
        o.insert("excessBlobGas".into(), Value::String(quantity_hex(v)));
    }
    if let Some(v) = header.parent_beacon_block_root.as_deref() {
        o.insert(
            "parentBeaconBlockRoot".into(),
            Value::String(bytes_to_hex(v)),
        );
    }
    if let Some(v) = header.l1_block_number {
        o.insert("l1BlockNumber".into(), Value::String(quantity_hex(v)));
    }
    if let Some(v) = header.send_count {
        o.insert("sendCount".into(), Value::String(quantity_hex(v)));
    }
    if let Some(v) = header.send_root.as_deref() {
        o.insert("sendRoot".into(), Value::String(bytes_to_hex(v)));
    }
    if let Some(v) = header.epoch {
        o.insert("epoch".into(), Value::String(quantity_hex(v)));
    }
    if let Some(v) = header.slot {
        o.insert("slot".into(), Value::String(quantity_hex(v)));
    }
    if let Some(v) = header.proposer_index {
        o.insert("proposerIndex".into(), Value::String(quantity_hex(v)));
    }
    if let Some(v) = header.transaction_count {
        o.insert(
            "transactionCount".into(),
            Value::String(quantity_hex(u64::from(v))),
        );
    }
    if let Some(v) = &header.proposer_public_key {
        o.insert("proposerPublicKey".into(), Value::String(v.clone()));
    }
    // `withdrawals` is a `repeated` proto field, which cannot distinguish
    // "not applicable" (pre-Shanghai chain) from "explicitly empty" on the
    // wire — an empty repeated field is simply never sent. json_rpc.go's
    // `withdrawals != nil` check is therefore equivalent to "non-empty" once
    // it has round-tripped through protobuf, which is what we check here.
    if !withdrawals.is_empty() {
        o.insert(
            "withdrawals".into(),
            Value::Array(withdrawals.iter().map(withdrawal_to_json).collect()),
        );
    }
    if let Some(v) = header.canonical_rlp.as_deref() {
        o.insert("canonicalRlp".into(), Value::String(bytes_to_hex(v)));
    }

    // Always present (possibly `[]`), unlike the optional fields above.
    o.insert(
        "uncles".into(),
        Value::Array(
            header
                .uncles
                .iter()
                .map(|u| Value::String(bytes_to_hex(u)))
                .collect(),
        ),
    );

    let transactions = if !full_transactions.is_empty() {
        Value::Array(full_transactions.iter().map(transaction_to_json).collect())
    } else if !transaction_hashes.is_empty() {
        Value::Array(
            transaction_hashes
                .iter()
                .map(|h| Value::String(bytes_to_hex(h)))
                .collect(),
        )
    } else {
        Value::Array(Vec::new())
    };
    o.insert("transactions".into(), transactions);

    Value::Object(o)
}

/// Converts a BDS `Transaction` into the exact JSON-RPC transaction object
/// shape. Mirrors `evm.TransactionToJsonRpc` in `evm/json_rpc.go`.
#[must_use]
pub fn transaction_to_json(tx: &Transaction) -> Value {
    let mut o = Map::new();
    o.insert("hash".into(), Value::String(bytes_to_hex(&tx.hash)));
    o.insert("nonce".into(), Value::String(quantity_hex(tx.nonce)));
    o.insert("from".into(), Value::String(bytes_to_hex(&tx.from)));
    o.insert("gas".into(), Value::String(quantity_hex(tx.gas_limit)));
    o.insert("input".into(), Value::String(bytes_to_hex(&tx.input)));
    o.insert(
        "type".into(),
        Value::String(quantity_hex(u64::from(tx.r#type))),
    );

    o.insert(
        "to".into(),
        nonempty(&tx.to).map_or(Value::Null, |b| Value::String(bytes_to_hex(b))),
    );

    if let Some(v) = tx.block_hash.as_deref() {
        o.insert("blockHash".into(), Value::String(bytes_to_hex(v)));
    }
    if let Some(v) = tx.block_number {
        o.insert("blockNumber".into(), Value::String(quantity_hex(v)));
    }
    if let Some(v) = tx.transaction_index {
        o.insert(
            "transactionIndex".into(),
            Value::String(quantity_hex(u64::from(v))),
        );
    }
    if let Some(v) = tx.block_timestamp {
        o.insert("blockTimestamp".into(), Value::String(quantity_hex(v)));
    }

    // Unlike the optional string fields below, `value` is a required proto
    // string; json_rpc.go omits the key entirely (not "0x0") when it's "".
    if !tx.value.is_empty() {
        if let Some(v) = decimal_string_to_hex(&tx.value) {
            o.insert("value".into(), Value::String(v));
        }
    }
    insert_decimal_omit(&mut o, "mint", tx.mint.as_deref());
    insert_decimal_omit(&mut o, "gasPrice", tx.gas_price.as_deref());
    insert_decimal_omit(&mut o, "maxFeePerGas", tx.max_fee_per_gas.as_deref());
    insert_decimal_omit(
        &mut o,
        "maxPriorityFeePerGas",
        tx.max_priority_fee_per_gas.as_deref(),
    );

    if let Some(v) = tx.gas_used {
        o.insert("gasUsed".into(), Value::String(quantity_hex(v)));
    }
    insert_decimal_omit(
        &mut o,
        "effectiveGasPrice",
        tx.effective_gas_price.as_deref(),
    );

    // r/s are non-optional `bytes` on the wire, but an absent bytes field and
    // an explicitly-empty one decode identically in proto3 — mirror Go's
    // `!= nil` check (on a []byte, itself only ever non-nil when non-empty
    // after a protobuf round-trip) as "non-empty" here.
    if !tx.r.is_empty() {
        o.insert("r".into(), Value::String(bytes_to_hex_fixed(&tx.r, 32)));
    }
    if !tx.s.is_empty() {
        o.insert("s".into(), Value::String(bytes_to_hex_fixed(&tx.s, 32)));
    }
    if let Some(v) = &tx.v {
        o.insert("v".into(), Value::String(bytes_to_quantity_hex(v)));
    }

    if let Some(v) = nonempty(&tx.source_hash) {
        o.insert("sourceHash".into(), Value::String(bytes_to_hex(v)));
    }

    // chainId/yParity are always present (null when absent), unlike most
    // other optional fields on this message.
    o.insert(
        "chainId".into(),
        tx.chain_id
            .map_or(Value::Null, |v| Value::String(quantity_hex(v))),
    );
    o.insert(
        "yParity".into(),
        tx.y_parity
            .map_or(Value::Null, |v| Value::String(quantity_hex(u64::from(v)))),
    );

    // Always present (possibly `[]`) — json_rpc.go pre-allocates this array
    // unconditionally, unlike blobVersionedHashes/authorizationList below.
    o.insert(
        "accessList".into(),
        Value::Array(
            tx.access_list
                .iter()
                .map(access_list_item_to_json)
                .collect(),
        ),
    );

    insert_decimal_omit(
        &mut o,
        "maxFeePerBlobGas",
        tx.max_fee_per_blob_gas.as_deref(),
    );
    if !tx.blob_versioned_hashes.is_empty() {
        o.insert(
            "blobVersionedHashes".into(),
            Value::Array(
                tx.blob_versioned_hashes
                    .iter()
                    .map(|h| Value::String(bytes_to_hex(h)))
                    .collect(),
            ),
        );
    }
    if let Some(v) = tx.blob_gas_used {
        o.insert("blobGasUsed".into(), Value::String(quantity_hex(v)));
    }
    insert_decimal_omit(&mut o, "blobGasPrice", tx.blob_gas_price.as_deref());

    if !tx.authorization_list.is_empty() {
        o.insert(
            "authorizationList".into(),
            Value::Array(
                tx.authorization_list
                    .iter()
                    .map(authorization_list_item_to_json)
                    .collect(),
            ),
        );
    }

    // l1Fee is always present (null on absence/parse-failure); everything
    // else in this L2 fee breakdown is simply omitted on absence.
    insert_decimal_or_null(&mut o, "l1Fee", tx.l1_fee.as_deref());
    insert_decimal_omit(&mut o, "l1GasUsed", tx.l1_gas_used.as_deref());
    insert_decimal_omit(&mut o, "l1GasPrice", tx.l1_gas_price.as_deref());
    if let Some(v) = tx.l1_fee_scalar {
        o.insert("l1FeeScalar".into(), Value::from(v));
    }
    insert_decimal_omit(&mut o, "l1BlobBaseFee", tx.l1_blob_base_fee.as_deref());
    if let Some(v) = tx.l1_blob_base_fee_scalar {
        o.insert("l1BlobBaseFeeScalar".into(), Value::String(quantity_hex(v)));
    }
    insert_decimal_omit(&mut o, "gatewayFee", tx.gateway_fee.as_deref());
    if let Some(v) = nonempty(&tx.fee_currency) {
        o.insert("feeCurrency".into(), Value::String(bytes_to_hex(v)));
    }
    if let Some(v) = nonempty(&tx.gateway_fee_recipient) {
        o.insert("gatewayFeeRecipient".into(), Value::String(bytes_to_hex(v)));
    }

    // Arbitrum retryable ticket fields.
    if let Some(v) = nonempty(&tx.beneficiary) {
        o.insert("beneficiary".into(), Value::String(bytes_to_hex(v)));
    }
    insert_decimal_omit(&mut o, "depositValue", tx.deposit_value.as_deref());
    insert_decimal_omit(&mut o, "l1BaseFee", tx.l1_base_fee.as_deref());
    insert_decimal_omit(&mut o, "maxSubmissionFee", tx.max_submission_fee.as_deref());
    if let Some(v) = nonempty(&tx.refund_to) {
        o.insert("refundTo".into(), Value::String(bytes_to_hex(v)));
    }
    if let Some(v) = nonempty(&tx.request_id) {
        o.insert("requestId".into(), Value::String(bytes_to_hex(v)));
    }
    if let Some(v) = nonempty(&tx.retry_data) {
        o.insert("retryData".into(), Value::String(bytes_to_hex(v)));
    }
    if let Some(v) = nonempty(&tx.retry_to) {
        o.insert("retryTo".into(), Value::String(bytes_to_hex(v)));
    }
    insert_decimal_omit(&mut o, "retryValue", tx.retry_value.as_deref());
    insert_decimal_omit(&mut o, "maxRefund", tx.max_refund.as_deref());
    insert_decimal_omit(
        &mut o,
        "submissionFeeRefund",
        tx.submission_fee_refund.as_deref(),
    );
    if let Some(v) = nonempty(&tx.ticket_id) {
        o.insert("ticketId".into(), Value::String(bytes_to_hex(v)));
    }

    // Base-specific fields.
    if let Some(v) = tx.is_system_tx {
        o.insert("isSystemTx".into(), Value::Bool(v));
    }
    insert_decimal_omit(
        &mut o,
        "depositReceiptVersion",
        tx.deposit_receipt_version.as_deref(),
    );

    Value::Object(o)
}

fn access_list_item_to_json(item: &AccessListItem) -> Value {
    let mut o = Map::new();
    o.insert("address".into(), Value::String(bytes_to_hex(&item.address)));
    o.insert(
        "storageKeys".into(),
        Value::Array(
            item.storage_keys
                .iter()
                .map(|k| Value::String(bytes_to_hex(k)))
                .collect(),
        ),
    );
    Value::Object(o)
}

fn authorization_list_item_to_json(item: &AuthorizationListItem) -> Value {
    let mut o = Map::new();
    o.insert("chainId".into(), Value::String(quantity_hex(item.chain_id)));
    o.insert("address".into(), Value::String(bytes_to_hex(&item.address)));
    o.insert("nonce".into(), Value::String(quantity_hex(item.nonce)));
    o.insert("r".into(), Value::String(bytes_to_hex_fixed(&item.r, 32)));
    o.insert("s".into(), Value::String(bytes_to_hex_fixed(&item.s, 32)));
    o.insert(
        "yParity".into(),
        Value::String(quantity_hex(u64::from(item.y_parity))),
    );
    if !item.authority.is_empty() {
        o.insert(
            "authority".into(),
            Value::String(bytes_to_hex(&item.authority)),
        );
    }
    Value::Object(o)
}

/// Converts a BDS `Receipt` into the exact JSON-RPC transaction receipt
/// object shape, including its `logs` array. Mirrors `evm.ReceiptToJsonRpc`
/// in `evm/json_rpc.go`.
#[must_use]
pub fn receipt_to_json(r: &Receipt) -> Value {
    let mut o = Map::new();
    o.insert(
        "transactionHash".into(),
        Value::String(bytes_to_hex(&r.transaction_hash)),
    );
    o.insert(
        "transactionIndex".into(),
        Value::String(quantity_hex(u64::from(r.transaction_index))),
    );
    o.insert(
        "blockHash".into(),
        Value::String(bytes_to_hex(&r.block_hash)),
    );
    o.insert(
        "blockNumber".into(),
        Value::String(quantity_hex(r.block_number)),
    );
    o.insert("from".into(), Value::String(bytes_to_hex(&r.from)));
    o.insert(
        "cumulativeGasUsed".into(),
        Value::String(quantity_hex(r.cumulative_gas_used)),
    );
    o.insert("gasUsed".into(), Value::String(quantity_hex(r.gas_used)));
    o.insert(
        "logsBloom".into(),
        Value::String(bytes_to_hex(&r.logs_bloom)),
    );
    o.insert(
        "logs".into(),
        Value::Array(r.logs.iter().map(log_to_json).collect()),
    );
    // contractAddress and to are always present (null when absent/empty).
    o.insert(
        "contractAddress".into(),
        nonempty(&r.contract_address).map_or(Value::Null, |b| Value::String(bytes_to_hex(b))),
    );
    o.insert(
        "to".into(),
        nonempty(&r.to).map_or(Value::Null, |b| Value::String(bytes_to_hex(b))),
    );

    // Unlike "to"/"contractAddress" above, status is simply omitted (not
    // null) when absent.
    if let Some(v) = r.status {
        o.insert("status".into(), Value::String(quantity_hex(u64::from(v))));
    }
    o.insert(
        "type".into(),
        Value::String(quantity_hex(u64::from(r.r#type))),
    );

    if !r.effective_gas_price.is_empty() {
        if let Some(v) = decimal_string_to_hex(&r.effective_gas_price) {
            o.insert("effectiveGasPrice".into(), Value::String(v));
        }
    }
    if let Some(v) = nonempty(&r.root) {
        o.insert("root".into(), Value::String(bytes_to_hex(v)));
    }
    if let Some(v) = r.block_timestamp {
        o.insert("blockTimestamp".into(), Value::String(quantity_hex(v)));
    }
    if let Some(v) = r.gas_used_for_l1 {
        o.insert("gasUsedForL1".into(), Value::String(quantity_hex(v)));
    }
    if let Some(v) = r.l1_block_number {
        o.insert("l1BlockNumber".into(), Value::String(quantity_hex(v)));
    }

    // Unlike the transaction's L2 fee breakdown, the RECEIPT's l1Fee,
    // l1GasUsed AND l1GasPrice all fall back to explicit `null` when absent
    // (json_rpc.go's `ReceiptToJsonRpc` has an `else { ... = nil }` branch
    // on each of these three; `TransactionToJsonRpc` only has it for l1Fee).
    insert_decimal_or_null(&mut o, "l1Fee", r.l1_fee.as_deref());
    insert_decimal_or_null(&mut o, "l1GasUsed", r.l1_gas_used.as_deref());
    insert_decimal_or_null(&mut o, "l1GasPrice", r.l1_gas_price.as_deref());
    insert_decimal_omit(&mut o, "gatewayFee", r.gateway_fee.as_deref());

    if let Some(v) = r.blob_gas_used {
        o.insert("blobGasUsed".into(), Value::String(quantity_hex(v)));
    }
    insert_decimal_omit(&mut o, "blobGasPrice", r.blob_gas_price.as_deref());
    if let Some(v) = r.l1_fee_scalar {
        o.insert("l1FeeScalar".into(), Value::from(v));
    }
    if let Some(v) = r.l1_base_fee_scalar {
        o.insert("l1BaseFeeScalar".into(), Value::String(quantity_hex(v)));
    }
    insert_decimal_omit(&mut o, "l1BlobBaseFee", r.l1_blob_base_fee.as_deref());
    if let Some(v) = r.l1_blob_base_fee_scalar {
        o.insert("l1BlobBaseFeeScalar".into(), Value::String(quantity_hex(v)));
    }
    if let Some(v) = r.da_footprint_gas_scalar {
        o.insert(
            "daFootprintGasScalar".into(),
            Value::String(quantity_hex(v)),
        );
    }
    insert_decimal_omit(&mut o, "depositNonce", r.deposit_nonce.as_deref());
    insert_decimal_omit(
        &mut o,
        "depositReceiptVersion",
        r.deposit_receipt_version.as_deref(),
    );
    if let Some(v) = r.timeboosted {
        o.insert("timeboosted".into(), Value::Bool(v));
    }

    Value::Object(o)
}

/// Converts a BDS `Log` into the exact JSON-RPC log object shape. Mirrors
/// `evm.LogToJsonRpc` in `evm/json_rpc.go`. `removed` is always `false` —
/// the BDS server only ever returns confirmed, non-reorged logs.
#[must_use]
pub fn log_to_json(log: &Log) -> Value {
    let mut o = Map::new();
    o.insert("address".into(), Value::String(bytes_to_hex(&log.address)));
    o.insert(
        "topics".into(),
        Value::Array(
            log.topics
                .iter()
                .map(|t| Value::String(bytes_to_hex(t)))
                .collect(),
        ),
    );
    o.insert("data".into(), Value::String(bytes_to_hex(&log.data)));
    o.insert(
        "blockNumber".into(),
        Value::String(quantity_hex(log.block_number)),
    );
    o.insert(
        "transactionHash".into(),
        Value::String(bytes_to_hex(&log.transaction_hash)),
    );
    o.insert(
        "transactionIndex".into(),
        Value::String(quantity_hex(u64::from(log.transaction_index))),
    );
    o.insert(
        "blockHash".into(),
        Value::String(bytes_to_hex(&log.block_hash)),
    );
    o.insert(
        "logIndex".into(),
        Value::String(quantity_hex(u64::from(log.log_index))),
    );
    o.insert("removed".into(), Value::Bool(false));
    if let Some(v) = log.block_timestamp {
        o.insert("blockTimestamp".into(), Value::String(quantity_hex(v)));
    }
    Value::Object(o)
}

/// Converts a BDS `Withdrawal` into the exact JSON-RPC withdrawal object
/// shape. Mirrors `evm.WithdrawalToJsonRpc` in `evm/json_rpc.go`.
#[must_use]
pub fn withdrawal_to_json(w: &Withdrawal) -> Value {
    let mut o = Map::new();
    o.insert("index".into(), Value::String(quantity_hex(w.index)));
    o.insert(
        "validatorIndex".into(),
        Value::String(quantity_hex(w.validator_index)),
    );
    o.insert("address".into(), Value::String(bytes_to_hex(&w.address)));
    o.insert("amount".into(), Value::String(quantity_hex(w.amount)));
    Value::Object(o)
}

// --- shared hex/decimal helpers -------------------------------------------

/// `b` when `opt` is `Some` and non-empty, else `None` — a proto-optional
/// `bytes` field is treated as absent once it round-trips as empty, mirroring
/// every `len(x) > 0` check in `evm/json_rpc.go`.
fn nonempty(opt: &Option<Bytes>) -> Option<&[u8]> {
    opt.as_deref().filter(|b| !b.is_empty())
}

/// Formats an integer as a JSON-RPC QUANTITY: `0x`-prefixed, lowercase, no
/// leading zeros; zero is `0x0` (which `{:x}` already produces for `0u64`).
/// Lowercase hex digits, indexed by nibble.
const HEX: &[u8; 16] = b"0123456789abcdef";

fn quantity_hex(n: u64) -> String {
    // At most 16 nibbles; write them into a stack buffer rather than paying
    // the `format!` machinery and its allocation for every integer field.
    if n == 0 {
        return "0x0".to_owned();
    }
    let mut digits = [0u8; 16];
    let mut i = 16;
    let mut v = n;
    while v != 0 {
        i -= 1;
        digits[i] = HEX[(v & 0xf) as usize];
        v >>= 4;
    }
    let mut out = String::with_capacity(2 + (16 - i));
    out.push_str("0x");
    out.push_str(std::str::from_utf8(&digits[i..]).expect("hex digits are ascii"));
    out
}

/// `0x`-prefixed hex of raw bytes (a JSON-RPC DATA field): always
/// even-length, one nibble pair per byte, never trimmed.
fn bytes_to_hex(b: &[u8]) -> String {
    // One allocation, two table lookups per byte. This previously called
    // `format!` once PER BYTE, which dominated the entire JSON mapping:
    // 172,483 allocations and 6.5 ms for a 318-transaction block.
    let mut out = Vec::with_capacity(2 + b.len() * 2);
    out.push(b'0');
    out.push(b'x');
    for &byte in b {
        out.push(HEX[(byte >> 4) as usize]);
        out.push(HEX[(byte & 0xf) as usize]);
    }
    String::from_utf8(out).expect("hex digits are ascii")
}

/// Left-pads `b` with zero bytes to `size` bytes before hex-encoding —
/// for fixed-width DATA fields (e.g. signature r/s = 32 bytes) that must
/// keep leading zeros. Longer-than-`size` input is passed through as-is.
fn bytes_to_hex_fixed(b: &[u8], size: usize) -> String {
    if b.len() >= size {
        return bytes_to_hex(b);
    }
    let mut buf = vec![0u8; size];
    buf[size - b.len()..].copy_from_slice(b);
    bytes_to_hex(&buf)
}

/// Encodes raw bytes as a JSON-RPC QUANTITY (big-endian integer, no leading
/// zeros, `0x0` for empty/zero) — used for the transaction `v` field, which
/// is a signature byte string interpreted as a number, not fixed-width DATA.
fn bytes_to_quantity_hex(b: &[u8]) -> String {
    let trimmed = {
        let mut i = 0;
        while i < b.len() && b[i] == 0 {
            i += 1;
        }
        &b[i..]
    };
    if trimmed.is_empty() {
        return "0x0".to_string();
    }
    let mut hex = String::with_capacity(trimmed.len() * 2);
    for (idx, byte) in trimmed.iter().enumerate() {
        if idx == 0 {
            hex.push_str(&format!("{byte:x}"));
        } else {
            hex.push_str(&format!("{byte:02x}"));
        }
    }
    format!("0x{hex}")
}

/// Canonicalizes a decimal (or already-hex) numeric string into a JSON-RPC
/// QUANTITY: `0x`-prefixed, no leading zeros, `"0x0"` for zero/empty.
/// Mirrors Go's `DecimalStringToHex`. Returns `None` on malformed input so
/// callers can silently omit the field, exactly like the Go reference does
/// when it discards the error.
fn decimal_string_to_hex(s: &str) -> Option<String> {
    let s = s.trim();
    if s.is_empty() {
        return Some("0x0".to_string());
    }
    if let Some(hex) = s.strip_prefix("0x").or_else(|| s.strip_prefix("0X")) {
        return normalize_hex_quantity(hex);
    }
    if !s.bytes().all(|b| b.is_ascii_digit()) {
        return None;
    }
    Some(decimal_digits_to_hex(s))
}

fn normalize_hex_quantity(hex: &str) -> Option<String> {
    if hex.is_empty() {
        return Some("0x0".to_string());
    }
    if !hex.bytes().all(|b| b.is_ascii_hexdigit()) {
        return None;
    }
    let trimmed = hex.trim_start_matches('0');
    if trimmed.is_empty() {
        Some("0x0".to_string())
    } else {
        Some(format!("0x{}", trimmed.to_ascii_lowercase()))
    }
}

/// Converts an arbitrary-precision base-10 digit string to a minimal-hex
/// QUANTITY via repeated long division by 16. `dec` must be all ASCII
/// digits (checked by the caller). No bignum crate is in this workspace's
/// dependency list, and these values (256-bit EVM words) are far too wide
/// for `u128`, hence the manual long division.
fn decimal_digits_to_hex(dec: &str) -> String {
    let mut digits: Vec<u8> = dec.bytes().map(|b| b - b'0').collect();
    while digits.len() > 1 && digits[0] == 0 {
        digits.remove(0);
    }
    if digits == [0] {
        return "0x0".to_string();
    }
    let mut hex_digits = Vec::new();
    while !(digits.len() == 1 && digits[0] == 0) {
        let mut remainder: u32 = 0;
        let mut next = Vec::with_capacity(digits.len());
        for &d in &digits {
            let cur = remainder * 10 + u32::from(d);
            next.push((cur / 16) as u8);
            remainder = cur % 16;
        }
        while next.len() > 1 && next[0] == 0 {
            next.remove(0);
        }
        hex_digits.push(remainder as u8);
        digits = next;
    }
    hex_digits.reverse();
    let hex_str: String = hex_digits
        .iter()
        .map(|d| char::from_digit(u32::from(*d), 16).unwrap_or('0'))
        .collect();
    format!("0x{hex_str}")
}

/// Inserts `key` as `null` when `v` is `None`, as the parsed decimal
/// QUANTITY when `v` parses, or omits the key entirely when `v` is
/// malformed. Used by the handful of fields (transaction/receipt `l1Fee`,
/// and the receipt's `l1GasUsed`/`l1GasPrice`) that fall back to an explicit
/// `null` rather than omission when the proto field itself is absent.
fn insert_decimal_or_null(o: &mut Map<String, Value>, key: &'static str, v: Option<&str>) {
    match v {
        None => {
            o.insert(key.to_string(), Value::Null);
        }
        Some(s) => {
            if let Some(hex) = decimal_string_to_hex(s) {
                o.insert(key.to_string(), Value::String(hex));
            }
        }
    }
}

/// Inserts `key` only when `v` is present and parses as a decimal QUANTITY;
/// omitted entirely otherwise. This is the common case across both messages.
fn insert_decimal_omit(o: &mut Map<String, Value>, key: &'static str, v: Option<&str>) {
    if let Some(hex) = v.and_then(decimal_string_to_hex) {
        o.insert(key.to_string(), Value::String(hex));
    }
}

// --- inverse: JSON-RPC response → proto ("ingest") ----------------------
//
// Mirrors `(*JsonRpcBlock).ToProto`, `ParseJsonRpcTransaction(s)`,
// `(*JsonRpcReceipt).ToProto`, `(*JsonRpcLog).ToProto`, and
// `(*JsonRpcWithdrawal).ToProto` in `evm/json_rpc.go`, plus the `HexToBytes`/
// `NumberishToUint64`/`NumberishString` helpers in `evm/util.go`. Field
// presence follows that source's actual behavior (documented per-field
// below) with one deliberate exception: `evm/json_rpc.go:328-330` sets
// `BlockHeader.BaseFeePerGas`/`Difficulty`/`TotalDifficulty` unconditionally
// to a pointer at `""` when the JSON key is absent, so a later
// re-serialization fabricates `"0x0"` out of nothing. This port treats all
// three like every other optional quantity string instead: absent/empty
// means `None`.

fn json_kind(v: &Value) -> &'static str {
    match v {
        Value::Null => "null",
        Value::Bool(_) => "a boolean",
        Value::Number(_) => "a number",
        Value::String(_) => "a string",
        Value::Array(_) => "an array",
        Value::Object(_) => "an object",
    }
}

fn as_object<'a>(
    v: &'a Value,
    expected: &'static str,
) -> Result<&'a Map<String, Value>, FromJsonError> {
    v.as_object().ok_or(FromJsonError::Shape {
        expected,
        found: json_kind(v),
    })
}

fn as_array<'a>(v: &'a Value, expected: &'static str) -> Result<&'a Vec<Value>, FromJsonError> {
    v.as_array().ok_or(FromJsonError::Shape {
        expected,
        found: json_kind(v),
    })
}

/// The value at `key`, treating JSON `null` the same as absent — mirrors
/// Go's map-lookup-plus-type-assertion pattern, where a `null` value never
/// type-asserts successfully either.
fn field<'a>(o: &'a Map<String, Value>, key: &'static str) -> Option<&'a Value> {
    o.get(key).filter(|v| !v.is_null())
}

fn field_str<'a>(o: &'a Map<String, Value>, key: &'static str) -> Option<&'a str> {
    field(o, key).and_then(Value::as_str)
}

fn required_str<'a>(
    o: &'a Map<String, Value>,
    key: &'static str,
) -> Result<&'a str, FromJsonError> {
    field_str(o, key).ok_or(FromJsonError::MissingField(key))
}

fn hex_to_bytes(field_name: &'static str, s: &str) -> Result<Bytes, FromJsonError> {
    let stripped = s
        .strip_prefix("0x")
        .or_else(|| s.strip_prefix("0X"))
        .unwrap_or(s);
    hex_digit_pairs_to_bytes(stripped)
        .map(Bytes::from)
        .map_err(|_| FromJsonError::InvalidHex {
            field: field_name,
            value: s.to_string(),
        })
}

/// A `bytes` field that is non-optional on the wire but, mirroring Go's
/// `HexToBytes(getString(key))` idiom, quietly becomes an empty value when
/// the JSON key is absent/empty/`"0x"` rather than erroring — Go only ever
/// errors here on a present-but-malformed hex string.
fn req_bytes(o: &Map<String, Value>, key: &'static str) -> Result<Bytes, FromJsonError> {
    hex_to_bytes(key, field_str(o, key).unwrap_or(""))
}

/// An optional `bytes` field: absent, `null`, or `""` all mean `None`;
/// present-but-malformed hex is an error.
fn opt_bytes(o: &Map<String, Value>, key: &'static str) -> Result<Option<Bytes>, FromJsonError> {
    match field_str(o, key) {
        None | Some("") => Ok(None),
        Some(s) => hex_to_bytes(key, s).map(Some),
    }
}

/// Like [`opt_bytes`], but `"0x"` (present, zero-length hex) is *also*
/// treated as absence — the `to`/`contractAddress` convention
/// (json_rpc.go:361,452,479).
fn opt_address(o: &Map<String, Value>, key: &'static str) -> Result<Option<Bytes>, FromJsonError> {
    match field_str(o, key) {
        None | Some("") | Some("0x") => Ok(None),
        Some(s) => hex_to_bytes(key, s).map(Some),
    }
}

/// Mirrors `NumberishToUint64`: a `0x`-prefixed hex quantity, or a decimal
/// string.
fn numberish_to_u64(field_name: &'static str, s: &str) -> Result<u64, FromJsonError> {
    if let Some(digits) = s.strip_prefix("0x").or_else(|| s.strip_prefix("0X")) {
        if digits.is_empty() {
            return Ok(0);
        }
        u64::from_str_radix(digits, 16).map_err(|_| FromJsonError::InvalidNumber {
            field: field_name,
            value: s.to_string(),
        })
    } else {
        s.parse::<u64>().map_err(|_| FromJsonError::InvalidNumber {
            field: field_name,
            value: s.to_string(),
        })
    }
}

fn numberish_to_u32(field_name: &'static str, s: &str) -> Result<u32, FromJsonError> {
    u32::try_from(numberish_to_u64(field_name, s)?).map_err(|_| FromJsonError::InvalidNumber {
        field: field_name,
        value: s.to_string(),
    })
}

/// A numeric field that's genuinely required — mirrors the handful of
/// `NumberishToUint64`/`32` calls in json_rpc.go that Go itself errors on
/// when the string is empty (`strconv.ParseUint("", ...)` fails), e.g.
/// block `number`/`timestamp`, log `blockNumber`/`logIndex`, receipt
/// `transactionIndex`.
fn required_u64_field(o: &Map<String, Value>, key: &'static str) -> Result<u64, FromJsonError> {
    numberish_to_u64(key, required_str(o, key)?)
}

fn required_u32_field(o: &Map<String, Value>, key: &'static str) -> Result<u32, FromJsonError> {
    numberish_to_u32(key, required_str(o, key)?)
}

/// The transaction/receipt `type` convention: defaults to `0` (legacy) when
/// absent or unparseable, rather than erroring (json_rpc.go:1402-1406).
fn u32_field_default_zero(o: &Map<String, Value>, key: &'static str) -> u32 {
    field_str(o, key)
        .and_then(|s| numberish_to_u32(key, s).ok())
        .unwrap_or(0)
}

fn optional_u64_field(
    o: &Map<String, Value>,
    key: &'static str,
) -> Result<Option<u64>, FromJsonError> {
    match field_str(o, key) {
        None | Some("") => Ok(None),
        Some(s) => Ok(Some(numberish_to_u64(key, s)?)),
    }
}

fn optional_u32_field(
    o: &Map<String, Value>,
    key: &'static str,
) -> Result<Option<u32>, FromJsonError> {
    match field_str(o, key) {
        None | Some("") => Ok(None),
        Some(s) => Ok(Some(numberish_to_u32(key, s)?)),
    }
}

fn optional_bool_field(o: &Map<String, Value>, key: &'static str) -> Option<bool> {
    field(o, key).and_then(Value::as_bool)
}

/// A raw string field passed through verbatim (no hex/decimal validation) —
/// e.g. the beacon-chain `proposerPublicKey`.
fn optional_raw_string_field(o: &Map<String, Value>, key: &'static str) -> Option<String> {
    match field_str(o, key) {
        None | Some("") => None,
        Some(s) => Some(s.to_string()),
    }
}

/// A 256-bit decimal-or-hex QUANTITY string, stored **verbatim** (never
/// renormalized) once validated — the convention for `value`, `gasPrice`,
/// `l1Fee`, and the rest of this module's big-number string fields, per the
/// "store quantity strings verbatim" rule. [`decimal_string_to_hex`] (the
/// to-JSON direction) is reused purely as a validator here, since it already
/// accepts exactly this field shape and rejects anything else.
fn quantity_string_field(
    o: &Map<String, Value>,
    key: &'static str,
) -> Result<Option<String>, FromJsonError> {
    match field_str(o, key) {
        None | Some("") => Ok(None),
        Some(s) => {
            if decimal_string_to_hex(s).is_none() {
                return Err(FromJsonError::InvalidNumber {
                    field: key,
                    value: s.to_string(),
                });
            }
            Ok(Some(s.to_string()))
        }
    }
}

/// `l1FeeScalar`-style field: unlike most fields here, a JSON-RPC node may
/// send this as either a quoted numeric string or a bare JSON number.
fn optional_f64_field(
    o: &Map<String, Value>,
    key: &'static str,
) -> Result<Option<f64>, FromJsonError> {
    match field(o, key) {
        None => Ok(None),
        Some(Value::Number(n)) => Ok(n.as_f64()),
        Some(Value::String(s)) if s.is_empty() => Ok(None),
        Some(Value::String(s)) => {
            s.parse::<f64>()
                .map(Some)
                .map_err(|_| FromJsonError::InvalidNumber {
                    field: key,
                    value: s.clone(),
                })
        }
        Some(_) => Err(FromJsonError::WrongType {
            field: key,
            reason: "expected a number or numeric string",
        }),
    }
}

/// A `uint64` field that, like [`optional_f64_field`], accepts either a
/// quoted numeric string or a bare JSON number — used for `blockTimestamp`
/// (a `NumberishString` in Go) and `l1BlobBaseFeeScalar`.
fn optional_numberish_u64_field(
    o: &Map<String, Value>,
    key: &'static str,
) -> Result<Option<u64>, FromJsonError> {
    match field(o, key) {
        None => Ok(None),
        Some(Value::Number(n)) => {
            n.as_u64()
                .map(Some)
                .ok_or_else(|| FromJsonError::InvalidNumber {
                    field: key,
                    value: n.to_string(),
                })
        }
        Some(Value::String(s)) if s.is_empty() => Ok(None),
        Some(Value::String(s)) => Ok(Some(numberish_to_u64(key, s)?)),
        Some(_) => Err(FromJsonError::WrongType {
            field: key,
            reason: "expected a number or numeric string",
        }),
    }
}

/// Converts a JSON-RPC withdrawal object into a BDS `Withdrawal`. Mirrors
/// `(*JsonRpcWithdrawal).ToProto` in `evm/json_rpc.go`.
///
/// # Errors
///
/// Returns [`FromJsonError`] when `v` isn't an object, or a required field is
/// missing/malformed.
pub fn withdrawal_from_json(v: &Value) -> Result<Withdrawal, FromJsonError> {
    let o = as_object(v, "a withdrawal object")?;
    Ok(Withdrawal {
        index: required_u64_field(o, "index")?,
        validator_index: required_u64_field(o, "validatorIndex")?,
        address: req_bytes(o, "address")?,
        amount: required_u64_field(o, "amount")?,
    })
}

/// Converts a JSON-RPC log object into a BDS `Log`. Mirrors
/// `(*JsonRpcLog).ToProto` in `evm/json_rpc.go`. `blockTimestamp` accepts a
/// JSON number or a hex/decimal string (Go's `NumberishString`).
///
/// # Errors
///
/// Returns [`FromJsonError`] when `v` isn't an object, or a required field is
/// missing/malformed.
pub fn log_from_json(v: &Value) -> Result<Log, FromJsonError> {
    let o = as_object(v, "a log object")?;
    let topics = match field(o, "topics") {
        None => Vec::new(),
        Some(v) => as_array(v, "a topics array")?
            .iter()
            .map(|t| {
                t.as_str()
                    .ok_or(FromJsonError::WrongType {
                        field: "topics",
                        reason: "expected an array of hex strings",
                    })
                    .and_then(|s| hex_to_bytes("topics", s))
            })
            .collect::<Result<Vec<_>, _>>()?,
    };
    Ok(Log {
        address: req_bytes(o, "address")?,
        topics,
        data: req_bytes(o, "data")?,
        block_number: required_u64_field(o, "blockNumber")?,
        block_hash: req_bytes(o, "blockHash")?,
        transaction_hash: req_bytes(o, "transactionHash")?,
        transaction_index: required_u32_field(o, "transactionIndex")?,
        log_index: required_u32_field(o, "logIndex")?,
        block_timestamp: optional_numberish_u64_field(o, "blockTimestamp")?,
    })
}

/// Converts a JSON-RPC transaction receipt object into a BDS `Receipt`,
/// including its `logs` array. Mirrors `(*JsonRpcReceipt).ToProto` in
/// `evm/json_rpc.go`. `blockTimestamp` accepts a JSON number or a
/// hex/decimal string.
///
/// # Errors
///
/// Returns [`FromJsonError`] when `v` isn't an object, or a required field is
/// missing/malformed.
pub fn receipt_from_json(v: &Value) -> Result<Receipt, FromJsonError> {
    let o = as_object(v, "a receipt object")?;
    let logs = match field(o, "logs") {
        None => Vec::new(),
        Some(v) => as_array(v, "a logs array")?
            .iter()
            .map(log_from_json)
            .collect::<Result<Vec<_>, _>>()?,
    };
    Ok(Receipt {
        transaction_hash: req_bytes(o, "transactionHash")?,
        block_number: required_u64_field(o, "blockNumber")?,
        block_hash: req_bytes(o, "blockHash")?,
        transaction_index: required_u32_field(o, "transactionIndex")?,
        r#type: u32_field_default_zero(o, "type"),
        from: req_bytes(o, "from")?,
        // "" and "0x" both mean absent (json_rpc.go:452).
        to: opt_address(o, "to")?,
        status: optional_u32_field(o, "status")?,
        gas_used: required_u64_field(o, "gasUsed")?,
        cumulative_gas_used: required_u64_field(o, "cumulativeGasUsed")?,
        // Non-optional proto string; passed through verbatim like Go's
        // `EffectiveGasPrice string` field (no validation, "" when absent).
        effective_gas_price: field_str(o, "effectiveGasPrice")
            .unwrap_or_default()
            .to_string(),
        logs_bloom: req_bytes(o, "logsBloom")?,
        logs,
        // "" and "0x" both mean absent (json_rpc.go:479).
        contract_address: opt_address(o, "contractAddress")?,
        root: opt_bytes(o, "root")?,
        block_timestamp: optional_numberish_u64_field(o, "blockTimestamp")?,
        blob_gas_used: optional_u64_field(o, "blobGasUsed")?,
        blob_gas_price: quantity_string_field(o, "blobGasPrice")?,
        timeboosted: optional_bool_field(o, "timeboosted"),
        l1_fee: quantity_string_field(o, "l1Fee")?,
        l1_gas_used: quantity_string_field(o, "l1GasUsed")?,
        l1_gas_price: quantity_string_field(o, "l1GasPrice")?,
        l1_fee_scalar: optional_f64_field(o, "l1FeeScalar")?,
        l1_base_fee_scalar: optional_u64_field(o, "l1BaseFeeScalar")?,
        gas_used_for_l1: optional_u64_field(o, "gasUsedForL1")?,
        l1_block_number: optional_u64_field(o, "l1BlockNumber")?,
        gateway_fee: quantity_string_field(o, "gatewayFee")?,
        deposit_nonce: quantity_string_field(o, "depositNonce")?,
        deposit_receipt_version: quantity_string_field(o, "depositReceiptVersion")?,
        l1_blob_base_fee: quantity_string_field(o, "l1BlobBaseFee")?,
        // Accepts a JSON number or a numeric string (more lenient than Go's
        // raw-string-only `getString`, which silently drops a bare number).
        l1_blob_base_fee_scalar: optional_numberish_u64_field(o, "l1BlobBaseFeeScalar")?,
        // Receipt-only — no equivalent Transaction field exists.
        da_footprint_gas_scalar: optional_u64_field(o, "daFootprintGasScalar")?,
    })
}

fn access_list_item_from_json(o: &Map<String, Value>) -> Result<AccessListItem, FromJsonError> {
    let address = req_bytes(o, "address")?;
    let storage_keys = match field(o, "storageKeys") {
        None => Vec::new(),
        Some(v) => match v.as_array() {
            Some(arr) => arr
                .iter()
                .filter_map(Value::as_str)
                .filter_map(|s| hex_to_bytes("storageKeys", s).ok())
                .collect(),
            None => Vec::new(),
        },
    };
    Ok(AccessListItem {
        address,
        storage_keys,
    })
}

fn authorization_list_item_from_json(
    o: &Map<String, Value>,
) -> Result<AuthorizationListItem, FromJsonError> {
    Ok(AuthorizationListItem {
        chain_id: required_u64_field(o, "chainId")?,
        address: req_bytes(o, "address")?,
        nonce: required_u64_field(o, "nonce")?,
        r: req_bytes(o, "r")?,
        s: req_bytes(o, "s")?,
        y_parity: required_u32_field(o, "yParity")?,
        // Optional; if present, malformed hex still errors (json_rpc.go's
        // authority branch only skips a *missing* authority, not a bad one).
        authority: opt_bytes(o, "authority")?.unwrap_or_default(),
    })
}

/// Converts a JSON-RPC transaction object into a BDS `Transaction`. Mirrors
/// `ParseJsonRpcTransaction` in `evm/json_rpc.go` (minus its `header`
/// backfill parameter — see [`get_block_response_from_json`], which applies
/// that backfill when parsing a block's transaction array).
///
/// # Errors
///
/// Returns [`FromJsonError`] when `v` isn't an object, or a required field is
/// missing/malformed.
pub fn transaction_from_json(v: &Value) -> Result<Transaction, FromJsonError> {
    transaction_from_json_object(as_object(v, "a transaction object")?, None)
}

fn transaction_from_json_object(
    o: &Map<String, Value>,
    header: Option<&BlockHeader>,
) -> Result<Transaction, FromJsonError> {
    // Required proto string; defaults to "0" when absent (json_rpc.go:1368-1371).
    let value = match field_str(o, "value") {
        None | Some("") => "0".to_string(),
        Some(s) => {
            if decimal_string_to_hex(s).is_none() {
                return Err(FromJsonError::InvalidNumber {
                    field: "value",
                    value: s.to_string(),
                });
            }
            s.to_string()
        }
    };

    // accessList: parsed item-by-item; a malformed item (bad address hex, or
    // not an object at all) is skipped silently rather than failing the
    // whole transaction.
    let access_list = match field(o, "accessList").and_then(Value::as_array) {
        None => Vec::new(),
        Some(arr) => arr
            .iter()
            .filter_map(Value::as_object)
            .filter_map(|item| access_list_item_from_json(item).ok())
            .collect(),
    };

    // authorizationList (EIP-7702): each item's chainId/address/nonce/r/s/
    // yParity are required — a malformed one errors the whole transaction,
    // matching json_rpc.go's unconditional error returns here. A
    // non-object entry is skipped (mirrors the Go type-assertion loop).
    let authorization_list = match field(o, "authorizationList").and_then(Value::as_array) {
        None => Vec::new(),
        Some(arr) => arr
            .iter()
            .filter_map(Value::as_object)
            .map(authorization_list_item_from_json)
            .collect::<Result<Vec<_>, _>>()?,
    };

    // blobVersionedHashes: non-string entries are skipped, but a malformed
    // hex string errors (json_rpc.go:1576-1588).
    let blob_versioned_hashes = match field(o, "blobVersionedHashes").and_then(Value::as_array) {
        None => Vec::new(),
        Some(arr) => arr
            .iter()
            .filter_map(Value::as_str)
            .map(|s| hex_to_bytes("blobVersionedHashes", s))
            .collect::<Result<Vec<_>, _>>()?,
    };

    // Block context: an explicit field on the transaction object always
    // wins; only missing values fall back to the enclosing block's header
    // (mirrors ParseJsonRpcTransaction's header-backed defaults, json_rpc.go:1541-1551).
    let block_number = match optional_u64_field(o, "blockNumber")? {
        Some(v) => Some(v),
        None => header.map(|h| h.number),
    };
    let block_hash = match opt_bytes(o, "blockHash")? {
        Some(v) => Some(v),
        None => header.map(|h| h.hash.clone()),
    };
    let block_timestamp = match optional_u64_field(o, "blockTimestamp")? {
        Some(v) => Some(v),
        None => header.map(|h| h.timestamp),
    };

    Ok(Transaction {
        hash: req_bytes(o, "hash")?,
        nonce: required_u64_field(o, "nonce")?,
        from: req_bytes(o, "from")?,
        // "" and "0x" both mean absent — contract creation (json_rpc.go:361).
        to: opt_address(o, "to")?,
        value,
        input: req_bytes(o, "input")?,
        r#type: u32_field_default_zero(o, "type"),
        gas_limit: required_u64_field(o, "gas")?,
        gas_price: quantity_string_field(o, "gasPrice")?,
        max_fee_per_gas: quantity_string_field(o, "maxFeePerGas")?,
        max_priority_fee_per_gas: quantity_string_field(o, "maxPriorityFeePerGas")?,
        gas_used: optional_u64_field(o, "gasUsed")?,
        effective_gas_price: quantity_string_field(o, "effectiveGasPrice")?,
        r: req_bytes(o, "r")?,
        s: req_bytes(o, "s")?,
        v: opt_bytes(o, "v")?,
        y_parity: optional_u32_field(o, "yParity")?,
        chain_id: optional_u64_field(o, "chainId")?,
        block_number,
        block_hash,
        transaction_index: optional_u32_field(o, "transactionIndex")?,
        block_timestamp,
        access_list,
        max_fee_per_blob_gas: quantity_string_field(o, "maxFeePerBlobGas")?,
        blob_versioned_hashes,
        blob_gas_used: optional_u64_field(o, "blobGasUsed")?,
        blob_gas_price: quantity_string_field(o, "blobGasPrice")?,
        authorization_list,
        l1_fee: quantity_string_field(o, "l1Fee")?,
        l1_gas_price: quantity_string_field(o, "l1GasPrice")?,
        l1_gas_used: quantity_string_field(o, "l1GasUsed")?,
        l1_fee_scalar: optional_f64_field(o, "l1FeeScalar")?,
        l1_blob_base_fee: quantity_string_field(o, "l1BlobBaseFee")?,
        // Accepts a JSON number or a numeric string, same leniency as the
        // receipt-side field of the same name.
        l1_blob_base_fee_scalar: optional_numberish_u64_field(o, "l1BlobBaseFeeScalar")?,
        gateway_fee: quantity_string_field(o, "gatewayFee")?,
        fee_currency: opt_bytes(o, "feeCurrency")?,
        gateway_fee_recipient: opt_bytes(o, "gatewayFeeRecipient")?,
        beneficiary: opt_bytes(o, "beneficiary")?,
        deposit_value: quantity_string_field(o, "depositValue")?,
        l1_base_fee: quantity_string_field(o, "l1BaseFee")?,
        max_submission_fee: quantity_string_field(o, "maxSubmissionFee")?,
        refund_to: opt_bytes(o, "refundTo")?,
        request_id: opt_bytes(o, "requestId")?,
        retry_data: opt_bytes(o, "retryData")?,
        retry_to: opt_bytes(o, "retryTo")?,
        retry_value: quantity_string_field(o, "retryValue")?,
        max_refund: quantity_string_field(o, "maxRefund")?,
        submission_fee_refund: quantity_string_field(o, "submissionFeeRefund")?,
        ticket_id: opt_bytes(o, "ticketId")?,
        is_system_tx: optional_bool_field(o, "isSystemTx"),
        deposit_receipt_version: quantity_string_field(o, "depositReceiptVersion")?,
        source_hash: opt_bytes(o, "sourceHash")?,
        mint: quantity_string_field(o, "mint")?,
    })
}

/// Converts a JSON-RPC block object's header fields into a BDS
/// `BlockHeader` (the `transactions`/`withdrawals` arrays are handled by
/// [`get_block_response_from_json`]). Mirrors the header portion of
/// `(*JsonRpcBlock).ToProto` in `evm/json_rpc.go`.
///
/// # Errors
///
/// Returns [`FromJsonError`] when `v` isn't an object, or a required field is
/// missing/malformed.
pub fn block_header_from_json(v: &Value) -> Result<BlockHeader, FromJsonError> {
    let o = as_object(v, "a block object")?;

    let uncles = match field(o, "uncles") {
        None => Vec::new(),
        Some(v) => as_array(v, "an uncles array")?
            .iter()
            .map(|u| {
                u.as_str()
                    .ok_or(FromJsonError::WrongType {
                        field: "uncles",
                        reason: "expected an array of hex strings",
                    })
                    .and_then(|s| hex_to_bytes("uncles", s))
            })
            .collect::<Result<Vec<_>, _>>()?,
    };

    Ok(BlockHeader {
        number: required_u64_field(o, "number")?,
        timestamp: required_u64_field(o, "timestamp")?,
        gas_limit: required_u64_field(o, "gasLimit")?,
        gas_used: required_u64_field(o, "gasUsed")?,
        // Defaults to 0 when absent, unlike the other required numeric
        // fields above (json_rpc.go:113-119).
        size: optional_u64_field(o, "size")?.unwrap_or(0),
        hash: req_bytes(o, "hash")?,
        parent_hash: req_bytes(o, "parentHash")?,
        state_root: req_bytes(o, "stateRoot")?,
        transactions_root: req_bytes(o, "transactionsRoot")?,
        receipts_root: req_bytes(o, "receiptsRoot")?,
        sha3_uncles: req_bytes(o, "sha3Uncles")?,
        miner: req_bytes(o, "miner")?,
        logs_bloom: req_bytes(o, "logsBloom")?,
        extra_data: req_bytes(o, "extraData")?,
        nonce: optional_u64_field(o, "nonce")?,
        blob_gas_used: optional_u64_field(o, "blobGasUsed")?,
        excess_blob_gas: optional_u64_field(o, "excessBlobGas")?,
        l1_block_number: optional_u64_field(o, "l1BlockNumber")?,
        epoch: optional_u64_field(o, "epoch")?,
        slot: optional_u64_field(o, "slot")?,
        proposer_index: optional_u64_field(o, "proposerIndex")?,
        send_count: optional_u64_field(o, "sendCount")?,
        transaction_count: optional_u32_field(o, "transactionCount")?,
        mix_hash: opt_bytes(o, "mixHash")?,
        parent_beacon_block_root: opt_bytes(o, "parentBeaconBlockRoot")?,
        withdrawals_root: opt_bytes(o, "withdrawalsRoot")?,
        send_root: opt_bytes(o, "sendRoot")?,
        // DEVIATION from json_rpc.go:328-330 (a known Go bug, not to be
        // copied): Go sets these three unconditionally to `&b.BaseFeePerGas`
        // etc., so an absent key becomes a pointer to "" and a later
        // re-serialization fabricates "0x0" out of nothing. Here they're
        // ordinary optional quantity strings: absent/empty means `None`.
        base_fee_per_gas: quantity_string_field(o, "baseFeePerGas")?,
        difficulty: quantity_string_field(o, "difficulty")?,
        total_difficulty: quantity_string_field(o, "totalDifficulty")?,
        proposer_public_key: optional_raw_string_field(o, "proposerPublicKey"),
        // Not sourced from JSON-RPC: this is a separate RLP-encoded-withdrawals
        // byte field with no Go/json_rpc.go equivalent (Go's BlockHeader has
        // no such field at all — the full withdrawal objects live on `Block`,
        // handled by `get_block_response_from_json` instead).
        withdrawals: None,
        canonical_rlp: opt_bytes(o, "canonicalRlp")?,
        uncles,
        requests_hash: opt_bytes(o, "requestsHash")?,
    })
}

/// Converts a JSON-RPC `eth_getBlockByNumber`/`eth_getBlockByHash` `result`
/// into a BDS `GetBlockResponse`. A JSON `null` result (the "block not
/// found" case) yields the default response with `block` unset. Mirrors
/// `(*JsonRpcBlock).ToProto` + `ParseJsonRpcTransactions` in
/// `evm/json_rpc.go`: a `transactions` array of strings becomes hashes
/// (`transactions`, with `full_transactions` empty); an array of objects
/// becomes `full_transactions` (with `transactions` empty), each backfilled
/// with this block's number/hash/timestamp when the transaction object
/// itself omits them.
///
/// # Errors
///
/// Returns [`FromJsonError`] when `v` isn't an object (or `null`), or a
/// required field is missing/malformed.
pub fn get_block_response_from_json(v: &Value) -> Result<GetBlockResponse, FromJsonError> {
    if v.is_null() {
        return Ok(GetBlockResponse::default());
    }
    let o = as_object(v, "a block object")?;
    let header = block_header_from_json(v)?;

    let mut transactions = Vec::new();
    let mut full_transactions = Vec::new();
    if let Some(txs) = field(o, "transactions") {
        let arr = as_array(txs, "a transactions array")?;
        // JSON-RPC never mixes hash-only and full-object transactions within
        // one block response, so the first element's shape decides for all.
        if matches!(arr.first(), Some(Value::Object(_))) {
            for tx in arr {
                let tx_obj = as_object(tx, "a transaction object")?;
                full_transactions.push(transaction_from_json_object(tx_obj, Some(&header))?);
            }
        } else {
            for tx in arr {
                let s = tx.as_str().ok_or(FromJsonError::WrongType {
                    field: "transactions",
                    reason: "expected an array of hex hashes or transaction objects",
                })?;
                transactions.push(hex_to_bytes("transactions", s)?);
            }
        }
    }

    let withdrawals = match field(o, "withdrawals") {
        None => Vec::new(),
        Some(v) => as_array(v, "a withdrawals array")?
            .iter()
            .map(withdrawal_from_json)
            .collect::<Result<Vec<_>, _>>()?,
    };

    Ok(GetBlockResponse {
        block: Some(header),
        transactions,
        chain_id: None,
        chain_genesis_hash: None,
        full_transactions,
        withdrawals,
    })
}

/// Converts an `eth_getTransactionByHash` `result` into a
/// `GetTransactionByHashResponse`. A JSON `null` result yields the default
/// response with `transaction` unset.
///
/// # Errors
///
/// Returns [`FromJsonError`] when `v` isn't an object (or `null`), or a
/// required field is missing/malformed.
pub fn get_transaction_by_hash_response_from_json(
    v: &Value,
) -> Result<GetTransactionByHashResponse, FromJsonError> {
    if v.is_null() {
        return Ok(GetTransactionByHashResponse::default());
    }
    Ok(GetTransactionByHashResponse {
        transaction: Some(transaction_from_json(v)?),
    })
}

/// Converts an `eth_getTransactionReceipt` `result` into a
/// `GetTransactionReceiptResponse`. A JSON `null` result yields the default
/// response with `receipt` unset.
///
/// # Errors
///
/// Returns [`FromJsonError`] when `v` isn't an object (or `null`), or a
/// required field is missing/malformed.
pub fn get_transaction_receipt_response_from_json(
    v: &Value,
) -> Result<GetTransactionReceiptResponse, FromJsonError> {
    if v.is_null() {
        return Ok(GetTransactionReceiptResponse::default());
    }
    Ok(GetTransactionReceiptResponse {
        receipt: Some(receipt_from_json(v)?),
    })
}

/// Converts an `eth_getLogs` `result` (an array of log objects) into a
/// `GetLogsResponse`.
///
/// # Errors
///
/// Returns [`FromJsonError`] when `v` isn't an array, or any element is
/// missing/malformed.
pub fn get_logs_response_from_json(v: &Value) -> Result<GetLogsResponse, FromJsonError> {
    let arr = as_array(v, "a logs array")?;
    Ok(GetLogsResponse {
        logs: arr
            .iter()
            .map(log_from_json)
            .collect::<Result<Vec<_>, _>>()?,
    })
}

/// Converts an `eth_getBlockReceipts` `result` (an array of receipt objects)
/// into a `GetBlockReceiptsResponse`.
///
/// # Errors
///
/// Returns [`FromJsonError`] when `v` isn't an array, or any element is
/// missing/malformed.
pub fn get_block_receipts_response_from_json(
    v: &Value,
) -> Result<GetBlockReceiptsResponse, FromJsonError> {
    let arr = as_array(v, "a receipts array")?;
    Ok(GetBlockReceiptsResponse {
        receipts: arr
            .iter()
            .map(receipt_from_json)
            .collect::<Result<Vec<_>, _>>()?,
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    // --- execute seam ------------------------------------------------

    /// The point of `execute` being generic: a client wrapped in tonic
    /// middleware must be accepted, so callers can inject per-call metadata
    /// (e.g. a W3C `traceparent`) without this crate knowing about any
    /// particular tracing stack. The property is purely a type-level one, so
    /// the check is that this function compiles; it is never called (channel
    /// construction would need a tokio reactor even for a lazy connect).
    #[allow(dead_code)]
    fn execute_accepts_intercepted_clients(
        call: RpcQueryCall,
        client: &mut RpcQueryServiceClient<
            tonic::service::interceptor::InterceptedService<
                tonic::transport::Channel,
                fn(tonic::Request<()>) -> Result<tonic::Request<()>, tonic::Status>,
            >,
        >,
    ) {
        drop(call.execute(client, None));
    }

    // --- request mapping ---------------------------------------------

    #[test]
    fn chain_id_ignores_params() {
        let call = map_request("eth_chainId", &json!([])).unwrap();
        assert!(matches!(call, RpcQueryCall::ChainId(_)));
        let call = map_request("eth_chainId", &json!(null)).unwrap();
        assert!(matches!(call, RpcQueryCall::ChainId(_)));
    }

    #[test]
    fn unsupported_method_errors() {
        let err = map_request("eth_call", &json!([])).unwrap_err();
        assert!(matches!(err, MapError::UnsupportedMethod));
    }

    #[test]
    fn get_block_by_number_tag_passes_through_verbatim() {
        let call = map_request("eth_getBlockByNumber", &json!(["latest", true])).unwrap();
        match call {
            RpcQueryCall::GetBlockByNumber(req) => {
                assert_eq!(req.block_number, "latest");
                assert!(req.include_transactions);
                assert!(req.chain_id.is_none());
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_block_by_number_hex_passes_through_verbatim() {
        let call = map_request("eth_getBlockByNumber", &json!(["0x1b4"])).unwrap();
        match call {
            RpcQueryCall::GetBlockByNumber(req) => {
                assert_eq!(req.block_number, "0x1b4");
                // includeTransactions defaults to false when omitted.
                assert!(!req.include_transactions);
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_block_by_number_missing_param_is_unmappable() {
        let err = map_request("eth_getBlockByNumber", &json!([])).unwrap_err();
        assert!(matches!(err, MapError::Unmappable(_)));
        let err = map_request("eth_getBlockByNumber", &json!([true])).unwrap_err();
        assert!(matches!(err, MapError::Unmappable(_)));
    }

    #[test]
    fn get_block_by_hash_valid() {
        let hash = format!("0x{}", "ab".repeat(32));
        let call = map_request("eth_getBlockByHash", &json!([hash, false])).unwrap();
        match call {
            RpcQueryCall::GetBlockByHash(req) => {
                assert_eq!(req.block_hash.len(), 32);
                assert!(!req.include_transactions);
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_block_by_hash_wrong_length_is_unmappable() {
        let hash = format!("0x{}", "ab".repeat(31));
        let err = map_request("eth_getBlockByHash", &json!([hash, false])).unwrap_err();
        assert!(matches!(err, MapError::Unmappable(_)));
    }

    #[test]
    fn get_block_by_hash_bad_hex_is_unmappable() {
        let err = map_request("eth_getBlockByHash", &json!(["0xzz", false])).unwrap_err();
        assert!(matches!(err, MapError::Unmappable(_)));
    }

    #[test]
    fn get_transaction_by_hash_and_receipt_extract_hash() {
        let hash = format!("0x{}", "cd".repeat(32));
        let call = map_request("eth_getTransactionByHash", &json!([hash.clone()])).unwrap();
        match call {
            RpcQueryCall::GetTransactionByHash(req) => assert_eq!(req.transaction_hash.len(), 32),
            _ => panic!("wrong variant"),
        }
        let call = map_request("eth_getTransactionReceipt", &json!([hash])).unwrap();
        match call {
            RpcQueryCall::GetTransactionReceipt(req) => assert_eq!(req.transaction_hash.len(), 32),
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_block_receipts_string_param() {
        let call = map_request("eth_getBlockReceipts", &json!(["latest"])).unwrap();
        match call {
            RpcQueryCall::GetBlockReceipts(req) => {
                assert_eq!(req.block_number.as_deref(), Some("latest"));
                assert!(req.block_hash.is_none());
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_block_receipts_block_hash_object_param() {
        let hash = format!("0x{}", "11".repeat(32));
        let call = map_request("eth_getBlockReceipts", &json!([{ "blockHash": hash }])).unwrap();
        match call {
            RpcQueryCall::GetBlockReceipts(req) => {
                assert!(req.block_number.is_none());
                assert_eq!(req.block_hash.unwrap().len(), 32);
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_block_receipts_block_number_object_param() {
        let call =
            map_request("eth_getBlockReceipts", &json!([{ "blockNumber": "0x10" }])).unwrap();
        match call {
            RpcQueryCall::GetBlockReceipts(req) => {
                assert_eq!(req.block_number.as_deref(), Some("0x10"));
                assert!(req.block_hash.is_none());
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_block_receipts_invalid_object_is_unmappable() {
        let err = map_request("eth_getBlockReceipts", &json!([{ "foo": "bar" }])).unwrap_err();
        assert!(matches!(err, MapError::Unmappable(_)));
    }

    fn logs_filter(extra: Value) -> Value {
        let mut base = json!({ "fromBlock": "0x1", "toBlock": "0x10" });
        if let (Value::Object(base_obj), Value::Object(extra_obj)) = (&mut base, extra) {
            base_obj.extend(extra_obj);
        }
        json!([base])
    }

    #[test]
    fn get_logs_from_to_hex_parsed() {
        let call = map_request("eth_getLogs", &logs_filter(json!({}))).unwrap();
        match call {
            RpcQueryCall::GetLogs(req) => {
                assert_eq!(req.from_block, Some(1));
                assert_eq!(req.to_block, Some(16));
                assert!(req.block_hash.is_none());
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_logs_tag_without_block_hash_is_unmappable() {
        let params = json!([{ "fromBlock": "latest", "toBlock": "0x10" }]);
        let err = map_request("eth_getLogs", &params).unwrap_err();
        assert!(matches!(err, MapError::Unmappable(_)));
    }

    #[test]
    fn get_logs_missing_bound_without_block_hash_is_unmappable() {
        let params = json!([{ "fromBlock": "0x1" }]);
        let err = map_request("eth_getLogs", &params).unwrap_err();
        assert!(matches!(err, MapError::Unmappable(_)));
    }

    #[test]
    fn get_logs_block_hash_skips_range_requirement() {
        let hash = format!("0x{}", "22".repeat(32));
        let params = json!([{ "blockHash": hash }]);
        let call = map_request("eth_getLogs", &params).unwrap();
        match call {
            RpcQueryCall::GetLogs(req) => {
                assert!(req.from_block.is_none());
                assert!(req.to_block.is_none());
                assert_eq!(req.block_hash.unwrap().len(), 32);
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_logs_address_string_and_array() {
        let addr = format!("0x{}", "33".repeat(20));
        let call = map_request(
            "eth_getLogs",
            &logs_filter(json!({ "address": addr.clone() })),
        )
        .unwrap();
        match call {
            RpcQueryCall::GetLogs(req) => assert_eq!(req.addresses.len(), 1),
            _ => panic!("wrong variant"),
        }

        let addr2 = format!("0x{}", "44".repeat(20));
        let call = map_request(
            "eth_getLogs",
            &logs_filter(json!({ "address": [addr, addr2] })),
        )
        .unwrap();
        match call {
            RpcQueryCall::GetLogs(req) => assert_eq!(req.addresses.len(), 2),
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_logs_address_wrong_length_is_unmappable() {
        let addr = format!("0x{}", "33".repeat(19));
        let err = map_request("eth_getLogs", &logs_filter(json!({ "address": addr }))).unwrap_err();
        assert!(matches!(err, MapError::Unmappable(_)));
    }

    #[test]
    fn get_logs_topics_null_preserves_position() {
        let topic0 = format!("0x{}", "55".repeat(32));
        let topic2 = format!("0x{}", "66".repeat(32));
        let params = logs_filter(json!({ "topics": [topic0, Value::Null, topic2] }));
        let call = map_request("eth_getLogs", &params).unwrap();
        match call {
            RpcQueryCall::GetLogs(req) => {
                assert_eq!(req.topics.len(), 3);
                assert_eq!(req.topics[0].values.len(), 1);
                assert!(
                    req.topics[1].values.is_empty(),
                    "null position must stay a wildcard"
                );
                assert_eq!(req.topics[2].values.len(), 1);
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_logs_topics_array_of_strings() {
        let a = format!("0x{}", "77".repeat(32));
        let b = format!("0x{}", "88".repeat(32));
        let params = logs_filter(json!({ "topics": [[a, b]] }));
        let call = map_request("eth_getLogs", &params).unwrap();
        match call {
            RpcQueryCall::GetLogs(req) => assert_eq!(req.topics[0].values.len(), 2),
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn get_logs_bad_topic_hex_is_unmappable() {
        let params = logs_filter(json!({ "topics": ["not-hex"] }));
        let err = map_request("eth_getLogs", &params).unwrap_err();
        assert!(matches!(err, MapError::Unmappable(_)));
    }

    // --- hex/decimal helpers -------------------------------------------

    // --- hex helpers: equivalence with the implementations they replaced ---

    /// The exact previous implementations, frozen. `quantity_hex` and
    /// `bytes_to_hex` were rewritten for speed (they allocated once per byte
    /// via `format!`, which dominated the whole mapping). Speed is worthless
    /// if the bytes changed, so pin the new ones against the old ones over
    /// every value that matters.
    mod reference {
        pub fn quantity_hex(n: u64) -> String {
            format!("0x{n:x}")
        }
        pub fn bytes_to_hex(b: &[u8]) -> String {
            let mut s = String::with_capacity(2 + b.len() * 2);
            s.push_str("0x");
            for byte in b {
                s.push_str(&format!("{byte:02x}"));
            }
            s
        }
    }

    #[test]
    fn quantity_hex_matches_the_previous_implementation() {
        let mut cases: Vec<u64> = vec![0, 1, 9, 10, 15, 16, 17, 255, 256, 4095, 4096, u64::MAX];
        // Every power of two and its neighbours: the carry and width edges.
        for bit in 0..64 {
            let p = 1u64 << bit;
            cases.extend([p.wrapping_sub(1), p, p.wrapping_add(1)]);
        }
        // A deterministic spread across the whole range (xorshift, no deps).
        let mut x = 0x2545_f491_4f6c_dd1d_u64;
        for _ in 0..20_000 {
            x ^= x << 13;
            x ^= x >> 7;
            x ^= x << 17;
            cases.push(x);
        }
        for n in cases {
            assert_eq!(
                quantity_hex(n),
                reference::quantity_hex(n),
                "quantity_hex({n})"
            );
        }
    }

    #[test]
    fn bytes_to_hex_matches_the_previous_implementation() {
        // Empty, and every single byte value on its own.
        assert_eq!(bytes_to_hex(&[]), reference::bytes_to_hex(&[]));
        for b in 0..=255u8 {
            assert_eq!(
                bytes_to_hex(&[b]),
                reference::bytes_to_hex(&[b]),
                "byte {b:#04x}"
            );
        }
        // Real field widths: address, hash, bloom, and a large calldata blob.
        let mut x = 0x9e37_79b9_7f4a_7c15_u64;
        for len in [2usize, 8, 20, 31, 32, 33, 64, 256, 1024, 100_000] {
            let mut buf = Vec::with_capacity(len);
            for _ in 0..len {
                x ^= x << 13;
                x ^= x >> 7;
                x ^= x << 17;
                buf.push((x & 0xff) as u8);
            }
            assert_eq!(
                bytes_to_hex(&buf),
                reference::bytes_to_hex(&buf),
                "length {len}"
            );
            // All-zero and all-ones of the same width, where nibble handling slips.
            assert_eq!(
                bytes_to_hex(&vec![0x00; len]),
                reference::bytes_to_hex(&vec![0x00; len])
            );
            assert_eq!(
                bytes_to_hex(&vec![0xff; len]),
                reference::bytes_to_hex(&vec![0xff; len])
            );
        }
    }

    #[test]
    fn quantity_hex_is_minimal() {
        assert_eq!(quantity_hex(0), "0x0");
        assert_eq!(quantity_hex(255), "0xff");
        assert_eq!(quantity_hex(16), "0x10");
    }

    #[test]
    fn bytes_to_hex_is_even_length() {
        assert_eq!(bytes_to_hex(&[]), "0x");
        assert_eq!(bytes_to_hex(&[0x01]), "0x01");
        assert_eq!(bytes_to_hex(&[0xab, 0xcd]), "0xabcd");
    }

    #[test]
    fn bytes_to_hex_fixed_pads() {
        assert_eq!(bytes_to_hex_fixed(&[0x01], 4), "0x00000001");
        assert_eq!(bytes_to_hex_fixed(&[0xff; 4], 4), "0xffffffff");
        assert_eq!(bytes_to_hex_fixed(&[0xff; 5], 4), "0xffffffffff");
    }

    #[test]
    fn bytes_to_quantity_hex_strips_leading_zeros() {
        assert_eq!(bytes_to_quantity_hex(&[]), "0x0");
        assert_eq!(bytes_to_quantity_hex(&[0x00, 0x00]), "0x0");
        assert_eq!(bytes_to_quantity_hex(&[0x00, 0x1b]), "0x1b");
        assert_eq!(bytes_to_quantity_hex(&[0x01, 0x00]), "0x100");
    }

    #[test]
    fn decimal_string_to_hex_covers_decimal_hex_and_empty() {
        assert_eq!(decimal_string_to_hex(""), Some("0x0".to_string()));
        assert_eq!(decimal_string_to_hex("0"), Some("0x0".to_string()));
        assert_eq!(decimal_string_to_hex("16"), Some("0x10".to_string()));
        assert_eq!(decimal_string_to_hex("255"), Some("0xff".to_string()));
        assert_eq!(
            decimal_string_to_hex("1000000000000000000"),
            Some("0xde0b6b3a7640000".to_string())
        );
        assert_eq!(decimal_string_to_hex("0x1B4"), Some("0x1b4".to_string()));
        assert_eq!(decimal_string_to_hex("not-a-number"), None);
    }

    // --- response mapping: log / withdrawal -----------------------------

    fn sample_log() -> Log {
        Log {
            address: Bytes::from_static(&[0x11; 20]),
            topics: vec![Bytes::from_static(&[0x22; 32])],
            data: Bytes::from_static(&[0xde, 0xad]),
            block_number: 100,
            block_hash: Bytes::from_static(&[0x33; 32]),
            transaction_hash: Bytes::from_static(&[0x44; 32]),
            transaction_index: 2,
            log_index: 5,
            block_timestamp: None,
        }
    }

    #[test]
    fn log_to_json_omits_block_timestamp_when_absent() {
        let v = log_to_json(&sample_log());
        assert_eq!(v["logIndex"], "0x5");
        assert_eq!(v["transactionIndex"], "0x2");
        assert_eq!(v["removed"], false);
        assert!(v.get("blockTimestamp").is_none());
    }

    #[test]
    fn log_to_json_includes_block_timestamp_when_present() {
        let mut log = sample_log();
        log.block_timestamp = Some(1_700_000_000);
        let v = log_to_json(&log);
        assert_eq!(v["blockTimestamp"], "0x6553f100");
    }

    #[test]
    fn withdrawal_to_json_shape() {
        let w = Withdrawal {
            index: 1,
            validator_index: 2,
            address: Bytes::from_static(&[0xaa; 20]),
            amount: 3,
        };
        let v = withdrawal_to_json(&w);
        assert_eq!(v["index"], "0x1");
        assert_eq!(v["validatorIndex"], "0x2");
        assert_eq!(v["amount"], "0x3");
        assert_eq!(v["address"], format!("0x{}", "aa".repeat(20)));
    }

    // --- response mapping: transaction -----------------------------------

    fn base_transaction() -> Transaction {
        Transaction {
            hash: Bytes::from_static(&[0x01; 32]),
            nonce: 1,
            from: Bytes::from_static(&[0x02; 20]),
            to: None,
            value: String::new(),
            input: Bytes::new(),
            r#type: 0,
            gas_limit: 21000,
            gas_price: None,
            max_fee_per_gas: None,
            max_priority_fee_per_gas: None,
            gas_used: None,
            effective_gas_price: None,
            r: Bytes::new(),
            s: Bytes::new(),
            v: None,
            y_parity: None,
            chain_id: None,
            block_number: None,
            block_hash: None,
            transaction_index: None,
            block_timestamp: None,
            access_list: Vec::new(),
            max_fee_per_blob_gas: None,
            blob_versioned_hashes: Vec::new(),
            blob_gas_used: None,
            blob_gas_price: None,
            authorization_list: Vec::new(),
            l1_fee: None,
            l1_gas_price: None,
            l1_gas_used: None,
            l1_fee_scalar: None,
            l1_blob_base_fee: None,
            l1_blob_base_fee_scalar: None,
            gateway_fee: None,
            fee_currency: None,
            gateway_fee_recipient: None,
            beneficiary: None,
            deposit_value: None,
            l1_base_fee: None,
            max_submission_fee: None,
            refund_to: None,
            request_id: None,
            retry_data: None,
            retry_to: None,
            retry_value: None,
            max_refund: None,
            submission_fee_refund: None,
            ticket_id: None,
            is_system_tx: None,
            deposit_receipt_version: None,
            source_hash: None,
            mint: None,
        }
    }

    #[test]
    fn transaction_to_json_minimal_shape() {
        let v = transaction_to_json(&base_transaction());
        assert_eq!(v["to"], Value::Null);
        assert_eq!(
            v["chainId"],
            Value::Null,
            "chainId is always present, null when absent"
        );
        assert_eq!(
            v["yParity"],
            Value::Null,
            "yParity is always present, null when absent"
        );
        assert_eq!(
            v["accessList"],
            json!([]),
            "accessList is always present, even empty"
        );
        assert_eq!(
            v["l1Fee"],
            Value::Null,
            "l1Fee is always present, null when absent"
        );
        // empty proto `value` string is omitted entirely, never "0x0".
        assert!(v.get("value").is_none());
        assert!(v.get("r").is_none());
        assert!(v.get("s").is_none());
        assert!(v.get("v").is_none());
        assert!(v.get("blobVersionedHashes").is_none());
        assert!(v.get("authorizationList").is_none());
        assert!(
            v.get("l1GasUsed").is_none(),
            "tx.l1GasUsed omits (no null fallback)"
        );
    }

    #[test]
    fn transaction_to_json_to_field_empty_vs_present() {
        let mut tx = base_transaction();
        tx.to = Some(Bytes::new());
        assert_eq!(
            transaction_to_json(&tx)["to"],
            Value::Null,
            "present-but-empty `to` is null"
        );

        tx.to = Some(Bytes::from_static(&[0xaa; 20]));
        assert_eq!(
            transaction_to_json(&tx)["to"],
            format!("0x{}", "aa".repeat(20))
        );
    }

    #[test]
    fn transaction_to_json_value_hex_when_nonempty() {
        let mut tx = base_transaction();
        tx.value = "1000000000000000000".to_string();
        let v = transaction_to_json(&tx);
        assert_eq!(v["value"], "0xde0b6b3a7640000");
    }

    #[test]
    fn transaction_to_json_r_s_v_fixed_width_and_quantity() {
        let mut tx = base_transaction();
        tx.r = Bytes::from_static(&[0x01]);
        tx.s = Bytes::from_static(&[0x02]);
        tx.v = Some(Bytes::from_static(&[0x1b]));
        let v = transaction_to_json(&tx);
        assert_eq!(v["r"], format!("0x{}{}", "0".repeat(62), "01"));
        assert_eq!(v["s"], format!("0x{}{}", "0".repeat(62), "02"));
        assert_eq!(v["v"], "0x1b");
    }

    #[test]
    fn transaction_to_json_l1_fee_omitted_on_parse_error() {
        let mut tx = base_transaction();
        tx.l1_fee = Some("not-a-number".to_string());
        let v = transaction_to_json(&tx);
        assert!(
            v.get("l1Fee").is_none(),
            "Some(unparseable) => omitted, not null"
        );
    }

    #[test]
    fn transaction_to_json_l1_fee_null_vs_hex() {
        let tx = base_transaction();
        assert_eq!(transaction_to_json(&tx)["l1Fee"], Value::Null);

        let mut tx = base_transaction();
        tx.l1_fee = Some("100".to_string());
        assert_eq!(transaction_to_json(&tx)["l1Fee"], "0x64");
    }

    #[test]
    fn transaction_to_json_optional_arrays_present_only_when_nonempty() {
        let mut tx = base_transaction();
        tx.blob_versioned_hashes = vec![Bytes::from_static(&[0x01; 32])];
        tx.authorization_list = vec![AuthorizationListItem {
            chain_id: 1,
            address: Bytes::from_static(&[0x02; 20]),
            nonce: 1,
            r: Bytes::from_static(&[0x03]),
            s: Bytes::from_static(&[0x04]),
            y_parity: 1,
            authority: Bytes::new(),
        }];
        let v = transaction_to_json(&tx);
        assert_eq!(v["blobVersionedHashes"].as_array().unwrap().len(), 1);
        let auth = &v["authorizationList"][0];
        assert_eq!(auth["chainId"], "0x1");
        assert!(
            auth.get("authority").is_none(),
            "empty authority is omitted"
        );
    }

    // --- response mapping: receipt -----------------------------------

    fn base_receipt() -> Receipt {
        Receipt {
            transaction_hash: Bytes::from_static(&[0x01; 32]),
            block_number: 10,
            block_hash: Bytes::from_static(&[0x02; 32]),
            transaction_index: 0,
            r#type: 2,
            from: Bytes::from_static(&[0x03; 20]),
            to: None,
            status: None,
            gas_used: 21000,
            cumulative_gas_used: 21000,
            effective_gas_price: String::new(),
            logs_bloom: Bytes::new(),
            logs: Vec::new(),
            contract_address: None,
            root: None,
            block_timestamp: None,
            blob_gas_used: None,
            blob_gas_price: None,
            timeboosted: None,
            l1_fee: None,
            l1_gas_used: None,
            l1_gas_price: None,
            l1_fee_scalar: None,
            l1_base_fee_scalar: None,
            gas_used_for_l1: None,
            l1_block_number: None,
            gateway_fee: None,
            deposit_nonce: None,
            deposit_receipt_version: None,
            l1_blob_base_fee: None,
            l1_blob_base_fee_scalar: None,
            da_footprint_gas_scalar: None,
        }
    }

    #[test]
    fn receipt_to_json_status_omitted_when_absent() {
        let v = receipt_to_json(&base_receipt());
        assert!(v.get("status").is_none());
        assert_eq!(v["contractAddress"], Value::Null);
        assert_eq!(v["to"], Value::Null);
        assert_eq!(v["l1Fee"], Value::Null);
        assert_eq!(v["l1GasUsed"], Value::Null);
        assert_eq!(v["l1GasPrice"], Value::Null);
        assert!(
            v.get("gatewayFee").is_none(),
            "gatewayFee omits (no null fallback)"
        );
    }

    #[test]
    fn receipt_to_json_status_present() {
        let mut r = base_receipt();
        r.status = Some(1);
        assert_eq!(receipt_to_json(&r)["status"], "0x1");
    }

    #[test]
    fn receipt_to_json_contract_address_empty_vs_present() {
        let mut r = base_receipt();
        r.contract_address = Some(Bytes::new());
        assert_eq!(receipt_to_json(&r)["contractAddress"], Value::Null);

        r.contract_address = Some(Bytes::from_static(&[0xbb; 20]));
        assert_eq!(
            receipt_to_json(&r)["contractAddress"],
            format!("0x{}", "bb".repeat(20))
        );
    }

    #[test]
    fn receipt_to_json_timeboosted_passthrough() {
        let mut r = base_receipt();
        r.timeboosted = Some(true);
        assert_eq!(receipt_to_json(&r)["timeboosted"], true);
    }

    #[test]
    fn receipt_to_json_logs_array_present() {
        let mut r = base_receipt();
        r.logs = vec![sample_log()];
        let v = receipt_to_json(&r);
        assert_eq!(v["logs"].as_array().unwrap().len(), 1);
    }

    // --- response mapping: block -----------------------------------

    fn base_header() -> BlockHeader {
        BlockHeader {
            number: 42,
            timestamp: 1_700_000_000,
            gas_limit: 30_000_000,
            gas_used: 21000,
            size: 1000,
            hash: Bytes::from_static(&[0x01; 32]),
            parent_hash: Bytes::from_static(&[0x02; 32]),
            state_root: Bytes::from_static(&[0x03; 32]),
            transactions_root: Bytes::from_static(&[0x04; 32]),
            receipts_root: Bytes::from_static(&[0x05; 32]),
            sha3_uncles: Bytes::from_static(&[0x06; 32]),
            miner: Bytes::from_static(&[0x07; 20]),
            logs_bloom: Bytes::new(),
            extra_data: Bytes::new(),
            nonce: None,
            blob_gas_used: None,
            excess_blob_gas: None,
            l1_block_number: None,
            epoch: None,
            slot: None,
            proposer_index: None,
            send_count: None,
            transaction_count: None,
            mix_hash: None,
            parent_beacon_block_root: None,
            withdrawals_root: None,
            send_root: None,
            base_fee_per_gas: None,
            difficulty: None,
            total_difficulty: None,
            proposer_public_key: None,
            withdrawals: None,
            canonical_rlp: None,
            uncles: Vec::new(),
            requests_hash: None,
        }
    }

    #[test]
    fn block_to_json_nonce_is_fixed_16_digits() {
        let mut h = base_header();
        h.nonce = Some(0x42);
        let v = block_to_json(&h, &[], &[], &[]);
        assert_eq!(v["nonce"], "0x0000000000000042");
    }

    #[test]
    fn block_to_json_uncles_always_present() {
        let v = block_to_json(&base_header(), &[], &[], &[]);
        assert_eq!(v["uncles"], json!([]));
    }

    #[test]
    fn block_to_json_withdrawals_omitted_when_empty() {
        let v = block_to_json(&base_header(), &[], &[], &[]);
        assert!(v.get("withdrawals").is_none());

        let withdrawals = vec![Withdrawal {
            index: 1,
            validator_index: 1,
            address: Bytes::from_static(&[0xcc; 20]),
            amount: 1,
        }];
        let v = block_to_json(&base_header(), &[], &[], &withdrawals);
        assert_eq!(v["withdrawals"].as_array().unwrap().len(), 1);
    }

    #[test]
    fn block_to_json_transactions_prefers_full_over_hashes() {
        let hashes = vec![Bytes::from_static(&[0xaa; 32])];
        let v = block_to_json(&base_header(), &hashes, &[], &[]);
        assert_eq!(v["transactions"], json!([format!("0x{}", "aa".repeat(32))]));

        let full = vec![base_transaction()];
        let v = block_to_json(&base_header(), &hashes, &full, &[]);
        // full transactions win when both are populated.
        assert_eq!(v["transactions"].as_array().unwrap().len(), 1);
        assert!(v["transactions"][0].is_object());
    }

    #[test]
    fn block_to_json_transactions_empty_when_neither_present() {
        let v = block_to_json(&base_header(), &[], &[], &[]);
        assert_eq!(v["transactions"], json!([]));
    }

    #[test]
    fn block_to_json_base_fee_per_gas_decimal_to_hex() {
        let mut h = base_header();
        h.base_fee_per_gas = Some("1000000000".to_string());
        let v = block_to_json(&h, &[], &[], &[]);
        assert_eq!(v["baseFeePerGas"], "0x3b9aca00");
    }

    // --- request mapping: call_to_json_rpc round trips ------------------

    #[test]
    fn call_to_json_rpc_chain_id_round_trips() {
        let call = map_request("eth_chainId", &json!([])).unwrap();
        let (method, params) = call_to_json_rpc(&call);
        assert_eq!(method, "eth_chainId");
        assert_eq!(params, json!([]));
        let call2 = map_request(method, &params).unwrap();
        assert!(matches!(call2, RpcQueryCall::ChainId(_)));
    }

    #[test]
    fn call_to_json_rpc_get_block_by_number_round_trips() {
        for params in [json!(["0x1b4", true]), json!(["latest", false])] {
            let call = map_request("eth_getBlockByNumber", &params).unwrap();
            let (method, out_params) = call_to_json_rpc(&call);
            let call2 = map_request(method, &out_params).unwrap();
            match (call, call2) {
                (RpcQueryCall::GetBlockByNumber(a), RpcQueryCall::GetBlockByNumber(b)) => {
                    assert_eq!(a, b);
                }
                _ => panic!("wrong variant"),
            }
        }
    }

    #[test]
    fn call_to_json_rpc_get_block_by_hash_round_trips() {
        let hash = format!("0x{}", "ab".repeat(32));
        let call = map_request("eth_getBlockByHash", &json!([hash, true])).unwrap();
        let (method, out_params) = call_to_json_rpc(&call);
        let call2 = map_request(method, &out_params).unwrap();
        match (call, call2) {
            (RpcQueryCall::GetBlockByHash(a), RpcQueryCall::GetBlockByHash(b)) => assert_eq!(a, b),
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn call_to_json_rpc_get_transaction_by_hash_round_trips() {
        let hash = format!("0x{}", "cd".repeat(32));
        let call = map_request("eth_getTransactionByHash", &json!([hash])).unwrap();
        let (method, out_params) = call_to_json_rpc(&call);
        let call2 = map_request(method, &out_params).unwrap();
        match (call, call2) {
            (RpcQueryCall::GetTransactionByHash(a), RpcQueryCall::GetTransactionByHash(b)) => {
                assert_eq!(a, b);
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn call_to_json_rpc_get_transaction_receipt_round_trips() {
        let hash = format!("0x{}", "ef".repeat(32));
        let call = map_request("eth_getTransactionReceipt", &json!([hash])).unwrap();
        let (method, out_params) = call_to_json_rpc(&call);
        let call2 = map_request(method, &out_params).unwrap();
        match (call, call2) {
            (RpcQueryCall::GetTransactionReceipt(a), RpcQueryCall::GetTransactionReceipt(b)) => {
                assert_eq!(a, b);
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn call_to_json_rpc_get_block_receipts_round_trips_block_number() {
        let call = map_request("eth_getBlockReceipts", &json!(["0x10"])).unwrap();
        let (method, out_params) = call_to_json_rpc(&call);
        let call2 = map_request(method, &out_params).unwrap();
        match (call, call2) {
            (RpcQueryCall::GetBlockReceipts(a), RpcQueryCall::GetBlockReceipts(b)) => {
                assert_eq!(a, b);
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn call_to_json_rpc_get_block_receipts_round_trips_block_hash() {
        let hash = format!("0x{}", "11".repeat(32));
        let call = map_request("eth_getBlockReceipts", &json!([{ "blockHash": hash }])).unwrap();
        let (method, out_params) = call_to_json_rpc(&call);
        let call2 = map_request(method, &out_params).unwrap();
        match (call, call2) {
            (RpcQueryCall::GetBlockReceipts(a), RpcQueryCall::GetBlockReceipts(b)) => {
                assert_eq!(a, b);
            }
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn call_to_json_rpc_get_logs_round_trips_with_topics_and_address() {
        let addr = format!("0x{}", "33".repeat(20));
        let topic0 = format!("0x{}", "55".repeat(32));
        let topic2 = format!("0x{}", "66".repeat(32));
        let params = json!([{
            "fromBlock": "0x1",
            "toBlock": "0x10",
            "address": [addr],
            "topics": [topic0, Value::Null, topic2],
        }]);
        let call = map_request("eth_getLogs", &params).unwrap();
        let (method, out_params) = call_to_json_rpc(&call);
        let call2 = map_request(method, &out_params).unwrap();
        match (call, call2) {
            (RpcQueryCall::GetLogs(a), RpcQueryCall::GetLogs(b)) => assert_eq!(a, b),
            _ => panic!("wrong variant"),
        }
    }

    #[test]
    fn call_to_json_rpc_get_logs_round_trips_with_block_hash() {
        let hash = format!("0x{}", "22".repeat(32));
        let call = map_request("eth_getLogs", &json!([{ "blockHash": hash }])).unwrap();
        let (method, out_params) = call_to_json_rpc(&call);
        let call2 = map_request(method, &out_params).unwrap();
        match (call, call2) {
            (RpcQueryCall::GetLogs(a), RpcQueryCall::GetLogs(b)) => assert_eq!(a, b),
            _ => panic!("wrong variant"),
        }
    }

    // --- ingest fixtures --------------------------------------------------

    fn minimal_block_json() -> Value {
        json!({
            "number": "0x1",
            "timestamp": "0x1",
            "gasLimit": "0x1",
            "gasUsed": "0x0",
            "hash": format!("0x{}", "01".repeat(32)),
            "parentHash": format!("0x{}", "02".repeat(32)),
            "stateRoot": format!("0x{}", "03".repeat(32)),
            "transactionsRoot": format!("0x{}", "04".repeat(32)),
            "receiptsRoot": format!("0x{}", "05".repeat(32)),
            "sha3Uncles": format!("0x{}", "06".repeat(32)),
            "miner": format!("0x{}", "07".repeat(20)),
            "logsBloom": "0x",
            "extraData": "0x",
        })
    }

    fn minimal_tx_json() -> Value {
        json!({
            "hash": format!("0x{}", "01".repeat(32)),
            "nonce": "0x1",
            "from": format!("0x{}", "02".repeat(20)),
            "gas": "0x5208",
            "input": "0x",
            "r": "0x0",
            "s": "0x0",
        })
    }

    fn minimal_receipt_json() -> Value {
        json!({
            "transactionHash": format!("0x{}", "01".repeat(32)),
            "blockNumber": "0xa",
            "blockHash": format!("0x{}", "02".repeat(32)),
            "transactionIndex": "0x0",
            "from": format!("0x{}", "03".repeat(20)),
            "gasUsed": "0x5208",
            "cumulativeGasUsed": "0x5208",
            "logsBloom": "0x",
        })
    }

    fn minimal_log_json() -> Value {
        json!({
            "address": format!("0x{}", "11".repeat(20)),
            "topics": [],
            "data": "0x",
            "blockNumber": "0xa",
            "transactionHash": format!("0x{}", "01".repeat(32)),
            "transactionIndex": "0x0",
            "blockHash": format!("0x{}", "02".repeat(32)),
            "logIndex": "0x0",
        })
    }

    // --- from_json: withdrawal --------------------------------------------

    #[test]
    fn withdrawal_from_json_shape() {
        let v = json!({
            "index": "0x1",
            "validatorIndex": "0x2",
            "address": format!("0x{}", "aa".repeat(20)),
            "amount": "0x3",
        });
        let w = withdrawal_from_json(&v).unwrap();
        assert_eq!(w.index, 1);
        assert_eq!(w.validator_index, 2);
        assert_eq!(w.amount, 3);
        assert_eq!(w.address.len(), 20);
    }

    #[test]
    fn withdrawal_from_json_missing_field_errors() {
        let v = json!({});
        let err = withdrawal_from_json(&v).unwrap_err();
        assert!(matches!(err, FromJsonError::MissingField("index")));
    }

    // --- from_json: log -----------------------------------------------------

    #[test]
    fn log_from_json_block_timestamp_number_or_string_or_absent() {
        let base = minimal_log_json();
        assert!(log_from_json(&base).unwrap().block_timestamp.is_none());

        let mut with_number = base.clone();
        with_number["blockTimestamp"] = json!(1_700_000_000_u64);
        assert_eq!(
            log_from_json(&with_number).unwrap().block_timestamp,
            Some(1_700_000_000)
        );

        let mut with_string = base;
        with_string["blockTimestamp"] = json!("0x6553f100");
        assert_eq!(
            log_from_json(&with_string).unwrap().block_timestamp,
            Some(1_700_000_000)
        );
    }

    #[test]
    fn log_from_json_topics_default_empty_when_absent() {
        let mut v = minimal_log_json();
        v.as_object_mut().unwrap().remove("topics");
        assert!(log_from_json(&v).unwrap().topics.is_empty());
    }

    #[test]
    fn log_from_json_wrong_shape_errors() {
        let err = log_from_json(&json!([1, 2, 3])).unwrap_err();
        assert!(matches!(err, FromJsonError::Shape { .. }));
    }

    // --- from_json: receipt --------------------------------------------

    #[test]
    fn receipt_from_json_to_and_contract_address_empty_and_0x_mean_absent() {
        let mut v = minimal_receipt_json();
        v["to"] = json!("");
        v["contractAddress"] = json!("0x");
        let r = receipt_from_json(&v).unwrap();
        assert!(r.to.is_none());
        assert!(r.contract_address.is_none());

        let addr = format!("0x{}", "bb".repeat(20));
        let mut v2 = minimal_receipt_json();
        v2["to"] = json!(addr);
        assert_eq!(receipt_from_json(&v2).unwrap().to.unwrap().len(), 20);
    }

    #[test]
    fn receipt_from_json_type_defaults_to_zero_when_absent() {
        assert_eq!(
            receipt_from_json(&minimal_receipt_json()).unwrap().r#type,
            0
        );
    }

    #[test]
    fn receipt_from_json_status_omitted_vs_present() {
        assert!(receipt_from_json(&minimal_receipt_json())
            .unwrap()
            .status
            .is_none());
        let mut v = minimal_receipt_json();
        v["status"] = json!("0x1");
        assert_eq!(receipt_from_json(&v).unwrap().status, Some(1));
    }

    #[test]
    fn receipt_from_json_l1_quantity_strings_stored_verbatim() {
        let mut v = minimal_receipt_json();
        v["l1Fee"] = json!("0x64");
        v["l1GasUsed"] = json!("1600"); // decimal, not hex — must be preserved as given.
        v["l1GasPrice"] = json!("0x1");
        v["l1FeeScalar"] = json!("1.5");
        v["l1BlobBaseFeeScalar"] = json!(7);
        v["daFootprintGasScalar"] = json!("0x9");
        let r = receipt_from_json(&v).unwrap();
        assert_eq!(r.l1_fee.as_deref(), Some("0x64"));
        assert_eq!(r.l1_gas_used.as_deref(), Some("1600"));
        assert_eq!(r.l1_gas_price.as_deref(), Some("0x1"));
        assert_eq!(r.l1_fee_scalar, Some(1.5));
        assert_eq!(r.l1_blob_base_fee_scalar, Some(7));
        assert_eq!(r.da_footprint_gas_scalar, Some(9));
    }

    #[test]
    fn receipt_from_json_l1_fields_absent_stay_none() {
        let r = receipt_from_json(&minimal_receipt_json()).unwrap();
        assert!(r.l1_fee.is_none());
        assert!(r.l1_gas_used.is_none());
        assert!(r.l1_gas_price.is_none());
    }

    #[test]
    fn receipt_from_json_logs_array_present() {
        let mut v = minimal_receipt_json();
        v["logs"] = json!([minimal_log_json()]);
        assert_eq!(receipt_from_json(&v).unwrap().logs.len(), 1);
    }

    // --- from_json: transaction --------------------------------------------

    #[test]
    fn transaction_from_json_value_defaults_to_zero_string_when_absent() {
        assert_eq!(
            transaction_from_json(&minimal_tx_json()).unwrap().value,
            "0"
        );
    }

    #[test]
    fn transaction_from_json_to_empty_and_0x_mean_absent() {
        let mut v = minimal_tx_json();
        v["to"] = json!("");
        assert!(transaction_from_json(&v).unwrap().to.is_none());

        v["to"] = json!("0x");
        assert!(transaction_from_json(&v).unwrap().to.is_none());

        let addr = format!("0x{}", "cc".repeat(20));
        v["to"] = json!(addr);
        assert_eq!(transaction_from_json(&v).unwrap().to.unwrap().len(), 20);
    }

    #[test]
    fn transaction_from_json_type_defaults_to_zero_when_absent_or_unparseable() {
        assert_eq!(transaction_from_json(&minimal_tx_json()).unwrap().r#type, 0);

        let mut v = minimal_tx_json();
        v["type"] = json!("not-a-number");
        assert_eq!(transaction_from_json(&v).unwrap().r#type, 0);
    }

    #[test]
    fn transaction_from_json_access_list_skips_malformed_items_silently() {
        let mut v = minimal_tx_json();
        v["accessList"] = json!([
            { "address": format!("0x{}", "aa".repeat(20)), "storageKeys": [] },
            { "address": "not-hex" },
            "not-an-object",
        ]);
        let tx = transaction_from_json(&v).unwrap();
        assert_eq!(
            tx.access_list.len(),
            1,
            "only the well-formed item survives"
        );
    }

    #[test]
    fn transaction_from_json_authorization_list_errors_on_malformed_item() {
        let mut v = minimal_tx_json();
        v["authorizationList"] = json!([{
            "chainId": "0x1",
            "address": "not-hex",
            "nonce": "0x0",
            "r": "0x0",
            "s": "0x0",
            "yParity": "0x0",
        }]);
        let err = transaction_from_json(&v).unwrap_err();
        assert!(matches!(err, FromJsonError::InvalidHex { .. }));
    }

    #[test]
    fn transaction_from_json_authorization_list_authority_optional() {
        let mut v = minimal_tx_json();
        v["authorizationList"] = json!([{
            "chainId": "0x1",
            "address": format!("0x{}", "aa".repeat(20)),
            "nonce": "0x0",
            "r": "0x0",
            "s": "0x0",
            "yParity": "0x0",
        }]);
        let tx = transaction_from_json(&v).unwrap();
        assert!(tx.authorization_list[0].authority.is_empty());
    }

    #[test]
    fn transaction_from_json_l1_fee_scalar_accepts_number_or_string() {
        let mut v = minimal_tx_json();
        v["l1FeeScalar"] = json!(1.5);
        assert_eq!(transaction_from_json(&v).unwrap().l1_fee_scalar, Some(1.5));

        let mut v2 = minimal_tx_json();
        v2["l1FeeScalar"] = json!("2.5");
        assert_eq!(transaction_from_json(&v2).unwrap().l1_fee_scalar, Some(2.5));
    }

    #[test]
    fn transaction_from_json_l1_blob_base_fee_scalar_accepts_number_or_string() {
        let mut v = minimal_tx_json();
        v["l1BlobBaseFeeScalar"] = json!(9);
        assert_eq!(
            transaction_from_json(&v).unwrap().l1_blob_base_fee_scalar,
            Some(9)
        );

        let mut v2 = minimal_tx_json();
        v2["l1BlobBaseFeeScalar"] = json!("0xa");
        assert_eq!(
            transaction_from_json(&v2).unwrap().l1_blob_base_fee_scalar,
            Some(10)
        );
    }

    #[test]
    fn transaction_from_json_blob_versioned_hashes_and_signature_fields() {
        let mut v = minimal_tx_json();
        let bvh = format!("0x{}", "ee".repeat(32));
        v["blobVersionedHashes"] = json!([bvh]);
        v["v"] = json!("0x1b");
        let tx = transaction_from_json(&v).unwrap();
        assert_eq!(tx.blob_versioned_hashes.len(), 1);
        assert_eq!(tx.v.unwrap().as_ref(), &[0x1b][..]);
    }

    // --- from_json: block header ------------------------------------------

    #[test]
    fn block_header_from_json_three_field_bug_fix() {
        let v = minimal_block_json();
        let h = block_header_from_json(&v).unwrap();
        assert!(
            h.base_fee_per_gas.is_none(),
            "absent key must stay None, not a fabricated \"0x0\""
        );
        assert!(h.difficulty.is_none());
        assert!(h.total_difficulty.is_none());

        let mut v2 = minimal_block_json();
        v2["baseFeePerGas"] = json!("0x3b9aca00");
        assert_eq!(
            block_header_from_json(&v2)
                .unwrap()
                .base_fee_per_gas
                .as_deref(),
            Some("0x3b9aca00")
        );
    }

    #[test]
    fn block_header_from_json_uncles_array() {
        let mut v = minimal_block_json();
        v["uncles"] = json!([format!("0x{}", "aa".repeat(32))]);
        assert_eq!(block_header_from_json(&v).unwrap().uncles.len(), 1);
    }

    #[test]
    fn block_header_from_json_size_defaults_to_zero_when_absent() {
        assert_eq!(
            block_header_from_json(&minimal_block_json()).unwrap().size,
            0
        );
    }

    #[test]
    fn block_header_from_json_nonce_round_trips_through_fixed_width_hex() {
        let mut v = minimal_block_json();
        v["nonce"] = json!("0x42");
        let h = block_header_from_json(&v).unwrap();
        assert_eq!(h.nonce, Some(0x42));
        assert_eq!(
            block_to_json(&h, &[], &[], &[])["nonce"],
            "0x0000000000000042"
        );
    }

    // --- from_json: response wrappers ---------------------------------

    #[test]
    fn get_block_response_from_json_null_result() {
        let resp = get_block_response_from_json(&Value::Null).unwrap();
        assert!(resp.block.is_none());
    }

    #[test]
    fn get_block_response_from_json_transaction_hashes_vs_objects() {
        let hash = format!("0x{}", "aa".repeat(32));
        let mut v = minimal_block_json();
        v["transactions"] = json!([hash]);
        let resp = get_block_response_from_json(&v).unwrap();
        assert_eq!(resp.transactions.len(), 1);
        assert!(resp.full_transactions.is_empty());

        let mut v2 = minimal_block_json();
        v2["transactions"] = json!([minimal_tx_json()]);
        let resp2 = get_block_response_from_json(&v2).unwrap();
        assert!(resp2.transactions.is_empty());
        assert_eq!(resp2.full_transactions.len(), 1);
        // Backfilled from the block header since the tx object omits them.
        assert_eq!(
            resp2.full_transactions[0].block_number,
            Some(resp2.block.as_ref().unwrap().number)
        );
    }

    #[test]
    fn get_block_response_from_json_withdrawals() {
        let mut v = minimal_block_json();
        v["withdrawals"] = json!([{
            "index": "0x1",
            "validatorIndex": "0x1",
            "address": format!("0x{}", "cc".repeat(20)),
            "amount": "0x1",
        }]);
        assert_eq!(
            get_block_response_from_json(&v).unwrap().withdrawals.len(),
            1
        );
    }

    #[test]
    fn get_transaction_by_hash_response_from_json_null_and_present() {
        assert!(get_transaction_by_hash_response_from_json(&Value::Null)
            .unwrap()
            .transaction
            .is_none());
        let resp = get_transaction_by_hash_response_from_json(&minimal_tx_json()).unwrap();
        assert!(resp.transaction.is_some());
    }

    #[test]
    fn get_transaction_receipt_response_from_json_null_and_present() {
        assert!(get_transaction_receipt_response_from_json(&Value::Null)
            .unwrap()
            .receipt
            .is_none());
        let resp = get_transaction_receipt_response_from_json(&minimal_receipt_json()).unwrap();
        assert!(resp.receipt.is_some());
    }

    #[test]
    fn get_logs_response_from_json_expects_array() {
        let resp = get_logs_response_from_json(&json!([minimal_log_json()])).unwrap();
        assert_eq!(resp.logs.len(), 1);
        assert!(matches!(
            get_logs_response_from_json(&json!({})).unwrap_err(),
            FromJsonError::Shape { .. }
        ));
    }

    #[test]
    fn get_block_receipts_response_from_json_expects_array() {
        let resp = get_block_receipts_response_from_json(&json!([minimal_receipt_json()])).unwrap();
        assert_eq!(resp.receipts.len(), 1);
        assert!(matches!(
            get_block_receipts_response_from_json(&json!({})).unwrap_err(),
            FromJsonError::Shape { .. }
        ));
    }

    // --- from_json error messages -----------------------------------------

    #[test]
    fn from_json_error_messages_are_useful() {
        let err = withdrawal_from_json(&json!({})).unwrap_err();
        assert_eq!(err.to_string(), "missing required field `index`");

        let err = withdrawal_from_json(&json!({
            "index": "0x1", "validatorIndex": "0x1", "address": "zz", "amount": "0x1",
        }))
        .unwrap_err();
        assert!(err.to_string().contains("not valid hex"));

        let err = withdrawal_from_json(&json!("not-an-object")).unwrap_err();
        assert!(err.to_string().contains("found a string"));
    }

    // --- full JSON → proto → JSON round trip ------------------------------

    #[test]
    fn full_block_json_round_trip_is_semantically_stable() {
        let tx = json!({
            "hash": format!("0x{}", "aa".repeat(32)),
            "nonce": "0x5",
            "from": format!("0x{}", "bb".repeat(20)),
            "to": format!("0x{}", "cc".repeat(20)),
            "value": "1000000000000000000",
            "input": "0xdeadbeef",
            "gas": "0x5208",
            "gasPrice": "0x3b9aca00",
            "type": "0x0",
            "r": format!("0x{}", "01".repeat(32)),
            "s": format!("0x{}", "02".repeat(32)),
            "v": "0x1b",
            "gasUsed": "0x5208",
            "effectiveGasPrice": "0x3b9aca00",
            "l1Fee": "0x64",
        });

        let mut block = minimal_block_json();
        {
            let o = block.as_object_mut().unwrap();
            o.insert("number".into(), json!("0x2a"));
            o.insert("timestamp".into(), json!("0x64c0ffee"));
            o.insert("gasLimit".into(), json!("0x1c9c380"));
            o.insert("gasUsed".into(), json!("0x5208"));
            o.insert("size".into(), json!("0x3e8"));
            o.insert("nonce".into(), json!("0x42"));
            o.insert("baseFeePerGas".into(), json!("0x3b9aca00"));
            o.insert(
                "withdrawalsRoot".into(),
                json!(format!("0x{}", "08".repeat(32))),
            );
            o.insert("uncles".into(), json!([]));
            o.insert("transactions".into(), json!([tx]));
            o.insert(
                "withdrawals".into(),
                json!([{
                    "index": "0x1", "validatorIndex": "0x2",
                    "address": format!("0x{}", "cc".repeat(20)), "amount": "0x3",
                }]),
            );
        }

        let resp = get_block_response_from_json(&block).unwrap();
        let header = resp.block.as_ref().unwrap();
        let out = block_to_json(
            header,
            &resp.transactions,
            &resp.full_transactions,
            &resp.withdrawals,
        );

        assert_eq!(out["number"], "0x2a");
        assert_eq!(out["nonce"], "0x0000000000000042");
        assert_eq!(out["baseFeePerGas"], "0x3b9aca00");
        assert!(
            out.get("difficulty").is_none(),
            "absent difficulty stays omitted, not fabricated 0x0"
        );
        assert!(out.get("totalDifficulty").is_none());
        assert_eq!(out["withdrawalsRoot"], format!("0x{}", "08".repeat(32)));
        assert_eq!(out["uncles"], json!([]));
        assert_eq!(out["withdrawals"][0]["index"], "0x1");
        assert_eq!(out["withdrawals"][0]["amount"], "0x3");

        let out_tx = &out["transactions"][0];
        assert_eq!(out_tx["hash"], format!("0x{}", "aa".repeat(32)));
        assert_eq!(out_tx["value"], "0xde0b6b3a7640000");
        assert_eq!(out_tx["to"], format!("0x{}", "cc".repeat(20)));
        assert_eq!(out_tx["gasPrice"], "0x3b9aca00");
        assert_eq!(out_tx["r"], format!("0x{}", "01".repeat(32)));
        assert_eq!(out_tx["s"], format!("0x{}", "02".repeat(32)));
        assert_eq!(out_tx["v"], "0x1b");
        assert_eq!(out_tx["l1Fee"], "0x64");
        assert_eq!(out_tx["type"], "0x0");
        assert_eq!(out_tx["effectiveGasPrice"], "0x3b9aca00");
    }
}
