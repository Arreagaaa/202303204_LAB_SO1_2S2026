#!/usr/bin/env bash
set -euo pipefail

MODULE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_PATH="$MODULE_DIR/continfo.ko"

case "${1:-load}" in
    load)
        make -C "$MODULE_DIR"
        sudo insmod "$MODULE_PATH"
        cat /proc/continfo_pr2_so1_202303204 | head -n 8
        ;;
    unload)
        sudo rmmod continfo
        ;;
    reload)
        "$0" unload || true
        "$0" load
        ;;
    *)
        echo "Uso: $0 {load|unload|reload}" >&2
        exit 2
        ;;
esac
