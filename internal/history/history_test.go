package history

import (
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
