package history

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestRingStore_GetEvents(t *testing.T) {
	ctx := context.Background()
	r := NewRingStore(3)

	t.Run("empty store returns empty slice", func(t *testing.T) {
		got := r.GetEvents()
		if len(got) != 0 {
			t.Errorf("GetEvents() = %v, want []", got)
		}
	})

	// prepare some timestamps
	e1 := Event{Timestamp: time.Unix(1, 0), HeartbeatID: "a"}
	e2 := Event{Timestamp: time.Unix(2, 0), HeartbeatID: "b"}

	t.Run("not full: two events in order", func(t *testing.T) {
		if err := r.RecordEvent(ctx, e1); err != nil {
			t.Fatalf("RecordEvent: %v", err)
		}
		if err := r.RecordEvent(ctx, e2); err != nil {
			t.Fatalf("RecordEvent: %v", err)
		}
		want := []Event{e1, e2}
		got := r.GetEvents()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetEvents() = %v, want %v", got, want)
		}
	})

	// add two more to force wrap: capacity=3, we already added 2
	e3 := Event{Timestamp: time.Unix(3, 0), HeartbeatID: "c"}
	e4 := Event{Timestamp: time.Unix(4, 0), HeartbeatID: "d"}

	t.Run("wrapped: only last 3 events in chrono order", func(t *testing.T) {
		if err := r.RecordEvent(ctx, e3); err != nil {
			t.Fatalf("RecordEvent: %v", err)
		}
		if err := r.RecordEvent(ctx, e4); err != nil {
			t.Fatalf("RecordEvent: %v", err)
		}
		// buffer should now hold [e4,e2,e3] with next==1 → chronological: [e2,e3,e4]
		want := []Event{e2, e3, e4}
		got := r.GetEvents()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetEvents() after wrap = %v, want %v", got, want)
		}
	})
}

func TestRingStore_GetEventsByID(t *testing.T) {
	ctx := context.Background()

	// First test unwrapped behavior
	t.Run("GetEventsByID before wrap", func(t *testing.T) {
		r := NewRingStore(3)
		e1 := Event{Timestamp: time.Unix(1, 0), HeartbeatID: "a"}
		e2 := Event{Timestamp: time.Unix(2, 0), HeartbeatID: "b"}
		e3 := Event{Timestamp: time.Unix(3, 0), HeartbeatID: "a"}

		if err := r.RecordEvent(ctx, e1); err != nil {
			t.Fatalf("RecordEvent: %v", err)
		}
		if err := r.RecordEvent(ctx, e2); err != nil {
			t.Fatalf("RecordEvent: %v", err)
		}
		if err := r.RecordEvent(ctx, e3); err != nil {
			t.Fatalf("RecordEvent: %v", err)
		}

		got := r.GetEventsByID("a")
		want := []Event{e1, e3} // out of [e1,e2,e3]
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetEventsByID(\"a\") = %v, want %v", got, want)
		}
	})

	// Now test wrapped behavior
	t.Run("GetEventsByID after wrap", func(t *testing.T) {
		r := NewRingStore(3)
		e1 := Event{Timestamp: time.Unix(1, 0), HeartbeatID: "a"}
		e2 := Event{Timestamp: time.Unix(2, 0), HeartbeatID: "b"}
		e3 := Event{Timestamp: time.Unix(3, 0), HeartbeatID: "a"}
		e4 := Event{Timestamp: time.Unix(4, 0), HeartbeatID: "c"}

		// record four events into a size-3 ring
		for _, e := range []Event{e1, e2, e3, e4} {
			if err := r.RecordEvent(ctx, e); err != nil {
				t.Fatalf("RecordEvent: %v", err)
			}
		}

		// after wrap, buffer chronological is [e2,e3,e4],
		// so only e3 for ID "a"
		got := r.GetEventsByID("a")
		want := []Event{e3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetEventsByID(\"a\") after wrap = %v, want %v", got, want)
		}
	})
}
