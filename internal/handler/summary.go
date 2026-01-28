package handler

import "net/http"

// Heartbeats returns a list of configured heartbeats for the UI.
func (a *API) HeartbeatsSum() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.respondJSON(w, http.StatusOK, a.service.HeartbeatSummaries())
	}
}

// Receivers returns a list of receivers for the UI.
func (a *API) ReceiversSum() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.respondJSON(w, http.StatusOK, a.service.ReceiverSummaries())
	}
}
