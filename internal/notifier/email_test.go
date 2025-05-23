package notifier

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/pkg/notify/email"
	"github.com/stretchr/testify/assert"
)

type mockEmailSender struct {
	called  bool
	message email.Message
	err     error
}

func (m *mockEmailSender) Send(ctx context.Context, msg email.Message) error {
	m.called = true
	m.message = msg
	return m.err
}

func TestEmailConfig_Notify(t *testing.T) {
	t.Parallel()

	mock := &mockEmailSender{}

	cfg := &EmailConfig{
		SMTPConfig: email.SMTPConfig{
			Host: "smtp.example.com",
			Port: 587,
			From: "noreply@example.com",
		},
		EmailDetails: EmailDetails{
			To: []string{"dev@example.com"},
		},
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
		sender: mock,
	}

	data := NotificationData{
		ID:       "my-check",
		Status:   "failed",
		LastBump: time.Now().Add(-10 * time.Minute),
	}

	err := cfg.Notify(context.Background(), data)
	assert.NoError(t, err)
	assert.True(t, mock.called)
	assert.Equal(t, "dev@example.com", mock.message.To[0])
	assert.Contains(t, mock.message.Subject, "my-check")
	assert.Contains(t, mock.message.Body, "Last ping")
	assert.Nil(t, cfg.lastErr)
	assert.WithinDuration(t, time.Now(), cfg.lastSent, time.Second)
}

func TestEmailConfig_Format(t *testing.T) {
	t.Parallel()

	cfg := &EmailConfig{
		EmailDetails: EmailDetails{},
	}

	data := NotificationData{
		ID:          "heartbeat-01",
		Description: "Checks DB",
		Status:      "active",
		LastBump:    time.Now().Add(-3 * time.Minute),
	}

	out, err := cfg.Format(data)
	assert.NoError(t, err)
	assert.Contains(t, out.Title, "heartbeat-01 ACTIVE")
	assert.Contains(t, out.Message, "Checks DB")
}
