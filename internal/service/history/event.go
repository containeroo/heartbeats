package history

import (
	"context"
	"errors"
	"fmt"

	storehistory "github.com/containeroo/heartbeats/internal/history"
)

// EventType categorizes what happened.
type EventType string

const (
	EventTypeHeartbeatReceived  EventType = "HeartbeatReceived"
	EventTypeHeartbeatFailed    EventType = "HeartbeatFailed"
	EventTypeStateChanged       EventType = "StateChanged"
	EventTypeNotificationSent   EventType = "NotificationSent"
	EventTypeNotificationFailed EventType = "NotificationFailed"
)

// Factory builds typed events from business actions.
type Factory struct{}

// NewFactory constructs a history event factory.
func NewFactory() *Factory {
	return &Factory{}
}

// HeartbeatReceived builds a heartbeat-received event.
func (f *Factory) HeartbeatReceived(id, source, method, userAgent string) Event {
	return Event{
		Type:        EventTypeHeartbeatReceived,
		HeartbeatID: id,
		Payload: RequestMetadataPayload{
			Source:    source,
			Method:    method,
			UserAgent: userAgent,
		},
	}
}

// HeartbeatFailed builds a heartbeat-failed event.
func (f *Factory) HeartbeatFailed(id, source, method, userAgent string) Event {
	return Event{
		Type:        EventTypeHeartbeatFailed,
		HeartbeatID: id,
		Payload: RequestMetadataPayload{
			Source:    source,
			Method:    method,
			UserAgent: userAgent,
		},
	}
}

// StateChanged builds a state-changed event.
func (f *Factory) StateChanged(id, from, to string) Event {
	return Event{
		Type:        EventTypeStateChanged,
		HeartbeatID: id,
		Payload: StateChangePayload{
			From: from,
			To:   to,
		},
	}
}

// NotificationSent builds a notification-sent event.
func (f *Factory) NotificationSent(id, receiver, typ, target string) Event {
	return Event{
		Type:        EventTypeNotificationSent,
		HeartbeatID: id,
		Payload: NotificationPayload{
			Receiver: receiver,
			Type:     typ,
			Target:   target,
		},
	}
}

// NotificationFailed builds a notification-failed event.
func (f *Factory) NotificationFailed(id, receiver, typ, target, errMsg string) Event {
	return Event{
		Type:        EventTypeNotificationFailed,
		HeartbeatID: id,
		Payload: NotificationPayload{
			Receiver: receiver,
			Type:     typ,
			Target:   target,
			Error:    errMsg,
		},
	}
}

// Payload is a typed payload marker.
type Payload interface {
	isPayload()
}

// RequestMetadataPayload captures request metadata for a heartbeat bump.
type RequestMetadataPayload struct {
	Source    string
	Method    string
	UserAgent string
}

func (RequestMetadataPayload) isPayload() {}

// StateChangePayload captures heartbeat state transitions.
type StateChangePayload struct {
	From string
	To   string
}

func (StateChangePayload) isPayload() {}

// NotificationPayload captures notification delivery details.
type NotificationPayload struct {
	Receiver string
	Type     string
	Target   string
	Error    string
}

func (NotificationPayload) isPayload() {}

// Event is a typed history event.
type Event struct {
	Type        EventType
	HeartbeatID string
	Payload     Payload
}

// Recorder appends typed events to the history store.
type Recorder struct {
	store storehistory.Store
}

// NewRecorder wraps a history store.
func NewRecorder(store storehistory.Store) *Recorder {
	return &Recorder{store: store}
}

// Append records a typed event.
func (r *Recorder) Append(ctx context.Context, e Event) error {
	if r == nil || r.store == nil {
		return errors.New("history store is nil")
	}

	storeEvent, err := toStoreEvent(e)
	if err != nil {
		return err
	}

	return r.store.Append(ctx, storeEvent)
}

func toStoreEvent(e Event) (storehistory.Event, error) {
	switch e.Type {
	case EventTypeHeartbeatReceived:
		p, ok := e.Payload.(RequestMetadataPayload)
		if !ok {
			return storehistory.Event{}, fmt.Errorf("invalid payload for %s", e.Type)
		}
		return storehistory.NewEvent(storehistory.EventTypeHeartbeatReceived, e.HeartbeatID, storehistory.RequestMetadataPayload{
			Source:    p.Source,
			Method:    p.Method,
			UserAgent: p.UserAgent,
		})
	case EventTypeHeartbeatFailed:
		p, ok := e.Payload.(RequestMetadataPayload)
		if !ok {
			return storehistory.Event{}, fmt.Errorf("invalid payload for %s", e.Type)
		}
		return storehistory.NewEvent(storehistory.EventTypeHeartbeatFailed, e.HeartbeatID, storehistory.RequestMetadataPayload{
			Source:    p.Source,
			Method:    p.Method,
			UserAgent: p.UserAgent,
		})
	case EventTypeStateChanged:
		p, ok := e.Payload.(StateChangePayload)
		if !ok {
			return storehistory.Event{}, fmt.Errorf("invalid payload for %s", e.Type)
		}
		return storehistory.NewEvent(storehistory.EventTypeStateChanged, e.HeartbeatID, storehistory.StateChangePayload{
			From: p.From,
			To:   p.To,
		})
	case EventTypeNotificationSent:
		p, ok := e.Payload.(NotificationPayload)
		if !ok {
			return storehistory.Event{}, fmt.Errorf("invalid payload for %s", e.Type)
		}
		return storehistory.NewEvent(storehistory.EventTypeNotificationSent, e.HeartbeatID, storehistory.NotificationPayload{
			Receiver: p.Receiver,
			Type:     p.Type,
			Target:   p.Target,
			Error:    p.Error,
		})
	case EventTypeNotificationFailed:
		p, ok := e.Payload.(NotificationPayload)
		if !ok {
			return storehistory.Event{}, fmt.Errorf("invalid payload for %s", e.Type)
		}
		return storehistory.NewEvent(storehistory.EventTypeNotificationFailed, e.HeartbeatID, storehistory.NotificationPayload{
			Receiver: p.Receiver,
			Type:     p.Type,
			Target:   p.Target,
			Error:    p.Error,
		})
	default:
		return storehistory.Event{}, fmt.Errorf("unknown event type %q", e.Type)
	}
}
