#!/usr/bin/env bash
set -Eeuo pipefail

registry=${1:?Usage: publish-images.sh REGISTRY_HOST}
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$root"

for command in docker oras; do
  command -v "$command" >/dev/null || { echo "Missing $command" >&2; exit 1; }
done

password_file="$HOME/.config/mumnk8s/registry-password"
[[ -s "$password_file" ]] || { echo "Missing registry password" >&2; exit 1; }
docker login "$registry" --username deploy --password-stdin < "$password_file" >/dev/null
oras login "$registry" --username deploy --password-stdin < "$password_file" >/dev/null

for name in api-rust go-ingest go-ingest-client go-writer go-writer-publisher consumer metrics; do
  echo "Building and pushing $name"
  docker build --quiet -f "$name/Dockerfile" -t "$registry/mumnk8s/$name:dev" . >/dev/null
  docker push "$registry/mumnk8s/$name:dev" >/dev/null
done

operator_context=$(mktemp -d)
trap 'rm -rf "$operator_context"' EXIT
cp -a operator/. "$operator_context/"
sed -i "s#zot.example#$registry#g" \
  "$operator_context/helm-charts/valkeyinstance/values.yaml" \
  "$operator_context/helm-charts/grafanainstance/values.yaml"
echo 'Building and pushing operator'
docker build --quiet -t "$registry/mumnk8s/operator:dev" "$operator_context" >/dev/null
docker push "$registry/mumnk8s/operator:dev" >/dev/null

for spec in \
  'rabbitmq:3.13-management|rabbitmq:3.13-management' \
  'valkey/valkey:8|valkey:8' \
  'grafana/grafana:11.2.0|grafana:11.2.0' \
  'prom/prometheus:v2.54.1|prometheus:v2.54.1' \
  'falcosecurity/falco:0.45.1|falco:0.45.1' \
  'falcosecurity/falcoctl:0.14.2|falcoctl:0.14.2'; do
  source_image=${spec%%|*}
  target_image="$registry/mumnk8s/${spec##*|}"
  echo "Mirroring $source_image"
  docker pull "$source_image" >/dev/null
  docker tag "$source_image" "$target_image"
  docker push "$target_image" >/dev/null
done

echo 'Publishing OCI input artifact'
oras push "$registry/mumnk8s/input:dev" \
  k8s/00-namespace.yaml:application/yaml >/dev/null
echo 'Mirroring Falco container plugin and its signature'
oras cp -r ghcr.io/falcosecurity/plugins/plugin/container:0.7.5 \
  "$registry/mumnk8s/falco-container-plugin:0.7.5" >/dev/null
echo 'All images and OCI artifact published'
