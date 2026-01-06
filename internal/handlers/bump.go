package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/service/bump"
)

// BumpHandler handles POST/GET /heartbeat/{id}.
func BumpHandler(mgr *heartbeat.Manager, hist history.Store, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		if err := bump.Receive(r.Context(), mgr, hist, logger, id, r.RemoteAddr, r.Method, r.UserAgent()); err != nil {
			if errors.Is(err, bump.ErrUnknownHeartbeat) {
				http.Error(w, fmt.Sprintf("unknown heartbeat id %q", id), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok") // nolint:errcheck
	})
}

// FailHandler handles POST/GET /heartbeat/{id}/fail.
func FailHandler(mgr *heartbeat.Manager, hist history.Store, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		if err := bump.Fail(r.Context(), mgr, hist, logger, id, r.RemoteAddr, r.Method, r.UserAgent()); err != nil {
			if errors.Is(err, bump.ErrUnknownHeartbeat) {
				http.Error(w, fmt.Sprintf("unknown heartbeat id %q", id), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok") // nolint:errcheck
	})
}
