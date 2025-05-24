package notifier

import (
	"context"
	"time"
)

// ReceiverConfig holds receiver-specific configurations.
type ReceiverConfig struct {
	SlackConfigs   []SlackConfig   `yaml:"slack_configs,omitempty"`   // SlackConfigs is the list of Slack configurations.
	EmailConfigs   []EmailConfig   `yaml:"email_configs,omitempty"`   // EmailConfigs is the list of email configurations.
	MSTeamsConfigs []MSTeamsConfig `yaml:"msteams_configs,omitempty"` // MSTeamsConfigs is the list of MSTeams configurations.
}

// NotificationData is the payload for alerts.
type NotificationData struct {
	ID          string    `json:"id"`          // heartbeat ID
	Name        string    `json:"name"`        // human-friendly name
	Description string    `json:"description"` // heartbeat description
	LastBump    time.Time `json:"lastPing"`    // time of last ping
	Status      string    `json:"status"`      // current status
	Receivers   []string  `json:"receivers"`   // list of receiver IDs
	Title       string    `json:"title"`       // rendered notification title
	Message     string    `json:"message"`     // rendered notification body
}

// Notifier defines methods for sending notifications.
type Notifier interface {
	Notify(ctx context.Context, data NotificationData) error // Notify sends a notification.
	Format(data NotificationData) (NotificationData, error)  // Format formats the notification title and text.
	Validate() error                                         // Validate checks whether the notifier is correctly configured.
	Resolve() error                                          // Resolve performs any necessary resolution (e.g., secrets, tokens).
	LastErr() error                                          // LastError reports whether the last notification attempt succeeded.
	Type() string                                            // Type returns the notifier's type, e.g., "slack", "email", "teams".
	LastSent() time.Time                                     // LastSent returns the timestamp of the last notification attempt.
}
