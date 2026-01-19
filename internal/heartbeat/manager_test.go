package heartbeat_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
	"github.com/stretchr/testify/assert"
)

func newTestManager(
	t *testing.T,
	ctx context.Context,
	cfg heartbeat.HeartbeatConfigMap,
	disp *notifier.Dispatcher,
	recorder *servicehistory.Recorder,
	logger *slog.Logger,
) *heartbeat.Manager {
	t.Helper()

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
		cfg,
		heartbeat.ManagerConfig{Logger: logger, Factory: factory},
	)
	assert.NoError(t, err)
	return mgr
}

func TestManager_HandleReceive(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(20)
	recorder := servicehistory.NewRecorder(hist)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, recorder, 1, 1, 10, nil)

	t.Run("sends receive to known actor", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]heartbeat.HeartbeatConfig{
			"a1": {
				Description: "test",
				Interval:    50 * time.Millisecond,
				Grace:       50 * time.Millisecond,
				Receivers:   []string{"r1"},
			},
		}

		mgr := newTestManager(t, ctx, cfg, disp, recorder, logger)
		mgr.StartAll()
		err := mgr.Receive("a1")
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond) // allow actor to process mailbox
		a := mgr.Get("a1")
		assert.Equal(t, common.HeartbeatStateActive, a.State)
	})

	t.Run("actor not found", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]heartbeat.HeartbeatConfig{}

		mgr := newTestManager(t, ctx, cfg, disp, recorder, logger)
		err := mgr.Receive("a1")
		assert.Error(t, err)
		assert.EqualError(t, err, "heartbeat ID \"a1\" not found")
	})
}

func TestManager_HandleFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(20)
	recorder := servicehistory.NewRecorder(hist)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, recorder, 1, 1, 10, nil)

	t.Run("sends fail to known actor", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]heartbeat.HeartbeatConfig{
			"a1": {
				Description: "test",
				Interval:    50 * time.Millisecond,
				Grace:       50 * time.Millisecond,
				Receivers:   []string{"r1"},
			},
		}

		mgr := newTestManager(t, ctx, cfg, disp, recorder, logger)
		mgr.StartAll()
		err := mgr.Fail("a1")
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond) // allow actor to process mailbox
		a := mgr.Get("a1")
		assert.Equal(t, common.HeartbeatStateFailed, a.State)
	})

	t.Run("actor not found", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]heartbeat.HeartbeatConfig{}

		mgr := newTestManager(t, ctx, cfg, disp, recorder, logger)
		err := mgr.Fail("a1")
		assert.Error(t, err)
		assert.EqualError(t, err, "heartbeat ID \"a1\" not found")
	})
}

func TestManager_HandleTest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	recorder := servicehistory.NewRecorder(hist)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, recorder, 1, 1, 10, nil)

	t.Run("sends test event to known actor", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]heartbeat.HeartbeatConfig{
			"a1": {
				Description: "test",
				Interval:    50 * time.Millisecond,
				Grace:       50 * time.Millisecond,
				Receivers:   []string{"r1"},
			},
		}

		mgr := newTestManager(t, ctx, cfg, disp, recorder, logger)
		mgr.StartAll()
		err := mgr.Test("a1")
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond) // allow actor to process mailbox
		a := mgr.Get("a1")
		assert.Equal(t, common.HeartbeatStateIdle, a.State)
	})

	t.Run("returns error if actor not found", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]heartbeat.HeartbeatConfig{}
		mgr := newTestManager(t, ctx, cfg, disp, recorder, logger)

		err := mgr.Test("does-not-exist")
		assert.Error(t, err)
		assert.EqualError(t, err, "heartbeat ID \"does-not-exist\" not found")
	})
}

func TestManager_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(20)
	recorder := servicehistory.NewRecorder(hist)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, recorder, 0, 0, 10, nil)

	cfg := map[string]heartbeat.HeartbeatConfig{
		"a1": {
			Description: "test-1",
			Interval:    50 * time.Millisecond,
			Grace:       50 * time.Millisecond,
			Receivers:   []string{"r1"},
		},
		"a2": {
			Description: "test-2",
			Interval:    50 * time.Millisecond,
			Grace:       50 * time.Millisecond,
			Receivers:   []string{"r1"},
		},
	}

	mgr := newTestManager(t, ctx, cfg, disp, recorder, logger)

	result := mgr.List()
	assert.Len(t, result, 2)
}

func TestManager_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(20)
	recorder := servicehistory.NewRecorder(hist)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, recorder, 0, 0, 10, nil)

	cfg := map[string]heartbeat.HeartbeatConfig{
		"a1": {
			Description: "test-1",
			Interval:    50 * time.Millisecond,
			Grace:       50 * time.Millisecond,
			Receivers:   []string{"r1"},
		},
		"a2": {
			Description: "test-2",
			Interval:    50 * time.Millisecond,
			Grace:       50 * time.Millisecond,
			Receivers:   []string{"r1"},
		},
	}

	mgr := newTestManager(t, ctx, cfg, disp, recorder, logger)

	t.Run("Get found", func(t *testing.T) {
		t.Parallel()

		result := mgr.Get("a2")
		assert.NotNil(t, result)
		assert.Equal(t, "test-2", result.Description)
	})

	t.Run("Get not found", func(t *testing.T) {
		t.Parallel()

		result := mgr.Get("a0")
		assert.Nil(t, result)
	})
}
