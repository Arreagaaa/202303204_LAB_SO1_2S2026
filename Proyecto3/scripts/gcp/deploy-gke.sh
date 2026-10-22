#!/usr/bin/env bash
set -Eeuo pipefail

registry=${1:?Usage: deploy-gke.sh REGISTRY_HOST GKE_ZONE CLUSTER_NAME}
zone=${2:?Usage: deploy-gke.sh REGISTRY_HOST GKE_ZONE CLUSTER_NAME}
cluster=${3:?Usage: deploy-gke.sh REGISTRY_HOST GKE_ZONE CLUSTER_NAME}
project=${GCP_PROJECT_ID:-ID_DEL_PROYECTO}
namespace=mumnk8s
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$root"

secret_dir="$HOME/.config/mumnk8s"
password_file="$secret_dir/registry-password"
[[ -s "$password_file" ]] || { echo "Missing $password_file" >&2; exit 1; }
mkdir -p "$secret_dir"
chmod 700 "$secret_dir"

gcloud container clusters get-credentials "$cluster" --zone "$zone" --project "$project"
context=$(kubectl config current-context)
[[ "$context" == "gke_${project}_${zone}_${cluster}" ]] || {
  echo "Unexpected kubectl context: $context" >&2
  exit 1
}

oras login "$registry" --username deploy --password-stdin < "$password_file" >/dev/null
input_dir=$(mktemp -d)
trap 'rm -rf "$input_dir"' EXIT
oras pull "$registry/mumnk8s/input:dev" -o "$input_dir"
kubectl apply -f "$input_dir/k8s/00-namespace.yaml"

python3 - "$registry" "$secret_dir" <<'PY'
import base64
import json
import pathlib
import secrets
import sys

registry, directory = sys.argv[1:]
directory = pathlib.Path(directory)
password = (directory / 'registry-password').read_text().strip()
auth = base64.b64encode(f'deploy:{password}'.encode()).decode()
config = {'auths': {registry: {'username': 'deploy', 'password': password, 'auth': auth}}}
docker_config = directory / 'registry-pull.json'
docker_config.write_text(json.dumps(config))
docker_config.chmod(0o600)

for name in ('rabbitmq-password', 'grafana-password'):
    path = directory / name
    if not path.exists():
        path.write_text(secrets.token_hex(24))
        path.chmod(0o600)

rabbit_password = (directory / 'rabbitmq-password').read_text().strip()
for name, value in {
    'rabbitmq-username': 'mumnk8s',
    'rabbitmq-uri': f'amqp://mumnk8s:{rabbit_password}@rabbitmq:5672/',
}.items():
    path = directory / name
    path.write_text(value)
    path.chmod(0o600)
PY

kubectl -n "$namespace" create secret generic registry-pull \
  --type=kubernetes.io/dockerconfigjson \
  --from-file=.dockerconfigjson="$secret_dir/registry-pull.json" \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl -n "$namespace" patch serviceaccount default --type=merge \
  -p '{"imagePullSecrets":[{"name":"registry-pull"}]}'

kubectl -n "$namespace" create secret generic rabbitmq-credentials \
  --from-file=username="$secret_dir/rabbitmq-username" \
  --from-file=password="$secret_dir/rabbitmq-password" \
  --from-file=uri="$secret_dir/rabbitmq-uri" \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl -n "$namespace" create secret generic grafana-admin \
  --from-file=admin-password="$secret_dir/grafana-password" \
  --dry-run=client -o yaml | kubectl apply -f -

apply_registry_manifest() {
  sed "s#zot.example#$registry#g" "$1" | kubectl apply -f -
}

apply_registry_manifest k8s/01-rabbitmq.yaml
apply_registry_manifest k8s/02-services.yaml
kubectl apply -f k8s/03-hpa.yaml
kubectl apply -f k8s/04-gateway.yaml
apply_registry_manifest k8s/05-consumer-valkey.yaml
apply_registry_manifest k8s/10-metrics.yaml
kubectl apply -f k8s/07-crds.yaml
apply_registry_manifest k8s/09-operator.yaml
kubectl -n "$namespace" patch serviceaccount mumnk8s-operator --type=merge \
  -p '{"imagePullSecrets":[{"name":"registry-pull"}]}'
kubectl -n "$namespace" rollout restart deployment/mumnk8s-operator
kubectl rollout status deployment/mumnk8s-operator -n "$namespace" --timeout=300s
kubectl apply -f k8s/06-custom-resources.yaml
kubectl rollout status statefulset/valkey-reports -n "$namespace" --timeout=600s
kubectl rollout status deployment/grafana-dashboard -n "$namespace" --timeout=600s
kubectl -n "$namespace" wait --for=condition=Ready pod --all --timeout=600s
kubectl -n "$namespace" get pods,deployments,statefulsets,hpa,pvc,gateway,httproute
