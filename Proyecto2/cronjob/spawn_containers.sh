#!/usr/bin/env bash
set -euo pipefail

PREFIX="so1-pr2"
HIGH_IMAGES=("so1-pr2/go-client:latest" "so1-pr2/alpine-highcpu:latest")
TS=$(date +%s)

PROFILES=(
    "so1-pr2/alpine-lowconsumo:latest|low"
    "so1-pr2/alpine-lowconsumo:latest|low"
    "so1-pr2/alpine-lowconsumo:latest|low"
    "${HIGH_IMAGES[$((RANDOM % ${#HIGH_IMAGES[@]}))]}|high"
    "${HIGH_IMAGES[$((RANDOM % ${#HIGH_IMAGES[@]}))]}|high"
)

for ((index = ${#PROFILES[@]} - 1; index > 0; index--)); do
    swap=$((RANDOM % (index + 1)))
    value="${PROFILES[$index]}"
    PROFILES[$index]="${PROFILES[$swap]}"
    PROFILES[$swap]="$value"
done

for index in "${!PROFILES[@]}"; do
    IFS='|' read -r image class <<< "${PROFILES[$index]}"
    name="${PREFIX}-${index}-${TS}"
    docker run -d --name "$name" --label so1.pr2.class="$class" "$image" >/dev/null 2>&1
done

intruder="${PREFIX}-intruder-${TS}"
docker run -d --name "$intruder" --label so1.pr2.class=intruder so1-pr2/intruder:latest >/dev/null 2>&1
