package flag

import (
	"os"
	"strings"
	"testing"

	"github.com/containeroo/tinyflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/containeroo/heartbeats/internal/logging"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "HEARTBEATS_") {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		value, ok := os.LookupEnv(key)
		if ok {
			t.Cleanup(func() {
				_ = os.Setenv(key, value)
			})
		} else {
			t.Cleanup(func() {
				_ = os.Unsetenv(key)
			})
		}
		_ = os.Unsetenv(key)
	}
}

func TestParseFlags(t *testing.T) {
	// clearEnv(t) // since using t.Setenv Parallel will panic

	t.Run("use defaults", func(t *testing.T) {
		clearEnv(t)

		cfg, err := ParseFlags([]string{"--config", "config.yaml"}, "vX.Y.Z")
		assert.NoError(t, err)
		assert.Equal(t, ":8080", cfg.ListenAddr, "default listen address")
		assert.False(t, cfg.Debug, "default debug flag")
		assert.Equal(t, logging.LogFormat("json"), cfg.LogFormat, "default log format")
		assert.Equal(t, "", cfg.RoutePrefix, "default route prefix")
		assert.Equal(t, "config.yaml", cfg.ConfigPath, "config path")
	})

	t.Run("show version", func(t *testing.T) {
		clearEnv(t)

		_, err := ParseFlags([]string{"--version"}, "1.2.3")
		assert.Error(t, err)
		assert.True(t, tinyflags.IsVersionRequested(err))
		assert.EqualError(t, err, "1.2.3")
	})

	t.Run("show help", func(t *testing.T) {
		clearEnv(t)

		_, err := ParseFlags([]string{"--help"}, "")
		assert.Error(t, err)
		assert.True(t, tinyflags.IsHelpRequested(err))
		require.Error(t, err)
		usage := err.Error()
		assert.True(t, strings.HasPrefix(usage, "Usage: heartbeats [flags]"))
		assert.Contains(t, usage, "--config")
		assert.Contains(t, usage, "--route-prefix")
		assert.Contains(t, usage, "--listen-address")
		assert.Contains(t, usage, "--log-format")
	})

	t.Run("custom values", func(t *testing.T) {
		clearEnv(t)

		args := []string{
			"-a", "127.0.0.1:9000",
			"--route-prefix", "heartbeats",
			"-d",
			"-l", "text",
			"--config", "config.yaml",
		}
		cfg, err := ParseFlags(args, "0.0.0")
		assert.NoError(t, err)
		assert.Equal(t, "127.0.0.1:9000", cfg.ListenAddr)
		assert.Equal(t, "/heartbeats", cfg.RoutePrefix)
		assert.True(t, cfg.Debug)
		assert.Equal(t, logging.LogFormat("text"), cfg.LogFormat)
	})

	t.Run("parsing error", func(t *testing.T) {
		clearEnv(t)
		args := []string{"--config", "config.yaml", "--invalid"}
		_, err := ParseFlags(args, "")

		assert.Error(t, err)
		assert.EqualError(t, err, "unknown flag: --invalid")
	})

	t.Run("missing config", func(t *testing.T) {
		clearEnv(t)

		args := []string{}
		_, err := ParseFlags(args, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config")
	})

	t.Run("valid log format", func(t *testing.T) {
		clearEnv(t)
		args := []string{"--config", "config.yaml", "--log-format", "json"}
		_, err := ParseFlags(args, "")

		assert.NoError(t, err)
	})

	t.Run("invalid log format", func(t *testing.T) {
		clearEnv(t)
		args := []string{"--config", "config.yaml", "--log-format", "xml"}
		_, err := ParseFlags(args, "")

		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value for flag --log-format: must be one of: text, json.")
	})

	t.Run("test route prefix", func(t *testing.T) {
		clearEnv(t)
		args := []string{"--config", "config.yaml", "--route-prefix", "heartbeats"}
		flags, err := ParseFlags(args, "")

		require.NoError(t, err)
		assert.Equal(t, "/heartbeats", flags.RoutePrefix)
	})
}
