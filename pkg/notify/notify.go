package notify

import (
	"context"
	"fmt"
	"heartbeats/pkg/notify/notifier"
	"heartbeats/pkg/notify/utils"
	"sync"
)

// Notification represents a single notification configuration, including the type (Slack, Email, etc.)
// and its enabled status.
type Notification struct {
	Name          string                  `yaml:"-"`
	Enabled       *bool                   `yaml:"enabled,omitempty"`
	Type          string                  `yaml:"type,omitempty"`
	SlackConfig   *notifier.SlackConfig   `yaml:"slack_config,omitempty"`
	MailConfig    *notifier.MailConfig    `yaml:"mail_config,omitempty"`
	MSTeamsConfig *notifier.MSTeamsConfig `yaml:"msteams_config,omitempty"`
	Notifier      notifier.Notifier       `yaml:"notifier,omitempty"`
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
		return fmt.Errorf("notification '%s' already exists", name)
	}

	instance, err := evaluateNotifier(notification)
	if err != nil {
		return err
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

// Update updates an existing notification in the store.
//
// Parameters:
//   - name: The name of the notification to update.
//   - notification: The Notification object containing the updated configuration.
//
// Returns:
//   - error: An error if the notification does not exist.
func (s *Store) Update(name string, notification *Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if found := s.notifications[name]; found == nil {
		return fmt.Errorf("notification '%s' does not exist", name)
	}

	instance, err := evaluateNotifier(notification)
	if err != nil {
		return err
	}

	notification.Notifier = instance
	s.notifications[name] = notification
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

// evaluateNotifier evaluates and returns the appropriate Notifier instance based on the notification configuration.
//
// Parameters:
//   - notification: The Notification object containing the configuration details.
//
// Returns:
//   - notifier.Notifier: The Notifier instance based on the configuration.
//   - error: An error if the configuration is missing or unsupported.
func evaluateNotifier(notification *Notification) (notifier.Notifier, error) {
	switch {
	case notification.SlackConfig != nil:
		return &notifier.SlackNotifier{Config: *notification.SlackConfig}, nil
	case notification.MailConfig != nil:
		return &notifier.EmailNotifier{Config: *notification.MailConfig}, nil
	case notification.MSTeamsConfig != nil:
		return &notifier.MSTeamsNotifier{Config: *notification.MSTeamsConfig}, nil
	default:
		return nil, fmt.Errorf("notification configuration is missing or unsupported")
	}
}
