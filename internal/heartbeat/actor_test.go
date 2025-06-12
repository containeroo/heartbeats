package heartbeat

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActor_Run_Smoke(t *testing.T) {
	t.Parallel()

	t.Run("Let heartbeat expire", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewRingStore(20)
		store := notifier.InitializeStore(nil, false, "0.0.0", logger)
		fn := func(ctx context.Context, data notifier.NotificationData) error {
			_ = hist.RecordEvent(ctx, history.Event{
				Timestamp:   time.Now(),
				Type:        history.EventTypeNotificationSent,
				HeartbeatID: data.ID,
				Payload:     data,
			})
			return nil
		}
		mock := &notifier.MockNotifier{NotifyFunc: fn}
		store.Register("r1", mock) // nolint:errcheck

		disp := notifier.NewDispatcher(store, logger, hist, 1, 1)

		actor := NewActor(
			ctx,
			"heartbeat-1",
			"Test Actor",
			100*time.Millisecond,
			100*time.Millisecond,
			[]string{"r1"},
			logger,
			hist,
			disp,
		)

		go actor.Run(ctx)

		// Send EventReceive — should trigger state Active and setup check timer
		actor.Mailbox() <- common.EventReceive
		time.Sleep(150 * time.Millisecond) // Allow time for timers and logic

		// Allow grace timeout to trigger
		time.Sleep(200 * time.Millisecond)

		// Check final state is 'missing'
		assert.Equal(t, common.HeartbeatStateMissing, actor.State, "expected state Missing, got %s", actor.State)

		// Wait unitl notification is sent
		require.Eventually(t, func() bool {
			return len(hist.GetEvents()) >= 3
		}, 1*time.Second, 100*time.Millisecond)

		// Check history includes state transitions and a notification
		events := hist.GetEvents()
		t.Logf("events: %+v", events)

		// Check events
		var hasGrace, hasActive, hasMissing, hasNotification bool
		for _, e := range events {
			switch e.NewState {
			case common.HeartbeatStateActive.String():
				hasActive = true
			case common.HeartbeatStateGrace.String():
				hasGrace = true
			case common.HeartbeatStateMissing.String():
				hasMissing = true
			}
			if e.Type == history.EventTypeNotificationSent {
				hasNotification = true
			}
		}

		for _, e := range events {
			t.Logf("%s → %s (%s)", e.PrevState, e.NewState, e.Type)
		}

		assert.True(t, hasActive, "expected transition to Active")
		assert.True(t, hasGrace, "expected transition to Missing")
		assert.True(t, hasMissing, "expected transition to Missing")
		assert.True(t, hasNotification, "expected at least one notification")
	})

	t.Run("Let heartbeat recover", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewRingStore(20)
		store := notifier.InitializeStore(nil, false, "0.0.0", logger)

		// Register a mock notifier that does not record its own history
		store.Register("r1", &notifier.MockNotifier{
			NotifyFunc: func(ctx context.Context, data notifier.NotificationData) error {
				return nil // Let Dispatcher handle recording
			},
		})

		disp := notifier.NewDispatcher(store, logger, hist, 1, 1)

		actor := NewActor(
			ctx,
			"heartbeat-1",
			"Test Actor",
			100*time.Millisecond,
			100*time.Millisecond,
			[]string{"r1"},
			logger,
			hist,
			disp,
		)

		go actor.Run(ctx)

		// trigger active → grace → missing
		time.Sleep(400 * time.Millisecond)
		actor.Mailbox() <- common.EventReceive

		// allow full missing cycle
		time.Sleep(400 * time.Millisecond)

		// trigger recovery
		actor.Mailbox() <- common.EventReceive

		// allow state transition to active and notification dispatch
		time.Sleep(400 * time.Millisecond)

		events := hist.GetEvents()
		t.Logf("events: %+v", events)

		var hasRecoveredState, hasRecoveredNotification bool
		for _, e := range events {
			if e.PrevState == common.HeartbeatStateMissing.String() && e.NewState == common.HeartbeatStateActive.String() {
				hasRecoveredState = true
			}
			if e.NewState == common.HeartbeatStateRecovered.String() && e.Type == history.EventTypeNotificationSent {
				hasRecoveredNotification = true
			}
		}

		for _, e := range events {
			t.Logf("%s → %s (%s)", e.PrevState, e.NewState, e.Type)
		}

		assert.True(t, hasRecoveredState, "expected missing → active state transition")
		assert.True(t, hasRecoveredNotification, "expected recovery notification to be sent")
	})
}
