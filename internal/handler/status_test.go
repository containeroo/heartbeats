package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat/service"
	"github.com/stretchr/testify/assert"
)

type fakeStatusService struct {
	statuses []service.Status
}

func (f fakeStatusService) HeartbeatSummaries() []service.HeartbeatSummary {
	return nil
}

func (f fakeStatusService) ReceiverSummaries() []service.ReceiverSummary {
	return nil
}

func (f fakeStatusService) Update(id string, payload string, now time.Time) error {
	return nil
}

func (f fakeStatusService) StatusAll() []service.Status {
	return f.statuses
}

func (f fakeStatusService) StatusByID(id string) (service.Status, error) {
	for _, status := range f.statuses {
		if status.ID == id {
			return status, nil
		}
	}
	return service.Status{}, fmt.Errorf("heartbeat %q not found", id)
}

func TestStatusAll(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		api := NewAPI("test", "test", "http://example.com", logger)
		api.SetService(fakeStatusService{statuses: []service.Status{{ID: "foo"}}})

		handler := api.StatusAll()
		req := httptest.NewRequest("GET", "/status", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
		var resp []service.Status
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Len(t, resp, 1, "Expected one status entry")
	})
}

func TestStatusByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		api := NewAPI(
			"test",
			"test",
			"http://example.com",
			logger,
		)
		api.SetService(fakeStatusService{statuses: []service.Status{{ID: "foo"}}})

		mux := http.NewServeMux()
		mux.HandleFunc("/status/{id}", api.Status())
		req := httptest.NewRequest("GET", "/status/foo", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
		var resp service.Status
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "foo", resp.ID, "Expected heartbeat id 'foo'")
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
		api.SetService(fakeStatusService{statuses: []service.Status{{ID: "foo"}}})

		mux := http.NewServeMux()
		mux.HandleFunc("/status/{id}", api.Status())
		req := httptest.NewRequest("GET", "/status/not-found", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code, "Expected status code 404")
		var resp errorResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.NotEmpty(t, resp.Error, "Expected error message")
	})
}
