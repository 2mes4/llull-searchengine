package engine

type TrieNode struct {
	Children map[rune]*TrieNode
	DocIDs   []string
}

func newTrieNode() *TrieNode {
	return &TrieNode{
		Children: make(map[rune]*TrieNode),
		DocIDs:   nil,
	}
}

func (n *TrieNode) addDocID(id string) {
	for _, existing := range n.DocIDs {
		if existing == id {
			return
		}
	}
	n.DocIDs = append(n.DocIDs, id)
}

func (n *TrieNode) removeDocID(id string) {
	filtered := n.DocIDs[:0]
	for _, existing := range n.DocIDs {
		if existing != id {
			filtered = append(filtered, existing)
		}
	}
	n.DocIDs = filtered
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
	for _, id := range node.DocIDs {
		result[id] = struct{}{}
	}
	for _, child := range node.Children {
		collectAllDocIDs(child, result)
	}
}
