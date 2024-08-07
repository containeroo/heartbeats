package notifier

import (
	"context"
	"fmt"
	"heartbeats/pkg/notify/services/msteams"
	"heartbeats/pkg/notify/utils"
	"heartbeats/pkg/resolver"
	"time"
)

const MSTeamsType = "msteams"

// MSTeamsConfig holds the configuration for MS Teams notifications inside the config file.
type MSTeamsConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Title      string `yaml:"title"`
	Text       string `yaml:"text"`
}

// MSTeamsNotifier Notifier for sending MS Teams notifications.
type MSTeamsNotifier struct {
	Config MSTeamsConfig `yaml:"msteams_config"`
}

// Send sends an MS Teams notification with the given data and resolution status.
//
// Parameters:
//   - ctx: Context for controlling the lifecycle of the notification sending.
//   - data: The data to be included in the notification.
//   - isResolved: Boolean indicating if the notification is a resolution.
//   - formatter: Function to format the title and text.
//
// Returns:
//   - error: An error if sending the notification fails.
func (m *MSTeamsNotifier) Send(ctx context.Context, data interface{}, isResolved bool, formatter Formatter) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	msteamsClient := msteams.New(nil, true)

	msteamsConfig, err := resolveMSTeamsConfig(m.Config)
	if err != nil {
		return fmt.Errorf("cannot resolve MS Teams config. %w", err)
	}

	title, err := formatter(msteamsConfig.Title, data, isResolved)
	if err != nil {
		return fmt.Errorf("cannot format MS Teams title. %w", err)
	}

	text, err := formatter(msteamsConfig.Text, data, false)
	if err != nil {
		return fmt.Errorf("cannot format MS Teams message. %w", err)
	}

	msteamsMessage := msteams.MSTeams{
		Title: title,
		Text:  text,
	}

	if _, err := msteamsClient.Send(ctx, msteamsMessage, msteamsConfig.WebhookURL); err != nil {
		return fmt.Errorf("cannot send MS Teams notification. %w", err)
	}

	return nil
}

// ValidateTemplate validates the notification templates against the provided data.
//
// Parameters:
//   - data: The data to be injected into the templates for validation.
//
// Returns:
//   - error: An error if the notification template cannot be validated.
func (m *MSTeamsNotifier) ValidateTemplate(data interface{}) error {
	if _, err := utils.FormatTemplate("title", m.Config.Title, data); err != nil {
		return fmt.Errorf("cannot validate MS Teams title template. %s", err)
	}

	if _, err := utils.FormatTemplate("text", m.Config.Text, data); err != nil {
		return fmt.Errorf("cannot validate MS Teams text template. %s", err)
	}

	return nil
}

// CheckResolveVariables checks if the configuration fields are resolvable.
//
// Returns:
//   - error: An error if the configuration fields are not resolvable.
func (m *MSTeamsNotifier) CheckResolveVariables() error {
	if _, err := resolveMSTeamsConfig(m.Config); err != nil {
		return err
	}

	return nil
}

// resolveMSTeamsConfig resolves webookURL, title and text.
//
// Parameters:
//   - config: The MS Teams configuration to resolve.
//
// Returns:
//   - MSTeamsConfig: The resolved MS Teams configuration.
//   - error: An error if resolving any field fails.
func resolveMSTeamsConfig(config MSTeamsConfig) (MSTeamsConfig, error) {
	webhookURL, err := resolver.ResolveVariable(config.WebhookURL)
	if err != nil {
		return MSTeamsConfig{}, fmt.Errorf("cannot resolve webhook URL. %w", err)
	}

	title, err := resolver.ResolveVariable(config.Title)
	if err != nil {
		return MSTeamsConfig{}, fmt.Errorf("cannot resolve MS Teams title. %w", err)
	}

	text, err := resolver.ResolveVariable(config.Text)
	if err != nil {
		return MSTeamsConfig{}, fmt.Errorf("cannot resolve MS Teams text. %w", err)
	}

	return MSTeamsConfig{
		WebhookURL: webhookURL,
		Title:      title,
		Text:       text,
	}, nil
}

// String returns the type of the notifier.
func (m *MSTeamsNotifier) String() string {
	return MSTeamsType
}
