//! Round-trips real production responses through the BDS mapping:
//! `JSON -> proto -> JSON`, then diffs against what production actually
//! served. This is the integrity net for any change to the mapping.
//!
//! The fixtures in `tests/fixtures/` are genuine `edge-boost` responses from
//! four chains, captured 2026-08-25 and trimmed to keep the repository small
//! while preserving every distinct shape. Between them they cover all six
//! transaction types in production today — `0x0` legacy, `0x1` (EIP-2930),
//! `0x2` (EIP-1559), `0x3` (EIP-4844, with blob fields), `0x4` (EIP-7702, with
//! authorization lists) and `0x7e` (L2 deposit) — plus withdrawals, access
//! lists, contract-creation receipts, and the OP-stack L1 fee fields.
//!
//! Every fixture must round-trip byte-identically except the one documented
//! case below. A new difference means the mapping changed behaviour.

use std::collections::BTreeSet;

use bds::evm::json_rpc as j;
use serde_json::Value;

/// The single known, intentional difference.
///
/// `transaction_from_json` falls back to the block header's timestamp when a
/// transaction carries no `blockTimestamp` of its own — an ingest-side
/// convenience, see the fallback in `json_rpc.rs`. Polygon's upstream is the
/// one provider in the corpus that omits it per transaction, so a Polygon
/// block gains the field on the way back out. This is the *inverse* direction
/// doing its job; the forward mapping is unaffected.
const KNOWN_DIFFS: &[(&str, &str)] = &[(
    "polygon-block-full.json",
    "transactions.[].blockTimestamp EXTRA-after",
)];

#[test]
fn every_production_fixture_round_trips() {
    let dir = concat!(env!("CARGO_MANIFEST_DIR"), "/tests/fixtures");
    let mut files: Vec<_> = std::fs::read_dir(dir)
        .expect("fixtures directory")
        .filter_map(|e| e.ok().map(|e| e.path()))
        .filter(|p| p.extension().is_some_and(|e| e == "json"))
        .collect();
    files.sort();
    assert!(
        files.len() >= 24,
        "expected the full corpus, found {}",
        files.len()
    );

    let mut failures = Vec::new();
    for path in files {
        let name = path
            .file_name()
            .expect("file name")
            .to_string_lossy()
            .to_string();
        let text = std::fs::read_to_string(&path).expect("read fixture");
        let envelope: Value = serde_json::from_str(&text).expect("fixture is JSON");
        let original = envelope.get("result").expect("fixture has a result");

        let Some(round) = round_trip(&name, original) else {
            failures.push(format!("{name}: the mapping refused it"));
            continue;
        };

        let mut diffs = BTreeSet::new();
        diff(&mut Vec::new(), original, &round, &mut diffs);
        for (fixture, allowed) in KNOWN_DIFFS {
            if *fixture == name {
                diffs.remove(*allowed);
            }
        }
        for d in diffs {
            failures.push(format!("{name}: {d}"));
        }
    }
    assert!(
        failures.is_empty(),
        "mapping changed behaviour:\n  {}",
        failures.join("\n  ")
    );
}

/// Every transaction type production serves must appear in the corpus, so the
/// round-trip above is actually exercising them.
#[test]
fn the_corpus_covers_every_transaction_type() {
    let dir = concat!(env!("CARGO_MANIFEST_DIR"), "/tests/fixtures");
    let mut seen = BTreeSet::new();
    for entry in std::fs::read_dir(dir).expect("fixtures directory") {
        let path = entry.expect("entry").path();
        if !path.to_string_lossy().contains("block-full") {
            continue;
        }
        let text = std::fs::read_to_string(&path).expect("read fixture");
        let v: Value = serde_json::from_str(&text).expect("fixture is JSON");
        for tx in v["result"]["transactions"].as_array().into_iter().flatten() {
            if let Some(t) = tx.get("type").and_then(Value::as_str) {
                seen.insert(t.to_owned());
            }
        }
    }
    for expected in ["0x0", "0x1", "0x2", "0x3", "0x4", "0x7e"] {
        assert!(
            seen.contains(expected),
            "corpus lost transaction type {expected}; have {seen:?}"
        );
    }
}

fn round_trip(name: &str, original: &Value) -> Option<Value> {
    if name.contains("block-full") || name.contains("block-hashes") {
        let r = j::get_block_response_from_json(original).ok()?;
        let header = r.block.as_ref()?;
        Some(j::block_to_json(
            header,
            &r.transactions,
            &r.full_transactions,
            &r.withdrawals,
        ))
    } else if name.contains("-tx") {
        let r = j::get_transaction_by_hash_response_from_json(original).ok()?;
        r.transaction.as_ref().map(j::transaction_to_json)
    } else if name.contains("-receipt") {
        let r = j::get_transaction_receipt_response_from_json(original).ok()?;
        r.receipt.as_ref().map(j::receipt_to_json)
    } else if name.contains("blockreceipts") {
        let r = j::get_block_receipts_response_from_json(original).ok()?;
        Some(Value::Array(
            r.receipts.iter().map(j::receipt_to_json).collect(),
        ))
    } else if name.contains("-logs") {
        let r = j::get_logs_response_from_json(original).ok()?;
        Some(Value::Array(r.logs.iter().map(j::log_to_json).collect()))
    } else {
        None
    }
}

/// Structural diff. Array indices collapse to `[]` so one wrong field across
/// three hundred transactions reports once instead of three hundred times.
fn diff(path: &mut Vec<String>, before: &Value, after: &Value, out: &mut BTreeSet<String>) {
    match (before, after) {
        (Value::Object(a), Value::Object(b)) => {
            for key in a.keys().chain(b.keys()).collect::<BTreeSet<_>>() {
                path.push(key.clone());
                match (a.get(key), b.get(key)) {
                    (Some(x), Some(y)) => diff(path, x, y, out),
                    (Some(_), None) => drop(out.insert(format!("{} MISSING-after", flatten(path)))),
                    (None, Some(_)) => drop(out.insert(format!("{} EXTRA-after", flatten(path)))),
                    (None, None) => {}
                }
                path.pop();
            }
        }
        (Value::Array(a), Value::Array(b)) => {
            if a.len() != b.len() {
                out.insert(format!("{} LEN {} != {}", flatten(path), a.len(), b.len()));
            }
            for (i, (x, y)) in a.iter().zip(b).enumerate() {
                path.push(format!("[{i}]"));
                diff(path, x, y, out);
                path.pop();
            }
        }
        _ => {
            if before != after {
                out.insert(format!("{} VALUE {before} != {after}", flatten(path)));
            }
        }
    }
}

fn flatten(path: &[String]) -> String {
    path.iter()
        .map(|s| if s.starts_with('[') { "[]" } else { s.as_str() })
        .collect::<Vec<_>>()
        .join(".")
}
