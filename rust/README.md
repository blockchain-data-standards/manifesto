# bds — Rust bindings for Blockchain Data Standards

Generated message types, tonic **clients**, and tonic **server traits** for the
BDS protobuf definitions in this repository, plus JSON-RPC mapping helpers for
the EVM family (the Rust counterpart of `evm/json_rpc.go`).

- `bds::common` — shared error details (`ErrorDetails`, `ErrorCode`).
- `bds::evm` — models (`Block`, `Transaction`, `Receipt`, `Log`, …) and the
  `RPCQueryService` / `QueryService` / `BulkQueryService` / `StreamService`
  stubs.
- `bds::evm::json_rpc` — map Ethereum JSON-RPC requests onto
  `RPCQueryService` calls and BDS responses back into exact JSON-RPC result
  shapes.

Building needs **no system `protoc`** — the protos compile at build time with
[`protox`](https://crates.io/crates/protox) (pure Rust).

## Usage

```toml
[dependencies]
bds = { git = "https://github.com/blockchain-data-standards/manifesto" }
```

Bridge a JSON-RPC request to a BDS server:

```rust,ignore
use bds::evm::json_rpc;

let call = json_rpc::map_request("eth_getBlockByNumber", &params)?;
let result: Option<serde_json::Value> = call.execute(&mut client, Some(timeout)).await?;
// Some(value) => the JSON-RPC `result`; None => the server had no such data.
```

Or implement a BDS server:

```rust,ignore
use bds::evm::rpc_query_service_server::{RpcQueryService, RpcQueryServiceServer};
```

## Notes

- `bds.discovery` bindings are intentionally absent: `discovery/discovery.proto`
  declares a `repeated` field inside a `oneof` (invalid proto3) and has never
  compiled under any protobuf toolchain. Bindings can be added once the schema
  is fixed.
- The crate is consumed as a **git dependency** for now; publishing to
  crates.io would require vendoring the `.proto` files into this directory
  (cargo packages can't reference `../`).
