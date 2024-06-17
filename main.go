package main

import (
	"context"
	"embed"
	"fmt"
	"heartbeats/pkg/config"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/notify"
	"heartbeats/pkg/server"
	"os"

	"github.com/alecthomas/kingpin/v2"
)

const version = "0.6.10"

//go:embed web
var templates embed.FS

var (
	configPath    = kingpin.Flag("config", "Path to the configuration file").Short('c').Envar("HEARTBEATS_CONFIG").Default("./deploy/config.yaml").String()
	listenAddress = kingpin.Flag("listen-address", "Address to listen on").Short('l').Envar("HEARTBEATS_LISTEN_ADDRESS").Default("localhost:8080").String()
	siteRoot      = kingpin.Flag("site-root", "Site root for the heartbeat service").Short('s').Envar("HEARTBEATS_SITE_ROOT").Default("http://<listenaddress>").String()
	maxSize       = kingpin.Flag("max-size", "Maximum size of the cache").Short('m').Envar("HEARTBEATS_MAX_SIZE").Default("1000").Int()
	reduce        = kingpin.Flag("reduce", "Percentage to reduce when max size is exceeded").Short('r').Envar("HEARTBEATS_REDUCE").Default("25").Int()
	verbose       = kingpin.Flag("verbose", "Enable verbose logging").Short('v').Envar("HEARTBEATS_VERBOSE").Bool()
)

// run initializes and runs the server with the provided context and verbosity settings.
func run(ctx context.Context, verbose bool) error {
	kingpin.Version(version)
	kingpin.Parse()

	log := logger.NewLogger(verbose)

	heartbeatsStore := heartbeat.NewStore()
	notificationStore := notify.NewStore()
	historyStore := history.NewStore()

	if err := config.Read(
		*configPath,
		history.Config{
			MaxSize: *maxSize,
			Reduce:  *reduce,
		},
		heartbeatsStore,
		notificationStore,
		historyStore,
	); err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	if err := config.Validate(heartbeatsStore, notificationStore); err != nil {
		return fmt.Errorf("Error validating config file: %v", err)
	}

	return server.Run(
		ctx,
		log,
		version,
		server.Config{
			ListenAddress: *listenAddress,
			SiteRoot:      *siteRoot,
		},
		templates,
		heartbeatsStore,
		notificationStore,
		historyStore,
	)
}

func main() {
	ctx := context.Background()
	if err := run(ctx, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
