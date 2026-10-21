//! Agave parity for the Solana history renderers.
//!
//! Each fixture in `tests/fixtures/svm/` is one live `api.mainnet-beta`
//! response (captured 2026-10-21) with the request that produced it. The
//! test lowers the result into the BDS proto the way a server stores it,
//! keeping only what the proto carries, renders it back through
//! `bds::svm::json_rpc` for the same request, and diffs the two. The proto
//! drops what the renderer has to re-derive (the `status` mirror of `err`,
//! top-level stack heights, the `uiAmount` float, null-vs-empty lists,
//! `version` gating), so a match proves the renderer re-derives each exactly
//! as Agave does. Zero differences allowed.
//!
//! Between them: a v0 transaction with an address-table lookup and loaded
//! addresses, a failed v0 transaction, a legacy transaction asked for with
//! and without `maxSupportedTransactionVersion`, and
//! `getSignatureStatuses` history over all three.

use bds::svm::json_rpc::{self as j, Reply, RpcQueryCall};
use bds::svm::{
    AddressTableLookup, CompiledInstruction, ConfirmedTransaction, GetSignatureStatusesResponse,
    GetTransactionResponse, InnerInstructions, Message, MessageHeader, ReturnData, Reward,
    SignatureStatus, TokenBalance, Transaction, TransactionStatusMeta, UiTokenAmount,
};
use bytes::Bytes;
use serde_json::Value;

fn b58(v: &Value) -> Bytes {
    Bytes::from(
        bs58::decode(v.as_str().expect("base58 string"))
            .into_vec()
            .expect("base58"),
    )
}

fn b58s(v: &Value) -> Vec<Bytes> {
    v.as_array().expect("array").iter().map(b58).collect()
}

fn u32_of(v: &Value) -> u32 {
    u32::try_from(v.as_u64().expect("u32")).expect("u32")
}

fn indexes(v: &Value) -> Bytes {
    v.as_array()
        .expect("index list")
        .iter()
        .map(|i| u8::try_from(i.as_u64().expect("index")).expect("u8 index"))
        .collect()
}

fn instruction(v: &Value) -> CompiledInstruction {
    CompiledInstruction {
        program_id_index: u32_of(&v["programIdIndex"]),
        accounts: indexes(&v["accounts"]),
        data: b58(&v["data"]),
        stack_height: v["stackHeight"]
            .as_u64()
            .map(|h| u32::try_from(h).expect("u32")),
        parsed: None,
    }
}

fn message(v: &Value) -> Message {
    let h = &v["header"];
    Message {
        header: Some(MessageHeader {
            num_required_signatures: u32_of(&h["numRequiredSignatures"]),
            num_readonly_signed_accounts: u32_of(&h["numReadonlySignedAccounts"]),
            num_readonly_unsigned_accounts: u32_of(&h["numReadonlyUnsignedAccounts"]),
        }),
        account_keys: b58s(&v["accountKeys"]),
        recent_blockhash: b58(&v["recentBlockhash"]),
        instructions: v["instructions"]
            .as_array()
            .expect("instructions")
            .iter()
            .map(instruction)
            .collect(),
        address_table_lookups: v["addressTableLookups"]
            .as_array()
            .map(|ls| {
                ls.iter()
                    .map(|l| AddressTableLookup {
                        account_key: b58(&l["accountKey"]),
                        writable_indexes: indexes(&l["writableIndexes"]),
                        readonly_indexes: indexes(&l["readonlyIndexes"]),
                    })
                    .collect()
            })
            .unwrap_or_default(),
        transaction_config: None,
    }
}

/// A list Agave may render `null` (pre-recording history): the items and
/// the proto's `*None` flag.
fn nullable<T>(v: &Value, item: impl Fn(&Value) -> T) -> (Vec<T>, bool) {
    v.as_array().map_or((Vec::new(), true), |items| {
        (items.iter().map(item).collect(), false)
    })
}

fn token_balance(v: &Value) -> TokenBalance {
    let a = &v["uiTokenAmount"];
    TokenBalance {
        account_index: u32_of(&v["accountIndex"]),
        mint: b58(&v["mint"]),
        owner: v.get("owner").map(b58),
        program_id: v.get("programId").map(b58),
        ui_token_amount: Some(UiTokenAmount {
            amount: a["amount"].as_str().expect("amount").to_owned(),
            decimals: u32_of(&a["decimals"]),
            ui_amount_string: a["uiAmountString"].as_str().map(str::to_owned),
        }),
    }
}

fn reward(v: &Value) -> Reward {
    Reward {
        pubkey: b58(&v["pubkey"]),
        lamports: v["lamports"].as_i64().expect("lamports"),
        post_balance: v["postBalance"].as_u64(),
        reward_type: v["rewardType"].as_str().map(str::to_owned),
        commission: v["commission"]
            .as_u64()
            .map(|c| u32::try_from(c).expect("u32")),
    }
}

fn meta(v: &Value) -> TransactionStatusMeta {
    let (inner_instructions, inner_instructions_none) =
        nullable(&v["innerInstructions"], |ii| InnerInstructions {
            index: u32_of(&ii["index"]),
            instructions: ii["instructions"]
                .as_array()
                .expect("inner instructions")
                .iter()
                .map(instruction)
                .collect(),
        });
    let (log_messages, log_messages_none) =
        nullable(&v["logMessages"], |l| l.as_str().expect("log").to_owned());
    let (pre_token_balances, pre_token_balances_none) =
        nullable(&v["preTokenBalances"], token_balance);
    let (post_token_balances, post_token_balances_none) =
        nullable(&v["postTokenBalances"], token_balance);
    let (rewards, rewards_none) = nullable(&v["rewards"], reward);
    let u64s = |v: &Value| {
        v.as_array()
            .expect("u64 list")
            .iter()
            .map(|n| n.as_u64().expect("u64"))
            .collect()
    };
    TransactionStatusMeta {
        err: (!v["err"].is_null()).then(|| v["err"].to_string()),
        fee: v["fee"].as_u64().expect("fee"),
        pre_balances: u64s(&v["preBalances"]),
        post_balances: u64s(&v["postBalances"]),
        inner_instructions,
        inner_instructions_none,
        log_messages,
        log_messages_none,
        pre_token_balances,
        pre_token_balances_none,
        post_token_balances,
        post_token_balances_none,
        rewards,
        rewards_none,
        loaded_writable_addresses: b58s(&v["loadedAddresses"]["writable"]),
        loaded_readonly_addresses: b58s(&v["loadedAddresses"]["readonly"]),
        return_data: v.get("returnData").map(|r| {
            use base64::Engine as _;
            ReturnData {
                program_id: b58(&r["programId"]),
                data: base64::engine::general_purpose::STANDARD
                    .decode(r["data"][0].as_str().expect("base64"))
                    .expect("base64")
                    .into(),
            }
        }),
        compute_units_consumed: v["computeUnitsConsumed"].as_u64(),
        cost_units: v["costUnits"].as_u64(),
    }
}

/// The version a server stores: what Agave renders when asked with
/// `maxSupportedTransactionVersion`; a legacy fixture asked without it
/// carries no `version`, and a v0 one could not have been served.
fn version(result: &Value) -> String {
    match &result["version"] {
        Value::Number(n) => n.to_string(),
        Value::String(s) => s.clone(),
        _ => "legacy".to_owned(),
    }
}

fn transaction_response(result: &Value) -> GetTransactionResponse {
    let tx = &result["transaction"];
    GetTransactionResponse {
        transaction: Some(ConfirmedTransaction {
            transaction: Some(Transaction {
                signatures: b58s(&tx["signatures"]),
                message: Some(message(&tx["message"])),
            }),
            meta: Some(meta(&result["meta"])),
            version: Some(version(result)),
        }),
        slot: result["slot"].as_u64().expect("slot"),
        block_time: result["blockTime"].as_i64(),
        transaction_index: result["transactionIndex"]
            .as_u64()
            .map(|i| u32::try_from(i).expect("u32")),
    }
}

/// Every path at which `got` and `want` differ.
fn diff(path: &str, got: &Value, want: &Value, out: &mut Vec<String>) {
    match (got, want) {
        (Value::Object(g), Value::Object(w)) => {
            for (k, wv) in w {
                match g.get(k) {
                    Some(gv) => diff(&format!("{path}.{k}"), gv, wv, out),
                    None => out.push(format!("{path}.{k} MISSING")),
                }
            }
            for k in g.keys().filter(|k| !w.contains_key(*k)) {
                out.push(format!("{path}.{k} EXTRA"));
            }
        }
        (Value::Array(g), Value::Array(w)) if g.len() == w.len() => {
            for (i, (gv, wv)) in g.iter().zip(w).enumerate() {
                diff(&format!("{path}[{i}]"), gv, wv, out);
            }
        }
        _ if got == want => {}
        _ => out.push(format!("{path}: got {got}, want {want}")),
    }
}

fn fixture(name: &str) -> (Value, Value) {
    let path = format!("{}/tests/fixtures/svm/{name}", env!("CARGO_MANIFEST_DIR"));
    let text = std::fs::read_to_string(&path).unwrap_or_else(|e| panic!("{path}: {e}"));
    let mut f: Value = serde_json::from_str(&text).expect("fixture JSON");
    (f["params"].take(), f["result"].take())
}

#[test]
fn get_transaction_matches_agave() {
    let names = [
        "get-transaction-v0-lookups.json",
        "get-transaction-failed.json",
        "get-transaction-legacy-unversioned.json",
        "get-transaction-legacy-versioned.json",
    ];
    let mut mismatches = Vec::new();
    for name in names {
        let (params, want) = fixture(name);
        let RpcQueryCall::GetTransaction(req) =
            j::map_request("getTransaction", &params).expect("fixture request maps")
        else {
            panic!("{name}: not a getTransaction call")
        };
        let resp = transaction_response(&want);
        let Reply::Transaction { slot, value } =
            j::transaction_reply(&resp, &req.signature, req.max_supported_transaction_version)
        else {
            panic!("{name}: a found transaction must render")
        };
        assert_eq!(Some(slot), want["slot"].as_u64(), "{name}");
        let mut out = Vec::new();
        diff("$", &value, &want, &mut out);
        mismatches.extend(out.into_iter().map(|d| format!("{name}: {d}")));
    }
    assert!(
        mismatches.is_empty(),
        "{} mismatches:\n{}",
        mismatches.len(),
        mismatches.join("\n")
    );
}

/// The v0 fixture asked without `maxSupportedTransactionVersion` is Agave's
/// `-32015`: never rendered here.
#[test]
fn a_v0_transaction_needs_the_version_opt_in() {
    let (params, want) = fixture("get-transaction-v0-lookups.json");
    let sig = params[0].clone();
    let RpcQueryCall::GetTransaction(req) =
        j::map_request("getTransaction", &Value::Array(vec![sig])).expect("maps")
    else {
        unreachable!()
    };
    let resp = transaction_response(&want);
    assert_eq!(
        j::transaction_reply(&resp, &req.signature, None),
        Reply::Unknown
    );
}

#[test]
fn get_signature_statuses_matches_agave() {
    let (params, want) = fixture("get-signature-statuses.json");
    let RpcQueryCall::GetSignatureStatuses(req) =
        j::map_request("getSignatureStatuses", &params).expect("fixture request maps")
    else {
        panic!("not a getSignatureStatuses call")
    };
    let entries = want["value"].as_array().expect("value list");
    let resp = GetSignatureStatusesResponse {
        statuses: entries
            .iter()
            .map(|e| SignatureStatus {
                found: true,
                slot: e["slot"].as_u64().expect("slot"),
                err: (!e["err"].is_null()).then(|| e["err"].to_string()),
                confirmation_status: bds::svm::Commitment::Finalized.into(),
            })
            .collect(),
    };
    let Reply::SignatureStatuses { slots, value } =
        j::signature_statuses_reply(&resp, req.signatures.len())
    else {
        panic!("every status found must render")
    };
    assert_eq!(slots.len(), entries.len());
    let context_slot = want["context"]["slot"].as_u64().expect("context slot");
    let got = j::signature_statuses_result(context_slot, value);
    let mut out = Vec::new();
    diff("$", &got, &want, &mut out);
    // `apiVersion` names the answering node's software, which a server
    // reading an archive cannot claim; Agave skips the key when absent.
    assert_eq!(
        out,
        ["$.context.apiVersion MISSING"],
        "{} mismatches",
        out.len()
    );
}
