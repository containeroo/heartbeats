package notifications

import (
	"fmt"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/mail"
)

type MailSettings struct {
	Name              string         `mapstructure:"name,omitempty"`
	Type              string         `mapstructure:"type,omitempty"`
	Enabled           bool           `mapstructure:"enabled"`
	Subject           string         `mapstructure:"subject,omitempty"`
	Message           string         `mapstructure:"message,omitempty"`
	Notifier          *notify.Notify `mapstructure:"-,omitempty" deepcopier:"skip"`
	SenderAddress     string         `mapstructure:"senderAddress,omitempty"`
	SmtpHostAddr      string         `mapstructure:"smtpHostAddr,omitempty"`
	SmtpHostPort      int            `mapstructure:"smtpHostPort,omitempty"`
	SmtpAuthUser      string         `mapstructure:"smtpAuthUser,omitempty"`
	SmtpAuthPassword  string         `mapstructure:"smtpAuthPassword,omitempty" redacted:"true"`
	ReceiverAddresses []string       `mapstructure:"receiverAddresses,omitempty"`
}

// GenerateMailService generates a mail service
func GenerateMailService(senderAddress string, smtpHostAddr string, smtpHostPort int, smtpAuthUser string, smtpAuthPassword string, receiverAddresses []string) (*mail.Mail, error) {

	mailService := mail.New(senderAddress, fmt.Sprintf("%s:%d", smtpHostAddr, smtpHostPort))
	if smtpAuthUser != "" && smtpAuthPassword != "" {
		mailService.AuthenticateSMTP("", smtpAuthUser, smtpAuthPassword, smtpHostAddr)
	}
	mailService.AddReceivers(receiverAddresses...)

	return mailService, nil
}
