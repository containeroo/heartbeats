package handlers

import (
	"heartbeats/pkg/config"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/notify"
	"heartbeats/pkg/timer"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPingHandler(t *testing.T) {
	log := logger.NewLogger(true)
	config.App.HeartbeatStore = heartbeat.NewStore()
	config.App.NotificationStore = notify.NewStore()
	config.HistoryStore = history.NewStore()

	h := &heartbeat.Heartbeat{
		Name:     "test",
		Enabled:  new(bool),
		Interval: &timer.Timer{Interval: new(time.Duration)},
		Grace:    &timer.Timer{Interval: new(time.Duration)},
	}
	*h.Enabled = true
	*h.Interval.Interval = time.Minute
	*h.Grace.Interval = time.Minute

	err := config.App.HeartbeatStore.Add("test", h)
	assert.NoError(t, err)

	hist, err := history.NewHistory(10, 2)
	assert.NoError(t, err)

	err = config.HistoryStore.Add("test", hist)
	assert.NoError(t, err)

	ns := notify.NewStore()
	config.App.NotificationStore = ns

	mux := http.NewServeMux()
	mux.Handle("GET /ping/{id}", Ping(log))

	t.Run("Heartbeat found and enabled", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping/test", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
		assert.Equal(t, "ok", rec.Body.String(), "Expected response body 'ok'")
	})

	t.Run("Heartbeat not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping/nonexistent", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code, "Expected status code 404")
		assert.Contains(t, rec.Body.String(), "Heartbeat 'nonexistent' not found", "Expected heartbeat not found message")
	})

	t.Run("Heartbeat found but not enabled", func(t *testing.T) {
		*h.Enabled = false
		req := httptest.NewRequest("GET", "/ping/test", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code, "Expected status code 503")
		assert.Contains(t, rec.Body.String(), "Heartbeat 'test' not enabled", "Expected heartbeat not enabled message")
	})
}
