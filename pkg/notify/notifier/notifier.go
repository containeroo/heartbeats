package notifier

import (
	"context"
)

// Formatter defines a function type for formatting notification content.
type Formatter func(content string, data interface{}, isResolved bool) (string, error)

// Notifier defines the interface for different notification types.
type Notifier interface {
	// Send triggers the notification with provided data and resolution status.
	//
	// Parameters:
	//   - ctx: Context for controlling the lifecycle of the notification sending.
	//   - data: The data to be included in the notification.
	//   - isResolved: Boolean indicating if the notification is a resolution.
	//   - formatter: Function to format the content.
	//
	// Returns:
	//   - error: An error if sending the notification fails.
	Send(ctx context.Context, data interface{}, isResolved bool, formatter Formatter) error

	// CheckResolveVariables checks if the configuration fields are resolvable.
	//
	// Returns:
	//   - error: An error if any of the configuration fields cannot be resolved.
	CheckResolveVariables() error

	// ValidateTemplate validates the notification templates against the provided data.
	//
	// Parameters:
	//   - data: The data to be injected into the templates for validation.
	//
	// Returns:
	//   - error: An error if the notification template cannot be validated.
	ValidateTemplate(data interface{}) error
}
