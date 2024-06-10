package notifier

import (
	"context"
)

// Formatter defines a function type for formatting notification content.
type Formatter func(content string, data interface{}, isResolved bool) (string, error)

// Notifier defines the interface for different notification types.
type Notifier interface {
	// Send triggers the notification with provided data and resolution status.
	Send(ctx context.Context, data interface{}, isResolved bool, formatter Formatter) error

	// CheckResolveVariables checks if the notifier configuration is valid.
	CheckResolveVariables() error
}
