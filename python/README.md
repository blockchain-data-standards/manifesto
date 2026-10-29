# bds — Python bindings for Blockchain Data Standards

Generated messages (`*_pb2.py`, typed by `*_pb2.pyi`) and gRPC clients and
servicers (`*_pb2_grpc.py`) for every `evm/*.proto`, under `bds.evm`.

To consume `StreamService` from eRPC, follow the
[eRPC streams integration guide](https://docs.erpc.cloud/operation/streams-integration).

## Usage

```sh
pip install "bds @ git+https://github.com/blockchain-data-standards/manifesto#subdirectory=python"
```

```python
import grpc
from bds.evm import stream_pb2, stream_pb2_grpc

stub = stream_pb2_grpc.StreamServiceStub(grpc.insecure_channel("localhost:4000"))
```

## Regenerate

```sh
bun run generate:evm:python   # = uv run python/generate.py
```

`generate.py` pins `grpcio-tools`, which bundles protoc, so the output is
byte-identical. The package is consumed from git for now; it is not on PyPI.
