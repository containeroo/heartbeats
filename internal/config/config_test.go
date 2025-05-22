package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/containeroo/heartbeats/pkg/notify/email"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		_, err := LoadConfig("does_not_exist.yaml")
		assert.Error(t, err)
	})

	t.Run("invalid YAML", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "bad.yaml")
		assert.NoError(t, os.WriteFile(path, []byte("::: not: yaml"), 0644))

		_, err := LoadConfig(path)
		assert.Error(t, err)
	})

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "good.yaml")
		const sample = `
receivers:
  receiver1: {}
heartbeats:
  hb1:
    interval: 1s
    grace: 500ms
    description: "test heartbeat"
    receivers: ["receiver1"]
`
		assert.NoError(t, os.WriteFile(path, []byte(sample), 0644))

		cfg, err := LoadConfig(path)
		assert.NoError(t, err)

		// check receivers
		assert.Contains(t, cfg.Receivers, "receiver1")
		// check heartbeat settings
		hb, ok := cfg.Heartbeats["hb1"]
		assert.True(t, ok, "hb1 should be present")
		assert.Equal(t, time.Second, hb.Interval)
		assert.Equal(t, 500*time.Millisecond, hb.Grace)
		assert.Equal(t, "test heartbeat", hb.Description)
		assert.Equal(t, []string{"receiver1"}, hb.Receivers)
	})
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Receivers: map[string]notifier.ReceiverConfig{
				"r": {},
			},
			Heartbeats: map[string]heartbeat.HeartbeatConfig{
				"hb": {
					Interval:  10 * time.Second,
					Grace:     1 * time.Second,
					Receivers: []string{"r"},
				},
			},
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("unknown receiver", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Receivers: map[string]notifier.ReceiverConfig{
				"r": {},
			},
			Heartbeats: map[string]heartbeat.HeartbeatConfig{
				"hb": {
					Interval:  1 * time.Second,
					Grace:     500 * time.Millisecond,
					Receivers: []string{"no_such"},
				},
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `heartbeat "hb" references unknown receiver "no_such"`)
	})

	t.Run("invalid slack config - validate error", func(t *testing.T) {
		t.Parallel()

		rc := notifier.ReceiverConfig{
			SlackConfigs: []notifier.SlackConfig{
				{Channel: "", Token: ""},
			},
		}
		cfg := &Config{
			Receivers: map[string]notifier.ReceiverConfig{"r": rc},
			Heartbeats: map[string]heartbeat.HeartbeatConfig{
				"hb": {
					Interval:  1 * time.Second,
					Grace:     1 * time.Second,
					Receivers: []string{"r"},
				},
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "receiver \"r\" slack config error: channel cannot be empty")
	})

	t.Run("invalid slack config - resolve error", func(t *testing.T) {
		t.Parallel()

		rc := notifier.ReceiverConfig{
			SlackConfigs: []notifier.SlackConfig{
				{Channel: "env:INVALID"},
			},
		}
		cfg := &Config{
			Receivers: map[string]notifier.ReceiverConfig{"r": rc},
			Heartbeats: map[string]heartbeat.HeartbeatConfig{
				"hb": {
					Interval:  1 * time.Second,
					Grace:     1 * time.Second,
					Receivers: []string{"r"},
				},
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "receiver \"r\" slack config error: resolve channel: environment variable 'INVALID' not found")
	})

	t.Run("invalid email config - validate error", func(t *testing.T) {
		t.Parallel()

		rc := notifier.ReceiverConfig{
			EmailConfigs: []notifier.EmailConfig{
				{
					SMTPConfig: email.SMTPConfig{
						Host:     "",
						Port:     587,
						From:     "noreply@example.com",
						Username: "nobody",
						Password: "secret",
					},
					EmailDetails: notifier.EmailDetails{
						To:          []string{"foo@example.com"},
						SubjectTmpl: "hi",
						BodyTmpl:    "body",
					},
				},
			},
		}

		cfg := &Config{
			Receivers: map[string]notifier.ReceiverConfig{"r": rc},
			Heartbeats: map[string]heartbeat.HeartbeatConfig{
				"hb": {
					Interval:  1 * time.Second,
					Grace:     1 * time.Second,
					Receivers: []string{"r"},
				},
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "receiver \"r\" email config error: SMTP host and port must be specified")
	})

	t.Run("invalid email config - resolve error", func(t *testing.T) {
		t.Parallel()

		rc := notifier.ReceiverConfig{
			EmailConfigs: []notifier.EmailConfig{
				{
					SMTPConfig: email.SMTPConfig{
						Host:     "env:INVALID",
						Port:     587,
						From:     "noreply@example.com",
						Username: "nobody",
						Password: "secret",
					},
					EmailDetails: notifier.EmailDetails{
						To:          []string{"foo@example.com"},
						SubjectTmpl: "hi",
						BodyTmpl:    "body",
					},
				},
			},
		}

		cfg := &Config{
			Receivers: map[string]notifier.ReceiverConfig{"r": rc},
			Heartbeats: map[string]heartbeat.HeartbeatConfig{
				"hb": {
					Interval:  1 * time.Second,
					Grace:     1 * time.Second,
					Receivers: []string{"r"},
				},
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "receiver \"r\" email config error: failed to resolve SMTP host: environment variable 'INVALID' not found")
	})

	t.Run("invalid teams config - validate", func(t *testing.T) {
		t.Parallel()

		rc := notifier.ReceiverConfig{
			MSTeamsConfigs: []notifier.MSTeamsConfig{{WebhookURL: ""}},
		}

		cfg := &Config{
			Receivers: map[string]notifier.ReceiverConfig{"r": rc},
			Heartbeats: map[string]heartbeat.HeartbeatConfig{
				"hb": {
					Interval:  1 * time.Second,
					Grace:     1 * time.Second,
					Receivers: []string{"r"},
				},
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "receiver \"r\" MSTeams config error: webhook URL cannot be empty")
	})

	t.Run("invalid teams config - resolve error", func(t *testing.T) {
		t.Parallel()

		rc := notifier.ReceiverConfig{
			MSTeamsConfigs: []notifier.MSTeamsConfig{{WebhookURL: "env:INVALID"}},
		}

		cfg := &Config{
			Receivers: map[string]notifier.ReceiverConfig{"r": rc},
			Heartbeats: map[string]heartbeat.HeartbeatConfig{
				"hb": {
					Interval:  1 * time.Second,
					Grace:     1 * time.Second,
					Receivers: []string{"r"},
				},
			},
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "receiver \"r\" MSTeams config error: failed to resolve WebhookURL: environment variable 'INVALID' not found")
	})
}
