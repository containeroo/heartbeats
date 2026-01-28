package routes

import (
	"bufio"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

// LoggingMiddleware wraps a handler with structured access logging.
func LoggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now() // Measure end-to-end latency.

		// Capture response status codes from downstream handlers.
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(startTime) // Measure total execution time

		logger.Debug("HTTP request",
			"method", r.Method,
			"url_path", r.URL.Path,
			"status_code", rec.statusCode,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"duration_ms", duration.String(),
			"heartbeat_id", extractHeartbeatID(r.URL.Path),
		)
	})
}

// responseRecorder wraps http.ResponseWriter to capture status codes.
type responseRecorder struct {
	http.ResponseWriter     // Wrapped writer.
	statusCode          int // Last status code written.
	bytesWritten        int // Total bytes written.
}

// WriteHeader records the status code and forwards it to the wrapped writer.
func (rw *responseRecorder) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write records the response size and forwards the write.
func (rw *responseRecorder) Write(p []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(p)
	rw.bytesWritten += n
	return n, err
}

// Flush passes through flush calls when supported.
func (rw *responseRecorder) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements http.Hijacker to support protocol upgrades.
func (rw *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response writer does not support hijacking")
	}
	return hijacker.Hijack()
}

// Unwrap exposes the underlying ResponseWriter for helpers like ResponseController.
func (rw *responseRecorder) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func extractHeartbeatID(path string) string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == "heartbeat" && parts[i+1] != "" {
			return parts[i+1]
		}
	}
	return ""
}
