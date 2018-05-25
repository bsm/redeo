package replication

import "io"

// Store implements an abstract data store.
type Store interface {
	// Snapshot creates a full snapshot of the store.
	Snapshot(io.Writer) error

	// Restore restores the full state from a snapshot.
	Restore(io.Reader) error
}
