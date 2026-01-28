package manager

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/heartbeat/types"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/metrics"
	ntypes "github.com/containeroo/heartbeats/internal/notify/types"
	"github.com/stretchr/testify/require"
)

type noopNotifier struct{}

func (noopNotifier) Enqueue(n ntypes.Notification) string { return "ok" }

func TestNewManager(t *testing.T) {
	t.Parallel()
	templateFS := os.DirFS(filepath.Join("..", "..", ".."))
	cfg := sampleConfig()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
		mgr, err := NewManager(cfg, templateFS, noopNotifier{}, history.NewStore(10), metrics.NewRegistry(), logger)
		require.NoError(t, err)
		require.NotNil(t, mgr)
	})

	t.Run("nil config error", func(t *testing.T) {
		t.Parallel()
		logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
		_, err := NewManager(nil, templateFS, noopNotifier{}, history.NewStore(10), metrics.NewRegistry(), logger)
		require.Error(t, err)
	})
}

func TestManagerGetAll(t *testing.T) {
	t.Parallel()
	manager := setupManager(t)
	defer manager.StopAll()

	t.Run("get existing", func(t *testing.T) {
		t.Parallel()
		hb, ok := manager.Get("api")
		require.True(t, ok)
		require.Equal(t, "api", hb.ID)
	})

	t.Run("missing returns false", func(t *testing.T) {
		t.Parallel()
		_, ok := manager.Get("missing")
		require.False(t, ok)
	})

	t.Run("all returns heartbeat list", func(t *testing.T) {
		t.Parallel()
		all := manager.All()
		require.Len(t, all, 1)
	})
}

func TestManagerReloadErrors(t *testing.T) {
	t.Parallel()
	templateFS := os.DirFS(filepath.Join("..", "..", ".."))
	var manager *Manager
	ctx := context.Background()

	t.Run("nil manager", func(t *testing.T) {
		t.Parallel()
		_, err := manager.Reload(ctx, sampleConfig(), templateFS)
		require.Error(t, err)
	})

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()
		mgr := setupManager(t)
		defer mgr.StopAll()
		_, err := mgr.Reload(ctx, nil, templateFS)
		require.Error(t, err)
	})
}

func TestDiffHeartbeatSets(t *testing.T) {
	t.Parallel()
	t.Run("diff counts", func(t *testing.T) {
		t.Parallel()
		old := map[string]*types.Heartbeat{
			"a": {ID: "a"},
		}
		newSet := map[string]*types.Heartbeat{
			"b": {ID: "b"},
		}
		result := diffHeartbeatSets(old, newSet)
		require.Equal(t, 1, result.Added)
		require.Equal(t, 0, result.Updated)
		require.Equal(t, 1, result.Removed)
	})
}

func sampleConfig() *config.Config {
	return &config.Config{
		Receivers: map[string]config.ReceiverConfig{
			"ops": {
				Webhooks: []config.WebhookConfig{
					{URL: "https://example.com"},
				},
			},
		},
		Heartbeats: map[string]config.HeartbeatConfig{
			"api": {
				Interval:  time.Second,
				LateAfter: time.Second,
				Receivers: []string{"ops"},
			},
		},
		History: config.HistoryConfig{
			Size:   10,
			Buffer: 2,
		},
	}
}

func setupManager(t *testing.T) *Manager {
	t.Helper()
	templateFS := os.DirFS(filepath.Join("..", "..", ".."))
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
	manager, err := NewManager(sampleConfig(), templateFS, noopNotifier{}, history.NewStore(10), metrics.NewRegistry(), logger)
	require.NoError(t, err)
	return manager
}
