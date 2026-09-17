#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$DIR"

echo "========================================="
echo "  Proyecto 2 - Inicio completo"
echo "========================================="

echo ""
echo "[1/7] Levantando infraestructura (Valkey + Grafana)..."
docker compose -f infra/docker-compose.yml up -d
sleep 3

echo ""
echo "[2/7] Compilando modulo de kernel..."
make -C kernel-module

echo ""
echo "[3/7] Compilando sonda eBPF..."
make -C ebpf-probe

echo ""
echo "[4/7] Compilando daemon Go..."
go build -C daemon -o daemon .

echo ""
echo "[5/7] Verificando modulo de kernel..."
if [ -f /proc/continfo_pr2_so1_202303204 ]; then
    echo "  OK: modulo ya cargado."
else
    echo "  FALTA: carga el modulo en una terminal NORMAL (sudo):"
    echo "  cd $DIR && sudo insmod kernel-module/continfo.ko"
    exit 1
fi

echo ""
echo "[6/7] Verificando sonda eBPF..."
if sudo bpftool prog show pinned /sys/fs/bpf/so1_pr2_kill_tracer >/dev/null 2>&1; then
    echo "  OK: sonda ya cargada."
else
    echo "  FALTA: carga la sonda en una terminal NORMAL (sudo):"
    echo "  cd $DIR && sudo bpftool prog load ebpf-probe/kill_tracer.bpf.o /sys/fs/bpf/so1_pr2_kill_tracer"
    exit 1
fi

echo ""
echo "[7/7] Construyendo imagenes Docker..."
docker build -q -t so1-pr2/go-client:latest docker-images/go-client
docker build -q -t so1-pr2/alpine-highcpu:latest docker-images/alpine-highcpu
docker build -q -t so1-pr2/alpine-lowconsumo:latest docker-images/alpine-lowconsumo
docker build -q -t so1-pr2/intruder:latest docker-images/intruder

echo ""
echo "========================================="
echo "  Todo listo. Para iniciar el daemon:"
echo "  cd $DIR && sudo ./daemon/daemon -interval 30s"
echo ""
echo "  Grafana: http://localhost:3001"
echo "  Credenciales: admin / x"
echo "========================================="
