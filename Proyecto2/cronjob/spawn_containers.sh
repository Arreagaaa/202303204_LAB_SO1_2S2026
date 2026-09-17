#!/usr/bin/env bash
set -euo pipefail

PREFIX="so1-pr2"
IMAGES=("so1-pr2/go-client:latest" "so1-pr2/alpine-highcpu:latest" "so1-pr2/alpine-lowconsumo:latest")
CLASSES=("high" "high" "low")
TS=$(date +%s)

for i in 0 1 2 3 4; do
    idx=$((RANDOM % ${#IMAGES[@]}))
    image="${IMAGES[$idx]}"
    class="${CLASSES[$idx]}"
    name="${PREFIX}-${i}-${TS}"
    docker run -d --name "$name" --label so1.pr2.class="$class" "$image" >/dev/null 2>&1
done

intruder="${PREFIX}-intruder-${TS}"
docker run -d --name "$intruder" --label so1.pr2.class=intruder so1-pr2/intruder:latest >/dev/null 2>&1
