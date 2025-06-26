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
	"github.com/containeroo/heartbeats/internal/debugserver"
	"github.com/containeroo/heartbeats/internal/flag"
	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/containeroo/heartbeats/internal/server"
)

// Run is the single entry point for the application.
func Run(ctx context.Context, webFS fs.FS, version, commit string, args []string, w io.Writer, getEnv func(string) string) error {
	// Parse and validate command-line flags.
	flags, err := flag.ParseFlags(args, version, getEnv)
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
	history, err := history.InitializeHistory(flags)
	if err != nil {
		return fmt.Errorf("failed to initialize history: %w", err)
	}

	// Inizalize notification
	store := notifier.InitializeStore(cfg.Receivers, flags.SkipTLS, version, logger)
	bufferSize := len(cfg.Heartbeats) // 1 slot per actor; each heartbeat sends max 1 notification concurrently
	dispatcher := notifier.NewDispatcher(
		store,
		logger,
		history,
		flags.RetryCount,
		flags.RetryDelay,
		bufferSize,
	)
	go dispatcher.Run(ctx)

	// Create Heartbeat Manager
	mgr := heartbeat.NewManagerFromHeartbeatMap(
		ctx,
		cfg.Heartbeats,
		dispatcher.Mailbox(),
		history,
		logger,
	)

	// Run debug server if enabled
	if flags.Debug {
		debugserver.Run(ctx, flags.DebugServerPort, mgr, dispatcher, logger)
	}

	// Create server and run forever
	router := server.NewRouter(
		webFS,
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
