package history

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBadgerStore(t *testing.T) {
	t.Parallel()

	t.Run("record and retrieve event", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		store, err := NewBadger(dir)
		assert.NoError(t, err)
		t.Cleanup(func() {
			_ = store.Close()
			_ = os.RemoveAll(dir)
		})

		ev := MustNewEvent(EventTypeStateChanged, "hb1", StateChangePayload{
			From: "idle",
			To:   "active",
		})

		err = store.RecordEvent(context.Background(), ev)
		assert.NoError(t, err)

		events := store.GetEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, "hb1", events[0].HeartbeatID)

		byID := store.GetEventsByID("hb1")
		assert.Len(t, byID, 1)
		assert.Equal(t, "hb1", byID[0].HeartbeatID)
	})

	t.Run("empty store", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		store, err := NewBadger(dir)
		assert.NoError(t, err)
		t.Cleanup(func() {
			_ = store.Close()
			_ = os.RemoveAll(dir)
		})

		assert.Empty(t, store.GetEvents())
		assert.Empty(t, store.GetEventsByID("does-not-exist"))
	})

	t.Run("second close should not error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		store, err := NewBadger(dir)
		assert.NoError(t, err)

		assert.NoError(t, store.Close())
		assert.NoError(t, store.Close(), "second close should not error")
	})
}
