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

		ctx := t.Context()

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

		ctx := t.Context()

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
			case (e.Type == history.EventTypeStateChanged &&
				e.PrevState == common.HeartbeatStateIdle.String() &&
				e.NewState == common.HeartbeatStateActive.String()):
				seenIdleToActive = true
			case (e.Type == history.EventTypeStateChanged &&
				e.PrevState == common.HeartbeatStateActive.String() &&
				e.NewState == common.HeartbeatStateGrace.String()):
				if !seenActiveToGrace {
					seenActiveToGrace = true
				} else {
					seenSecondGrace = true
				}
			case (e.Type == history.EventTypeStateChanged &&
				e.PrevState == common.HeartbeatStateGrace.String() &&
				e.NewState == common.HeartbeatStateActive.String()):
				seenGraceToActive = true
			case (e.Type == history.EventTypeStateChanged &&
				e.PrevState == common.HeartbeatStateGrace.String() &&
				e.NewState == common.HeartbeatStateMissing.String()):
				seenGraceToMissing = true
			case (e.Type == history.EventTypeStateChanged &&
				e.PrevState == common.HeartbeatStateMissing.String() &&
				e.NewState == common.HeartbeatStateActive.String()):
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

func TestActor_OnReceive(t *testing.T) {
	t.Parallel()

	t.Run("previous state was missing", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		dispatchCh := make(chan notifier.NotificationData, 1)
		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewRingStore(10)

		a := &Actor{
			ctx:         ctx,
			ID:          "hb1",
			Description: "Test Heartbeat",
			Interval:    100 * time.Millisecond,
			Grace:       50 * time.Millisecond,
			Receivers:   []string{"r1"},
			logger:      logger,
			hist:        hist,
			dispatchCh:  dispatchCh,
			State:       common.HeartbeatStateMissing,
		}

		a.onReceive()

		assert.Equal(t, common.HeartbeatStateActive, a.State)
		assert.WithinDuration(t, time.Now(), a.LastBump, 50*time.Millisecond)

		select {
		case msg := <-dispatchCh:
			assert.Equal(t, "hb1", msg.ID)
			assert.Equal(t, common.HeartbeatStateRecovered.String(), msg.Status)
		default:
			t.Fatal("expected recovery notification to be sent")
		}

		events := hist.GetEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, "missing", events[0].PrevState)
		assert.Equal(t, "active", events[0].NewState)
	})

	t.Run("previous state was not missing", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		dispatchCh := make(chan notifier.NotificationData, 1)
		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewRingStore(10)

		a := &Actor{
			ctx:         ctx,
			ID:          "hb2",
			Description: "Test Heartbeat",
			Interval:    100 * time.Millisecond,
			Grace:       50 * time.Millisecond,
			Receivers:   []string{"r2"},
			logger:      logger,
			hist:        hist,
			dispatchCh:  dispatchCh,
			State:       common.HeartbeatStateIdle,
		}

		a.onReceive()

		assert.Equal(t, common.HeartbeatStateActive, a.State)
		assert.WithinDuration(t, time.Now(), a.LastBump, 50*time.Millisecond)

		select {
		case msg := <-dispatchCh:
			t.Fatalf("did not expect notification, but got %+v", msg)
		default:
			// expected: no notification
		}

		events := hist.GetEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, "idle", events[0].PrevState)
		assert.Equal(t, "active", events[0].NewState)
	})
}

func TestActor_OnFail(t *testing.T) {
	t.Parallel()

	t.Run("sends missing notification and updates state", func(t *testing.T) {
		t.Parallel()

		recv := make(chan notifier.NotificationData, 1)
		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewRingStore(10)

		act := &Actor{
			ID:          "x",
			Description: "test",
			Receivers:   []string{"r1"},
			State:       common.HeartbeatStateActive,
			dispatchCh:  recv,
			hist:        hist,
			logger:      logger,
		}

		act.onFail()

		assert.Equal(t, common.HeartbeatStateFailed, act.State)

		select {
		case n := <-recv:
			assert.Equal(t, "x", n.ID)
			assert.Equal(t, "test", n.Description)
			assert.Equal(t, common.HeartbeatStateMissing.String(), n.Status)
			assert.Equal(t, []string{"r1"}, n.Receivers)
			assert.WithinDuration(t, time.Now(), n.LastBump, time.Second)
		case <-time.After(10 * time.Millisecond):
			t.Fatal("expected notification not sent")
		}
	})
}

func TestActor_OnEnterGrace(t *testing.T) {
	t.Parallel()

	t.Run("transitions to grace if state is active", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewRingStore(10)

		act := &Actor{
			ID:         "x",
			State:      common.HeartbeatStateActive,
			Grace:      50 * time.Millisecond,
			logger:     logger,
			hist:       hist,
			mailbox:    make(chan common.EventType, 1),
			dispatchCh: make(chan notifier.NotificationData, 1),
		}

		act.onEnterGrace()

		assert.Equal(t, common.HeartbeatStateGrace, act.State)
		assert.NotNil(t, act.graceTimer)
	})

	t.Run("does nothing if state is not active", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewRingStore(10)

		act := &Actor{
			ID:         "x",
			State:      common.HeartbeatStateIdle,
			Grace:      50 * time.Millisecond,
			logger:     logger,
			hist:       hist,
			mailbox:    make(chan common.EventType, 1),
			dispatchCh: make(chan notifier.NotificationData, 1),
		}

		act.onEnterGrace()

		assert.Equal(t, common.HeartbeatStateIdle, act.State)
		assert.Nil(t, act.graceTimer)
	})
}

func TestActor_OnEnterMissing(t *testing.T) {
	t.Parallel()

	t.Run("transitions to missing and sends notification if in grace", func(t *testing.T) {
		t.Parallel()

		recv := make(chan notifier.NotificationData, 1)
		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewRingStore(10)

		now := time.Now()

		act := &Actor{
			ID:          "x",
			Description: "test",
			State:       common.HeartbeatStateGrace,
			LastBump:    now,
			Receivers:   []string{"r1"},
			dispatchCh:  recv,
			hist:        hist,
			logger:      logger,
		}

		act.onEnterMissing()

		assert.Equal(t, common.HeartbeatStateMissing, act.State)

		select {
		case n := <-recv:
			assert.Equal(t, "x", n.ID)
			assert.Equal(t, "test", n.Description)
			assert.Equal(t, now, n.LastBump)
			assert.Equal(t, common.HeartbeatStateMissing.String(), n.Status)
			assert.Equal(t, []string{"r1"}, n.Receivers)
		case <-time.After(10 * time.Millisecond):
			t.Fatal("expected notification not sent")
		}
	})

	t.Run("does nothing if state is not grace", func(t *testing.T) {
		t.Parallel()

		recv := make(chan notifier.NotificationData, 1)
		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewRingStore(10)

		act := &Actor{
			ID:         "x",
			State:      common.HeartbeatStateActive,
			dispatchCh: recv,
			hist:       hist,
			logger:     logger,
		}

		act.onEnterMissing()

		assert.Equal(t, common.HeartbeatStateActive, act.State)

		select {
		case n := <-recv:
			t.Fatalf("unexpected notification sent: %+v", n)
		case <-time.After(5 * time.Millisecond):
			// pass
		}
	})
}

func TestActor_RunPending(t *testing.T) {
	t.Parallel()

	t.Run("executes pending transition and clears it", func(t *testing.T) {
		t.Parallel()

		called := false
		act := &Actor{
			pending: func() {
				called = true
			},
		}

		act.delayTimer = time.NewTimer(time.Hour)
		act.runPending()

		assert.True(t, called, "expected pending function to be called")
		assert.Nil(t, act.pending, "expected pending to be cleared")
	})

	t.Run("does nothing if no pending transition is set", func(t *testing.T) {
		t.Parallel()

		act := &Actor{}
		act.delayTimer = time.NewTimer(time.Hour)

		act.runPending()

		assert.Nil(t, act.pending, "expected pending to still be nil")
	})
}

func TestActor_SetPending(t *testing.T) {
	t.Parallel()

	t.Run("sets pending function and creates delay timer", func(t *testing.T) {
		t.Parallel()

		var called bool
		fn := func() { called = true }

		act := &Actor{}
		act.setPending(fn)

		assert.NotNil(t, act.delayTimer, "expected delayTimer to be set")
		assert.NotNil(t, act.pending, "expected pending function to be set")

		// verify that calling pending runs the intended function
		act.runPending()
		assert.True(t, called, "expected pending function to be executed")

		// cleanup in case runPending didn’t drain
		if act.delayTimer != nil && !act.delayTimer.Stop() {
			select {
			case <-act.delayTimer.C:
			default:
			}
		}
	})
}
