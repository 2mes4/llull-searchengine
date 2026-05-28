package engine

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

func (se *SearchEngine) SavePersistent() error {
	if se.persistPath == "" {
		return nil
	}

	db, err := bolt.Open(se.persistPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return fmt.Errorf("open db for save: %w", err)
	}
	defer db.Close()

	se.mu.RLock()
	defer se.mu.RUnlock()

	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("documents"))
		if err != nil {
			return err
		}

		for id, meta := range se.Metadata {
			data, err := json.Marshal(meta)
			if err != nil {
				return fmt.Errorf("marshal %s: %w", id, err)
			}
			if err := bucket.Put([]byte(id), data); err != nil {
				return err
			}
		}

		count := int64(len(se.Metadata))
		countData, _ := json.Marshal(count)
		if err := bucket.Put([]byte("__count__"), countData); err != nil {
			return err
		}

		return nil
	})
}

func (se *SearchEngine) LoadPersistent(path string) (int, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{ReadOnly: true, Timeout: 1 * time.Second})
	if err != nil {
		return 0, fmt.Errorf("open db for load: %w", err)
	}
	defer db.Close()

	var count int
	var metadata map[string]DocMetadata

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("documents"))
		if bucket == nil {
			return fmt.Errorf("no documents bucket found")
		}

		metadata = make(map[string]DocMetadata)

		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if string(k) == "__count__" {
				continue
			}
			var meta DocMetadata
			if err := json.Unmarshal(v, &meta); err != nil {
				return fmt.Errorf("unmarshal %s: %w", string(k), err)
			}
			metadata[string(k)] = meta
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	se.mu.Lock()
	defer se.mu.Unlock()

	se.Metadata = metadata
	se.docCount = int64(len(metadata))

	for id, meta := range metadata {
		for _, value := range meta.Fields {
			if str, ok := value.(string); ok && str != "" {
				tokens := tokenize(str)
				seen := make(map[string]struct{})
				for _, token := range tokens {
					if _, dup := seen[token]; dup {
						continue
					}
					seen[token] = struct{}{}
					insertIntoTrie(se.Root, token, id)
				}
			}
		}
	}

	count = len(metadata)
	return count, nil
}
