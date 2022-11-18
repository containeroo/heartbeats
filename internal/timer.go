package internal

import (
	"sync"
	"time"
)

type Timer struct {
	mutex     sync.Mutex
	timer     *time.Timer
	duration  *time.Duration
	cancel    chan struct{}
	Cancelled bool
	Completed bool
}

// NewTimer creates a new timer with a duration and a callback function that is called when the timer is expired
func NewTimer(duration time.Duration, complete func()) *Timer {
	t := &Timer{}
	t.duration = &duration
	t.timer = time.NewTimer(duration)
	t.cancel = make(chan struct{})
	go t.wait(complete, func() {})
	return t
}

// NewTimerWithCancel creates a new timer with a duration and a callback function that is called when the timer is expired and a cancel function that is called when the timer is cancelled
func NewTimerWithCancel(duration time.Duration, complete func(), cancel func()) *Timer {
	t := &Timer{}
	t.duration = &duration
	t.timer = time.NewTimer(duration)
	t.cancel = make(chan struct{})
	go t.wait(complete, cancel)
	return t
}

// Reset resets the timer with a new duration
func (t *Timer) Reset(duration time.Duration) {
	if !t.timer.Stop() {
		<-t.timer.C
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
	t.timer.Stop()
	t.cancel <- struct{}{}
}

// wait waits for the timer to expire or be cancelled
func (t *Timer) wait(complete func(), cancel func()) {
	for {
		select {
		case <-t.timer.C:
			t.mutex.Lock()
			if !t.Cancelled {
				t.Completed = true
				t.mutex.Unlock()
				complete()
				return
			}
			t.mutex.Unlock()
		case <-t.cancel:
			cancel()
			return
		}
	}
}
