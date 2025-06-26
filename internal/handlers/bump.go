package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
)

// BumpHandler handles POST/GET /heartbeat/{id}.
func BumpHandler(mgr *heartbeat.Manager, hist history.Store, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		src := r.RemoteAddr

		if mgr.Get(id) == nil {
			logger.Warn("received bump for unknown heartbeat ID", "id", id, "from", src)
			http.Error(w, fmt.Sprintf("unknown heartbeat id %q", id), http.StatusNotFound)
			return
		}

		logger.Info("received bump", "id", id, "from", src)

		payload := history.RequestMetadataPayload{
			Source:    src,
			Method:    r.Method,
			UserAgent: r.UserAgent(),
		}
		ev := history.MustNewEvent(history.EventTypeHeartbeatReceived, id, payload)

		if err := hist.RecordEvent(r.Context(), ev); err != nil {
			logger.Error("failed to record state change", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// We check if the heartbeat exists before calling HandleReceive
		mgr.HandleReceive(id) // nolint:errcheck

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

		src := r.RemoteAddr

		if mgr.Get(id) == nil {
			logger.Warn("received /fail bump for unknown heartbeat ID", "id", id, "from", src)
			http.Error(w, fmt.Sprintf("unknown heartbeat id %q", id), http.StatusNotFound)
			return
		}

		logger.Info("manual fail", "id", id, "from", src)

		payload := history.RequestMetadataPayload{
			Source:    src,
			Method:    r.Method,
			UserAgent: r.UserAgent(),
		}
		ev := history.MustNewEvent(history.EventTypeHeartbeatFailed, id, payload)

		if err := hist.RecordEvent(r.Context(), ev); err != nil {
			logger.Error("failed to record state change", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// We check if the heartbeat exists before calling HandleFail
		mgr.HandleFail(id) // nolint:errcheck

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok") // nolint:errcheck
	})
}
