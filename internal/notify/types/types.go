package types

import (
	"context"
	"time"
)

// Notification describes a notification payload and routing info.
type Notification interface {
	HeartbeatID() string
	Status() string
	Payload() string
	Since() time.Duration
	Timestamp() time.Time
	ReceiverNames() []string
}

// Delivery sends notifications to receivers.
type Delivery interface {
	Dispatch(ctx context.Context, n Payload, receivers []*Receiver) error
}

// Target delivers a notification to a single destination.
type Target interface {
	Send(n Payload) error
	Type() string
}

// DeliveryResult captures target response details for history.
type DeliveryResult struct {
	Status     string // Target-specific status (e.g., HTTP status).
	StatusCode int    // Numeric status code when available.
	Response   string // Target-specific response summary/body.
}

// ResultTarget returns delivery details for history recording.
type ResultTarget interface {
	SendResult(n Payload) (DeliveryResult, error)
}

// Notifier enqueues notifications for delivery.
type Notifier interface {
	Enqueue(n Notification) string
}

// ReceiverRegistry registers receivers per heartbeat.
type ReceiverRegistry interface {
	Register(heartbeatID string, receivers map[string]*Receiver)
}

// HeartbeatRegistry registers heartbeat metadata for dispatch.
type HeartbeatRegistry interface {
	RegisterHeartbeat(heartbeatID string, meta HeartbeatMeta)
}

// Payload describes a heartbeat event to deliver.
type Payload struct {
	HeartbeatID string        // Heartbeat identifier.
	Title       string        // Heartbeat title.
	Status      string        // Status identifier.
	Payload     string        // Raw last payload.
	Timestamp   time.Time     // Event time.
	Interval    time.Duration // Expected heartbeat interval.
	LateAfter   time.Duration // Late window duration.
	Since       time.Duration // Time since last heartbeat.
}

// HeartbeatMeta provides heartbeat context for rendering.
type HeartbeatMeta struct {
	Title     string        // Heartbeat title.
	Interval  time.Duration // Expected heartbeat interval.
	LateAfter time.Duration // Late window duration.
}

// Receiver describes the runtime delivery configuration.
type Receiver struct {
	Name    string         // Receiver name.
	Retry   RetryConfig    // Retry policy.
	Targets []Target       // Delivery targets.
	Vars    map[string]any // Receiver variables.
}

// RetryConfig defines retry behavior for a receiver.
type RetryConfig struct {
	Count int           // Number of retries.
	Delay time.Duration // Delay between retries.
}
