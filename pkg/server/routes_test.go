package server

import (
	"heartbeats/pkg/config"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/notify"
	"heartbeats/pkg/timer"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// customAferoFS implements the fs.FS interface for afero.Fs
type customAferoFS struct {
	fs afero.Fs
}

// Open implements the fs.FS interface
func (a *customAferoFS) Open(name string) (fs.File, error) {
	return a.fs.Open(name)
}

// Convert the afero.Fs to customAferoFS
func aferoToCustomAferoFS(afs afero.Fs) fs.FS {
	return &customAferoFS{fs: afs}
}

func setupAferoFSForRoutes() afero.Fs {
	aferoFS := afero.NewMemMapFs()

	staticFiles := []string{
		"web/static/css/heartbeats.css",
		"web/templates/history.html",
		"web/templates/heartbeats.html",
		"web/templates/footer.html",
	}

	for _, file := range staticFiles {
		content, err := os.ReadFile(filepath.Join("../../", file))
		if err != nil {
			panic(err)
		}

		err = afero.WriteFile(aferoFS, file, content, 0644)
		if err != nil {
			panic(err)
		}
	}

	return aferoFS
}

func TestNewRouter(t *testing.T) {
	log := logger.NewLogger(true)
	config.App.HeartbeatStore = heartbeat.NewStore()
	config.App.NotificationStore = notify.NewStore()
	config.HistoryStore = history.NewStore()

	h := &heartbeat.Heartbeat{
		Name:     "test",
		Enabled:  new(bool),
		Interval: &timer.Timer{Interval: new(time.Duration)},
		Grace:    &timer.Timer{Interval: new(time.Duration)},
	}
	*h.Enabled = true
	*h.Interval.Interval = time.Minute
	*h.Grace.Interval = time.Minute

	err := config.App.HeartbeatStore.Add("test", h)
	assert.NoError(t, err)

	hist, err := history.NewHistory(10, 2)
	assert.NoError(t, err)

	err = config.HistoryStore.Add("test", hist)
	assert.NoError(t, err)

	aferoFS := setupAferoFSForRoutes()
	customFS := aferoToCustomAferoFS(aferoFS)

	mux := newRouter(log, customFS)

	t.Run("GET /", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Heartbeat")
	})

	t.Run("GET /ping/test", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping/test", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("POST /ping/test", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/ping/test", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("GET /history/test", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/history/test", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GET /healthz", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/healthz", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("POST /healthz", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/healthz", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("GET /metrics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GET /static/example.txt", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/heartbeats.css", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Heartbeat not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping/nonexistent", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "Heartbeat 'nonexistent' not found")
	})
}
