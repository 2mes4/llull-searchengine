package engine

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"
)

type IndexManager struct {
	mu         sync.RWMutex
	indices    map[string]*managedIndex
	basePath   string
	ttl        time.Duration
	defaultIdx string
}

type managedIndex struct {
	engine   *SearchEngine
	lastUsed time.Time
	loaded   bool
}

func NewIndexManager(basePath string, ttl time.Duration, defaultIdx string) *IndexManager {
	m := &IndexManager{
		indices:    make(map[string]*managedIndex),
		basePath:   basePath,
		ttl:        ttl,
		defaultIdx: defaultIdx,
	}
	if defaultIdx == "" {
		m.defaultIdx = "default"
	}
	go m.autoUnloadLoop()
	return m
}

func (im *IndexManager) GetOrCreateIndex(name string) *SearchEngine {
	im.mu.Lock()
	defer im.mu.Unlock()

	if mi, exists := im.indices[name]; exists {
		mi.lastUsed = time.Now()
		if !mi.loaded {
			mi.engine = NewSearchEngine(name)
			dbPath := filepath.Join(im.basePath, fmt.Sprintf("llull-%s.db", name))
			mi.engine.SetPersistPath(dbPath)
			count, err := mi.engine.LoadPersistent(dbPath)
			if err == nil && count > 0 {
				log.Printf("Loaded index %q from disk (%d docs)", name, count)
			} else {
				log.Printf("Created new index %q (no persisted data)", name)
			}
			mi.loaded = true
		}
		return mi.engine
	}

	eng := NewSearchEngine(name)
	dbPath := filepath.Join(im.basePath, fmt.Sprintf("llull-%s.db", name))
	eng.SetPersistPath(dbPath)
	eng.SetIndexName(name)

	count, err := eng.LoadPersistent(dbPath)
	if err == nil && count > 0 {
		log.Printf("Loaded index %q from disk (%d docs)", name, count)
	} else {
		log.Printf("Created new index %q (no persisted data)", name)
	}

	_, err = im.tryLoadSeedForIndex(name, eng)
	if err != nil {
		log.Printf("No seed for index %q: %v", name, err)
	}

	im.indices[name] = &managedIndex{
		engine:   eng,
		lastUsed: time.Now(),
		loaded:   true,
	}
	return eng
}

func (im *IndexManager) tryLoadSeedForIndex(name string, eng *SearchEngine) (int, error) {
	return 0, nil
}

func (im *IndexManager) GetIndex(name string) *SearchEngine {
	im.mu.RLock()
	mi, exists := im.indices[name]
	im.mu.RUnlock()
	if !exists {
		return nil
	}
	im.mu.Lock()
	mi.lastUsed = time.Now()
	im.mu.Unlock()
	return mi.engine
}

func (im *IndexManager) ListIndices() []string {
	im.mu.RLock()
	defer im.mu.RUnlock()
	names := make([]string, 0, len(im.indices))
	for name := range im.indices {
		names = append(names, name)
	}
	return names
}

func (im *IndexManager) IndexInfo(name string) (docCount int64, loaded bool) {
	im.mu.RLock()
	defer im.mu.RUnlock()
	if mi, exists := im.indices[name]; exists {
		return mi.engine.DocCount(), mi.loaded
	}
	return 0, false
}

func (im *IndexManager) DefaultIndex() string {
	return im.defaultIdx
}

func (im *IndexManager) autoUnloadLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		im.unloadStaleIndices()
	}
}

func (im *IndexManager) unloadStaleIndices() {
	im.mu.Lock()
	defer im.mu.Unlock()

	now := time.Now()
	for name, mi := range im.indices {
		if name == im.defaultIdx {
			continue
		}
		if !mi.loaded {
			continue
		}
		if now.Sub(mi.lastUsed) > im.ttl {
			log.Printf("Unloading index %q (idle for %v)", name, now.Sub(mi.lastUsed).Round(time.Second))
			if mi.engine.Dirty() {
				if err := mi.engine.SavePersistent(); err != nil {
					log.Printf("Error saving index %q before unload: %v", name, err)
				}
			}
			mi.engine = nil
			mi.loaded = false
		}
	}
}

func (im *IndexManager) SaveAll() {
	im.mu.RLock()
	defer im.mu.RUnlock()
	for name, mi := range im.indices {
		if mi.loaded && mi.engine.Dirty() {
			if err := mi.engine.SavePersistent(); err != nil {
				log.Printf("Error saving index %q: %v", name, err)
			} else {
				mi.engine.MarkClean()
			}
		}
	}
}

func (im *IndexManager) Stop() {
	im.SaveAll()
}
