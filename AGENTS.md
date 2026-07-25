# AGENTS.md

## Project: Llull Search Engine

**Capa:** Execució
**Funció:** Cercador semàntic multi-índex per a documents i continguts del sistema.
**No fa:** Xat amb usuaris, execució d'agents, emmagatzematge de dades primàries.

### Responsabilitats

1. **Indexació de documents**: Rep documents via API i els indexa en un trie amb persistència BoltDB.
   Cada espai de treball té el seu propi índex (`space-{spaceId}`).
2. **Cerca semàntica**: Permet cerques per text complet amb fuzzy matching, ranking i paginació.
3. **Multi-índex**: Gestiona múltiples índexs independents amb TTL de descàrrega automàtica.
4. **Connectors de dades**: Pot indexar des de PostgreSQL, MySQL, MongoDB i Firestore.

### Multi-Tenant

Els índexs es prefiquen automàticament amb el tenant.

```bash
export LLULL_TENANT_PREFIX=makeyourcrew
go run ./cmd/server
```

`/v1/space-abc123/search` → internament cerca `makeyourcrew-space-abc123`

### Sereno Logs

Events a stdout en format JSON sereno: `search`, `document_indexed`, `document_deleted`,
amb `latency_ms`, `query`, `total_hits`, `namespace`.

### Build & Run

```bash
go build ./...
go test ./... -v -race
go run ./cmd/server -port 8080
```

### API

```bash
GET    /v1/{index}/search?q=query&limit=10
POST   /v1/{index}/index     # { id, action: "INDEX"|"DELETE", fields }
DELETE /v1/{index}/documents/{id}
GET    /v1/health
```

### Key Design Decisions

- **Trie** (no Radix Tree) per simplicitat
- **Levenshtein automaton** amb DFS pruning per fuzzy search
- **BoltDB** per persistència embedded sense dependències externes
- **Worker pool** amb buffered channel per indexació asíncrona
- Connectors implementen `datasource.Connector` interface
