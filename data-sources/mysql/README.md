# Data Source: MySQL

Syncs changes from a MySQL table into the Llull search engine.

## Configuration

| Parameter | Required | Description |
|-----------|----------|-------------|
| `connection` | Yes | MySQL DSN (e.g. `user:password@tcp(host:3306)/dbname`) |
| `collection` | Yes | Table name to watch |
| `fields` | No | Columns to index (defaults to all text columns) |
| `weight_field` | No | Numeric column to use as document weight (0.0–1.0) |
| `poll_interval` | No | How often to check for changes (default: `5s`) |

## Table Requirements

Your table must have an `id` column (primary key) and an `updated_at` timestamp column for change detection.

```sql
CREATE TABLE documents (
    id          VARCHAR(255) PRIMARY KEY,
    title       TEXT,
    content     TEXT,
    weight      FLOAT DEFAULT 0.5,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

## Usage

```bash
llull -datasource config.json
```

Where `config.json` contains:

```json
{
  "type": "mysql",
  "connection": "user:password@tcp(localhost:3306)/mydb",
  "collection": "documents",
  "fields": ["title", "content"],
  "weight_field": "weight",
  "poll_interval": "10s"
}
```
