package engine

import (
	"testing"
)

func TestTokenizeBasic(t *testing.T) {
	tokens := tokenize("Caballero de la Insula Firme")
	expected := []string{"caballero", "de", "la", "insula", "firme"}
	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i, exp := range expected {
		if tokens[i] != exp {
			t.Errorf("Token %d: expected %q, got %q", i, exp, tokens[i])
		}
	}
}

func TestTokenizePunctuation(t *testing.T) {
	tokens := tokenize("El caballero, (el valiente) fue a la cueva.")
	for _, token := range tokens {
		for _, ch := range token {
			if ch == ',' || ch == '(' || ch == ')' || ch == '.' {
				t.Errorf("Token %q contains punctuation", token)
			}
		}
	}
}

func TestTokenizeDiacritics(t *testing.T) {
	tokens := tokenize("España Ínsula Río")
	expected := []string{"espana", "insula", "rio"}
	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i, exp := range expected {
		if tokens[i] != exp {
			t.Errorf("Token %d: expected %q, got %q", i, exp, tokens[i])
		}
	}
}

func TestTokenizeEmpty(t *testing.T) {
	tokens := tokenize("")
	if len(tokens) != 0 {
		t.Fatalf("Expected 0 tokens for empty string, got %d", len(tokens))
	}
}

func TestTokenizeNumbers(t *testing.T) {
	tokens := tokenize("Capitulo 150 del ano 1545")
	if len(tokens) != 5 {
		t.Fatalf("Expected 5 tokens, got %d: %v", len(tokens), tokens)
	}
}
