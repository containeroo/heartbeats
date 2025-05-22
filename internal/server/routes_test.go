package server

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/notifier"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// customAferoFS adapts afero.Fs to fs.FS
type customAferoFS struct{ fs afero.Fs }

func (a *customAferoFS) Open(name string) (fs.File, error) {
	return a.fs.Open(name)
}

func aferoToFS(afs afero.Fs) fs.FS {
	return &customAferoFS{fs: afs}
}

func setupTestFS(t *testing.T) fs.FS {
	t.Helper()

	mem := afero.NewMemMapFs()
	files := map[string]string{
		"web/static/example.txt":        "static file",
		"web/static/css/heartbeats.css": "body {}",
		"web/templates/base.html":       `{{define "base"}}<html>v:{{.Version}}</html>{{end}}`,
		"web/templates/navbar.html":     `{{define "navbar"}}navbar{{end}}`,
		"web/templates/footer.html":     `{{define "footer"}}footer{{end}}`,
		"web/templates/heartbeats.html": `heartbeat page`,
		"web/templates/receivers.html":  `receiver page`,
		"web/templates/history.html":    `history page`,
	}

	for name, content := range files {
		err := afero.WriteFile(mem, name, []byte(content), 0644)
		require.NoError(t, err)
	}
	return aferoToFS(mem)
}

func TestNewRouter(t *testing.T) {
	t.Parallel()

	fs := setupTestFS(t)
	version := "test-version"
	siteRoot := "/"

	ctx := context.Background()
	cfg := map[string]heartbeat.HeartbeatConfig{
		"a1": {
			Description: "test",
			Interval:    50 * time.Millisecond,
			Grace:       50 * time.Millisecond,
			Receivers:   []string{"r1"},
		},
	}
	disp := notifier.NewDispatcher(notifier.InitializeStore(nil, false, nil), nil)
	hist := history.NewRingStore(10)
	logger := logging.SetupLogger(logging.LogFormatText, false, io.Discard)

	mgr := heartbeat.NewManager(ctx, cfg, disp, hist, logger)

	router := NewRouter(fs, siteRoot, version, mgr, hist, disp, logger, true)

	t.Run("GET /", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), version)
	})

	t.Run("GET /static/css/heartbeats.css", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("GET", "/static/css/heartbeats.css", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "body")
	})

	t.Run("GET /healthz", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest("GET", "/healthz", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "ok", rec.Body.String())
	})

	t.Run("GET /api/v1/bump/test", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/api/v1/bump/a1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "ok", rec.Body.String())
	})
}
