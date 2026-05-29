---
name: llull-searchengine-deployment
description: >-
  Guide for deploying and integrating the Llull search engine. Use this skill
  when the user wants to deploy Llull to Docker, Kubernetes, or a VPS;
  integrate with existing databases; call the search API from JavaScript,
  TypeScript, or Python; configure data sources; or generate search UI
  components. Also triggers on "llull deploy", "llull install", "llull
  integration", "search engine setup", "llull api", "llull search".
---

# Llull Search Engine Deployment & Integration

This skill guides you through deploying the [Llull](https://github.com/2mes4/llull) search engine and integrating it with your application.

## Quick Start

```bash
git clone https://github.com/2mes4/llull.git
cd llull
docker compose up --build
# UI at http://localhost:4320, API at http://localhost:8180
```

## Configuration Parameters

### Server Flags

| Flag | Env Variable | Default | Description |
|------|-------------|---------|-------------|
| `-port` | `PORT` | `8080` | HTTP server port |
| `-auth-token` | `AUTH_TOKEN` | `llull-dev-token` | Bearer token for `/v1/index` |
| `-workers` | — | `4` | Number of worker goroutines |
| `-buffer` | — | `5000` | Worker queue buffer capacity |
| `-seed-file` | `SEED_FILE` | — | Path to JSON seed file |
| `-seed-dir` | `SEED_DIR` | — | Source text directory for seed generation |
| `-seed-count` | — | `1000` | Max documents to generate from seed |

### Data Source Parameters

Each data source connector accepts a `datasource.Config`:

```json
{
  "type": "postgres|mysql|mongodb",
  "connection": "connection-string-uri",
  "collection": "table-or-collection-name",
  "fields": ["title", "content", "tags"],
  "weight_field": "popularity_score",
  "poll_interval": "10s",
  "batch_size": 1000,
  "options": {
    "database": "myapp"
  }
}
```

## API Integration Snippets

### JavaScript / TypeScript (Browser)

```js
class LlullSearch {
  constructor(host) { this.host = host; }

  async search(query, options = {}) {
    const params = new URLSearchParams({
      q: query, page: options.page || 1,
      hits_per_page: options.hitsPerPage || 10,
      fuzzy: options.fuzzy !== false ? 'true' : 'false'
    });
    if (options.useWeight) {
      params.set('use_weight', 'true');
      params.set('weight_impact', String(options.weightImpact || 0.3));
    }
    const res = await fetch(`${this.host}/v1/search?${params}`);
    return res.json();
  }

  async indexDocument(id, fields, token) {
    const res = await fetch(`${this.host}/v1/index`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ id, action: 'INDEX', fields })
    });
    return res.status === 202;
  }
}

// Usage
const search = new LlullSearch('http://localhost:8080');
const results = await search.search('ciencia', { page: 1, fuzzy: true });
console.log(results.hits);
```

### Python

```python
import httpx
from typing import List, Dict, Optional

class LlullSearchEngine:
    def __init__(self, host: str, auth_token: Optional[str] = None):
        self.host = host.rstrip('/')
        self.auth_token = auth_token
        self._client = httpx.Client()

    def search(
        self,
        query: str,
        page: int = 1,
        hits_per_page: int = 10,
        fuzzy: bool = True,
        use_weight: bool = False,
        weight_impact: float = 0.3
    ) -> Dict:
        params = {
            'q': query,
            'page': str(page),
            'hits_per_page': str(hits_per_page),
            'fuzzy': 'true' if fuzzy else 'false',
        }
        if use_weight:
            params['use_weight'] = 'true'
            params['weight_impact'] = str(weight_impact)
        return self._client.get(f'{self.host}/v1/search', params=params).json()

    def index(self, doc_id: str, fields: Dict) -> bool:
        headers = {'Content-Type': 'application/json'}
        if self.auth_token:
            headers['Authorization'] = f'Bearer {self.auth_token}'
        res = self._client.post(
            f'{self.host}/v1/index',
            json={'id': doc_id, 'action': 'INDEX', 'fields': fields},
            headers=headers
        )
        return res.status_code == 202

    def delete(self, doc_id: str) -> bool:
        headers = {'Content-Type': 'application/json'}
        if self.auth_token:
            headers['Authorization'] = f'Bearer {self.auth_token}'
        res = self._client.post(
            f'{self.host}/v1/index',
            json={'id': doc_id, 'action': 'DELETE'},
            headers=headers
        )
        return res.status_code == 202

# Usage
engine = LlullSearchEngine('http://localhost:8080', auth_token='my-token')
results = engine.search('Ramon Llull', fuzzy=True)
print(f"Found {results['total_hits']} results")

# Index a document
engine.index('doc-001', {
    'title': 'Arbre de la Ciencia',
    'content': 'Set arbres de la ciència...',
    'weight': 0.9
})
```

### curl

```bash
# Search
curl "http://localhost:8080/v1/search?q=llull&page=1&fuzzy=true"

# Index
curl -X POST http://localhost:8080/v1/index \
  -H "Authorization: Bearer my-token" \
  -H "Content-Type: application/json" \
  -d '{"id":"doc-1","action":"INDEX","fields":{"title":"Example","content":"Hello world","weight":0.5}}'

# Health
curl http://localhost:8080/v1/health
```

### Real-time autocomplete (JavaScript)

```js
// Debounced autocomplete with <5ms latency
const input = document.getElementById('search');
let timer;

input.addEventListener('input', () => {
  clearTimeout(timer);
  const q = input.value.trim();
  if (q.length < 2) return;
  timer = setTimeout(async () => {
    const res = await fetch(`/v1/search?q=${encodeURIComponent(q)}&hits_per_page=5&fuzzy=true`);
    const data = await res.json();
    renderDropdown(data.hits);
  }, 150);
});
```

## Deployment Options

### Docker

```bash
# Production
docker run -d --name llull \
  -p 8080:8080 \
  -e AUTH_TOKEN=my-secret-token \
  ericmora/llull-searchengine \
  -port 8080 -workers 4

# With compose
docker compose up -d
```

### Kubernetes

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/
kubectl port-forward -n llull svc/llull 8080:8080
```

### VPS (Linux systemd)

```bash
# Install binary
sudo cp llull /usr/local/bin/
sudo useradd -r -s /bin/false llull

# Create service (see deploy/docs/linux-install.md)
sudo systemctl enable llull
sudo systemctl start llull

# TLS with Caddy
# search.example.com { reverse_proxy localhost:8080 }
```

## Data Source Configuration

Each data source README includes connection setup. Common pattern:

```bash
# PostgreSQL
llull -datasource '{"type":"postgres","connection":"postgres://user:pass@host/db","collection":"documents","fields":["title","content"],"weight_field":"score"}'

# MySQL
llull -datasource '{"type":"mysql","connection":"user:pass@tcp(host:3306)/db","collection":"documents","fields":["title","content"]}'

# MongoDB
llull -datasource '{"type":"mongodb","connection":"mongodb://host:27017","collection":"documents","options":{"database":"myapp"}}'
```

## UI Components

Two library packages are available:

- **React** (`npm install llull-search-components`)
- **Flutter** (`flutter pub add llull_search_components`)

See `ui-components/react/` and `ui-components/flutter/` for full docs.

## Deployment Scripts

The `skills/llull-searchengine-deployment/scripts/` directory contains:

- `deploy-docker.sh` — Build and push Docker image
- `deploy-k8s.sh` — Apply Kubernetes manifests
- `deploy-vps.sh` — Install via systemd on a Linux VPS

```bash
chmod +x skills/llull-searchengine-deployment/scripts/*.sh
./skills/llull-searchengine-deployment/scripts/deploy-docker.sh
```
