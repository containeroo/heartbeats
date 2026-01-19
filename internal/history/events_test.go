package history

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventType_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "HeartbeatReceived", EventTypeHeartbeatReceived.String())
	assert.Equal(t, "HeartbeatFailed", EventTypeHeartbeatFailed.String())
	assert.Equal(t, "StateChanged", EventTypeStateChanged.String())
	assert.Equal(t, "NotificationSent", EventTypeNotificationSent.String())
	assert.Equal(t, "NotificationFailed", EventTypeNotificationFailed.String())
}

type dummyPayload struct {
	Message string `json:"message"`
}

func TestNewEvent(t *testing.T) {
	t.Parallel()

	t.Run("valid payload", func(t *testing.T) {
		ev, err := NewEvent(EventTypeHeartbeatReceived, "hb1", dummyPayload{Message: "ping"})
		assert.NoError(t, err)
		assert.Equal(t, "HeartbeatReceived", ev.Type.String())
		assert.Equal(t, "hb1", ev.HeartbeatID)
		assert.Equal(t, 1, ev.Version)

		var dp dummyPayload
		assert.NoError(t, json.Unmarshal(ev.RawPayload, &dp))
		assert.Equal(t, "ping", dp.Message)
	})

	t.Run("nil payload", func(t *testing.T) {
		ev, err := NewEvent(EventTypeNotificationSent, "hb2", nil)
		assert.NoError(t, err)
		assert.Nil(t, ev.RawPayload)
		assert.Equal(t, 1, ev.Version)
	})

	t.Run("invalid payload", func(t *testing.T) {
		_, err := NewEvent(EventTypeStateChanged, "hb3", func() {}) // non-marshallable
		assert.Error(t, err)
	})
}

func TestMustNewEvent(t *testing.T) {
	t.Parallel()

	t.Run("panics on marshal error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for unmarshalable payload")
			}
		}()
		_ = MustNewEvent(EventTypeHeartbeatFailed, "hbX", func() {})
	})
}

func TestEvent_ToJSON(t *testing.T) {
	t.Parallel()

	t.Run("returns JSON", func(t *testing.T) {
		ev := MustNewEvent(EventTypeNotificationFailed, "hbY", dummyPayload{Message: "fail"})
		assert.Equal(t, "{\"message\":\"fail\"}", ev.ToJSON())
	})

	t.Run("returns empty string for nil payload", func(t *testing.T) {
		ev := Event{}
		assert.Equal(t, "", ev.ToJSON())
	})
}
