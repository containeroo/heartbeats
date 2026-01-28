package event

import "time"

// Event represents a notification with delivery configuration.
type Event struct {
	Heartbeat string
	StatusVal string
	Body      string
	SinceVal  time.Duration
	Time      time.Time
	Receivers []string
}

// NewEvent constructs an Event with delivery configuration.
func NewEvent(
	heartbeatID, status, body string,
	since time.Duration,
	timestamp time.Time,
	receivers []string,
) *Event {
	return &Event{
		Heartbeat: heartbeatID,
		StatusVal: status,
		Body:      body,
		SinceVal:  since,
		Time:      timestamp,
		Receivers: receivers,
	}
}

// HeartbeatID returns the heartbeat id.
func (e *Event) HeartbeatID() string { return e.Heartbeat }

// Status returns the status identifier.
func (e *Event) Status() string { return e.StatusVal }

// Payload returns the raw payload.
func (e *Event) Payload() string { return e.Body }

// Since returns the elapsed time since last heartbeat.
func (e *Event) Since() time.Duration { return e.SinceVal }

// Timestamp returns the event timestamp.
func (e *Event) Timestamp() time.Time { return e.Time }

// ReceiverNames returns the receiver names for the notification.
func (e *Event) ReceiverNames() []string { return e.Receivers }
