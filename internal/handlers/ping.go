package handlers

import (
	"context"
	"fmt"
	"heartbeats/internal/config"
	"heartbeats/internal/heartbeat"
	"heartbeats/internal/history"
	"heartbeats/internal/logger"
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
				logger.Warnf("%s Heartbeat «%s» not found", heartbeat.EventBeat, heartbeatName)
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(fmt.Sprintf("heartbeat '%s' not found", heartbeatName))) // Make linter happy
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
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("nok")) // Make linter happy
				return
			}

			ns := config.App.NotificationStore
			h.StartInterval(ctx, logger, ns, hs)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok")) // Make linter happy
		})
}
