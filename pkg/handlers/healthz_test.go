package handlers

import (
	"heartbeats/pkg/logger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthz(t *testing.T) {
	log := logger.NewLogger(true)
	handler := Healthz(log)

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
	assert.Equal(t, "ok", rec.Body.String(), "Expected response body 'ok'")
}
