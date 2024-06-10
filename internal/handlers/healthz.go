package handlers

import (
	"heartbeats/internal/logger"
	"net/http"
)

func Healthz(logger logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Health check endpoint called")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok")) // Make linter happy
	})
}
