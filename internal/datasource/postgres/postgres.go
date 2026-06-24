package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/2mes4/llull/internal/datasource"
	_ "github.com/lib/pq"
)

type Connector struct {
	db          *sql.DB
	cfg         datasource.Config
	table       string
	lastSync    time.Time
	tsColumn    string
	idColumn    string
	initialDone bool
}

func (c *Connector) Name() string { return "postgres" }

func (c *Connector) Connect(ctx context.Context, cfg datasource.Config) error {
	c.cfg = cfg
	c.table = cfg.Collection
	if c.table == "" {
		return fmt.Errorf("postgres: collection (table name) is required")
	}

	c.tsColumn = cfg.Options["timestamp_column"]
	if c.tsColumn == "" {
		c.tsColumn = "updated_at"
	}
	c.idColumn = cfg.Options["id_column"]
	if c.idColumn == "" {
		c.idColumn = "id"
	}

	db, err := sql.Open("postgres", cfg.Connection)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	c.db = db

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}
	log.Printf("[postgres] connected to table %s (ts=%s id=%s)", c.table, c.tsColumn, c.idColumn)
	return nil
}

func (c *Connector) buildColumns() string {
	if len(c.cfg.Fields) == 0 {
		return "*"
	}
	cols := c.idColumn
	for _, f := range c.cfg.Fields {
		cols += ", " + f
	}
	if c.cfg.WeightField != "" {
		cols += ", " + c.cfg.WeightField
	}
	return cols
}

func (c *Connector) scanRows(rows *sql.Rows, callback func(datasource.Event)) int {
	colNames, err := rows.Columns()
	if err != nil {
		return 0
	}

	count := 0
	for rows.Next() {
		vals := make([]interface{}, len(colNames))
		ptrs := make([]interface{}, len(colNames))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			log.Printf("[postgres] scan error: %v", err)
			continue
		}

		fields := make(map[string]interface{})
		var docID string
		for i, name := range colNames {
			v := vals[i]
			switch val := v.(type) {
		 case []byte:
				v = string(val)
			}
			fields[name] = v
			if name == c.idColumn {
				docID = fmt.Sprintf("%v", v)
			}
		}
		if docID == "" {
			continue
		}

		if c.cfg.WeightField != "" {
			if w, ok := fields[c.cfg.WeightField]; ok {
				fields["weight"] = w
			}
		}

		callback(datasource.Event{
			Action: "INDEX",
			Document: datasource.Document{
				ID:     docID,
				Fields: fields,
			},
		})
		count++
	}
	return count
}

func (c *Connector) Sync(ctx context.Context, callback func(datasource.Event)) error {
	interval := 10 * time.Second
	if c.cfg.PollInterval != "" {
		if d, err := time.ParseDuration(c.cfg.PollInterval); err == nil {
			interval = d
		}
	}

	cols := c.buildColumns()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !c.initialDone {
			log.Printf("[postgres] initial full sync from %s", c.table)
			query := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s", cols, c.table, c.idColumn)
			rows, err := c.db.QueryContext(ctx, query)
			if err != nil {
				log.Printf("[postgres] full sync query error: %v", err)
				time.Sleep(interval)
				continue
			}
			n := c.scanRows(rows, callback)
			rows.Close()
			c.initialDone = true
			c.lastSync = time.Now()
			log.Printf("[postgres] full sync done: %d rows", n)
		} else {
			query := fmt.Sprintf("SELECT %s FROM %s WHERE %s > $1 ORDER BY %s ASC",
				cols, c.table, c.tsColumn, c.tsColumn)
			rows, err := c.db.QueryContext(ctx, query, c.lastSync)
			if err != nil {
				log.Printf("[postgres] poll query error: %v", err)
				time.Sleep(interval)
				continue
			}
			n := c.scanRows(rows, callback)
			rows.Close()
			if n > 0 {
				log.Printf("[postgres] incremental sync: %d changed rows", n)
			}
			c.lastSync = time.Now()
		}

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
