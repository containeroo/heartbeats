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

	SkipTLS   *bool  `yaml:"skip_tls"`             // SkipTLS skipt TLS check when doing the web request.
	Channel   string `yaml:"channel"`              // Slack channel
	Token     string `yaml:"token"`                // Slack API token
	Username  string `yaml:"username,omitempty"`   // display username
	TitleTmpl string `yaml:"title_tmpl,omitempty"` // title template
	TextTmpl  string `yaml:"text_tmpl,omitempty"`  // text template
}

// NewSlackNotifier creates a Slack notifier.
func NewSlackNotifier(id string, cfg SlackConfig, logger *slog.Logger, sender slack.Sender) Notifier {
	cfg.id = id
	cfg.logger = logger
	cfg.sender = sender
	return &cfg
}

// Type returns the type of the notifier
func (s *SlackConfig) Type() string        { return "slack" }
func (s *SlackConfig) Target() string      { return s.Channel }
func (s *SlackConfig) LastSent() time.Time { return s.lastSent }
func (s *SlackConfig) LastErr() error      { return s.lastErr }
func (s *SlackConfig) Format(data NotificationData) (NotificationData, error) {
	return formatNotification(data, s.TitleTmpl, s.TextTmpl, defaultSlackTitleTmpl, defaultSlackTextTmpl)
}

func (s *SlackConfig) Notify(ctx context.Context, data NotificationData) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	s.lastSent = time.Now()
	s.lastErr = nil

	formatted, err := s.Format(data)
	if err != nil {
		s.lastErr = err
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
		Channel:     s.Channel,
		Attachments: []slack.Attachment{attachment},
	}

	if _, err := s.sender.Send(ctx, payload); err != nil {
		s.lastErr = err
		return fmt.Errorf("send Slack notification: %w", err)
	}

	s.logger.Info("Notification sent",
		"receiver", s.id,
		"type", s.Type(),
		"channel", s.Target(),
	)
	return nil
}

// Resolve interpolates variables in the config.
func (s *SlackConfig) Resolve() error {
	var err error
	if s.Token, err = resolver.ResolveVariable(s.Token); err != nil {
		return fmt.Errorf("resolve token: %w", err)
	}
	if s.Channel, err = resolver.ResolveVariable(s.Channel); err != nil {
		return fmt.Errorf("resolve channel: %w", err)
	}
	if s.TitleTmpl, err = resolver.ResolveVariable(s.TitleTmpl); err != nil {
		return fmt.Errorf("resolve title template: %w", err)
	}
	if s.TextTmpl, err = resolver.ResolveVariable(s.TextTmpl); err != nil {
		return fmt.Errorf("resolve text template: %w", err)
	}
	return nil
}

// Validate ensures required fields are set.
func (s *SlackConfig) Validate() error {
	if s.Token == "" {
		return errors.New("token cannot be empty")
	}
	if s.Channel == "" {
		return errors.New("channel cannot be empty")
	}
	s.Channel = "#" + strings.TrimPrefix(s.Channel, "#")
	return nil
}
