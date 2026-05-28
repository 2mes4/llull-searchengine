package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/2mes4/llull/internal/engine"
	"github.com/2mes4/llull/internal/worker"
)

type Handlers struct {
	engine    *engine.SearchEngine
	pool      *worker.Pool
	authToken string
	startedAt time.Time
}

func NewHandlers(eng *engine.SearchEngine, pool *worker.Pool, authToken string) *Handlers {
	return &Handlers{
		engine:    eng,
		pool:      pool,
		authToken: authToken,
		startedAt: time.Now(),
	}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "ok",
		"docs_indexed": h.engine.DocCount(),
		"data_source":  h.engine.DataSource(),
		"queue_length": h.pool.QueueLen(),
		"goroutines":   runtime.NumGoroutine(),
		"memory_mb":    fmt.Sprintf("%.1f", float64(m.Alloc)/1024/1024),
		"uptime_sec":   int(h.startedAt.Unix()),
	})
}

func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	expected := fmt.Sprintf("Bearer %s", h.authToken)
	if authHeader != expected {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var payload engine.IndexPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if payload.ID == "" {
		http.Error(w, `{"error":"missing id"}`, http.StatusBadRequest)
		return
	}

	if payload.Action == "" {
		payload.Action = "INDEX"
	}

	if ok := h.pool.Enqueue(payload); !ok {
		http.Error(w, `{"error":"server busy, retry later"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "enqueued"})
}

func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(engine.PaginatedResponse{
			Hits:        []engine.SearchResult{},
			TotalHits:   0,
			Page:        1,
			HitsPerPage: 10,
			Query:       "",
		})
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	hitsPerPage, _ := strconv.Atoi(r.URL.Query().Get("hits_per_page"))
	useWeight := r.URL.Query().Get("use_weight") == "true"
	weightImpact, _ := strconv.ParseFloat(r.URL.Query().Get("weight_impact"), 64)
	fuzzy := r.URL.Query().Get("fuzzy") == "true"

	start := time.Now()
	req := engine.SearchRequest{
		Query:        query,
		UseWeight:    useWeight,
		WeightImpact: weightImpact,
		Page:         page,
		HitsPerPage:  hitsPerPage,
		Fuzzy:        fuzzy,
	}

	result := h.engine.Search(req)
	result.QueryTime = time.Since(start).Microseconds()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
