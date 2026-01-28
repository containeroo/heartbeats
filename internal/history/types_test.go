package history

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventTypeString(t *testing.T) {
	cases := map[EventType]string{
		EventHeartbeatReceived:     "heartbeat_received",
		EventHeartbeatTransition:   "heartbeat_transition",
		EventHTTPAccess:            "http_access",
		EventNotificationDelivered: "notification_delivered",
		EventNotificationFailed:    "notification_failed",
		EventType(99):              "unknown",
	}
	for typ, expected := range cases {
		require.Equal(t, expected, typ.String())
	}
}
