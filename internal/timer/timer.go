package timer

import (
	"sync"
	"time"
)

// Timer is a struct for a timer
type Timer struct {
	mutex     sync.Mutex
	timer     *time.Timer
	duration  *time.Duration
	Cancelled bool
	Completed bool
}

// IsCompleted returns true if the timer is completed
func (t *Timer) IsCompleted() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.Completed
}

// IsCancelled returns true if the timer is cancelled
func (t *Timer) IsCancelled() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.Cancelled
}

// SetCompleted sets the completed flag
func (t *Timer) SetCompleted(completed bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Completed = completed
}

// SetCancelled sets the cancelled flag
func (t *Timer) SetCancelled(cancelled bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Cancelled = cancelled
}

// NewTimer creates a new timer with a duration and a callback function that is called when the timer is expired
func NewTimer(duration time.Duration, complete func()) *Timer {
	t := &Timer{}
	t.duration = &duration
	t.timer = time.NewTimer(duration)
	go t.wait(complete)
	return t
}

// Reset resets the timer with a new duration
func (t *Timer) Reset(duration time.Duration) {
	if !t.timer.Stop() {
		select {
		case <-t.timer.C:
		default:
		}
	}
	t.timer.Reset(duration)
}

// Cancel cancels the timer
func (t *Timer) Cancel() {
	t.mutex.Lock()
	if t.Completed {
		t.mutex.Unlock()
		return
	}
	t.Cancelled = true
	t.mutex.Unlock()

	if !t.timer.Stop() {
		select {
		case <-t.timer.C:
		default:
		}
	}
}

// wait waits for the timer to expire or be cancelled
func (t *Timer) wait(complete func()) {
	<-t.timer.C
	t.mutex.Lock()
	if !t.Cancelled {
		t.Completed = true
		t.mutex.Unlock()
		complete()
		return
	}
	t.mutex.Unlock()
}
