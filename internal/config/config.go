package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/notifier"
	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration.
type Config struct {
	Receivers  map[string]notifier.ReceiverConfig `yaml:"receivers"`  // Receivers is the map of receiver IDs to their configurations.
	Heartbeats heartbeat.HeartbeatConfigMap       `yaml:"heartbeats"` // Heartbeats is the map of heartbeat IDs to their configurations.
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate checks that each heartbeat references an existing receiver and
// validates and resolves secrets in receiver configurations.
func (c *Config) Validate() error {
	for receiverID, rc := range c.Receivers {
		for i := range rc.SlackConfigs {
			if err := rc.SlackConfigs[i].Resolve(); err != nil {
				return fmt.Errorf("receiver %q slack config error: %w", receiverID, err)
			}
			if err := rc.SlackConfigs[i].Validate(); err != nil {
				return fmt.Errorf("receiver %q slack config error: %w", receiverID, err)
			}
		}
		for i := range rc.EmailConfigs {
			if err := rc.EmailConfigs[i].Resolve(); err != nil {
				return fmt.Errorf("receiver %q email config error: %w", receiverID, err)
			}
			if err := rc.EmailConfigs[i].Validate(); err != nil {
				return fmt.Errorf("receiver %q email config error: %w", receiverID, err)
			}
		}
		for i := range rc.MSTeamsConfigs {
			if err := rc.MSTeamsConfigs[i].Resolve(); err != nil {
				return fmt.Errorf("receiver %q MSTeams config error: %w", receiverID, err)
			}
			if err := rc.MSTeamsConfigs[i].Validate(); err != nil {
				return fmt.Errorf("receiver %q MSTeams config error: %w", receiverID, err)
			}
		}
		for i := range rc.MSTeamsGraphConfig {
			if err := rc.MSTeamsGraphConfig[i].Resolve(); err != nil {
				return fmt.Errorf("receiver %q MSTeams Graph config error: %w", receiverID, err)
			}
			if err := rc.MSTeamsGraphConfig[i].Validate(); err != nil {
				return fmt.Errorf("receiver %q MSTeams Graph config error: %w", receiverID, err)
			}
		}

		c.Receivers[receiverID] = rc // Write back updated receiver config.
	}

	// Validate that each heartbeat references valid receivers.
	for hbName, hb := range c.Heartbeats {
		for _, receiverName := range hb.Receivers {
			if _, ok := c.Receivers[receiverName]; !ok {
				return fmt.Errorf("heartbeat %q references unknown receiver %q", hbName, receiverName)
			}
		}
	}
	if len(c.Heartbeats) == 0 {
		return fmt.Errorf("at least one heartbeat must be defined")
	}
	return nil
}

// Reload loads, validates, and applies the config to the dispatcher and manager.
func Reload(
	filePath string,
	skipTLS bool,
	version string,
	logger *slog.Logger,
	dispatcher *notifier.Dispatcher,
	mgr *heartbeat.Manager,
) (heartbeat.ReconcileResult, error) {
	cfg, err := LoadConfig(filePath)
	if err != nil {
		return heartbeat.ReconcileResult{}, fmt.Errorf("failed to load config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return heartbeat.ReconcileResult{}, fmt.Errorf("invalid YAML config: %w", err)
	}

	newStore := notifier.InitializeStore(cfg.Receivers, skipTLS, version, logger)
	dispatcher.UpdateStore(newStore)

	res, err := mgr.Reconcile(cfg.Heartbeats)
	if err != nil {
		return res, err
	}
	return res, nil
}

// NewReloadFunc builds a reload function with logging and serialization.
func NewReloadFunc(
	filePath string,
	skipTLS bool,
	version string,
	logger *slog.Logger,
	dispatcher *notifier.Dispatcher,
	mgr *heartbeat.Manager,
) func() error {
	var reloadMu sync.Mutex
	return func() error {
		reloadMu.Lock()
		defer reloadMu.Unlock()

		res, err := Reload(filePath, skipTLS, version, logger, dispatcher, mgr)
		if err != nil {
			return err
		}
		if res.Added+res.Updated+res.Removed > 0 {
			logging.SystemLogger(logger, nil).Info("heartbeats reloaded",
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
			logging.SystemLogger(logger, nil).Info("reload requested")
			if err := reloadFn(); err != nil {
				logging.SystemLogger(logger, nil).Error("reload failed", "err", err)
			} else {
				logging.SystemLogger(logger, nil).Info("reload completed")
			}
		}
	}
}
