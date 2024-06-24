package handlers

import (
	"context"
	"fmt"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/notify"
	"net/http"
)

// Ping handles the /ping/{id} endpoint
func Ping(logger logger.Logger, heartbeatStore *heartbeat.Store, notificationStore *notify.Store, historyStore *history.Store) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Use context.Background() to ensure the timers are not tied to the HTTP request context.
			// This prevents timers from being cancelled when the request context is done or cancelled.
			ctx := context.Background()

			heartbeatName := r.PathValue("id")
			clientIP := getClientIP(r)
			logger.Debugf("%s /ping/%s %s %s", r.Method, heartbeatName, clientIP, r.UserAgent())

			h := heartbeatStore.Get(heartbeatName)
			if h == nil {
				errMsg := fmt.Sprintf("Heartbeat '%s' not found", heartbeatName)
				logger.Warn(errMsg)
				http.Error(w, errMsg, http.StatusNotFound)
				return
			}
			msg := "got ping"
			logger.Infof("%s %s", heartbeatName, msg)

			if h.Enabled != nil && !*h.Enabled {
				http.Error(w, fmt.Sprintf("Heartbeat '%s' not enabled", heartbeatName), http.StatusServiceUnavailable)
				return
			}

			hs := historyStore.Get(heartbeatName)

			h.StartInterval(ctx, logger, notificationStore, hs)

			details := map[string]string{
				"proto":     r.Proto,
				"clientIP":  clientIP,
				"method":    r.Method,
				"userAgent": r.UserAgent(),
			}
			hs.Add(history.Beat, msg, details)

			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("ok")); err != nil {
				logger.Errorf("Failed to write response. %v", err)
			}
		})
}
