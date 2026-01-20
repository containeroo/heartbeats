package middleware

import (
	"net/http"

	"github.com/containeroo/heartbeats/internal/logging"
)

// RequestIDMiddleware ensures each request has a request_id in context and response headers.
func RequestIDMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(logging.RequestIDHeader)
			if requestID == "" {
				requestID = logging.NewRequestID()
			}
			if requestID != "" {
				r = r.WithContext(logging.WithRequestID(r.Context(), requestID))
				w.Header().Set(logging.RequestIDHeader, requestID)
			}
			next.ServeHTTP(w, r)
		})
	}
}
