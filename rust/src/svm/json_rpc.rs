//! Solana JSON-RPC ⇄ BDS mapping.
//!
//! Maps the Solana history methods Agave serves from its ledger (`getBlock`,
//! `getTransaction`, `getSignatureStatuses` with history search,
//! `getBlocksWithLimit`, `getBlockTime`, `getFirstAvailableBlock`) onto
//! `bds.svm.RPCQueryService` ([`map_request`]), and each BDS response back
//! into the exact result Agave renders ([`RpcQueryCall::execute`],
//! [`Reply`], [`block_to_json`]). It is
//! the Rust counterpart of this repository's Go `svm/json_rpc.go`, and keeps
//! its scope: the `encoding: "json"` form only. Instructions stay compiled
//! (`programIdIndex` / `accounts` / base58 `data`); `jsonParsed`, `base64` and
//! `base58` are refused by the mapper, so a caller forwards them.
//!
//! Two places differ from the Go, both where the Go cannot see what the client
//! asked for: an empty `transactions` or `rewards` list still renders `[]`
//! when the request asked for it (Agave keys those on the request, not on
//! emptiness), and `uiAmount` follows the branch Agave took when it wrote the
//! balance (see [`ui_amount`]).
//!
//! Null-vs-empty fidelity: Agave distinguishes `null` from `[]` on several
//! meta fields for pre-recording history. The proto carries that as `*None`
//! booleans; a set flag renders `null`, an empty list renders `[]`.

use std::time::Duration;

use bytes::Bytes;
use serde_json::{Map, Value};

use super::rpc_query_service_client::RpcQueryServiceClient;
use super::{
    Commitment, CompiledInstruction, ConfirmedBlock, ConfirmedTransaction, GetBlockRequest,
    GetBlockResponse, GetBlockTimeRequest, GetBlockTimeResponse, GetBlocksWithLimitRequest,
    GetFirstAvailableBlockRequest, GetSignatureStatusesRequest, GetSignatureStatusesResponse,
    GetTransactionRequest, GetTransactionResponse, Message, Reward, SlotStatus, TokenBalance,
    TransactionDetails, TransactionStatusMeta, UiTokenAmount,
};

/// Agave's `MAX_GET_SIGNATURE_STATUSES_QUERY_ITEMS`.
const MAX_SIGNATURE_STATUSES: usize = 256;
/// Agave's `MAX_GET_CONFIRMED_BLOCKS_RANGE`.
const MAX_BLOCKS_LIMIT: u64 = 500_000;

/// Whether `method` has a `bds.svm.RPCQueryService` mapping here.
#[must_use]
pub fn method_supported(method: &str) -> bool {
    matches!(
        method,
        "getBlock"
            | "getTransaction"
            | "getSignatureStatuses"
            | "getBlocksWithLimit"
            | "getBlockTime"
            | "getFirstAvailableBlock"
    )
}

/// An error mapping a JSON-RPC request onto a BDS call.
#[derive(Debug, thiserror::Error)]
pub enum MapError {
    /// `method` has no mapping (see [`method_supported`]).
    #[error("method has no BDS RPC equivalent")]
    UnsupportedMethod,
    /// `params` could not be mapped onto the target BDS request message, or
    /// ask for a rendering this module does not produce.
    #[error("params not expressible as a BDS request: {0}")]
    Unmappable(&'static str),
}

/// A typed, ready-to-send `RPCQueryService` call.
#[derive(Debug)]
pub enum RpcQueryCall {
    /// `getBlock`.
    GetBlock(GetBlockRequest),
    /// `getTransaction`.
    GetTransaction(GetTransactionRequest),
    /// `getSignatureStatuses` with `searchTransactionHistory: true`.
    GetSignatureStatuses(GetSignatureStatusesRequest),
    /// `getBlocksWithLimit`.
    GetBlocksWithLimit(GetBlocksWithLimitRequest),
    /// `getBlockTime`.
    GetBlockTime(GetBlockTimeRequest),
    /// `getFirstAvailableBlock`.
    GetFirstAvailableBlock(GetFirstAvailableBlockRequest),
}

/// Maps a JSON-RPC `method` + `params` pair onto a [`RpcQueryCall`].
///
/// Shapes and defaults are Agave's (`rpc/src/rpc.rs`):
///
/// - `getBlock [slot, config?]`, `config` an object or, in the deprecated
///   form Agave still accepts, a bare encoding string. Defaults
///   `encoding: "json"`, `transactionDetails: "full"`, `rewards: true`,
///   `commitment: "finalized"`.
/// - `getTransaction [signature, config?]`, `config` likewise an object or a
///   bare encoding string; `encoding: "json"`, `commitment: "finalized"`.
/// - `getSignatureStatuses [signatures, config]`: at most 256 signatures,
///   and only with `searchTransactionHistory: true`. Without it Agave reads
///   only its recent status cache and answers `null` for older history the
///   server would find, so that form is the node's to answer.
/// - `getBlocksWithLimit [startSlot, limit, config?]`: `limit` at most
///   500,000, `commitment: "finalized"` by default.
/// - `getBlockTime [slot]`, `getFirstAvailableBlock []`.
///
/// `genesisHash` and `minContextSlot` are never set on a built request: the
/// caller's connection pins the chain, and a context slot is a claim about
/// the answering node's view that only that node can check.
///
/// # Errors
///
/// [`MapError::UnsupportedMethod`] for a method [`method_supported`] does
/// not list; [`MapError::Unmappable`] for malformed params, an encoding other
/// than `json`, `transactionDetails: "accounts"`, a commitment below
/// `confirmed` (or any but Agave's three for `getSignatureStatuses`),
/// `minContextSlot`, a limit or list above Agave's cap, or a config key the
/// mapping does not know. Every one of those is the node's to answer.
pub fn map_request(method: &str, params: &Value) -> Result<RpcQueryCall, MapError> {
    match method {
        "getBlock" => map_get_block(params),
        "getTransaction" => map_get_transaction(params),
        "getSignatureStatuses" => map_get_signature_statuses(params),
        "getBlocksWithLimit" => map_get_blocks_with_limit(params),
        "getBlockTime" => map_get_block_time(params),
        "getFirstAvailableBlock" => map_get_first_available_block(params),
        _ => Err(MapError::UnsupportedMethod),
    }
}

fn map_get_block(params: &Value) -> Result<RpcQueryCall, MapError> {
    let params = params
        .as_array()
        .ok_or(MapError::Unmappable("getBlock: params must be an array"))?;
    let slot = params
        .first()
        .and_then(Value::as_u64)
        .ok_or(MapError::Unmappable(
            "getBlock: params[0] must be a slot number",
        ))?;
    let mut req = GetBlockRequest {
        slot,
        ..GetBlockRequest::default()
    };
    match params.get(1) {
        None | Some(Value::Null) => {}
        Some(Value::String(encoding)) => check_encoding(encoding)?,
        Some(Value::Object(config)) => apply_block_config(&mut req, config)?,
        Some(_) => {
            return Err(MapError::Unmappable(
                "getBlock: params[1] must be a config object or encoding string",
            ))
        }
    }
    if params.len() > 2 {
        return Err(MapError::Unmappable("getBlock: too many params"));
    }
    Ok(RpcQueryCall::GetBlock(req))
}

fn apply_block_config(
    req: &mut GetBlockRequest,
    config: &Map<String, Value>,
) -> Result<(), MapError> {
    for (key, value) in config {
        if value.is_null() {
            continue; // Agave reads `null` as the key's default.
        }
        match key.as_str() {
            "encoding" => check_encoding(
                value
                    .as_str()
                    .ok_or(MapError::Unmappable("getBlock: encoding must be a string"))?,
            )?,
            "transactionDetails" => {
                req.transaction_details = match value.as_str() {
                    Some("full") => TransactionDetails::Full,
                    Some("signatures") => TransactionDetails::Signatures,
                    Some("none") => TransactionDetails::None,
                    _ => {
                        return Err(MapError::Unmappable(
                            "getBlock: transactionDetails must be full, signatures or none",
                        ))
                    }
                } as i32;
            }
            "rewards" => {
                req.exclude_rewards = !value
                    .as_bool()
                    .ok_or(MapError::Unmappable("getBlock: rewards must be a boolean"))?;
            }
            "commitment" => req.commitment = commitment(value)? as i32,
            "maxSupportedTransactionVersion" => {
                let version = value.as_u64().and_then(|v| u8::try_from(v).ok()).ok_or(
                    MapError::Unmappable("getBlock: maxSupportedTransactionVersion must be a u8"),
                )?;
                req.max_supported_transaction_version = Some(u32::from(version));
            }
            _ => return Err(MapError::Unmappable("getBlock: unknown config key")),
        }
    }
    Ok(())
}

fn check_encoding(encoding: &str) -> Result<(), MapError> {
    if encoding == "json" {
        Ok(())
    } else {
        Err(MapError::Unmappable("only encoding json is rendered"))
    }
}

/// `confirmed` or `finalized`: Agave's history methods refuse anything
/// lower (`check_is_at_least_confirmed`), and the node renders that error.
fn commitment(value: &Value) -> Result<Commitment, MapError> {
    match value.as_str() {
        Some("finalized") => Ok(Commitment::Finalized),
        Some("confirmed") => Ok(Commitment::Confirmed),
        _ => Err(MapError::Unmappable(
            "commitment must be confirmed or finalized",
        )),
    }
}

/// `params` as a positional array of at most `max` entries; `null` or absent
/// params read as `[]`, as Agave's positional parser does.
fn positional(params: &Value, max: usize) -> Result<&[Value], MapError> {
    let params = match params {
        Value::Null => &[][..],
        Value::Array(items) => items.as_slice(),
        _ => return Err(MapError::Unmappable("params must be an array")),
    };
    if params.len() > max {
        return Err(MapError::Unmappable("too many params"));
    }
    Ok(params)
}

fn slot_param(params: &[Value], at: usize) -> Result<u64, MapError> {
    params
        .get(at)
        .and_then(Value::as_u64)
        .ok_or(MapError::Unmappable("slot must be a u64"))
}

/// A base58 transaction signature, decoded to its 64 bytes.
fn signature(value: &Value) -> Result<Bytes, MapError> {
    value
        .as_str()
        .and_then(|s| bs58::decode(s).into_vec().ok())
        .filter(|b| b.len() == 64)
        .map(Bytes::from)
        .ok_or(MapError::Unmappable("signature must be 64 base58 bytes"))
}

/// The config object at `params[at]`, absent or `null` read as empty.
fn config_param(params: &[Value], at: usize) -> Result<Option<&Map<String, Value>>, MapError> {
    match params.get(at) {
        None | Some(Value::Null) => Ok(None),
        Some(Value::Object(config)) => Ok(Some(config)),
        Some(_) => Err(MapError::Unmappable("config must be an object")),
    }
}

fn map_get_transaction(params: &Value) -> Result<RpcQueryCall, MapError> {
    let params = positional(params, 2)?;
    let mut req = GetTransactionRequest {
        signature: signature(params.first().unwrap_or(&Value::Null))?,
        ..GetTransactionRequest::default()
    };
    // Agave's `RpcEncodingConfigWrapper`: a config object, or the deprecated
    // bare encoding string.
    match params.get(1) {
        Some(Value::String(encoding)) => check_encoding(encoding)?,
        _ => {
            for (key, value) in config_param(params, 1)?.into_iter().flatten() {
                if value.is_null() {
                    continue;
                }
                match key.as_str() {
                    "encoding" => check_encoding(
                        value
                            .as_str()
                            .ok_or(MapError::Unmappable("encoding must be a string"))?,
                    )?,
                    "commitment" => req.commitment = commitment(value)? as i32,
                    "maxSupportedTransactionVersion" => {
                        let version = value.as_u64().and_then(|v| u8::try_from(v).ok()).ok_or(
                            MapError::Unmappable("maxSupportedTransactionVersion must be a u8"),
                        )?;
                        req.max_supported_transaction_version = Some(u32::from(version));
                    }
                    _ => return Err(MapError::Unmappable("unknown getTransaction config key")),
                }
            }
        }
    }
    Ok(RpcQueryCall::GetTransaction(req))
}

fn map_get_signature_statuses(params: &Value) -> Result<RpcQueryCall, MapError> {
    let params = positional(params, 2)?;
    let signatures = params
        .first()
        .and_then(Value::as_array)
        .ok_or(MapError::Unmappable("params[0] must be a signature list"))?;
    if signatures.len() > MAX_SIGNATURE_STATUSES {
        return Err(MapError::Unmappable("more signatures than Agave accepts"));
    }
    let mut history = false;
    for (key, value) in config_param(params, 1)?.into_iter().flatten() {
        if value.is_null() {
            continue;
        }
        match key.as_str() {
            "searchTransactionHistory" => {
                history = value.as_bool().ok_or(MapError::Unmappable(
                    "searchTransactionHistory must be a boolean",
                ))?;
            }
            // Picks the bank Agave reads the status cache and the response
            // context from. An entry this module renders is rooted, which
            // every commitment reports alike.
            "commitment"
                if matches!(
                    value.as_str(),
                    Some("processed" | "confirmed" | "finalized")
                ) => {}
            _ => {
                return Err(MapError::Unmappable(
                    "unknown getSignatureStatuses config key",
                ))
            }
        }
    }
    if !history {
        return Err(MapError::Unmappable(
            "getSignatureStatuses: only searchTransactionHistory: true reads history",
        ));
    }
    Ok(RpcQueryCall::GetSignatureStatuses(
        GetSignatureStatusesRequest {
            signatures: signatures.iter().map(signature).collect::<Result<_, _>>()?,
            genesis_hash: None,
        },
    ))
}

fn map_get_blocks_with_limit(params: &Value) -> Result<RpcQueryCall, MapError> {
    let params = positional(params, 3)?;
    let mut req = GetBlocksWithLimitRequest {
        start_slot: slot_param(params, 0)?,
        limit: params
            .get(1)
            .and_then(Value::as_u64)
            .filter(|&limit| limit <= MAX_BLOCKS_LIMIT)
            .ok_or(MapError::Unmappable(
                "limit must be a u64 within Agave's cap",
            ))?,
        ..GetBlocksWithLimitRequest::default()
    };
    for (key, value) in config_param(params, 2)?.into_iter().flatten() {
        if value.is_null() {
            continue;
        }
        match key.as_str() {
            "commitment" => req.commitment = commitment(value)? as i32,
            _ => {
                return Err(MapError::Unmappable(
                    "unknown getBlocksWithLimit config key",
                ))
            }
        }
    }
    Ok(RpcQueryCall::GetBlocksWithLimit(req))
}

fn map_get_block_time(params: &Value) -> Result<RpcQueryCall, MapError> {
    let params = positional(params, 1)?;
    Ok(RpcQueryCall::GetBlockTime(GetBlockTimeRequest {
        slot: slot_param(params, 0)?,
        genesis_hash: None,
    }))
}

fn map_get_first_available_block(params: &Value) -> Result<RpcQueryCall, MapError> {
    positional(params, 0)?;
    Ok(RpcQueryCall::GetFirstAvailableBlock(
        GetFirstAvailableBlockRequest::default(),
    ))
}

// --- execute -----------------------------------------------------------

/// What a call's BDS response says, rendered as far as this module can
/// render it without knowing what the caller can prove.
///
/// The caller still owns two decisions: which slots it may report as
/// finalized (a server's confirmed store reaches past what is rooted), and
/// any per-cluster history quirk (a slot range whose `blockTime` Agave never
/// recorded). Each variant carries the slots those decisions need.
#[derive(Debug, Clone, PartialEq)]
pub enum Reply {
    /// `getBlock` / `getBlockTime`: the server reports the slot skipped.
    /// Agave answers that with an error whose code depends on where the node
    /// looked (`-32007` from its ledger, `-32009` from long-term storage),
    /// so the caller picks.
    Skipped,
    /// The server gave no answer it can prove: a transaction or status it did
    /// not find (its index may simply not cover that history), an answer
    /// missing a field Agave always renders, or a transaction version the
    /// request did not opt into. Never Agave's `null`; the node decides.
    Unknown,
    /// The `getBlock` result.
    Block(Value),
    /// The `getTransaction` result, and the slot the transaction landed in.
    Transaction { slot: u64, value: Value },
    /// `getSignatureStatuses`' `value` list, every entry rendered rooted
    /// (`confirmations: null`, `confirmationStatus: "finalized"`), and each
    /// entry's slot. Wrap with [`signature_statuses_result`] once every slot
    /// is proven rooted.
    SignatureStatuses { slots: Vec<u64>, value: Vec<Value> },
    /// The `getBlocksWithLimit` result. It may be shorter than the limit
    /// where the server's view ends.
    Blocks(Vec<u64>),
    /// `getBlockTime` for a present slot; [`None`] when the server holds no
    /// time for it.
    BlockTime(Option<i64>),
    /// The `getFirstAvailableBlock` result: the server's own floor.
    FirstAvailableBlock(u64),
}

impl RpcQueryCall {
    /// Sends this call over `client`, with the same client bounds, `timeout`
    /// semantics and `bds.rpc` / `bds.to_json` spans as
    /// [`crate::evm::json_rpc::RpcQueryCall::execute`].
    ///
    /// # Errors
    ///
    /// Returns the `tonic::Status` produced by the underlying gRPC call.
    pub async fn execute<T>(
        self,
        client: &mut RpcQueryServiceClient<T>,
        timeout: Option<Duration>,
    ) -> Result<Reply, tonic::Status>
    where
        T: tonic::client::GrpcService<tonic::body::Body>,
        T::Error: Into<tonic::codegen::StdError>,
        T::ResponseBody: tonic::codegen::Body<Data = Bytes> + Send + 'static,
        <T::ResponseBody as tonic::codegen::Body>::Error: Into<tonic::codegen::StdError> + Send,
    {
        use tracing::Instrument as _;
        fn request<M>(message: M, timeout: Option<Duration>) -> tonic::Request<M> {
            let mut request = tonic::Request::new(message);
            if let Some(t) = timeout {
                request.set_timeout(t);
            }
            request
        }
        let rpc = || tracing::info_span!("bds.rpc");
        let to_json = tracing::info_span!("bds.to_json");
        match self {
            Self::GetBlock(req) => {
                let options = BlockOptions::from(&req);
                let resp = client
                    .get_block(request(req, timeout))
                    .instrument(rpc())
                    .await?
                    .into_inner();
                Ok(to_json.in_scope(|| {
                    get_block_response_to_json(&resp, options).map_or(Reply::Skipped, Reply::Block)
                }))
            }
            Self::GetTransaction(req) => {
                let wanted = req.signature.clone();
                let max_version = req.max_supported_transaction_version;
                let resp = client
                    .get_transaction(request(req, timeout))
                    .instrument(rpc())
                    .await?
                    .into_inner();
                Ok(to_json.in_scope(|| transaction_reply(&resp, &wanted, max_version)))
            }
            Self::GetSignatureStatuses(req) => {
                let asked = req.signatures.len();
                let resp = client
                    .get_signature_statuses(request(req, timeout))
                    .instrument(rpc())
                    .await?
                    .into_inner();
                Ok(to_json.in_scope(|| signature_statuses_reply(&resp, asked)))
            }
            Self::GetBlocksWithLimit(req) => {
                let resp = client
                    .get_blocks_with_limit(request(req, timeout))
                    .instrument(rpc())
                    .await?
                    .into_inner();
                Ok(Reply::Blocks(resp.slots))
            }
            Self::GetBlockTime(req) => {
                let resp = client
                    .get_block_time(request(req, timeout))
                    .instrument(rpc())
                    .await?
                    .into_inner();
                Ok(block_time_reply(&resp))
            }
            Self::GetFirstAvailableBlock(req) => {
                let resp = client
                    .get_first_available_block(request(req, timeout))
                    .instrument(rpc())
                    .await?
                    .into_inner();
                Ok(Reply::FirstAvailableBlock(resp.slot))
            }
        }
    }
}

// --- response mapping ----------------------------------------------------

/// What a `getBlock` request asked to see, which Agave renders by rather
/// than by what the block holds.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct BlockOptions {
    /// `transactionDetails`: `transactions` (full), `signatures`, or neither.
    pub details: TransactionDetails,
    /// `rewards`: the block-level list is emitted, `[]` included, only when
    /// set.
    pub rewards: bool,
    /// `maxSupportedTransactionVersion` was given. Agave renders each
    /// transaction's `version` only then: clients that predate versioned
    /// transactions get no such key.
    pub versioned: bool,
}

impl From<&GetBlockRequest> for BlockOptions {
    fn from(req: &GetBlockRequest) -> Self {
        Self {
            details: req.transaction_details(),
            rewards: !req.exclude_rewards,
            versioned: req.max_supported_transaction_version.is_some(),
        }
    }
}

/// Renders a `GetBlockResponse` as the `getBlock` result. [`None`] for a
/// skipped slot (or a present slot carrying no block, which no conforming
/// server sends).
#[must_use]
pub fn get_block_response_to_json(resp: &GetBlockResponse, options: BlockOptions) -> Option<Value> {
    if resp.slot_status() == SlotStatus::SlotSkipped {
        return None;
    }
    let block = resp.block.as_ref()?;
    Some(block_to_json(block, &resp.signatures, options))
}

/// Renders a `GetTransactionResponse` as the `getTransaction` result for a
/// request whose `maxSupportedTransactionVersion` was `max_version`. [`None`]
/// where Agave's answer cannot be proven from the response: no transaction,
/// no `transactionIndex`, an unknown version, or a version the request did
/// not opt into (Agave's `-32015`, which the node renders).
#[must_use]
pub fn get_transaction_response_to_json(
    resp: &GetTransactionResponse,
    max_version: Option<u32>,
) -> Option<Value> {
    let ct = resp.transaction.as_ref()?;
    let index = resp.transaction_index?;
    // Agave's `validate_version`.
    match (ct.version.as_deref()?, max_version) {
        ("legacy", _) => {}
        (v, Some(max)) if v.parse::<u32>().is_ok_and(|v| v <= max) => {}
        _ => return None,
    }
    let Value::Object(mut o) = confirmed_transaction_to_json(ct, max_version.is_some()) else {
        return None;
    };
    o.insert("slot".into(), resp.slot.into());
    o.insert("blockTime".into(), resp.block_time.into());
    o.insert("transactionIndex".into(), index.into());
    Some(Value::Object(o))
}

/// Renders `getSignatureStatuses`' `value` list, each entry as Agave renders
/// a rooted status from history (`confirmations: null`, `confirmationStatus:
/// "finalized"`). [`None`] if any signature was not found: the server's
/// index may not cover it where Agave's history would.
#[must_use]
pub fn signature_statuses_to_json(resp: &GetSignatureStatusesResponse) -> Option<Vec<Value>> {
    resp.statuses
        .iter()
        .map(|s| {
            if !s.found {
                return None;
            }
            let (err, status) = err_and_status(s.err.as_deref());
            let mut o = Map::new();
            o.insert("slot".into(), s.slot.into());
            o.insert("confirmations".into(), Value::Null);
            o.insert("status".into(), status);
            o.insert("err".into(), err);
            o.insert("confirmationStatus".into(), "finalized".into());
            Some(Value::Object(o))
        })
        .collect()
}

/// Wraps a [`Reply::SignatureStatuses`] value in Agave's `RpcResponse`
/// envelope, its context at `context_slot`. `apiVersion` is left out: it
/// names the software of the node that answered, and Agave's context type
/// skips it when absent.
#[must_use]
pub fn signature_statuses_result(context_slot: u64, value: Vec<Value>) -> Value {
    let mut context = Map::new();
    context.insert("slot".into(), context_slot.into());
    let mut o = Map::new();
    o.insert("context".into(), Value::Object(context));
    o.insert("value".into(), Value::Array(value));
    Value::Object(o)
}

/// The [`Reply`] for a `GetTransactionResponse` to a request for `wanted`.
///
/// The server finds a transaction by a hash of its first signature, so a
/// transaction whose first signature is not `wanted` is not the one asked
/// for, whatever the server says.
#[must_use]
pub fn transaction_reply(
    resp: &GetTransactionResponse,
    wanted: &[u8],
    max_version: Option<u32>,
) -> Reply {
    let first = resp
        .transaction
        .as_ref()
        .and_then(|ct| ct.transaction.as_ref())
        .and_then(|tx| tx.signatures.first());
    if first.map(|s| s.as_ref()) != Some(wanted) {
        return Reply::Unknown;
    }
    get_transaction_response_to_json(resp, max_version).map_or(Reply::Unknown, |value| {
        Reply::Transaction {
            slot: resp.slot,
            value,
        }
    })
}

/// The [`Reply`] for a `GetSignatureStatusesResponse` to a request for
/// `asked` signatures: [`Reply::Unknown`] unless the server answered each
/// one and found it.
#[must_use]
pub fn signature_statuses_reply(resp: &GetSignatureStatusesResponse, asked: usize) -> Reply {
    if resp.statuses.len() != asked {
        return Reply::Unknown;
    }
    signature_statuses_to_json(resp).map_or(Reply::Unknown, |value| Reply::SignatureStatuses {
        slots: resp.statuses.iter().map(|s| s.slot).collect(),
        value,
    })
}

fn block_time_reply(resp: &GetBlockTimeResponse) -> Reply {
    if resp.slot_status() == SlotStatus::SlotSkipped {
        Reply::Skipped
    } else {
        Reply::BlockTime(resp.block_time)
    }
}

/// Converts a `ConfirmedBlock` to the `getBlock` result. `signatures` is the
/// flat list a `TRANSACTION_DETAILS_SIGNATURES` response carries (the block
/// then holds no transactions). Mirrors `svm.BlockToJsonRpc`.
///
/// Agave emits `transactions`, `signatures` and `rewards` by what was asked
/// ([`BlockOptions`]), so an empty list still renders `[]`.
#[must_use]
pub fn block_to_json(block: &ConfirmedBlock, signatures: &[Bytes], options: BlockOptions) -> Value {
    let mut o = Map::new();
    o.insert("blockhash".into(), base58(&block.blockhash));
    o.insert(
        "previousBlockhash".into(),
        base58(&block.previous_blockhash),
    );
    o.insert("parentSlot".into(), block.parent_slot.into());
    // Agave always emits these two keys, `null` when unknown.
    o.insert("blockHeight".into(), block.block_height.into());
    o.insert("blockTime".into(), block.block_time.into());
    match options.details {
        TransactionDetails::Full => {
            o.insert(
                "transactions".into(),
                Value::Array(
                    block
                        .transactions
                        .iter()
                        .map(|ct| confirmed_transaction_to_json(ct, options.versioned))
                        .collect(),
                ),
            );
        }
        TransactionDetails::Signatures => {
            o.insert("signatures".into(), base58_array(signatures));
        }
        TransactionDetails::None | TransactionDetails::Accounts => {}
    }
    if options.rewards {
        o.insert("rewards".into(), rewards_to_json(&block.rewards));
    }
    if let Some(n) = block.num_reward_partitions {
        o.insert("numRewardPartitions".into(), n.into());
    }
    Value::Object(o)
}

/// Renders one entry of `getBlock.transactions`; `versioned` as in
/// [`BlockOptions`]. Mirrors `svm.ConfirmedTransactionToJsonRpc`.
#[must_use]
pub fn confirmed_transaction_to_json(ct: &ConfirmedTransaction, versioned: bool) -> Value {
    let mut o = Map::new();
    let version = ct.version.as_deref();
    if let Some(tx) = &ct.transaction {
        let mut t = Map::new();
        t.insert("signatures".into(), base58_array(&tx.signatures));
        if let Some(m) = &tx.message {
            t.insert("message".into(), message_to_json(m, version == Some("0")));
        }
        o.insert("transaction".into(), Value::Object(t));
    }
    o.insert(
        "meta".into(),
        ct.meta.as_ref().map_or(Value::Null, meta_to_json),
    );
    // Agave renders legacy as the string "legacy" and a versioned
    // transaction as its number.
    if let Some(v) = version.filter(|_| versioned) {
        let rendered = match v.parse::<u8>() {
            Ok(n) => Value::from(n),
            Err(_) => Value::String(v.to_owned()),
        };
        o.insert("version".into(), rendered);
    }
    Value::Object(o)
}

/// Renders `transaction.message`. `v0` keeps `addressTableLookups` even when
/// empty: Agave keys that field on the message version, not its length.
fn message_to_json(m: &Message, v0: bool) -> Value {
    let mut o = Map::new();
    if let Some(h) = &m.header {
        let mut header = Map::new();
        header.insert(
            "numRequiredSignatures".into(),
            h.num_required_signatures.into(),
        );
        header.insert(
            "numReadonlySignedAccounts".into(),
            h.num_readonly_signed_accounts.into(),
        );
        header.insert(
            "numReadonlyUnsignedAccounts".into(),
            h.num_readonly_unsigned_accounts.into(),
        );
        o.insert("header".into(), Value::Object(header));
    }
    o.insert("accountKeys".into(), base58_array(&m.account_keys));
    o.insert("recentBlockhash".into(), base58(&m.recent_blockhash));
    o.insert(
        "instructions".into(),
        instructions_to_json(&m.instructions, Some(1)),
    );
    if v0 || !m.address_table_lookups.is_empty() {
        let lookups = m
            .address_table_lookups
            .iter()
            .map(|l| {
                let mut lookup = Map::new();
                lookup.insert("accountKey".into(), base58(&l.account_key));
                lookup.insert("writableIndexes".into(), byte_indexes(&l.writable_indexes));
                lookup.insert("readonlyIndexes".into(), byte_indexes(&l.readonly_indexes));
                Value::Object(lookup)
            })
            .collect();
        o.insert("addressTableLookups".into(), Value::Array(lookups));
    }
    if let Some(c) = &m.transaction_config {
        // Agave renders every key, `null` when the transaction left it unset.
        let mut config = Map::new();
        config.insert("priorityFee".into(), c.priority_fee.into());
        config.insert("computeUnitLimit".into(), c.compute_unit_limit.into());
        config.insert(
            "loadedAccountsDataSizeLimit".into(),
            c.loaded_accounts_data_size_limit.into(),
        );
        config.insert("heapSize".into(), c.heap_size.into());
        o.insert("transactionConfig".into(), Value::Object(config));
    }
    Value::Object(o)
}

/// `top_level` is the stack height Agave stamps on every top-level
/// instruction whatever the store holds (`TRANSACTION_LEVEL_STACK_HEIGHT`);
/// [`None`] renders each inner instruction's stored height.
fn instructions_to_json(instructions: &[CompiledInstruction], top_level: Option<u32>) -> Value {
    Value::Array(
        instructions
            .iter()
            .map(|ci| {
                let mut o = Map::new();
                o.insert("programIdIndex".into(), ci.program_id_index.into());
                o.insert("accounts".into(), byte_indexes(&ci.accounts));
                o.insert("data".into(), base58(&ci.data));
                o.insert("stackHeight".into(), top_level.or(ci.stack_height).into());
                Value::Object(o)
            })
            .collect(),
    )
}

/// A JSON-stringified `TransactionError` (absent = success) as Agave's `err`
/// object and the deprecated `status` derived from it.
fn err_and_status(raw: Option<&str>) -> (Value, Value) {
    let mut status = Map::new();
    let err = match raw {
        Some(raw) => {
            let err = serde_json::from_str(raw).unwrap_or_else(|_| Value::String(raw.to_owned()));
            status.insert("Err".into(), err.clone());
            err
        }
        None => {
            status.insert("Ok".into(), Value::Null);
            Value::Null
        }
    };
    (err, Value::Object(status))
}

/// Renders `transaction.meta` under `encoding: "json"`. Mirrors
/// `svm.MetaToJsonRpc`.
#[must_use]
pub fn meta_to_json(m: &TransactionStatusMeta) -> Value {
    let mut o = Map::new();
    let (err, status) = err_and_status(m.err.as_deref());
    o.insert("err".into(), err);
    o.insert("status".into(), status);
    o.insert("fee".into(), m.fee.into());
    o.insert("preBalances".into(), u64_array(&m.pre_balances));
    o.insert("postBalances".into(), u64_array(&m.post_balances));
    o.insert(
        "innerInstructions".into(),
        if m.inner_instructions_none {
            Value::Null
        } else {
            Value::Array(
                m.inner_instructions
                    .iter()
                    .map(|ii| {
                        let mut inner = Map::new();
                        inner.insert("index".into(), ii.index.into());
                        inner.insert(
                            "instructions".into(),
                            instructions_to_json(&ii.instructions, None),
                        );
                        Value::Object(inner)
                    })
                    .collect(),
            )
        },
    );
    o.insert(
        "logMessages".into(),
        if m.log_messages_none {
            Value::Null
        } else {
            Value::Array(
                m.log_messages
                    .iter()
                    .map(|l| Value::String(l.clone()))
                    .collect(),
            )
        },
    );
    o.insert(
        "preTokenBalances".into(),
        token_balances_to_json(&m.pre_token_balances, m.pre_token_balances_none),
    );
    o.insert(
        "postTokenBalances".into(),
        token_balances_to_json(&m.post_token_balances, m.post_token_balances_none),
    );
    o.insert(
        "rewards".into(),
        if m.rewards_none {
            Value::Null
        } else {
            rewards_to_json(&m.rewards)
        },
    );
    let mut loaded = Map::new();
    loaded.insert(
        "writable".into(),
        base58_array(&m.loaded_writable_addresses),
    );
    loaded.insert(
        "readonly".into(),
        base58_array(&m.loaded_readonly_addresses),
    );
    o.insert("loadedAddresses".into(), Value::Object(loaded));
    // Absent return data omits the key (Agave's `or_skip`), never `null`.
    if let Some(rd) = &m.return_data {
        use base64::Engine as _;
        let mut r = Map::new();
        r.insert("programId".into(), base58(&rd.program_id));
        r.insert(
            "data".into(),
            Value::Array(vec![
                Value::String(base64::engine::general_purpose::STANDARD.encode(&rd.data)),
                Value::String("base64".into()),
            ]),
        );
        o.insert("returnData".into(), Value::Object(r));
    }
    if let Some(v) = m.compute_units_consumed {
        o.insert("computeUnitsConsumed".into(), v.into());
    }
    if let Some(v) = m.cost_units {
        o.insert("costUnits".into(), v.into());
    }
    Value::Object(o)
}

/// Renders one reward entry. Mirrors `svm.RewardToJsonRpc`.
#[must_use]
pub fn reward_to_json(r: &Reward) -> Value {
    let mut o = Map::new();
    o.insert("pubkey".into(), base58(&r.pubkey));
    o.insert("lamports".into(), r.lamports.into());
    if let Some(v) = r.post_balance {
        o.insert("postBalance".into(), v.into());
    }
    o.insert(
        "rewardType".into(),
        r.reward_type
            .as_deref()
            .map_or(Value::Null, |t| Value::String(t.to_owned())),
    );
    o.insert("commission".into(), r.commission.into());
    Value::Object(o)
}

fn rewards_to_json(rewards: &[Reward]) -> Value {
    Value::Array(rewards.iter().map(reward_to_json).collect())
}

fn token_balances_to_json(balances: &[TokenBalance], none: bool) -> Value {
    if none {
        return Value::Null;
    }
    Value::Array(balances.iter().map(token_balance_to_json).collect())
}

/// Renders one pre/postTokenBalances entry. Mirrors
/// `svm.TokenBalanceToJsonRpc`.
#[must_use]
pub fn token_balance_to_json(tb: &TokenBalance) -> Value {
    let mut o = Map::new();
    o.insert("accountIndex".into(), tb.account_index.into());
    o.insert("mint".into(), base58(&tb.mint));
    if let Some(owner) = &tb.owner {
        o.insert("owner".into(), base58(owner));
    }
    if let Some(program) = &tb.program_id {
        o.insert("programId".into(), base58(program));
    }
    if let Some(a) = &tb.ui_token_amount {
        o.insert("uiTokenAmount".into(), ui_token_amount_to_json(a));
    }
    Value::Object(o)
}

fn ui_token_amount_to_json(a: &UiTokenAmount) -> Value {
    let plain = a.amount.parse::<u64>().ok().and_then(|amount| {
        u8::try_from(a.decimals)
            .ok()
            .map(|d| real_number_string_trimmed(amount, d))
    });
    let ui_amount_string = a
        .ui_amount_string
        .clone()
        .or_else(|| plain.clone())
        .unwrap_or_else(|| a.amount.clone());
    let mut o = Map::new();
    o.insert("amount".into(), Value::String(a.amount.clone()));
    o.insert("decimals".into(), a.decimals.into());
    o.insert(
        "uiAmount".into(),
        ui_amount(
            a,
            plain.as_deref() == Some(ui_amount_string.as_str()),
            &ui_amount_string,
        )
        .map_or(Value::Null, Value::from),
    );
    o.insert("uiAmountString".into(), Value::String(ui_amount_string));
    Value::Object(o)
}

/// The lossy `uiAmount` float, reconstructed from what the proto keeps.
///
/// Agave's `token_amount_to_ui_amount_v3` computes it one of two ways, and the
/// string says which: a plain mint's `uiAmountString` is the trimmed decimal
/// of `amount` and its float is `amount as f64 / 10^decimals` (double rounding
/// included — do not "fix" it to an exact parse); an interest-bearing or
/// scaled mint's string is something else, and its float is that string
/// parsed. Storage then turns `0.0` into `None` (`storage-proto`
/// `convert.rs`), which is why every zero balance in a node's `getBlock`
/// output carries `uiAmount: null`.
fn ui_amount(a: &UiTokenAmount, plain: bool, ui_amount_string: &str) -> Option<f64> {
    let value = if plain {
        let amount = a.amount.parse::<u64>().ok()?;
        let divisor = 10_u64.checked_pow(a.decimals)?;
        #[allow(clippy::cast_precision_loss)]
        let value = amount as f64 / divisor as f64;
        value
    } else {
        ui_amount_string.parse::<f64>().ok()?
    };
    (value != 0.0).then_some(value)
}

/// Agave's `real_number_string_trimmed`: `amount` scaled down by
/// `10^decimals`, trailing fraction zeros and a bare point removed.
fn real_number_string_trimmed(amount: u64, decimals: u8) -> String {
    let digits = amount.to_string();
    let decimals = usize::from(decimals);
    if decimals == 0 {
        return digits;
    }
    let mut s = String::with_capacity(digits.len().max(decimals + 1) + 1);
    for _ in digits.len()..=decimals {
        s.push('0');
    }
    s.push_str(&digits);
    s.insert(s.len() - decimals, '.');
    let trimmed = s.trim_end_matches('0').trim_end_matches('.').len();
    s.truncate(trimmed);
    s
}

fn base58(b: &[u8]) -> Value {
    Value::String(bs58::encode(b).into_string())
}

fn base58_array(items: &[Bytes]) -> Value {
    Value::Array(items.iter().map(|b| base58(b)).collect())
}

/// A packed account-index list as JSON numbers: the proto carries u8 indexes
/// as bytes.
fn byte_indexes(b: &[u8]) -> Value {
    Value::Array(b.iter().map(|&v| Value::from(v)).collect())
}

fn u64_array(vs: &[u64]) -> Value {
    Value::Array(vs.iter().map(|&v| Value::from(v)).collect())
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    fn get_block(params: Value) -> Result<GetBlockRequest, MapError> {
        match map_request("getBlock", &params)? {
            RpcQueryCall::GetBlock(req) => Ok(req),
            other => panic!("getBlock mapped to {other:?}"),
        }
    }

    const SIG: &str =
        "2f8cU2S3RAeRp8dio7PaEHUqssAuJZEM3DHWEFh9jgDGcD1xxEczQfywF7ZDbeXj2ekQshu2jpEdFgAebfoYJgjn";

    /// Each method's accepted shapes, Agave's defaults included.
    #[test]
    fn history_methods_map_agaves_shapes() {
        let sig = bs58::decode(SIG).into_vec().unwrap();
        for params in [
            json!([SIG]),
            json!([SIG, "json"]),
            json!([SIG, {"encoding": "json", "commitment": "confirmed", "maxSupportedTransactionVersion": 0}]),
        ] {
            let RpcQueryCall::GetTransaction(req) = map_request("getTransaction", &params).unwrap()
            else {
                panic!("{params}")
            };
            assert_eq!(req.signature.as_ref(), sig.as_slice());
        }
        let RpcQueryCall::GetTransaction(req) =
            map_request("getTransaction", &json!([SIG])).unwrap()
        else {
            unreachable!()
        };
        assert_eq!(req.commitment(), Commitment::Finalized);
        assert_eq!(req.max_supported_transaction_version, None);

        let RpcQueryCall::GetSignatureStatuses(req) = map_request(
            "getSignatureStatuses",
            &json!([[SIG, SIG], {"searchTransactionHistory": true, "commitment": "processed"}]),
        )
        .unwrap() else {
            unreachable!()
        };
        assert_eq!(req.signatures.len(), 2);

        let RpcQueryCall::GetBlocksWithLimit(req) =
            map_request("getBlocksWithLimit", &json!([5, 500_000])).unwrap()
        else {
            unreachable!()
        };
        assert_eq!((req.start_slot, req.limit), (5, 500_000));
        assert_eq!(req.commitment(), Commitment::Finalized);
        assert!(matches!(
            map_request("getBlockTime", &json!([9])).unwrap(),
            RpcQueryCall::GetBlockTime(GetBlockTimeRequest { slot: 9, .. })
        ));
        for params in [Value::Null, json!([])] {
            assert!(matches!(
                map_request("getFirstAvailableBlock", &params).unwrap(),
                RpcQueryCall::GetFirstAvailableBlock(_)
            ));
        }
    }

    /// Everything the node must answer: renderings refused, history the
    /// status cache would not search, and Agave's own errors.
    #[test]
    fn history_methods_refuse_what_they_cannot_answer() {
        let too_many: Vec<_> = (0..257).map(|_| SIG).collect();
        for (method, params) in [
            ("getTransaction", json!([])),
            ("getTransaction", json!(["not-a-signature"])),
            ("getTransaction", json!([SIG, "base64"])),
            ("getTransaction", json!([SIG, {"encoding": "jsonParsed"}])),
            ("getTransaction", json!([SIG, {"encoding": "base58"}])),
            ("getTransaction", json!([SIG, {"commitment": "processed"}])),
            ("getTransaction", json!([SIG, {"minContextSlot": 1}])),
            (
                "getTransaction",
                json!([SIG, {"maxSupportedTransactionVersion": 256}]),
            ),
            ("getSignatureStatuses", json!([[SIG]])),
            (
                "getSignatureStatuses",
                json!([[SIG], {"searchTransactionHistory": false}]),
            ),
            (
                "getSignatureStatuses",
                json!([[SIG], {"searchTransactionHistory": true, "minContextSlot": 1}]),
            ),
            (
                "getSignatureStatuses",
                json!([too_many, {"searchTransactionHistory": true}]),
            ),
            (
                "getSignatureStatuses",
                json!([["bad"], {"searchTransactionHistory": true}]),
            ),
            ("getBlocksWithLimit", json!([5])),
            ("getBlocksWithLimit", json!([5, 500_001])),
            (
                "getBlocksWithLimit",
                json!([5, 2, {"commitment": "processed"}]),
            ),
            ("getBlocksWithLimit", json!([5, 2, {"minContextSlot": 1}])),
            ("getBlockTime", json!([])),
            ("getBlockTime", json!([-1])),
            ("getFirstAvailableBlock", json!([1])),
        ] {
            assert!(
                matches!(map_request(method, &params), Err(MapError::Unmappable(_))),
                "{method} {params}"
            );
        }
    }

    fn found_transaction(version: &str) -> GetTransactionResponse {
        GetTransactionResponse {
            transaction: Some(ConfirmedTransaction {
                transaction: Some(super::super::Transaction {
                    signatures: vec![Bytes::from(bs58::decode(SIG).into_vec().unwrap())],
                    message: Some(Message::default()),
                }),
                meta: Some(TransactionStatusMeta::default()),
                version: Some(version.into()),
            }),
            slot: 400_000_000,
            block_time: Some(1_770_991_680),
            transaction_index: Some(3),
        }
    }

    /// A transaction is served only when the server proves it is the one
    /// asked for, carries every key Agave renders, and the request opted
    /// into its version.
    #[test]
    fn transaction_reply_serves_only_what_it_can_prove() {
        let sig = bs58::decode(SIG).into_vec().unwrap();
        let Reply::Transaction { slot, value } =
            transaction_reply(&found_transaction("legacy"), &sig, None)
        else {
            panic!("a found legacy transaction is served")
        };
        assert_eq!(slot, 400_000_000);
        assert_eq!(value["transactionIndex"], json!(3));
        assert_eq!(value["blockTime"], json!(1_770_991_680));
        assert!(
            value.get("version").is_none(),
            "unversioned clients get no version"
        );
        let Reply::Transaction { value, .. } =
            transaction_reply(&found_transaction("0"), &sig, Some(0))
        else {
            panic!("v0 within the asked version is served")
        };
        assert_eq!(value["version"], json!(0));

        let mut other = sig.clone();
        other[0] ^= 1;
        let no_index = GetTransactionResponse {
            transaction_index: None,
            ..found_transaction("legacy")
        };
        for (resp, wanted, max) in [
            (GetTransactionResponse::default(), &sig, None),
            (found_transaction("legacy"), &other, None),
            (no_index, &sig, None),
            (found_transaction("0"), &sig, None),
            (found_transaction("1"), &sig, Some(0)),
        ] {
            assert_eq!(transaction_reply(&resp, wanted, max), Reply::Unknown);
        }
    }

    /// One status the server did not find forwards the whole request: its
    /// index has holes where Agave's history does not.
    #[test]
    fn signature_statuses_reply_needs_every_entry() {
        use super::super::SignatureStatus;
        let found = |slot| SignatureStatus {
            found: true,
            slot,
            err: None,
            confirmation_status: Commitment::Finalized.into(),
        };
        let resp = GetSignatureStatusesResponse {
            statuses: vec![
                found(7),
                SignatureStatus {
                    err: Some(r#"{"InstructionError":[3,{"Custom":6239}]}"#.into()),
                    ..found(8)
                },
            ],
        };
        let Reply::SignatureStatuses { slots, value } = signature_statuses_reply(&resp, 2) else {
            panic!("all found is served")
        };
        assert_eq!(slots, [7, 8]);
        assert_eq!(
            signature_statuses_result(9, value),
            json!({"context": {"slot": 9}, "value": [
                {"slot": 7, "confirmations": null, "err": null, "status": {"Ok": null}, "confirmationStatus": "finalized"},
                {"slot": 8, "confirmations": null, "err": {"InstructionError": [3, {"Custom": 6239}]},
                 "status": {"Err": {"InstructionError": [3, {"Custom": 6239}]}}, "confirmationStatus": "finalized"}
            ]})
        );
        assert_eq!(signature_statuses_reply(&resp, 3), Reply::Unknown);
        let missing = GetSignatureStatusesResponse {
            statuses: vec![found(7), SignatureStatus::default()],
        };
        assert_eq!(signature_statuses_reply(&missing, 2), Reply::Unknown);
    }

    #[test]
    fn get_block_defaults_are_agaves() {
        let req = get_block(json!([7])).unwrap();
        assert_eq!(req.slot, 7);
        assert_eq!(req.transaction_details(), TransactionDetails::Full);
        assert_eq!(req.commitment(), Commitment::Finalized);
        assert!(!req.exclude_rewards);
        assert_eq!(req.max_supported_transaction_version, None);
    }

    #[test]
    fn get_block_maps_config() {
        let req = get_block(json!([7, {
            "encoding": "json",
            "transactionDetails": "signatures",
            "rewards": false,
            "commitment": "confirmed",
            "maxSupportedTransactionVersion": 0
        }]))
        .unwrap();
        assert_eq!(req.transaction_details(), TransactionDetails::Signatures);
        assert_eq!(req.commitment(), Commitment::Confirmed);
        assert!(req.exclude_rewards);
        assert_eq!(req.max_supported_transaction_version, Some(0));
        assert!(get_block(json!([7, "json"])).is_ok());
    }

    #[test]
    fn get_block_refuses_what_it_cannot_render() {
        for params in [
            json!([]),
            json!(["7"]),
            json!([7, {"encoding": "jsonParsed"}]),
            json!([7, {"encoding": "base64"}]),
            json!([7, "base58"]),
            json!([7, {"transactionDetails": "accounts"}]),
            json!([7, {"commitment": "processed"}]),
            json!([7, {"maxSupportedTransactionVersion": 256}]),
            json!([7, {"minContextSlot": 1}]),
        ] {
            assert!(
                matches!(get_block(params.clone()), Err(MapError::Unmappable(_))),
                "{params}"
            );
        }
        assert!(matches!(
            map_request("getSignaturesForAddress", &json!([])),
            Err(MapError::UnsupportedMethod)
        ));
    }

    #[test]
    fn empty_lists_follow_the_request() {
        let block = ConfirmedBlock::default();
        let full = block_to_json(
            &block,
            &[],
            BlockOptions {
                details: TransactionDetails::Full,
                rewards: true,
                versioned: true,
            },
        );
        assert_eq!(full["transactions"], json!([]));
        assert_eq!(full["rewards"], json!([]));
        assert_eq!(full["blockHeight"], Value::Null);
        let none = block_to_json(
            &block,
            &[],
            BlockOptions {
                details: TransactionDetails::None,
                rewards: false,
                versioned: true,
            },
        );
        assert!(none.get("transactions").is_none());
        assert!(none.get("signatures").is_none());
        assert!(none.get("rewards").is_none());
    }

    #[test]
    fn ui_amount_follows_agaves_branch() {
        let plain = UiTokenAmount {
            amount: "711190217735753".into(),
            decimals: 6,
            ui_amount_string: Some("711190217.735753".into()),
        };
        let v = ui_token_amount_to_json(&plain);
        assert_eq!(v["uiAmount"], json!(711_190_217.735_753));
        let scaled = UiTokenAmount {
            amount: "1000".into(),
            decimals: 2,
            ui_amount_string: Some("12.5".into()),
        };
        assert_eq!(ui_token_amount_to_json(&scaled)["uiAmount"], json!(12.5));
        let zero = UiTokenAmount {
            amount: "0".into(),
            decimals: 9,
            ui_amount_string: None,
        };
        let v = ui_token_amount_to_json(&zero);
        assert_eq!(v["uiAmount"], Value::Null);
        assert_eq!(v["uiAmountString"], json!("0"));
        assert_eq!(real_number_string_trimmed(1_500_000, 6), "1.5");
        assert_eq!(real_number_string_trimmed(5, 3), "0.005");
        assert_eq!(real_number_string_trimmed(1_000_000, 6), "1");
    }
}
