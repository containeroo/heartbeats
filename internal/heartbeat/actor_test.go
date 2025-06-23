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
)

func TestActor_Run_Smoke(t *testing.T) {
	t.Parallel()

	t.Run("Let heartbeat expire", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewRingStore(20)
		store := notifier.InitializeStore(nil, false, "0.0.0", logger)

		store.Register("r1", &notifier.MockNotifier{
			NotifyFunc: func(ctx context.Context, data notifier.NotificationData) error {
				return hist.RecordEvent(ctx, history.Event{
					Timestamp:   time.Now(),
					Type:        history.EventTypeNotificationSent,
					HeartbeatID: data.ID,
					Payload:     data,
				})
			},
		})
		disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)
		go disp.Run(ctx)

		actor := NewActor(
			ctx,
			"heartbeat-1",
			"Test Actor",
			100*time.Millisecond,
			100*time.Millisecond,
			[]string{"r1"},
			logger,
			hist,
			disp.Mailbox(),
		)

		go actor.Run(ctx)

		// Send one heartbeat to enter active state
		actor.Mailbox() <- common.EventReceive

		// Let the heartbeat expire:
		// checkTimer (100ms) + transitionDelay (1s) +
		// graceTimer (100ms) + transitionDelay (1s) + buffer
		time.Sleep(2300 * time.Millisecond)

		events := hist.GetEvents()
		t.Logf("events: %+v", events)

		var seenIdleToActive,
			seenActiveToGrace,
			seenGraceToMissing,
			seenNotification bool

		for _, e := range events {
			switch {
			case e.Type == history.EventTypeStateChanged && e.PrevState == common.HeartbeatStateIdle.String() && e.NewState == common.HeartbeatStateActive.String():
				seenIdleToActive = true
			case e.Type == history.EventTypeStateChanged && e.PrevState == common.HeartbeatStateActive.String() && e.NewState == common.HeartbeatStateGrace.String():
				seenActiveToGrace = true
			case e.Type == history.EventTypeStateChanged && e.PrevState == common.HeartbeatStateGrace.String() && e.NewState == common.HeartbeatStateMissing.String():
				seenGraceToMissing = true
			case e.Type == history.EventTypeNotificationSent:
				seenNotification = true
			}
			t.Logf("%s → %s (%s)", e.PrevState, e.NewState, e.Type)
		}

		assert.True(t, seenIdleToActive, "expected Idle → Active transition")
		assert.True(t, seenActiveToGrace, "expected Active → Grace transition")
		assert.True(t, seenGraceToMissing, "expected Grace → Missing transition")
		assert.True(t, seenNotification, "expected notification to be sent")
	})

	t.Run("Let heartbeat recover", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewRingStore(20)
		store := notifier.InitializeStore(nil, false, "0.0.0", logger)

		store.Register("r1", &notifier.MockNotifier{
			NotifyFunc: func(ctx context.Context, data notifier.NotificationData) error {
				return nil
			},
		})
		disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)
		go disp.Run(ctx)

		actor := NewActor(
			ctx,
			"heartbeat-2",
			"Test Actor",
			100*time.Millisecond,
			100*time.Millisecond,
			[]string{"r1"},
			logger,
			hist,
			disp.Mailbox(),
		)

		go actor.Run(ctx)

		// First bump: idle → active
		actor.Mailbox() <- common.EventReceive
		time.Sleep(transitionDelay + 50*time.Millisecond)

		// Wait for active → grace
		time.Sleep(actor.Interval + transitionDelay + 50*time.Millisecond)

		// bump during grace: grace → active (recovery)
		actor.Mailbox() <- common.EventReceive
		time.Sleep(transitionDelay + 50*time.Millisecond)

		// Wait again for active → grace
		time.Sleep(actor.Interval + transitionDelay + 50*time.Millisecond)

		// Let it go to missing
		time.Sleep(actor.Grace + transitionDelay + 50*time.Millisecond)

		// final bump: missing → active (triggers recovery)
		actor.Mailbox() <- common.EventReceive
		time.Sleep(transitionDelay + 50*time.Millisecond)

		events := hist.GetEvents()
		t.Logf("events: %+v", events)

		var seenIdleToActive,
			seenActiveToGrace,
			seenGraceToActive,
			seenSecondGrace,
			seenGraceToMissing,
			seenRecovered bool

		for _, e := range events {
			switch {
			case e.Type == history.EventTypeStateChanged && e.PrevState == common.HeartbeatStateIdle.String() && e.NewState == common.HeartbeatStateActive.String():
				seenIdleToActive = true
			case e.Type == history.EventTypeStateChanged && e.PrevState == common.HeartbeatStateActive.String() && e.NewState == common.HeartbeatStateGrace.String():
				if !seenActiveToGrace {
					seenActiveToGrace = true
				} else {
					seenSecondGrace = true
				}
			case e.Type == history.EventTypeStateChanged && e.PrevState == common.HeartbeatStateGrace.String() && e.NewState == common.HeartbeatStateActive.String():
				seenGraceToActive = true
			case e.Type == history.EventTypeStateChanged && e.PrevState == common.HeartbeatStateGrace.String() && e.NewState == common.HeartbeatStateMissing.String():
				seenGraceToMissing = true
			case e.Type == history.EventTypeStateChanged && e.PrevState == common.HeartbeatStateMissing.String() && e.NewState == common.HeartbeatStateActive.String():
				seenRecovered = true
			}
		}

		assert.True(t, seenIdleToActive, "expected idle → active")
		assert.True(t, seenActiveToGrace, "expected active → grace")
		assert.True(t, seenGraceToActive, "expected grace → active")
		assert.True(t, seenSecondGrace, "expected second active → grace")
		assert.True(t, seenGraceToMissing, "expected grace → missing")
		assert.True(t, seenRecovered, "expected missing → active (recovery)")
	})
}

func TestActor_onFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(5)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)

	actor := NewActor(
		ctx,
		"fail-case",
		"fail test",
		time.Second,
		time.Second,
		nil,
		logger,
		hist,
		disp.Mailbox(),
	)

	actor.State = common.HeartbeatStateActive
	actor.onFail()

	assert.Equal(t, common.HeartbeatStateFailed, actor.State)
}

func TestActor_onEnterGrace_and_Missing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)

	actor := NewActor(
		ctx,
		"grace-case",
		"grace test",
		time.Second,
		time.Second,
		nil,
		logger,
		hist,
		disp.Mailbox(),
	)

	actor.State = common.HeartbeatStateActive
	actor.onEnterGrace()
	assert.Equal(t, common.HeartbeatStateGrace, actor.State)

	actor.State = common.HeartbeatStateGrace
	actor.onEnterMissing()
	assert.Equal(t, common.HeartbeatStateMissing, actor.State)
}

func TestActor_runPending_and_setPending(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(5)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)

	actor := NewActor(
		ctx,
		"delay-case",
		"delay test",
		time.Second,
		time.Second,
		nil,
		logger,
		hist,
		disp.Mailbox(),
	)

	called := false
	fn := func() { called = true }

	actor.setPending(fn)
	actor.runPending()

	assert.True(t, called)
	assert.Nil(t, actor.pending)
}
