package notifications

import (
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/msteams"
)

type MsteamsSettings struct {
	Name     string         `mapstructure:"name,omitempty"`
	Type     string         `mapstructure:"type,omitempty"`
	Enabled  bool           `mapstructure:"enabled"`
	Subject  string         `mapstructure:"subject,omitempty"`
	Message  string         `mapstructure:"message,omitempty"`
	Notifier *notify.Notify `mapstructure:"-,omitempty"`
	WebHooks []string       `mapstructure:"webhooks,omitempty"`
}

// GenerateMailService generates a mail service
func GenerateMsteamsService(webHooks []string) (*msteams.MSTeams, error) {

	mailService := msteams.New()
	mailService.AddReceivers(webHooks...)

	return mailService, nil
}
