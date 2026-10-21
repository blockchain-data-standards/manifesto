//! Solana JSON-RPC ⇄ BDS mapping.
//!
//! Maps a Solana `getBlock` request onto `bds.svm.RPCQueryService.GetBlock`
//! ([`map_request`]) and the BDS response back into the exact `getBlock`
//! result Agave renders ([`RpcQueryCall::execute`], [`block_to_json`]). It is
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
    GetBlockResponse, Message, Reward, SlotStatus, TokenBalance, TransactionDetails,
    TransactionStatusMeta, UiTokenAmount,
};

/// Whether `method` has a `bds.svm.RPCQueryService` mapping here.
#[must_use]
pub fn method_supported(method: &str) -> bool {
    method == "getBlock"
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
}

/// Maps a JSON-RPC `method` + `params` pair onto a [`RpcQueryCall`].
///
/// `getBlock` takes `[slot, config?]`, where `config` is an object or, in the
/// deprecated form Agave still accepts, a bare encoding string. Defaults are
/// Agave's: `encoding: "json"`, `transactionDetails: "full"`, `rewards: true`,
/// `commitment: "finalized"`.
///
/// `genesisHash` and `minContextSlot` are never set on the built request: the
/// caller's connection pins the chain, and `getBlock` has no context slot.
///
/// # Errors
///
/// [`MapError::UnsupportedMethod`] for any method but `getBlock`;
/// [`MapError::Unmappable`] for a malformed slot or config, an encoding other
/// than `json`, `transactionDetails: "accounts"`, a commitment below
/// `confirmed`, or a config key Agave's `RpcBlockConfig` does not have. Every
/// one of those is the node's to answer.
pub fn map_request(method: &str, params: &Value) -> Result<RpcQueryCall, MapError> {
    match method {
        "getBlock" => map_get_block(params),
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
            "commitment" => {
                req.commitment = match value.as_str() {
                    Some("finalized") => Commitment::Finalized,
                    Some("confirmed") => Commitment::Confirmed,
                    _ => {
                        return Err(MapError::Unmappable(
                            "getBlock: commitment must be confirmed or finalized",
                        ))
                    }
                } as i32;
            }
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
        Err(MapError::Unmappable(
            "getBlock: only encoding json is rendered",
        ))
    }
}

// --- execute -----------------------------------------------------------

impl RpcQueryCall {
    /// Sends this call over `client`, with the same client bounds, `timeout`
    /// semantics and `bds.rpc` / `bds.to_json` spans as
    /// [`crate::evm::json_rpc::RpcQueryCall::execute`].
    ///
    /// `Ok(Some(_))` carries the `getBlock` result. `Ok(None)` means the
    /// server reported the slot skipped (`SLOT_SKIPPED`): Agave answers that
    /// with an error whose code depends on where the node looked (`-32007`
    /// from its ledger, `-32009` from long-term storage), so the caller picks.
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
            Self::GetBlock(req) => {
                let options = BlockOptions::from(&req);
                let mut request = tonic::Request::new(req);
                if let Some(t) = timeout {
                    request.set_timeout(t);
                }
                let resp = client
                    .get_block(request)
                    .instrument(tracing::info_span!("bds.rpc"))
                    .await?
                    .into_inner();
                Ok(tracing::info_span!("bds.to_json")
                    .in_scope(|| get_block_response_to_json(&resp, options)))
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

/// Renders `transaction.meta` under `encoding: "json"`. Mirrors
/// `svm.MetaToJsonRpc`.
#[must_use]
pub fn meta_to_json(m: &TransactionStatusMeta) -> Value {
    let mut o = Map::new();
    // `err` is carried JSON-stringified; re-inflate it so the wire shape is
    // an object, and derive the deprecated `status` from it.
    match m.err.as_deref() {
        Some(raw) => {
            let err = serde_json::from_str(raw).unwrap_or_else(|_| Value::String(raw.to_owned()));
            let mut status = Map::new();
            status.insert("Err".into(), err.clone());
            o.insert("err".into(), err);
            o.insert("status".into(), Value::Object(status));
        }
        None => {
            let mut status = Map::new();
            status.insert("Ok".into(), Value::Null);
            o.insert("err".into(), Value::Null);
            o.insert("status".into(), Value::Object(status));
        }
    }
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
        map_request("getBlock", &params).map(|RpcQueryCall::GetBlock(req)| req)
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
            map_request("getTransaction", &json!([])),
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
