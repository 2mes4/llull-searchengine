# Linux Installation Guide

Complete guide for deploying Llull on a Linux VPS or bare-metal server.

## Prerequisites

- Linux server (Ubuntu 22.04+, Debian 12+, or similar)
- At least 1 GB RAM (2 GB recommended for 100K+ documents)
- Open port for HTTP (8080 by default)

## Option A: Binary Installation

### 1. Build the binary

On a machine with Go 1.24+ installed:

```bash
git clone git@github.com:2mes4/llull-searchengine.git
cd llull-searchengine
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o llull ./cmd/server
```

### 2. Install on the server

```bash
sudo useradd -r -s /bin/false llull
sudo mkdir -p /etc/llull /var/lib/llull
sudo cp llull /usr/local/bin/llull
sudo chown llull:llull /usr/local/bin/llull /var/lib/llull
```

### 3. Generate seed data (optional)

```bash
llull -generate-seed /var/lib/llull/seed.json \
      -seed-dir /path/to/llull/data/llibres-llull \
      -seed-count 1000
```

### 4. Create systemd service

```ini
# /etc/systemd/system/llull.service
[Unit]
Description=Llull Search Engine
After=network.target

[Service]
Type=simple
User=llull
Group=llull
ExecStart=/usr/local/bin/llull \
    -port 8080 \
    -auth-token ${AUTH_TOKEN} \
    -workers 4 \
    -seed-file /var/lib/llull/seed.json
Restart=always
RestartSec=5
LimitNOFILE=65536

EnvironmentFile=/etc/llull/env

[Install]
WantedBy=multi-user.target
```

Create the environment file:

```bash
# /etc/llull/env
AUTH_TOKEN=your-secret-token-here
```

### 5. Enable and start

```bash
sudo systemctl daemon-reload
sudo systemctl enable llull
sudo systemctl start llull
sudo systemctl status llull
```

### 6. Verify

```bash
curl http://localhost:8080/v1/health
```

## Option B: Docker Installation

```bash
git clone git@github.com:2mes4/llull-searchengine.git
cd llull-searchengine
docker compose -f deploy/docker-compose.yml up -d
```

Customize for production:

```yaml
# docker-compose.prod.yml
services:
  engine:
    environment:
      - AUTH_TOKEN=${AUTH_TOKEN}
    restart: always
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: "2"
```

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## TLS with Caddy

Install [Caddy](https://caddyserver.com/) and create a Caddyfile:

```Caddyfile
search.example.com {
    reverse_proxy localhost:8080
}
```

```bash
sudo apt install caddy
sudo systemctl restart caddy
```

Caddy automatically provisions and renews Let's Encrypt TLS certificates.

## TLS with nginx

```nginx
server {
    listen 443 ssl http2;
    server_name search.example.com;

    ssl_certificate /etc/letsencrypt/live/search.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/search.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Firewall

```bash
# Allow HTTP/HTTPS for TLS proxy
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Block direct engine access (only if behind proxy)
# Do NOT run this if you access :8080 directly
sudo ufw deny 8080/tcp

sudo ufw enable
```

## Monitoring

### Health check

```bash
# Simple check
curl -sf http://localhost:8080/v1/health || echo "ENGINE DOWN"

# Cron-based alert (every minute)
* * * * * curl -sf http://localhost:8080/v1/health > /dev/null || echo "Llull is down" | mail -s "Alert" you@example.com
```

### Logs

```bash
# systemd
journalctl -u llull -f

# Docker
docker compose logs -f engine
```

### Performance tuning

| Parameter | CLI Flag | Default | Recommendation |
|-----------|----------|---------|----------------|
| Workers | `-workers` | 4 | Match CPU cores |
| Queue size | `-buffer` | 5000 | Increase for bursty writes |
| Memory | — | ~1 GB per 100K docs | Monitor with `docker stats` |
