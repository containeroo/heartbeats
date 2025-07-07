package flag

import (
	"os"
	"testing"

	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/tinyflags"
	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()

	t.Run("use defaults", func(t *testing.T) {
		t.Parallel()

		cfg, err := ParseFlags([]string{}, "vX.Y.Z", os.Getenv)
		assert.NoError(t, err)
		assert.Equal(t, "config.yaml", cfg.ConfigPath, "default config path")
		assert.Equal(t, ":8080", cfg.ListenAddr, "default listen address")
		assert.Equal(t, "http://localhost:8080", cfg.SiteRoot, "default site root")
		assert.Equal(t, 10000, cfg.HistorySize, "default history size")
		assert.False(t, cfg.Debug, "default debug flag")
		assert.Equal(t, 8081, cfg.DebugServerPort, "default debug server port")
		assert.False(t, cfg.SkipTLS, "default skipTLS setting")
		assert.Equal(t, logging.LogFormat("json"), cfg.LogFormat, "default log format")
	})

	t.Run("show version", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--version"}, "1.2.3", os.Getenv)
		assert.Error(t, err)
		assert.True(t, tinyflags.IsVersionRequested(err))
		assert.EqualError(t, err, "1.2.3")
	})

	t.Run("show help", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--help"}, "", os.Getenv)
		assert.Error(t, err)
		assert.True(t, tinyflags.IsHelpRequested(err))
		usage := `Usage: Heartbeats [flags]
Flags:
  -c, --config CONFIG                    Path to configuration file (Default: config.yaml)
  -a, --listen-address ADDR              Listen address (Default: :8080)
  -r, --site-root URL                    Site root URL (Default: http://localhost:8080)
  -s, --history-size INT                 Size of the history. Minimum is 100 (Default: 10000)
      --skip-tls                         Skip TLS verification
  -d, --debug                            Enable debug logging
      --debug-server-port PORT           Port for the debug server (Default: 8081)
  -l, --log-format <text|json>           Log format (Allowed: text, json) (Default: json)
      --retry-count INT                  Retries for failed notifications (-1 = infinite) (Default: 3)
      --retry-delay DUR                  Delay between retries (Default: 2s)
  -h, --help                             show help
`
		assert.EqualError(t, err, usage)
	})

	t.Run("custom values", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"-c", "myconf.yml",
			"-a", "127.0.0.1:9000",
			"-r", "https://example.com",
			"-s", "2000",
			"-d",
			"--debug-server-port", "8082",
			"-l", "text",
		}
		cfg, err := ParseFlags(args, "0.0.0", os.Getenv)
		assert.NoError(t, err)
		assert.Equal(t, "myconf.yml", cfg.ConfigPath)
		assert.Equal(t, "127.0.0.1:9000", cfg.ListenAddr)
		assert.Equal(t, "https://example.com", cfg.SiteRoot)
		assert.Equal(t, 2000, cfg.HistorySize)
		assert.True(t, cfg.Debug)
		assert.Equal(t, 8082, cfg.DebugServerPort)
		assert.Equal(t, logging.LogFormat("text"), cfg.LogFormat)
	})

	t.Run("parsing error", func(t *testing.T) {
		t.Parallel()
		args := []string{"--invalid"}
		_, err := ParseFlags(args, "", os.Getenv)

		assert.Error(t, err)
		assert.EqualError(t, err, "unknown flag: --invalid")
	})

	t.Run("valid min history size", func(t *testing.T) {
		t.Parallel()

		args := []string{"--history-size", "100"}
		_, err := ParseFlags(args, "", os.Getenv)

		assert.NoError(t, err)
	})

	t.Run("invalid min history size", func(t *testing.T) {
		t.Parallel()

		args := []string{"--history-size", "99"}
		_, err := ParseFlags(args, "", os.Getenv)

		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value for flag --history-size: invalid value \"99\": history size must be at least 100, got 99")
	})

	t.Run("valid log format", func(t *testing.T) {
		t.Parallel()
		args := []string{"--log-format", "json"}
		_, err := ParseFlags(args, "", os.Getenv)

		assert.NoError(t, err)
	})

	t.Run("invalid log format", func(t *testing.T) {
		t.Parallel()
		args := []string{"--log-format", "xml"}
		_, err := ParseFlags(args, "", os.Getenv)

		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value for flag --log-format: invalid value \"xml\": must be one of text, json")
	})

	t.Run("valid retry count", func(t *testing.T) {
		t.Parallel()
		args := []string{"--retry-count", "1"}
		_, err := ParseFlags(args, "", os.Getenv)

		assert.NoError(t, err)
	})

	t.Run("invalid retry count", func(t *testing.T) {
		t.Parallel()
		args := []string{"--retry-count", "0"}
		_, err := ParseFlags(args, "", os.Getenv)

		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value for flag --retry-count: invalid value \"0\": retry count must be -1 for infinite or >= 1, got 0")
	})

	t.Run("invalid retry delay", func(t *testing.T) {
		t.Parallel()
		args := []string{"--retry-delay", "50ms"}
		_, err := ParseFlags(args, "", os.Getenv)

		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value for flag --retry-delay: invalid value \"50ms\": retry delay must be at least 200ms, got 50ms")
	})

	t.Run("valid retry delay", func(t *testing.T) {
		t.Parallel()
		args := []string{"--retry-delay", "200ms"}
		_, err := ParseFlags(args, "", os.Getenv)

		assert.NoError(t, err)
	})
}
