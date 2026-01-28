package reconcile

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/heartbeat/manager"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/metrics"
	ntypes "github.com/containeroo/heartbeats/internal/notify/types"
	"github.com/stretchr/testify/require"
)

type noopNotifier struct{}

func (noopNotifier) Enqueue(ntypes.Notification) string { return "ok" }

const sampleConfigYAML = `
receivers:
  ops:
    webhooks:
      - url: https://example.com
heartbeats:
  api:
    interval: 1s
    late_after: 2s
    receivers: ["ops"]
history:
  size: 10
`

func TestReload(t *testing.T) {
	t.Parallel()

	templateFS := os.DirFS(filepath.Join("..", "..", ".."))
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		cfgPath := writeConfig(t, sampleConfigYAML)
		cfg, err := config.Load(cfgPath)
		require.NoError(t, err)
		mgr, err := manager.NewManager(cfg, templateFS, noopNotifier{}, history.NewStore(10), metrics.NewRegistry(), logger)
		require.NoError(t, err)

		res, err := Reload(context.Background(), cfgPath, templateFS, config.LoadOptions{}, logger, mgr)
		require.NoError(t, err)
		require.Zero(t, res.Added+res.Updated+res.Removed)
	})

	t.Run("load error", func(t *testing.T) {
		t.Parallel()

		_, err := Reload(context.Background(), filepath.Join(t.TempDir(), "missing.yaml"), templateFS, config.LoadOptions{}, logger, &manager.Manager{})
		require.Error(t, err)
	})
}

func TestNewReloadFunc(t *testing.T) {
	t.Parallel()

	templateFS := os.DirFS(filepath.Join("..", "..", ".."))
	ctx := context.Background()
	logBuf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfgPath := writeConfig(t, sampleConfigYAML)
	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)
	mgr, err := manager.NewManager(cfg, templateFS, noopNotifier{}, history.NewStore(10), metrics.NewRegistry(), logger)
	require.NoError(t, err)

	reloadFn := NewReloadFunc(ctx, cfgPath, templateFS, config.LoadOptions{}, logger, mgr)

	require.NoError(t, os.WriteFile(cfgPath, []byte(`
receivers:
  ops:
    webhooks:
      - url: https://example.com
  pager:
    webhooks:
      - url: https://example.net
heartbeats:
  api:
    interval: 1s
    late_after: 2s
    receivers: ["ops"]
  db:
    interval: 2s
    late_after: 3s
    receivers: ["pager"]
history:
  size: 10
`), 0o600))

	require.NoError(t, reloadFn())
	require.Contains(t, logBuf.String(), "heartbeats reloaded")
}

func TestWatchReload(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))

	t.Run("trigger reload", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		calls := make(chan struct{}, 1)
		reloadCh := make(chan os.Signal, 1)
		go WatchReload(ctx, reloadCh, logger, func() error {
			calls <- struct{}{}
			return nil
		})

		reloadCh <- os.Interrupt
		select {
		case <-time.After(time.Second):
			t.Fatal("reload not triggered")
		case <-calls:
		}
	})

	t.Run("log on failure", func(t *testing.T) {
		t.Parallel()

		logBuf := &bytes.Buffer{}
		logger := slog.New(slog.NewJSONHandler(logBuf, &slog.HandlerOptions{}))
		reloadCh := make(chan os.Signal, 1)
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		go WatchReload(ctx, reloadCh, logger, func() error {
			return errors.New("boom")
		})

		reloadCh <- os.Interrupt
		time.Sleep(10 * time.Millisecond)
		require.Contains(t, logBuf.String(), "reload failed")
	})
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}
