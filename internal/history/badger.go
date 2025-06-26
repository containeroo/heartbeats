package history

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
)

const (
	prefixEvent = "event:"
	prefixID    = "hb:"
)

// BadgerStore stores events persistently using BadgerDB.
type BadgerStore struct {
	db  *badger.DB
	mu  sync.Mutex // protects access during read aggregations
	now func() time.Time
}

// NewBadger opens or creates a Badger store at the given dir.
func NewBadger(dir string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(dir).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open badger: %w", err)
	}
	return &BadgerStore{db: db, now: time.Now}, nil
}

// RecordEvent persists the given event.
func (b *BadgerStore) RecordEvent(_ context.Context, e Event) error {
	key := fmt.Sprintf("%s%d", prefixEvent, b.now().UnixNano())

	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return b.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set([]byte(key), data); err != nil {
			return err
		}
		// Also index by heartbeat ID
		indexKey := fmt.Sprintf("%s%s:%d", prefixID, e.HeartbeatID, b.now().UnixNano())
		return txn.Set([]byte(indexKey), data)
	})
}

// GetEvents returns all events in chronological order.
func (b *BadgerStore) GetEvents() []Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	var out []Event
	_ = b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(prefixEvent)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				continue
			}
			var e Event
			if err := json.Unmarshal(val, &e); err == nil {
				out = append(out, e)
			}
		}
		return nil
	})
	slices.SortFunc(out, func(a, b Event) int {
		return a.Timestamp.Compare(b.Timestamp)
	})
	return out
}

// GetEventsByID returns all events for a specific heartbeat ID.
func (b *BadgerStore) GetEventsByID(id string) []Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	var out []Event
	_ = b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(prefixID + id + ":")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				continue
			}
			var e Event
			if err := json.Unmarshal(val, &e); err == nil {
				out = append(out, e)
			}
		}
		return nil
	})
	slices.SortFunc(out, func(a, b Event) int {
		return a.Timestamp.Compare(b.Timestamp)
	})
	return out
}

// Close closes the Badger DB.
func (b *BadgerStore) Close() error {
	return b.db.Close()
}
