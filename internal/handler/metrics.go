package handler

import "net/http"

// Metrics exposes Prometheus metrics.
func (a *API) Metrics() http.Handler {
	return a.metrics.Metrics()
}
