package flag

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/logging"

	flag "github.com/spf13/pflag"
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

		cfg, err := ParseFlags([]string{}, os.Getenv, "vX.Y.Z")
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

		_, err := ParseFlags([]string{"--version"}, os.Getenv, "1.2.3")
		assert.Error(t, err)
		helpErr, ok := err.(*HelpRequested)
		assert.True(t, ok, "should return HelpRequested")
		expected := fmt.Sprintf("Heartbeats version %s\n", "1.2.3")
		assert.Equal(t, expected, helpErr.Message)
	})

	t.Run("show help", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--help"}, os.Getenv, "")
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
			"-l", "text",
		}
		cfg, err := ParseFlags(args, os.Getenv, "")
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
		_, err := ParseFlags(args, os.Getenv, "")

		assert.Error(t, err)
		assert.EqualError(t, err, "unknown flag: --invalid")
	})
}

func TestBindEnvFromUsage(t *testing.T) {
	t.Parallel()

	getEnv := func(key string) string {
		m := map[string]string{
			"TEST_CONFIG":         "env.yaml",
			"TEST_CONFIG_MISSING": "default.yaml",
		}

		return m[key]
	}

	t.Run("Overrides Unset", func(t *testing.T) {
		t.Parallel()

		var value string
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.StringVar(&value, "config", "default.yaml", "Config path [env: TEST_CONFIG]")

		assert.NoError(t, fs.Parse([]string{}))

		bindEnvFromUsage(fs, getEnv)

		assert.Equal(t, "env.yaml", value)
	},
	)

	t.Run("Does not override explicit flag", func(t *testing.T) {
		t.Parallel()

		var value string
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.StringVar(&value, "config", "default.yaml", "Config path [env: TEST_CONFIG]")

		assert.NoError(t, fs.Parse([]string{"--config=explicit.yaml"}))

		bindEnvFromUsage(fs, os.Getenv)

		assert.Equal(t, "explicit.yaml", value)
	})

	t.Run("Does not override missing env var", func(t *testing.T) {
		t.Parallel()

		var value string
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.StringVar(&value, "config", "default.yaml", "Config path [env: TEST_CONFIG_MISSING]")

		assert.NoError(t, fs.Parse([]string{}))

		bindEnvFromUsage(fs, os.Getenv)

		assert.Equal(t, "default.yaml", value)
	})

	t.Run("Corrupted env var", func(t *testing.T) {
		t.Parallel()

		var value string
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.StringVar(&value, "config", "default.yaml", "Config path [env: TEST_CONFIG")

		assert.NoError(t, fs.Parse([]string{}))

		bindEnvFromUsage(fs, getEnv)

		assert.Equal(t, "default.yaml", value)
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
		assert.ErrorContains(t, err, "retry count")
	})

	t.Run("invalid retry count < -1", func(t *testing.T) {
		t.Parallel()

		cfg := Options{
			LogFormat:  logging.LogFormatText,
			RetryCount: -2,
			RetryDelay: 2 * time.Second,
		}
		err := cfg.Validate()
		assert.ErrorContains(t, err, "retry count")
	})

	t.Run("invalid retry delay < 1s", func(t *testing.T) {
		t.Parallel()

		cfg := Options{
			LogFormat:  logging.LogFormatText,
			RetryCount: 3,
			RetryDelay: 500 * time.Millisecond,
		}
		err := cfg.Validate()
		assert.ErrorContains(t, err, "retry delay")
	})

	t.Run("invalid log format", func(t *testing.T) {
		t.Parallel()

		cfg := Options{
			LogFormat:  "xml", // invalid
			RetryCount: 3,
			RetryDelay: 2 * time.Second,
		}
		err := cfg.Validate()
		assert.ErrorContains(t, err, "invalid log format")
	})
}
