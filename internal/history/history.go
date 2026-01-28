package history

import (
	"sync"
	"time"
)

// Event represents a recorded history item.
type Event struct {
	Time        time.Time      `json:"timestamp"`             // Event timestamp.
	Type        string         `json:"type"`                  // Event type identifier.
	HeartbeatID string         `json:"heartbeatId,omitempty"` // Heartbeat id.
	Receiver    string         `json:"receiver,omitempty"`    // Receiver name.
	TargetType  string         `json:"targetType,omitempty"`  // Delivery target type.
	Status      string         `json:"status,omitempty"`      // Status value.
	Message     string         `json:"message,omitempty"`     // Human-readable message.
	Fields      map[string]any `json:"fields,omitempty"`      // Optional fields.
}

// Recorder records history events.
type Recorder interface {
	Add(Event)
	List() []Event
	ListByID(string) []Event
}

// Streamer provides event subscriptions.
type Streamer interface {
	Subscribe(buffer int) (<-chan Event, func())
}

// Store keeps a fixed-size history buffer.
type Store struct {
	mu     sync.RWMutex
	size   int
	events []Event
	next   int
	full   bool
}

// NewStore constructs a Store with a fixed size.
func NewStore(size int) *Store {
	size = max(size, 1)
	return &Store{
		size:   size,
		events: make([]Event, 0, size),
	}
}

// Add records a new event.
func (s *Store) Add(e Event) {
	if e.Time.IsZero() {
		e.Time = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.events) < s.size {
		s.events = append(s.events, e)
		if len(s.events) == s.size {
			s.full = true
		}
		return
	}
	s.events[s.next] = e
	s.next = (s.next + 1) % s.size
	s.full = true
}

// List returns a snapshot of events in chronological order.
func (s *Store) List() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.events) == 0 {
		return nil
	}
	if !s.full {
		out := make([]Event, len(s.events))
		copy(out, s.events)
		return out
	}
	out := make([]Event, 0, len(s.events))
	out = append(out, s.events[s.next:]...)
	out = append(out, s.events[:s.next]...)
	return out
}

// ListByID returns a snapshot of events in chronological order filtered by heartbeat id.
func (s *Store) ListByID(heartbeatID string) []Event {
	items := s.List()
	out := make([]Event, 0, len(items))
	for _, item := range items {
		if item.HeartbeatID == heartbeatID {
			out = append(out, item)
		}
	}
	return out
}
