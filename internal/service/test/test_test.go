package test

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notifier"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendTestNotification(t *testing.T) {
	t.Parallel()

	notifyCh := make(chan notifier.NotificationData, 1)
	mock := &notifier.MockNotifier{
		NotifyFunc: func(ctx context.Context, data notifier.NotificationData) error {
			notifyCh <- data
			return nil
		},
	}

	store := notifier.NewReceiverStore()
	store.Register("rec-1", mock)

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	metricsReg := metrics.New(hist)
	recorder := servicehistory.NewRecorder(hist)
	disp := notifier.NewDispatcher(store, logger, recorder, 1, 1, 10, metricsReg)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go disp.Run(ctx)

	SendTestNotification(disp, logger, "rec-1")

	select {
	case data := <-notifyCh:
		assert.Len(t, data.Receivers, 1)
		assert.Equal(t, "rec-1", data.Receivers[0])
		assert.Equal(t, "Test Notification", data.Title)
		assert.Equal(t, "This is a test notification", data.Message)
		assert.True(t, strings.HasPrefix(data.ID, "manual-test-"))
	case <-time.After(2 * time.Second):
		require.Fail(t, "timeout waiting for notification")
	}
}

func TestTriggerTestHeartbeat(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	store := notifier.NewReceiverStore()
	recorder := servicehistory.NewRecorder(hist)
	metricsReg := metrics.New(hist)
	disp := notifier.NewDispatcher(store, logger, recorder, 1, 1, 10, metricsReg)

	cfg := heartbeat.HeartbeatConfigMap{
		"hb1": {
			ID:          "hb1",
			Description: "desc",
			Interval:    time.Second,
			Grace:       time.Second,
			Receivers:   []string{"r1"},
		},
	}
	factory := heartbeat.DefaultActorFactory{
		Logger:     logger,
		History:    recorder,
		Metrics:    metricsReg,
		DispatchCh: disp.Mailbox(),
	}
	mgr, err := heartbeat.NewManagerFromHeartbeatMap(
		context.Background(),
		cfg,
		logger,
		factory,
	)
	assert.NoError(t, err)

	assert.NoError(t, TriggerTestHeartbeat(mgr, logger, "hb1"))
	assert.EqualError(t, TriggerTestHeartbeat(mgr, logger, "missing"), "heartbeat ID \"missing\" not found")
}
