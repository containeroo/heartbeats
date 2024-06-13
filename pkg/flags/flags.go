package flags

import (
	"fmt"
	"heartbeats/pkg/config"
	"io"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

// ParseResult contains the result of the ParseFlags function.
type ParseResult struct {
	ShowHelp    bool
	ShowVersion bool
	Err         error
}

// ParseFlags initializes the command-line flags and sets the values in the global config.App.
func ParseFlags(args []string, output io.Writer) ParseResult {
	var showVersion, showHelp bool

	pflag.StringVarP(&config.App.Path, "config", "c", "./deploy/config.yaml", "Path to the configuration file")
	pflag.StringVarP(&config.App.Server.ListenAddress, "listen-address", "l", "localhost:8080", "Address to listen on")
	pflag.StringVarP(&config.App.Server.SiteRoot, "site-root", "s", "", "Site root for the heartbeat service (default \"http://<listenAddress>\")")
	pflag.IntVarP(&config.App.Cache.MaxSize, "max-size", "m", 100, "Maximum size of the cache")
	pflag.IntVarP(&config.App.Cache.Reduce, "reduce", "r", 10, "Amount to reduce when max size is exceeded")
	pflag.BoolVarP(&config.App.Verbose, "verbose", "v", false, "Enable verbose logging")
	pflag.BoolVar(&showVersion, "version", false, "Show version and exit")
	pflag.BoolVarP(&showHelp, "help", "h", false, "Show help and exit")

	pflag.CommandLine.SetOutput(output)
	pflag.CommandLine.SortFlags = false
	pflag.CommandLine.Init("heartbeats", pflag.ExitOnError)

	// Disable default help flag
	pflag.CommandLine.Usage = func() {
		fmt.Fprintf(output, "Usage of %s:\n", args[0])
		pflag.PrintDefaults()
	}

	processEnvVariables()

	err := pflag.CommandLine.Parse(args[1:])
	if err != nil {
		return ParseResult{Err: err}
	}

	if showHelp {
		pflag.Usage()
		return ParseResult{ShowHelp: true}
	}

	if showVersion {
		return ParseResult{ShowVersion: true}
	}

	if config.App.Server.SiteRoot == "" {
		config.App.Server.SiteRoot = fmt.Sprintf("http://%s", config.App.Server.ListenAddress)
	}

	return ParseResult{}
}

// processEnvVariables checks for environment variables with the prefix "HEARTBEATS_" and sets the corresponding flags.
func processEnvVariables() {
	prefix := "HEARTBEATS_"
	pflag.VisitAll(func(f *pflag.Flag) {
		envVar := prefix + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		if val, ok := os.LookupEnv(envVar); ok {
			if err := f.Value.Set(val); err != nil {
				fmt.Fprintf(os.Stderr, "Error setting flag from environment variable %s: %v\n", envVar, err)
			}
		}
	})
}
