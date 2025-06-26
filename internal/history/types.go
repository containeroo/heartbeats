package history

import (
	"context"
	"errors"
	"fmt"
)

type BackendType string

var (
	BackendTypeRingStore BackendType = "ring"
	BackendTypeBadger    BackendType = "badger"
)

// Store records and exposes all system events.
type Store interface {
	RecordEvent(ctx context.Context, e Event) error // RecordEvent appends a new event to history.
	GetEvents() []Event                             // GetEvents returns a snapshot of all recorded events.
	GetEventsByID(id string) []Event                // GetEventsByID returns all events recorded for the specified heartbeat ID.
}

func InitializeHistory(backend BackendType, size int, path string) (Store, error) {
	switch backend {
	case BackendTypeRingStore:
		return NewRingStore(size), nil
	case BackendTypeBadger:
		if path == "" {
			return nil, errors.New("badger backend requires a path")
		}
		return NewBadger(path)
	default:
		return nil, fmt.Errorf("unknown history backend: %s", backend)
	}
}
