package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/heartbeats/pkg/notify/email"
	"github.com/containeroo/resolver"
)

const (
	defaultEmailSubjectTmpl = "[HEARTBEATS]: {{ .ID }} {{ upper .Status }}"
	defaultEmailBodyTmpl    = "<b>Description:</b><br>{{ .Description }}.<br>Last ping: {{ ago .LastBump }}"
)

// EmailConfig holds the configuration for sending email notifications.
type EmailConfig struct {
	id string `yaml:"-"` // id is the ID of the email configuration.

	lastSent time.Time    `yaml:"-"`
	lastErr  error        `yaml:"-"`
	logger   *slog.Logger `yaml:"-"`
	sender   email.Sender `yaml:"-"`

	SMTPConfig   email.SMTPConfig `yaml:"smtp"`  // SMTPConfig is the SMTP configuration.
	EmailDetails EmailDetails     `yaml:"email"` // EmailDetails is the email message details.
}

// EmailDetails is a simplified version of the email.Message used by heartbeats.
type EmailDetails struct {
	To          []string `yaml:"to"`                        // List of recipients.
	Cc          []string `yaml:"cc,omitempty"`              // List of CCs.
	Bcc         []string `yaml:"bcc,omitempty"`             // List of BCCs.
	IsHTML      bool     `yaml:"isHTML,omitempty"`          // Whether to send as HTML.
	SubjectTmpl string   `yaml:"subjectTemplate,omitempty"` // Subject template.
	BodyTmpl    string   `yaml:"bodyTemplate,omitempty"`    // Body template.
}

// NewEmailNotifier creates a new EmailConfig notifier.
func NewEmailNotifier(id string, cfg EmailConfig, logger *slog.Logger, sender email.Sender) Notifier {
	return &EmailConfig{
		id:           id,
		SMTPConfig:   cfg.SMTPConfig,
		EmailDetails: cfg.EmailDetails,
		logger:       logger,
	}
}

func (en *EmailConfig) Type() string        { return "email" }
func (en *EmailConfig) LastSent() time.Time { return en.lastSent }
func (en *EmailConfig) LastErr() error      { return en.lastErr }
func (ec *EmailConfig) Format(data NotificationData) (NotificationData, error) {
	return formatNotification(data, ec.EmailDetails.SubjectTmpl, ec.EmailDetails.BodyTmpl, defaultEmailSubjectTmpl, defaultEmailBodyTmpl)
}

// Notify formats and sends the email using the configured SMTP settings.
func (en *EmailConfig) Notify(ctx context.Context, data NotificationData) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	en.lastSent = time.Now()
	en.lastErr = nil

	formatted, err := en.Format(data)
	if err != nil {
		en.lastErr = err
		return fmt.Errorf("failed to format notification: %w", err)
	}

	msg := email.Message{
		To:      en.EmailDetails.To,
		Cc:      en.EmailDetails.Cc,
		Bcc:     en.EmailDetails.Bcc,
		IsHTML:  en.EmailDetails.IsHTML,
		Subject: formatted.Title,
		Body:    formatted.Message,
	}

	if err := en.sender.Send(ctx, msg); err != nil {
		en.lastErr = err
		return fmt.Errorf("cannot send Email notification: %w", err)
	}

	en.logger.Info("Email notification sent", "receiver", en.id, "to", en.EmailDetails.To)
	return nil
}

func (ec *EmailConfig) Resolve() error {
	var err error
	if ec.SMTPConfig.Host, err = resolver.ResolveVariable(ec.SMTPConfig.Host); err != nil {
		return fmt.Errorf("failed to resolve SMTP host: %w", err)
	}
	if ec.SMTPConfig.From, err = resolver.ResolveVariable(ec.SMTPConfig.From); err != nil {
		return fmt.Errorf("failed to resolve SMTP from: %w", err)
	}
	if ec.SMTPConfig.Username, err = resolver.ResolveVariable(ec.SMTPConfig.Username); err != nil {
		return fmt.Errorf("failed to resolve SMTP username: %w", err)
	}
	if ec.SMTPConfig.Password, err = resolver.ResolveVariable(ec.SMTPConfig.Password); err != nil {
		return fmt.Errorf("failed to resolve SMTP password: %w", err)
	}

	if ec.EmailDetails.To, err = resolver.ResolveSlice(ec.EmailDetails.To); err != nil {
		return fmt.Errorf("failed to resolve email recipient: %w", err)
	}
	if ec.EmailDetails.Cc, err = resolver.ResolveSlice(ec.EmailDetails.Cc); err != nil {
		return fmt.Errorf("failed to resolve email cc: %w", err)
	}
	if ec.EmailDetails.Bcc, err = resolver.ResolveSlice(ec.EmailDetails.Bcc); err != nil {
		return fmt.Errorf("failed to resolve email bcc: %w", err)
	}
	if ec.EmailDetails.SubjectTmpl, err = resolver.ResolveVariable(ec.EmailDetails.SubjectTmpl); err != nil {
		return fmt.Errorf("failed to resolve subject template: %w", err)
	}
	if ec.EmailDetails.BodyTmpl, err = resolver.ResolveVariable(ec.EmailDetails.BodyTmpl); err != nil {
		return fmt.Errorf("failed to resolve body template: %w", err)
	}

	return nil
}

func (ec *EmailConfig) Validate() error {
	if ec.SMTPConfig.Host == "" || ec.SMTPConfig.Port == 0 {
		return fmt.Errorf("SMTP host and port must be specified")
	}
	if ec.SMTPConfig.From == "" {
		return fmt.Errorf("SMTP from must be specified")
	}
	if _, err := ec.Format(NotificationData{}); err != nil {
		return err
	}
	if len(ec.EmailDetails.To) == 0 && len(ec.EmailDetails.Cc) == 0 && len(ec.EmailDetails.Bcc) == 0 {
		return fmt.Errorf("at least one recipient must be specified")
	}
	return nil
}
