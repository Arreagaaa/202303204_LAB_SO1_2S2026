#!/usr/bin/env bash
set -euo pipefail

cluster="${1:-mumnk8s-test}"
kind_bin="${HOME}/.local/bin/kind"

images=(
  api-rust consumer go-ingest go-ingest-client go-writer
  go-writer-publisher metrics operator
)
for name in "${images[@]}"; do
  docker tag "mumnk8s/${name}:local" "zot.example/mumnk8s/${name}:dev"
  "$kind_bin" load docker-image "zot.example/mumnk8s/${name}:dev" --name "$cluster"
done

# Las imágenes públicas de infraestructura se descargan desde los manifiestos
# locales. RabbitMQ usa un índice multiarquitectura que falla con kind load.
