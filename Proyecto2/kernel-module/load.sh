#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
PROC_FILE="/proc/continfo_pr2_so1_202303204"
BPF_PIN="/sys/fs/bpf/so1_pr2_kill_tracer"

if [ ! -f "$PROC_FILE" ]; then
    insmod "$DIR/continfo.ko"
    echo "modulo cargado"
else
    echo "modulo ya estaba cargado"
fi

if [ -e "$BPF_PIN" ] && bpftool link list 2>/dev/null | grep -q "tracepoint sys_enter_kill"; then
    echo "sonda eBPF ya estaba cargada y adjuntada"
    exit 0
fi

rm -f "$BPF_PIN"
bpftool prog load "$DIR/../ebpf-probe/kill_tracer.bpf.o" "$BPF_PIN" autoattach
echo "sonda eBPF cargada y adjuntada"
