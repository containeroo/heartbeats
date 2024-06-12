package timer

import (
	"context"
	"sync"
	"time"
)

// TimerCallback is a function type that defines the signature for callback functions
// to be executed by Timer.
type TimerCallback func()

// Timer encapsulates a time.Timer for scheduling callbacks at specific intervals.
type Timer struct {
	Timer    *time.Timer    `yaml:"-"`
	Interval *time.Duration `yaml:"interval,omitempty"`
	Mutex    sync.Mutex     `yaml:"-"`
	cancel   context.CancelFunc
}

// UnmarshalYAML custom unmarshals a duration string into a Timer.
func (t *Timer) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var durationStr string
	if err := unmarshal(&durationStr); err != nil {
		return err
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return err
	}

	t.Interval = &duration

	return nil
}

// RunTimer runs the interval timer and executes the callback when elapsed.
// The timer respects the context for cancellation.
//
// Parameters:
//   - ctx: Context controlling the lifecycle of the timer.
//   - callback: Function to be executed when the timer elapses.
func (t *Timer) RunTimer(ctx context.Context, callback TimerCallback) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	// If there's an existing timer, stop it
	if t.Timer != nil {
		t.Timer.Stop()
	}

	// Create a new context with cancel function
	var cancelCtx context.Context
	cancelCtx, t.cancel = context.WithCancel(ctx)

	// Create a new timer
	t.Timer = time.AfterFunc(*t.Interval, func() {
		select {
		case <-cancelCtx.Done():
			// Context was cancelled
			return
		default:
			// Timer elapsed
			callback()
		}
	})
}

// StopTimer stops the interval timer.
//
// This function cancels any in-progress timer callbacks and stops the timer.
func (t *Timer) StopTimer() {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	// Cancel the context to stop any in-progress timer callbacks
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}

	if t.Timer != nil {
		t.Timer.Stop()
		t.Timer = nil
	}
}
