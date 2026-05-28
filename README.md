<p align="center">
  <img src="https://img.shields.io/badge/go-1.24+-00ADD8?style=flat-square" />
  <img src="https://img.shields.io/badge/license-source--available-orange?style=flat-square" />
  <img src="https://img.shields.io/badge/latency-%3C5ms-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/multi--index-supported-blue?style=flat-square" />
</p>

<h1 align="center">Llull Search Engine</h1>

<p align="center">
  <strong>Search engine for your database.</strong><br/>
  Multi-index, search-as-you-type with fuzzy matching and weighted ranking.<br/>
  Works alongside <strong>Firestore</strong>, <strong>PostgreSQL</strong>, <strong>MySQL</strong>, <strong>MongoDB</strong> — any data source.
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> · <a href="#architecture">Architecture</a> · <a href="#configuration">Configuration</a> · <a href="#api">API</a> · <a href="#data-sources">Data Sources</a> · <a href="#ui-components">UI Components</a> · <a href="#deployment">Deployment</a>
</p>

---

## Why Llull?

Your database is great at storing data. It's terrible at searching it.

Firestore charges per read and can't do full-text search. PostgreSQL `LIKE %query%` scans entire tables. MongoDB text indexes are limited. You could use Algolia or ElasticSearch — but that means sending data to a third party, paying per query, and depending on their uptime.

**Llull runs on your infrastructure.** An in-memory prefix trie delivers sub-millisecond searches with zero external dependencies. It syncs with your database through pluggable data sources and serves a clean search API. Supports **multiple named indices**, each with its own trie and BoltDB persistence, with automatic unloading of idle indices to conserve memory.

### What you get

- **Multi-index** — Multiple named indices, each with isolated trie and persistence. Auto-unload idle indices after configurable TTL
- **Search-as-you-type** — Prefix trie with O(K) lookup, independent of document count
- **Typo tolerance** — Levenshtein automaton with DFS pruning, max distance 2
- **Weighted ranking** — `Sf = (St × (1-I)) + (Sb × I)` with user-configurable impact
- **BoltDB persistence** — Index survives restarts. Metadata persisted to embedded key-value store
- **Sub-millisecond latency** — In-memory index, concurrent reads with `sync.RWMutex`
- **Async indexing** — Worker pool with buffered channel, non-blocking enqueue
- **Multiple data sources** — Firestore (push), PostgreSQL, MySQL, MongoDB (poll)
- **React & Flutter components** — Plug-and-play search UI libraries with index prop
- **Google-style UI** — Included vanilla JS search interface with match highlighting and modal

---

## Quick Start

### Docker (recommended)

```bash
git clone https://github.com/2mes4/llull-searchengine.git
cd llull-searchengine
docker compose -f deploy/docker-compose.yml up --build
```

Open `http://localhost:4320`. API at `http://localhost:8180`.

### Local (Go)

```bash
# Generate seed data (embedded Ramon Llull texts in Catalan)
go run ./cmd/server -generate-seed seed.json -seed-count 1000

# Start engine
go run ./cmd/server -seed-file seed.json -port 8080

# Search
curl "http://localhost:8080/v1/search?q=llull"

# Multi-index search
curl "http://localhost:8080/v1/default/search?q=meravelles"
```

---

## Architecture

```
                         ┌──────────────┐
                         │  IndexManager │
                         │  (auto-unload │
                         │   idle indices│
                         │   after TTL)  │
                         └──────┬───────┘
                                │
              ┌─────────────────┼─────────────────┐
              │                 │                   │
         ┌────┴─────┐     ┌────┴─────┐       ┌────┴─────┐
         │ "default"│     │"products"│       │  "users" │
         │ ┌───────┐│     │ ┌───────┐│       │ ┌───────┐│
         │ │ Trie  ││     │ │ Trie  ││       │ │ Trie  ││
         │ │BoltDB ││     │ │BoltDB ││       │ │BoltDB ││
         │ └───────┘│     │ └───────┘│       │ └───────┘│
         └──────────┘     └──────────┘       └──────────┘

 Database Platform ──► Data Source ──► Llull Engine ──► Search API
    (Firestore,        (connector       (IndexManager     (GET /v1/{index}/search)
     Postgres,          pushes or        manages in-
     MySQL,             polls for        memory trie      (POST /v1/{index}/index)
     MongoDB)           changes)         instances)
```

### Core components

| Component | File | Description |
|-----------|------|-------------|
| **IndexManager** | `internal/engine/index_manager.go` | Manages multiple named indices, auto-unload, lazy loading |
| **Prefix Trie** | `internal/engine/trie.go` | Each node stores `DocIDs []string`. O(K) search |
| **Fuzzy Search** | `internal/engine/fuzzy.go` | Levenshtein automaton DFS on trie, prunes branches |
| **Worker Pool** | `internal/worker/pool.go` | Buffered channel + goroutines, 503 on overflow |
| **Ranking** | `internal/engine/search.go` | `Sf = (St × (1-I)) + (Sb × I)` |
| **Persistence** | `internal/engine/persist.go` | BoltDB key-value store per index |
| **Data Sources** | `internal/datasource/` | `Connector` interface: `Connect`, `Sync`, `Close` |

---

## Configuration

### Server Parameters

| Flag | Env Variable | Default | Description |
|------|-------------|---------|-------------|
| `-port` | `PORT` | `8080` | HTTP port |
| `-auth-token` | `AUTH_TOKEN` | `llull-dev-token` | Bearer token for index endpoint |
| `-workers` | — | `4` | Worker goroutines for indexing |
| `-buffer` | — | `5000` | Index queue buffer capacity |
| `-seed-file` | `SEED_FILE` | — | JSON seed file to load on startup |
| `-seed-dir` | `SEED_DIR` | — | Source directory for seed generation |
| `-seed-count` | — | `1000` | Max seed documents to generate |
| `-db` | `DB_PATH` | `/data` | Directory for BoltDB persistent files |
| `-default-index` | `DEFAULT_INDEX` | `default` | Name of the default index |
| `-index-ttl` | — | `30m` | Time before unloading idle indices |
| `-data-source` | `DATA_SOURCE` | `seed` | Label for the active data source |

### Data Source Configuration

```json
{
  "type": "postgres",
  "connection": "postgres://user:pass@host:5432/db?sslmode=disable",
  "collection": "documents",
  "index": "products",
  "fields": ["title", "content"],
  "weight_field": "popularity_score",
  "poll_interval": "10s",
  "batch_size": 1000
}
```

Each table/collection can target a specific Llull index via the `index` field.

---

## API

### Search

```
GET /v1/search?q=ciencia&page=1&hits_per_page=10&use_weight=true&weight_impact=0.3&fuzzy=true
GET /v1/{index}/search?q=ciencia&page=1&hits_per_page=10&fuzzy=true
```

```json
{
  "hits": [{"id":"doc-01","score":0.935,"weight":0.72,"fields":{"title":"...","content":"..."}}],
  "total_hits": 35,
  "page": 1,
  "nb_pages": 4,
  "index": "default",
  "query_time": 5793
}
```

### Index

```
POST /v1/index
POST /v1/{index}/index
Authorization: Bearer <token>
{"id":"doc-id","action":"INDEX","fields":{"title":"...","content":"...","weight":0.8}}
```

### Health

```
GET /v1/health  →  {"status":"ok","docs_indexed":1000,"data_source":"seed","indices":{"default":{"docs":1000,"loaded":true}}}
GET /v1/indices →  {"indices":{"default":{"docs":1000,"loaded":true}},"default_index":"default"}
```

---

## Data Sources

Llull works in parallel with your database platform. It doesn't replace it — it indexes a copy of the fields you want searchable. Each table/collection maps to a named Llull index.

| Data Source | Sync Model | Status | Docs |
|-------------|-----------|--------|------|
| **Firestore** | Push (Cloud Function) | Production-ready | [`data-sources/firestore/`](data-sources/firestore/) |
| **PostgreSQL** | Poll (timestamp) | Connector + docs | [`data-sources/postgres/`](data-sources/postgres/) |
| **MySQL** | Poll (timestamp) | Connector + docs | [`data-sources/mysql/`](data-sources/mysql/) |
| **MongoDB** | Poll (cursor) | Connector + docs | [`data-sources/mongodb/`](data-sources/mongodb/) |

Each data source directory contains a `README.md` with connection parameters, schema requirements, and setup instructions. All connectors implement the `datasource.Connector` interface in `internal/datasource/`.

See [`skills/llull-searchengine-datasources-creator/SKILL.md`](skills/llull-searchengine-datasources-creator/SKILL.md) for the guide on building custom connectors.

---

## UI Components

Plug-and-play search UI for React and Flutter applications. All components accept an optional `index` prop for multi-index support.

| Package | Platform | Docs |
|---------|----------|------|
| `@llull/search-components` | React (npm) | [`ui-components/react/`](ui-components/react/) |
| `llull_search_components` | Flutter (pub.dev) | [`ui-components/flutter/`](ui-components/flutter/) |

### Available components

- **Dropdown** — Autocomplete search-as-you-type with debounce, match highlighting, and multi-index support
- **Headless hook / controller** — Full control over rendering, returns raw results with index awareness
- **Results view** — Full search interface with pagination, score display, and index selector

---

## Project Structure

```
cmd/server/main.go              Entry point with IndexManager + BoltDB persistence
internal/engine/                Core search engine (IndexManager, trie, search, ranking, fuzzy, persist)
internal/api/                   HTTP handlers (chi router, multi-index routes)
internal/worker/                Buffered worker pool
internal/datasource/            Data source abstraction layer + connectors (Postgres, MySQL, MongoDB)
internal/seed/                  Seed data generator from text files
data-sources/                   Data source connectors (Firestore, Postgres, MySQL, MongoDB)
skills/                         AI skills for contributors and deployment
ui-components/react/            React component library (npm)
ui-components/flutter/          Flutter widget library (pub.dev)
web/                            Search UI (static HTML/CSS/JS, Google-style with modal)
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
  -e DB_PATH=/data \
  llull-searchengine -port 8080 -workers 4
```

With compose:

```bash
docker compose -f deploy/docker-compose.yml up -d
```

### Kubernetes

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/
kubectl -n llull port-forward svc/llull-searchengine 8080:8080
```

### Linux (VPS)

Full guide in [`deploy/docs/linux-install.md`](deploy/docs/linux-install.md):

```bash
./skills/llull-searchengine-deployment/scripts/deploy-vps.sh
```

---

## Skills

| Skill | Description |
|-------|-------------|
| [`llull-searchengine-contributor`](skills/llull-searchengine-contributor/SKILL.md) | How to contribute data sources and frontend components |
| [`llull-searchengine-deployment`](skills/llull-searchengine-deployment/SKILL.md) | How to deploy and integrate with JS/Python |
| [`llull-searchengine-datasources-creator`](skills/llull-searchengine-datasources-creator/SKILL.md) | Step-by-step guide for building new connectors |

---

## License

Source-available license. See [LICENSE](LICENSE) for details.

You may use, modify, and self-host freely. You may not offer Llull as a managed cloud service.
