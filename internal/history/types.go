package history

// EventType represents the history event type.
type EventType int

const (
	EventHeartbeatReceived EventType = iota
	EventHeartbeatTransition
	EventHTTPAccess
	EventNotificationDelivered
	EventNotificationFailed
)

// String returns the history event identifier.
func (e EventType) String() string {
	switch e {
	case EventHeartbeatReceived:
		return "heartbeat_received"
	case EventHeartbeatTransition:
		return "heartbeat_transition"
	case EventHTTPAccess:
		return "http_access"
	case EventNotificationDelivered:
		return "notification_delivered"
	case EventNotificationFailed:
		return "notification_failed"
	default:
		return "unknown"
	}
}
