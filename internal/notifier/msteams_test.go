package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/pkg/notify/msteams"
	"github.com/stretchr/testify/assert"
)

type mockMSTeamsSender struct {
	called  bool
	sentMsg msteams.MSTeams
	err     error
}

func (m *mockMSTeamsSender) Send(ctx context.Context, msg msteams.MSTeams, webhookURL string) (string, error) {
	m.called = true
	m.sentMsg = msg
	return "ok", m.err
}

func TestMSTeamsConfig_Type(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "msteams", NewMSTeamsNotifier("id", MSTeamsConfig{}, nil, nil).Type())
}

func TestMSTeamsConfig_Target(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "http://example.com/...path", NewMSTeamsNotifier("id", MSTeamsConfig{WebhookURL: "http://example.com/some-path"}, nil, nil).Target())
}

func TestMSTeamsConfig_LastSent(t *testing.T) {
	t.Parallel()
	assert.Equal(t, time.Time{}, NewMSTeamsNotifier("id", MSTeamsConfig{}, nil, nil).LastSent())
}

func TestMSTeamsConfig_LastErr(t *testing.T) {
	t.Parallel()
	assert.Nil(t, NewMSTeamsNotifier("id", MSTeamsConfig{}, nil, nil).LastErr())
}

func TestMSTeamsConfig_Notify(t *testing.T) {
	t.Parallel()
	t.Run("fails on invalid template", func(t *testing.T) {
		t.Parallel()

		mock := &mockMSTeamsSender{}
		config := &MSTeamsConfig{
			WebhookURL: "https://example.com",
			logger:     slog.New(slog.NewTextHandler(os.Stdout, nil)),
			sender:     mock,
			TextTmpl:   "{{ .Missing }}",
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
		assert.EqualError(t, err, "failed to format notification: format message: template: notification:1:3: executing \"notification\" at <.Missing>: can't evaluate field Missing in type notifier.NotificationData")
	})

	t.Run("Sender returns error", func(t *testing.T) {
		t.Parallel()

		mock := &mockMSTeamsSender{err: fmt.Errorf("mock error")}
		config := &MSTeamsConfig{
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
		assert.EqualError(t, err, "cannot send MSTeams notification. mock error")
	})

	t.Run("sends notification", func(t *testing.T) {
		t.Parallel()

		mock := &mockMSTeamsSender{}
		config := &MSTeamsConfig{
			WebhookURL: "https://example.com",
			logger:     slog.New(slog.NewTextHandler(os.Stdout, nil)),
			sender:     mock,
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
		assert.NoError(t, err)
		assert.True(t, mock.called)
		assert.Equal(t, "[FAILED] db-check", mock.sentMsg.Title)
		assert.Equal(t, "db-check is failed (last bump: 2m0s)", mock.sentMsg.Text)
		assert.Nil(t, config.lastErr)
	})
}

func TestMSTeamsConfig_Resolve(t *testing.T) {
	t.Setenv("TEAMS_URL", "https://hooks.example.com")
	t.Setenv("TEAMS_TITLE", "[{{ .Status }}] Alert")
	t.Setenv("TEAMS_TEXT", "Last seen at {{ ago .LastBump }}")

	t.Run("resolves all fields", func(t *testing.T) {
		t.Parallel()

		cfg := &MSTeamsConfig{
			WebhookURL: "env:TEAMS_URL",
			TitleTmpl:  "env:TEAMS_TITLE",
			TextTmpl:   "env:TEAMS_TEXT",
		}

		err := cfg.Resolve()
		assert.NoError(t, err)
		assert.Equal(t, "https://hooks.example.com", cfg.WebhookURL)
		assert.Equal(t, "[{{ .Status }}] Alert", cfg.TitleTmpl)
		assert.Equal(t, "Last seen at {{ ago .LastBump }}", cfg.TextTmpl)
	})

	t.Run("fails on WebhookURL", func(t *testing.T) {
		t.Parallel()

		cfg := &MSTeamsConfig{
			WebhookURL: "env:INVALID",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve WebhookURL: environment variable \"INVALID\" not found")
	})

	t.Run("fails on TitleTmpl", func(t *testing.T) {
		t.Parallel()

		cfg := &MSTeamsConfig{
			WebhookURL: "https://ok",
			TitleTmpl:  "env:INVALID",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve title template: environment variable \"INVALID\" not found")
	})

	t.Run("fails on TextTmpl", func(t *testing.T) {
		t.Parallel()

		cfg := &MSTeamsConfig{
			WebhookURL: "https://ok",
			TextTmpl:   "env:INVALID",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve text template: environment variable \"INVALID\" not found")
	})
}

func TestMSTeamsConfig_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		cfg := &MSTeamsConfig{
			WebhookURL: "https://hooks.example.com",
			TitleTmpl:  "[{{ .ID }}]",
			TextTmpl:   "Body",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("fails with invalid template", func(t *testing.T) {
		t.Parallel()

		cfg := &MSTeamsConfig{
			WebhookURL: "https://hooks.example.com",
			TitleTmpl:  "{{ .", // invalid
		}
		err := cfg.Validate()
		assert.EqualError(t, err, "format title: template: notification:1: illegal number syntax: \".\"")
	})

	t.Run("fails when URL is empty", func(t *testing.T) {
		t.Parallel()

		cfg := &MSTeamsConfig{
			WebhookURL: "",
			TitleTmpl:  "OK",
			TextTmpl:   "OK",
		}
		err := cfg.Validate()
		assert.EqualError(t, err, "webhook URL cannot be empty")
	})

	t.Run("Invalid URL", func(t *testing.T) {
		t.Parallel()

		cfg := &MSTeamsConfig{
			WebhookURL: "::::",
			TitleTmpl:  "OK",
			TextTmpl:   "OK",
		}
		err := cfg.Validate()
		assert.EqualError(t, err, "webhook URL is not a valid URL: parse \"::::\": missing protocol scheme")
	})
}

func TestMSTeamsConfig_Format(t *testing.T) {
	t.Parallel()

	config := &MSTeamsConfig{}

	now := time.Now()
	data := NotificationData{
		ID:       "api-heartbeat",
		Status:   "active",
		LastBump: now.Add(-5 * time.Minute),
	}

	out, err := config.Format(data)
	assert.NoError(t, err)
	assert.Equal(t, "[ACTIVE] api-heartbeat", out.Title)
	assert.Equal(t, "api-heartbeat is active (last bump: 5m0s)", out.Message)
}
