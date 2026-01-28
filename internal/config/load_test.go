package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Receivers: map[string]ReceiverConfig{
				"ops": {
					Webhooks: []WebhookConfig{
						{URL: "https://example.com/webhook"},
					},
				},
			},
			Heartbeats: map[string]HeartbeatConfig{
				"api": {
					Interval:        3 * time.Second,
					LateAfter:       2 * time.Second,
					AlertOnRecovery: utils.ToPtr(true),
					Receivers:       []string{"ops"},
				},
			},
			History: HistoryConfig{
				Size:   100,
				Buffer: 1,
			},
		}

		require.NoError(t, cfg.Validate())
	})
	t.Run("error - missing receiver", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Receivers: map[string]ReceiverConfig{},
			Heartbeats: map[string]HeartbeatConfig{
				"api": {
					Interval:  3 * time.Second,
					LateAfter: 2 * time.Second,
					Receivers: []string{"ops"},
				},
			},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "receiver")
	})
	t.Run("error - heartbeat missing", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Receivers: map[string]ReceiverConfig{
				"ops": {
					Webhooks: []WebhookConfig{{URL: "https://example.com"}},
				},
			},
			Heartbeats: map[string]HeartbeatConfig{},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "heartbeat")
	})
}

func TestLoad(t *testing.T) {
	t.Parallel()
	t.Run("successful", func(t *testing.T) {
		t.Parallel()
		_, err := Load("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config path is required")
	})

	t.Run("error - non existent", func(t *testing.T) {
		t.Parallel()
		_, err := Load("non-existent.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read config")
	})

	t.Run("load valid file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		require.NoError(t, os.WriteFile(path, []byte(`
receivers:
  ops:
    emails:
      - host: smtp
        from: foo
        to: [bar]
        pass: baz
        user: usr
        ssl: true

heartbeats:
  api:
    interval: 3s
    late_after: 2s
    receivers: ["ops"]

history:
  size: 10
  buffer: 1
`), 0o600))

		cfg, err := Load(path)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, 1, len(cfg.Heartbeats))
		require.Equal(t, 1, len(cfg.Receivers))
	})
}
