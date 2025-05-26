package heartbeat

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/stretchr/testify/assert"
)

func TestActor_Run_Smoke(t *testing.T) {
	t.Parallel()

	t.Run("Let heartbeat expire", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		var logBuffer strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
		hist := history.NewRingStore(20)
		store := notifier.InitializeStore(nil, false, logger)
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

		// Check history includes state transitions and a notification
		events := hist.GetEvents()
		assert.Len(t, events, 4)

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

		assert.True(t, hasActive, "expected transition to Active")
		assert.True(t, hasGrace, "expected transition to Missing")
		assert.True(t, hasMissing, "expected transition to Missing")
		assert.True(t, hasNotification, "expected at least one notification")
	})

	t.Run("Let heartbeat recover", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		var logBuffer strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
		hist := history.NewRingStore(20)
		store := notifier.InitializeStore(nil, false, logger)
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

		time.Sleep(400 * time.Millisecond)     // wait until actor is idle
		actor.Mailbox() <- common.EventReceive // send receive
		time.Sleep(400 * time.Millisecond)     // enough for idle → active → grace → missing
		actor.Mailbox() <- common.EventReceive // recover
		time.Sleep(400 * time.Millisecond)     // give time for recovery transition

		events := hist.GetEvents()

		var hasRecoveredState, hasRecoveredNotification bool
		for _, e := range events {
			if e.PrevState == common.HeartbeatStateMissing.String() && e.NewState == common.HeartbeatStateActive.String() {
				hasRecoveredState = true
			}
			p := e.Payload
			if p == nil {
				continue
			}
			n, ok := p.(*notifier.NotificationData)
			assert.True(t, ok)
			if n.Status == common.HeartbeatStateRecovered.String() {
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
