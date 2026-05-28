package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/2mes4/llull/internal/datasource"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connector struct {
	client   *mongo.Client
	db       *mongo.Database
	coll     *mongo.Collection
	cfg      datasource.Config
	lastSync time.Time
}

func (c *Connector) Name() string { return "mongodb" }

func (c *Connector) Connect(ctx context.Context, cfg datasource.Config) error {
	c.cfg = cfg
	collName := cfg.Collection
	if collName == "" {
		return fmt.Errorf("mongodb: collection name is required")
	}

	clientOpts := options.Client().ApplyURI(cfg.Connection)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("mongodb: %w", err)
	}
	c.client = client

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("mongodb ping: %w", err)
	}

	dbName := "llull"
	if v, ok := cfg.Options["database"]; ok {
		dbName = v
	}
	c.db = client.Database(dbName)
	c.coll = c.db.Collection(collName)

	return nil
}

func (c *Connector) Sync(ctx context.Context, callback func(datasource.Event)) error {
	var pipeline []bson.M

	if !c.lastSync.IsZero() {
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{
				"updatedAt": bson.M{"$gt": primitive.NewDateTimeFromTime(c.lastSync)},
			},
		})
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		cursor, err := c.coll.Find(ctx, bson.M{})
		if err != nil {
			return fmt.Errorf("mongodb find: %w", err)
		}

		var results []bson.M
		if err := cursor.All(ctx, &results); err != nil {
			cursor.Close(ctx)
			return fmt.Errorf("mongodb cursor: %w", err)
		}
		cursor.Close(ctx)
		_ = results

		c.lastSync = time.Now()

		interval := 5 * time.Second
		if c.cfg.PollInterval != "" {
			if d, err := time.ParseDuration(c.cfg.PollInterval); err == nil {
				interval = d
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

func (c *Connector) Close() error {
	if c.client != nil {
		return c.client.Disconnect(context.Background())
	}
	return nil
}
