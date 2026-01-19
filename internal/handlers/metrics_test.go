package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler(t *testing.T) {
	t.Parallel()

	// Create a ring store and fill with some data
	store := history.NewRingStore(100)
	for range 10 {
		event := history.MustNewEvent(history.EventTypeHeartbeatReceived, "test_heartbeat", history.RequestMetadataPayload{})
		_ = store.Append(context.Background(), event)
	}

	// Create metrics handler with injected store
	metricsReg := metrics.New(store)
	api := NewAPI("test", "test", nil, slog.New(slog.NewTextHandler(&strings.Builder{}, nil)), nil, store, nil, nil, metricsReg)
	handler := api.Metrics()

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected HTTP 200 OK")
	assert.Contains(t, rec.Body.String(), "heartbeats_history_byte_size", "Expected metrics to include 'heartbeats_history_byte_size'")
}
