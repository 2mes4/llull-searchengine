package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/2mes4/llull/internal/engine"
	"github.com/2mes4/llull/internal/worker"
)

func setupTestAPI(t *testing.T) (*httptest.Server, *engine.SearchEngine) {
	t.Helper()
	eng := engine.NewSearchEngine("")
	pool := worker.NewPool(eng, 100, 2)
	t.Cleanup(func() { pool.Stop() })
	handlers := NewHandlers(eng, pool, "test-token")
	router := NewRouter(handlers)
	return httptest.NewServer(router), eng
}

func TestHealthEndpoint(t *testing.T) {
	srv, _ := setupTestAPI(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/v1/health")
	if err != nil {
		t.Fatalf("Health request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("Expected status ok, got %v", body["status"])
	}
}

func TestIndexEndpointAuth(t *testing.T) {
	srv, _ := setupTestAPI(t)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/v1/index", "application/json", strings.NewReader(`{"id":"doc1"}`))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Fatalf("Expected 401 without auth, got %d", resp.StatusCode)
	}
}

func TestIndexEndpointSuccess(t *testing.T) {
	srv, _ := setupTestAPI(t)
	defer srv.Close()

	req, _ := http.NewRequest("POST", srv.URL+"/v1/index", strings.NewReader(`{
		"id": "doc1",
		"action": "INDEX",
		"fields": {
			"title": "Test Document",
			"content": "contenido de prueba para buscar"
		}
	}`))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		t.Fatalf("Expected 202, got %d", resp.StatusCode)
	}

	time.Sleep(100 * time.Millisecond)
}

func TestSearchEndpoint(t *testing.T) {
	srv, eng := setupTestAPI(t)
	defer srv.Close()

	eng.IndexDocument(engine.IndexPayload{
		ID:     "doc1",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Amadis de Gaula",
			"content": "El caballero Amadis",
		},
	})

	resp, err := http.Get(srv.URL + "/v1/search?q=amadis&page=1&hits_per_page=10")
	if err != nil {
		t.Fatalf("Search request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	var result engine.PaginatedResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.TotalHits != 1 {
		t.Errorf("Expected 1 hit, got %d", result.TotalHits)
	}
	if len(result.Hits) != 1 {
		t.Errorf("Expected 1 hit in page, got %d", len(result.Hits))
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	srv, _ := setupTestAPI(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/v1/search?q=")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	var result engine.PaginatedResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.TotalHits != 0 {
		t.Errorf("Expected 0 hits for empty query, got %d", result.TotalHits)
	}
}

func TestSearchWithWeight(t *testing.T) {
	srv, eng := setupTestAPI(t)
	defer srv.Close()

	eng.IndexDocument(engine.IndexPayload{
		ID:     "doc1",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Caballero Bajo",
			"content": "caballero aventura",
			"weight":  0.1,
		},
	})

	eng.IndexDocument(engine.IndexPayload{
		ID:     "doc2",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Caballero Alto",
			"content": "caballero aventura",
			"weight":  0.9,
		},
	})

	resp, err := http.Get(srv.URL + "/v1/search?q=caballero&use_weight=true&weight_impact=0.8&page=1&hits_per_page=10")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	var result engine.PaginatedResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Hits) < 2 {
		t.Fatalf("Expected 2 hits, got %d", len(result.Hits))
	}
	if result.Hits[0].ID != "doc2" {
		t.Errorf("Expected doc2 (high weight) first, got %s", result.Hits[0].ID)
	}
}

func TestIndexEndpointMissingID(t *testing.T) {
	srv, _ := setupTestAPI(t)
	defer srv.Close()

	req, _ := http.NewRequest("POST", srv.URL+"/v1/index", strings.NewReader(`{"action":"INDEX"}`))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Fatalf("Expected 400 for missing ID, got %d", resp.StatusCode)
	}
}
