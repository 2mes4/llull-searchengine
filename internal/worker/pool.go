package worker

import (
	"context"
	"runtime"
	"sync"

	"github.com/2mes4/llull/internal/engine"
)

type Pool struct {
	queue   chan engine.IndexPayload
	engine  *engine.SearchEngine
	wg      sync.WaitGroup
	cancel  context.CancelFunc
	stopped bool
	mu      sync.Mutex
}

func NewPool(eng *engine.SearchEngine, bufferSize int, numWorkers int) *Pool {
	if bufferSize <= 0 {
		bufferSize = 5000
	}
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	ctx, cancel := context.WithCancel(context.Background())

	p := &Pool{
		queue:  make(chan engine.IndexPayload, bufferSize),
		engine: eng,
		cancel: cancel,
	}

	p.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go p.worker(ctx)
	}

	return p
}

func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case payload, ok := <-p.queue:
			if !ok {
				return
			}
			p.process(payload)
		}
	}
}

func (p *Pool) process(payload engine.IndexPayload) {
	switch payload.Action {
	case "INDEX":
		p.engine.IndexDocument(payload)
	case "DELETE":
		p.engine.DeleteDocument(payload.ID)
	}
}

func (p *Pool) Enqueue(payload engine.IndexPayload) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stopped {
		return false
	}

	select {
	case p.queue <- payload:
		return true
	default:
		return false
	}
}

func (p *Pool) QueueLen() int {
	return len(p.queue)
}

func (p *Pool) Stop() {
	p.mu.Lock()
	p.stopped = true
	p.mu.Unlock()

	p.cancel()
	p.wg.Wait()
}

func (p *Pool) Drain() {
	p.Stop()
}
