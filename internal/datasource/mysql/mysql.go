package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/2mes4/llull/internal/datasource"
	_ "github.com/go-sql-driver/mysql"
)

type Connector struct {
	db       *sql.DB
	cfg      datasource.Config
	table    string
	lastSync time.Time
}

func (c *Connector) Name() string { return "mysql" }

func (c *Connector) Connect(ctx context.Context, cfg datasource.Config) error {
	c.cfg = cfg
	c.table = cfg.Collection
	if c.table == "" {
		return fmt.Errorf("mysql: collection (table name) is required")
	}

	db, err := sql.Open("mysql", cfg.Connection)
	if err != nil {
		return fmt.Errorf("mysql: %w", err)
	}
	c.db = db

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("mysql ping: %w", err)
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

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

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
