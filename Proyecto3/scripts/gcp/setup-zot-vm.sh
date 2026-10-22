#!/usr/bin/env bash
set -euo pipefail

host=${1:?Usage: setup-zot-vm.sh REGISTRY_HOST}
cd "$(dirname "$0")"

sudo apt-get update
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y docker.io apache2-utils
sudo systemctl enable --now docker
sudo install -d -m 0750 /opt/mumnk8s-zot /var/lib/mumnk8s-zot
sudo install -m 0644 zot-config.json /opt/mumnk8s-zot/config.json
sed "s/REGISTRY_HOST/$host/g" Caddyfile.template | sudo tee /opt/mumnk8s-zot/Caddyfile >/dev/null

if [[ ! -s /opt/mumnk8s-zot/htpasswd ]]; then
  if [[ ! -s "$HOME/registry-password" ]]; then
    echo "Missing $HOME/registry-password" >&2
    exit 1
  fi
  sudo htpasswd -iBc /opt/mumnk8s-zot/htpasswd deploy < "$HOME/registry-password"
  shred -u "$HOME/registry-password"
fi

sudo docker pull ghcr.io/project-zot/zot-linux-amd64:v2.1.18
sudo docker pull caddy:2.10.2
sudo docker rm -f mumnk8s-zot mumnk8s-caddy 2>/dev/null || true
sudo docker run -d --name mumnk8s-zot --restart unless-stopped --network host \
  -v /opt/mumnk8s-zot/config.json:/etc/zot/config.json:ro \
  -v /opt/mumnk8s-zot/htpasswd:/etc/zot/htpasswd:ro \
  -v /var/lib/mumnk8s-zot:/var/lib/zot \
  ghcr.io/project-zot/zot-linux-amd64:v2.1.18 serve /etc/zot/config.json
sudo docker run -d --name mumnk8s-caddy --restart unless-stopped --network host \
  -v /opt/mumnk8s-zot/Caddyfile:/etc/caddy/Caddyfile:ro \
  -v mumnk8s-caddy-data:/data -v mumnk8s-caddy-config:/config \
  caddy:2.10.2
