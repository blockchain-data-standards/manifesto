#!/bin/sh
# Regenerates src/main/java from ../evm/*.proto: `sh java/generate.sh`.
# protoc and protoc-gen-grpc-java come from Maven Central at the pinned
# versions below, so the output is byte-identical on every machine.
set -eu

PROTOC_VERSION=4.36.2
GRPC_JAVA_VERSION=1.85.1

cd "$(dirname "$0")"

case "$(uname -s)-$(uname -m)" in
  Darwin-arm64) platform=osx-aarch_64 ;;
  Darwin-x86_64) platform=osx-x86_64 ;;
  Linux-aarch64) platform=linux-aarch_64 ;;
  Linux-x86_64) platform=linux-x86_64 ;;
  *) echo "generate.sh: no protoc build for $(uname -s)-$(uname -m)" >&2; exit 1 ;;
esac

fetch() { # group/path artifact version -> .tools/<artifact>-<version>
  bin=".tools/$2-$3"
  if [ ! -x "$bin" ]; then
    mkdir -p .tools
    curl -fsSL -o "$bin.part" "https://repo1.maven.org/maven2/$1/$2/$3/$2-$3-$platform.exe"
    chmod +x "$bin.part"
    mv "$bin.part" "$bin"
  fi
  echo "$bin"
}

protoc=$(fetch com/google/protobuf protoc "$PROTOC_VERSION")
plugin=$(fetch io/grpc protoc-gen-grpc-java "$GRPC_JAVA_VERSION")

rm -rf src/main/java
mkdir -p src/main/java
"$protoc" -I../evm \
  --plugin=protoc-gen-grpc-java="$plugin" \
  --java_out=src/main/java \
  --grpc-java_out=src/main/java \
  ../evm/*.proto
