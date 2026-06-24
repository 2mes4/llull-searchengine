package engine

type TrieNode struct {
	Children map[rune]*TrieNode
	DocIDs   map[string]struct{}
}

func newTrieNode() *TrieNode {
	return &TrieNode{
		Children: make(map[rune]*TrieNode),
		DocIDs:   nil,
	}
}

func (n *TrieNode) addDocID(id string) {
	if n.DocIDs == nil {
		n.DocIDs = make(map[string]struct{})
	}
	n.DocIDs[id] = struct{}{}
}

func (n *TrieNode) removeDocID(id string) {
	if n.DocIDs != nil {
		delete(n.DocIDs, id)
	}
}

func insertIntoTrie(root *TrieNode, token string, docID string) {
	current := root
	for _, char := range token {
		if _, exists := current.Children[char]; !exists {
			current.Children[char] = newTrieNode()
		}
		current = current.Children[char]
		current.addDocID(docID)
	}
}

func removeFromTrie(root *TrieNode, token string, docID string) {
	current := root
	for _, char := range token {
		child, exists := current.Children[char]
		if !exists {
			return
		}
		child.removeDocID(docID)
		current = child
	}
}

func findPrefixNode(root *TrieNode, prefix string) *TrieNode {
	current := root
	for _, char := range prefix {
		child, exists := current.Children[char]
		if !exists {
			return nil
		}
		current = child
	}
	return current
}

func collectAllDocIDs(node *TrieNode, result map[string]struct{}) {
	for id := range node.DocIDs {
		result[id] = struct{}{}
	}
	for _, child := range node.Children {
		collectAllDocIDs(child, result)
	}
}

func collectDocIDsLimit(root *TrieNode, prefix string, limit int) []string {
	node := findPrefixNode(root, prefix)
	if node == nil {
		return nil
	}
	ids := make(map[string]struct{}, limit)
	var collect func(n *TrieNode)
	collect = func(n *TrieNode) {
		if len(ids) >= limit {
			return
		}
		for id := range n.DocIDs {
			ids[id] = struct{}{}
			if len(ids) >= limit {
				return
			}
		}
		for _, child := range n.Children {
			if len(ids) >= limit {
				return
			}
			collect(child)
		}
	}
	collect(node)
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	return result
}
