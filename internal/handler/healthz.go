package handler

import (
	"net/http"

	"github.com/containeroo/heartbeats/internal/service/health"
)

// Healthz returns the HTTP handler for the /healthz endpoint.
func (a *API) Healthz(service *health.Service) http.HandlerFunc {
	status := service.Status()
	return func(w http.ResponseWriter, r *http.Request) {
		a.respondJSON(w, http.StatusOK, statusResponse{Status: status.Status})
	}
}
