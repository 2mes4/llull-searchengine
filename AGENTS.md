# AGENTS.md

## Project: Llull Search Engine

**Capa:** Execució
**Rol:** Cercador semàntic que indexa informació de l'usuari en diferents capes i la fa
cercable per text complet, fuzzy matching i ranking.

---

### Dades que indexa

| Capa | Dades indexades | Origen de les dades | Com s'actualitza l'índex |
|------|----------------|---------------------|--------------------------|
| **Espai de treball** (space-{spaceId}) | Notes (notes/), Tasques (tasks/), Missatges (messages/), Accions humanes (human_actions/), Fitxers (files/) | Firestore → Cloud Function `syncToLlull` | Automàticament via Firestore trigger `syncToLlull`. Quan un document es crea/modifica/elimina a Firestore, la cloud function `syncToLlull` crida `POST /v1/space-{spaceId}/index` |
| **Marketplace** | Equips d'agents publicats (marketplace_teams/) | Firestore → Cloud Function `syncMarketplaceToLlull` | Automàticament via Firestore trigger en `marketplace_teams/{teamId}` |
| **Global** (global/) | Totes les dades indexades de tots els espais | Acumulació dels indexos individuals | Cada indexació a `space-{spaceId}` també replica a `global/` |

---

### Processos que executa

**1. Indexació automàtica (via Firestore trigger)**
```
Usuari crea/modifica document a Firestore
  → syncToLlull (Cloud Function)
  → POST /v1/space-{spaceId}/index { id, action: "INDEX", fields: { text, name, path, ... } }
  → POST /v1/global/index (replica a índex global)
  → Llull: afegeix al trie + persisteix a BoltDB
```

**2. Indexació de fitxers del data repo**
```
Masovera completa un run
  → xerraire crida syncDataFilesToFirestore()
  → Escriu fitxers a Firestore spaces/{spaceId}/files/
  → Firestore trigger syncToLlull
  → Llull indexa els fitxers
```

**3. Cerca**
```
Xerraire (searchDocuments)
  → GET /v1/{index}/search?q=query&limit=10
  → Llull: cerca al trie, aplica fuzzy matching, ranking
  → Torna hits ordenats per score
```

---

### Relacions amb altres components

| Component | Com es relaciona | Dades compartides |
|-----------|-----------------|-------------------|
| **Firestore** | Firestore triggers disparen `syncToLlull` → indexen a Llull | Totes les dades d'espai |
| **Functions** | `syncToLlull`, `searchSpace`, `syncMarketplaceToLlull` cloud functions | Index/delete operations |
| **Xerraire** | Crida `searchDocuments` → `GET /v1/space-{spaceId}/search` | Resultats de cerca |
| **Mycrew-api** | Crida `searchMarketplace` → `GET /v1/marketplace/search` | Resultats de marketplace |

---

### Sereno Logs

Events a stdout en format JSON sereno:

| Event | action.type/name | Quan |
|-------|-----------------|------|
| Search | `tool_execution/search` | Cerca executada (amb query, total_hits, latency_ms) |
| Index | `tool_execution/document_index` | Document indexat (amb doc_id, namespace) |
| Delete | `tool_execution/document_delete` | Document eliminat |
| Health | `tool_execution/health` | Health check |

---

### API

```bash
GET    /v1/{index}/search?q=query&limit=10&offset=0
POST   /v1/{index}/index          # { id, action: "INDEX"|"DELETE", fields: { name, content, path, type, ... } }
DELETE /v1/{index}/documents/{id}
GET    /v1/health
```

El paràmetre `{index}` pot ser:
- `space-{spaceId}` — índex d'un espai de treball
- `global` — índex global de tots els espais
- `marketplace` — índex del marketplace

Amb multi-tenant, el prefix s'afegeix automàticament:
- `makeyourcrew-space-{spaceId}`
- `makeyourcrew-global`
- `makeyourcrew-marketplace`

---

### Què NO fa llull

- ❌ No executa agents (ho fa masovera)
- ❌ No xateja amb usuaris (ho fa xerraire)
- ❌ No emmagatzema dades primàries (les dades són a Firestore)
- ❌ No gestiona usuaris ni permisos
- ❌ No fa indexació de codi font (només documents i fitxers de dades)
