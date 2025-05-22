package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/pkg/notify/slack"
	"github.com/containeroo/resolver"
)

const (
	defaultSlackTitleTmpl = "[{{ upper .Status }}] {{ .ID }}"
	defaultSlackTextTmpl  = "{{ .ID }} is {{ .Status }} (last Ping: {{ ago .LastBump }})"
)

// SlackConfig sends notifications to Slack.
type SlackConfig struct {
	id       string       `yaml:"-"` // config ID for logging
	lastSent time.Time    `yaml:"-"`
	success  *bool        `yaml:"-"`
	logger   *slog.Logger `yaml:"-"` // logger for internal events
	sender   slack.Sender `yaml:"-"` // Optional override for injecting a custom sender (used in tests)

	SkipTLS   *bool  `yaml:"skipTLS"`                 // SkipTLS skipt TLS check when doing the web request.
	Channel   string `yaml:"channel"`                 // Slack channel
	Token     string `yaml:"token"`                   // Slack API token
	Username  string `yaml:"username,omitempty"`      // display username
	TitleTmpl string `yaml:"titleTemplate,omitempty"` // title template
	TextTmpl  string `yaml:"textTemplate,omitempty"`  // text template
}

// NewSlackNotifier creates a Slack notifier.
func NewSlackNotifier(id string, cfg SlackConfig, logger *slog.Logger, sender slack.Sender) Notifier {
	cfg.id = id
	cfg.logger = logger
	cfg.sender = sender
	return &cfg
}

// Type returns the type of the notifier
func (sn *SlackConfig) Type() string        { return "slack" }
func (sn *SlackConfig) LastSent() time.Time { return sn.lastSent }
func (sn *SlackConfig) Successful() *bool   { return sn.success }

// Notify sends a Slack notification.
func (sn *SlackConfig) Notify(ctx context.Context, data NotificationData) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	formatted, err := sn.Format(data)
	if err != nil {
		return fmt.Errorf("format notification: %w", err)
	}

	status := common.HeartbeatState(formatted.Status)
	color := "danger"
	if status == common.HeartbeatStateActive || status == common.HeartbeatStateRecovered {
		color = "good"
	}

	attachment := slack.Attachment{
		Color: color,
		Title: formatted.Title,
		Text:  formatted.Message,
	}

	payload := slack.Slack{
		Channel:     sn.Channel,
		Attachments: []slack.Attachment{attachment},
	}

	if _, err := sn.sender.Send(ctx, payload); err != nil {
		sn.success = boolPtr(false)
		return fmt.Errorf("send slack notification: %w", err)
	}

	sn.lastSent = time.Now()
	sn.success = boolPtr(true)

	sn.logger.Info("slack notification sent", "receiver", sn.id, "channel", sn.Channel)
	return nil
}

// Format applies templates to produce a notification.
func (sn *SlackConfig) Format(data NotificationData) (NotificationData, error) {
	if sn.TitleTmpl == "" {
		sn.TitleTmpl = defaultSlackTitleTmpl
	}
	t, err := applyTemplate(sn.TitleTmpl, data)
	if err != nil {
		return data, err
	}
	if sn.TextTmpl == "" {
		sn.TextTmpl = defaultSlackTextTmpl
	}
	msg, err := applyTemplate(sn.TextTmpl, data)
	if err != nil {
		return data, err
	}
	data.Title = t
	data.Message = msg
	return data, nil
}

// Validate ensures required fields are set.
func (sn *SlackConfig) Validate() error {
	if sn.Channel == "" {
		return fmt.Errorf("channel cannot be empty")
	}
	sn.Channel = "#" + strings.TrimPrefix(sn.Channel, "#")
	return nil
}

// Resolve interpolates env variables in the config.
func (sn *SlackConfig) Resolve() error {
	var err error
	if sn.Token, err = resolver.ResolveVariable(sn.Token); err != nil {
		return fmt.Errorf("resolve token: %w", err)
	}
	if sn.Channel, err = resolver.ResolveVariable(sn.Channel); err != nil {
		return fmt.Errorf("resolve channel: %w", err)
	}
	if sn.TitleTmpl, err = resolver.ResolveVariable(sn.TitleTmpl); err != nil {
		return fmt.Errorf("resolve title template: %w", err)
	}
	if sn.TextTmpl, err = resolver.ResolveVariable(sn.TextTmpl); err != nil {
		return fmt.Errorf("resolve text template: %w", err)
	}
	return nil
}
