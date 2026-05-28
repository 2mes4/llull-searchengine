package engine

const maxFuzzyDistance = 2

func fuzzySearch(root *TrieNode, query string, maxDist int) []string {
	if maxDist <= 0 {
		maxDist = maxFuzzyDistance
	}
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return nil
	}

	allResults := make(map[string]struct{})
	for _, token := range tokens {
		tokenRunes := []rune(token)
		ids := make(map[string]struct{})
		levenshteinDFS(root, tokenRunes, 0, make([]int, len(tokenRunes)+1), maxDist, ids)
		for id := range ids {
			allResults[id] = struct{}{}
		}
	}

	result := make([]string, 0, len(allResults))
	for id := range allResults {
		result = append(result, id)
	}
	return result
}

func levenshteinDFS(
	node *TrieNode,
	query []rune,
	depth int,
	prevRow []int,
	maxDist int,
	result map[string]struct{},
) {
	for i := range prevRow {
		prevRow[i] = i
	}
	levenshteinDFSRecursive(node, query, depth, prevRow, maxDist, result)
}

func levenshteinDFSRecursive(
	node *TrieNode,
	query []rune,
	depth int,
	prevRow []int,
	maxDist int,
	result map[string]struct{},
) {
	currentRow := make([]int, len(query)+1)
	currentRow[0] = depth + 1

	minVal := currentRow[0]

	for i := 1; i <= len(query); i++ {
		del := prevRow[i] + 1
		ins := currentRow[i-1] + 1
		sub := prevRow[i-1]

		if sub < del {
			if sub < ins {
				currentRow[i] = sub
			} else {
				currentRow[i] = ins
			}
		} else {
			if del < ins {
				currentRow[i] = del
			} else {
				currentRow[i] = ins
			}
		}

		if currentRow[i] < minVal {
			minVal = currentRow[i]
		}
	}

	if len(node.DocIDs) > 0 {
		lastVal := currentRow[len(query)]
		if lastVal <= maxDist {
			for _, id := range node.DocIDs {
				result[id] = struct{}{}
			}
		}
	}

	if minVal <= maxDist {
		for char, child := range node.Children {
			childRow := make([]int, len(query)+1)
			childRow[0] = depth + 2
			childMin := childRow[0]

			for i := 1; i <= len(query); i++ {
				cost := 1
				if query[i-1] == char {
					cost = 0
				}
				del := currentRow[i] + 1
				ins := childRow[i-1] + 1
				sub := currentRow[i-1] + cost

				if sub < del {
					if sub < ins {
						childRow[i] = sub
					} else {
						childRow[i] = ins
					}
				} else {
					if del < ins {
						childRow[i] = del
					} else {
						childRow[i] = ins
					}
				}

				if childRow[i] < childMin {
					childMin = childRow[i]
				}
			}

			if childMin <= maxDist {
				levenshteinDFSRecursive(child, query, depth+1, childRow, maxDist, result)
			}
		}
	}
}
