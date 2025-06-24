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

func setupRouter(t *testing.T, hbName string, hist history.Store) http.Handler {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	heartbeats := heartbeat.HeartbeatConfigMap{
		hbName: {
			ID:          hbName,
			Description: hbName,
			Interval:    30 * time.Second,
			Grace:       10 * time.Second,
			Receivers:   []string{"rec"},
		},
	}

	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)
	mgr := heartbeat.NewManager(context.Background(), heartbeats, disp.Mailbox(), hist, logger)
	router := http.NewServeMux()
	router.Handle("GET /no-id/", BumpHandler(mgr, hist, logger))
	router.Handle("GET /no-id/fail", FailHandler(mgr, hist, logger))
	router.Handle("GET /bump/{id}", BumpHandler(mgr, hist, logger))
	router.Handle("POST /bump/{id}", BumpHandler(mgr, hist, logger))
	router.Handle("GET /bump/{id}/fail", FailHandler(mgr, hist, logger))
	router.Handle("POST /bump/{id}/fail", FailHandler(mgr, hist, logger))
	return router
}

func TestBumpHandler(t *testing.T) {
	t.Parallel()

	hist := history.NewRingStore(10)

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()

		router := setupRouter(t, "invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/no-id/", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "missing id\n", rec.Body.String())
	})

	t.Run("not found id", func(t *testing.T) {
		t.Parallel()

		router := setupRouter(t, "invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/bump/not-found", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "unknown heartbeat id \"not-found\"\n", rec.Body.String())
	})

	t.Run("successful bump GET", func(t *testing.T) {
		t.Parallel()

		hbName := "get-success"
		router := setupRouter(t, hbName, hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/bump/%s", hbName), nil)

		req.RemoteAddr = "1.2.3.4:5678"
		req.Header.Set("User-Agent", "Go-test")
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())

		e := hist.GetEventsByID(hbName)

		ev := e[0]
		assert.Equal(t, history.EventTypeHeartbeatReceived, ev.Type)
		assert.Equal(t, hbName, ev.HeartbeatID)
		assert.Equal(t, "GET", ev.Method)
		assert.Equal(t, "1.2.3.4:5678", ev.Source)
		assert.Equal(t, "Go-test", ev.UserAgent)
	})

	t.Run("successful bump POST", func(t *testing.T) {
		t.Parallel()

		hbName := "post-success"
		router := setupRouter(t, hbName, hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/bump/%s", hbName), nil)
		req.RemoteAddr = "5.6.7.8:1234"
		req.Header.Set("User-Agent", "Go-post")
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())

		e := hist.GetEventsByID(hbName)
		ev := e[0]

		assert.Equal(t, history.EventTypeHeartbeatReceived, ev.Type)
		assert.Equal(t, hbName, ev.HeartbeatID)
		assert.Equal(t, "POST", ev.Method)
	})
}

func TestFailHandler(t *testing.T) {
	t.Parallel()

	hist := history.NewRingStore(10)

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()

		router := setupRouter(t, "invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/no-id/fail", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "missing id\n", rec.Body.String())
	})

	t.Run("not found id", func(t *testing.T) {
		t.Parallel()

		router := setupRouter(t, "invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/bump/not-found/fail", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "unknown heartbeat id \"not-found\"\n", rec.Body.String())
	})

	t.Run("successful fail GET", func(t *testing.T) {
		t.Parallel()

		hbName := "check-success-fail"
		router := setupRouter(t, hbName, hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/bump/%s/fail", hbName), nil)
		req.RemoteAddr = "1.2.3.4:5678"
		req.Header.Set("User-Agent", "Go-test")
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())

		e := hist.GetEventsByID(hbName)

		ev := e[0]
		assert.Equal(t, history.EventTypeHeartbeatFailed, ev.Type)
		assert.Equal(t, hbName, ev.HeartbeatID)
		assert.Equal(t, "GET", ev.Method)
		assert.Equal(t, "1.2.3.4:5678", ev.Source)
		assert.Equal(t, "Go-test", ev.UserAgent)
	})

	t.Run("successful fail POST", func(t *testing.T) {
		t.Parallel()

		hbName := "post-success-fail"
		router := setupRouter(t, hbName, hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/bump/%s/fail", hbName), nil)
		req.RemoteAddr = "5.6.7.8:1234"
		req.Header.Set("User-Agent", "Go-post")
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())

		e := hist.GetEventsByID(hbName)
		ev := e[0]

		assert.Equal(t, history.EventTypeHeartbeatFailed, ev.Type)
		assert.Equal(t, hbName, ev.HeartbeatID)
		assert.Equal(t, "POST", ev.Method)
	})
}
