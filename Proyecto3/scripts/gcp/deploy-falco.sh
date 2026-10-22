#!/usr/bin/env bash
set -Eeuo pipefail

registry=${1:?Usage: deploy-falco.sh REGISTRY_HOST}
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$root"
secret_dir="$HOME/.config/mumnk8s"
[[ -s "$secret_dir/registry-password" && -s "$secret_dir/registry-pull.json" ]] || {
  echo "Missing registry credentials in $secret_dir" >&2
  exit 1
}

kubectl create namespace falco --dry-run=client -o yaml | kubectl apply -f -
kubectl -n falco create secret generic registry-pull \
  --type=kubernetes.io/dockerconfigjson \
  --from-file=.dockerconfigjson="$secret_dir/registry-pull.json" \
  --dry-run=client -o yaml | kubectl apply -f -

python3 - "$registry" "$secret_dir" <<'PY'
import pathlib
import sys

registry, directory = sys.argv[1:]
directory = pathlib.Path(directory)
password = (directory / 'registry-password').read_text().strip()
path = directory / 'falcoctl-registry-auth'
path.write_text(f'{registry},deploy,{password}')
path.chmod(0o600)
PY
kubectl -n falco create secret generic falcoctl-registry-auth \
  --from-file=FALCOCTL_REGISTRY_AUTH_BASIC="$secret_dir/falcoctl-registry-auth" \
  --dry-run=client -o yaml | kubectl apply -f -

helm repo add falcosecurity https://falcosecurity.github.io/charts --force-update >/dev/null
helm repo update falcosecurity >/dev/null
helm upgrade --install falco falcosecurity/falco \
  --version 9.2.1 --namespace falco --values k8s/falco-values.yaml \
  --set image.registry="$registry" \
  --set image.repository=mumnk8s/falco \
  --set image.tag=0.45.1 \
  --set falcoctl.image.registry="$registry" \
  --set falcoctl.image.repository=mumnk8s/falcoctl \
  --set falcoctl.image.tag=0.14.2 \
  --set driver.kind=modern_ebpf \
  --set driver.loader.enabled=false \
  --set falcoctl.artifact.follow.enabled=false \
  --set "imagePullSecrets[0].name=registry-pull" \
  --set "falcoctl.artifact.install.envFrom[0].secretRef.name=falcoctl-registry-auth" \
  --set-string collectors.containerEngine.pluginRef="$registry/mumnk8s/falco-container-plugin:0.7.5" \
  --wait --timeout=10m

kubectl get daemonsets,pods -n falco
