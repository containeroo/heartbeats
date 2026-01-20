package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"syscall"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/debugserver"
	"github.com/containeroo/heartbeats/internal/flag"
	"github.com/containeroo/heartbeats/internal/handler"
	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/containeroo/heartbeats/internal/server"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"

	"github.com/containeroo/tinyflags"
)

// Run is the single entry point for the application.
func Run(ctx context.Context, webFS fs.FS, version, commit string, args []string, w io.Writer) error {
	// Create a context to listen for shutdown signals
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	// Create another context to listen for reload signals
	reloadCh := make(chan os.Signal, 1)
	signal.Notify(reloadCh, syscall.SIGHUP)

	flags, err := flag.ParseFlags(args, version)
	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			fmt.Fprint(w, err.Error()) // nolint:errcheck
			return nil
		}
		return fmt.Errorf("CLI flags error: %w", err)
	}

	cfg, err := config.LoadConfig(flags.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid YAML config: %w", err)
	}

	logger := logging.SetupLogger(flags.LogFormat, flags.Debug, w)
	logging.SystemLogger(logger, nil).Info("Starting Heartbeats", "version", version, "commit", commit)

	if len(flags.OverriddenValues) > 0 {
		logging.SystemLogger(logger, nil).Info("CLI Overrides", "overrides", flags.OverriddenValues)
	}

	histStore, err := history.InitializeHistory(flags.HistorySize)
	if err != nil {
		return fmt.Errorf("failed to initialize history: %w", err)
	}
	histRecorder := servicehistory.NewRecorder(histStore)
	metricsReg := metrics.New(histStore)

	store := notifier.InitializeStore(cfg.Receivers, flags.SkipTLS, version, logger)
	bufferSize := len(cfg.Heartbeats) // 1 slot per actor; each heartbeat sends max 1 notification concurrently
	dispatcher := notifier.NewDispatcher(
		store,
		logger,
		histRecorder,
		flags.RetryCount,
		flags.RetryDelay,
		bufferSize,
		metricsReg,
	)
	go dispatcher.Run(ctx)

	// Create Heartbeat Manager
	actorFactory := heartbeat.DefaultActorFactory{
		Logger:     logger,
		History:    histRecorder,
		Metrics:    metricsReg,
		DispatchCh: dispatcher.Mailbox(),
	}
	mgr, err := heartbeat.NewManagerFromHeartbeatMap(
		ctx,
		cfg.Heartbeats,
		logger,
		actorFactory,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize heartbeats: %w", err)
	}
	mgr.StartAll()

	reloadConfig := config.NewReloadFunc(flags.ConfigPath, flags.SkipTLS, version, logger, dispatcher, mgr)
	go config.WatchReload(ctx, reloadCh, logger, reloadConfig)

	api := handler.NewAPI(
		version,
		commit,
		webFS,
		flags.SiteRoot,
		flags.RoutePrefix,
		flags.Debug,
		logger,
		mgr,
		histStore,
		histRecorder,
		dispatcher,
		metricsReg,
		reloadConfig,
	)

	// Run debug server if enabled
	if flags.Debug {
		debugserver.Run(ctx, flags.DebugServerPort, api)
	}

	// Create server and run forever
	router, err := server.NewRouter(webFS, flags.RoutePrefix, api, flags.Debug)
	if err != nil {
		return fmt.Errorf("configure router: %w", err)
	}
	if err := server.Run(ctx, flags.ListenAddr, router, logger); err != nil {
		return fmt.Errorf("failed to run Heartbeats: %w", err)
	}

	return nil
}
