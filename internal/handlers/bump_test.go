package handlers

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/stretchr/testify/assert"
)

func setupRouter(hbName string, hist *history.RingStore) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	heartbeats := heartbeat.HeartbeatConfigMap{
		hbName: {
			ID:          hbName,
			Description: hbName,
			Interval:    30 * time.Second,
			Grace:       10 * time.Second,
			Receivers:   []string{"rec"},
		},
	}

	disp := notifier.NewDispatcher(notifier.InitializeStore(nil, false, logger), logger)
	mgr := heartbeat.NewManager(context.Background(), heartbeats, disp, hist, logger)
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

		router := setupRouter("invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/no-id/", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "missing id\n", rec.Body.String())
	})

	t.Run("not found id", func(t *testing.T) {
		t.Parallel()

		router := setupRouter("invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/bump/not-found", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "unknown heartbeat id \"not-found\"\n", rec.Body.String())
	})

	t.Run("successful bump GET", func(t *testing.T) {
		t.Parallel()

		hbName := "get-success"
		router := setupRouter(hbName, hist)

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
		router := setupRouter(hbName, hist)

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

		router := setupRouter("invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/no-id/fail", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "missing id\n", rec.Body.String())
	})

	t.Run("not found id", func(t *testing.T) {
		t.Parallel()

		router := setupRouter("invalid", hist)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/bump/not-found/fail", nil)
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "unknown heartbeat id \"not-found\"\n", rec.Body.String())
	})

	t.Run("successful fail GET", func(t *testing.T) {
		t.Parallel()

		hbName := "check-success-fail"
		router := setupRouter(hbName, hist)

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
		router := setupRouter(hbName, hist)

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
