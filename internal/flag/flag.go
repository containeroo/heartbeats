package flag

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/routeutil"

	"github.com/containeroo/tinyflags"
)

// Options holds the application configuration.
type Options struct {
	Debug            bool              // Set LogLevel to Debug
	DebugServerPort  int               // Port for the debug server
	LogFormat        logging.LogFormat // Specify the log output format
	ConfigPath       string            // Path to the configuration file
	ListenAddr       string            // Address to listen on
	SiteRoot         string            // Root URL of the site
	RoutePrefix      string            // Route prefix
	HistorySize      int               // Size of the history ring buffer
	SkipTLS          bool              // Skip TLS for all receivers
	RetryCount       int               // Number of retries for notifications
	RetryDelay       time.Duration     // Delay between retries
	OverriddenValues map[string]any    // Overridden values from environment
}

// ParseFlags parses flags and environment variables.
func ParseFlags(args []string, version string) (Options, error) {
	opts := Options{}

	tf := tinyflags.NewFlagSet("Heartbeats", tinyflags.ContinueOnError)
	tf.Version(version)
	tf.EnvPrefix("HEARTBEATS")

	tf.StringVar(&opts.ConfigPath, "config", "config.yaml", "Path to configuration file").
		Short("c").
		Value()
	listenAddr := tf.TCPAddr("listen-address", &net.TCPAddr{IP: nil, Port: 8080}, "Listen address").
		Short("a").
		Placeholder("ADDR").
		Value()

	tf.StringVar(&opts.RoutePrefix, "route-prefix", "", "Path prefix to mount the app (e.g., /heartbeats). Empty = root.").
		Finalize(func(input string) string {
			return routeutil.NormalizeRoutePrefix(input) // canonical "" or "/heartbeats"
		}).
		Placeholder("PATH").
		Value()

	tf.StringVar(&opts.SiteRoot, "site-root", "http://localhost:8080", "Site root URL").
		Finalize(func(input string) string {
			return strings.TrimRight(input, "/")
		}).
		Short("r").
		Placeholder("URL").
		Value()

	tf.IntVar(&opts.HistorySize, "history-size", 10000, "Size of the history. Minimum is 100").
		Validate(func(v int) error {
			if v < 100 {
				return fmt.Errorf("history size must be at least 100, got %d", v)
			}
			return nil
		}).
		Short("s").
		Placeholder("INT").
		Value()

	tf.BoolVar(&opts.SkipTLS, "skip-tls", false, "Skip TLS verification").
		Value()

	tf.BoolVar(&opts.Debug, "debug", false, "Enable debug logging").
		Short("d").
		Value()

	tf.IntVar(&opts.DebugServerPort, "debug-server-port", 8081, "Port for the debug server").
		Placeholder("PORT").
		Value()

	logFormat := tf.String("log-format", "json", "Log format").
		Choices(string(logging.LogFormatText), string(logging.LogFormatJSON)).
		Short("l").
		Value()

	tf.IntVar(&opts.RetryCount, "retry-count", 3, "Retries for failed notifications (-1 = infinite)").
		Validate(func(v int) error {
			if v < -1 || v == 0 {
				return fmt.Errorf("retry count must be -1 for infinite or >= 1, got %d", v)
			}
			return nil
		}).
		Placeholder("INT").
		Value()

	tf.DurationVar(&opts.RetryDelay, "retry-delay", 2*time.Second, "Delay between retries").
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

	opts.ListenAddr = (*listenAddr).String()
	opts.LogFormat = logging.LogFormat(*logFormat)
	opts.OverriddenValues = tf.OverriddenValues()
	if opts.RoutePrefix != "" && !strings.HasSuffix(opts.SiteRoot, opts.RoutePrefix) {
		opts.SiteRoot += opts.RoutePrefix
	}

	return opts, nil
}
