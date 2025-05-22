package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containeroo/heartbeats/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler(t *testing.T) {
	// Register a test metric
	testMetric := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_metric",
		Help: "This is a test metric",
	})
	metrics.PromMetrics.Registry.MustRegister(testMetric)

	// Increment the test metric
	testMetric.Inc()

	// Create the Metrics handler
	handler := Metrics()

	// Create a new HTTP request
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	// Serve the HTTP request
	handler.ServeHTTP(rec, req)

	// Check the status code and response body
	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
	assert.Contains(t, rec.Body.String(), "test_metric", "Expected response body to contain 'test_metric'")
}
