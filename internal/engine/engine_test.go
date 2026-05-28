package engine

import (
	"testing"
)

func TestEngineIndexAndSearch(t *testing.T) {
	eng := NewSearchEngine("")

	eng.IndexDocument(IndexPayload{
		ID:     "doc1",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Amadis de Gaula",
			"content": "El valiente caballero Amadis partio hacia la Insula Firme",
			"weight":  0.8,
		},
	})

	eng.IndexDocument(IndexPayload{
		ID:     "doc2",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Tirante el Blanco",
			"content": "Tirante fue a Constantinopla con gran compania",
			"weight":  0.6,
		},
	})

	result := eng.Search(SearchRequest{
		Query:       "amadis",
		Page:        1,
		HitsPerPage: 10,
	})

	if result.TotalHits != 1 {
		t.Fatalf("Expected 1 hit, got %d", result.TotalHits)
	}
	if result.Hits[0].ID != "doc1" {
		t.Errorf("Expected doc1, got %s", result.Hits[0].ID)
	}
}

func TestEngineDelete(t *testing.T) {
	eng := NewSearchEngine("")

	eng.IndexDocument(IndexPayload{
		ID:     "doc1",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Prueba",
			"content": "Contenido de prueba para eliminar",
		},
	})

	if eng.DocCount() != 1 {
		t.Fatalf("Expected 1 doc, got %d", eng.DocCount())
	}

	eng.DeleteDocument("doc1")

	if eng.DocCount() != 0 {
		t.Fatalf("Expected 0 docs after delete, got %d", eng.DocCount())
	}

	result := eng.Search(SearchRequest{
		Query:       "prueba",
		Page:        1,
		HitsPerPage: 10,
	})

	if result.TotalHits != 0 {
		t.Fatalf("Expected 0 hits after delete, got %d", result.TotalHits)
	}
}

func TestEnginePagination(t *testing.T) {
	eng := NewSearchEngine("")

	for i := 0; i < 25; i++ {
		eng.IndexDocument(IndexPayload{
			ID:     "doc-" + string(rune('A'+i)),
			Action: "INDEX",
			Fields: map[string]interface{}{
				"title":   "Caballero",
				"content": "El caballero partio en busca de aventura",
			},
		})
	}

	result := eng.Search(SearchRequest{
		Query:       "caballero",
		Page:        1,
		HitsPerPage: 10,
	})

	if len(result.Hits) != 10 {
		t.Fatalf("Expected 10 hits on page 1, got %d", len(result.Hits))
	}
	if result.NbPages != 3 {
		t.Errorf("Expected 3 pages, got %d", result.NbPages)
	}

	page2 := eng.Search(SearchRequest{
		Query:       "caballero",
		Page:        2,
		HitsPerPage: 10,
	})

	if len(page2.Hits) != 10 {
		t.Fatalf("Expected 10 hits on page 2, got %d", len(page2.Hits))
	}
}

func TestEngineWeightedRanking(t *testing.T) {
	eng := NewSearchEngine("")

	eng.IndexDocument(IndexPayload{
		ID:     "low-weight",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Caballero Bajo",
			"content": "caballero",
			"weight":  0.1,
		},
	})

	eng.IndexDocument(IndexPayload{
		ID:     "high-weight",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Caballero Alto",
			"content": "caballero",
			"weight":  0.9,
		},
	})

	result := eng.Search(SearchRequest{
		Query:        "caballero",
		UseWeight:    true,
		WeightImpact: 0.8,
		Page:         1,
		HitsPerPage:  10,
	})

	if len(result.Hits) < 2 {
		t.Fatalf("Expected at least 2 hits, got %d", len(result.Hits))
	}

	if result.Hits[0].ID != "high-weight" {
		t.Errorf("Expected high-weight doc first, got %s (score: %.3f)", result.Hits[0].ID, result.Hits[0].Score)
	}
}

func TestEngineFuzzySearch(t *testing.T) {
	eng := NewSearchEngine("")

	eng.IndexDocument(IndexPayload{
		ID:     "doc1",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Amadis de Gaula",
			"content": "El caballero Amadis",
		},
	})

	result := eng.Search(SearchRequest{
		Query:       "amadiz",
		Page:        1,
		HitsPerPage: 10,
		Fuzzy:       true,
	})

	if result.TotalHits < 1 {
		t.Fatalf("Expected at least 1 fuzzy hit for 'amadiz', got %d", result.TotalHits)
	}
}

func TestEngineEmptyQuery(t *testing.T) {
	eng := NewSearchEngine("")

	result := eng.Search(SearchRequest{
		Query:       "",
		Page:        1,
		HitsPerPage: 10,
	})

	if result.TotalHits != 0 {
		t.Fatalf("Expected 0 hits for empty query, got %d", result.TotalHits)
	}
}

func TestEngineUpdateDocument(t *testing.T) {
	eng := NewSearchEngine("")

	eng.IndexDocument(IndexPayload{
		ID:     "doc1",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Titulo Original",
			"content": "Contenido original",
		},
	})

	eng.IndexDocument(IndexPayload{
		ID:     "doc1",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Titulo Actualizado",
			"content": "Contenido actualizado con dragon",
		},
	})

	result := eng.Search(SearchRequest{
		Query:       "dragon",
		Page:        1,
		HitsPerPage: 10,
	})

	if result.TotalHits != 1 {
		t.Fatalf("Expected 1 hit for 'dragon', got %d", result.TotalHits)
	}
}
