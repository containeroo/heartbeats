package flags

import (
	"flag"
	"fmt"
	"heartbeats/internal/config"
	"os"

	"github.com/spf13/pflag"
)

// ParseFlags initializes the command-line flags and sets the values in the global config.App.
func ParseFlags(currentVersion string) error {
	var showVersion, showHelp bool

	pflag.StringVarP(&config.App.Path, "config", "c", "./deploy/config.yaml", "Path to the configuration file")
	pflag.StringVarP(&config.App.Server.ListenAddress, "listen-address", "l", "localhost:8080", "Address to listen on")
	pflag.StringVarP(&config.App.Server.SiteRoot, "siteRoot", "s", "", "Site root for the heartbeat service (default \"http://<listenAddress>\")")
	pflag.IntVarP(&config.App.Cache.MaxSize, "maxSize", "m", 100, "Maximum size of the cache")
	pflag.IntVarP(&config.App.Cache.Reduce, "reduce", "r", 10, "Amount to reduce when max size is exceeded")
	pflag.BoolVarP(&config.App.Verbose, "verbose", "v", false, "Enable verbose logging")
	pflag.BoolVar(&showVersion, "version", false, "Show version and exit")
	pflag.BoolVarP(&showHelp, "help", "h", false, "Show help and exit")

	// Disable default help flag
	pflag.CommandLine.SortFlags = false
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine.Init("heartbeats", pflag.ExitOnError)
	pflag.CommandLine.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		pflag.PrintDefaults()
	}

	pflag.Parse()

	// Show help and exit
	if showHelp {
		pflag.Usage()
		os.Exit(0)
	}

	// Show version and exit
	if showVersion {
		fmt.Println(currentVersion)
		os.Exit(0)
	}

	if config.App.Server.SiteRoot == "" {
		config.App.Server.SiteRoot = fmt.Sprintf("http://%s", config.App.Server.ListenAddress)
	}

	config.App.Version = currentVersion

	return nil
}
