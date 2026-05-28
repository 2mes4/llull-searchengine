<p align="center">
  <img src="https://img.shields.io/badge/go-1.24+-00ADD8?style=flat-square" />
  <img src="https://img.shields.io/badge/license-source--available-orange?style=flat-square" />
  <img src="https://img.shields.io/badge/latency-%3C5ms-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/docker-llull--searchengine-blue?style=flat-square" />
</p>

<h1 align="center">Llull</h1>

<p align="center">
  <strong>Search engine for your database.</strong><br/>
  Drop-in search-as-you-type with fuzzy matching and weighted ranking.<br/>
  Works alongside <strong>Firestore</strong>, <strong>PostgreSQL</strong>, <strong>MySQL</strong>, <strong>MongoDB</strong> — any data source.
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> · <a href="#architecture">Architecture</a> · <a href="#configuration">Configuration</a> · <a href="#api">API</a> · <a href="#data-sources">Data Sources</a> · <a href="#ui-components">UI Components</a> · <a href="#deployment">Deployment</a>
</p>

---

## Why Llull?

Your database is great at storing data. It's terrible at searching it.

Firestore charges per read and can't do full-text search. PostgreSQL `LIKE %query%` scans entire tables. MongoDB text indexes are limited. You could use Algolia or ElasticSearch — but that means sending data to a third party, paying per query, and depending on their uptime.

**Llull runs on your infrastructure.** An in-memory prefix trie delivers sub-millisecond searches with zero external dependencies. It syncs with your database through pluggable data sources and serves a clean search API.

### What you get

- **Search-as-you-type** — Prefix trie with O(K) lookup, independent of document count
- **Typo tolerance** — Levenshtein automaton with DFS pruning, max distance 2
- **Weighted ranking** — `Sf = (St × (1-I)) + (Sb × I)` with user-configurable impact
- **Sub-millisecond latency** — In-memory index, concurrent reads with `sync.RWMutex`
- **Async indexing** — Worker pool with buffered channel, non-blocking enqueue
- **Multiple data sources** — Firestore (push), PostgreSQL, MySQL, MongoDB (poll)
- **React & Flutter components** — Plug-and-play search UI libraries
- **Google-style UI** — Included vanilla JS search interface with match highlighting

---

## Quick Start

### Docker (recommended)

```bash
docker run -d --name llull -p 8080:8080 llull-searchengine -port 8080 -workers 4
```

With compose:

```bash
git clone https://github.com/2mes4/llull.git
cd llull
docker compose -f deploy/docker-compose.yml up --build
```

Open `http://localhost:4320`. API at `http://localhost:8180`.

### Local (Go)

```bash
# Generate seed data
go run ./cmd/server -generate-seed seed.json -seed-dir data/llibres-llull

# Start engine
go run ./cmd/server -seed-file seed.json -port 8080

# Search
curl "http://localhost:8080/v1/search?q=llull"
```

---

## Architecture

```
 Database Platform ──► Data Source ──► Llull Engine ──► Search API
    (Firestore,        (connector       (in-memory        (GET /v1/search)
     Postgres,          pushes or        prefix trie
     MySQL,             polls for       + ranking
     MongoDB)           changes)        + fuzzy)
```

### Core components

| Component | File | Description |
|-----------|------|-------------|
| **Prefix Trie** | `internal/engine/trie.go` | Each node stores `DocIDs []string`. O(K) search |
| **Fuzzy Search** | `internal/engine/fuzzy.go` | Levenshtein automaton DFS on trie, prunes branches |
| **Worker Pool** | `internal/worker/pool.go` | Buffered channel + goroutines, 503 on overflow |
| **Ranking** | `internal/engine/search.go` | `Sf = (St × (1-I)) + (Sb × I)` |
| **Data Sources** | `internal/datasource/` | `Connector` interface: `Connect`, `Sync`, `Close` |

---

## Configuration

### Server Parameters

| Flag | Env Variable | Default | Description |
|------|-------------|---------|-------------|
| `-port` | `PORT` | `8080` | HTTP port |
| `-auth-token` | `AUTH_TOKEN` | `llull-dev-token` | Bearer token for `/v1/index` |
| `-workers` | — | `4` | Worker goroutines for indexing |
| `-buffer` | — | `5000` | Index queue buffer capacity |
| `-seed-file` | `SEED_FILE` | — | JSON file to load on startup |
| `-seed-dir` | `SEED_DIR` | — | Source directory for seed generation |
| `-seed-count` | — | `1000` | Max seed documents to generate |

### Data Source Configuration

```json
{
  "type": "postgres",
  "connection": "postgres://user:pass@host:5432/db?sslmode=disable",
  "collection": "documents",
  "fields": ["title", "content"],
  "weight_field": "popularity_score",
  "poll_interval": "10s",
  "batch_size": 1000
}
```

---

## API

### Search

```
GET /v1/search?q=ciencia&page=1&hits_per_page=10&use_weight=true&weight_impact=0.3&fuzzy=true
```

```json
{
  "hits": [{"id":"doc-01","score":0.935,"weight":0.72,"fields":{"title":"...","content":"..."}}],
  "total_hits": 35,
  "page": 1,
  "nb_pages": 4
}
```

### Index

```
POST /v1/index
Authorization: Bearer <token>
{"id":"doc-id","action":"INDEX","fields":{"title":"...","content":"...","weight":0.8}}
```

### Health

```
GET /v1/health  →  {"status":"ok","docs_indexed":414}
```

---

## Data Sources

Llull works in parallel with your database platform. It doesn't replace it — it indexes a copy of the fields you want searchable.

| Data Source | Sync Model | Status | Docs |
|-------------|-----------|--------|------|
| **Firestore** | Push (Cloud Function) | Production-ready | [`data-sources/firestore/`](data-sources/firestore/) |
| **PostgreSQL** | Poll (timestamp) | Connector + docs | [`data-sources/postgres/`](data-sources/postgres/) |
| **MySQL** | Poll (timestamp) | Connector + docs | [`data-sources/mysql/`](data-sources/mysql/) |
| **MongoDB** | Poll (cursor) | Connector + docs | [`data-sources/mongodb/`](data-sources/mongodb/) |

Each data source directory contains a `README.md` with connection parameters, schema requirements, and setup instructions. All connectors implement the `datasource.Connector` interface in `internal/datasource/`.

See [`skills/create-datasource/SKILL.md`](skills/create-datasource/SKILL.md) for the guide on building custom connectors.

---

## UI Components

Plug-and-play search UI for React and Flutter applications.

| Package | Platform | Docs |
|---------|----------|------|
| `@llull/search-components` | React (npm) | [`ui-components/react/`](ui-components/react/) |
| `llull_search_components` | Flutter (pub.dev) | [`ui-components/flutter/`](ui-components/flutter/) |

### Available components

- **Dropdown** — Autocomplete search-as-you-type with debounce and match highlighting
- **Headless hook / controller** — Full control over rendering, returns raw results
- **Results view** — Full search interface with pagination and score display

---

## Project Structure

```
cmd/server/main.go              Entry point
internal/engine/                Core search engine (trie, search, ranking, fuzzy)
internal/api/                   HTTP handlers (chi router)
internal/worker/                Buffered worker pool
internal/datasource/            Data source abstraction layer + connectors
internal/seed/                  Seed data generator from text files
data-sources/                   Data source connectors (Firestore, Postgres, MySQL, MongoDB)
skills/                         AI skills for contributors and deployment
ui-components/react/            React component library (npm)
ui-components/flutter/          Flutter widget library (pub.dev)
web/                            Search UI (static HTML/CSS/JS, Google-style)
deploy/docker/                  Dockerfiles and compose
deploy/k8s/                     Kubernetes manifests
deploy/docs/                    Linux installation guide
```

---

## Deployment

### Docker (public image)

```bash
docker run -d --name llull -p 8080:8080 \
  -e AUTH_TOKEN=my-secret \
  llull-searchengine -port 8080 -workers 4
```

Or with compose:

```bash
docker compose -f deploy/docker-compose.yml up -d
```

### Kubernetes

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/
kubectl -n llull port-forward svc/llull 8080:8080
```

### Linux (VPS)

Full guide in [`deploy/docs/linux-install.md`](deploy/docs/linux-install.md):

```bash
./skills/llull-searchengine-deployment/scripts/deploy-vps.sh
```

### Kubernetes

Full manifests in [`deploy/k8s/`](deploy/k8s/):

```bash
./skills/llull-searchengine-deployment/scripts/deploy-k8s.sh
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) and the [contributor skill](skills/llull-searchengine-contributor/SKILL.md).

### Skills

| Skill | Description |
|-------|-------------|
| [`llull-searchengine-contributor`](skills/llull-searchengine-contributor/SKILL.md) | How to contribute data sources and frontend components |
| [`llull-searchengine-deployment`](skills/llull-searchengine-deployment/SKILL.md) | How to deploy and integrate with JS/Python |
| [`create-datasource`](skills/create-datasource/SKILL.md) | Step-by-step guide for building new connectors |

---

## License

Source-available license. See [LICENSE](LICENSE) for details.

You may use, modify, and self-host freely. You may not offer Llull as a managed cloud service.
