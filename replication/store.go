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

// InMemKeyValueStore is the most simplistic implementation of a DataStore.
// Please only use for tests.
type InMemKeyValueStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewInMemKeyValueStore creates a new InMemKeyValueStore
func NewInMemKeyValueStore() *InMemKeyValueStore {
	return &InMemKeyValueStore{
		data: make(map[string][]byte),
	}
}

// Get retrieves a value by key.
func (s *InMemKeyValueStore) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	val, ok := s.data[key]
	s.mu.RUnlock()
	return val, ok
}

// Set stores a value for a key.
func (s *InMemKeyValueStore) Set(key string, value []byte) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
}

// Keys returns all keys
func (s *InMemKeyValueStore) Keys() []string {
	s.mu.RLock()
	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}
	s.mu.RUnlock()
	return keys
}

// Snapshot implements DataStore interface.
func (s *InMemKeyValueStore) Snapshot(w io.Writer) error {
	enc := json.NewEncoder(w)

	s.mu.RLock()
	err := enc.Encode(s.data)
	s.mu.RUnlock()
	return err
}

// Restore implements DataStore interface.
func (s *InMemKeyValueStore) Restore(r io.Reader) error {
	var data map[string][]byte
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return err
	}

	s.mu.Lock()
	s.data = data
	s.mu.Unlock()
	return nil
}

// --------------------------------------------------------------------

var ErrNotFound = errors.New("not found")

// StableStore is used for configuration purposes
type StableStore interface {
	// Decode reads a key into a target value.
	// It returns ErrNotFound when no such key is stored.
	Decode(key string, value interface{}) error
	// Encode stores a value under a given key.
	Encode(key string, value interface{}) error
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

// Decode implements the StableStore interface
func (s *inMemStableStore) Decode(key string, value interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.data[key]
	if !ok {
		return ErrNotFound
	}
	return json.Unmarshal(data, value)
}

// Encode implements the StableStore interface
func (s *inMemStableStore) Encode(key string, value interface{}) error {
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

// Decode implements the StableStore interface
func (s *fsStableStore) Decode(key string, value interface{}) error {
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

// Encode implements the StableStore interface
func (s *fsStableStore) Encode(key string, value interface{}) error {
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
