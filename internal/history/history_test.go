package history

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreListAndListByID(t *testing.T) {
	store := NewStore(3)
	store.Add(Event{Type: "first", HeartbeatID: "hb-1"})
	store.Add(Event{Type: "second", HeartbeatID: "hb-2"})
	store.Add(Event{Type: "third", HeartbeatID: "hb-1"})

	t.Run("List returns chronological snapshot", func(t *testing.T) {
		items := store.List()
		require.Len(t, items, 3)
		require.Equal(t, "first", items[0].Type)
		require.Equal(t, "third", items[2].Type)
	})

	t.Run("ListByID filters events", func(t *testing.T) {
		items := store.ListByID("hb-1")
		require.Len(t, items, 2)
		assert.Equal(t, "hb-1", items[0].HeartbeatID)
		assert.Equal(t, "hb-1", items[1].HeartbeatID)
	})

	t.Run("Overflow retains newest events", func(t *testing.T) {
		store.Add(Event{Type: "fourth", HeartbeatID: "hb-3"})
		items := store.List()

		require.Len(t, items, 3)
		assert.Equal(t, "second", items[0].Type)
		assert.Equal(t, "fourth", items[2].Type)
	})
}

func TestAsyncRecorderAddsAndBroadcasts(t *testing.T) {
	store := NewStore(10)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	rec := NewAsyncRecorder(store, logger, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rec.Start(ctx)

	sub, cancelSub := rec.Subscribe(2)
	defer cancelSub()

	rec.Add(Event{Type: "notify", HeartbeatID: "hb"})
	rec.Add(Event{Type: "another", HeartbeatID: "hb"})

	waitForEvent := func(expected string) {
		select {
		case ev := <-sub:
			assert.Equal(t, expected, ev.Type)
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for %s", expected)
		}
	}

	waitForEvent("notify")
	waitForEvent("another")

	events := waitForStored(t, store, "hb", 2)
	require.Len(t, events, 2)
}

func waitForStored(t *testing.T, store *Store, hb string, expect int) []Event {
	t.Helper()
	var events []Event
	for i := 0; i < 10; i++ {
		events = store.ListByID(hb)
		if len(events) == expect {
			return events
		}
		time.Sleep(20 * time.Millisecond)
	}
	return events
}
