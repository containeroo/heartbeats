package notifier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/pkg/notify/slack"
	"github.com/containeroo/resolver"
)

const (
	defaultSlackTitleTmpl string = "[{{ upper .Status }}] {{ .ID }}"
	defaultSlackTextTmpl  string = "{{ .ID }} is {{ .Status }} (last bump: {{ ago .LastBump }})"
)

// SlackConfig sends notifications to Slack.
type SlackConfig struct {
	id string `yaml:"-"` // config ID for logging

	lastSent time.Time    `yaml:"-"`
	lastErr  error        `yaml:"-"`
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
func (sn *SlackConfig) Target() string      { return sn.Channel }
func (sn *SlackConfig) LastSent() time.Time { return sn.lastSent }
func (sn *SlackConfig) LastErr() error      { return sn.lastErr }
func (sn *SlackConfig) Format(data NotificationData) (NotificationData, error) {
	return formatNotification(data, sn.TitleTmpl, sn.TextTmpl, defaultSlackTitleTmpl, defaultSlackTextTmpl)
}

func (sn *SlackConfig) Notify(ctx context.Context, data NotificationData) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	sn.lastSent = time.Now()
	sn.lastErr = nil

	formatted, err := sn.Format(data)
	if err != nil {
		sn.lastErr = err
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
		sn.lastErr = err
		return fmt.Errorf("send Slack notification: %w", err)
	}

	sn.logger.Info("Notification sent",
		"receiver", sn.id,
		"type", sn.Type(),
		"target", sn.Target(),
	)
	return nil
}

// Resolve interpolates variables in the config.
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

// Validate ensures required fields are set.
func (sn *SlackConfig) Validate() error {
	if sn.Token == "" {
		return errors.New("token cannot be empty")
	}
	if sn.Channel == "" {
		return errors.New("channel cannot be empty")
	}
	sn.Channel = "#" + strings.TrimPrefix(sn.Channel, "#")
	return nil
}
