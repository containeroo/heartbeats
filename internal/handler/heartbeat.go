package handler

import (
	"io"
	"net/http"
	"time"
)

// Heartbeat receives heartbeat pings for a specific heartbeat id.
func (a *API) Heartbeat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		heartbeatID := r.PathValue("id")
		if heartbeatID == "" {
			a.respondJSON(w, http.StatusBadRequest, errorResponse{Error: "missing heartbeat id"})
			return
		}

		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		_ = r.Body.Close()

		if err := a.service.Update(heartbeatID, string(body), now); err != nil {
			a.respondJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
			return
		}
		a.respondJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	}
}
