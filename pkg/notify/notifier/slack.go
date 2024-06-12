package notifier

import (
	"context"
	"fmt"
	"heartbeats/pkg/notify/resolve"
	"heartbeats/pkg/notify/services/slack"
	"heartbeats/pkg/notify/utils"
	"time"
)

const SlackType = "slack"

// SlackConfig holds the configuration for Slack notifications inside the config file.
type SlackConfig struct {
	Channel       string `yaml:"channel"`
	Token         string `yaml:"token"`
	Title         string `yaml:"title"`
	Text          string `yaml:"text"`
	ColorTemplate string `yaml:"colorTemplate"`
}

// SlackNotifier implements Notifier for sending Slack notifications.
type SlackNotifier struct {
	Config SlackConfig `yaml:"slack_config"`
}

// Send sends a Slack notification with the given data and resolution status.
//
// Parameters:
//   - ctx: Context for controlling the lifecycle of the notification sending.
//   - data: The data to be included in the notification.
//   - isResolved: Boolean indicating if the notification is a resolution.
//   - formatter: Function to format the title and text.
//
// Returns:
//   - error: An error if sending the notification fails.
func (s SlackNotifier) Send(ctx context.Context, data interface{}, isResolved bool, formatter Formatter) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	slackConfig, err := resolveSlackConfig(s.Config)
	if err != nil {
		return fmt.Errorf("cannot resolve Slack config. %w", err)
	}
	headers := map[string]string{
		"Authorization": "Bearer " + slackConfig.Token,
		"Content-Type":  "application/json",
	}

	slackClient := slack.New(headers, true)

	title, err := formatter(slackConfig.Title, data, isResolved)
	if err != nil {
		return fmt.Errorf("cannot format Slack title. %w", err)
	}

	text, err := formatter(slackConfig.Text, data, false)
	if err != nil {
		return fmt.Errorf("cannot format Slack message. %w", err)
	}

	color, err := determineColor(data, slackConfig.ColorTemplate)
	if err != nil {
		return fmt.Errorf("cannot determine Slack color. %w", err)
	}

	attachment := slack.Attachment{
		Text:  text,
		Color: color,
		Title: title,
	}

	slackItem := slack.Slack{
		Channel:     s.Config.Channel,
		Attachments: []slack.Attachment{attachment},
	}

	_, err = slackClient.Send(ctx, slackItem)
	if err != nil {
		return fmt.Errorf("cannot send Slack notification. %w", err)
	}

	return nil
}

// CheckResolveVariables checks if the configuration fields are resolvable.
//
// Returns:
//   - error: An error if any of the configuration fields cannot be resolved.
func (e SlackNotifier) CheckResolveVariables() error {
	if _, err := resolveSlackConfig(e.Config); err != nil {
		return err
	}

	return nil
}

// resolveSlackConfig resolves token, channel, title, and text.
//
// Parameters:
//   - config: The SlackConfig containing the raw configuration values.
//
// Returns:
//   - SlackConfig: The resolved SlackConfig.
//   - error: An error if any of the configuration values cannot be resolved.
func resolveSlackConfig(config SlackConfig) (SlackConfig, error) {
	token, err := resolve.ResolveVariable(config.Token)
	if err != nil {
		return SlackConfig{}, fmt.Errorf("cannot resolve Slack token. %w", err)
	}

	channel, err := resolve.ResolveVariable(config.Channel)
	if err != nil {
		return SlackConfig{}, fmt.Errorf("cannot resolve Slack channel. %w", err)
	}

	title, err := resolve.ResolveVariable(config.Title)
	if err != nil {
		return SlackConfig{}, fmt.Errorf("cannot resolve Slack title. %w", err)
	}

	text, err := resolve.ResolveVariable(config.Text)
	if err != nil {
		return SlackConfig{}, fmt.Errorf("cannot resolve Slack text. %w", err)
	}

	colorTemplate, err := resolve.ResolveVariable(config.ColorTemplate)
	if err != nil {
		return SlackConfig{}, fmt.Errorf("cannot resolve Slack color template. %w", err)
	}

	return SlackConfig{
		Token:         token,
		Channel:       channel,
		Title:         title,
		Text:          text,
		ColorTemplate: colorTemplate,
	}, nil
}

// String returns the type of the notifier.
func (s SlackNotifier) String() string {
	return SlackType
}

// determineColor determines the color for the Slack message based on the notification data.
func determineColor(data interface{}, template string) (string, error) {
	color, err := utils.FormatTemplate("determineColor", template, data)
	if err != nil {
		return "", fmt.Errorf("error determine color. %s", err)
	}

	return color, nil
}
