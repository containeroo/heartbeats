package flag

import (
	"net"
	"strings"

	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/routes"

	"github.com/containeroo/tinyflags"
)

// Options holds the application configuration.
type Options struct {
	Debug            bool              // Enable debug logging.
	LogFormat        logging.LogFormat // Log output format.
	ListenAddr       string            // Address to listen on.
	RoutePrefix      string            // Route prefix for mounting.
	ConfigPath       string            // Path to YAML config file.
	StrictEnv        bool              // Enforce env placeholders in config.
	SiteRoot         string            // Site root URL.
	OverriddenValues map[string]any    // Overridden values from environment.
}

// ParseFlags parses flags and environment variables into Options.
func ParseFlags(args []string, version string) (Options, error) {
	opts := Options{}

	tf := tinyflags.NewFlagSet("heartbeats", tinyflags.ContinueOnError)
	tf.Version(version)
	tf.EnvPrefix("HEARTBEATS_")

	listenAddr := tf.TCPAddr("listen-address", &net.TCPAddr{IP: nil, Port: 8080}, "Listen address").
		Short("a").
		Placeholder("ADDR").
		Value()

	tf.StringVar(&opts.RoutePrefix, "route-prefix", "", "Path prefix to mount the app (e.g., /hiccup). Empty = root.").
		Finalize(func(input string) string {
			return routes.NormalizeRoutePrefix(input)
		}).
		Placeholder("PATH").
		Value()

	tf.StringVar(&opts.ConfigPath, "config", "", "Path to YAML config file.").
		Short("c").
		Placeholder("PATH").
		Required().
		Value()

	tf.StringVar(&opts.SiteRoot, "site-root", "http://localhost:8080", "Site root URL").
		Finalize(func(input string) string {
			return strings.TrimRight(input, "/")
		}).
		Short("r").
		Placeholder("URL").
		Value()

	tf.BoolVar(&opts.Debug, "debug", false, "Enable debug logging").
		Short("d").
		Value()

	tf.BoolVar(&opts.StrictEnv, "strict-env", false, "Fail if config references unset env vars").
		Value()

	logFormat := tf.String("log-format", "json", "Log format").
		Choices(string(logging.LogFormatText), string(logging.LogFormatJSON)).
		Short("l").
		Value()

	if err := tf.Parse(args); err != nil {
		return Options{}, err
	}

	opts.ListenAddr = (*listenAddr).String()
	opts.LogFormat = logging.LogFormat(*logFormat)
	opts.OverriddenValues = tf.OverriddenValues()

	return opts, nil
}
