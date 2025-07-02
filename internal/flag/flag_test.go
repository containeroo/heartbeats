package flag

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/spf13/pflag"
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
		helpErr, ok := err.(*HelpRequested)
		assert.True(t, ok, "should return HelpRequested")
		expected := fmt.Sprintf("Heartbeats version %s\n", "1.2.3")
		assert.Equal(t, expected, helpErr.Message)
	})

	t.Run("show help", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--help"}, "", os.Getenv)
		assert.Error(t, err)
		helpErr, ok := err.(*HelpRequested)
		assert.True(t, ok, "should return HelpRequested")
		// Ensure usage header is present
		assert.True(t, strings.HasPrefix(helpErr.Message, "Usage: heartbeats [flags]"))
	})

	t.Run("custom values", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"-c", "myconf.yml",
			"-a", "127.0.0.1:9000",
			"-r", "https://example.com",
			"-s", "42",
			"-d",
			"--debug-server-port", "8082",
			"-l", "text",
		}
		cfg, err := ParseFlags(args, "0.0.0", os.Getenv)
		assert.NoError(t, err)
		assert.Equal(t, "myconf.yml", cfg.ConfigPath)
		assert.Equal(t, "127.0.0.1:9000", cfg.ListenAddr)
		assert.Equal(t, "https://example.com", cfg.SiteRoot)
		assert.Equal(t, 42, cfg.HistorySize)
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
}

func TestBuildOptions(t *testing.T) {
	t.Parallel()

	t.Run("valid flags", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

		// Register all flags required by buildOptions
		fs.String("config", "test.yaml", "Path to config")
		fs.String("listen-address", ":9090", "Listen address")
		fs.String("site-root", "https://example.com", "Site root")
		fs.Int("history-size", 123, "History size")
		fs.Bool("debug", true, "Enable debug")
		fs.Int("debug-server-port", 9999, "Debug server port")
		fs.Bool("skip-tls", true, "Skip TLS")
		fs.Int("retry-count", 5, "Retry count")
		fs.Duration("retry-delay", 3*time.Second, "Retry delay")
		fs.String("log-format", "text", "Log format")

		// Simulate parsing (uses the default values above)
		assert.NoError(t, fs.Parse([]string{}))

		opts, err := buildOptions(fs)
		assert.NoError(t, err)

		assert.Equal(t, "test.yaml", opts.ConfigPath)
		assert.Equal(t, ":9090", opts.ListenAddr)
		assert.Equal(t, "https://example.com", opts.SiteRoot)
		assert.Equal(t, 123, opts.HistorySize)
		assert.Equal(t, true, opts.Debug)
		assert.Equal(t, 9999, opts.DebugServerPort)
		assert.Equal(t, true, opts.SkipTLS)
		assert.Equal(t, 5, opts.RetryCount)
		assert.Equal(t, 3*time.Second, opts.RetryDelay)
		assert.Equal(t, logging.LogFormat("text"), opts.LogFormat)
	})

	t.Run("panics on error", func(t *testing.T) {
		t.Parallel()

		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

		opts, err := buildOptions(fs)

		assert.Error(t, err)
		assert.EqualError(t, err, "failed to parse flags: flag not found: config")
		assert.Equal(t, Options{}, opts)
	})

	t.Run("partial flags", func(t *testing.T) {
		t.Parallel()

		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

		// Register only a subset of expected flags
		fs.String("config", "cfg.yaml", "config file")
		fs.String("site-root", "http://localhost", "site root URL")
		fs.Int("history-size", 100, "history size")

		// leave out "listen-address" → must(envflag.HostPort(...)) will panic

		_, err := buildOptions(fs)

		assert.Error(t, err)
		assert.EqualError(t, err, "failed to parse flags: flag not found: listen-address")
	})
}

func TestMust(t *testing.T) {
	t.Parallel()

	t.Run("returns value if no error", func(t *testing.T) {
		t.Parallel()

		got := must("value", nil)
		assert.Equal(t, "value", got)
	})

	t.Run("panics if error is not nil", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic but got none")
			}
		}()
		_ = must("fail", errors.New("error"))
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		cfg := Options{
			LogFormat:  logging.LogFormatJSON,
			RetryCount: 3,
			RetryDelay: 2 * time.Second,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("invalid retry count = 0", func(t *testing.T) {
		t.Parallel()

		cfg := Options{
			LogFormat:  logging.LogFormatText,
			RetryCount: 0,
			RetryDelay: 2 * time.Second,
		}
		err := cfg.Validate()
		assert.EqualError(t, err, "retry count must be -1 (infinite) or >= 1, got 0")
	})

	t.Run("invalid retry count < -1", func(t *testing.T) {
		t.Parallel()

		cfg := Options{
			LogFormat:  logging.LogFormatText,
			RetryCount: -2,
			RetryDelay: 2 * time.Second,
		}
		err := cfg.Validate()
		assert.EqualError(t, err, "retry count must be -1 (infinite) or >= 1, got -2")
	})

	t.Run("invalid retry delay < 1s", func(t *testing.T) {
		t.Parallel()

		cfg := Options{
			LogFormat:  logging.LogFormatText,
			RetryCount: 3,
			RetryDelay: 500 * time.Millisecond,
		}
		err := cfg.Validate()
		assert.EqualError(t, err, "retry delay must be at least 1s, got 500ms")
	})

	t.Run("invalid log format", func(t *testing.T) {
		t.Parallel()

		cfg := Options{
			LogFormat:  "xml", // invalid
			RetryCount: 3,
			RetryDelay: 2 * time.Second,
		}
		err := cfg.Validate()
		assert.EqualError(t, err, "invalid log format: 'xml'")
	})
}

func TestIsHelpRequested(t *testing.T) {
	t.Parallel()

	t.Run("returns true and writes message for HelpRequested error", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		helpMsg := "this is the help message\n"
		err := &HelpRequested{Message: helpMsg}

		ok := IsHelpRequested(err, buf)

		assert.True(t, ok)
		assert.Equal(t, helpMsg, buf.String())
	})

	t.Run("returns false and writes nothing for unrelated error", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		err := errors.New("some other error")

		ok := IsHelpRequested(err, buf)

		assert.False(t, ok)
		assert.Equal(t, "", buf.String())
	})
}
