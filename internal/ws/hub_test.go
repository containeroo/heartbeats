package ws

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/containeroo/heartbeats/internal/heartbeat/service"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/stretchr/testify/require"
)

func TestHubPublishEventBroadcasts(t *testing.T) {
	hub, writer := captureHub(t)
	ev := history.Event{
		Type:        history.EventHTTPAccess.String(),
		HeartbeatID: "api",
		Receiver:    "ops",
		TargetType:  "webhook",
		Fields: map[string]any{
			"target": "https://example.com",
		},
	}

	hub.PublishEvent(ev)

	require.Equal(t, []string{"history", "heartbeat", "receiver"}, writer.types())
}

func TestHubStartConsumesStream(t *testing.T) {
	hub, writer := captureHub(t)
	stream := make(chan history.Event, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hub.Start(ctx, func(int) (<-chan history.Event, func()) {
		return stream, func() {
			close(stream)
		}
	})

	stream <- history.Event{
		Type:        history.EventHTTPAccess.String(),
		HeartbeatID: "api",
		Receiver:    "ops",
		TargetType:  "webhook",
		Fields: map[string]any{
			"target": "https://example.com",
		},
	}

	require.Eventually(t, func() bool {
		return len(writer.types()) >= 3
	}, time.Second, 10*time.Millisecond)
	require.Equal(t, []string{"history", "heartbeat", "receiver"}, writer.types()[:3])
}

func captureHub(t *testing.T) (*Hub, *typeWriter) {
	t.Helper()

	writer := &typeWriter{}
	orig := wsjsonWrite
	wsjsonWrite = writer.write
	t.Cleanup(func() {
		wsjsonWrite = orig
	})

	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
	hub := NewHub(logger, Providers{
		HeartbeatByID: func(id string) (service.HeartbeatSummary, bool) {
			return service.HeartbeatSummary{ID: id, Status: "missing"}, true
		},
		ReceiverByKey: func(receiver, typ, target string) (service.ReceiverSummary, bool) {
			return service.ReceiverSummary{ID: receiver, Type: typ, Destination: target}, true
		},
	})
	var conn websocket.Conn
	hub.add(&conn)
	require.NotEmpty(t, hub.snapshot())
	t.Cleanup(func() {
		hub.remove(&conn)
	})

	return hub, writer
}

type typeWriter struct {
	typesSent []string
}

func (w *typeWriter) write(_ context.Context, _ *websocket.Conn, msg any) error {
	if m, ok := msg.(message); ok {
		w.typesSent = append(w.typesSent, m.Type)
	}
	return nil
}

func (w *typeWriter) types() []string {
	return append([]string(nil), w.typesSent...)
}
