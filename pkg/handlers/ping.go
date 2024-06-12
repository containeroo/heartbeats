package handlers

import (
	"context"
	"fmt"
	"heartbeats/pkg/config"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"net/http"
)

// Ping handles the /ping/{id} endpoint
func Ping(logger logger.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Use context.Background() to ensure the timers are not tied to the HTTP request context.
			// This prevents timers from being cancelled when the request context is done or cancelled.
			ctx := context.Background()

			heartbeatName := r.PathValue("id")
			clientIP := getClientIP(r)
			logger.Debugf("%s /ping/%s %s %s", r.Method, heartbeatName, clientIP, r.UserAgent())

			h := config.App.HeartbeatStore.Get(heartbeatName)
			if h == nil {
				errMsg := fmt.Sprintf("Heartbeat '%s' not found", heartbeatName)
				logger.Warn(errMsg)
				http.Error(w, errMsg, http.StatusNotFound)
				return
			}

			details := map[string]string{
				"proto":     r.Proto,
				"clientIP":  clientIP,
				"method":    r.Method,
				"userAgent": r.UserAgent(),
			}

			msg := "got ping"
			logger.Infof("%s %s", heartbeatName, msg)

			hs := config.HistoryStore.Get(heartbeatName)
			hs.AddEntry(history.Beat, msg, details)

			if h.Enabled != nil && !*h.Enabled {
				http.Error(w, fmt.Sprintf("Heartbeat '%s' not enabled", heartbeatName), http.StatusServiceUnavailable)
				return
			}

			ns := config.App.NotificationStore
			h.StartInterval(ctx, logger, ns, hs)

			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("ok")); err != nil {
				logger.Errorf("Failed to write response. %v", err)
			}
		})
}
