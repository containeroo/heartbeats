package history

import (
	"context"
	"slices"
	"sync"
	"unsafe"
)

// RingStore is a fixed-size circular buffer of history.Event.
// When full, new events overwrite the oldest one.
type RingStore struct {
	mu   sync.Mutex
	buf  []Event // underlying storage
	next int     // next write index
	full bool    // have we wrapped at least once?
	max  int     // capacity of the buffer
}

// NewRingStore returns a RingStore that holds at most maxEvents.
func NewRingStore(maxEvents int) *RingStore {
	return &RingStore{
		buf: make([]Event, maxEvents),
		max: maxEvents,
	}
}

// ByteSize returns an estimated size of the ring buffer in bytes.
func (r *RingStore) ByteSize() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	var total int
	for _, e := range r.buf {
		total += int(unsafe.Sizeof(e)) // size of struct (fixed-size fields)
		total += len(e.HeartbeatID)    // string data
		total += len(e.RawPayload)     // payload slice
		total += len(e.Type)           // EventType is a string alias
	}
	return total
}

// Append appends a new event into the ring.
func (r *RingStore) Append(_ context.Context, e Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Overwrite at 'next' slot
	r.buf[r.next] = e

	// Advance pointer, wrap & mark full if needed
	r.next++
	if r.next >= r.max {
		r.next = 0
		r.full = true
	}
	return nil
}

// List returns all events oldest-first.
// If the buffer has never wrapped, that's just buf[:next].
// Once wrapped, we stitch buf[next:] before buf[:next].
func (r *RingStore) List() []Event {
	r.mu.Lock()
	defer r.mu.Unlock()

	// not yet wrapped → simple slice up to next
	if !r.full {
		return slices.Clone(r.buf[:r.next])
	}

	// wrapped → combine [next:] + [:next]
	head := slices.Clone(r.buf[r.next:]) // from oldest
	tail := slices.Clone(r.buf[:r.next]) // through most recent
	return append(head, tail...)
}

// ListByID returns only those events for the given heartbeat ID,
// still in chronological (oldest-first) order.
func (r *RingStore) ListByID(id string) []Event {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Build a chronological view in seq:
	// start with all entries up to next
	seq := r.buf[:r.next]
	// if wrapped, prepend the overflow slice
	if r.full {
		seq = append(r.buf[r.next:], seq...)
	}

	// Now filter by HeartbeatID
	var filtered []Event
	for _, e := range seq {
		if e.HeartbeatID == id {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
