#!/usr/bin/env bash
set -euo pipefail

PREFIX="so1-pr2"
PROFILES=(low low low high high)

for ((index = ${#PROFILES[@]} - 1; index > 0; index--)); do
    swap=$((RANDOM % (index + 1)))
    value="${PROFILES[$index]}"
    PROFILES[$index]="${PROFILES[$swap]}"
    PROFILES[$swap]="$value"
done

for index in "${!PROFILES[@]}"; do
    name="${PREFIX}-${index}-$(date +%s)"
    profile="${PROFILES[$index]}"

    if [[ "$profile" == "low" ]]; then
        docker run -d --name "$name" --label so1.pr2.class=low so1-pr2/alpine-lowconsumo:latest
    else
        docker run -d --name "$name" --label so1.pr2.class=high so1-pr2/alpine-highcpu:latest
    fi
done

intruder="${PREFIX}-intruder-$(date +%s)"
docker run -d --name "$intruder" --label so1.pr2.class=intruder so1-pr2/intruder:latest
