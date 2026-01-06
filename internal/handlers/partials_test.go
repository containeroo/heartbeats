package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/stretchr/testify/assert"
)

func TestPartialHandler(t *testing.T) {
	t.Parallel()

	webFS := fstest.MapFS{
		"web/templates/heartbeats.html": &fstest.MapFile{Data: []byte(`{{define "heartbeats"}}HEARTBEATS{{end}}`)},
		"web/templates/receivers.html":  &fstest.MapFile{Data: []byte(`{{define "receivers"}}RECEIVERS{{end}}`)},
		"web/templates/history.html":    &fstest.MapFile{Data: []byte(`{{define "history"}}HISTORY{{end}}`)},
	}

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	_ = hist.Append(context.Background(), history.MustNewEvent(
		history.EventTypeHeartbeatReceived,
		"hb1",
		history.RequestMetadataPayload{
			Method:    "GET",
			Source:    "127.0.0.1",
			UserAgent: "Go-http-client",
		},
	))
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)

	mgr := heartbeat.NewManagerFromHeartbeatMap(context.Background(), map[string]heartbeat.HeartbeatConfig{
		"hb1": {
			Description: "desc",
			Interval:    10 * time.Second,
			Grace:       5 * time.Second,
			Receivers:   []string{"r1"},
		},
	}, disp.Mailbox(), hist, logger)

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/partials/invalid", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(webFS, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "404 page not found\n", rr.Body.String())
	})

	t.Run("heartbeats", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/partials/heartbeats", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(webFS, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "HEARTBEATS")
	})

	t.Run("Receivers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/partials/receivers", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(webFS, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "RECEIVERS")
	})

	t.Run("History", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/partials/history", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(webFS, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "HISTORY")
	})
}
