package handler

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler(t *testing.T) {
	t.Parallel()

	metricsReg := metrics.NewRegistry()
	metricsReg.IncHeartbeatReceived("api")
	metricsReg.SetReceiverStatus("ops", "webhook", "api", metrics.SUCCESS)
	api := NewAPI(
		"test",
		"test",
		"http://example.com",
		slog.New(slog.NewTextHandler(&strings.Builder{}, nil)),
	)
	api.SetMetrics(metricsReg)

	handler := api.Metrics()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected HTTP 200 OK")
	assert.Contains(t, rec.Body.String(), "heartbeats_heartbeat_received_total", "Expected metrics to include 'heartbeats_heartbeat_received_total'")
	assert.Contains(t, rec.Body.String(), "heartbeats_receiver_last_status", "Expected metrics to include 'heartbeats_receiver_last_status'")
}
