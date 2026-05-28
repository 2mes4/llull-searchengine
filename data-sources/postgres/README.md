# Data Source: PostgreSQL

Syncs changes from a PostgreSQL table into the Llull search engine.

## Configuration

| Parameter | Required | Description |
|-----------|----------|-------------|
| `connection` | Yes | PostgreSQL connection string (e.g. `postgres://user:pass@host:5432/db?sslmode=disable`) |
| `collection` | Yes | Table name to watch |
| `fields` | No | Columns to index (defaults to all text columns) |
| `weight_field` | No | Numeric column to use as document weight (0.0–1.0) |
| `poll_interval` | No | How often to check for changes (default: `5s`) |

## Table Requirements

Your table must have an `id` column (primary key) and an `updated_at` timestamp column for change detection.

```sql
CREATE TABLE documents (
    id          TEXT PRIMARY KEY,
    title       TEXT,
    content     TEXT,
    weight      REAL DEFAULT 0.5,
    updated_at  TIMESTAMP DEFAULT NOW()
);
```

## Usage

```bash
llull -datasource config.json
```

Where `config.json` contains:

```json
{
  "type": "postgres",
  "connection": "postgres://user:pass@localhost:5432/mydb",
  "collection": "documents",
  "fields": ["title", "content"],
  "weight_field": "weight",
  "poll_interval": "10s"
}
```
