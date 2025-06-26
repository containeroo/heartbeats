package history

import (
	"encoding/json"
	"fmt"
	"time"
)

// EventType categorizes what happened.
type EventType string

const (
	EventTypeHeartbeatReceived  EventType = "HeartbeatReceived"  // a ping from a client
	EventTypeHeartbeatFailed    EventType = "HeartbeatFailed"    // a manual failure
	EventTypeStateChanged       EventType = "StateChanged"       // state transition of an actor
	EventTypeNotificationSent   EventType = "NotificationSent"   // a notification was sent
	EventTypeNotificationFailed EventType = "NotificationFailed" // a notification failed
)

// String returns the EventType as string.
func (h EventType) String() string { return string(h) }

// Event is a generic record of something that happened.
type Event struct {
	Timestamp   time.Time       `json:"timestamp"`         // when it happened
	Type        EventType       `json:"type"`              // what kind of event
	HeartbeatID string          `json:"heartbeat_id"`      // which heartbeat it belongs to
	RawPayload  json.RawMessage `json:"payload,omitempty"` // optional payload
}

// ToJSON returns the raw payload as a JSON string.
func (e *Event) ToJSON() string {
	if e.RawPayload == nil {
		return ""
	}
	return string(e.RawPayload)
}

// NewEvent creates a new Event with optional payload.
func NewEvent(t EventType, id string, payload any) (Event, error) {
	var raw json.RawMessage
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return Event{}, fmt.Errorf("marshal payload: %w", err)
		}
		raw = data
	}

	return Event{
		Timestamp:   time.Now(),
		Type:        t,
		HeartbeatID: id,
		RawPayload:  raw,
	}, nil
}

// MustNewEvent creates a new Event or panics if payload marshalling fails.
func MustNewEvent(t EventType, id string, payload any) Event {
	ev, err := NewEvent(t, id, payload)
	if err != nil {
		panic(err)
	}
	return ev
}

// DecodePayload unmarshals the raw payload into the given target.
func (e *Event) DecodePayload(v any) error {
	if len(e.RawPayload) == 0 {
		return fmt.Errorf("empty payload")
	}
	return json.Unmarshal(e.RawPayload, v)
}
