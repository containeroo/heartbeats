package internal

import (
	"sync"
	"time"
)

type Timer struct {
	mutex     sync.Mutex
	timer     *time.Timer
	duration  *time.Duration
	Cancelled bool
	Completed bool
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
