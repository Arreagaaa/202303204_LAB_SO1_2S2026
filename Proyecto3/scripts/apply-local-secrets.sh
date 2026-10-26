#!/usr/bin/env bash
set -euo pipefail

CONTEXT="${1:?Uso: bash scripts/apply-local-secrets.sh CONTEXTO_KIND}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$ROOT/scripts/local-secrets.sh"

kubectl --context "$CONTEXT" -n mumnk8s create secret generic rabbitmq-credentials \
  --from-file=username="$LOCAL_SECRET_DIR/rabbitmq-username" \
  --from-file=password="$LOCAL_SECRET_DIR/rabbitmq-password" \
  --from-file=uri="$LOCAL_SECRET_DIR/rabbitmq-uri" \
  --dry-run=client -o yaml | kubectl --context "$CONTEXT" apply -f -
kubectl --context "$CONTEXT" -n mumnk8s create secret generic grafana-admin \
  --from-file=admin-password="$LOCAL_SECRET_DIR/grafana-password" \
  --dry-run=client -o yaml | kubectl --context "$CONTEXT" apply -f -
