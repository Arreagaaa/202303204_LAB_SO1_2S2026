#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "Validando archivos YAML"
if command -v kubeconform >/dev/null 2>&1; then
  find k8s -name '*.yaml' -print0 | xargs -0 kubeconform -strict
else
  echo "kubeconform no esta instalado; se omite la validacion YAML"
fi

echo "Compilando servicios Go"
go work sync
go build ./go-ingest
go build ./go-writer
go build ./consumer

echo "Compilando API Rust"
cargo check --manifest-path api-rust/Cargo.toml

echo "Validando charts Helm"
if command -v helm >/dev/null 2>&1; then
  helm lint operator/helm-charts/valkeyinstance
  helm lint operator/helm-charts/grafanainstance
else
  echo "helm no esta instalado; se omite la validacion de charts"
fi

echo "Validacion local completada"
