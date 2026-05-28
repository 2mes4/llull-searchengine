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
	manager   *engine.IndexManager
	pool      *worker.Pool
	authToken string
	startedAt time.Time
}

func NewHandlers(mgr *engine.IndexManager, pool *worker.Pool, authToken string) *Handlers {
	return &Handlers{
		manager:   mgr,
		pool:      pool,
		authToken: authToken,
		startedAt: time.Now(),
	}
}

func (h *Handlers) resolveIndex(r *http.Request) string {
	idx := r.PathValue("index")
	if idx == "" {
		idx = h.manager.DefaultIndex()
	}
	return idx
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	indices := h.manager.ListIndices()
	totalDocs := int64(0)
	indexInfo := make(map[string]map[string]interface{})
	for _, name := range indices {
		count, loaded := h.manager.IndexInfo(name)
		indexInfo[name] = map[string]interface{}{
			"docs":   count,
			"loaded": loaded,
		}
		totalDocs += count
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "ok",
		"docs_indexed":   totalDocs,
		"data_source":    "seed",
		"default_index":  h.manager.DefaultIndex(),
		"indices":        indexInfo,
		"queue_length":   h.pool.QueueLen(),
		"goroutines":     runtime.NumGoroutine(),
		"memory_mb":      fmt.Sprintf("%.1f", float64(m.Alloc)/1024/1024),
		"total_memory_mb": fmt.Sprintf("%.1f", float64(m.Sys)/1024/1024),
		"uptime_sec":     int(h.startedAt.Unix()),
	})
}

func (h *Handlers) Indices(w http.ResponseWriter, r *http.Request) {
	indices := h.manager.ListIndices()
	info := make(map[string]map[string]interface{})
	for _, name := range indices {
		count, loaded := h.manager.IndexInfo(name)
		info[name] = map[string]interface{}{
			"docs":   count,
			"loaded": loaded,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"indices":       info,
		"default_index": h.manager.DefaultIndex(),
	})
}

func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	idx := h.resolveIndex(r)

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

	eng := h.manager.GetOrCreateIndex(idx)
	if payload.Action == "INDEX" {
		eng.IndexDocument(payload)
	} else if payload.Action == "DELETE" {
		eng.DeleteDocument(payload.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "index": idx})
}

func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	idx := h.resolveIndex(r)
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

	eng := h.manager.GetIndex(idx)
	if eng == nil {
		eng = h.manager.GetOrCreateIndex(idx)
	}

	start := time.Now()
	req := engine.SearchRequest{
		Query:        query,
		UseWeight:    useWeight,
		WeightImpact: weightImpact,
		Page:         page,
		HitsPerPage:  hitsPerPage,
		Fuzzy:        fuzzy,
	}

	result := eng.Search(req)
	result.QueryTime = time.Since(start).Microseconds()
	result.Index = idx

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handlers) IndexHandlerLegacy(w http.ResponseWriter, r *http.Request) {
	idx := h.manager.DefaultIndex()
	eng := h.manager.GetOrCreateIndex(idx)

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

	if payload.Action == "INDEX" {
		eng.IndexDocument(payload)
	} else if payload.Action == "DELETE" {
		eng.DeleteDocument(payload.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "index": idx})
}
