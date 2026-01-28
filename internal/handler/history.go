package handler

import "net/http"

// HistoryAll returns all recorded history items.
func (a *API) HistoryAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.respondJSON(w, http.StatusOK, a.history.List())
	}
}

// HistoryByHeartbeat returns recorded history items filtered by heartbeat id.
func (a *API) HistoryByHeartbeat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		heartbeatID := r.PathValue("id")
		if heartbeatID == "" {
			a.respondJSON(w, http.StatusBadRequest, errorResponse{Error: "missing heartbeat id"})
			return
		}

		a.respondJSON(w, http.StatusOK, a.history.ListByID(heartbeatID))
	}
}
