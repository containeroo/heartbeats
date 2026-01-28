package reconcile

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sync"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/heartbeat/manager"
)

// Reload loads, validates, and applies the config to the manager.
func Reload(
	ctx context.Context,
	filePath string,
	templateFS fs.FS,
	opts config.LoadOptions,
	logger *slog.Logger,
	mgr *manager.Manager,
) (manager.ReloadResult, error) {
	cfg, err := config.LoadWithOptions(filePath, opts)
	if err != nil {
		return manager.ReloadResult{}, fmt.Errorf("failed to load config: %w", err)
	}
	res, err := mgr.Reload(ctx, cfg, templateFS)
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
	opts config.LoadOptions,
	logger *slog.Logger,
	mgr *manager.Manager,
) func() error {
	var reloadMu sync.Mutex
	return func() error {
		reloadMu.Lock()
		defer reloadMu.Unlock()

		res, err := Reload(ctx, filePath, templateFS, opts, logger, mgr)
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
