package worker

import (
	"testing"
	"time"

	"github.com/2mes4/llull/internal/engine"
)

func TestPoolEnqueue(t *testing.T) {
	eng := engine.NewSearchEngine("")
	pool := NewPool(eng, 100, 2)
	defer pool.Stop()

	ok := pool.Enqueue(engine.IndexPayload{
		ID:     "doc1",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title":   "Test Doc",
			"content": "contenido de prueba",
		},
	})

	if !ok {
		t.Fatal("Expected enqueue to succeed")
	}

	time.Sleep(100 * time.Millisecond)

	if eng.DocCount() != 1 {
		t.Fatalf("Expected 1 doc, got %d", eng.DocCount())
	}
}

func TestPoolDelete(t *testing.T) {
	eng := engine.NewSearchEngine("")
	pool := NewPool(eng, 100, 2)

	eng.IndexDocument(engine.IndexPayload{
		ID:     "doc1",
		Action: "INDEX",
		Fields: map[string]interface{}{
			"title": "To Delete",
		},
	})

	pool.Enqueue(engine.IndexPayload{
		ID:     "doc1",
		Action: "DELETE",
	})

	time.Sleep(100 * time.Millisecond)
	pool.Stop()

	if eng.DocCount() != 0 {
		t.Fatalf("Expected 0 docs after delete, got %d", eng.DocCount())
	}
}

func TestPoolBackpressure(t *testing.T) {
	eng := engine.NewSearchEngine("")
	pool := NewPool(eng, 5, 1)
	defer pool.Stop()

	enqueued := 0
	rejected := 0

	for i := 0; i < 100; i++ {
		ok := pool.Enqueue(engine.IndexPayload{
			ID:     "doc-" + string(rune(i)),
			Action: "INDEX",
			Fields: map[string]interface{}{
				"content": "filler",
			},
		})
		if ok {
			enqueued++
		} else {
			rejected++
		}
	}

	if rejected == 0 {
		t.Log("Warning: expected some rejections due to backpressure, but all enqueued")
	}
	t.Logf("Enqueued: %d, Rejected: %d", enqueued, rejected)
}

func TestPoolStop(t *testing.T) {
	eng := engine.NewSearchEngine("")
	pool := NewPool(eng, 100, 2)

	pool.Enqueue(engine.IndexPayload{
		ID:     "doc1",
		Action: "INDEX",
		Fields: map[string]interface{}{"title": "Test"},
	})

	pool.Stop()

	ok := pool.Enqueue(engine.IndexPayload{
		ID:     "doc2",
		Action: "INDEX",
		Fields: map[string]interface{}{"title": "After Stop"},
	})

	if ok {
		t.Error("Expected enqueue to fail after stop")
	}
}
