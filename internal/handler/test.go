package handler

import (
	"errors"
	"net/http"

	testservice "github.com/containeroo/heartbeats/internal/service/test"
)

// TestReceiverHandler allows sending a test notification to a specific receiver
func (a *API) TestReceiverHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			a.logRequestError(r, "missing_id", "missing id", errors.New("missing id"))
			a.respondJSON(w, http.StatusBadRequest, errorResponse{Error: "missing id"})
			return
		}

		testservice.SendTestNotification(a.disp, a.Logger, id)

		a.respondJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	})
}

// TestHeartbeatHandler allows sending a test notification to a specific heartbeat
func (a *API) TestHeartbeatHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			a.logRequestError(r, "missing_id", "missing id", errors.New("missing id"))
			a.respondJSON(w, http.StatusBadRequest, errorResponse{Error: "missing id"})
			return
		}

		if err := testservice.TriggerTestHeartbeat(a.mgr, a.Logger, id); err != nil {
			a.logRequestError(r, "handle_test_failed", "handle test failed", err)
			a.respondJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
			return
		}

		a.respondJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	})
}
