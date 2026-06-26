package sender

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	kit "github.com/containeroo/notifykit/notify"
	"github.com/stretchr/testify/require"

	"github.com/containeroo/heartbeats/internal/heartbeat/types"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/metrics"
	appnotify "github.com/containeroo/heartbeats/internal/notify"
	"github.com/containeroo/heartbeats/internal/runner"
)

type captureNotifier struct {
	events []*appnotify.Event
}

func (c *captureNotifier) Enqueue(_ context.Context, n kit.Notification) (string, error) {
	event, ok := n.(*appnotify.Event)
	if ok {
		c.events = append(c.events, event)
	}
	return n.ID(), nil
}

func TestHeartbeatSenderMissingEnqueuesNotification(t *testing.T) {
	t.Parallel()

	hb := &types.Heartbeat{
		ID:          "api",
		Title:       "API",
		Receivers:   []string{"ops"},
		ReceiverIDs: []kit.ReceiverID{"heartbeat.api.receiver.ops"},
	}
	sender, notifier, _, _ := newTestSender(t, hb)

	sender.Missing(time.Now(), time.Second, "payload")

	require.Len(t, notifier.events, 1)
	require.Equal(t, "missing", notifier.events[0].StatusValue)
	require.Equal(t, "api", notifier.events[0].Heartbeat)
}

func TestHeartbeatSenderLateTrailer(t *testing.T) {
	t.Parallel()

	baseHeartbeat := &types.Heartbeat{
		ID:          "api",
		Receivers:   []string{"ops"},
		ReceiverIDs: []kit.ReceiverID{"heartbeat.api.receiver.ops"},
	}

	t.Run("suppressed when alert off", func(t *testing.T) {
		t.Parallel()
		hb := *baseHeartbeat
		hb.AlertOnLate = false
		sender, notifier, _, _ := newTestSender(t, &hb)
		sender.Late(time.Now(), time.Second, "payload")
		require.Empty(t, notifier.events)
	})

	t.Run("enqueues when alert on", func(t *testing.T) {
		t.Parallel()
		hb := *baseHeartbeat
		hb.AlertOnLate = true
		sender, notifier, _, _ := newTestSender(t, &hb)
		sender.Late(time.Now(), time.Second, "payload")
		require.Len(t, notifier.events, 1)
		require.Equal(t, "late", notifier.events[0].StatusValue)
	})
}

func TestHeartbeatSenderRecovered(t *testing.T) {
	t.Parallel()

	baseHeartbeat := &types.Heartbeat{
		ID:          "api",
		Receivers:   []string{"ops"},
		ReceiverIDs: []kit.ReceiverID{"heartbeat.api.receiver.ops"},
	}

	t.Run("suppressed when recovery disabled", func(t *testing.T) {
		t.Parallel()
		hb := *baseHeartbeat
		hb.AlertOnRecovery = false
		sender, notifier, _, _ := newTestSender(t, &hb)
		sender.Recovered(time.Now(), "payload")
		require.Empty(t, notifier.events)
	})

	t.Run("enqueues when recovery enabled", func(t *testing.T) {
		t.Parallel()
		hb := *baseHeartbeat
		hb.AlertOnRecovery = true
		sender, notifier, _, _ := newTestSender(t, &hb)
		sender.Recovered(time.Now(), "payload")
		require.Len(t, notifier.events, 1)
		require.Equal(t, "recovered", notifier.events[0].StatusValue)
	})
}

func TestHeartbeatSenderTransitionRecordsHistory(t *testing.T) {
	t.Parallel()

	hb := &types.Heartbeat{
		ID:        "api",
		Receivers: []string{"ops"},
	}
	sender, _, hist, logBuf := newTestSender(t, hb)

	now := time.Now()
	sender.Transition(now, runner.StageOK, runner.StageLate, 3*time.Second)

	require.Contains(t, logBuf.String(), "Runner stage transitioned")
	events := hist.List()
	require.Len(t, events, 1)
	event := events[0]
	require.Equal(t, history.EventHeartbeatTransition.String(), event.Type)
	require.Equal(t, "late", event.Status)
	require.Equal(t, "ok", event.Fields["from"])
	require.Equal(t, "late", event.Fields["to"])
	require.Equal(t, "3s", event.Fields["since"])
}

func newTestSender(t *testing.T, hb *types.Heartbeat) (*HeartbeatSender, *captureNotifier, *history.Store, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	notifier := &captureNotifier{}
	hist := history.NewStore(10)
	return &HeartbeatSender{
		Heartbeat: hb,
		Notifier:  notifier,
		History:   hist,
		Logger:    slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{})),
		Metrics:   metrics.NewRegistry(),
	}, notifier, hist, buf
}
