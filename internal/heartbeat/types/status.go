package types

// Status represents a heartbeat status.
type Status int

const (
	StatusUnknown Status = iota
	StatusOK
	StatusLate
	StatusMissing
	StatusRecovered
)

// String returns the status identifier.
func (s Status) String() string {
	switch s {
	case StatusOK:
		return "ok"
	case StatusLate:
		return "late"
	case StatusMissing:
		return "missing"
	case StatusRecovered:
		return "recovered"
	default:
		return "unknown"
	}
}
