package debugserver

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/handlers"
	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
	"github.com/stretchr/testify/require"
)

func TestDebugServer_Run(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	// Setup basic in-memory state
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(5)
	recorder := servicehistory.NewRecorder(hist)
	store := notifier.NewReceiverStore()
	disp := notifier.NewDispatcher(store, logger, recorder, 1, 1*time.Millisecond, 10, nil)

	hbCfg := map[string]heartbeat.HeartbeatConfig{
		"test-hb": {
			Description: "demo",
			Interval:    time.Second,
			Grace:       time.Second,
			Receivers:   []string{"r1"},
		},
	}
	factory := heartbeat.DefaultActorFactory{
		Deps: heartbeat.ActorDeps{
			Logger:     logger,
			History:    recorder,
			Metrics:    nil,
			DispatchCh: disp.Mailbox(),
		},
	}
	mgr, err := heartbeat.NewManagerFromHeartbeatMap(
		ctx,
		hbCfg,
		heartbeat.ManagerConfig{Logger: logger, Factory: factory},
	)
	require.NoError(t, err)
	api := handlers.NewAPI("test", "test", nil, logger, mgr, hist, recorder, disp, nil)

	// Pick a random free port
	port := 8089
	Run(ctx, port, api)

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
