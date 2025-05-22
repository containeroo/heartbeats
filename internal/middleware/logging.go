package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggingMiddleware wraps the handler, logs the request and response, and writes the response to the client.
func LoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now() // Start measuring time before calling next.ServeHTTP

			// Capture response status
			rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rec, r)

			duration := time.Since(startTime) // Measure total execution time

			// Log request details
			logger.Debug("HTTP request",
				"method", r.Method,
				"url_path", r.URL.Path,
				"status_code", rec.statusCode,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"duration_ms", duration.String(),
			)
		})
	}
}
