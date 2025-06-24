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

func (c *MSTeamsGraphConfig) Type() string        { return "msteamsgraph" }
func (c *MSTeamsGraphConfig) Target() string      { return c.TeamID + "/" + c.ChannelID }
func (c *MSTeamsGraphConfig) LastSent() time.Time { return c.lastSent }
func (c *MSTeamsGraphConfig) LastErr() error      { return c.lastErr }

func (c *MSTeamsGraphConfig) Format(data NotificationData) (NotificationData, error) {
	return formatNotification(data, c.TitleTmpl, c.TextTmpl, defaultTeamsTitleTmpl, defaultTeamsTextTmpl)
}

// Notify sends a channel message via Microsoft Graph API.
func (c *MSTeamsGraphConfig) Notify(ctx context.Context, data NotificationData) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	c.lastSent = time.Now()
	c.lastErr = nil

	formatted, err := c.Format(data)
	if err != nil {
		c.lastErr = err
		return fmt.Errorf("format notification: %w", err)
	}

	body := msteamsgraph.ItemBody{
		ContentType: "html",
		Content:     "<b>" + formatted.Title + "</b><br>" + formatted.Message,
	}
	msg := msteamsgraph.Message{
		Body: body,
	}

	if _, err := c.sender.SendChannel(ctx, c.TeamID, c.ChannelID, msg); err != nil {
		c.lastErr = err
		return fmt.Errorf("send MS Teams message: %w", err)
	}

	c.logger.Info("Notification sent",
		"receiver", c.id,
		"type", c.Type(),
		"target", c.Target(),
	)
	return nil
}

// Resolve resolves templated fields and credentials using a variable resolver.
func (c *MSTeamsGraphConfig) Resolve() error {
	var err error
	if c.Token, err = resolver.ResolveVariable(c.Token); err != nil {
		return fmt.Errorf("resolve token: %w", err)
	}
	if c.TeamID, err = resolver.ResolveVariable(c.TeamID); err != nil {
		return fmt.Errorf("resolve teamID: %w", err)
	}
	if c.ChannelID, err = resolver.ResolveVariable(c.ChannelID); err != nil {
		return fmt.Errorf("resolve channelID: %w", err)
	}
	if c.TitleTmpl, err = resolver.ResolveVariable(c.TitleTmpl); err != nil {
		return fmt.Errorf("resolve title template: %w", err)
	}
	if c.TextTmpl, err = resolver.ResolveVariable(c.TextTmpl); err != nil {
		return fmt.Errorf("resolve text template: %w", err)
	}
	return nil
}

// Validate ensures required configuration fields are set correctly.
func (c *MSTeamsGraphConfig) Validate() error {
	if strings.TrimSpace(c.Token) == "" {
		return errors.New("token cannot be empty")
	}
	if strings.TrimSpace(c.TeamID) == "" {
		return errors.New("teamID cannot be empty")
	}
	if strings.TrimSpace(c.ChannelID) == "" {
		return errors.New("channelID cannot be empty")
	}
	return nil
}
