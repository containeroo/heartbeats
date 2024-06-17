package heartbeat

// Event represents different types of heartbeat events.
type Event int16

const (
	// EventBeat indicates a regular heartbeat event.
	EventBeat Event = iota
	// EventInterval indicates an interval event.
	EventInterval
	// EventGrace indicates a grace period event.
	EventGrace
	// EventExpired indicates an expired event.
	EventExpired
	// EventSend indicates a send event.
	EventSend
)

// String returns the string representation of the Event.
func (e Event) String() string {
	return [...]string{"BEAT", "INTERVAL", "GRACE", "EXPIRED", "SEND"}[e]
}

// Status represents the different statuses of a heartbeat.
type Status int16

const (
	// StatusNever indicates the heartbeat has never been active.
	StatusNever Status = iota
	// StatusOK indicates the heartbeat is functioning correctly.
	StatusOK
	// StatusGrace indicates the heartbeat is in a grace period.
	StatusGrace
	// StatusNOK indicates the heartbeat is not functioning correctly.
	StatusNOK
	// StatusUnknown indicates the heartbeat status is unknown.
	StatusUnknown
)

// String returns the string representation of the Status.
func (s Status) String() string {
	return [...]string{"never", "ok", "grace", "nok", "unknown"}[s]
}

// MarshalText implements the encoding.TextMarshaler interface for Status,
// returning the string representation as a byte slice.
func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}
