package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	handler := Healthz()

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
	assert.Equal(t, "ok", rec.Body.String(), "Expected response body 'ok'")
}
