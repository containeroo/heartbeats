package history

// Event represents different types of events in the history.
type Event int16

const (
	// Beat indicates a regular heartbeat event.
	Beat Event = iota
	// Interval indicates an interval event.
	Interval
	// Grace indicates a grace period event.
	Grace
	// Expired indicates an expired event.
	Expired
	// Send indicates a send event.
	Send
)

// String returns the string representation of the Event.
func (e Event) String() string {
	return [...]string{"BEAT", "INTERVAL", "GRACE", "EXPIRED", "SEND"}[e]
}
