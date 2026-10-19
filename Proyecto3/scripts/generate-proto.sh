#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

command -v protoc >/dev/null || { echo "Falta protoc" >&2; exit 1; }
command -v protoc-gen-go >/dev/null || { echo "Falta protoc-gen-go" >&2; exit 1; }
command -v protoc-gen-go-grpc >/dev/null || { echo "Falta protoc-gen-go-grpc" >&2; exit 1; }

protoc --proto_path=proto \
  --go_out=proto --go_opt=paths=source_relative \
  --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
  report.proto

echo "Contratos Go generados en proto/"
