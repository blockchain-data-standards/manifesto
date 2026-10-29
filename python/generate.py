# /// script
# requires-python = ">=3.10"
# dependencies = ["grpcio-tools==1.85.0"]
# ///
"""Regenerates bds/evm from ../evm/*.proto: `uv run python/generate.py`.

grpcio-tools bundles its own protoc, so the pinned version above fixes the
output byte for byte.
"""

import re
import sys
from pathlib import Path

from grpc_tools import protoc

ROOT = Path(__file__).resolve().parent.parent
PROTOS = sorted(p.name for p in (ROOT / "evm").glob("*.proto"))
OUT = ROOT / "python" / "bds" / "evm"

for stale in OUT.glob("*_pb2*"):
    stale.unlink()

status = protoc.main(
    [
        "grpc_tools.protoc",
        f"-I{ROOT / 'evm'}",
        f"--python_out={OUT}",
        f"--pyi_out={OUT}",
        f"--grpc_python_out={OUT}",
        *PROTOS,
    ]
)
if status != 0:
    sys.exit(status)

# The protos import siblings by bare name (`import "models.proto"`), which
# protoc renders as top-level `import models_pb2`. Inside the `bds.evm`
# package those must be relative.
SIBLING_IMPORT = re.compile(r"^import (\w+_pb2) as ", re.MULTILINE)
for generated in sorted(OUT.glob("*_pb2*")):
    text = generated.read_text()
    generated.write_text(SIBLING_IMPORT.sub(r"from . import \1 as ", text))
