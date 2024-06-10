package notify

import (
	"context"
	"fmt"
	"heartbeats/internal/notify/notifier"
	"heartbeats/internal/notify/utils"
	"sync"
)

// Notification represents a single notification configuration, including the type (Slack, Email, etc.)
// and its enabled status.
type Notification struct {
	Name          string                  `mapstructure:"name" yaml:"-"`
	Enabled       *bool                   `mapstructure:"enabled,omitempty" yaml:"enabled,omitempty"`
	Type          string                  `mapstructure:"type,omitempty" yaml:"type,omitempty"`
	SlackConfig   *notifier.SlackConfig   `mapstructure:"slack_config,omitempty" yaml:"slack_config,omitempty"`
	MailConfig    *notifier.MailConfig    `mapstructure:"mail_config,omitempty" yaml:"mail_config,omitempty"`
	MSTeamsConfig *notifier.MSTeamsConfig `mapstructure:"msteams_config,omitempty" yaml:"msteams_config,omitempty"`
	Notifier      notifier.Notifier       `mapstructure:"notifier,omitempty" yaml:"notifier,omitempty"`
}

// String returns the name of the notification.
func (n *Notification) String() string {
	return n.Name
}

// Store manages the storage and retrieval of notifications.
type Store struct {
	notifications map[string]*Notification
	mu            sync.RWMutex
}

// MarshalYAML implements the yaml.Marshaler interface.
func (s *Store) MarshalYAML() (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.notifications, nil
}

// NewStore creates a new Store.
//
// Returns:
//   - *Store: A new instance of the Store.
func NewStore() *Store {
	return &Store{
		notifications: make(map[string]*Notification),
	}
}

// Send sends the notification with the given data and resolution status.
//
// Parameters:
//   - ctx: Context for controlling the lifecycle of the notification sending.
//   - data: The data to be included in the notification.
//   - isResolved: Boolean indicating if the notification is a resolution.
//   - formatter: Function to format the subject and body.
//
// Returns:
//   - error: An error if sending the notification fails.
func (n *Notification) Send(ctx context.Context, data interface{}, isResolved bool, formatter notifier.Formatter) error {
	return n.Notifier.Send(ctx, data, isResolved, formatter)
}

// CheckResolveVariables checks if the notification configuration variables are resolvable.
//
// Returns:
//   - error: An error if the notification configuration variables are not resolvable.
func (n *Notification) CheckResolveVariables() error {
	return n.Notifier.CheckResolveVariables()
}

// GetAll retrieves a copy of all notifications in the store.
//
// Returns:
//   - map[string]*Notification: A map of all notifications in the store.
func (s *Store) GetAll() map[string]*Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.notifications
}

// Get retrieves a notification by name.
//
// Parameters:
//   - name: The name of the notification to retrieve.
//
// Returns:
//   - *Notification: The notification object if found, otherwise nil.
func (s *Store) Get(name string) *Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.notifications[name]
}

// Add adds a notification to the store.
//
// Parameters:
//   - name: The name of the notification to add.
//   - notification: The Notification object containing the configuration details.
//
// Returns:
//   - error: An error if the notification already exists, has missing/unsupported configuration, or fails validation.
func (s *Store) Add(name string, notification *Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if found := s.notifications[name]; found != nil {
		return fmt.Errorf("notification '%s' already exists.", notification.Name)
	}

	var instance notifier.Notifier

	switch {
	case notification.SlackConfig != nil:
		instance = &notifier.SlackNotifier{Config: *notification.SlackConfig}
	case notification.MailConfig != nil:
		instance = &notifier.EmailNotifier{Config: *notification.MailConfig}
	case notification.MSTeamsConfig != nil:
		instance = &notifier.MSTeamsNotifier{Config: *notification.MSTeamsConfig}
	default:
		return fmt.Errorf("notification configuration is missing or unsupported.")
	}

	notification.Name = name
	notification.Type = fmt.Sprint(instance)
	notification.Notifier = instance
	s.notifications[name] = notification

	if notification.Enabled != nil && *notification.Enabled {
		if err := notification.CheckResolveVariables(); err != nil {
			return err
		}
	}

	return nil
}

// DefaultFormatter expands the content with environment variables, formats the notification content, appending [RESOLVED] if needed.
func DefaultFormatter(content string, data interface{}, isResolved bool) (string, error) {
	formattedContent, err := utils.FormatTemplate("default", content, data)
	if err != nil {
		return "", err
	}

	if isResolved {
		formattedContent = "[RESOLVED] " + formattedContent
	}

	return formattedContent, nil
}
