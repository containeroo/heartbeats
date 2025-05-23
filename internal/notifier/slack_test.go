package notifier

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/pkg/notify/slack"
	"github.com/stretchr/testify/require"
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

func TestSlackConfig_Notify(t *testing.T) {
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
		Status:      string(common.HeartbeatStateMissing),
		Description: "Heartbeat failed",
		LastBump:    time.Now().Add(-10 * time.Minute),
	}

	err := cfg.Notify(context.Background(), data)
	require.NoError(t, err)
	require.True(t, mock.called)
	require.Len(t, mock.payload.Attachments, 1)
	require.Equal(t, "#devops", mock.payload.Channel)
	require.Contains(t, mock.payload.Attachments[0].Title, "api-check")
	require.Contains(t, mock.payload.Attachments[0].Text, "api-check is missing")
	require.Nil(t, cfg.lastErr)
	require.WithinDuration(t, time.Now(), cfg.lastSent, time.Second)
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
	require.NoError(t, err)
	require.Contains(t, out.Title, "[ACTIVE]")
	require.Contains(t, out.Message, "health-check is active")
}
