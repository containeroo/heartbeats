package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/pkg/notify/msteamsgraph"
	"github.com/stretchr/testify/assert"
)

// mockMSTeamsGraphSender implements msteamsgraph.Sender
type mockMSTeamsGraphSender struct {
	called  bool
	payload msteamsgraph.Message
	err     error
}

func (m *mockMSTeamsGraphSender) SendChannel(ctx context.Context, teamID, channelID string, msg msteamsgraph.Message) (*msteamsgraph.Response, error) {
	m.called = true
	m.payload = msg
	return &msteamsgraph.Response{ID: "12345"}, m.err
}

func TestMSTeamsGraphConfig_Type(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "msteamsgraph", NewMSTeamsGraphNotifier("id", MSTeamsGraphConfig{}, nil, nil).Type())
}

func TestMSTeamsGraphConfig_Target(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "********m-id/***********l-id", NewMSTeamsGraphNotifier("id", MSTeamsGraphConfig{
		TeamID:    "mock-team-id",
		ChannelID: "mock-channel-id",
	}, nil, nil).Target())
}

func TestMSTeamsGraphConfig_LastSent(t *testing.T) {
	t.Parallel()
	assert.Equal(t, time.Time{}, NewMSTeamsGraphNotifier("id", MSTeamsGraphConfig{}, nil, nil).LastSent())
}

func TestMSTeamsGraphConfig_LastErr(t *testing.T) {
	t.Parallel()
	assert.Nil(t, NewMSTeamsGraphNotifier("id", MSTeamsGraphConfig{}, nil, nil).LastErr())
}

func TestMSTeamsGraphConfig_Notify(t *testing.T) {
	t.Parallel()

	t.Run("invalid template", func(t *testing.T) {
		t.Parallel()
		mock := &mockMSTeamsGraphSender{}

		cfg := &MSTeamsGraphConfig{
			logger:    slog.New(slog.NewTextHandler(os.Stdout, nil)),
			sender:    mock,
			TitleTmpl: "{{ .Missing }}",
		}

		err := cfg.Notify(context.Background(), NotificationData{
			ID:       "test",
			Status:   "active",
			LastBump: time.Now(),
		})
		assert.ErrorContains(t, err, "format notification")
	})

	t.Run("send error", func(t *testing.T) {
		t.Parallel()
		mock := &mockMSTeamsGraphSender{err: fmt.Errorf("mock error")}

		cfg := &MSTeamsGraphConfig{
			TeamID:    "team",
			ChannelID: "chan",
			logger:    slog.New(slog.NewTextHandler(os.Stdout, nil)),
			sender:    mock,
		}

		err := cfg.Notify(context.Background(), NotificationData{
			ID:       "test",
			Status:   "failed",
			LastBump: time.Now(),
		})
		assert.ErrorContains(t, err, "send MS Teams message")
	})

	t.Run("successful send", func(t *testing.T) {
		t.Parallel()
		mock := &mockMSTeamsGraphSender{}

		cfg := &MSTeamsGraphConfig{
			TeamID:    "team",
			ChannelID: "chan",
			logger:    slog.New(slog.NewTextHandler(os.Stdout, nil)),
			sender:    mock,
		}

		err := cfg.Notify(context.Background(), NotificationData{
			ID:          "check-1",
			Status:      common.HeartbeatStateRecovered.String(),
			Description: "OK now",
			LastBump:    time.Now().Add(-2 * time.Minute),
		})
		assert.NoError(t, err)
		assert.True(t, mock.called)
		assert.Contains(t, mock.payload.Body.Content, "check-1")
		assert.Contains(t, mock.payload.Body.Content, "RECOVERED")
	})
}

func TestMSTeamsGraphConfig_Resolve(t *testing.T) {
	t.Setenv("TEAMS_TOKEN", "abc123")
	t.Setenv("TEAMS_TEAM", "t1")
	t.Setenv("TEAMS_CHANNEL", "c1")
	t.Setenv("TEAMS_TITLE", "{{ .ID }}")
	t.Setenv("TEAMS_TEXT", "{{ .Status }}")

	t.Run("resolves all fields", func(t *testing.T) {
		cfg := &MSTeamsGraphConfig{
			Token:     "env:TEAMS_TOKEN",
			TeamID:    "env:TEAMS_TEAM",
			ChannelID: "env:TEAMS_CHANNEL",
			TitleTmpl: "env:TEAMS_TITLE",
			TextTmpl:  "env:TEAMS_TEXT",
		}
		err := cfg.Resolve()
		assert.NoError(t, err)
		assert.Equal(t, "abc123", cfg.Token)
		assert.Equal(t, "t1", cfg.TeamID)
		assert.Equal(t, "c1", cfg.ChannelID)
		assert.Equal(t, "{{ .ID }}", cfg.TitleTmpl)
		assert.Equal(t, "{{ .Status }}", cfg.TextTmpl)
	})

	t.Run("fails on missing token", func(t *testing.T) {
		cfg := &MSTeamsGraphConfig{
			Token: "env:INVALID_VAR",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "resolve token: environment variable 'INVALID_VAR' not found")
	})
	t.Run("fails on missing team", func(t *testing.T) {
		cfg := &MSTeamsGraphConfig{
			Token:  "xoxb",
			TeamID: "env:INVALID_VAR",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "resolve teamID: environment variable 'INVALID_VAR' not found")
	})

	t.Run("fails on missing team", func(t *testing.T) {
		cfg := &MSTeamsGraphConfig{
			Token:     "xoxb",
			TeamID:    "id",
			ChannelID: "env:INVALID_VAR",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "resolve channelID: environment variable 'INVALID_VAR' not found")
	})

	t.Run("fails on title template resolution", func(t *testing.T) {
		cfg := &MSTeamsGraphConfig{
			Token:     "xoxb",
			TeamID:    "id",
			ChannelID: "id",
			TitleTmpl: "env:INVALID_VAR",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "resolve title template: environment variable 'INVALID_VAR' not found")
	})

	t.Run("fails on text template resolution", func(t *testing.T) {
		cfg := &MSTeamsGraphConfig{
			Token:     "xoxb",
			TeamID:    "id",
			ChannelID: "id",
			TextTmpl:  "env:INVALID_VAR",
		}
		err := cfg.Resolve()
		assert.EqualError(t, err, "resolve text template: environment variable 'INVALID_VAR' not found")
	})
}

func TestMSTeamsGraphConfig_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		cfg := &MSTeamsGraphConfig{
			Token:     "tok",
			TeamID:    "t1",
			ChannelID: "c1",
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("missing token", func(t *testing.T) {
		cfg := &MSTeamsGraphConfig{TeamID: "t1", ChannelID: "c1"}
		assert.EqualError(t, cfg.Validate(), "token cannot be empty")
	})

	t.Run("missing team", func(t *testing.T) {
		cfg := &MSTeamsGraphConfig{Token: "tok", ChannelID: "c1"}
		assert.EqualError(t, cfg.Validate(), "teamID cannot be empty")
	})

	t.Run("missing channel", func(t *testing.T) {
		cfg := &MSTeamsGraphConfig{Token: "tok", TeamID: "t1"}
		assert.EqualError(t, cfg.Validate(), "channelID cannot be empty")
	})
}
