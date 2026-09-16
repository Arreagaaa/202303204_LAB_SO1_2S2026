#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$DIR"

echo "========================================="
echo "  Proyecto 2 - Inicio completo"
echo "========================================="

echo ""
echo "[1/8] Levantando infraestructura (Valkey + Grafana)..."
docker compose -f infra/docker-compose.yml up -d
sleep 3

echo ""
echo "[2/8] Compilando modulo de kernel..."
make -C kernel-module

echo ""
echo "[3/8] Compilando sonda eBPF..."
make -C ebpf-probe

echo ""
echo "[4/8] Compilando daemon Go..."
go build -o daemon/daemon ./daemon

echo ""
echo "[5/8] Verificando modulo de kernel..."
if [ -f /proc/continfo_pr2_so1_202303204 ]; then
    echo "  OK: modulo ya cargado."
else
    echo "  FALTA: carga el modulo en una terminal NORMAL (sudo):"
    echo "  cd $DIR && sudo insmod kernel-module/continfo.ko"
    echo "  Aborta hasta que el modulo este cargado."
    exit 1
fi

echo ""
echo "[6/8] Verificando sonda eBPF..."
if [ -f /sys/fs/bpf/so1_pr2_kill_tracer ]; then
    echo "  OK: sonda ya cargada."
else
    echo "  FALTA: carga la sonda en una terminal NORMAL (sudo):"
    echo "  cd $DIR && sudo bpftool prog load ebpf-probe/kill_tracer.bpf.o /sys/fs/bpf/so1_pr2_kill_tracer"
    echo "  Aborta hasta que la sonda este cargada."
    exit 1
fi

echo ""
echo "[7/8] Construyendo imagenes Docker..."
docker build -q -t so1-pr2/go-client:latest docker-images/go-client
docker build -q -t so1-pr2/alpine-highcpu:latest docker-images/alpine-highcpu
docker build -q -t so1-pr2/alpine-lowconsumo:latest docker-images/alpine-lowconsumo
docker build -q -t so1-pr2/intruder:latest docker-images/intruder

echo ""
echo "[8/8] Creando contenedores de prueba..."
chmod +x cronjob/spawn_containers.sh
./cronjob/spawn_containers.sh
sleep 3
echo "Contenedores creados:"
docker ps --filter "label=so1.pr2.class" --format "  {{.Names}} ({{.Label \"so1.pr2.class\"}})"

echo ""
echo "========================================="
echo "  Todo listo. Para iniciar el daemon:"
echo "  cd $DIR && sudo ./daemon/daemon -interval 10s"
echo ""
echo "  Grafana: http://localhost:3001"
echo "  Credenciales: admin / x"
echo "========================================="
