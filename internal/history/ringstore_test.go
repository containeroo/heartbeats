package history

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRingStore_GetEvents(t *testing.T) {
	t.Parallel()

	// prepare some timestamps
	e1 := Event{Timestamp: time.Unix(1, 0), HeartbeatID: "a"}
	e2 := Event{Timestamp: time.Unix(2, 0), HeartbeatID: "b"}

	// add two more to force wrap: capacity=3, we already added 2
	e3 := Event{Timestamp: time.Unix(3, 0), HeartbeatID: "c"}
	e4 := Event{Timestamp: time.Unix(4, 0), HeartbeatID: "d"}

	t.Run("empty store returns empty slice", func(t *testing.T) {
		t.Parallel()

		r := NewRingStore(4)
		got := r.List()
		assert.Len(t, got, 0)
	})

	t.Run("not full: two events in order", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		r := NewRingStore(4)

		err := r.Append(ctx, e1)
		assert.NoError(t, err)
		err = r.Append(ctx, e2)
		assert.NoError(t, err)

		want := []Event{e1, e2}
		got := r.List()
		assert.Equal(t, want, got)
	})

	t.Run("wrapped: only last 3 events in chrono order", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		r := NewRingStore(3)

		err := r.Append(ctx, e1)
		assert.NoError(t, err)
		err = r.Append(ctx, e2)
		assert.NoError(t, err)
		err = r.Append(ctx, e3)
		assert.NoError(t, err)
		err = r.Append(ctx, e4)
		assert.NoError(t, err)

		// buffer should now hold [e4,e2,e3] with next==1 → chronological: [e2,e3,e4]
		want := []Event{e2, e3, e4}
		got := r.List()
		assert.Equal(t, got, want)
	})
}

func TestRingStore_GetEventsByID(t *testing.T) {
	t.Parallel()

	t.Run("GetEventsByID before wrap", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		r := NewRingStore(3)
		e1 := Event{Timestamp: time.Unix(1, 0), HeartbeatID: "a"}
		e2 := Event{Timestamp: time.Unix(2, 0), HeartbeatID: "b"}
		e3 := Event{Timestamp: time.Unix(3, 0), HeartbeatID: "a"}

		err := r.Append(ctx, e1)
		assert.NoError(t, err)

		err = r.Append(ctx, e2)
		assert.NoError(t, err)

		err = r.Append(ctx, e3)
		assert.NoError(t, err)

		got := r.ListByID("a")
		want := []Event{e1, e3} // out of [e1,e2,e3]
		assert.Equal(t, got, want)
	})

	t.Run("GetEventsByID after wrap", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		r := NewRingStore(3)
		e1 := Event{Timestamp: time.Unix(1, 0), HeartbeatID: "a"}
		e2 := Event{Timestamp: time.Unix(2, 0), HeartbeatID: "b"}
		e3 := Event{Timestamp: time.Unix(3, 0), HeartbeatID: "a"}
		e4 := Event{Timestamp: time.Unix(4, 0), HeartbeatID: "c"}

		// record four events into a size-3 ring
		for _, e := range []Event{e1, e2, e3, e4} {
			err := r.Append(ctx, e)
			assert.NoError(t, err)
		}

		// after wrap, buffer chronological is [e2,e3,e4],
		// so only e3 for ID "a"
		got := r.ListByID("a")
		want := []Event{e3}
		assert.Equal(t, got, want)
	})
}

func TestRingStore_ByteSize(t *testing.T) {
	t.Parallel()

	size := 10_000
	store := NewRingStore(size)

	payload := RequestMetadataPayload{
		Source:    "http://localhost:9090",
		Method:    "GET",
		UserAgent: "go-test",
	}

	for range size + 5 {
		ev := MustNewEvent(EventTypeHeartbeatReceived, "test", payload)
		err := store.Append(context.Background(), ev)
		assert.NoError(t, err)
	}

	got := store.ByteSize()
	assert.Equal(t, 1810000, got)
}

func TestRingStore_ByteSizePerformance(t *testing.T) {
	t.Parallel()

	size := 10_000
	store := NewRingStore(size)

	for range size {
		ev := MustNewEvent(EventTypeHeartbeatReceived, "test", RequestMetadataPayload{})
		_ = store.Append(context.Background(), ev)
	}

	start := time.Now()
	_ = store.ByteSize()
	elapsed := time.Since(start)

	t.Logf("ByteSize calculated in %s", elapsed)
	assert.Less(t, elapsed, 100*time.Millisecond)
}
