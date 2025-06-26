package flag

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/pkg/envflag"

	flag "github.com/spf13/pflag"
)

// HelpRequested indicates that help was requested.
type HelpRequested struct {
	Message string // Message is the help message.
}

// Error returns the help message.
func (e *HelpRequested) Error() string { return e.Message }

// Options holds the application configuration.
type Options struct {
	Debug           bool                // Set LogLevel to Debug
	LogFormat       logging.LogFormat   // Specify the log output format
	ConfigPath      string              // Path to the configuration file
	ListenAddr      string              // Address to listen on
	SiteRoot        string              // Root URL of the site
	SkipTLS         bool                // Skip TLS for all receivers
	RetryCount      int                 // Number of retries for notifications
	RetryDelay      time.Duration       // Delay between retries
	DebugServerPort int                 // Port for the debug server
	HistoryBackend  history.BackendType // Backend for history: ring | badger
	RingStoreSize   int                 // Size of the history ring buffer
	BadgerPath      string              // Path to the badger directory
}

// ParseFlags parses flags and environment variables.
func ParseFlags(args []string, version string, getEnv func(string) string) (Options, error) {
	fs := flag.NewFlagSet("Heartbeats", flag.ContinueOnError)
	fs.SortFlags = false

	// Register flags
	fs.StringP("config", "c", "config.yaml", "Path to configuration file (env: HEARTBEATS_CONFIG)")
	fs.StringP("listen-address", "a", ":8080", "Address to listen on (env: HEARTBEATS_LISTEN_ADDRESS)")
	fs.StringP("site-root", "r", "http://localhost:8080", "Site root URL (env: HEARTBEATS_SITE_ROOT)")
	fs.Bool("skip-tls", false, "Skip TLS verification (env: HEARTBEATS_SKIP_TLS)")
	fs.BoolP("debug", "d", false, "Enable debug logging (env: HEARTBEATS_DEBUG)")
	fs.Int("debug-server-port", 8081, "Port for the debug server (env: HEARTBEATS_DEBUG_SERVER_PORT)")
	fs.StringP("log-format", "l", "json", "Log format: json | text (env: HEARTBEATS_LOG_FORMAT)")
	fs.Int("retry-count", 3, "Retries for failed notifications (-1 = infinite) (env: HEARTBEATS_RETRY_COUNT)")
	fs.Duration("retry-delay", 2*time.Second, "Delay between retries (env: HEARTBEATS_RETRY_DELAY)")

	// History flags
	fs.String("history-backend", string(history.BackendTypeRingStore), "Backend for history: ring | badger (env: HEARTBEATS_HISTORY_BACKEND)")
	fs.String("badger-path", "db", "Path to the badger directory (env: HEARTBEATS_BADGER_PATH)")
	fs.Int("ring-size", 10000, "Size of the ringstore (env: HEARTBEAT_RING_SIZE)")

	// Meta flags
	var showHelp, showVersion bool
	fs.BoolVarP(&showHelp, "help", "h", false, "Show help and exit")
	fs.BoolVar(&showVersion, "version", false, "Print version and exit")

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
		var buf bytes.Buffer
		fs.SetOutput(&buf)
		fs.Usage()
		return Options{}, &HelpRequested{Message: buf.String()}
	}

	return buildOptions(fs)
}

// buildOptions resolves all values from flags, env, or defaults.
func buildOptions(fs *flag.FlagSet) (opts Options, err error) {
	defer func() {
		if r := recover(); r != nil {
			// catch panic from must(...) calls to avoid repetitive `if err != nil` checks
			// and convert them into a single error return instead
			err = fmt.Errorf("failed to parse flags: %v", r)
			opts = Options{} // zero value
		}
	}()
	return Options{
		ConfigPath:      must(envflag.String(fs, "config", "HEARTBEATS_CONFIG")),
		ListenAddr:      must(envflag.HostPort(fs, "listen-address", "HEARTBEATS_LISTEN_ADDRESS")),
		SiteRoot:        must(envflag.URL(fs, "site-root", "HEARTBEATS_SITE_ROOT")),
		Debug:           must(envflag.Bool(fs, "debug", "HEARTBEATS_DEBUG")),
		DebugServerPort: must(envflag.Int(fs, "debug-server-port", "HEARTBEATS_DEBUG_SERVER_PORT")),
		SkipTLS:         must(envflag.Bool(fs, "skip-tls", "HEARTBEATS_SKIP_TLS")),
		RetryCount:      must(envflag.Int(fs, "retry-count", "HEARTBEATS_RETRY_COUNT")),
		RetryDelay:      must(envflag.Duration(fs, "retry-delay", "HEARTBEATS_RETRY_DELAY")),
		LogFormat:       logging.LogFormat(must(envflag.String(fs, "log-format", "HEARTBEATS_LOG_FORMAT"))),
		HistoryBackend:  history.BackendType(must(envflag.String(fs, "history-backend", "HEARTBEATS_HISTORY_BACKEND"))),
		RingStoreSize:   must(envflag.Int(fs, "ring-size", "HEARTBEATS_HISTORY_SIZE")),
		BadgerPath:      must(envflag.String(fs, "badger-path", "HEARTBEATS_BADGER_PATH")),
	}, nil
}

// must panics on err and is used to keep config assembly clean.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// Validate checks whether the Options are valid.
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
	if c.HistoryBackend != history.BackendTypeBadger && c.HistoryBackend != history.BackendTypeRingStore {
		return fmt.Errorf("invalid history backend: '%s' (must be 'ring' or 'badger')", c.HistoryBackend)
	}

	return nil
}
