package flag

import (
	"fmt"
	"strings"
	"testing"

	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()

	t.Run("Error Message", func(t *testing.T) {
		t.Parallel()

		err := &HelpRequested{Message: "This is a help message"}
		assert.Equal(t, "This is a help message", err.Error(), "Error() should return the correct message")
	})

	t.Run("use defaults", func(t *testing.T) {
		cfg, err := ParseFlags([]string{}, "vX.Y.Z")
		assert.NoError(t, err)
		assert.Equal(t, "config.yaml", cfg.ConfigPath, "default config path")
		assert.Equal(t, ":8080", cfg.ListenAddr, "default listen address")
		assert.Equal(t, "http://localhost:8080", cfg.SiteRoot, "default site root")
		assert.Equal(t, 10000, cfg.HistorySize, "default history size")
		assert.False(t, cfg.Debug, "default debug flag")
		assert.False(t, cfg.SkipTLS, "default skipTLS setting")
		assert.Equal(t, logging.LogFormat("json"), cfg.LogFormat, "default log format")
	})

	t.Run("show version", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--version"}, "1.2.3")
		assert.Error(t, err)
		helpErr, ok := err.(*HelpRequested)
		assert.True(t, ok, "should return HelpRequested")
		expected := fmt.Sprintf("Heartbeats version %s\n", "1.2.3")
		assert.Equal(t, expected, helpErr.Message)
	})

	t.Run("show help", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--help"}, "")
		assert.Error(t, err)
		helpErr, ok := err.(*HelpRequested)
		assert.True(t, ok, "should return HelpRequested")
		// Ensure usage header is present
		assert.True(t, strings.HasPrefix(helpErr.Message, "Usage: heartbeats [flags]"))
	})

	t.Run("invalid log format", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--log-format", "xml"}, "")
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid log format: 'xml'")
	})

	t.Run("custom values", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"-c", "myconf.yml",
			"-a", "127.0.0.1:9000",
			"-r", "https://example.com",
			"-s", "42",
			"-d",
			"-l", "text",
		}
		cfg, err := ParseFlags(args, "")
		assert.NoError(t, err)
		assert.Equal(t, "myconf.yml", cfg.ConfigPath)
		assert.Equal(t, "127.0.0.1:9000", cfg.ListenAddr)
		assert.Equal(t, "https://example.com", cfg.SiteRoot)
		assert.Equal(t, 42, cfg.HistorySize)
		assert.True(t, cfg.Debug)
		assert.Equal(t, logging.LogFormat("text"), cfg.LogFormat)
	})

	t.Run("parsing error", func(t *testing.T) {
		t.Parallel()
		args := []string{"--invalid"}
		_, err := ParseFlags(args, "")

		assert.Error(t, err)
		assert.EqualError(t, err, "unknown flag: --invalid")
	})
}
