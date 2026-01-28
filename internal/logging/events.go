package logging

// Event represents a structured log event identifier.
type Event int

const (
	EventEncodeResponseFailed Event = iota
	EventHeartbeatMailboxFull
	EventHeartbeatMetadataMissing
	EventHeartbeatStarted
	EventHistoryDropped
	EventNotificationDelivered
	EventNotificationDeliveryFailed
	EventNotificationMissing
	EventNotificationTargetDelivered
	EventNotificationTargetDispatch
	EventNotificationTargetFailed
	EventReceiverEmpty
	EventReceiverMissing
	EventRoutesMounted
	EventStageTransition
	EventWebhookResponse
)

// String returns the log event identifier.
func (e Event) String() string {
	switch e {
	case EventEncodeResponseFailed:
		return "encode_response_failed"
	case EventHeartbeatMailboxFull:
		return "heartbeat_mailbox_full"
	case EventHeartbeatMetadataMissing:
		return "heartbeat_metadata_missing"
	case EventHeartbeatStarted:
		return "heartbeat_started"
	case EventHistoryDropped:
		return "history_dropped"
	case EventNotificationDelivered:
		return "notification_delivered"
	case EventNotificationDeliveryFailed:
		return "notification_delivery_failed"
	case EventNotificationMissing:
		return "notification_missing"
	case EventNotificationTargetDelivered:
		return "notification_target_delivered"
	case EventNotificationTargetDispatch:
		return "notification_target_dispatch"
	case EventNotificationTargetFailed:
		return "notification_target_failed"
	case EventReceiverEmpty:
		return "receiver_empty"
	case EventReceiverMissing:
		return "receiver_missing"
	case EventRoutesMounted:
		return "routes_mounted"
	case EventStageTransition:
		return "stage_transition"
	case EventWebhookResponse:
		return "webhook_response"
	default:
		return "unknown"
	}
}
