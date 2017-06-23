package info

import "bytes"

// Main info registry.
// Please note: in order to minimise performance impact info registries
// are not using locks are therefore not thread-safe. Please make sure
// you register all metrics and values before you start the server.
type Registry struct{ sections []*Section }

// New creates a new Registry
func New() *Registry {
	return new(Registry)
}

// Section returns a section, or appends a new one
// when the given name cannot be found
func (r *Registry) Section(name string) *Section {
	for _, s := range r.sections {
		if s.name == name {
			return s
		}
	}
	section := &Section{name: name}
	r.sections = append(r.sections, section)
	return section
}

// Clear removes all sections from the registry
func (r *Registry) Clear() {
	r.sections = nil
}

// String generates an info string output
func (r *Registry) String() string {
	buf := new(bytes.Buffer)
	for i, section := range r.sections {
		if len(section.kvs) == 0 {
			continue
		}

		if i != 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString("# " + section.name + "\n")
		section.writeTo(buf)
	}
	return buf.String()
}

// An info section contains multiple values
type Section struct {
	name string
	kvs  []kv
}

// Register registers a value under a name
func (s *Section) Register(name string, value Value) {
	s.kvs = append(s.kvs, kv{name, value})
}

// Clear removes all values from a section
func (s *Section) Clear() {
	s.kvs = nil
}

func (s *Section) writeTo(buf *bytes.Buffer) {
	for _, kv := range s.kvs {
		buf.WriteString(kv.name + ":" + kv.value.String() + "\n")
	}
}

type kv struct {
	name  string
	value Value
}
