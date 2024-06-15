package handlers

import (
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
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

func setupAferoFSForHistory() afero.Fs {
	aferoFS := afero.NewMemMapFs()

	templateFiles := []string{
		"history.html",
		"footer.html",
	}

	for _, file := range templateFiles {
		content, err := os.ReadFile(filepath.Join("../../web/templates", file))
		if err != nil {
			panic(err)
		}

		err = afero.WriteFile(aferoFS, "web/templates/"+file, content, 0644)
		if err != nil {
			panic(err)
		}
	}

	return aferoFS
}

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

func TestHistoryHandler(t *testing.T) {
	log := logger.NewLogger(true)
	heartbeatStore := heartbeat.NewStore()
	historyStore := history.NewStore()

	h := &heartbeat.Heartbeat{
		Name:     "test",
		Enabled:  new(bool),
		Interval: &timer.Timer{Interval: new(time.Duration)},
		Grace:    &timer.Timer{Interval: new(time.Duration)},
	}
	*h.Enabled = true
	*h.Interval.Interval = time.Minute
	*h.Grace.Interval = time.Minute

	err := heartbeatStore.Add("test", h)
	assert.NoError(t, err)

	hist, err := history.NewHistory(10, 2)
	assert.NoError(t, err)

	err = historyStore.Add("test", hist)
	assert.NoError(t, err)

	version := "1.0.0"

	mux := http.NewServeMux()
	aferoFS := setupAferoFSForHistory()
	mux.Handle("GET /history/{id}", History(log, aferoToCustomAferoFS(aferoFS), version, heartbeatStore, historyStore))

	t.Run("Heartbeat not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/history/nonexistent", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code, "Expected status code 404")
		assert.Contains(t, rec.Body.String(), "Heartbeat 'nonexistent' not found", "Expected heartbeat not found message")
	})

	t.Run("Heartbeat found and history retrieved", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/history/test", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
		assert.Contains(t, rec.Body.String(), "History for test", "Expected history content")
	})

	// Simulate a template parsing error by using an invalid template path
	t.Run("Template parsing error", func(t *testing.T) {
		invalidFS := afero.NewMemMapFs() // Empty FS to simulate missing templates
		mux := http.NewServeMux()
		mux.Handle("GET /history/{id}", History(log, aferoToCustomAferoFS(invalidFS), version, heartbeatStore, historyStore))

		req := httptest.NewRequest("GET", "/history/test", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code, "Expected status code 500")
		assert.Contains(t, rec.Body.String(), "Internal Server Error", "Expected internal server error message")
	})
}
