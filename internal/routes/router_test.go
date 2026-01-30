package routes

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/containeroo/heartbeats/internal/handler"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()

	// In-memory file system with a minimal SPA template.
	webFS := fstest.MapFS{
		"web/dist/index.html": &fstest.MapFile{Data: []byte(`<!doctype html><base href="{{ .BaseHref }}"><meta name="routePrefix" content="{{ .RoutePrefix }}">`)},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := &handler.API{
		Logger:  logger,
		Version: "v1.2.3",
		Commit:  "abc123",
	}
	api.SetMetrics(metrics.NewRegistry())

	router, err := NewRouter(webFS, api, "/heartbeats", logger)
	require.NoError(t, err)

	t.Run("GET /heartbeats/", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/heartbeats/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `base href="/heartbeats/"`)
		assert.Contains(t, rec.Body.String(), `content="/heartbeats"`)
	})

	t.Run("GET /heartbeats/api/config", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/heartbeats/api/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var payload map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &payload)
		require.NoError(t, err)
		assert.Equal(t, "v1.2.3", payload["version"])
		assert.Equal(t, "abc123", payload["commit"])
	})
}
