# bds — Java bindings for Blockchain Data Standards

Generated messages and grpc-java stubs for every `evm/*.proto`, in package
`bds.evm`: one outer class per file (`Stream`, `Query`, `Models`, …) and one
`*Grpc` class per service (`StreamServiceGrpc`, …).

To consume `StreamService` from eRPC, follow the
[eRPC streams integration guide](https://docs.erpc.cloud/operation/streams-integration).

## Usage

The artifact is not on Maven Central yet. Install it from a checkout:

```sh
mvn -f java/pom.xml install
```

```xml
<dependency>
  <groupId>io.github.blockchain-data-standards</groupId>
  <artifactId>bds</artifactId>
  <version>0.1.0</version>
</dependency>
```

## Regenerate

```sh
bun run generate:evm:java   # = sh java/generate.sh
```

`generate.sh` downloads protoc and protoc-gen-grpc-java from Maven Central at
pinned versions into `java/.tools/`, so the output is byte-identical.
