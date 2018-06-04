package replication

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DataStore implements an abstract data store.
type DataStore interface {
	// Snapshot creates a full snapshot of the store.
	Snapshot(io.Writer) error

	// Restore restores the full state from a snapshot.
	Restore(io.Reader) error
}

// --------------------------------------------------------------------

var ErrNotFound = errors.New("not found")

// StableStore is used for configuration purposes
type StableStore interface {
	// Get reads a key into a target value.
	// It returns ErrNotFound when no such key is stored.
	Get(key string, value interface{}) error
	// Set stores a value under a given key.
	Set(key string, value interface{}) error
}

type inMemStableStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewInMemStableStore returns a new in-memory stable store.
// Please only use this for testing.
func NewInMemStableStore() StableStore {
	return &inMemStableStore{data: make(map[string][]byte)}
}

// Get implements the StableStore interface
func (s *inMemStableStore) Get(key string, value interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.data[key]
	if !ok {
		return ErrNotFound
	}
	return json.Unmarshal(data, value)
}

// Set implements the StableStore interface
func (s *inMemStableStore) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.data[key] = data
	s.mu.Unlock()

	return nil
}

type fsStableStore struct {
	dir string
}

// NewFSStableStore returns a new in-memory stable store.
// Please only use this for testing.
func NewFSStableStore(dir string) StableStore {
	entries, _ := filepath.Glob(filepath.Join(dir, ".*.~t*"))
	for _, tn := range entries {
		_ = os.Remove(tn)
	}
	return &fsStableStore{dir: dir}
}

// Get implements the StableStore interface
func (s *fsStableStore) Get(key string, value interface{}) error {
	f, err := os.Open(filepath.Join(s.dir, key))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(value)
}

// Set implements the StableStore interface
func (s *fsStableStore) Set(key string, value interface{}) error {
	tn := filepath.Join(s.dir, fmt.Sprintf(".%s.~t%d", key, time.Now().UnixNano()))
	defer os.Remove(tn)

	t, err := os.Create(tn)
	if err != nil {
		return err
	}
	defer t.Close()

	enc := json.NewEncoder(t)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return err
	}

	if err := t.Close(); err != nil {
		return err
	}
	return os.Rename(tn, filepath.Join(s.dir, key))
}
