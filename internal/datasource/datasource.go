package datasource

import "context"

type Document struct {
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

type Event struct {
	Action   string   `json:"action"`
	Document Document `json:"document"`
}

type Config struct {
	Type           string            `json:"type"`
	Connection     string            `json:"connection"`
	Index          string            `json:"index"`
	Fields         []string          `json:"fields"`
	WeightField    string            `json:"weight_field"`
	Collection     string            `json:"collection"`
	Options        map[string]string `json:"options"`
	PollInterval   string            `json:"poll_interval"`
	BatchSize      int               `json:"batch_size"`
}

type Connector interface {
	Name() string
	Connect(ctx context.Context, cfg Config) error
	Sync(ctx context.Context, callback func(Event)) error
	Close() error
}
