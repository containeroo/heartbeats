package notifier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/containeroo/heartbeats/pkg/notify/msteamsgraph"
	"github.com/containeroo/resolver"
)

const (
	defaultTeamsTitleTmpl string = "[{{ upper .Status }}] {{ .ID }}"
	defaultTeamsTextTmpl  string = "{{ .ID }} is {{ .Status }} (last Ping: {{ ago .LastBump }})"
)

// MSTeamsGraphConfig sends notifications to Microsoft Teams via Graph API.
type MSTeamsGraphConfig struct {
	id string `yaml:"-"`

	lastSent time.Time           `yaml:"-"`
	lastErr  error               `yaml:"-"`
	logger   *slog.Logger        `yaml:"-"`
	sender   msteamsgraph.Sender `yaml:"-"`

	SkipTLS   *bool  `yaml:"skipTLS"`
	Token     string `yaml:"token"`     // bearer token for Graph API
	TeamID    string `yaml:"teamID"`    // ID of the target team
	ChannelID string `yaml:"channelID"` // ID of the channel in the team
	TitleTmpl string `yaml:"titleTemplate,omitempty"`
	TextTmpl  string `yaml:"textTemplate,omitempty"`
}

// NewMSTeamsGraphNotifier constructs a new Teams Graph notifier.
func NewMSTeamsGraphNotifier(id string, cfg MSTeamsGraphConfig, logger *slog.Logger, sender msteamsgraph.Sender) Notifier {
	cfg.id = id
	cfg.logger = logger
	cfg.sender = sender
	return &cfg
}

func (m *MSTeamsGraphConfig) Type() string { return "msteamsgraph" }
func (m *MSTeamsGraphConfig) Target() string {
	return MaskedTail(m.TeamID, 4) + "/" + MaskedTail(m.ChannelID, 4)
}
func (m *MSTeamsGraphConfig) LastSent() time.Time { return m.lastSent }
func (m *MSTeamsGraphConfig) LastErr() error      { return m.lastErr }

func (m *MSTeamsGraphConfig) Format(data NotificationData) (NotificationData, error) {
	return formatNotification(data, m.TitleTmpl, m.TextTmpl, defaultTeamsTitleTmpl, defaultTeamsTextTmpl)
}

// Notify sends a channel message via Microsoft Graph API.
func (m *MSTeamsGraphConfig) Notify(ctx context.Context, data NotificationData) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	m.lastSent = time.Now()
	m.lastErr = nil

	formatted, err := m.Format(data)
	if err != nil {
		m.lastErr = err
		return fmt.Errorf("format notification: %w", err)
	}

	body := msteamsgraph.ItemBody{
		ContentType: "html",
		Content:     "<b>" + formatted.Title + "</b><br>" + formatted.Message,
	}
	msg := msteamsgraph.Message{
		Body: body,
	}

	if _, err := m.sender.SendChannel(ctx, m.TeamID, m.ChannelID, msg); err != nil {
		m.lastErr = err
		return fmt.Errorf("send MS Teams message: %w", err)
	}

	m.logger.Info("Notification sent",
		"receiver", m.id,
		"type", m.Type(),
		"team", MaskedTail(m.TeamID, 4),
		"channel", MaskedTail(m.ChannelID, 4),
	)
	return nil
}

// Resolve resolves templated fields and credentials using a variable resolver.
func (m *MSTeamsGraphConfig) Resolve() error {
	var err error
	if m.Token, err = resolver.ResolveVariable(m.Token); err != nil {
		return fmt.Errorf("resolve token: %w", err)
	}
	if m.TeamID, err = resolver.ResolveVariable(m.TeamID); err != nil {
		return fmt.Errorf("resolve teamID: %w", err)
	}
	if m.ChannelID, err = resolver.ResolveVariable(m.ChannelID); err != nil {
		return fmt.Errorf("resolve channelID: %w", err)
	}
	if m.TitleTmpl, err = resolver.ResolveVariable(m.TitleTmpl); err != nil {
		return fmt.Errorf("resolve title template: %w", err)
	}
	if m.TextTmpl, err = resolver.ResolveVariable(m.TextTmpl); err != nil {
		return fmt.Errorf("resolve text template: %w", err)
	}
	return nil
}

// Validate ensures required configuration fields are set correctly.
func (m *MSTeamsGraphConfig) Validate() error {
	if strings.TrimSpace(m.Token) == "" {
		return errors.New("token cannot be empty")
	}
	if strings.TrimSpace(m.TeamID) == "" {
		return errors.New("teamID cannot be empty")
	}
	if strings.TrimSpace(m.ChannelID) == "" {
		return errors.New("channelID cannot be empty")
	}
	return nil
}
