package notify

import (
	"fmt"
	"sync/atomic"
	"time"

	kit "github.com/containeroo/notifykit/notify"
)

var eventCounter uint64

// Event wraps a heartbeat lifecycle event as a notifykit notification.
type Event struct {
	IDValue      string
	Heartbeat    string
	TitleValue   string
	StatusValue  string
	Body         string
	SinceValue   time.Duration
	Time         time.Time
	Interval     time.Duration
	LateAfter    time.Duration
	ReceiverList []kit.ReceiverID
}

// NewEvent constructs an Event notification.
func NewEvent(
	heartbeatID, title, status, body string,
	since time.Duration,
	timestamp time.Time,
	interval time.Duration,
	lateAfter time.Duration,
	receivers []kit.ReceiverID,
) *Event {
	return &Event{
		IDValue:      nextEventID(),
		Heartbeat:    heartbeatID,
		TitleValue:   title,
		StatusValue:  status,
		Body:         body,
		SinceValue:   since,
		Time:         timestamp,
		Interval:     interval,
		LateAfter:    lateAfter,
		ReceiverList: append([]kit.ReceiverID(nil), receivers...),
	}
}

// ID returns the notification id used by notifykit.
func (e *Event) ID() string {
	if e == nil {
		return ""
	}
	return e.IDValue
}

// ReceiverIDs returns explicit receiver routing.
func (e *Event) ReceiverIDs() []kit.ReceiverID {
	if e == nil || len(e.ReceiverList) == 0 {
		return nil
	}
	return append([]kit.ReceiverID(nil), e.ReceiverList...)
}

// Data returns the receiver-scoped template data.
func (e *Event) Data(receiver string, vars map[string]any, subject string) any {
	if e == nil {
		return nil
	}
	return NewData(*e, receiver, vars, subject)
}

func nextEventID() string {
	seq := atomic.AddUint64(&eventCounter, 1)
	return fmt.Sprintf("%d-%d", time.Now().UTC().UnixNano(), seq)
}
