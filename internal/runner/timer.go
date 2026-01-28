package runner

import "time"

// stageTimer manages a resettable one-shot timer with an active flag.
type stageTimer struct {
	timer  *time.Timer // Underlying timer instance.
	active bool        // Whether the timer is currently armed.
}

// Reset arms or re-arms the timer to fire after the duration.
func (t *stageTimer) Reset(d time.Duration) {
	if t.timer == nil {
		t.timer = time.NewTimer(d)
		t.active = true
		return
	}
	if !t.timer.Stop() {
		select {
		case <-t.timer.C:
		default:
		}
	}
	t.timer.Reset(d)
	t.active = true
}

// Stop disarms the timer and drains the channel if needed.
func (t *stageTimer) Stop() {
	if !t.active || t.timer == nil {
		return
	}
	stopTimer(&t.timer)
	t.active = false
}

// C returns the timer channel when active, or nil when inactive.
func (t *stageTimer) C() <-chan time.Time {
	if !t.active || t.timer == nil {
		return nil
	}
	return t.timer.C
}

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
