package notifications

import (
	"fmt"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/slack"
)

type SlackSettings struct {
	Name        string         `mapstructure:"name,omitempty"`
	Type        string         `mapstructure:"type,omitempty"`
	Enabled     *bool          `mapstructure:"enabled,omitempty"`
	SendResolve *bool          `mapstructure:"sendResolve,omitempty"`
	Notifier    *notify.Notify `mapstructure:"-,omitempty" deepcopier:"skip"`
	Subject     string         `mapstructure:"subject,omitempty"`
	Message     string         `mapstructure:"message,omitempty"`
	OauthToken  string         `mapstructure:"oauthToken,omitempty" redacted:"true"`
	Channels    []string       `mapstructure:"channels,omitempty"`
}

func GenerateSlackService(token string, channels []string) (*slack.Slack, error) {
	if token == "" {
		return nil, fmt.Errorf("Token is empty")
	}
	if len(channels) == 0 {
		return nil, fmt.Errorf("Channels are empty")
	}

	slackService := slack.New(token)
	slackService.AddReceivers(channels...)

	return slackService, nil
}
