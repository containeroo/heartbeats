package notifier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/heartbeats/pkg/notify/email"
	"github.com/containeroo/resolver"
)

const (
	defaultEmailSubjectTmpl string = "[HEARTBEATS]: {{ .ID }} {{ upper .Status }}"
	defaultEmailBodyTmpl    string = "<b>Description:</b><br>{{ .Description }}.<br>Last bump: {{ ago .LastBump }}"
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
	To          []string `yaml:"to"`                         // List of recipients.
	CC          []string `yaml:"cc,omitempty"`               // List of CCs.
	BCC         []string `yaml:"bcc,omitempty"`              // List of BCCs.
	IsHTML      bool     `yaml:"is_html,omitempty"`          // Whether to send as HTML.
	SubjectTmpl string   `yaml:"subject_template,omitempty"` // Subject template.
	BodyTmpl    string   `yaml:"body_template,omitempty"`    // Body template.
}

// NewEmailNotifier creates a new EmailConfig notifier.
func NewEmailNotifier(id string, cfg EmailConfig, logger *slog.Logger, sender email.Sender) Notifier {
	return &EmailConfig{
		id:           id,
		SMTPConfig:   cfg.SMTPConfig,
		EmailDetails: cfg.EmailDetails,
		logger:       logger,
		sender:       sender,
	}
}

func (e *EmailConfig) Type() string        { return "email" }
func (e *EmailConfig) Target() string      { return FormatEmailRecipients(e.EmailDetails) }
func (e *EmailConfig) LastSent() time.Time { return e.lastSent }
func (e *EmailConfig) LastErr() error      { return e.lastErr }
func (ec *EmailConfig) Format(data NotificationData) (NotificationData, error) {
	return formatNotification(data, ec.EmailDetails.SubjectTmpl, ec.EmailDetails.BodyTmpl, defaultEmailSubjectTmpl, defaultEmailBodyTmpl)
}

// Notify formats and sends the email using the configured SMTP settings.
func (e *EmailConfig) Notify(ctx context.Context, data NotificationData) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	e.lastSent = time.Now()
	e.lastErr = nil

	formatted, err := e.Format(data)
	if err != nil {
		e.lastErr = err
		return fmt.Errorf("failed to format notification: %w", err)
	}

	msg := email.Message{
		To:      e.EmailDetails.To,
		Cc:      e.EmailDetails.CC,
		Bcc:     e.EmailDetails.BCC,
		IsHTML:  e.EmailDetails.IsHTML,
		Subject: formatted.Title,
		Body:    formatted.Message,
	}

	if err := e.sender.Send(ctx, msg); err != nil {
		e.lastErr = err
		return fmt.Errorf("cannot send Email notification: %w", err)
	}

	// Log only attributes that are relevant to the notification.
	attrs := []slog.Attr{
		slog.String("receiver", e.id),
		slog.String("type", e.Type()),
	}
	if len(e.EmailDetails.To) > 0 {
		attrs = append(attrs, slog.Any("to", e.EmailDetails.To))
	}
	if len(e.EmailDetails.CC) > 0 {
		attrs = append(attrs, slog.Any("cc", e.EmailDetails.CC))
	}
	if len(e.EmailDetails.BCC) > 0 {
		attrs = append(attrs, slog.Any("bcc", e.EmailDetails.BCC))
	}

	e.logger.LogAttrs(ctx, slog.LevelInfo, "Notification sent", attrs...)
	return nil
}

// Resolve interpolates variables in the config.
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
	if ec.EmailDetails.CC, err = resolver.ResolveSlice(ec.EmailDetails.CC); err != nil {
		return fmt.Errorf("failed to resolve email cc: %w", err)
	}
	if ec.EmailDetails.BCC, err = resolver.ResolveSlice(ec.EmailDetails.BCC); err != nil {
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

// Resolve interpolates env variables in the config.
func (ec *EmailConfig) Validate() error {
	if ec.SMTPConfig.Host == "" || ec.SMTPConfig.Port == 0 {
		return errors.New("SMTP host and port must be specified")
	}
	if ec.SMTPConfig.From == "" {
		return errors.New("SMTP from must be specified")
	}
	if _, err := ec.Format(NotificationData{}); err != nil {
		return err
	}
	if len(ec.EmailDetails.To) == 0 && len(ec.EmailDetails.CC) == 0 && len(ec.EmailDetails.BCC) == 0 {
		return errors.New("at least one recipient must be specified")
	}
	return nil
}
