package handlers

import (
	"net/http"

	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics returns the HTTP handler for the /metrics endpoint.
func Metrics() http.HandlerFunc {
	h := promhttp.HandlerFor(metrics.PromMetrics.Registry, promhttp.HandlerOpts{})
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}
