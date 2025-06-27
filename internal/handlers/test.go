package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/notifier"
)

// TestReceiverHandler allows sending a test notification to a specific receiver
func TestReceiverHandler(dispatcher *notifier.Dispatcher, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		logger.Info("Test request received", "receiver", id)

		dispatcher.Mailbox() <- notifier.NotificationData{
			ID:        fmt.Sprintf("manual-test-%s", time.Now().Format(time.RFC3339)),
			Receivers: []string{id},
			Title:     "Test Notification",
			Message:   "This is a test notification",
		}

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

		logger.Info("Test request heartbeat", "heartbeat", id)

		if err := mgr.Test(id); err != nil {
			logger.Error("handle test failed", "id", id, "err", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok") // nolint:errcheck
	})
}
