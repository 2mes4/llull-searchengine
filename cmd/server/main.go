package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/2mes4/llull/internal/api"
	"github.com/2mes4/llull/internal/datasource"
	"github.com/2mes4/llull/internal/datasource/postgres"
	"github.com/2mes4/llull/internal/engine"
	"github.com/2mes4/llull/internal/logging"
	"github.com/2mes4/llull/internal/seed"
	"github.com/2mes4/llull/internal/worker"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	authToken := flag.String("auth-token", os.Getenv("AUTH_TOKEN"), "Bearer token for index endpoint")
	workers := flag.Int("workers", envInt("WORKERS", 4), "Number of worker goroutines")
	bufferSize := flag.Int("buffer", envInt("BUFFER", 5000), "Worker queue buffer size")
	seedFile := flag.String("seed-file", os.Getenv("SEED_FILE"), "JSON file to seed the default index")
	generateSeed := flag.String("generate-seed", "", "Generate seed file at path and exit")
	seedDir := flag.String("seed-dir", os.Getenv("SEED_DIR"), "Directory with source text files")
	seedCount := flag.Int("seed-count", 1000, "Max number of seed documents to generate")
	dbPath := flag.String("db", os.Getenv("DB_PATH"), "Directory for persistent database files")
	defaultIndex := flag.String("default-index", os.Getenv("DEFAULT_INDEX"), "Default index name")
	indexTTL := flag.Duration("index-ttl", 30*time.Minute, "Time before unloading idle indices")
	flag.Parse()

	if *authToken == "" {
		*authToken = "llull-dev-token"
	}
	if *defaultIndex == "" {
		*defaultIndex = "default"
	}
	if *dbPath == "" {
		*dbPath = "/data"
	}

	tid := logging.NewTraceID()
	sid := logging.NewSpanID()

	if *generateSeed != "" {
		textsDir := *seedDir
		if textsDir == "" {
			textsDir = "data/llibres-llull"
		}
		texts, names, err := seed.LoadTextFiles(textsDir)
		if err != nil {
			texts, names = seed.EmbedFallbackTexts()
		}
		if len(texts) == 0 {
			texts, names = seed.EmbedFallbackTexts()
		}
		logging.Emit("info", fmt.Sprintf("Loaded %d text files, generating up to %d documents to %s", len(texts), *seedCount, *generateSeed), tid, sid, &logging.Fields{
			Action: &logging.Action{Type: "lifecycle", Name: "seed"},
		})
		if err := seed.GenerateSeedFile(*generateSeed, texts, names, *seedCount); err != nil {
			logging.Emit("fatal", fmt.Sprintf("Failed to generate seed file: %v", err), tid, sid, &logging.Fields{
				Action: &logging.Action{Type: "lifecycle", Name: "seed"},
			})
		}
		logging.Emit("info", "Seed file generated successfully", tid, sid, &logging.Fields{
			Action: &logging.Action{Type: "lifecycle", Name: "seed"},
		})
		return
	}

	mgr := engine.NewIndexManager(*dbPath, *indexTTL, *defaultIndex)
	eng := mgr.GetOrCreateIndex(*defaultIndex)
	pool := worker.NewPool(eng, *bufferSize, *workers)

	tenantPrefix := os.Getenv("LLULL_TENANT_PREFIX")
	if tenantPrefix != "" {
		logging.Emit("info", fmt.Sprintf("Tenant prefix enabled: %q", tenantPrefix), tid, sid, &logging.Fields{
			Action: &logging.Action{Type: "lifecycle", Name: "startup"},
		})
	}

	handlers := api.NewHandlers(mgr, pool, *authToken, tenantPrefix)
	router := api.NewRouter(handlers)

	if *seedFile != "" {
		loadSeedData(pool, *seedFile)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *dbPath != "" {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				mgr.SaveAll()
			}
		}()
	}

	dsType := os.Getenv("DATASOURCE_TYPE")
	if dsType != "" {
		go startDatasource(ctx, dsType, mgr, pool)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logging.Emit("info", fmt.Sprintf("Llull search engine running on :%d (default index: %q)", *port, *defaultIndex), tid, sid, &logging.Fields{
			Action: &logging.Action{Type: "lifecycle", Name: "startup"},
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Emit("fatal", fmt.Sprintf("Server error: %v", err), tid, sid, &logging.Fields{
				Action: &logging.Action{Type: "lifecycle", Name: "startup"},
			})
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Emit("info", "Shutting down server...", tid, sid, &logging.Fields{
		Action: &logging.Action{Type: "lifecycle", Name: "shutdown"},
	})
	cancel()
	mgr.Stop()
	pool.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
	logging.Emit("info", "Server stopped", tid, sid, &logging.Fields{
		Action: &logging.Action{Type: "lifecycle", Name: "shutdown"},
	})
}

func loadSeedData(pool *worker.Pool, path string) {
	tid := logging.NewTraceID()
	sid := logging.NewSpanID()

	logging.Emit("info", fmt.Sprintf("Loading seed data from %s", path), tid, sid, &logging.Fields{
		Action: &logging.Action{Type: "lifecycle", Name: "seed"},
	})

	docs, err := seed.LoadSeedFile(path)
	if err != nil {
		logging.Emit("warn", fmt.Sprintf("Could not load seed file: %v", err), tid, sid, &logging.Fields{
			Action: &logging.Action{Type: "lifecycle", Name: "seed"},
		})
		return
	}

	if len(docs) == 0 {
		logging.Emit("warn", "Seed file is empty, no documents to index", tid, sid, &logging.Fields{
			Action: &logging.Action{Type: "lifecycle", Name: "seed"},
		})
		return
	}

	payloads := seed.ToIndexPayloads(docs)

	enqueued := 0
	for _, p := range payloads {
		if pool.Enqueue(p) {
			enqueued++
		}
	}

	logging.Emit("info", fmt.Sprintf("Enqueued %d/%d documents for indexing", enqueued, len(payloads)), tid, sid, &logging.Fields{
		Action: &logging.Action{Type: "lifecycle", Name: "seed"},
	})

	b, _ := json.MarshalIndent(docs[0], "", "  ")
	logging.Emit("info", fmt.Sprintf("Sample document: %s", string(b)), tid, sid, &logging.Fields{
		Action: &logging.Action{Type: "lifecycle", Name: "seed"},
	})
}

func startDatasource(ctx context.Context, dsType string, mgr *engine.IndexManager, pool *worker.Pool) {
	tid := logging.NewTraceID()
	sid := logging.NewSpanID()

	dsIndex := os.Getenv("DATASOURCE_INDEX")
	if dsIndex == "" {
		dsIndex = "default"
	}
	mgr.GetOrCreateIndex(dsIndex)

	cfg := datasource.Config{
		Type:        dsType,
		Connection:  os.Getenv("DATASOURCE_CONNECTION"),
		Index:       dsIndex,
		Collection:  os.Getenv("DATASOURCE_COLLECTION"),
		Fields:      splitCSV(os.Getenv("DATASOURCE_FIELDS")),
		WeightField: os.Getenv("DATASOURCE_WEIGHT_FIELD"),
		PollInterval: os.Getenv("DATASOURCE_POLL_INTERVAL"),
		BatchSize:   500,
		Options: map[string]string{
			"timestamp_column": os.Getenv("DATASOURCE_TIMESTAMP_COLUMN"),
			"id_column":        os.Getenv("DATASOURCE_ID_COLUMN"),
		},
	}

	if cfg.Connection == "" || cfg.Collection == "" {
		logging.Emit("warn", fmt.Sprintf("Datasource %s: missing CONNECTION or COLLECTION, skipping", dsType), tid, sid, &logging.Fields{
			Action: &logging.Action{Type: "lifecycle", Name: "datasource"},
		})
		return
	}

	var conn datasource.Connector
	switch dsType {
	case "postgres":
		conn = &postgres.Connector{}
	default:
		logging.Emit("error", fmt.Sprintf("Unsupported datasource type: %s", dsType), tid, sid, &logging.Fields{
			Action: &logging.Action{Type: "lifecycle", Name: "datasource"},
		})
		return
	}

	if err := conn.Connect(ctx, cfg); err != nil {
		logging.Emit("error", fmt.Sprintf("Datasource connect error: %v", err), tid, sid, &logging.Fields{
			Action: &logging.Action{Type: "lifecycle", Name: "datasource"},
		})
		return
	}
	defer conn.Close()

	logging.Emit("info", fmt.Sprintf("Starting %s datasource sync (index=%s, collection=%s)", dsType, dsIndex, cfg.Collection), tid, sid, &logging.Fields{
		Action: &logging.Action{Type: "lifecycle", Name: "datasource"},
	})

	conn.Sync(ctx, func(ev datasource.Event) {
		payload := engine.IndexPayload{
			ID:     ev.Document.ID,
			Action: ev.Action,
			Fields: ev.Document.Fields,
		}
		for !pool.Enqueue(payload) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Millisecond):
			}
		}
	})
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, f := range strings.Split(s, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
