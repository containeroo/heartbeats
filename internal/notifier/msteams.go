package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/heartbeats/pkg/notify/msteams"

	"github.com/containeroo/resolver"
)

const (
	defaultMSTeamsTitleTmpl string = "[{{ upper .Status }}] {{ .ID }}"
	defaultMSTeamsTextTmpl  string = "{{ .ID }} is {{ .Status }} (last Ping: {{ ago .LastBump }})"
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

func (mc *MSTeamsConfig) Type() string        { return "msteams" }
func (mc *MSTeamsConfig) LastSent() time.Time { return mc.lastSent }
func (mc *MSTeamsConfig) LastErr() error      { return mc.lastErr }
func (mc *MSTeamsConfig) Format(data NotificationData) (NotificationData, error) {
	return formatNotification(data, mc.TitleTmpl, mc.TextTmpl, defaultMSTeamsTitleTmpl, defaultMSTeamsTextTmpl)
}

func (mc *MSTeamsConfig) Notify(ctx context.Context, data NotificationData) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	mc.lastSent = time.Now()
	mc.lastErr = nil

	formatted, err := mc.Format(data)
	if err != nil {
		mc.lastErr = err
		return fmt.Errorf("failed to format notification: %w", err)
	}

	msg := msteams.MSTeams{
		Title: formatted.Title,
		Text:  formatted.Message,
	}

	if _, err := mc.sender.Send(ctx, msg, mc.WebhookURL); err != nil {
		mc.lastErr = err
		return fmt.Errorf("cannot send MSTeams notification. %w", err)
	}

	mc.logger.Info("MSTeams notification sent", "receiver", mc.id)
	return nil
}

func (mc *MSTeamsConfig) Resolve() error {
	webhookURL, err := resolver.ResolveVariable(mc.WebhookURL)
	if err != nil {
		return fmt.Errorf("failed to resolve WebhookURL: %w", err)
	}
	mc.WebhookURL = webhookURL

	titleTmpl, err := resolver.ResolveVariable(mc.TitleTmpl)
	if err != nil {
		return fmt.Errorf("failed to resolve title template: %w", err)
	}
	mc.TitleTmpl = titleTmpl

	textTmpl, err := resolver.ResolveVariable(mc.TextTmpl)
	if err != nil {
		return fmt.Errorf("failed to resolve text template: %w", err)
	}
	mc.TextTmpl = textTmpl

	return nil
}

func (mc *MSTeamsConfig) Validate() error {
	dummydata := NotificationData{}
	if _, err := mc.Format(dummydata); err != nil {
		return err
	}
	if mc.WebhookURL == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}
	return nil
}
