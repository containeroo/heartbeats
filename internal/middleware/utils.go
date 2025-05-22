package middleware

import "net/http"

// responseRecorder is a wrapper around http.ResponseWriter that captures the status code.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// ResponseWriter sends an HTTP response header with the provided status code.
func (rw *responseRecorder) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
