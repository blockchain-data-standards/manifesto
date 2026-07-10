//! Rust bindings for the [Blockchain Data Standards](https://github.com/blockchain-data-standards/manifesto)
//! protobuf definitions.
//!
//! Message types, tonic clients, and tonic server traits are generated for
//! every BDS family (`bds.common`, `bds.discovery`, `bds.evm`). The
//! [`evm::json_rpc`] module additionally maps Ethereum JSON-RPC requests onto
//! `bds.evm.RPCQueryService` calls and BDS responses back into exact JSON-RPC
//! result shapes — the Rust counterpart of the repository's Go `json_rpc.go`.

/// `bds.common` — shared error details.
pub mod common {
    #![allow(clippy::all, clippy::pedantic, missing_docs)]
    include!(concat!(env!("OUT_DIR"), "/bds.common.rs"));
}

// NOTE: `bds.discovery` bindings are intentionally absent — the proto declares
// a `repeated` field inside a `oneof` (invalid proto3) and has never compiled.

/// `bds.evm` — EVM models plus the RPC / query / bulk / stream services.
pub mod evm {
    #![allow(clippy::all, clippy::pedantic, missing_docs)]
    include!(concat!(env!("OUT_DIR"), "/bds.evm.rs"));

    pub mod json_rpc;
}
