# AGENTS.md

## Project: Llull Search Engine

Algolia-like search engine for databases. Written in Go. Designed to run as a sidecar on a VPS or in Kubernetes. Named after Ramon Llull.

## Build & Run Commands

```bash
# Build
go build ./...

# Run tests (always with -race)
go test ./... -v -race

# Generate seed data from llibres-llull text files
go run ./cmd/server -generate-seed seed.json -seed-dir data/llibres-llull -seed-count 1000

# Run server with seed data
go run ./cmd/server -seed-file seed.json -port 8080

# Docker
docker compose -f deploy/docker-compose.yml up --build
```

## Project Structure

```
cmd/server/main.go              Entry point
internal/engine/                Core search engine (trie, search, ranking, fuzzy)
internal/api/                   HTTP handlers (chi router)
internal/worker/                Buffered worker pool
internal/datasource/            Data source abstraction layer + connectors
internal/seed/                  Seed data generator from text files
data-sources/firestore/         Firebase Extension (push model)
data-sources/postgres/          PostgreSQL connector docs + config
data-sources/mysql/             MySQL connector docs + config
data-sources/mongodb/           MongoDB connector docs + config
skills/                         AI skills for contributor, deployment, datasource creation
ui-components/react/            React component library (npm)
ui-components/flutter/          Flutter widget library (pub.dev)
web/                            Search UI (static HTML/CSS/JS, Google-style)
deploy/docker/                  Dockerfiles and compose
deploy/k8s/                     Kubernetes manifests
deploy/docs/                    Linux installation guide
```

## Code Style

- No comments unless explicitly requested
- Follow existing patterns in each package
- Use `sync.RWMutex` for all concurrent access to the trie
- Write-locked for index/delete, read-locked for search
- Error types: plain `fmt.Errorf` with `%w` wrapping
- All code and documentation in English
- External dependencies: only `chi`, `chi/cors`, and `golang.org/x/text` for core
- Data source connectors may import their respective database drivers

## Testing

- Every package must have `*_test.go` files
- Always run with `-race` flag
- Use `httptest.NewServer` for API integration tests
- Use `t.Cleanup()` for resource cleanup

## Key Design Decisions

- Trie (not Radix Tree) for simplicity — can be upgraded later
- Levenshtein automaton with DFS pruning for fuzzy search
- Early truncation at 1000 results before sorting for pagination performance
- Worker pool with buffered channel for async indexing
- Elastic License 2.0 style — source available, no cloud resale
- Search results include document fields, weight, and score for UI rendering
- Data sources implement the `datasource.Connector` interface for extensibility
- All data source connectors follow the same internal structure and configuration format

## Data Source Development

See `skills/create-datasource/SKILL.md` for the guide on building new data source connectors.

Every connector must:
1. Implement `datasource.Connector` interface
2. Accept `datasource.Config` for connection parameters
3. Emit `datasource.Event` objects through the callback
4. Handle context cancellation for graceful shutdown
5. Include a README.md in its data-sources/ directory
