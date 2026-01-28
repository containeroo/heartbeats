package dispatch

import (
	"sync"

	"github.com/containeroo/heartbeats/internal/notify/types"
)

// Store keeps notifications by id.
type Store struct {
	mu    sync.RWMutex
	items map[string]types.Notification
}

// NewStore initializes an empty notification store.
func NewStore() *Store {
	return &Store{
		items: make(map[string]types.Notification),
	}
}

// Put stores a notification by its id.
func (s *Store) Put(id string, n types.Notification) {
	if n == nil || id == "" {
		return
	}
	s.mu.Lock()
	s.items[id] = n
	s.mu.Unlock()
}

// Get returns a notification by id.
func (s *Store) Get(id string) (types.Notification, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.items[id]
	return n, ok
}

// Delete removes a notification by id.
func (s *Store) Delete(id string) {
	s.mu.Lock()
	delete(s.items, id)
	s.mu.Unlock()
}
