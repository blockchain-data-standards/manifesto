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
    GetBlockByNumberRequest, GetBlockReceiptsRequest, GetBlockResponse, GetLogsRequest,
    GetTransactionByHashRequest, GetTransactionReceiptRequest, Log, Receipt, TopicFilter,
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
    let hash_str = param(params, 0).and_then(Value::as_str).ok_or(MapError::Unmappable(
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
    let hash_str = param(params, 0).and_then(Value::as_str).ok_or(MapError::Unmappable(
        "eth_getTransactionByHash: params[0] must be a 0x-hex transaction hash",
    ))?;
    let transaction_hash = parse_hex_bytes_exact(hash_str, 32)?;
    Ok(RpcQueryCall::GetTransactionByHash(GetTransactionByHashRequest {
        transaction_hash: transaction_hash.into(),
        chain_id: None,
        chain_genesis_hash: None,
    }))
}

fn map_get_transaction_receipt(params: &Value) -> Result<RpcQueryCall, MapError> {
    let hash_str = param(params, 0).and_then(Value::as_str).ok_or(MapError::Unmappable(
        "eth_getTransactionReceipt: params[0] must be a 0x-hex transaction hash",
    ))?;
    let transaction_hash = parse_hex_bytes_exact(hash_str, 32)?;
    Ok(RpcQueryCall::GetTransactionReceipt(GetTransactionReceiptRequest {
        transaction_hash: transaction_hash.into(),
        chain_id: None,
        chain_genesis_hash: None,
    }))
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
        .ok_or(MapError::Unmappable("eth_getLogs: params[0] must be a filter object"))?;

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
                    .ok_or(MapError::Unmappable("eth_getLogs: address array entries must be strings"))
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
                        .ok_or(MapError::Unmappable("eth_getLogs: topic array entries must be strings"))
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

// --- hex parsing (request side) --------------------------------------------

fn hex_digit_pairs_to_bytes(hex_digits: &str) -> Result<Vec<u8>, MapError> {
    if !hex_digits.bytes().all(|b| b.is_ascii_hexdigit()) {
        return Err(MapError::Unmappable("invalid hex digits"));
    }
    // Odd-length hex is padded with a leading zero nibble, mirroring Go's
    // `HexToBytes` (some RPC nodes emit values like "0x1" instead of "0x01").
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
            u8::from_str_radix(s, 16).map_err(|_| MapError::Unmappable("invalid hex digits"))
        })
        .collect()
}

fn parse_hex_bytes(s: &str) -> Result<Vec<u8>, MapError> {
    let stripped = s.strip_prefix("0x").or_else(|| s.strip_prefix("0X")).unwrap_or(s);
    hex_digit_pairs_to_bytes(stripped)
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
    /// `timeout`, when set, becomes the gRPC `grpc-timeout` deadline for the
    /// call (via [`tonic::Request::set_timeout`]).
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
    pub async fn execute(
        self,
        client: &mut RpcQueryServiceClient<tonic::transport::Channel>,
        timeout: Option<Duration>,
    ) -> Result<Option<Value>, tonic::Status> {
        match self {
            Self::ChainId(req) => {
                let resp = client.chain_id(with_timeout(req, timeout)).await?.into_inner();
                Ok(Some(Value::String(quantity_hex(resp.chain_id))))
            }
            Self::GetBlockByNumber(req) => {
                let resp = client.get_block_by_number(with_timeout(req, timeout)).await?.into_inner();
                Ok(get_block_response_to_json(&resp))
            }
            Self::GetBlockByHash(req) => {
                let resp = client.get_block_by_hash(with_timeout(req, timeout)).await?.into_inner();
                Ok(get_block_response_to_json(&resp))
            }
            Self::GetTransactionByHash(req) => {
                let resp = client.get_transaction_by_hash(with_timeout(req, timeout)).await?.into_inner();
                Ok(resp.transaction.as_ref().map(transaction_to_json))
            }
            Self::GetTransactionReceipt(req) => {
                let resp =
                    client.get_transaction_receipt(with_timeout(req, timeout)).await?.into_inner();
                Ok(resp.receipt.as_ref().map(receipt_to_json))
            }
            Self::GetLogs(req) => {
                let resp = client.get_logs(with_timeout(req, timeout)).await?.into_inner();
                Ok(Some(Value::Array(resp.logs.iter().map(log_to_json).collect())))
            }
            Self::GetBlockReceipts(req) => {
                let resp = client.get_block_receipts(with_timeout(req, timeout)).await?.into_inner();
                Ok(Some(Value::Array(resp.receipts.iter().map(receipt_to_json).collect())))
            }
        }
    }
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
    Some(block_to_json(header, &resp.transactions, &resp.full_transactions, &resp.withdrawals))
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
    o.insert("parentHash".into(), Value::String(bytes_to_hex(&header.parent_hash)));
    o.insert("sha3Uncles".into(), Value::String(bytes_to_hex(&header.sha3_uncles)));
    o.insert("logsBloom".into(), Value::String(bytes_to_hex(&header.logs_bloom)));
    o.insert("transactionsRoot".into(), Value::String(bytes_to_hex(&header.transactions_root)));
    o.insert("stateRoot".into(), Value::String(bytes_to_hex(&header.state_root)));
    o.insert("receiptsRoot".into(), Value::String(bytes_to_hex(&header.receipts_root)));
    o.insert("miner".into(), Value::String(bytes_to_hex(&header.miner)));
    o.insert("extraData".into(), Value::String(bytes_to_hex(&header.extra_data)));
    o.insert("size".into(), Value::String(quantity_hex(header.size)));
    o.insert("gasLimit".into(), Value::String(quantity_hex(header.gas_limit)));
    o.insert("gasUsed".into(), Value::String(quantity_hex(header.gas_used)));
    o.insert("timestamp".into(), Value::String(quantity_hex(header.timestamp)));

    if let Some(nonce) = header.nonce {
        // Unlike every other quantity field, the nonce is zero-padded to a
        // fixed 16 hex digits (8 bytes) — `fmt.Sprintf("0x%016x", ...)`.
        o.insert("nonce".into(), Value::String(format!("0x{nonce:016x}")));
    }
    insert_decimal_omit(&mut o, "baseFeePerGas", header.base_fee_per_gas.as_deref());
    insert_decimal_omit(&mut o, "difficulty", header.difficulty.as_deref());
    insert_decimal_omit(&mut o, "totalDifficulty", header.total_difficulty.as_deref());
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
        o.insert("parentBeaconBlockRoot".into(), Value::String(bytes_to_hex(v)));
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
        o.insert("transactionCount".into(), Value::String(quantity_hex(u64::from(v))));
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
        Value::Array(header.uncles.iter().map(|u| Value::String(bytes_to_hex(u))).collect()),
    );

    let transactions = if !full_transactions.is_empty() {
        Value::Array(full_transactions.iter().map(transaction_to_json).collect())
    } else if !transaction_hashes.is_empty() {
        Value::Array(transaction_hashes.iter().map(|h| Value::String(bytes_to_hex(h))).collect())
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
    o.insert("type".into(), Value::String(quantity_hex(u64::from(tx.r#type))));

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
        o.insert("transactionIndex".into(), Value::String(quantity_hex(u64::from(v))));
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
    insert_decimal_omit(&mut o, "maxPriorityFeePerGas", tx.max_priority_fee_per_gas.as_deref());

    if let Some(v) = tx.gas_used {
        o.insert("gasUsed".into(), Value::String(quantity_hex(v)));
    }
    insert_decimal_omit(&mut o, "effectiveGasPrice", tx.effective_gas_price.as_deref());

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
        tx.chain_id.map_or(Value::Null, |v| Value::String(quantity_hex(v))),
    );
    o.insert(
        "yParity".into(),
        tx.y_parity.map_or(Value::Null, |v| Value::String(quantity_hex(u64::from(v)))),
    );

    // Always present (possibly `[]`) — json_rpc.go pre-allocates this array
    // unconditionally, unlike blobVersionedHashes/authorizationList below.
    o.insert(
        "accessList".into(),
        Value::Array(tx.access_list.iter().map(access_list_item_to_json).collect()),
    );

    insert_decimal_omit(&mut o, "maxFeePerBlobGas", tx.max_fee_per_blob_gas.as_deref());
    if !tx.blob_versioned_hashes.is_empty() {
        o.insert(
            "blobVersionedHashes".into(),
            Value::Array(
                tx.blob_versioned_hashes.iter().map(|h| Value::String(bytes_to_hex(h))).collect(),
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
                tx.authorization_list.iter().map(authorization_list_item_to_json).collect(),
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
    insert_decimal_omit(&mut o, "submissionFeeRefund", tx.submission_fee_refund.as_deref());
    if let Some(v) = nonempty(&tx.ticket_id) {
        o.insert("ticketId".into(), Value::String(bytes_to_hex(v)));
    }

    // Base-specific fields.
    if let Some(v) = tx.is_system_tx {
        o.insert("isSystemTx".into(), Value::Bool(v));
    }
    insert_decimal_omit(&mut o, "depositReceiptVersion", tx.deposit_receipt_version.as_deref());

    Value::Object(o)
}

fn access_list_item_to_json(item: &AccessListItem) -> Value {
    let mut o = Map::new();
    o.insert("address".into(), Value::String(bytes_to_hex(&item.address)));
    o.insert(
        "storageKeys".into(),
        Value::Array(item.storage_keys.iter().map(|k| Value::String(bytes_to_hex(k))).collect()),
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
    o.insert("yParity".into(), Value::String(quantity_hex(u64::from(item.y_parity))));
    if !item.authority.is_empty() {
        o.insert("authority".into(), Value::String(bytes_to_hex(&item.authority)));
    }
    Value::Object(o)
}

/// Converts a BDS `Receipt` into the exact JSON-RPC transaction receipt
/// object shape, including its `logs` array. Mirrors `evm.ReceiptToJsonRpc`
/// in `evm/json_rpc.go`.
#[must_use]
pub fn receipt_to_json(r: &Receipt) -> Value {
    let mut o = Map::new();
    o.insert("transactionHash".into(), Value::String(bytes_to_hex(&r.transaction_hash)));
    o.insert("transactionIndex".into(), Value::String(quantity_hex(u64::from(r.transaction_index))));
    o.insert("blockHash".into(), Value::String(bytes_to_hex(&r.block_hash)));
    o.insert("blockNumber".into(), Value::String(quantity_hex(r.block_number)));
    o.insert("from".into(), Value::String(bytes_to_hex(&r.from)));
    o.insert("cumulativeGasUsed".into(), Value::String(quantity_hex(r.cumulative_gas_used)));
    o.insert("gasUsed".into(), Value::String(quantity_hex(r.gas_used)));
    o.insert("logsBloom".into(), Value::String(bytes_to_hex(&r.logs_bloom)));
    o.insert("logs".into(), Value::Array(r.logs.iter().map(log_to_json).collect()));
    // contractAddress and to are always present (null when absent/empty).
    o.insert(
        "contractAddress".into(),
        nonempty(&r.contract_address).map_or(Value::Null, |b| Value::String(bytes_to_hex(b))),
    );
    o.insert("to".into(), nonempty(&r.to).map_or(Value::Null, |b| Value::String(bytes_to_hex(b))));

    // Unlike "to"/"contractAddress" above, status is simply omitted (not
    // null) when absent.
    if let Some(v) = r.status {
        o.insert("status".into(), Value::String(quantity_hex(u64::from(v))));
    }
    o.insert("type".into(), Value::String(quantity_hex(u64::from(r.r#type))));

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
        o.insert("daFootprintGasScalar".into(), Value::String(quantity_hex(v)));
    }
    insert_decimal_omit(&mut o, "depositNonce", r.deposit_nonce.as_deref());
    insert_decimal_omit(&mut o, "depositReceiptVersion", r.deposit_receipt_version.as_deref());
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
        Value::Array(log.topics.iter().map(|t| Value::String(bytes_to_hex(t))).collect()),
    );
    o.insert("data".into(), Value::String(bytes_to_hex(&log.data)));
    o.insert("blockNumber".into(), Value::String(quantity_hex(log.block_number)));
    o.insert("transactionHash".into(), Value::String(bytes_to_hex(&log.transaction_hash)));
    o.insert("transactionIndex".into(), Value::String(quantity_hex(u64::from(log.transaction_index))));
    o.insert("blockHash".into(), Value::String(bytes_to_hex(&log.block_hash)));
    o.insert("logIndex".into(), Value::String(quantity_hex(u64::from(log.log_index))));
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
    o.insert("validatorIndex".into(), Value::String(quantity_hex(w.validator_index)));
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
fn quantity_hex(n: u64) -> String {
    format!("0x{n:x}")
}

/// `0x`-prefixed hex of raw bytes (a JSON-RPC DATA field): always
/// even-length, one nibble pair per byte, never trimmed.
fn bytes_to_hex(b: &[u8]) -> String {
    let mut s = String::with_capacity(2 + b.len() * 2);
    s.push_str("0x");
    for byte in b {
        s.push_str(&format!("{byte:02x}"));
    }
    s
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
    let hex_str: String =
        hex_digits.iter().map(|d| char::from_digit(u32::from(*d), 16).unwrap_or('0')).collect();
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

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

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
        let call =
            map_request("eth_getLogs", &logs_filter(json!({ "address": addr.clone() }))).unwrap();
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
        let err =
            map_request("eth_getLogs", &logs_filter(json!({ "address": addr }))).unwrap_err();
        assert!(matches!(err, MapError::Unmappable(_)));
    }

    #[test]
    fn get_logs_topics_null_preserves_position() {
        let topic0 = format!("0x{}", "55".repeat(32));
        let topic2 = format!("0x{}", "66".repeat(32));
        let params =
            logs_filter(json!({ "topics": [topic0, Value::Null, topic2] }));
        let call = map_request("eth_getLogs", &params).unwrap();
        match call {
            RpcQueryCall::GetLogs(req) => {
                assert_eq!(req.topics.len(), 3);
                assert_eq!(req.topics[0].values.len(), 1);
                assert!(req.topics[1].values.is_empty(), "null position must stay a wildcard");
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
        let w = Withdrawal { index: 1, validator_index: 2, address: Bytes::from_static(&[0xaa; 20]), amount: 3 };
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
        assert_eq!(v["chainId"], Value::Null, "chainId is always present, null when absent");
        assert_eq!(v["yParity"], Value::Null, "yParity is always present, null when absent");
        assert_eq!(v["accessList"], json!([]), "accessList is always present, even empty");
        assert_eq!(v["l1Fee"], Value::Null, "l1Fee is always present, null when absent");
        // empty proto `value` string is omitted entirely, never "0x0".
        assert!(v.get("value").is_none());
        assert!(v.get("r").is_none());
        assert!(v.get("s").is_none());
        assert!(v.get("v").is_none());
        assert!(v.get("blobVersionedHashes").is_none());
        assert!(v.get("authorizationList").is_none());
        assert!(v.get("l1GasUsed").is_none(), "tx.l1GasUsed omits (no null fallback)");
    }

    #[test]
    fn transaction_to_json_to_field_empty_vs_present() {
        let mut tx = base_transaction();
        tx.to = Some(Bytes::new());
        assert_eq!(transaction_to_json(&tx)["to"], Value::Null, "present-but-empty `to` is null");

        tx.to = Some(Bytes::from_static(&[0xaa; 20]));
        assert_eq!(transaction_to_json(&tx)["to"], format!("0x{}", "aa".repeat(20)));
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
        assert!(v.get("l1Fee").is_none(), "Some(unparseable) => omitted, not null");
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
        assert!(auth.get("authority").is_none(), "empty authority is omitted");
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
        assert!(v.get("gatewayFee").is_none(), "gatewayFee omits (no null fallback)");
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
        assert_eq!(receipt_to_json(&r)["contractAddress"], format!("0x{}", "bb".repeat(20)));
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

        let withdrawals =
            vec![Withdrawal { index: 1, validator_index: 1, address: Bytes::from_static(&[0xcc; 20]), amount: 1 }];
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
}
