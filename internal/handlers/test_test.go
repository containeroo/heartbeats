package handlers

import (
	"context"
	"encoding/json"
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
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
	"github.com/stretchr/testify/assert"
)

func TestTestReceiverHandler(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	recorder := servicehistory.NewRecorder(hist)
	disp := notifier.NewDispatcher(store, logger, recorder, 1, 1, 10)
	api := NewAPI("test", "test", nil, logger, nil, hist, recorder, disp)

	handler := api.TestReceiverHandler()

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/internal/receiver/", nil)
		rec := httptest.NewRecorder()

		// simulate missing path value
		req.SetPathValue("id", "")
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp errorResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "missing id", resp.Error)
	})

	t.Run("sends test notification", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/internal/receiver/test-rec", nil)
		rec := httptest.NewRecorder()

		req.SetPathValue("id", "test-rec")
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp statusResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "ok", resp.Status)
	})
}

func TestTestHeartbeatHandler(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	recorder := servicehistory.NewRecorder(hist)
	disp := notifier.NewDispatcher(store, logger, recorder, 1, 1, 10)

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
	mgr := heartbeat.NewManagerFromHeartbeatMap(context.Background(), cfg, disp.Mailbox(), recorder, logger)
	api := NewAPI(logger, "test", "test", nil, mgr, hist, recorder, disp)
	handler := api.TestHeartbeatHandler()

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/internal/heartbeat/", nil)
		rec := httptest.NewRecorder()

		req.SetPathValue("id", "")
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp errorResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "missing id", resp.Error)
	})

	t.Run("unknown id", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/internal/heartbeat/invalid", nil)
		rec := httptest.NewRecorder()

		req.SetPathValue("id", "invalid")
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		var resp errorResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "heartbeat ID \"invalid\" not found", resp.Error)
	})

	t.Run("trigger test heartbeat", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", fmt.Sprintf("/internal/heartbeat/%s", hbName), nil)
		rec := httptest.NewRecorder()

		req.SetPathValue("id", hbName)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp statusResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "ok", resp.Status)
	})
}
