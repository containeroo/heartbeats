package notifications

import (
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/msteams"
)

type MsteamsSettings struct {
	Name        string         `mapstructure:"name,omitempty"`
	Type        string         `mapstructure:"type,omitempty"`
	Enabled     *bool          `mapstructure:"enabled,omitempty"`
	SendResolve *bool          `mapstructure:"sendResolve,omitempty"`
	Subject     string         `mapstructure:"subject,omitempty"`
	Message     string         `mapstructure:"message,omitempty"`
	Notifier    *notify.Notify `mapstructure:"-,omitempty" deepcopier:"skip"`
	WebHooks    []string       `mapstructure:"webhooks,omitempty" redacted:"true"`
}

// GenerateMailService generates a mail service
func GenerateMsteamsService(webHooks []string) (*msteams.MSTeams, error) {

	MSTeams := msteams.New()
	MSTeams.AddReceivers(webHooks...)

	return MSTeams, nil
}
