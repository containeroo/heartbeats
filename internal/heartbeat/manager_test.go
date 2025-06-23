package heartbeat_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

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

		mgr := heartbeat.NewManager(ctx, cfg, disp.Mailbox(), hist, logger)
		err := mgr.HandleReceive("a1")
		assert.NoError(t, err)
	})

	t.Run("actor not found", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]heartbeat.HeartbeatConfig{}

		mgr := heartbeat.NewManager(ctx, cfg, disp.Mailbox(), hist, logger)
		err := mgr.HandleReceive("a1")
		assert.Error(t, err)
		assert.EqualError(t, err, "unknown heartbeat id \"a1\"")
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

		mgr := heartbeat.NewManager(ctx, cfg, disp.Mailbox(), hist, logger)
		err := mgr.HandleFail("a1")
		assert.NoError(t, err)
	})

	t.Run("actor not found", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]heartbeat.HeartbeatConfig{}

		mgr := heartbeat.NewManager(ctx, cfg, disp.Mailbox(), hist, logger)
		err := mgr.HandleFail("a1")
		assert.Error(t, err)
		assert.EqualError(t, err, "unknown heartbeat id \"a1\"")
	})
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

	mgr := heartbeat.NewManager(ctx, cfg, disp.Mailbox(), hist, logger)

	t.Run("List", func(t *testing.T) {
		t.Parallel()

		result := mgr.List()
		assert.Len(t, result, 2)
	})

	t.Run("Get found", func(t *testing.T) {
		t.Parallel()

		result := mgr.Get("a2")
		assert.NotNil(t, result)
		assert.Equal(t, "test-2", result.Description)
	})

	t.Run("Get found", func(t *testing.T) {
		t.Parallel()

		result := mgr.Get("a0")
		assert.Nil(t, result)
	})
}
