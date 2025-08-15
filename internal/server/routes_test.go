package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"

	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()

	// in-memory file system with minimal template files
	webFS := fstest.MapFS{
		"web/static/css/heartbeats.css": &fstest.MapFile{Data: []byte(`body {}`)},
		"web/templates/base.html":       &fstest.MapFile{Data: []byte(`{{define "base"}}<html>{{template "navbar"}}<footer>{{.Version}}</footer>{{end}}`)},
		"web/templates/navbar.html":     &fstest.MapFile{Data: []byte(`{{define "navbar"}}<nav>nav</nav>{{end}}`)},
		"web/templates/heartbeats.html": &fstest.MapFile{Data: []byte(`heartbeat page`)},
		"web/templates/receivers.html":  &fstest.MapFile{Data: []byte(`receiver page`)},
		"web/templates/history.html":    &fstest.MapFile{Data: []byte(`history page`)},
		"web/templates/footer.html":     &fstest.MapFile{Data: []byte(`{{define "footer"}}<!-- footer -->{{end}}`)},
	}
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
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, nil, 0, 0, 10)
	hist := history.NewRingStore(10)

	mgr := heartbeat.NewManagerFromHeartbeatMap(ctx, cfg, disp.Mailbox(), hist, logger)

	router := NewRouter(webFS, siteRoot, "", version, mgr, hist, disp, logger, true)

	t.Run("GET /", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), version)
	})

	t.Run("GET /static/css/heartbeats.css", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/static/css/heartbeats.css", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "body")
	})

	t.Run("GET /healthz", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/healthz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("GET /bump/test", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/bump/a1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})
}
