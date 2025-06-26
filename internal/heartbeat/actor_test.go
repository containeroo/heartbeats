package heartbeat

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/stretchr/testify/assert"
)

// waitForPayloadEvent waits for a matching event to appear in history, with timeout.
func waitForPayloadEvent[T any](t *testing.T, hist history.Store, match func(T) bool, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, ev := range hist.GetEvents() {
			var payload T
			if err := ev.DecodePayload(&payload); err != nil {
				continue // skip non-matching or invalid payloads
			}
			if match(payload) {
				return true
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// waitForRawEvent waits for an event that matches the given predicate on the full event.
func waitForRawEvent(t *testing.T, hist history.Store, match func(history.Event) bool, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if slices.ContainsFunc(hist.GetEvents(), match) {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

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
				return hist.RecordEvent(ctx, history.MustNewEvent(history.EventTypeNotificationSent, data.ID, data))
			},
		})

		disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)
		go disp.Run(ctx)

		actor := NewActorFromConfig(ActorConfig{
			Ctx:         ctx,
			ID:          "heartbeat-1",
			Description: "Test Actor",
			Interval:    100 * time.Millisecond,
			Grace:       100 * time.Millisecond,
			Receivers:   []string{"r1"},
			Logger:      logger,
			History:     hist,
			DispatchCh:  disp.Mailbox(),
		})
		go actor.Run(ctx)

		// First receive to transition out of idle
		actor.Mailbox() <- common.EventReceive

		timeout := transitionDelay + 150*time.Millisecond

		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "idle" && p.To == "active"
		}, timeout))

		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "active" && p.To == "grace"
		}, timeout))

		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "grace" && p.To == "missing"
		}, timeout))

		assert.True(t, waitForRawEvent(t, hist, func(e history.Event) bool {
			return e.Type == history.EventTypeNotificationSent || e.Type == history.EventTypeNotificationFailed
		}, timeout))
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

		actor := NewActorFromConfig(ActorConfig{
			Ctx:         ctx,
			ID:          "heartbeat-2",
			Description: "Test Actor",
			Interval:    100 * time.Millisecond,
			Grace:       100 * time.Millisecond,
			Receivers:   []string{"r1"},
			Logger:      logger,
			History:     hist,
			DispatchCh:  disp.Mailbox(),
		})
		go actor.Run(ctx)

		actor.Mailbox() <- common.EventReceive
		timeout := transitionDelay + 200*time.Millisecond

		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "idle" && p.To == "active"
		}, timeout), "expected Idle → Active")

		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "active" && p.To == "grace"
		}, timeout), "expected Active → Grace")

		timeout += 500 * time.Millisecond // allow time before recover
		actor.Mailbox() <- common.EventReceive

		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "grace" && p.To == "active"
		}, timeout), "expected Grace → Active")

		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "active" && p.To == "grace"
		}, timeout), "expected Active → Grace")

		timeout += time.Second // wait for missing state
		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "grace" && p.To == "missing"
		}, timeout), "expected Grace → Missing")

		// wait for notification
		assert.True(t, waitForRawEvent(t, hist, func(e history.Event) bool {
			return e.Type == history.EventTypeNotificationSent || e.Type == history.EventTypeNotificationFailed
		}, timeout))

		timeout += time.Second // final recover
		actor.Mailbox() <- common.EventReceive

		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "missing" && p.To == "active"
		}, timeout), "expected Missing → Active")
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

		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "missing" && p.To == "active"
		}, 100*time.Millisecond), "expected state change: missing → active")
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

		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "idle" && p.To == "active"
		}, 100*time.Millisecond), "expected state change: missing → active")
	})

	t.Run("record error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		dispatchCh := make(chan notifier.NotificationData, 1)
		var buf strings.Builder
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		mockHist := history.MockStore{
			RecordEventFunc: func(ctx context.Context, e history.Event) error {
				return errors.New("fail!")
			},
		}

		a := &Actor{
			ctx:         ctx,
			ID:          "hb2",
			Description: "Test Heartbeat",
			Interval:    100 * time.Millisecond,
			Grace:       50 * time.Millisecond,
			Receivers:   []string{"r2"},
			logger:      logger,
			hist:        &mockHist,
			dispatchCh:  dispatchCh,
			State:       common.HeartbeatStateIdle,
		}

		a.onReceive()

		assert.Equal(t, common.HeartbeatStateActive, a.State)
		assert.Contains(t, buf.String(), "level=ERROR msg=\"failed to record state change\" err=fail!\n")
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
			assert.Equal(t, common.HeartbeatStateFailed.String(), n.Status)
			assert.Equal(t, []string{"r1"}, n.Receivers)
			assert.WithinDuration(t, time.Now(), n.LastBump, time.Second)
		case <-time.After(10 * time.Millisecond):
			t.Fatal("expected notification not sent")
		}
	})

	t.Run("record error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		dispatchCh := make(chan notifier.NotificationData, 1)
		var buf strings.Builder
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		mockHist := history.MockStore{
			RecordEventFunc: func(ctx context.Context, e history.Event) error {
				return errors.New("fail!")
			},
		}

		a := &Actor{
			ctx:         ctx,
			ID:          "hb2",
			Description: "Test Heartbeat",
			Interval:    100 * time.Millisecond,
			Grace:       50 * time.Millisecond,
			Receivers:   []string{"r2"},
			logger:      logger,
			hist:        &mockHist,
			dispatchCh:  dispatchCh,
			State:       common.HeartbeatStateActive,
		}

		a.onFail()

		assert.Equal(t, common.HeartbeatStateFailed, a.State)
		assert.Contains(t, buf.String(), "level=ERROR msg=\"failed to record state change\" err=fail!\n")
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

	t.Run("record error", func(t *testing.T) {
		t.Parallel()

		var buf strings.Builder
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		mockHist := history.MockStore{
			RecordEventFunc: func(ctx context.Context, e history.Event) error {
				return errors.New("fail!")
			},
		}

		act := &Actor{
			ID:         "x",
			State:      common.HeartbeatStateActive,
			Grace:      50 * time.Millisecond,
			logger:     logger,
			hist:       &mockHist,
			mailbox:    make(chan common.EventType, 1),
			dispatchCh: make(chan notifier.NotificationData, 1),
		}

		act.onEnterGrace()

		assert.Equal(t, common.HeartbeatStateGrace, act.State)
		assert.Contains(t, buf.String(), "level=ERROR msg=\"failed to record state change\" err=fail!\n")
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

	t.Run("record error", func(t *testing.T) {
		t.Parallel()

		var buf strings.Builder
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		mockHist := history.MockStore{
			RecordEventFunc: func(ctx context.Context, e history.Event) error {
				return errors.New("fail!")
			},
		}

		act := &Actor{
			ID:         "x",
			State:      common.HeartbeatStateGrace,
			dispatchCh: make(chan notifier.NotificationData, 1),
			hist:       &mockHist,
			logger:     logger,
		}

		act.onEnterMissing()

		assert.Equal(t, common.HeartbeatStateMissing, act.State)
		assert.Contains(t, buf.String(), "level=ERROR msg=\"failed to record state change\" err=fail!\n")
	})
}

func TestActor_OnTest(t *testing.T) {
	t.Parallel()

	t.Run("dispatches notification with correct fields", func(t *testing.T) {
		t.Parallel()

		dispatchCh := make(chan notifier.NotificationData, 1)

		a := &Actor{
			ID:          "test-hb",
			Description: "desc",
			LastBump:    time.Now().Add(-5 * time.Minute),
			State:       common.HeartbeatStateIdle,
			Receivers:   []string{"r1", "r2"},
			dispatchCh:  dispatchCh,
		}

		a.onTest()

		select {
		case msg := <-dispatchCh:
			assert.Equal(t, "test-hb", msg.ID)
			assert.Equal(t, "desc", msg.Description)
			assert.Equal(t, a.LastBump, msg.LastBump)
			assert.Equal(t, "idle", msg.Status)
			assert.Equal(t, []string{"r1", "r2"}, msg.Receivers)
			assert.Equal(t, common.HeartbeatStateIdle.String(), msg.Status)
		default:
			t.Fatal("expected notification to be sent")
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
