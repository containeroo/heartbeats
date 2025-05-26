package heartbeat

import (
	"testing"
	"time"

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
