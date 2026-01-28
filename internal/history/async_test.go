package history

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAsyncRecorderSubscribeAndCancel(t *testing.T) {
	store := NewStore(4)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	rec := NewAsyncRecorder(store, logger, 2)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	rec.Start(ctx)

	sub, cancelSub := rec.Subscribe(1)
	defer cancelSub()

	rec.Add(Event{Type: "evt", HeartbeatID: "hb"})
	select {
	case ev := <-sub:
		require.Equal(t, "evt", ev.Type)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}

	cancelSub()
	_, ok := <-sub
	require.False(t, ok, "subscription channel should close after cancel")
}
