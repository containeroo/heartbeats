package history

import (
	"fmt"
	"sync"
	"time"
)

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
	reduceRatio int // percentage reduction, e.g., 20 for 20%
}

// NewHistory creates a new History instance.
func NewHistory(maxSize, reduceRatio int) (*History, error) {
	if reduceRatio > 100 {
		return nil, fmt.Errorf("reduce is a percentage and cannot be more 100.")
	}

	return &History{
		entries:     make([]HistoryEntry, 0, maxSize),
		maxSize:     maxSize,
		reduceRatio: reduceRatio,
	}, nil
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
		h.reduceSize()
	}
}

// reduceSize reduces the number of entries in the history by the specified percentage.
func (h *History) reduceSize() {
	reduceTo := h.maxSize - (h.maxSize * h.reduceRatio / 100)
	if reduceTo < 0 {
		reduceTo = 0
	}
	if len(h.entries) > reduceTo {
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
