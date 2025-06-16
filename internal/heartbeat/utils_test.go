package heartbeat

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/internal/history"
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

		a := &Actor{
			ID:     "demo",
			ctx:    context.Background(),
			logger: logger,
			hist:   hist,
		}

		a.recordStateChange(common.HeartbeatStateGrace, common.HeartbeatStateActive)

		events := hist.GetEvents()

		assert.Equal(t, "grace", events[0].PrevState)
		assert.Equal(t, "active", events[0].NewState)
		assert.Equal(t, "demo", events[0].HeartbeatID)

		logStr := logBuf.String()
		assert.Contains(t, logStr, "level=INFO msg=\"state change\" heartbeat=demo from=grace to=active\n")
	})

	t.Run("Status does not change", func(t *testing.T) {
		t.Parallel()

		var logBuf strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))
		hist := history.NewRingStore(10)

		a := &Actor{
			ID:     "noop",
			ctx:    context.Background(),
			logger: logger,
			hist:   hist,
		}

		a.recordStateChange(common.HeartbeatStateActive, common.HeartbeatStateActive)

		assert.Empty(t, logBuf.String())
	})
}
