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
	"github.com/containeroo/heartbeats/internal/flag"
	"github.com/containeroo/heartbeats/internal/handler"
	"github.com/containeroo/heartbeats/internal/heartbeat/manager"
	"github.com/containeroo/heartbeats/internal/heartbeat/reconcile"
	"github.com/containeroo/heartbeats/internal/heartbeat/service"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notify/dispatch"
	"github.com/containeroo/heartbeats/internal/routes"
	"github.com/containeroo/heartbeats/internal/ws"

	"github.com/containeroo/httpgrace/server"
	"github.com/containeroo/tinyflags"
)

// Run is the single entry point for the application.
func Run(ctx context.Context, appFS fs.FS, version, commit string, args []string, w io.Writer) error {
	// Create a context to listen for shutdown signals
	// Cancel on SIGINT/SIGTERM for graceful shutdown.
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create another context to listen for reload signals
	reloadCh := make(chan os.Signal, 1)
	signal.Notify(reloadCh, syscall.SIGHUP)

	flags, err := flag.ParseFlags(args, version)
	logger := logging.SetupLogger(flags.LogFormat, flags.Debug, w)
	sysLogger := logging.SystemLogger(logger)
	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			fmt.Fprint(w, err.Error()) // nolint:errcheck
			return nil
		}
		sysLogger.Error("application failed",
			"event", "app_failed",
			"stage", "parse_flags",
			"err", err,
		)
		return fmt.Errorf("CLI flags error: %w", err)
	}

	sysLogger.Info("Starting heartbeats",
		"version", version,
		"commit", commit,
	)
	if len(flags.OverriddenValues) > 0 {
		sysLogger.Info("CLI Overrides", "overrides", flags.OverriddenValues)
	}

	businessLogger := logging.BusinessLogger(logger)
	accessLogger := logging.AccessLogger(logger)

	cfg, err := config.LoadWithOptions(flags.ConfigPath, config.LoadOptions{StrictEnv: flags.StrictEnv})
	if err != nil {
		sysLogger.Error("application failed",
			"event", "app_failed",
			"stage", "load_config",
			"err", err,
		)
		return fmt.Errorf("load config: %w", err)
	}

	api := handler.NewAPI(
		version,
		commit,
		flags.SiteRoot,
		accessLogger,
	)

	metricsReg := metrics.NewRegistry()
	api.SetMetrics(metricsReg)

	historyStore := history.NewStore(cfg.History.Size)
	historyRecorder := history.NewAsyncRecorder(historyStore, businessLogger, cfg.History.Buffer)
	historyRecorder.Start(ctx)
	api.SetHistory(historyRecorder)

	notifyManager := dispatch.NewManager(businessLogger, historyRecorder, metricsReg)
	notifyManager.Start(ctx)

	manager, err := manager.NewManager(cfg, appFS, notifyManager, historyRecorder, metricsReg, businessLogger)
	if err != nil {
		sysLogger.Error("application failed",
			"event", "app_failed",
			"stage", "build_manager",
			"err", err,
		)
		return fmt.Errorf("build manager: %w", err)
	}
	manager.StartAll(ctx)

	svc := service.NewService(manager, notifyManager, historyRecorder)
	api.SetService(svc)

	reloadConfigFn := reconcile.NewReloadFunc(
		ctx,
		flags.ConfigPath,
		appFS,
		config.LoadOptions{StrictEnv: flags.StrictEnv},
		sysLogger,
		manager,
	)
	go reconcile.WatchReload(ctx, reloadCh, sysLogger, reloadConfigFn)
	api.SetReloadFn(reloadConfigFn)

	wsHub := ws.NewHub(accessLogger, ws.Providers{
		Heartbeats:    svc.HeartbeatSummaries,
		Receivers:     svc.ReceiverSummaries,
		History:       svc.HistorySnapshot,
		HeartbeatByID: svc.HeartbeatSummaryByID,
		ReceiverByKey: svc.ReceiverSummaryByKey,
	})
	wsHub.Start(ctx, svc.HistoryStream)
	api.SetHub(wsHub)

	// Create server and run forever
	router, err := routes.NewRouter(appFS, api, flags.RoutePrefix, sysLogger)
	if err != nil {
		sysLogger.Error("application failed",
			"event", "app_failed",
			"stage", "create_router",
			"err", err,
		)
		return fmt.Errorf("create router: %w", err)
	}
	if err := server.Run(ctx, flags.ListenAddr, router, sysLogger); err != nil {
		sysLogger.Error("application failed",
			"event", "app_failed",
			"stage", "run_server",
			"err", err,
		)
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}
