//! Compiles every BDS proto family into Rust with `protox` (pure Rust — no
//! system `protoc` required) and generates tonic client **and** server stubs,
//! so the crate serves both consumers and implementers.

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // discovery/discovery.proto is deliberately absent: it declares a
    // `repeated` field inside a `oneof` (AvailabilityInfo.ranges), which is
    // invalid proto3 — it has never compiled under protoc (no generated Go
    // code exists for it either). Add it back once the schema is fixed.
    let protos = [
        "../common/errors.proto",
        "../evm/models.proto",
        "../evm/rpc.proto",
        "../evm/query.proto",
        "../evm/bulk.proto",
        "../evm/stream.proto",
    ];
    let includes = ["../common", "../evm"];

    let fds = protox::compile(protos, includes)?;

    tonic_prost_build::configure()
        .build_server(true)
        .build_client(true)
        .bytes(".")
        .compile_fds(fds)?;

    for proto in protos {
        println!("cargo::rerun-if-changed={proto}");
    }

    Ok(())
}
