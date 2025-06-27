package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/stretchr/testify/assert"
)

func TestTestReceiverHandler(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)

	handler := TestReceiverHandler(disp, logger)

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/internal/receiver/", nil)
		rec := httptest.NewRecorder()

		// simulate missing path value
		req.SetPathValue("id", "")
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "missing id\n", rec.Body.String())
	})

	t.Run("sends test notification", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/internal/receiver/test-rec", nil)
		rec := httptest.NewRecorder()

		req.SetPathValue("id", "test-rec")
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})
}

func TestTestHeartbeatHandler(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)

	hbName := "test-hb"
	cfg := heartbeat.HeartbeatConfigMap{
		hbName: {
			ID:          hbName,
			Description: "test heartbeat",
			Interval:    time.Second,
			Grace:       time.Second,
			Receivers:   []string{"r1"},
		},
	}
	mgr := heartbeat.NewManagerFromHeartbeatMap(context.Background(), cfg, disp.Mailbox(), hist, logger)
	handler := TestHeartbeatHandler(mgr, logger)

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/internal/heartbeat/", nil)
		rec := httptest.NewRecorder()

		req.SetPathValue("id", "")
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "missing id\n", rec.Body.String())
	})

	t.Run("unknown id", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/internal/heartbeat/invalid", nil)
		rec := httptest.NewRecorder()

		req.SetPathValue("id", "invalid")
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "heartbeat ID \"invalid\" not found\n", rec.Body.String())
	})

	t.Run("trigger test heartbeat", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", fmt.Sprintf("/internal/heartbeat/%s", hbName), nil)
		rec := httptest.NewRecorder()

		req.SetPathValue("id", hbName)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})
}
