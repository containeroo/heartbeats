package history

import (
	"context"
	"time"

	"github.com/containeroo/heartbeats/internal/notifier"
)

// EventType categorizes what happened.
type EventType string

const (
	EventTypeHeartbeatReceived EventType = "HeartbeatReceived" // EventTypeHeartbeatReceived is a ping from a client.
	EventTypeHeartbeatFailed   EventType = "HeartbeatFailed"   // EventTypeHeartbeatFailed is a manual failure.
	EventTypeStateChanged      EventType = "StateChanged"      // EventTypeStateChanged is when an Actor’s state changes.
	EventTypeNotificationSent  EventType = "NotificationSent"  // EventTypeNotificationSent is when a notification is dispatched.
)

// String returns the EventType as string.
func (h EventType) String() string {
	return string(h)
}

// Event is a generic record of something that happened.
type Event struct {
	Timestamp    time.Time                  // when it happened
	Type         EventType                  // what kind of event
	HeartbeatID  string                     // which heartbeat this belongs to
	Source       string                     // e.g. remote address
	Method       string                     // HTTP method, if relevant
	UserAgent    string                     // user-agent, if relevant
	PrevState    string                     // prior state (for state changes)
	NewState     string                     // new state (for state changes)
	Notification *notifier.NotificationData // payload for notification events
}

// Store records and exposes all system events.
type Store interface {
	// RecordEvent appends a new event to history.
	RecordEvent(ctx context.Context, e Event) error
	// GetEvents returns a snapshot of all recorded events.
	GetEvents() []Event
	// GetEventsByID returns all events recorded for the specified heartbeat ID.
	GetEventsByID(id string) []Event
}
