package reconcile

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sync"

	kit "github.com/containeroo/notifykit/notify"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/heartbeat/manager"
	"github.com/containeroo/heartbeats/internal/notify"
)

// Reload loads, validates, and applies the config to the manager.
func Reload(
	ctx context.Context,
	filePath string,
	templateFS fs.FS,
	receivers kit.Receivers,
	routes notify.ReceiverRoutes,
	opts config.LoadOptions,
	logger *slog.Logger,
	mgr *manager.Manager,
) (manager.ReloadResult, error) {
	cfg, err := config.LoadWithOptions(filePath, opts)
	if err != nil {
		return manager.ReloadResult{}, fmt.Errorf("failed to load config: %w", err)
	}
	nextReceivers, nextRoutes, err := notify.ReceiversFromConfig(templateFS, cfg, logger)
	if err != nil {
		return manager.ReloadResult{}, fmt.Errorf("build receivers: %w", err)
	}
	notify.ReplaceReceivers(receivers, nextReceivers)
	notify.ReplaceRoutes(routes, nextRoutes)
	res, err := mgr.Reload(ctx, cfg, routes)
	if err != nil {
		return res, err
	}
	return res, nil
}

// NewReloadFunc builds a reload function with logging and serialization.
func NewReloadFunc(
	ctx context.Context,
	filePath string,
	templateFS fs.FS,
	receivers kit.Receivers,
	routes notify.ReceiverRoutes,
	opts config.LoadOptions,
	logger *slog.Logger,
	mgr *manager.Manager,
) func() error {
	var reloadMu sync.Mutex
	return func() error {
		reloadMu.Lock()
		defer reloadMu.Unlock()

		res, err := Reload(ctx, filePath, templateFS, receivers, routes, opts, logger, mgr)
		if err != nil {
			return err
		}
		if res.Added+res.Updated+res.Removed > 0 {
			logger.Info("heartbeats reloaded",
				"added", res.Added,
				"updated", res.Updated,
				"removed", res.Removed,
			)
		}
		return nil
	}
}

// WatchReload listens for reload signals and invokes the reload function.
func WatchReload(ctx context.Context, reloadCh <-chan os.Signal, logger *slog.Logger, reloadFn func() error) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-reloadCh:
			logger.Info("reload requested")
			if err := reloadFn(); err != nil {
				logger.Error("reload failed", "err", err)
				continue
			}
			logger.Info("reload completed")
		}
	}
}
