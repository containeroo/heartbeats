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
	"github.com/stretchr/testify/assert"
)

func TestManager_HandleReceive(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(20)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)

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

		mgr := heartbeat.NewManagerFromHeartbeatMap(ctx, cfg, disp.Mailbox(), hist, logger)
		err := mgr.Receive("a1")
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond) // allow actor to process mailbox
		a := mgr.Get("a1")
		assert.Equal(t, common.HeartbeatStateActive, a.State)
	})

	t.Run("actor not found", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]heartbeat.HeartbeatConfig{}

		mgr := heartbeat.NewManagerFromHeartbeatMap(ctx, cfg, disp.Mailbox(), hist, logger)
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
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)

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

		mgr := heartbeat.NewManagerFromHeartbeatMap(ctx, cfg, disp.Mailbox(), hist, logger)
		err := mgr.Fail("a1")
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond) // allow actor to process mailbox
		a := mgr.Get("a1")
		assert.Equal(t, common.HeartbeatStateFailed, a.State)
	})

	t.Run("actor not found", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]heartbeat.HeartbeatConfig{}

		mgr := heartbeat.NewManagerFromHeartbeatMap(ctx, cfg, disp.Mailbox(), hist, logger)
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
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)

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

		mgr := heartbeat.NewManagerFromHeartbeatMap(ctx, cfg, disp.Mailbox(), hist, logger)
		err := mgr.Test("a1")
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond) // allow actor to process mailbox
		a := mgr.Get("a1")
		assert.Equal(t, common.HeartbeatStateIdle, a.State)
	})

	t.Run("returns error if actor not found", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]heartbeat.HeartbeatConfig{}
		mgr := heartbeat.NewManagerFromHeartbeatMap(ctx, cfg, disp.Mailbox(), hist, logger)

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
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 0, 0, 10)

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

	mgr := heartbeat.NewManagerFromHeartbeatMap(ctx, cfg, disp.Mailbox(), hist, logger)

	result := mgr.List()
	assert.Len(t, result, 2)
}

func TestManager_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(20)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 0, 0, 10)

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

	mgr := heartbeat.NewManagerFromHeartbeatMap(ctx, cfg, disp.Mailbox(), hist, logger)

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
