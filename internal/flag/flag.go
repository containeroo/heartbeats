package flag

import (
	"bytes"
	"fmt"
	"strings"

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

// Config holds the application configuration.
type Config struct {
	Debug       bool              // Set LogLevel to Debug
	LogFormat   logging.LogFormat // Specify the log output format
	ConfigPath  string            // ConfigPath is the path to the configuration file.
	ListenAddr  string            // ListenAddress is the address to listen on.
	SiteRoot    string            // SiteRoot is the root URL of the site.
	HistorySize int               // HistorySize is the size of the history ring buffer.
	SkipTLS     bool              // SkipTLS for all receivers
}

// ParseFlags parses command-line flags.
func ParseFlags(args []string, version string) (Config, error) {
	fs := flag.NewFlagSet("Heartbeats", flag.ContinueOnError)
	fs.SortFlags = false

	// Server settings
	configPath := fs.StringP("config", "c", "config.yaml", "Path to configuration file")
	listenAddress := fs.StringP("listen-address", "a", ":8080", "Address to listen on")
	siteRoot := fs.StringP("site-root", "r", "http://localhost:8080", "Site root URL")
	historySize := fs.IntP("history-size", "s", 10000, "Size of the history")

	// Application logging.
	skipTLS := fs.Bool("skip-tls", false, "Skip TLS verification for all receivers (can be overridden per receiver)")
	debug := fs.BoolP("debug", "d", false, "Enable debug logging (default: false)")
	logFormat := fs.StringP("log-format", "l", "json", "Log format (json | text)")

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
		return Config{}, err
	}

	if showVersion {
		return Config{}, &HelpRequested{Message: fmt.Sprintf("%s version %s\n", fs.Name(), version)}
	}
	if showHelp {
		// Capture custom usage output into buffer
		var buf bytes.Buffer
		fs.SetOutput(&buf)
		fs.Usage()
		return Config{}, &HelpRequested{Message: buf.String()}
	}

	if *logFormat != "json" && *logFormat != "text" {
		return Config{}, fmt.Errorf("invalid log format: '%s'", *logFormat)
	}

	return Config{
		ConfigPath:  *configPath,
		ListenAddr:  *listenAddress,
		SiteRoot:    *siteRoot,
		HistorySize: *historySize,
		Debug:       *debug,
		LogFormat:   logging.LogFormat(*logFormat),
		SkipTLS:     *skipTLS,
	}, nil
}
