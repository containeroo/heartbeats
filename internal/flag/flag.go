package flag

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/containeroo/heartbeats/internal/logging"

	flag "github.com/spf13/pflag"
)

// HelpRequested indicates that help was requested.
type HelpRequested struct {
	Message string // Message is the help message.
}

// Error returns the help message.
func (e *HelpRequested) Error() string {
	return e.Message
}

// Options holds the application configuration.
type Options struct {
	Debug       bool              // Set LogLevel to Debug
	LogFormat   logging.LogFormat // Specify the log output format
	ConfigPath  string            // Path to the configuration file
	ListenAddr  string            // Address to listen on
	SiteRoot    string            // Root URL of the site
	HistorySize int               // Size of the history ring buffer
	SkipTLS     bool              // Skip TLS for all receivers
	RetryCount  int               // Number of retries for notifications
	RetryDelay  time.Duration     // Delay between retries
}

// ParseFlags parses command-line flags.
func ParseFlags(args []string, getEnv func(string) string, version string) (Options, error) {
	fs := flag.NewFlagSet("Heartbeats", flag.ContinueOnError)
	fs.SortFlags = false

	// Server settings
	configPath := fs.StringP("config", "c", "config.yaml", "Path to configuration file [env: HEARTBEATS_CONFIG]")
	listenAddress := fs.StringP("listen-address", "a", ":8080", "Address to listen on [env: HEARTBEATS_LISTEN_ADDRESS]")
	siteRoot := fs.StringP("site-root", "r", "http://localhost:8080", "Site root URL [env: HEARTBEATS_SITE_ROOT]")
	historySize := fs.IntP("history-size", "s", 10000, "Size of the history [env: HEARTBEATS_HISTORY_SIZE]")
	skipTLS := fs.Bool("skip-tls", false, "Skip TLS verification for all receivers (can be overridden per receiver) [env: HEARTBEATS_SKIP_TLS]")

	// Application logging.
	debug := fs.BoolP("debug", "d", false, "Enable debug logging (default: false) [env: HEARTBEAT_DEBUG]")
	logFormat := fs.StringP("log-format", "l", "json", "Log format (json | text) [env: HEARTBEATS_LOG_FORMAT]")

	// Retry settings
	retryCount := fs.Int("retry-count", 3, "How many times to retry a failed notification. Use -1 for infinite retries. [env: HEARTBEATS_RETRY_COUNT]")
	retryDelay := fs.Duration("retry-delay", 2*time.Second, "Delay between retries. Must be >= 1s. [env: HEARTBEATS_RETRY_DELAY]")

	// Meta
	var showHelp, showVersion bool
	fs.BoolVarP(&showHelp, "help", "h", false, "Show help and exit")
	fs.BoolVar(&showVersion, "version", false, "Print version and exit")

	// Custom usage message.
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [flags]\n\nFlags:\n", strings.ToLower(fs.Name())) // nolint:errcheck
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}

	if showVersion {
		return Options{}, &HelpRequested{Message: fmt.Sprintf("%s version %s\n", fs.Name(), version)}
	}
	if showHelp {
		// Capture custom usage output into buffer
		var buf bytes.Buffer
		fs.SetOutput(&buf)
		fs.Usage()
		return Options{}, &HelpRequested{Message: buf.String()}
	}

	bindEnvFromUsage(fs, getEnv)

	return Options{
		ConfigPath:  *configPath,
		ListenAddr:  *listenAddress,
		SiteRoot:    *siteRoot,
		HistorySize: *historySize,
		Debug:       *debug,
		LogFormat:   logging.LogFormat(*logFormat),
		SkipTLS:     *skipTLS,
		RetryCount:  *retryCount,
		RetryDelay:  *retryDelay,
	}, nil
}

// bindEnvFromUsage overrides unset flags using env vars extracted from usage annotations.
func bindEnvFromUsage(fs *flag.FlagSet, getEnv func(string) string) {
	fs.VisitAll(func(f *flag.Flag) {
		// Try to extract "env: VAR" from the usage string
		usage := f.Usage
		prefix := "[env:"
		start := strings.Index(usage, prefix)
		if start == -1 {
			return
		}
		end := strings.Index(usage[start:], "]")
		if end == -1 {
			return
		}
		envVar := strings.TrimSpace(usage[start+len(prefix) : start+end])

		// Don't override if flag was explicitly set
		if fs.Changed(f.Name) {
			return
		}

		envVal := getEnv(envVar)
		if envVal == "" {
			return
		}

		fs.Set(f.Name, envVal) // nolint:errcheck
	})
}

// Validate checks whether the Config is semantically valid.
func (c *Options) Validate() error {
	if c.RetryCount == 0 || c.RetryCount < -1 {
		return fmt.Errorf("retry count must be -1 (infinite) or >= 1, got %d", c.RetryCount)
	}
	if c.RetryDelay < time.Second {
		return fmt.Errorf("retry delay must be at least 1s, got %s", c.RetryDelay)
	}
	if c.LogFormat != logging.LogFormatText && c.LogFormat != logging.LogFormatJSON {
		return fmt.Errorf("invalid log format: '%s'", c.LogFormat)
	}
	return nil
}
