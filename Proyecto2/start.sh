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
echo "[5/7] Cargando modulo de kernel y sonda eBPF..."
sudo bash kernel-module/load.sh

echo ""
echo "[6/7] Verificando..."
if [ -f /proc/continfo_pr2_so1_202303204 ] && sudo bpftool link list 2>/dev/null | grep -q "tracepoint sys_enter_kill"; then
    echo "  OK: modulo cargado y sonda adjuntada a sys_enter_kill."
else
    echo "  ERROR: modulo o sonda no listos."
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
