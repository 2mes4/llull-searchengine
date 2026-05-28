package seed

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTextFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("Hello world from Llull"), 0644)
	os.WriteFile(filepath.Join(dir, "skip.pdf"), []byte("binary"), 0644)

	texts, names, err := LoadTextFiles(dir)
	if err != nil {
		t.Fatalf("LoadTextFiles failed: %v", err)
	}
	if len(texts) != 1 {
		t.Fatalf("Expected 1 text file, got %d", len(texts))
	}
	if texts[0] != "Hello world from Llull" {
		t.Errorf("Unexpected text: %s", texts[0])
	}
	if len(names) != 1 || names[0] != "test" {
		t.Errorf("Unexpected names: %v", names)
	}
}

func TestGenerateDocumentsFromTexts(t *testing.T) {
	texts := []string{
		strings.Repeat("Ramon Llull fue un filosofo y místico catalan. ", 50),
		strings.Repeat("El arbol de la ciencia es una obra fundamental. ", 50),
	}
	names := []string{"obra-1", "obra-2"}

	docs := GenerateDocumentsFromTexts(texts, names, 100)
	if len(docs) == 0 {
		t.Fatal("Expected documents, got 0")
	}

	for i, doc := range docs {
		if doc.ID == "" {
			t.Errorf("Doc %d: missing ID", i)
		}
		if doc.Fields == nil {
			t.Errorf("Doc %d: missing fields", i)
			continue
		}
		if content, ok := doc.Fields["content"]; !ok || content == "" {
			t.Errorf("Doc %d: missing content", i)
		}
		if _, ok := doc.Fields["weight"]; !ok {
			t.Errorf("Doc %d: missing weight", i)
		}
	}
}

func TestGenerateSeedFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "book.txt"), []byte("Contenido del libro de Ramon Llull"), 0644)
	texts, names, _ := LoadTextFiles(dir)

	path := filepath.Join(t.TempDir(), "seed.json")
	err := GenerateSeedFile(path, texts, names, 10)
	if err != nil {
		t.Fatalf("GenerateSeedFile: %v", err)
	}

	docs, err := LoadSeedFile(path)
	if err != nil {
		t.Fatalf("LoadSeedFile: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("Expected 1 doc, got %d", len(docs))
	}
}

func TestToIndexPayloads(t *testing.T) {
	docs := []BookDocument{
		{ID: "doc-0", Fields: map[string]interface{}{"content": "test"}},
	}
	payloads := ToIndexPayloads(docs)
	if len(payloads) != 1 {
		t.Fatalf("Expected 1 payload, got %d", len(payloads))
	}
	if payloads[0].Action != "INDEX" {
		t.Errorf("Expected INDEX, got %s", payloads[0].Action)
	}
}

func TestSplitIntoChunks(t *testing.T) {
	text := "Paragraph one.\n\nParagraph two.\n\nParagraph three."
	chunks := splitIntoChunks(text, 30)
	if len(chunks) < 2 {
		t.Fatalf("Expected at least 2 chunks, got %d", len(chunks))
	}
}
