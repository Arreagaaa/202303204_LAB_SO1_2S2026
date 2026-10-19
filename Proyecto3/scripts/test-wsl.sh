#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "Validando archivos YAML"
if command -v kubeconform >/dev/null 2>&1; then
  find k8s -name '*.yaml' ! -name 'falco-values.yaml' ! -name '06-custom-resources.yaml' -print0 |
    xargs -0 kubeconform -strict -ignore-missing-schemas
else
  echo "kubeconform no esta instalado; se omite la validacion YAML"
fi

echo "Probando servicios Go"
go work sync
go test ./... ./go-ingest/... ./go-ingest-client/... ./go-writer/... ./go-writer-publisher/... ./consumer/...
go vet ./... ./go-ingest/... ./go-ingest-client/... ./go-writer/... ./go-writer-publisher/... ./consumer/...

echo "Compilando API Rust"
cargo test --locked --manifest-path api-rust/Cargo.toml

echo "Validando dashboard JSON"
python3 -m json.tool operator/helm-charts/grafanainstance/dashboards/reports.json >/dev/null

echo "Validando charts Helm"
if command -v helm >/dev/null 2>&1; then
  helm lint operator/helm-charts/valkeyinstance
  helm lint operator/helm-charts/grafanainstance
  helm template reports operator/helm-charts/valkeyinstance >/dev/null
  helm template reports operator/helm-charts/grafanainstance >/dev/null
else
  echo "helm no esta instalado; se omite la validacion de charts"
fi

echo "Validacion local completada"
