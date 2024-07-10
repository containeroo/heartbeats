package timer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimer(t *testing.T) {
	t.Run("UnmarshalYAML", func(t *testing.T) {
		var tm Timer
		durationStr := "2s"
		err := tm.UnmarshalYAML(func(v interface{}) error {
			*v.(*string) = durationStr
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 2*time.Second, *tm.Interval)
	})

	t.Run("RunTimer", func(t *testing.T) {
		tm := Timer{
			Interval: durationPtr(1 * time.Second),
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var called bool
		tm.RunTimer(ctx, func(ctx context.Context) {
			called = true
		})

		time.Sleep(2 * time.Second)
		assert.True(t, called)
	})

	t.Run("RunTimerWithCancel", func(t *testing.T) {
		tm := Timer{
			Interval: durationPtr(1 * time.Second),
		}
		ctx, cancel := context.WithCancel(context.Background())

		var called bool
		tm.RunTimer(ctx, func(ctx context.Context) {
			called = true
		})

		cancel()
		time.Sleep(2 * time.Second)
		assert.False(t, called)
	})

	t.Run("StopTimer", func(t *testing.T) {
		tm := Timer{
			Interval: durationPtr(1 * time.Second),
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var called bool
		tm.RunTimer(ctx, func(ctx context.Context) {
			called = true
		})

		tm.StopTimer()
		time.Sleep(2 * time.Second)
		assert.False(t, called)
	})
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}
