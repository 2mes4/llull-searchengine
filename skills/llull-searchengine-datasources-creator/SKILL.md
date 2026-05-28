# Skill: Create a New Data Source Connector

This skill guides you through building a new data source connector for Llull.

## Overview

A data source connector syncs documents from an external database into the Llull search engine. Every connector implements the `datasource.Connector` interface found in `internal/datasource/datasource.go`.

## Architecture

### Two sync models

1. **Push model** (like Firestore): An external trigger (Cloud Function, webhook, etc.) sends HTTP POST requests to `POST /v1/index`. No Go connector code needed — just configuration and external glue code.

2. **Poll model** (like PostgreSQL, MySQL, MongoDB): A Go connector polls the database at a configurable interval, detects changes (usually via timestamp), and feeds them into the engine through a callback.

### The interface

```go
type Connector interface {
    Name() string
    Connect(ctx context.Context, cfg Config) error
    Sync(ctx context.Context, callback func(Event)) error
    Close() error
}
```

### Shared types

```go
type Config struct {
    Type         string            // "postgres", "mysql", "mongodb", etc.
    Connection   string            // Connection string / URI
    Fields       []string          // Fields to index
    WeightField  string            // Field to use as document weight
    Collection   string            // Table / collection name
    Options      map[string]string // Connector-specific options
    PollInterval string            // How often to poll (e.g. "10s")
    BatchSize    int               // Number of rows to fetch per poll
}

type Document struct {
    ID     string
    Fields map[string]interface{}
}

type Event struct {
    Action   string   // "INDEX" or "DELETE"
    Document Document
}
```

## Step-by-step guide

### Step 1: Create the connector package

Create a new package under `internal/datasource/<name>/`:

```
internal/datasource/<name>/<name>.go
```

### Step 2: Implement the Connector interface

```go
package <name>

import (
    "context"
    "fmt"
    "github.com/2mes4/llull/internal/datasource"
)

type Connector struct {
    cfg datasource.Config
}

func (c *Connector) Name() string { return "<name>" }

func (c *Connector) Connect(ctx context.Context, cfg datasource.Config) error {
    c.cfg = cfg
    // Validate required fields
    if cfg.Connection == "" {
        return fmt.Errorf("<name>: connection string is required")
    }
    if cfg.Collection == "" {
        return fmt.Errorf("<name>: collection name is required")
    }
    // Connect to the database
    return nil
}

func (c *Connector) Sync(ctx context.Context, callback func(datasource.Event)) error {
    // Poll or listen for changes
    // For each changed document, call:
    //   callback(datasource.Event{
    //       Action: "INDEX",
    //       Document: datasource.Document{
    //           ID: "...",
    //           Fields: map[string]interface{}{...},
    //       },
    //   })
    // Respect ctx.Done() for graceful shutdown
    <-ctx.Done()
    return ctx.Err()
}

func (c *Connector) Close() error {
    // Clean up connections
    return nil
}
```

### Step 3: Create the data-sources directory

Create `data-sources/<name>/README.md` with:
- Overview
- Configuration parameters table
- Schema/table requirements
- Usage example with `config.json`
- Connection string format

### Step 4: Add the database driver

Add the Go import for the database driver in your connector file:

```go
import _ "github.com/driver/package"
```

Run `go mod tidy` to update dependencies.

### Step 5: Write tests

Create `internal/datasource/<name>/<name>_test.go`:

```go
package <name>

import "testing"

func TestConnectorName(t *testing.T) {
    c := &Connector{}
    if c.Name() != "<name>" {
        t.Errorf("expected name '<name>', got %s", c.Name())
    }
}

func TestConnectValidation(t *testing.T) {
    c := &Connector{}
    err := c.Connect(t.Context(), datasource.Config{})
    if err == nil {
        t.Error("expected error for empty config")
    }
}
```

### Step 6: Register the connector

If using the poll model, add the connector to the factory in `internal/datasource/registry.go` (create if needed) so it can be selected by `Config.Type`.

## Checklist

- [ ] Package `internal/datasource/<name>/<name>.go` implements `Connector` interface
- [ ] `Connect()` validates required config and establishes connection
- [ ] `Sync()` respects context cancellation
- [ ] `Close()` releases resources
- [ ] `data-sources/<name>/README.md` with full documentation
- [ ] Database driver added to `go.mod`
- [ ] Tests in `internal/datasource/<name>/<name>_test.go`
- [ ] All code and documentation in English

## Examples

Study these existing connectors for reference:
- `internal/datasource/postgres/postgres.go` — SQL database with timestamp polling
- `internal/datasource/mysql/mysql.go` — SQL database with timestamp polling
- `internal/datasource/mongodb/mongodb.go` — NoSQL with cursor-based polling
- `data-sources/firestore/` — Push model with Cloud Function (no Go connector needed)
