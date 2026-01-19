package history

import "context"

type BackendType string

var BackendTypeRingStore BackendType = "ring"

// Store records and exposes all system events.
type Store interface {
	Append(ctx context.Context, e Event) error // Append appends a new event to history.
	List() []Event                             // List returns a snapshot of all recorded events.
	ListByID(id string) []Event                // ListByID returns all events recorded for the specified heartbeat ID.
	ByteSize() int                             // ByteSize returns the current size of the history store in bytes.
}

// InitializeHistory initializes a new history store.
func InitializeHistory(size int) (Store, error) {
	return NewRingStore(size), nil
}
