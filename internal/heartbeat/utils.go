package heartbeat

import "time"

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
