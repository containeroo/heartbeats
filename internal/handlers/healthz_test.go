package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/containeroo/heartbeats/internal/service/health"
	"github.com/stretchr/testify/assert"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	api := NewAPI(
		"test",
		"test",
		nil,
		"",
		"",
		true,
		slog.New(slog.NewTextHandler(&strings.Builder{}, nil)),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	handler := api.Healthz(health.NewService())

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
	var resp statusResponse
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "ok", resp.Status, "Expected response status 'ok'")
}
