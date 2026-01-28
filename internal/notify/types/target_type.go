package types

// TargetType represents a delivery target kind.
type TargetType int

const (
	TargetUnknown TargetType = iota
	TargetWebhook
	TargetEmail
)

// String returns the target type identifier.
func (t TargetType) String() string {
	switch t {
	case TargetWebhook:
		return "webhook"
	case TargetEmail:
		return "email"
	default:
		return "unknown"
	}
}
