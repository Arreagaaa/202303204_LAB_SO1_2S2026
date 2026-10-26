#!/usr/bin/env bash
# Cargar con: source scripts/local-secrets.sh
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  echo "Carga este archivo con: source scripts/local-secrets.sh" >&2
  exit 2
fi

LOCAL_SECRET_DIR="${HOME}/.config/mumnk8s-local"
umask 077
mkdir -p "$LOCAL_SECRET_DIR"
chmod 700 "$LOCAL_SECRET_DIR"

for name in rabbitmq-password grafana-password; do
  if [[ ! -s "$LOCAL_SECRET_DIR/$name" ]]; then
    command -v openssl >/dev/null || { echo "Falta openssl" >&2; return 1; }
    openssl rand -hex 24 > "$LOCAL_SECRET_DIR/$name"
  fi
done

printf 'mumnk8s' > "$LOCAL_SECRET_DIR/rabbitmq-username"
RABBITMQ_PASSWORD="$(<"$LOCAL_SECRET_DIR/rabbitmq-password")"
GRAFANA_PASSWORD="$(<"$LOCAL_SECRET_DIR/grafana-password")"
printf 'amqp://mumnk8s:%s@rabbitmq:5672/' "$RABBITMQ_PASSWORD" > "$LOCAL_SECRET_DIR/rabbitmq-uri"
printf 'machine 127.0.0.1 login admin password %s\n' "$GRAFANA_PASSWORD" > "$LOCAL_SECRET_DIR/grafana.netrc"
chmod 600 "$LOCAL_SECRET_DIR"/*
export RABBITMQ_PASSWORD GRAFANA_PASSWORD LOCAL_SECRET_DIR
