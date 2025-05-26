package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os/signal"
	"syscall"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/flag"
	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/containeroo/heartbeats/internal/server"
)

// Run is the single entry point for the application.
func Run(ctx context.Context, staticFS fs.FS, version, commit string, args []string, w io.Writer) error {
	// Parse and validate command-line flags.
	flags, err := flag.ParseFlags(args, version)
	if err != nil {
		var helpErr *flag.HelpRequested
		if errors.As(err, &helpErr) {
			fmt.Fprint(w, helpErr.Error()) // nolint:errcheck
			return nil
		}
		return fmt.Errorf("parsing error: %w", err)
	}
	if err := flags.Validate(); err != nil {
		return fmt.Errorf("invalid CLI flags: %w", err)
	}

	// Load and validate configuration.
	cfg, err := config.LoadConfig(flags.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid YAML config: %w", err)
	}

	// Setup logger
	logger := logging.SetupLogger(flags.LogFormat, flags.Debug, w)
	logger.Info("Starting Heartbeats",
		"version", version,
		"commit", commit,
	)

	// Create a context to listen for shutdown signals
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create history
	history := history.NewRingStore(flags.HistorySize)

	// Inizalize notification
	store := notifier.InitializeStore(cfg.Receivers, false, logger)
	dispatcher := notifier.NewDispatcher(
		store,
		logger,
		history,
		flags.RetryCount,
		flags.RetryDelay,
	)

	// Create Heartbeat Manager
	mgr := heartbeat.NewManager(
		ctx,
		cfg.Heartbeats,
		dispatcher,
		history,
		logger,
	)

	// Create server and run forever
	router := server.NewRouter(
		staticFS,
		flags.SiteRoot,
		version,
		mgr,
		history,
		dispatcher,
		logger,
		flags.Debug,
	)
	if err := server.Run(ctx, flags.ListenAddr, router, logger); err != nil {
		return fmt.Errorf("failed to run Heartbeats: %w", err)
	}

	return nil
}
