package handler

import "net/http"

// Status returns a status snapshot of the heartbeat receiver.
func (a *API) StatusAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.respondJSON(w, http.StatusOK, a.service.StatusAll())
	}
}

// Status returns a status snapshot of the heartbeat receiver.
func (a *API) Status() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		heartbeatID := r.PathValue("id")
		if heartbeatID == "" {
			a.respondJSON(w, http.StatusBadRequest, errorResponse{Error: "missing heartbeat id"})
			return
		}

		status, err := a.service.StatusByID(heartbeatID)
		if err != nil {
			a.respondJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
			return
		}

		a.respondJSON(w, http.StatusOK, status)
	}
}
