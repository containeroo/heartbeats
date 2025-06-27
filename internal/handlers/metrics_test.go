package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containeroo/heartbeats/internal/history"
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
	handler := Metrics(store)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected HTTP 200 OK")
	assert.Equal(t, "# HELP heartbeats_history_byte_size Current size of the history store in bytes\n# TYPE heartbeats_history_byte_size gauge\nheartbeats_history_byte_size 8720\n", rec.Body.String(), "Expected metrics to include 'heartbeats_history_byte_size'")
}
