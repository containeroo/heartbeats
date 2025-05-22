package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

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

		now := time.Now()
		src := r.RemoteAddr
		ua := r.Header.Get("User-Agent")
		logger.Info("received heartbeat", "id", id, "from", src)

		_ = hist.RecordEvent(r.Context(), history.Event{
			Timestamp:   now,
			Type:        history.EventTypeHeartbeatReceived,
			HeartbeatID: id,
			Source:      src,
			Method:      r.Method,
			UserAgent:   ua,
		})

		if err := mgr.HandleReceive(id); err != nil {
			logger.Error("handle receive failed", "id", id, "err", err)
			http.Error(w, err.Error(), http.StatusNotFound)
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

		now := time.Now()
		src := r.RemoteAddr
		ua := r.Header.Get("User-Agent")
		logger.Info("manual fail", "id", id, "from", src)

		_ = hist.RecordEvent(r.Context(), history.Event{
			Timestamp:   now,
			Type:        history.EventTypeHeartbeatFailed,
			HeartbeatID: id,
			Source:      src,
			Method:      r.Method,
			UserAgent:   ua,
		})

		if err := mgr.HandleFail(id); err != nil {
			logger.Error("handle receive failed", "id", id, "err", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok") // nolint:errcheck
	})
}
