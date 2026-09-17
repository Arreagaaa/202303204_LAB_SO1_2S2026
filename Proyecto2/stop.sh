#!/usr/bin/env bash
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$DIR"

echo "========================================="
echo "  Proyecto 2 - Limpieza completa"
echo "========================================="

echo ""
echo "[1/5] Deteniendo daemon si esta corriendo..."
pkill -f "daemon/daemon" 2>/dev/null || true
sleep 1

echo ""
echo "[2/5] Eliminando cronjob del daemon..."
crontab -l 2>/dev/null | grep -v "proyecto2-so1" | crontab - 2>/dev/null || true

echo ""
echo "[3/5] Eliminando contenedores de prueba..."
docker ps -a --filter "label=so1.pr2.class" -q | xargs -r docker stop 2>/dev/null || true
docker ps -a --filter "label=so1.pr2.class" -q | xargs -r docker rm 2>/dev/null || true

echo ""
echo "[4/5] Descargando modulo y sonda..."
sudo rmmod continfo 2>/dev/null || echo "  Modulo ya descargado o sin permisos."
sudo bpftool prog detach pinned /sys/fs/bpf/so1_pr2_kill_tracer 2>/dev/null || true
sudo rm -f /sys/fs/bpf/so1_pr2_kill_tracer 2>/dev/null || true

echo ""
echo "[5/5] Bajando infraestructura..."
docker compose -f infra/docker-compose.yml down 2>/dev/null || true

echo ""
echo "========================================="
echo "  Limpieza completada."
echo "========================================="
