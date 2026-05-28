package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/2mes4/llull/internal/api"
	"github.com/2mes4/llull/internal/engine"
	"github.com/2mes4/llull/internal/seed"
	"github.com/2mes4/llull/internal/worker"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	authToken := flag.String("auth-token", os.Getenv("AUTH_TOKEN"), "Bearer token for index endpoint")
	workers := flag.Int("workers", 4, "Number of worker goroutines")
	bufferSize := flag.Int("buffer", 5000, "Worker queue buffer size")
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
		log.Printf("Loaded %d text files, generating up to %d documents to %s...", len(texts), *seedCount, *generateSeed)
		if err := seed.GenerateSeedFile(*generateSeed, texts, names, *seedCount); err != nil {
			log.Fatalf("Failed to generate seed file: %v", err)
		}
		log.Printf("Seed file generated successfully")
		return
	}

	mgr := engine.NewIndexManager(*dbPath, *indexTTL, *defaultIndex)
	eng := mgr.GetOrCreateIndex(*defaultIndex)
	pool := worker.NewPool(eng, *bufferSize, *workers)

	handlers := api.NewHandlers(mgr, pool, *authToken)
	router := api.NewRouter(handlers)

	if *seedFile != "" {
		loadSeedData(pool, *seedFile)
	}

	if *dbPath != "" {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				mgr.SaveAll()
			}
		}()
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Llull search engine running on :%d (default index: %q)", *port, *defaultIndex)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	mgr.Stop()
	pool.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Server stopped")
}

func loadSeedData(pool *worker.Pool, path string) {
	log.Printf("Loading seed data from %s...", path)

	docs, err := seed.LoadSeedFile(path)
	if err != nil {
		log.Printf("Warning: could not load seed file: %v", err)
		return
	}

	if len(docs) == 0 {
		log.Println("Warning: seed file is empty, no documents to index")
		return
	}

	payloads := seed.ToIndexPayloads(docs)

	enqueued := 0
	for _, p := range payloads {
		if pool.Enqueue(p) {
			enqueued++
		}
	}

	log.Printf("Enqueued %d/%d documents for indexing", enqueued, len(payloads))

	b, _ := json.MarshalIndent(docs[0], "", "  ")
	log.Printf("Sample document: %s", string(b))
}
