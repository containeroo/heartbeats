package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/pkg/notify/slack"
	"github.com/stretchr/testify/assert"
)

// mockSlackSender implements slack.Sender
type mockSlackSender struct {
	called  bool
	payload slack.Slack
	err     error
}

func (m *mockSlackSender) Send(ctx context.Context, msg slack.Slack) (*slack.Response, error) {
	m.called = true
	m.payload = msg
	return &slack.Response{Ok: true}, m.err
}

func TestSlackConfig_Type(t *testing.T) {
	t.Parallel()
	t.Run("returns slack", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "slack", NewSlackNotifier("id", SlackConfig{}, nil, nil).Type())
	})
}

func TestSlackConfig_LastSent(t *testing.T) {
	t.Parallel()
	t.Run("returns last sent time", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, time.Time{}, NewSlackNotifier("id", SlackConfig{}, nil, nil).LastSent())
	})
}

func TestSlackConfig_LastErr(t *testing.T) {
	t.Parallel()
	t.Run("returns last error", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, NewSlackNotifier("id", SlackConfig{}, nil, nil).LastErr())
	})
}

func TestSlackConfig_Notify(t *testing.T) {
	t.Parallel()

	t.Run("fails on invalid template", func(t *testing.T) {
		t.Parallel()

		mock := &mockSlackSender{}

		config := &SlackConfig{
			logger:   slog.New(slog.NewTextHandler(os.Stdout, nil)),
			sender:   mock,
			TextTmpl: "{{ .Missing }}",
		}

		now := time.Now()
		data := NotificationData{
			ID:       "db-check",
			Status:   "failed",
			LastBump: now.Add(-2 * time.Minute),
			Message:  "",
			Title:    "",
		}

		err := config.Notify(context.Background(), data)
		assert.EqualError(t, err, "format notification: format message: template: notification:1:3: executing \"notification\" at <.Missing>: can't evaluate field Missing in type notifier.NotificationData")
	})

	t.Run("Sender returns error", func(t *testing.T) {
		t.Parallel()

		mock := &mockSlackSender{err: fmt.Errorf("mock error")}

		config := &SlackConfig{
			logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
			sender: mock,
		}

		now := time.Now()
		data := NotificationData{
			ID:       "db-check",
			Status:   "failed",
			LastBump: now.Add(-2 * time.Minute),
			Message:  "",
			Title:    "",
		}

		err := config.Notify(context.Background(), data)
		assert.EqualError(t, err, "send Slack notification: mock error")
	})

	t.Run("sends notification", func(t *testing.T) {
		t.Parallel()

		mock := &mockSlackSender{}

		cfg := &SlackConfig{
			Token:   "xoxb-mock",
			Channel: "#devops",
			logger:  slog.New(slog.NewTextHandler(os.Stdout, nil)),
			sender:  mock,
		}

		data := NotificationData{
			ID:          "api-check",
			Status:      common.HeartbeatStateRecovered.String(),
			Description: "Heartbeat failed",
			LastBump:    time.Now().Add(-10 * time.Minute),
		}

		err := cfg.Notify(context.Background(), data)
		assert.NoError(t, err)
		assert.True(t, mock.called)
		assert.Len(t, mock.payload.Attachments, 1)
		assert.Equal(t, "#devops", mock.payload.Channel)
		assert.Equal(t, "[RECOVERED] api-check", mock.payload.Attachments[0].Title)
		assert.Equal(t, "api-check is recovered (last Ping: 10m0s)", mock.payload.Attachments[0].Text)
		assert.Nil(t, cfg.lastErr)
		assert.WithinDuration(t, time.Now(), cfg.lastSent, time.Second)
	})
}

func TestSlackConfig_Resolve(t *testing.T) {
	t.Setenv("SLACK_TOKEN", "xoxb-123456")
	t.Setenv("SLACK_CHANNEL", "#alerts")
	t.Setenv("SLACK_TITLE", "[{{ .Status }}] Alert")
	t.Setenv("SLACK_TEXT", "Last bump: {{ ago .LastBump }}")

	t.Run("resolves all fields", func(t *testing.T) {
		cfg := &SlackConfig{
			Token:     "env:SLACK_TOKEN",
			Channel:   "env:SLACK_CHANNEL",
			TitleTmpl: "env:SLACK_TITLE",
			TextTmpl:  "env:SLACK_TEXT",
		}

		err := cfg.Resolve()
		assert.NoError(t, err)
		assert.Equal(t, "xoxb-123456", cfg.Token)
		assert.Equal(t, "#alerts", cfg.Channel)
		assert.Equal(t, "[{{ .Status }}] Alert", cfg.TitleTmpl)
		assert.Equal(t, "Last bump: {{ ago .LastBump }}", cfg.TextTmpl)
	})

	t.Run("fails on token resolution", func(t *testing.T) {
		cfg := &SlackConfig{
			Token: "env:INVALID_VAR",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "resolve token: environment variable 'INVALID_VAR' not found")
	})

	t.Run("fails on channel resolution", func(t *testing.T) {
		cfg := &SlackConfig{
			Token:   "xoxb",
			Channel: "env:INVALID_VAR",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "resolve channel: environment variable 'INVALID_VAR' not found")
	})

	t.Run("fails on title template resolution", func(t *testing.T) {
		cfg := &SlackConfig{
			Token:     "xoxb",
			Channel:   "#ch",
			TitleTmpl: "env:INVALID_VAR",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "resolve title template: environment variable 'INVALID_VAR' not found")
	})

	t.Run("fails on text template resolution", func(t *testing.T) {
		cfg := &SlackConfig{
			Token:    "xoxb",
			Channel:  "#ch",
			TextTmpl: "env:INVALID_VAR",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "resolve text template: environment variable 'INVALID_VAR' not found")
	})
}

func TestSlackConfig_Validate(t *testing.T) {
	t.Run("valid with plain channel", func(t *testing.T) {
		cfg := &SlackConfig{
			Channel: "alerts",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "#alerts", cfg.Channel)
	})

	t.Run("valid with hash-prefixed channel", func(t *testing.T) {
		cfg := &SlackConfig{
			Channel: "#alerts",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "#alerts", cfg.Channel)
	})

	t.Run("missing channel", func(t *testing.T) {
		cfg := &SlackConfig{
			Channel: "",
		}
		err := cfg.Validate()
		assert.EqualError(t, err, "channel cannot be empty")
	})
}

func TestSlackConfig_Format(t *testing.T) {
	t.Parallel()

	cfg := &SlackConfig{}

	data := NotificationData{
		ID:          "health-check",
		Status:      "active",
		Description: "everything is fine",
		LastBump:    time.Now().Add(-3 * time.Minute),
	}

	out, err := cfg.Format(data)
	assert.NoError(t, err)
	assert.Equal(t, "[ACTIVE] health-check", out.Title)
	assert.Equal(t, "health-check is active (last Ping: 3m0s)", out.Message)
}
