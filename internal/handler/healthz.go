package handler

import (
	"net/http"
)

// Healthz returns the HTTP handler for the /healthz endpoint.
func (a *API) Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.respondJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	}
}
