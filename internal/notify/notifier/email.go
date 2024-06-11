package notifier

import (
	"context"
	"fmt"
	"heartbeats/internal/notify/resolve"
	"heartbeats/internal/notify/services/email"
	"time"
)

const EmailType = "email"

// MailConfig holds configuration settings for email notifications.
type MailConfig struct {
	SMTP  email.SMTPConfig `mapstructure:"smtp"`
	Email email.Email      `mapstructure:"email"`
}

// EmailNotifier implements Notifier for sending email notifications.
type EmailNotifier struct {
	Config MailConfig
}

// Send sends an email notification with the given data and resolution status.
//
// Parameters:
//   - ctx: Context for controlling the lifecycle of the notification sending.
//   - data: The data to be included in the notification.
//   - isResolved: Boolean indicating if the notification is a resolution.
//   - formatter: Function to format the subject and body.
//
// Returns:
//   - error: An error if sending the notification fails.
func (e EmailNotifier) Send(ctx context.Context, data interface{}, isResolved bool, formatter Formatter) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	smtpConfig, err := resolveSMTPConfig(e.Config.SMTP)
	if err != nil {
		return fmt.Errorf("cannot resolve SMTP config. %w", err)
	}

	emailConfig, err := resolveEmailConfig(e.Config.Email)
	if err != nil {
		return fmt.Errorf("cannot resolve email config. %w", err)
	}

	mailClient := email.New(smtpConfig)
	emailConfig.Subject, err = formatter(emailConfig.Subject, data, isResolved)
	if err != nil {
		return fmt.Errorf("cannot format email subject. %w", err)
	}
	emailConfig.Body, err = formatter(emailConfig.Body, data, false)
	if err != nil {
		return fmt.Errorf("cannot format email body. %w", err)
	}

	if err := mailClient.Send(ctx, emailConfig); err != nil {
		return fmt.Errorf("cannot send email notification. %w", err)
	}

	return nil
}

// CheckResolveVariables checks if the configuration fields are resolvable.
//
// Returns:
//   - error: An error if the configuration fields are not resolvable.
func (e EmailNotifier) CheckResolveVariables() error {
	if _, err := resolveSMTPConfig(e.Config.SMTP); err != nil {
		return fmt.Errorf("cannot resolve SMTP config. %w", err)
	}

	if _, err := resolveEmailConfig(e.Config.Email); err != nil {
		return fmt.Errorf("cannot resolve email config. %w", err)
	}

	return nil
}

// resolveSMTPConfig resolves host, from, username and password.
//
// Parameters:
//   - config: The SMTP configuration to resolve.
//
// Returns:
//   - email.SMTPConfig: The resolved SMTP configuration.
//   - error: An error if resolving any field fails.
func resolveSMTPConfig(config email.SMTPConfig) (email.SMTPConfig, error) {
	var err error
	config.Host, err = resolve.ResolveVariable(config.Host)
	if err != nil {
		return email.SMTPConfig{}, err
	}
	config.From, err = resolve.ResolveVariable(config.From)
	if err != nil {
		return email.SMTPConfig{}, err
	}
	config.Username, err = resolve.ResolveVariable(config.Username)
	if err != nil {
		return email.SMTPConfig{}, err
	}
	config.Password, err = resolve.ResolveVariable(config.Password)
	if err != nil {
		return email.SMTPConfig{}, err
	}
	return config, nil
}

// resolveEmailConfig resolves to, cc, bcc, subject and body.
//
// Parameters:
//   - config: The email configuration to resolve.
//
// Returns:
//   - email.Email: The resolved email configuration.
//   - error: An error if resolving any field fails.
func resolveEmailConfig(config email.Email) (email.Email, error) {
	var err error

	for i, to := range config.To {
		config.To[i], err = resolve.ResolveVariable(to)
		if err != nil {
			return email.Email{}, err
		}
	}

	for i, cc := range config.Cc {
		config.Cc[i], err = resolve.ResolveVariable(cc)
		if err != nil {
			return email.Email{}, err
		}
	}

	for i, bcc := range config.Bcc {
		config.Bcc[i], err = resolve.ResolveVariable(bcc)
		if err != nil {
			return email.Email{}, err
		}
	}

	config.Subject, err = resolve.ResolveVariable(config.Subject)
	if err != nil {
		return email.Email{}, err
	}

	config.Body, err = resolve.ResolveVariable(config.Body)
	if err != nil {
		return email.Email{}, err
	}

	return config, nil
}

// String returns the type of the notifier.
func (e EmailNotifier) String() string {
	return EmailType
}
