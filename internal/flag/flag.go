package flag

import (
	"fmt"
	"time"

	"github.com/containeroo/heartbeats/internal/logging"

	"github.com/containeroo/tinyflags"
)

// Options holds the application configuration.
type Options struct {
	Debug           bool              // Set LogLevel to Debug
	DebugServerPort int               // Port for the debug server
	LogFormat       logging.LogFormat // Specify the log output format
	ConfigPath      string            // Path to the configuration file
	ListenAddr      string            // Address to listen on
	SiteRoot        string            // Root URL of the site
	HistorySize     int               // Size of the history ring buffer
	SkipTLS         bool              // Skip TLS for all receivers
	RetryCount      int               // Number of retries for notifications
	RetryDelay      time.Duration     // Delay between retries
}

// ParseFlags parses flags and environment variables.
func ParseFlags(args []string, version string, getEnv func(string) string) (Options, error) {
	tf := tinyflags.NewFlagSet("Heartbeats", tinyflags.ContinueOnError)
	tf.Version(version)
	tf.Sorted(false)

	configPath := tf.StringP("config", "c", "config.yaml", "Path to configuration file").Value()
	listenAddr := tf.ListenAddrP("listen-address", "a", ":8080", "Listen address").
		Metavar("ADDR").
		Value()

	siteRoot := tf.StringP("site-root", "r", "http://localhost:8080", "Site root URL").
		Metavar("URL").
		Value()
	histSize := tf.IntP("history-size", "s", 10000, "Size of the history. Minimum is 100").
		Validator(func(v int) error {
			if v < 100 {
				return fmt.Errorf("history size must be at least 100, got %d", v)
			}
			return nil
		}).
		Metavar("INT").
		Value()

	skipTLS := tf.Bool("skip-tls", false, "Skip TLS verification").Value()

	debug := tf.BoolP("debug", "d", false, "Enable debug logging").Value()

	debugServer := tf.Int("debug-server-port", 8081, "Port for the debug server").
		Metavar("PORT").
		Value()

	logFormat := tf.StringP("log-format", "l", "json", "Log format").
		Choices(string(logging.LogFormatText), string(logging.LogFormatJSON)).
		Value()

	retryCount := tf.Int("retry-count", 3, "Retries for failed notifications (-1 = infinite)").
		Validator(func(v int) error {
			if v < -1 || v == 0 {
				return fmt.Errorf("retry count must be -1 for infinite or >= 1, got %d", v)
			}
			return nil
		}).
		Metavar("INT").
		Value()

	retryDelay := tf.Duration("retry-delay", 2*time.Second, "Delay between retries").
		Validator(func(v time.Duration) error {
			if v < 200*time.Millisecond {
				return fmt.Errorf("retry delay must be at least 200ms, got %s", v)
			}
			return nil
		}).
		Metavar("DUR").
		Value()

	if err := tf.Parse(args); err != nil {
		return Options{}, err
	}

	return Options{
		ConfigPath:      *configPath,
		ListenAddr:      *listenAddr,
		SiteRoot:        *siteRoot,
		HistorySize:     *histSize,
		Debug:           *debug,
		DebugServerPort: *debugServer,
		SkipTLS:         *skipTLS,
		RetryCount:      *retryCount,
		RetryDelay:      *retryDelay,
		LogFormat:       logging.LogFormat(*logFormat),
	}, nil
}
