package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"time"
)

type Action struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

type Fields struct {
	Action       *Action
	LatencyMs    int64
	Namespace    string
	ParentSpanID string
	Extra        map[string]interface{}
}

type ctxKey string

const (
	ctxKeyTraceID ctxKey = "trace_id"
	ctxKeySpanID  ctxKey = "span_id"
)

func NewTraceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func NewSpanID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Emit(level, msg, traceID, spanID string, fields *Fields) {
	entry := map[string]interface{}{
		"timestamp":  time.Now().UTC().Format("2006-01-02T15:04:05.000") + "Z",
		"level":      level,
		"app":        "llull",
		"agent_role": "search-engine",
		"trace_id":   traceID,
		"span_id":    spanID,
		"message":    msg,
	}
	if fields != nil {
		if fields.Action != nil {
			entry["action"] = fields.Action
		}
		if fields.LatencyMs > 0 {
			entry["latency_ms"] = fields.LatencyMs
		}
		if fields.Namespace != "" {
			entry["namespace"] = fields.Namespace
		}
		if fields.ParentSpanID != "" {
			entry["parent_span_id"] = fields.ParentSpanID
		}
		for k, v := range fields.Extra {
			entry[k] = v
		}
	}

	json.NewEncoder(os.Stdout).Encode(entry)

	if level == "fatal" {
		os.Exit(1)
	}
}

func WithTraceContext(ctx context.Context) (context.Context, string, string) {
	traceID := NewTraceID()
	spanID := NewSpanID()
	ctx = context.WithValue(ctx, ctxKeyTraceID, traceID)
	ctx = context.WithValue(ctx, ctxKeySpanID, spanID)
	return ctx, traceID, spanID
}

func TraceFromContext(ctx context.Context) (string, string) {
	traceID, _ := ctx.Value(ctxKeyTraceID).(string)
	spanID, _ := ctx.Value(ctxKeySpanID).(string)
	return traceID, spanID
}
