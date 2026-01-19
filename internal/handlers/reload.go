package handlers

import "net/http"

// ReloadHandler triggers a config reload.
func (a *API) ReloadHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.reload == nil {
			a.respondJSON(w, http.StatusNotImplemented, errorResponse{Error: "reload not configured"})
			return
		}
		if err := a.reload(); err != nil {
			a.respondJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}
		a.respondJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	})
}
