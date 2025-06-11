package notifier

import (
	"context"
	"fmt"
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

func TestEmailConfig_Type(t *testing.T) {
	t.Parallel()

	t.Run("returns email", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "email", NewEmailNotifier("id", EmailConfig{}, nil, nil).Type())
	})
}

func TestEmailConfig_LastSent(t *testing.T) {
	t.Parallel()

	t.Run("returns last sent time", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, time.Time{}, NewEmailNotifier("id", EmailConfig{}, nil, nil).LastSent())
	})
}

func TestEmailConfig_LastErr(t *testing.T) {
	t.Parallel()

	t.Run("returns last error", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, NewEmailNotifier("id", EmailConfig{}, nil, nil).LastErr())
	})
}

func TestEmailConfig_Notify(t *testing.T) {
	t.Parallel()

	t.Run("fails on invalid template", func(t *testing.T) {
		t.Parallel()

		mock := &mockEmailSender{}

		config := &EmailConfig{
			logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
			sender: mock,
			EmailDetails: EmailDetails{
				SubjectTmpl: "{{ .Missing }}",
			},
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
		assert.EqualError(t, err, "failed to format notification: format title: template: notification:1:3: executing \"notification\" at <.Missing>: can't evaluate field Missing in type notifier.NotificationData")
	})

	t.Run("Sender returns error", func(t *testing.T) {
		t.Parallel()

		mock := &mockEmailSender{err: fmt.Errorf("mock error")}

		config := &EmailConfig{
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
		assert.EqualError(t, err, "cannot send Email notification: mock error")
	})

	t.Run("sends notification", func(t *testing.T) {
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
	})
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

func TestEmailConfig_Resolve(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_FROM", "noreply@example.com")
	t.Setenv("SMTP_USER", "smtp-user")
	t.Setenv("SMTP_PASS", "smtp-pass")
	t.Setenv("EMAIL_TO", "user@example.com")
	t.Setenv("EMAIL_SUBJECT", "Alert: {{ .ID }")
	t.Setenv("EMAIL_BODY", "Body for {{ .ID }")

	t.Run("resolves all fields", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host:     "env:SMTP_HOST",
				From:     "env:SMTP_FROM",
				Username: "env:SMTP_USER",
				Password: "env:SMTP_PASS",
			},
			EmailDetails: EmailDetails{
				To:          []string{"env:EMAIL_TO"},
				SubjectTmpl: "env:EMAIL_SUBJECT",
				BodyTmpl:    "env:EMAIL_BODY",
			},
		}

		err := cfg.Resolve()
		assert.NoError(t, err)

		assert.Equal(t, "smtp.example.com", cfg.SMTPConfig.Host)
		assert.Equal(t, "noreply@example.com", cfg.SMTPConfig.From)
		assert.Equal(t, "smtp-user", cfg.SMTPConfig.Username)
		assert.Equal(t, "smtp-pass", cfg.SMTPConfig.Password)
		assert.Equal(t, []string{"user@example.com"}, cfg.EmailDetails.To)
		assert.Equal(t, "Alert: {{ .ID }", cfg.EmailDetails.SubjectTmpl)
		assert.Equal(t, "Body for {{ .ID }", cfg.EmailDetails.BodyTmpl)
	})

	t.Run("fails on not resolvable 'host'", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "env:INVALID",
			},
		}

		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve SMTP host: environment variable 'INVALID' not found")
	})

	t.Run("fails on not resolvable 'From'", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "smtp.example.com",
				From: "env:INVALID",
			},
		}

		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve SMTP from: environment variable 'INVALID' not found")
	})

	t.Run("fails on not resolvable 'Username'", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host:     "smtp.example.com",
				From:     "from@example.com",
				Username: "env:INVALID",
			},
		}

		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve SMTP username: environment variable 'INVALID' not found")
	})

	t.Run("fails on not resolvable 'Password'", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host:     "smtp.example.com",
				From:     "from@example.com",
				Password: "env:INVALID",
			},
		}

		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve SMTP password: environment variable 'INVALID' not found")
	})

	t.Run("fails on not resolvable 'To'", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "smtp.example.com",
				Port: 587,
				From: "from@example.com",
			},
			EmailDetails: EmailDetails{
				To: []string{"env:INVALID"},
			},
		}

		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve email recipient: environment variable 'INVALID' not found")
	})

	t.Run("fails on not resolvable 'Cc'", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "smtp.example.com",
				Port: 587,
				From: "from@example.com",
			},
			EmailDetails: EmailDetails{
				CC: []string{"env:INVALID"},
			},
		}

		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve email cc: environment variable 'INVALID' not found")
	})

	t.Run("fails on not resolvable 'Bcc'", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "smtp.example.com",
				Port: 587,
				From: "from@example.com",
			},
			EmailDetails: EmailDetails{
				BCC: []string{"env:INVALID"},
			},
		}

		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve email bcc: environment variable 'INVALID' not found")
	})

	t.Run("fails on not resolvable 'SubjectTmpl'", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "smtp.example.com",
				Port: 587,
				From: "from@example.com",
			},
			EmailDetails: EmailDetails{
				SubjectTmpl: "env:INVALID",
			},
		}

		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve subject template: environment variable 'INVALID' not found")
	})

	t.Run("fails on not resolvable 'BodyTmpl'", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "smtp.example.com",
				Port: 587,
				From: "from@example.com",
			},
			EmailDetails: EmailDetails{
				BodyTmpl: "env:INVALID",
			},
		}

		err := cfg.Resolve()
		assert.EqualError(t, err, "failed to resolve body template: environment variable 'INVALID' not found")
	})
}

func TestEmailConfig_Validate(t *testing.T) {
	t.Parallel()

	t.Run("invalid format", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "smtp.example.com",
				Port: 587,
				From: "from@example.com",
			},
			EmailDetails: EmailDetails{
				To:          []string{"user@example.com"},
				SubjectTmpl: "{{.",
			},
		}

		err := cfg.Validate()
		assert.EqualError(t, err, "format title: template: notification:1: illegal number syntax: \".\"")
	})

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "smtp.example.com",
				From: "from@example.com",
				Port: 587,
			},
			EmailDetails: EmailDetails{
				To:          []string{"user@example.com"},
				SubjectTmpl: "Hello",
				BodyTmpl:    "Body",
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("fails on 'Host' with no value", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "",
			},
		}

		err := cfg.Validate()
		assert.EqualError(t, err, "SMTP host and port must be specified")
	})

	t.Run("missing port", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "host.example.com",
			},
			EmailDetails: EmailDetails{
				To:          []string{"user@example.com"},
				SubjectTmpl: "Hello",
				BodyTmpl:    "Body",
			},
		}

		err := cfg.Validate()
		assert.EqualError(t, err, "SMTP host and port must be specified")
	})

	t.Run("missing recipient", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "smtp.example.com",
				Port: 25,
				From: "from@example.com",
			},
			EmailDetails: EmailDetails{
				To:          []string{},
				SubjectTmpl: "Subject",
				BodyTmpl:    "Body",
			},
		}

		err := cfg.Validate()
		assert.EqualError(t, err, "at least one recipient must be specified")
	})

	t.Run("fails on 'From' with no value", func(t *testing.T) {
		t.Parallel()

		cfg := &EmailConfig{
			SMTPConfig: email.SMTPConfig{
				Host: "smtp.example.com",
				Port: 587,
				From: "",
			},
		}

		err := cfg.Validate()
		assert.EqualError(t, err, "SMTP from must be specified")
	})
}
