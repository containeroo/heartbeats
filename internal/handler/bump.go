package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/containeroo/heartbeats/internal/service/bump"
)

// BumpHandler handles POST/GET /heartbeat/{id}.
func (a *API) BumpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			a.logRequestError(r, "missing_id", "missing id", errors.New("missing id"))
			a.respondJSON(w, http.StatusBadRequest, errorResponse{Error: "missing id"})
			return
		}

		if err := bump.Receive(r.Context(), a.mgr, a.rec, a.Logger, id, r.RemoteAddr, r.Method, r.UserAgent()); err != nil {
			if errors.Is(err, bump.ErrUnknownHeartbeat) {
				a.logRequestError(r, "unknown_heartbeat", "unknown heartbeat id", err)
				a.respondJSON(w, http.StatusNotFound, errorResponse{Error: fmt.Sprintf("unknown heartbeat id %q", id)})
				return
			}
			a.logRequestError(r, "failed_to_receive", "failed to receive heartbeat", err)
			a.respondJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}

		a.businessLogger(r).Info("received heartbeat", "id", id)
		a.respondJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	})
}

// FailHandler handles POST/GET /heartbeat/{id}/fail.
func (a *API) FailHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			a.logRequestError(r, "missing_id", "missing id", errors.New("missing id"))
			a.respondJSON(w, http.StatusBadRequest, errorResponse{Error: "missing id"})
			return
		}

		if err := bump.Fail(r.Context(), a.mgr, a.rec, a.Logger, id, r.RemoteAddr, r.Method, r.UserAgent()); err != nil {
			if errors.Is(err, bump.ErrUnknownHeartbeat) {
				a.logRequestError(r, "unknown_heartbeat", "unknown heartbeat id", err)
				a.respondJSON(w, http.StatusNotFound, errorResponse{Error: fmt.Sprintf("unknown heartbeat id %q", id)})
				return
			}
			a.logRequestError(r, "failed_to_fail", "failed to fail heartbeat", err)
			a.respondJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
			return
		}

		a.businessLogger(r).Info("failed heartbeat", "id", id)
		a.respondJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	})
}
