package engine

import (
	"testing"
)

func TestTrieInsertAndSearch(t *testing.T) {
	root := newTrieNode()
	insertIntoTrie(root, "caballero", "doc1")
	insertIntoTrie(root, "caballeria", "doc2")
	insertIntoTrie(root, "caballo", "doc3")

	node := findPrefixNode(root, "cab")
	if node == nil {
		t.Fatal("Expected to find prefix 'cab'")
	}

	ids := make(map[string]struct{})
	collectAllDocIDs(node, ids)

	if len(ids) != 3 {
		t.Fatalf("Expected 3 doc IDs, got %d", len(ids))
	}

	for _, expected := range []string{"doc1", "doc2", "doc3"} {
		if _, ok := ids[expected]; !ok {
			t.Errorf("Missing doc ID: %s", expected)
		}
	}
}

func TestTriePrefixSearchReturnsOnlyMatching(t *testing.T) {
	root := newTrieNode()
	insertIntoTrie(root, "espada", "doc1")
	insertIntoTrie(root, "escudo", "doc2")
	insertIntoTrie(root, "espejo", "doc3")

	ids := searchByPrefix(root, "esp")
	if len(ids) != 2 {
		t.Fatalf("Expected 2 results for 'esp', got %d", len(ids))
	}

	set := make(map[string]struct{})
	for _, id := range ids {
		set[id] = struct{}{}
	}
	if _, ok := set["doc1"]; !ok {
		t.Error("Expected doc1 (espada)")
	}
	if _, ok := set["doc3"]; !ok {
		t.Error("Expected doc3 (espejo)")
	}
}

func TestTrieNoMatch(t *testing.T) {
	root := newTrieNode()
	insertIntoTrie(root, "caballero", "doc1")

	ids := searchByPrefix(root, "zorro")
	if ids != nil {
		t.Fatalf("Expected nil for no match, got %v", ids)
	}
}

func TestTrieRemoveDocID(t *testing.T) {
	root := newTrieNode()
	insertIntoTrie(root, "espada", "doc1")
	insertIntoTrie(root, "espada", "doc2")
	removeFromTrie(root, "espada", "doc1")

	ids := searchByPrefix(root, "espada")
	if len(ids) != 1 {
		t.Fatalf("Expected 1 result after removal, got %d", len(ids))
	}
	if ids[0] != "doc2" {
		t.Errorf("Expected doc2, got %s", ids[0])
	}
}

func TestTrieDuplicateDocID(t *testing.T) {
	root := newTrieNode()
	insertIntoTrie(root, "prueba", "doc1")
	insertIntoTrie(root, "prueba", "doc1")

	node := findPrefixNode(root, "prueba")
	if len(node.DocIDs) != 1 {
		t.Fatalf("Expected 1 doc ID (deduped), got %d", len(node.DocIDs))
	}
}

func TestTrieMultiTokenSearch(t *testing.T) {
	root := newTrieNode()
	insertIntoTrie(root, "caballero", "doc1")
	insertIntoTrie(root, "espada", "doc1")
	insertIntoTrie(root, "caballero", "doc2")

	ids := searchMultiToken(root, []string{"caballero", "espada"})
	if len(ids) != 1 {
		t.Fatalf("Expected 1 result for multi-token, got %d", len(ids))
	}
	if ids[0] != "doc1" {
		t.Errorf("Expected doc1, got %s", ids[0])
	}
}
