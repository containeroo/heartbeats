package common

// HeartbeatState represents the internal state of a heartbeat actor.
type HeartbeatState string

const (
	HeartbeatStateIdle      HeartbeatState = "idle"      // Idle is the heartbeat initial state.
	HeartbeatStateActive    HeartbeatState = "active"    // Active is the heartbeat active state.
	HeartbeatStateGrace     HeartbeatState = "grace"     // Grace is the heartbeat grace state.
	HeartbeatStateMissing   HeartbeatState = "missing"   // Missing is the heartbeat missing state.
	HeartbeatStateFailed    HeartbeatState = "failed"    // Failed is the heartbeat failed state.
	HeartbeatStateRecovered HeartbeatState = "recovered" // Recovered is the heartbeat recovered state. Only used for sending "resolved" messages.
)

// String returns the HeartbeatState as string.
func (h HeartbeatState) String() string {
	return string(h)
}

// EventType represents a heartbeat event.
type EventType int

const (
	EventReceive EventType = iota // EventReceive is a heartbeat ping.
	EventFail                     // EventFail is a manual failure.
)
