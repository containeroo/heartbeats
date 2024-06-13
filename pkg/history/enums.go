package history

// Event type represents various events that can be logged in history.
type Event int16

const (
	Beat Event = iota
	Interval
	Grace
	Expired
	Send
)

func (e Event) String() string {
	return [...]string{"BEAT", "INTERVAL", "GRACE", "EXPIRED", "SEND"}[e]
}
