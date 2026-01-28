package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/heartbeat/manager"
	"github.com/containeroo/heartbeats/internal/heartbeat/service"
	"github.com/stretchr/testify/assert"
)

func TestStatusAll(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Receivers: map[string]config.ReceiverConfig{
			"ops": {
				Webhooks: []config.WebhookConfig{
					{
						URL:      "https://example.com",
						Template: "default",
					},
				},
			},
		},
		Heartbeats: map[string]config.HeartbeatConfig{
			"foo": {
				Interval:  time.Second,
				LateAfter: time.Second,
				Receivers: []string{"ops"},
			},
		},
	}

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		mgr, err := manager.NewManager(&cfg, os.DirFS("../.."), nil, nil, nil, logger)
		assert.NoError(t, err)
		svc := service.NewService(mgr, nil, nil)
		api := NewAPI("test", "test", "http://example.com", logger)
		api.SetService(svc)

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

	cfg := config.Config{
		Receivers: map[string]config.ReceiverConfig{
			"ops": {
				Webhooks: []config.WebhookConfig{
					{
						URL:      "https://example.com",
						Template: "default",
					},
				},
			},
		},
		Heartbeats: map[string]config.HeartbeatConfig{
			"foo": {
				Interval:  time.Second,
				LateAfter: time.Second,
				Receivers: []string{"ops"},
			},
		},
	}

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		mgr, err := manager.NewManager(&cfg, os.DirFS("../.."), nil, nil, nil, logger)
		assert.NoError(t, err)
		svc := service.NewService(mgr, nil, nil)
		api := NewAPI(
			"test",
			"test",
			"http://example.com",
			logger,
		)
		api.SetService(svc)

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
		mgr, err := manager.NewManager(&cfg, os.DirFS("../.."), nil, nil, nil, logger)
		assert.NoError(t, err)
		svc := service.NewService(mgr, nil, nil)
		api := NewAPI(
			"test",
			"test",
			"http://example.com",
			logger,
		)
		api.SetService(svc)

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
