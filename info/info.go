package info

import (
	"bytes"
	"strings"
	"sync"
)

// Registry : main info registry.
// Please note: in order to minimise performance impact info registries
// are not using locks are therefore not thread-safe. Please make sure
// you register all metrics and values before you start the server.
type Registry struct {
	sections []*Section
	mu       sync.RWMutex
}

// New creates a new Registry
func New() *Registry {
	return new(Registry)
}

// FindSection returns a section by name or nil when not found.
func (r *Registry) FindSection(name string) *Section {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.findSection(name)
}

// FetchSection returns a section, or appends a new one
// when the given name cannot be found
func (r *Registry) FetchSection(name string) *Section {
	if s := r.FindSection(name); s != nil {
		return s
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	s := r.findSection(name)
	if s == nil {
		s = &Section{name: name}
		r.sections = append(r.sections, s)
	}
	return s
}

// Clear removes all sections from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	r.sections = nil
	r.mu.Unlock()
}

// String generates an info string output
func (r *Registry) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	buf := new(bytes.Buffer)
	for i, s := range r.sections {
		if len(s.kvs) == 0 {
			continue
		}

		if i != 0 {
			buf.WriteByte('\n')
		}
		s.writeTo(buf)
	}
	return buf.String()
}

func (r *Registry) findSection(name string) *Section {
	for _, s := range r.sections {
		if strings.ToLower(s.name) == strings.ToLower(name) {
			return s
		}
	}
	return nil
}

// Section : an info section contains multiple values
type Section struct {
	name string
	kvs  []kv
	mu   sync.RWMutex
}

// Register registers a value under a name
func (s *Section) Register(name string, value Value) {
	s.mu.Lock()
	s.kvs = append(s.kvs, kv{name, value})
	s.mu.Unlock()
}

// Clear removes all values from a section
func (s *Section) Clear() {
	s.mu.Lock()
	s.kvs = nil
	s.mu.Unlock()
}

// Replace replaces the section enties
func (s *Section) Replace(fn func(*Section)) {
	t := &Section{name: s.name}
	fn(t)

	s.mu.Lock()
	s.kvs = t.kvs
	s.mu.Unlock()
}

func (s *Section) writeTo(buf *bytes.Buffer) {
	buf.WriteString("# " + s.name + "\n")
	for _, kv := range s.kvs {
		buf.WriteString(kv.name + ":" + kv.value.String() + "\n")
	}
}

// String generates an info string output
func (s *Section) String() string {
	if s == nil {
		return ""
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.kvs) == 0 {
		return ""
	}

	buf := new(bytes.Buffer)
	s.writeTo(buf)
	return buf.String()
}

type kv struct {
	name  string
	value Value
}
