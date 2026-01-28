package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	api := NewAPI(
		"test",
		"test",
		"http://example.com",
		slog.New(slog.NewTextHandler(&strings.Builder{}, nil)),
	)
	handler := api.Healthz()
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
	var resp statusResponse
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "ok", resp.Status, "Expected response status 'ok'")
}
