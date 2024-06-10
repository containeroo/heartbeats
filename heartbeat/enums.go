package heartbeat

type Event int16

const (
	EventBeat Event = iota
	EventInterval
	EventGrace
	EventExpired
	EventSend
)

func (e Event) String() string {
	return [...]string{"BEAT", "INTERVAL", "GRACE", "EXPIRED", "SEND"}[e]
}

type Status int16

const (
	StatusNever Status = iota
	StatusOK
	StatusGrace
	StatusNOK
	StatusUnknown
)

func (s Status) String() string {
	return [...]string{"never", "ok", "grace", "nok", "unknown"}[s]
}

// Implement the TextMarshaler interface
func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}
