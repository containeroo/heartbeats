package history

import (
	"context"
	"errors"
	"fmt"

	"github.com/containeroo/heartbeats/internal/flag"
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

func InitHistory(flags flag.Flags) (Store, error) {

	switch flags.HistoryBackend {
	case BackendTypeRingStore:
		return NewRingStore(size), nil
	case BackendTypeBadger:
		if flags.BadgerPath == "" {
			return nil, errors.New("badger backend requires a path")
		}
		return NewBadger(path)
	default:
		return nil, fmt.Errorf("unknown history backend: %s", backend)
	}
