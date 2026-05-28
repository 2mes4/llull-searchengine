#!/usr/bin/env bash
set -euo pipefail

LLULL_BIN="${LLULL_BIN:-./llull}"
AUTH_TOKEN="${AUTH_TOKEN:-change-me}"

echo "Installing Llull on Linux VPS..."

sudo useradd -r -s /bin/false llull 2>/dev/null || true
sudo mkdir -p /etc/llull /var/lib/llull
sudo cp "$LLULL_BIN" /usr/local/bin/llull
sudo chown llull:llull /usr/local/bin/llull /var/lib/llull

echo "AUTH_TOKEN=$AUTH_TOKEN" | sudo tee /etc/llull/env

sudo tee /etc/systemd/system/llull.service > /dev/null <<'SERVICE'
[Unit]
Description=Llull Search Engine
After=network.target

[Service]
Type=simple
User=llull
Group=llull
ExecStart=/usr/local/bin/llull -port 8080 -auth-token ${AUTH_TOKEN} -workers 4
Restart=always
RestartSec=5
LimitNOFILE=65536
EnvironmentFile=/etc/llull/env

[Install]
WantedBy=multi-user.target
SERVICE

sudo systemctl daemon-reload
sudo systemctl enable llull
sudo systemctl start llull
sudo systemctl status llull --no-pager

echo "Llull installed. API at http://localhost:8080"
