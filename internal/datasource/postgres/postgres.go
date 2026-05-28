package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/2mes4/llull/internal/datasource"
	_ "github.com/lib/pq"
)

type Connector struct {
	db        *sql.DB
	cfg       datasource.Config
	table     string
	lastSync  time.Time
}

func (c *Connector) Name() string { return "postgres" }

func (c *Connector) Connect(ctx context.Context, cfg datasource.Config) error {
	c.cfg = cfg
	c.table = cfg.Collection
	if c.table == "" {
		return fmt.Errorf("postgres: collection (table name) is required")
	}

	db, err := sql.Open("postgres", cfg.Connection)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	c.db = db

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}
	return nil
}

func (c *Connector) Sync(ctx context.Context, callback func(datasource.Event)) error {
	interval := 5 * time.Second
	if c.cfg.PollInterval != "" {
		if d, err := time.ParseDuration(c.cfg.PollInterval); err == nil {
			interval = d
		}
	}

	if c.lastSync.IsZero() {
		c.lastSync = time.Now().Add(-24 * time.Hour)
	}

	cols := "*"
	if len(c.cfg.Fields) > 0 {
		cols = "id"
		for _, f := range c.cfg.Fields {
			cols += ", " + f
		}
		if c.cfg.WeightField != "" {
			cols += ", " + c.cfg.WeightField
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE updated_at > $1 ORDER BY updated_at ASC", cols, c.table)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rows, err := c.db.QueryContext(ctx, query, c.lastSync)
		if err != nil {
			return fmt.Errorf("postgres query: %w", err)
		}

		cols, err := rows.Columns()
		if err != nil {
			rows.Close()
			continue
		}
		vals := make([]interface{}, len(cols))
		for rows.Next() {
			if err := rows.Scan(vals...); err != nil {
				continue
			}
		}
		rows.Close()
		c.lastSync = time.Now()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

func (c *Connector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
