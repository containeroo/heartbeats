package debugserver

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/stretchr/testify/require"
)

func TestDebugServer_Run(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	// Setup basic in-memory state
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(5)
	store := notifier.NewReceiverStore()
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1*time.Millisecond, 10)

	hbCfg := map[string]heartbeat.HeartbeatConfig{
		"test-hb": {
			Description: "demo",
			Interval:    time.Second,
			Grace:       time.Second,
			Receivers:   []string{"r1"},
		},
	}
	mgr := heartbeat.NewManagerFromHeartbeatMap(ctx, hbCfg, disp.Mailbox(), hist, logger)

	// Pick a random free port
	port := 8089
	Run(ctx, port, mgr, disp, logger)

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	t.Run("receiver test handler", func(t *testing.T) {
		url := fmt.Sprintf("http://127.0.0.1:%d/internal/receiver/r1", port)
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close() // nolint:errcheck
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("heartbeat test handler", func(t *testing.T) {
		url := fmt.Sprintf("http://127.0.0.1:%d/internal/heartbeat/test-hb", port)
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close() // nolint:errcheck
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
