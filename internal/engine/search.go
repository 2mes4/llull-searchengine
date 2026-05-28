package engine

import (
	"math"
	"sort"
)

const defaultMaxExecutionPool = 1000

type DocMetadata struct {
	Fields map[string]interface{}
	Weight float64
}

type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float64                `json:"score"`
	Weight  float64                `json:"weight"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

type SearchRequest struct {
	Query        string  `json:"query"`
	UseWeight    bool    `json:"use_weight"`
	WeightImpact float64 `json:"weight_impact"`
	Page         int     `json:"page"`
	HitsPerPage  int     `json:"hits_per_page"`
	Fuzzy        bool    `json:"fuzzy"`
}

type PaginatedResponse struct {
	Hits        []SearchResult `json:"hits"`
	TotalHits   int            `json:"total_hits"`
	Page        int            `json:"page"`
	NbPages     int            `json:"nb_pages"`
	HitsPerPage int            `json:"hits_per_page"`
	Query       string         `json:"query"`
	QueryTime   int64          `json:"query_time"`
}

func searchByPrefix(root *TrieNode, prefix string) []string {
	node := findPrefixNode(root, prefix)
	if node == nil {
		return nil
	}
	ids := make(map[string]struct{})
	collectAllDocIDs(node, ids)
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	return result
}

func searchMultiToken(root *TrieNode, tokens []string) []string {
	if len(tokens) == 0 {
		return nil
	}

	firstIDs := searchByPrefix(root, tokens[0])
	if len(firstIDs) == 0 {
		return nil
	}

	if len(tokens) == 1 {
		return firstIDs
	}

	firstSet := make(map[string]struct{}, len(firstIDs))
	for _, id := range firstIDs {
		firstSet[id] = struct{}{}
	}

	for _, token := range tokens[1:] {
		tokenIDs := searchByPrefix(root, token)
		if len(tokenIDs) == 0 {
			return nil
		}
		tokenSet := make(map[string]struct{}, len(tokenIDs))
		for _, id := range tokenIDs {
			tokenSet[id] = struct{}{}
		}
		for id := range firstSet {
			if _, exists := tokenSet[id]; !exists {
				delete(firstSet, id)
			}
		}
	}

	result := make([]string, 0, len(firstSet))
	for id := range firstSet {
		result = append(result, id)
	}
	return result
}

func calculateTextScore(query string, docFields map[string]interface{}, tokens []string) float64 {
	score := 0.0
	queryLower := query

	for _, value := range docFields {
		str, ok := value.(string)
		if !ok {
			continue
		}
		strLower := str
		if containsExact(strLower, queryLower) {
			score += 1.0
		}
		for _, token := range tokens {
			if containsExact(strLower, token) {
				score += 0.5
			}
		}
	}

	return math.Min(score, 1.0)
}

func containsExact(text, substr string) bool {
	return len(text) >= len(substr) &&
		(text == substr ||
			(len(text) > len(substr) && hasSubstringAt(text, substr)))
}

func hasSubstringAt(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if text[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func rankResults(matchedIDs []string, req SearchRequest, metadata map[string]DocMetadata, tokens []string) PaginatedResponse {
	totalHits := len(matchedIDs)

	if totalHits == 0 {
		return PaginatedResponse{
			Hits:        []SearchResult{},
			TotalHits:   0,
			Page:        req.Page,
			HitsPerPage: req.HitsPerPage,
			Query:       req.Query,
		}
	}

	if req.HitsPerPage <= 0 {
		req.HitsPerPage = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	poolSize := defaultMaxExecutionPool
	if totalHits < poolSize {
		poolSize = totalHits
	}

	workingIDs := matchedIDs[:poolSize]

	results := make([]SearchResult, 0, len(workingIDs))
	for _, id := range workingIDs {
		meta, exists := metadata[id]
		if !exists {
			continue
		}

		textScore := calculateTextScore(req.Query, meta.Fields, tokens)
		if textScore == 0 {
			textScore = 0.7
		}

		finalScore := textScore

		if req.UseWeight {
			impact := req.WeightImpact
			if impact < 0 {
				impact = 0
			}
			if impact > 1 {
				impact = 1
			}
			finalScore = (textScore * (1.0 - impact)) + (meta.Weight * impact)
		}

		results = append(results, SearchResult{
			ID:     id,
			Score:  finalScore,
			Weight: meta.Weight,
			Fields: meta.Fields,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	offset := (req.Page - 1) * req.HitsPerPage
	end := offset + req.HitsPerPage

	if offset > len(results) {
		offset = len(results)
	}
	if end > len(results) {
		end = len(results)
	}

	paginated := results[offset:end]
	nbPages := int(math.Ceil(float64(totalHits) / float64(req.HitsPerPage)))

	return PaginatedResponse{
		Hits:        paginated,
		TotalHits:   totalHits,
		Page:        req.Page,
		NbPages:     nbPages,
		HitsPerPage: req.HitsPerPage,
		Query:       req.Query,
	}
}
