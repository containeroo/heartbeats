package heartbeat

import (
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
)

// stopTimer stops a running timer and ensures its channel is drained.
// If the timer has already fired, Stop() will return false, indicating that
// the timer's channel still contains an event. In that case, we drain the
// channel to prevent unexpected behavior, such as lingering events being
// received later. The timer reference is then set to nil.
//
// This function prevents goroutine leaks and ensures proper cleanup of timers.
//
// Parameters:
//   - t: A pointer to the *time.Timer that should be stopped and cleared.
//     A pointer to a pointer (**time.Timer) is used so the function can both stop the timer and set the original reference to nil, ensuring the caller's variable is correctly updated.
func stopTimer(t **time.Timer) {
	if *t == nil {
		return
	}
	if !(*t).Stop() {
		// drain channel so future reads don't fire spuriously
		select {
		case <-(*t).C: // Drain the timer channel
		default: // No-op if already drained
		}
	}
	*t = nil
}

// stopAllTimers stops any running timers.
func (a *Actor) stopAllTimers() {
	stopTimer(&a.checkTimer)
	stopTimer(&a.graceTimer)
	stopTimer(&a.delayTimer)
}

// recordStateChange logs and records a state change if it actually changed.
func (a *Actor) recordStateChange(prev, next common.HeartbeatState) error {
	if prev == next {
		// avoid noisy logs when state hasn’t changed (e.g. repeated heartbeats in active state)
		return nil
	}

	from := prev.String()
	to := next.String()

	a.logger.Info("state change",
		"heartbeat", a.ID,
		"from", from,
		"to", to,
	)

	factory := servicehistory.NewFactory()
	ev := factory.StateChanged(a.ID, from, to)

	return a.hist.Append(a.ctx, ev)
}
