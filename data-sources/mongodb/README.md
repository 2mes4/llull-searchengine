# Data Source: MongoDB

Syncs changes from a MongoDB collection into the Llull search engine.

## Configuration

| Parameter | Required | Description |
|-----------|----------|-------------|
| `connection` | Yes | MongoDB connection URI (e.g. `mongodb://localhost:27017`) |
| `collection` | Yes | Collection name to watch |
| `options.database` | Yes | Database name |
| `fields` | No | Fields to index (defaults to all string fields) |
| `weight_field` | No | Numeric field to use as document weight (0.0–1.0) |
| `poll_interval` | No | How often to check for changes (default: `5s`) |

## Document Requirements

Your documents should have a string `_id` and an `updatedAt` field for change detection.

```javascript
{
  "_id": "doc-001",
  "title": "My Document",
  "content": "The text to search...",
  "weight": 0.8,
  "updatedAt": ISODate("2024-01-01T00:00:00Z")
}
```

## Usage

```bash
llull -datasource config.json
```

Where `config.json` contains:

```json
{
  "type": "mongodb",
  "connection": "mongodb://localhost:27017",
  "collection": "documents",
  "options": {
    "database": "myapp"
  },
  "fields": ["title", "content"],
  "weight_field": "weight",
  "poll_interval": "10s"
}
```
