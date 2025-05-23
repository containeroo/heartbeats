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
		return fmt.Errorf("cannot send email notification: %w", err)
	}

	en.logger.Info("Email notification sent", "receiver", en.id, "to", en.EmailDetails.To)
	return nil
}

// Format applies templates to subject and body fields.
func (en *EmailConfig) Format(data NotificationData) (NotificationData, error) {
	if en.EmailDetails.SubjectTmpl == "" {
		en.EmailDetails.SubjectTmpl = defaultEmailSubjectTmpl
	}
	subject, err := applyTemplate(en.EmailDetails.SubjectTmpl, data)
	if err != nil {
		return data, err
	}

	if en.EmailDetails.BodyTmpl == "" {
		en.EmailDetails.BodyTmpl = defaultEmailBodyTmpl
	}
	body, err := applyTemplate(en.EmailDetails.BodyTmpl, data)
	if err != nil {
		return data, err
	}

	data.Title = subject
	data.Message = body
	return data, nil
}

// Resolve interpolates all environment variables used in the configuration.
func (ec *EmailConfig) Resolve() error {
	resolve := func(in string) (string, error) {
		return resolver.ResolveVariable(in)
	}

	var err error
	if ec.SMTPConfig.Host, err = resolve(ec.SMTPConfig.Host); err != nil {
		return fmt.Errorf("failed to resolve SMTP host: %w", err)
	}
	if ec.SMTPConfig.From, err = resolve(ec.SMTPConfig.From); err != nil {
		return fmt.Errorf("failed to resolve SMTP from: %w", err)
	}
	if ec.SMTPConfig.From == "" {
		return fmt.Errorf("SMTP from must be specified")
	}
	if ec.SMTPConfig.Username, err = resolve(ec.SMTPConfig.Username); err != nil {
		return fmt.Errorf("failed to resolve SMTP username: %w", err)
	}
	if ec.SMTPConfig.Password, err = resolve(ec.SMTPConfig.Password); err != nil {
		return fmt.Errorf("failed to resolve SMTP password: %w", err)
	}

	resolveList := func(list []string) ([]string, error) {
		out := make([]string, 0, len(list))
		for _, item := range list {
			resolved, err := resolve(item)
			if err != nil {
				return nil, err
			}
			out = append(out, resolved)
		}
		return out, nil
	}

	if ec.EmailDetails.To, err = resolveList(ec.EmailDetails.To); err != nil {
		return fmt.Errorf("failed to resolve email recipient: %w", err)
	}
	if ec.EmailDetails.Cc, err = resolveList(ec.EmailDetails.Cc); err != nil {
		return fmt.Errorf("failed to resolve email cc: %w", err)
	}
	if ec.EmailDetails.Bcc, err = resolveList(ec.EmailDetails.Bcc); err != nil {
		return fmt.Errorf("failed to resolve email bcc: %w", err)
	}
	if ec.EmailDetails.SubjectTmpl, err = resolve(ec.EmailDetails.SubjectTmpl); err != nil {
		return fmt.Errorf("failed to resolve subject template: %w", err)
	}
	if ec.EmailDetails.BodyTmpl, err = resolve(ec.EmailDetails.BodyTmpl); err != nil {
		return fmt.Errorf("failed to resolve body template: %w", err)
	}

	return nil
}

// Validate checks the basic configuration for completeness.
func (ec *EmailConfig) Validate() error {
	if _, err := ec.Format(NotificationData{}); err != nil {
		return err
	}
	if ec.SMTPConfig.Host == "" || ec.SMTPConfig.Port == 0 {
		return fmt.Errorf("SMTP host and port must be specified")
	}
	if len(ec.EmailDetails.To) == 0 {
		return fmt.Errorf("at least one recipient must be specified")
	}
	return nil
}
