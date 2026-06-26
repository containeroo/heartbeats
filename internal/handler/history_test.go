package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/history"
	"github.com/stretchr/testify/assert"
)

func TestHistoryList(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewStore(10)
		api := NewAPI("test", "test", "http://example.com", logger)
		api.SetHistory(hist)

		hist.Add(history.Event{
			Time:        time.Now().UTC(),
			Type:        history.EventHeartbeatReceived.String(),
			HeartbeatID: "foo",
		})
		hist.Add(history.Event{
			Time:        time.Now().UTC(),
			Type:        history.EventNotificationDelivered.String(),
			HeartbeatID: "foo",
		})

		handler := api.HistoryAll()
		req := httptest.NewRequest("GET", "/history", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
		var resp []history.Event
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Len(t, resp, 2, "Expected two history entries")
	})
}

func TestHistoryListByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		hist := history.NewStore(10)
		api := NewAPI(
			"test",
			"test",
			"http://example.com",
			logger,
		)
		api.SetHistory(hist)

		hist.Add(history.Event{
			Time:        time.Now().UTC(),
			Type:        history.EventHeartbeatReceived.String(),
			HeartbeatID: "foo",
		})
		hist.Add(history.Event{
			Time:        time.Now().UTC(),
			Type:        history.EventHeartbeatReceived.String(),
			HeartbeatID: "bar",
		})

		mux := http.NewServeMux()
		mux.HandleFunc("/history/{id}", api.HistoryByHeartbeat())
		req := httptest.NewRequest("GET", "/history/foo", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
		var resp []history.Event
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Len(t, resp, 1, "Expected one history entry")
		assert.Equal(t, "foo", resp[0].HeartbeatID, "Expected heartbeat id 'foo'")
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		api := NewAPI(
			"test",
			"test",
			"http://example.com",
			logger,
		)
		api.SetHistory(history.NewStore(1))

		mux := http.NewServeMux()
		mux.HandleFunc("/history/{id}", api.HistoryByHeartbeat())
		req := httptest.NewRequest("GET", "/history/not-found", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
		var resp []history.Event
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Len(t, resp, 0, "Expected empty history response")
	})
}
