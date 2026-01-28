package runner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStageTimerReset_NewTimer(t *testing.T) {
	t.Parallel()

	var timer stageTimer
	timer.Reset(10 * time.Millisecond)
	defer timer.Stop()

	require.True(t, timer.active)
	require.NotNil(t, timer.timer)
	require.NotNil(t, timer.C())
}

func TestStageTimerStopDrainsChannel(t *testing.T) {
	t.Parallel()

	var timer stageTimer
	timer.Reset(1 * time.Millisecond)
	// Wait until the timer fires so the channel contains a value.
	select {
	case <-timer.C():
	case <-time.After(50 * time.Millisecond):
		t.Fatal("timer did not fire")
	}

	timer.Stop()
	require.False(t, timer.active)
	require.Nil(t, timer.timer)
	require.Nil(t, timer.C())
}

func TestStageTimerStopInactive(t *testing.T) {
	t.Parallel()

	var timer stageTimer
	timer.Stop() // should be safe when nothing is armed
	require.False(t, timer.active)
	require.Nil(t, timer.timer)
}

func Test_stopTimer_Nil(t *testing.T) {
	t.Parallel()

	var ptr *time.Timer
	stopTimer(&ptr)
	require.Nil(t, ptr)
}
