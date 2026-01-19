package heartbeat

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/history"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
	"github.com/stretchr/testify/assert"
)

func TestStopTimer(t *testing.T) {
	t.Parallel()

	t.Run("nil timer is safe", func(t *testing.T) {
		t.Parallel()

		var timer *time.Timer = nil
		stopTimer(&timer)
		assert.Nil(t, timer)
	})

	t.Run("active timer is stopped and nil", func(t *testing.T) {
		t.Parallel()

		timer := time.NewTimer(1 * time.Hour)
		stopTimer(&timer)
		assert.Nil(t, timer)
	})

	t.Run("expired timer is drained and nil", func(t *testing.T) {
		t.Parallel()

		timer := time.NewTimer(10 * time.Millisecond)
		time.Sleep(20 * time.Millisecond) // let it fire

		select {
		case <-timer.C:
			// already drained
		default:
			t.Fatal("expected timer to fire before calling stopTimer")
		}

		// NOTE: timer is fired but unread, which we simulate by creating a new fired timer
		timer = time.NewTimer(10 * time.Millisecond)
		time.Sleep(20 * time.Millisecond)

		stopTimer(&timer)
		assert.Nil(t, timer)
	})
}

func TestRecordStateChange_ChangesState(t *testing.T) {
	t.Parallel()

	t.Run("Status changes", func(t *testing.T) {
		var logBuf strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))
		hist := history.NewRingStore(10)
		recorder := servicehistory.NewRecorder(hist)

		a := &Actor{
			ID:     "demo",
			ctx:    context.Background(),
			logger: logger,
			hist:   recorder,
		}

		a.recordStateChange(HeartbeatStateGrace, HeartbeatStateActive) // nolint:errcheck

		assert.True(t, waitForPayloadEvent(t, hist, func(p history.StateChangePayload) bool {
			return p.From == "grace" && p.To == "active"
		}, 100*time.Millisecond), "expected state change: missing → active")
	})

	t.Run("Status does not change", func(t *testing.T) {
		t.Parallel()

		var logBuf strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))
		hist := history.NewRingStore(10)
		recorder := servicehistory.NewRecorder(hist)

		a := &Actor{
			ID:     "noop",
			ctx:    context.Background(),
			logger: logger,
			hist:   recorder,
		}

		a.recordStateChange(HeartbeatStateActive, HeartbeatStateActive) // nolint:errcheck

		assert.Empty(t, logBuf.String())
	})
	t.Run("record error", func(t *testing.T) {
		t.Parallel()

		var logBuf strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))
		mockHist := &history.MockStore{
			RecordEventFunc: func(ctx context.Context, e history.Event) error {
				return errors.New("fail!")
			},
		}

		recorder := servicehistory.NewRecorder(mockHist)

		a := &Actor{
			ID:     "noop",
			ctx:    context.Background(),
			logger: logger,
			hist:   recorder,
		}

		a.recordStateChange(HeartbeatStateActive, HeartbeatStateActive) // nolint:errcheck

		assert.Empty(t, logBuf.String())
	})
}
