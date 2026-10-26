#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
source "$ROOT/scripts/local-secrets.sh"
export PATH="$HOME/.local/bin:$PATH"

CLUSTER=mumnk8s-demo
CONTEXT="kind-$CLUSTER"
COMPOSE_PROJECT=mumnk8s-demo-local
FRESH=false
DASHBOARD_ONLY=false
PF_PID=""
COMPOSE_UP=false

for arg in "$@"; do
  case "$arg" in
    --fresh) FRESH=true ;;
    --dashboard-only) DASHBOARD_ONLY=true ;;
    *) echo "Uso: bash scripts/test-local-full.sh [--fresh | --dashboard-only]" >&2; exit 2 ;;
  esac
done
if [[ "$FRESH" == true && "$DASHBOARD_ONLY" == true ]]; then
  echo "Usa --fresh o --dashboard-only, no ambos" >&2
  exit 2
fi

cleanup() {
  if [[ -n "$PF_PID" ]]; then
    kill "$PF_PID" 2>/dev/null || true
    wait "$PF_PID" 2>/dev/null || true
  fi
  if [[ "$COMPOSE_UP" == true ]]; then
    docker compose -p "$COMPOSE_PROJECT" -f compose.local.yaml down >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

step() { printf '\n==> %s\n' "$*"; }
need() { command -v "$1" >/dev/null || { echo "Falta $1 en WSL" >&2; exit 1; }; }
for cmd in docker kind kubectl helm curl python3; do need "$cmd"; done

open_grafana() {
  step "Abrir dashboard de Grafana"
  local grafana_port
  grafana_port=$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()')
  kubectl --context "$CONTEXT" port-forward -n mumnk8s svc/grafana-dashboard "$grafana_port:3000" >/tmp/mumnk8s-demo-grafana.log 2>&1 &
  PF_PID=$!
  for _ in $(seq 1 20); do
    if curl -fsS --netrc-file "$LOCAL_SECRET_DIR/grafana.netrc" "http://127.0.0.1:$grafana_port/api/health" >/dev/null 2>&1; then break; fi
    kill -0 "$PF_PID" 2>/dev/null || { cat /tmp/mumnk8s-demo-grafana.log >&2; exit 1; }
    sleep 1
  done
  curl -fsS --netrc-file "$LOCAL_SECRET_DIR/grafana.netrc" "http://127.0.0.1:$grafana_port/api/search" |
    python3 -c 'import json,sys; assert any(d.get("uid") == "mumnk8s-reports" for d in json.load(sys.stdin)), "Dashboard no encontrado"'
  if [[ "${1:-}" == full ]]; then
    echo "PRUEBA LOCAL COMPLETA: codigo, Compose, Gateway, Locust, HPA, RabbitMQ, Valkey y Grafana OK"
  fi
  echo "Abre Grafana en Windows: http://localhost:$grafana_port/d/mumnk8s-reports/m-u-m-n-k8s-reportes-militares?orgId=1&refresh=10s"
  echo "Usuario: admin | contrasena local en $LOCAL_SECRET_DIR/grafana-password"
  echo "Deja esta terminal abierta mientras ves Grafana; Ctrl+C cierra el acceso local."
  wait "$PF_PID"
}

if [[ "$DASHBOARD_ONLY" == true ]]; then
  open_grafana
  exit 0
fi

docker info >/dev/null

step "Pruebas de codigo Go, Rust, Helm y dashboard"
bash scripts/test-wsl.sh

step "Docker Compose: construir, arrancar y probar"
docker compose -p "$COMPOSE_PROJECT" -f compose.local.yaml down
docker compose -p "$COMPOSE_PROJECT" -f compose.local.yaml up -d --build --wait
COMPOSE_UP=true
curl -fsS http://127.0.0.1:8080/health
docker compose -p "$COMPOSE_PROJECT" -f compose.local.yaml --profile load run --rm locust
docker compose -p "$COMPOSE_PROJECT" -f compose.local.yaml down
COMPOSE_UP=false

step "Preparar kind"
if kind get clusters | grep -Fxq "$CLUSTER"; then
  if [[ "$FRESH" == true ]]; then
    echo "Borrando $CLUSTER y sus datos locales de prueba"
    kind delete cluster --name "$CLUSTER"
  else
    echo "Reutilizando $CLUSTER; usa --fresh para recrearlo desde cero"
  fi
fi
if ! kind get clusters | grep -Fxq "$CLUSTER"; then
  kind create cluster --name "$CLUSTER" --wait 120s
fi
kubectl --context "$CONTEXT" get --raw=/readyz

step "Cargar las siete imagenes de Compose y construir el operator"
docker build -f operator/Dockerfile -t mumnk8s/operator:local operator
bash scripts/load-kind-images.sh "$CLUSTER"

step "Aplicar servicios y recursos locales"
kubectl --context "$CONTEXT" apply -f k8s/00-namespace.yaml -f k8s/07-crds.yaml
bash scripts/apply-local-secrets.sh "$CONTEXT"
sed 's#zot.example/mumnk8s/rabbitmq:3.13-management#rabbitmq:3.13-management#' k8s/01-rabbitmq.yaml | kubectl --context "$CONTEXT" apply -f -
kubectl --context "$CONTEXT" apply -f k8s/02-services.yaml -f k8s/05-consumer-valkey.yaml -f k8s/09-operator.yaml
sed 's#zot.example/mumnk8s/prometheus:v2.54.1#prom/prometheus:v2.54.1#' k8s/10-metrics.yaml | kubectl --context "$CONTEXT" apply -f -
kubectl --context "$CONTEXT" apply -f k8s/local/custom-resources.yaml

helm --kube-context "$CONTEXT" upgrade --install eg oci://docker.io/envoyproxy/gateway-helm --version v1.9.2 -n envoy-gateway-system --create-namespace
kubectl --context "$CONTEXT" apply -f k8s/local/gateway-class.yaml
sed 's/gke-l7-global-external-managed/envoy-local/' k8s/04-gateway.yaml | kubectl --context "$CONTEXT" apply -f -
kubectl --context "$CONTEXT" apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.8.0/components.yaml
if ! kubectl --context "$CONTEXT" get deployment metrics-server -n kube-system -o jsonpath='{.spec.template.spec.containers[0].args}' | grep -q -- '--kubelet-insecure-tls'; then
  kubectl --context "$CONTEXT" patch deployment metrics-server -n kube-system --type=json -p '[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'
fi
kubectl --context "$CONTEXT" apply -f k8s/03-hpa.yaml

if [[ "$FRESH" == false ]]; then
  # Las etiquetas :dev no cambian; reiniciar para tomar las imagenes recien cargadas.
  kubectl --context "$CONTEXT" rollout restart deployment/api-rust deployment/consumer deployment/go-ingest deployment/go-writer deployment/metrics deployment/mumnk8s-operator -n mumnk8s
fi

step "Esperar pods, Metrics Server y Service de Envoy"
kubectl --context "$CONTEXT" rollout status deployment/metrics-server -n kube-system --timeout=180s
for deployment in api-rust consumer go-ingest go-writer metrics mumnk8s-operator prometheus grafana-dashboard; do
  kubectl --context "$CONTEXT" rollout status "deployment/$deployment" -n mumnk8s --timeout=300s
done
kubectl --context "$CONTEXT" rollout status statefulset/rabbitmq statefulset/valkey-reports -n mumnk8s --timeout=300s

GW_SVC=""
for _ in $(seq 1 60); do
  GW_SVC=$(kubectl --context "$CONTEXT" get svc -n envoy-gateway-system -o name | grep '^service/envoy-mumnk8s-mumnk8s-gateway-' | head -1 || true)
  [[ -n "$GW_SVC" ]] && break
  sleep 3
done
[[ -n "$GW_SVC" ]] || { echo "Envoy no creo el Service del Gateway en 180 segundos" >&2; exit 1; }
echo "Gateway: $GW_SVC"

step "Prueba de carga por Gateway; redireccion automatica"
gateway_port=$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()')
kubectl --context "$CONTEXT" port-forward -n envoy-gateway-system "$GW_SVC" "$gateway_port:80" >/tmp/mumnk8s-demo-gateway.log 2>&1 &
PF_PID=$!
for _ in $(seq 1 20); do
  if curl -sS -o /dev/null "http://127.0.0.1:$gateway_port/" 2>/dev/null; then break; fi
  kill -0 "$PF_PID" 2>/dev/null || { cat /tmp/mumnk8s-demo-gateway.log >&2; exit 1; }
  sleep 1
done
payload=$(printf '{"country":"CHN","warplanes_in_air":7,"warships_in_water":3,"timestamp":"%s"}' "$(date -u +%Y-%m-%dT%H:%M:%SZ)")
status=$(curl -sS -o /dev/null -w '%{http_code}' -X POST "http://127.0.0.1:$gateway_port/grpc-202303204" -H 'Content-Type: application/json' -d "$payload")
[[ "$status" == 202 ]] || { echo "Gateway respondio HTTP $status, se esperaba 202" >&2; exit 1; }
docker run --rm --network host -v "$ROOT/locust:/mnt:ro" locustio/locust:2.31.8 -f /mnt/locustfile.py --headless -u 100 -r 10 -t 90s --host "http://127.0.0.1:$gateway_port" --only-summary
kill "$PF_PID" 2>/dev/null || true
wait "$PF_PID" 2>/dev/null || true
PF_PID=""

step "Confirmar HPA, cola y datos de CHN"
kubectl --context "$CONTEXT" get hpa -n mumnk8s
replicas=$(kubectl --context "$CONTEXT" get deployment api-rust -n mumnk8s -o jsonpath='{.status.availableReplicas}')
[[ "${replicas:-0}" -ge 2 ]] || { echo "HPA no escalo: $replicas replicas disponibles" >&2; exit 1; }
for _ in $(seq 1 30); do
  queue=$(kubectl --context "$CONTEXT" exec -n mumnk8s statefulset/rabbitmq -- rabbitmqctl list_queues name messages_ready messages_unacknowledged)
  read -r ready unacked < <(awk '$1=="military-reports" {print $2, $3}' <<< "$queue")
  echo "RabbitMQ: ready=${ready:-?}, unacknowledged=${unacked:-?}"
  [[ "${ready:-1}" == 0 && "${unacked:-1}" == 0 ]] && break
  sleep 10
done
[[ "${ready:-1}" == 0 && "${unacked:-1}" == 0 ]] || { echo "La cola no se vacio" >&2; exit 1; }
chn=$(kubectl --context "$CONTEXT" exec -n mumnk8s statefulset/valkey-reports -- valkey-cli llen reports:CHN)
echo "Valkey reports:CHN = $chn"
[[ "$chn" =~ ^[0-9]+$ && "$chn" -gt 0 ]] || { echo "No hay reportes CHN" >&2; exit 1; }

kubectl --context "$CONTEXT" get valkeyinstances,grafanainstances,pvc -n mumnk8s
open_grafana full
