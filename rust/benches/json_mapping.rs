//! What a full-transaction block costs to turn into JSON-RPC.
//!
//! This is the hot path for `edge-boost`: a cache hit decodes the BDS
//! response and maps it to JSON, and on a large block that mapping used to be
//! the single largest slice of the whole request. Keep it benched so it stays
//! fixed.
//!
//! Run with `cargo bench -p bds`.

use bds::evm::json_rpc as j;
use criterion::{criterion_group, criterion_main, BatchSize, Criterion};
use serde_json::Value;

/// Build a realistically large block from the committed fixture by repeating
/// its transactions. The fixture is trimmed for repository size; production
/// blocks run to a few hundred transactions, which is the shape that matters.
fn big_block() -> (
    bds::evm::BlockHeader,
    Vec<bds::evm::Transaction>,
    Vec<bds::evm::Withdrawal>,
) {
    let path = concat!(
        env!("CARGO_MANIFEST_DIR"),
        "/tests/fixtures/ethereum-block-full.json"
    );
    let text = std::fs::read_to_string(path).expect("fixture");
    let v: Value = serde_json::from_str(&text).expect("json");
    let resp = j::get_block_response_from_json(&v["result"]).expect("map");
    let header = resp.block.expect("header");
    let seed = resp.full_transactions;
    assert!(!seed.is_empty(), "fixture must carry full transactions");
    let mut txs = Vec::with_capacity(300);
    while txs.len() < 300 {
        txs.extend(seed.iter().cloned());
    }
    txs.truncate(300);
    (header, txs, resp.withdrawals)
}

fn bench(c: &mut Criterion) {
    let (header, txs, withdrawals) = big_block();
    let hashes: Vec<bytes::Bytes> = Vec::new();

    let mut group = c.benchmark_group("block_to_json");
    group.sample_size(40);
    group.bench_function("300 full transactions", |b| {
        b.iter(|| {
            let v = j::block_to_json(&header, &hashes, &txs, &withdrawals);
            std::hint::black_box(&v);
        });
    });
    // The second half of the real cost: handing those bytes to the caller.
    let value = j::block_to_json(&header, &hashes, &txs, &withdrawals);
    group.bench_function("serialize the mapped Value", |b| {
        b.iter_batched(
            || (),
            |()| std::hint::black_box(serde_json::to_vec(&value).expect("serialize")),
            BatchSize::SmallInput,
        );
    });
    group.finish();
}

criterion_group!(benches, bench);
criterion_main!(benches);
