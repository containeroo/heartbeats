package history

// NotificationPayload is used for NotificationSent and NotificationFailed events.
type NotificationPayload struct {
	Receiver string `json:"receiver"`        // receiver ID
	Type     string `json:"type"`            // notifier type (e.g., slack, msteams)
	Target   string `json:"target"`          // target destination
	Error    string `json:"error,omitempty"` // optional error message
}

// StateChangePayload is used for StateChanged events.
type StateChangePayload struct {
	From string `json:"from"` // previous state
	To   string `json:"to"`   // new state
}

// RequestMetadataPayload is used for HeartbeatReceived and HeartbeatFailed events.
type RequestMetadataPayload struct {
	Source    string `json:"source"`     // remote IP address
	Method    string `json:"method"`     // HTTP method
	UserAgent string `json:"user_agent"` // user-agent header
}
