package handlers

import (
	"heartbeats/pkg/logger"
	"heartbeats/pkg/metrics"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics is the handler for the /metrics route
// It returns the metrics for Prometheus to scrape
func Metrics(logger logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		promhttp.HandlerFor(metrics.PromMetrics.Registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})
}
