package handlers

import (
	"net/http"

	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics returns the HTTP handler for the /metrics endpoint.
func Metrics(hist history.Store) http.HandlerFunc {
	metrics := metrics.NewMetrics(hist)
	h := promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{})
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}
