package notifier

import (
	"context"
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
	sendErr error
}

func (m *mockMSTeamsSender) Send(ctx context.Context, msg msteams.MSTeams, webhookURL string) (string, error) {
	m.called = true
	m.sentMsg = msg
	return "ok", m.sendErr
}

func TestMSTeamsConfig_Notify(t *testing.T) {
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
	assert.Contains(t, mock.sentMsg.Title, "db-check")
	assert.Contains(t, mock.sentMsg.Text, "db-check is failed")
	assert.Nil(t, config.lastErr)
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
	assert.Contains(t, out.Title, "[ACTIVE]")
	assert.Contains(t, out.Message, "api-heartbeat is active")
}
