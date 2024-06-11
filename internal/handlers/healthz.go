package handlers

import (
	"heartbeats/internal/logger"
	"net/http"
)

// Healthz handles the /healthz endpoint
func Healthz(logger logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			logger.Errorf("Failed to write health check response. %v", err)
		}
	}
}
