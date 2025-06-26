package history

import (
	"context"

	"github.com/containeroo/heartbeats/internal/flag"
)

type BackendType string

var BackendTypeRingStore BackendType = "ring"

// Store records and exposes all system events.
type Store interface {
	RecordEvent(ctx context.Context, e Event) error // RecordEvent appends a new event to history.
	GetEvents() []Event                             // GetEvents returns a snapshot of all recorded events.
	GetEventsByID(id string) []Event                // GetEventsByID returns all events recorded for the specified heartbeat ID.
}

func InitializeHistory(flags flag.Options) (Store, error) {
	return NewRingStore(flags.HistorySize), nil
}
