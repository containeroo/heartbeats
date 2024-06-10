package history

import (
	"fmt"
	"sync"
	"time"
)

// Event type represents various events that can be logged in history.
type Event int16

const (
	Beat Event = iota
	Interval
	Grace
	Expired
	Send
)

func (e Event) String() string {
	return [...]string{"BEAT", "INTERVAL", "GRACE", "EXPIRED", "SEND"}[e]
}

// HistoryEntry represents a single entry in the history log.
type HistoryEntry struct {
	Time    time.Time
	Event   Event
	Message string
	Details map[string]string
}

// History maintains a history of events with a maximum size.
type History struct {
	mu          sync.Mutex
	entries     []HistoryEntry
	maxSize     int
	reduceRatio int
}

// NewHistory creates a new History instance.
func NewHistory(maxSize, reduceRatio int) *History {
	return &History{
		entries:     make([]HistoryEntry, 0, maxSize),
		maxSize:     maxSize,
		reduceRatio: reduceRatio,
	}
}

// AddEntry adds a new entry to the history.
func (h *History) AddEntry(event Event, message string, details map[string]string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = append(h.entries, HistoryEntry{
		Time:    time.Now(),
		Event:   event,
		Message: message,
		Details: details,
	})

	if len(h.entries) > h.maxSize {
		reduceTo := int(h.maxSize * h.reduceRatio)
		h.entries = h.entries[len(h.entries)-reduceTo:]
	}
}

// GetAllEntries returns all history entries.
func (h *History) GetAllEntries() []HistoryEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.entries
}

// Store manages histories for multiple heartbeats.
type Store struct {
	mu        sync.RWMutex
	histories map[string]*History
}

// MarshalYAML implements the yaml.Marshaler interface.
func (s *Store) MarshalYAML() (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.histories, nil
}

// NewStore creates a new HistoryStore.
func NewStore() *Store {
	return &Store{
		histories: make(map[string]*History),
	}
}

// Get retrieves the history for a given heartbeat.
func (s *Store) Get(name string) *History {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.histories[name]
}

// Add adds a history entry for a given heartbeat.
func (s *Store) Add(name string, history *History) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if exists := s.histories[name]; exists != nil {
		return fmt.Errorf("history '%s' already exists", name)
	}

	s.histories[name] = history

	return nil
}

// Delete removes the history for a given heartbeat.
func (s *Store) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.histories, name)
}
