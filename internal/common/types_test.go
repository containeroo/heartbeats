package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeartbeatState_String(t *testing.T) {
	t.Parallel()

	t.Run("idle", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "idle", HeartbeatStateIdle.String())
	})

	t.Run("active", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "active", HeartbeatStateActive.String())
	})

	t.Run("grace", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "grace", HeartbeatStateGrace.String())
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "missing", HeartbeatStateMissing.String())
	})

	t.Run("failed", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "failed", HeartbeatStateFailed.String())
	})

	t.Run("recovered", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "recovered", HeartbeatStateRecovered.String())
	})
}

func TestEventTypeConstants(t *testing.T) {
	t.Parallel()

	t.Run("EventReceive equals 0", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, EventType(0), EventReceive)
	})

	t.Run("EventFail equals 1", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, EventType(1), EventFail)
	})
}
