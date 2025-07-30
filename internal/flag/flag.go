package flag

import (
	"fmt"
	"net"
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
func ParseFlags(args []string, version string) (Options, error) {
	tf := tinyflags.NewFlagSet("Heartbeats", tinyflags.ContinueOnError)
	tf.Version(version)
	tf.EnvPrefix("HEARTBEATS")

	configPath := tf.String("config", "config.yaml", "Path to configuration file").
		Short("c").
		Value()
	listenAddr := tf.TCPAddr("listen-address", &net.TCPAddr{IP: nil, Port: 8080}, "Listen address").
		Short("a").
		Placeholder("ADDR").
		Value()

	siteRoot := tf.String("site-root", "http://localhost:8080", "Site root URL").
		Short("r").
		Placeholder("URL").
		Value()
	histSize := tf.Int("history-size", 10000, "Size of the history. Minimum is 100").
		Validate(func(v int) error {
			if v < 100 {
				return fmt.Errorf("history size must be at least 100, got %d", v)
			}
			return nil
		}).
		Short("s").
		Placeholder("INT").
		Value()

	skipTLS := tf.Bool("skip-tls", false, "Skip TLS verification").Value()

	debug := tf.Bool("debug", false, "Enable debug logging").
		Short("d").
		Value()

	debugServer := tf.Int("debug-server-port", 8081, "Port for the debug server").
		Placeholder("PORT").
		Value()

	logFormat := tf.String("log-format", "json", "Log format").
		Choices(string(logging.LogFormatText), string(logging.LogFormatJSON)).
		Short("l").
		Value()

	retryCount := tf.Int("retry-count", 3, "Retries for failed notifications (-1 = infinite)").
		Validate(func(v int) error {
			if v < -1 || v == 0 {
				return fmt.Errorf("retry count must be -1 for infinite or >= 1, got %d", v)
			}
			return nil
		}).
		Placeholder("INT").
		Value()

	retryDelay := tf.Duration("retry-delay", 2*time.Second, "Delay between retries").
		Validate(func(v time.Duration) error {
			if v < 200*time.Millisecond {
				return fmt.Errorf("retry delay must be at least 200ms, got %s", v)
			}
			return nil
		}).
		Placeholder("DUR").
		Value()

	if err := tf.Parse(args); err != nil {
		return Options{}, err
	}

	return Options{
		ConfigPath:      *configPath,
		ListenAddr:      (*listenAddr).String(),
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
