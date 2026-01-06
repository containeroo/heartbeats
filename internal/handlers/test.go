package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/notifier"
	testservice "github.com/containeroo/heartbeats/internal/service/test"
)

// TestReceiverHandler allows sending a test notification to a specific receiver
func TestReceiverHandler(dispatcher *notifier.Dispatcher, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		testservice.SendTestNotification(dispatcher, logger, id)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok") // nolint:errcheck
	})
}

// TestHeartbeatHandler allows sending a test notification to a specific heartbeat
func TestHeartbeatHandler(mgr *heartbeat.Manager, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		if err := testservice.TriggerTestHeartbeat(mgr, logger, id); err != nil {
			logger.Error("handle test failed", "id", id, "err", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok") // nolint:errcheck
	})
}
