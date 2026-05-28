package engine

import (
	"sync"
	"time"
)

type IndexPayload struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	UpdatedAt int64                  `json:"updated_at,omitempty"`
}

type SearchEngine struct {
	mu          sync.RWMutex
	Root        *TrieNode
	Metadata    map[string]DocMetadata
	docCount    int64
	persistPath string
	startedAt   time.Time
	dataSource  string
	dirty       bool
}

func NewSearchEngine(dataSource string) *SearchEngine {
	return &SearchEngine{
		Root:       newTrieNode(),
		Metadata:   make(map[string]DocMetadata),
		startedAt:  time.Now(),
		dataSource: dataSource,
	}
}

func (se *SearchEngine) IndexDocument(payload IndexPayload) {
	se.mu.Lock()
	defer se.mu.Unlock()

	weight := 0.0
	if w, ok := payload.Fields["weight"]; ok {
		switch v := w.(type) {
		case float64:
			weight = v
		case int:
			weight = float64(v)
		}
	}

	meta := DocMetadata{
		Fields: payload.Fields,
		Weight: weight,
	}

	oldMeta, exists := se.Metadata[payload.ID]
	if exists {
		se.removeDocFromTrie(oldMeta, payload.ID)
	}

	se.Metadata[payload.ID] = meta

	for _, value := range payload.Fields {
		if str, ok := value.(string); ok && str != "" {
			tokens := tokenize(str)
			uniqueTokens := make(map[string]struct{})
			for _, token := range tokens {
				uniqueTokens[token] = struct{}{}
			}
			for token := range uniqueTokens {
				insertIntoTrie(se.Root, token, payload.ID)
			}
		}
	}

	se.docCount++
	se.dirty = true
}

func (se *SearchEngine) DeleteDocument(id string) {
	se.mu.Lock()
	defer se.mu.Unlock()

	meta, exists := se.Metadata[id]
	if !exists {
		return
	}

	se.removeDocFromTrie(meta, id)
	delete(se.Metadata, id)
	se.docCount--
	se.dirty = true
}

func (se *SearchEngine) removeDocFromTrie(meta DocMetadata, id string) {
	for _, value := range meta.Fields {
		if str, ok := value.(string); ok && str != "" {
			tokens := tokenize(str)
			uniqueTokens := make(map[string]struct{})
			for _, token := range tokens {
				uniqueTokens[token] = struct{}{}
			}
			for token := range uniqueTokens {
				removeFromTrie(se.Root, token, id)
			}
		}
	}
}

func (se *SearchEngine) Search(req SearchRequest) PaginatedResponse {
	se.mu.RLock()
	defer se.mu.RUnlock()

	tokens := tokenize(req.Query)
	if len(tokens) == 0 {
		return PaginatedResponse{
			Hits:        []SearchResult{},
			TotalHits:   0,
			Page:        req.Page,
			HitsPerPage: req.HitsPerPage,
			Query:       req.Query,
		}
	}

	var matchedIDs []string
	matchedIDs = searchMultiToken(se.Root, tokens)

	if req.Fuzzy && len(matchedIDs) < 10 {
		fuzzyIDs := fuzzySearch(se.Root, req.Query, maxFuzzyDistance)
		combined := make(map[string]struct{})
		for _, id := range matchedIDs {
			combined[id] = struct{}{}
		}
		for _, id := range fuzzyIDs {
			combined[id] = struct{}{}
		}
		matchedIDs = make([]string, 0, len(combined))
		for id := range combined {
			matchedIDs = append(matchedIDs, id)
		}
	}

	return rankResults(matchedIDs, req, se.Metadata, tokens)
}

func (se *SearchEngine) DocCount() int64 {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.docCount
}

func (se *SearchEngine) Uptime() time.Duration {
	return time.Since(se.startedAt)
}

func (se *SearchEngine) DataSource() string {
	return se.dataSource
}

func (se *SearchEngine) SetPersistPath(path string) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.persistPath = path
}

func (se *SearchEngine) PersistPath() string {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.persistPath
}

func (se *SearchEngine) Dirty() bool {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.dirty
}

func (se *SearchEngine) MarkClean() {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.dirty = false
}
