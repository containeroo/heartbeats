package notifier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/containeroo/heartbeats/pkg/notify/msteams"

	"github.com/containeroo/resolver"
)

const (
	defaultMSTeamsTitleTmpl string = "[{{ upper .Status }}] {{ .ID }}"
	defaultMSTeamsTextTmpl  string = "{{ .ID }} is {{ .Status }} (last bump: {{ ago .LastBump }})"
)

// MSTeamsConfig sends notifications to MSTeams.
type MSTeamsConfig struct {
	id string `yaml:"-"` // id is the ID of the email configuration.

	lastSent time.Time      `yaml:"-"`
	lastErr  error          `yaml:"-"`
	logger   *slog.Logger   `yaml:"-"` // logger is the logger for logging events.
	sender   msteams.Sender `yaml:"-"` // Optional override for injecting a custom sender (used in tests)

	WebhookURL string `yaml:"webhook_url"`             // WebhookURL is the webhook URL for the MSTeams webhook.
	SkipTLS    *bool  `yaml:"skipTLS"`                 // SkipTLS skipt TLS check when doing the web request.
	TitleTmpl  string `yaml:"titleTemplate,omitempty"` // TitleTmpl is the title template for the notification.
	TextTmpl   string `yaml:"textTemplate,omitempty"`  // TextTmpl is the text template for the notification.
}

// NewMSTeamsNotifier creates a new MSTeamsNotifier for a single MSTeams configuration.
func NewMSTeamsNotifier(id string, cfg MSTeamsConfig, logger *slog.Logger, sender msteams.Sender) *MSTeamsConfig {
	cfg.id = id
	cfg.logger = logger
	cfg.sender = sender
	return &cfg
}

func (m *MSTeamsConfig) Type() string        { return "msteams" }
func (m *MSTeamsConfig) Target() string      { return MasqueradeURL(m.WebhookURL, 4) }
func (m *MSTeamsConfig) LastSent() time.Time { return m.lastSent }
func (m *MSTeamsConfig) LastErr() error      { return m.lastErr }
func (m *MSTeamsConfig) Format(data NotificationData) (NotificationData, error) {
	return formatNotification(data, m.TitleTmpl, m.TextTmpl, defaultMSTeamsTitleTmpl, defaultMSTeamsTextTmpl)
}

func (m *MSTeamsConfig) Notify(ctx context.Context, data NotificationData) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	m.lastSent = time.Now()
	m.lastErr = nil

	formatted, err := m.Format(data)
	if err != nil {
		m.lastErr = err
		return fmt.Errorf("failed to format notification: %w", err)
	}

	msg := msteams.MSTeams{
		Title: formatted.Title,
		Text:  formatted.Message,
	}

	if _, err := m.sender.Send(ctx, msg, m.WebhookURL); err != nil {
		m.lastErr = err
		return fmt.Errorf("cannot send MSTeams notification. %w", err)
	}

	m.logger.Info("Notification sent",
		"receiver", m.id,
		"type", m.Type(),
		"webhook_url", m.Target(),
	)
	return nil
}

// Resolve interpolates variables in the config.
func (m *MSTeamsConfig) Resolve() error {
	webhookURL, err := resolver.ResolveVariable(m.WebhookURL)
	if err != nil {
		return fmt.Errorf("failed to resolve WebhookURL: %w", err)
	}
	m.WebhookURL = webhookURL

	titleTmpl, err := resolver.ResolveVariable(m.TitleTmpl)
	if err != nil {
		return fmt.Errorf("failed to resolve title template: %w", err)
	}
	m.TitleTmpl = titleTmpl

	textTmpl, err := resolver.ResolveVariable(m.TextTmpl)
	if err != nil {
		return fmt.Errorf("failed to resolve text template: %w", err)
	}
	m.TextTmpl = textTmpl

	return nil
}

// Validate ensures required fields are set.
func (m *MSTeamsConfig) Validate() error {
	dummydata := NotificationData{}
	if _, err := m.Format(dummydata); err != nil {
		return err
	}
	if m.WebhookURL == "" {
		return errors.New("webhook URL cannot be empty")
	}
	if _, err := url.Parse(m.WebhookURL); err != nil {
		return fmt.Errorf("webhook URL is not a valid URL: %w", err)
	}
	return nil
}
