package routes

import (
	"net/http"
	"time"

	"github.com/containeroo/heartbeats/internal/history"
)

// HistoryMiddleware records HTTP access events into history.
func HistoryMiddleware(recorder history.Recorder, next http.Handler) http.HandlerFunc {
	if recorder == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			if next != nil {
				next.ServeHTTP(w, r)
			}
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		if next != nil {
			next.ServeHTTP(rec, r)
		}

		event := history.Event{
			Time:    time.Now().UTC(),
			Type:    history.EventHTTPAccess.String(),
			Message: "http_access",
			Fields: map[string]any{
				"method":      r.Method,
				"url_path":    r.URL.Path,
				"status_code": rec.statusCode,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
				"duration":    time.Since(startTime).String(),
				"bytes":       rec.bytesWritten,
			},
		}
		if heartbeatID := r.PathValue("id"); heartbeatID != "" {
			event.HeartbeatID = heartbeatID
		}
		recorder.Add(event)
	})
}
