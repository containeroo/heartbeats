package handlers

import (
	"bytes"
	"context"
	"html/template"
	"io/fs"
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

func setupTestFS() fs.FS {
	return fstest.MapFS{
		"web/templates/heartbeats.html": &fstest.MapFile{Data: []byte(`{{define "heartbeats"}}HEARTBEATS{{end}}`)},
		"web/templates/receivers.html":  &fstest.MapFile{Data: []byte(`{{define "receivers"}}RECEIVERS{{end}}`)},
		"web/templates/history.html":    &fstest.MapFile{Data: []byte(`{{define "history"}}HISTORY{{end}}`)},
	}
}

func loadTestTemplate(t *testing.T, name string, content string) *template.Template {
	fs := fstest.MapFS{
		"web/templates/" + name + ".html": &fstest.MapFile{Data: []byte(content)},
	}
	tmpl, err := template.New(name).
		Funcs(notifier.FuncMap()).
		ParseFS(fs, "web/templates/"+name+".html")
	assert.NoError(t, err)
	return tmpl
}

func TestPartialHandler(t *testing.T) {
	t.Parallel()

	fs := setupTestFS()
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	_ = hist.RecordEvent(context.Background(), history.Event{
		Timestamp:   time.Now(),
		Type:        "HeartbeatReceived",
		HeartbeatID: "hb1",
		Method:      "GET",
		Source:      "127.0.0.1",
		UserAgent:   "Go-http-client",
	})

	disp := notifier.NewDispatcher(notifier.InitializeStore(nil, false, logger), logger)

	mgr := heartbeat.NewManager(context.Background(), map[string]heartbeat.HeartbeatConfig{
		"hb1": {
			Description: "desc",
			Interval:    10 * time.Second,
			Grace:       5 * time.Second,
			Receivers:   []string{"r1"},
		},
	}, disp, hist, logger)

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/invalid", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(fs, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "404 page not found\n", rr.Body.String())
	})

	t.Run("heartbeats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/heartbeats", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(fs, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "HEARTBEATS")
	})

	t.Run("Receivers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/partials/receivers", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(fs, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "RECEIVERS")
	})

	t.Run("History", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/partials/history", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(fs, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "HISTORY")
	})
}

func TestRenderHeartbeats(t *testing.T) {
	t.Parallel()

	tmpl := loadTestTemplate(t, "heartbeats", `{{define "heartbeats"}}{{range .Heartbeats}}{{.ID}}:{{.Status}};{{end}}{{end}}`)

	hist := history.NewRingStore(10)
	disp := notifier.NewDispatcher(notifier.InitializeStore(nil, false, nil), nil)
	mgr := heartbeat.NewManager(context.Background(), map[string]heartbeat.HeartbeatConfig{
		"b": {
			Description: "b-desc",
			Interval:    1 * time.Second,
			Grace:       1 * time.Second,
			Receivers:   []string{"r1"},
		},
		"a": {
			Description: "a-desc",
			Interval:    1 * time.Second,
			Grace:       1 * time.Second,
			Receivers:   []string{"r1"},
		},
	}, disp, hist, nil)

	var buf bytes.Buffer
	a := mgr.Get("b")
	a.LastBump = time.Now()

	err := renderHeartbeats(&buf, tmpl, "http://localhost", mgr, hist)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "a:idle;b:idle")
}

func TestRenderReceivers(t *testing.T) {
	t.Parallel()

	tmpl := loadTestTemplate(t, "receivers", `{{define "receivers"}}{{range .Receivers}}{{.Status}};{{end}}{{end}}`)

	r := notifier.ReceiverConfig{
		SlackConfigs:   []notifier.SlackConfig{{Channel: "channel"}},
		MSTeamsConfigs: []notifier.MSTeamsConfig{{WebhookURL: "example.com"}},
		EmailConfigs:   []notifier.EmailConfig{{EmailDetails: notifier.EmailDetails{To: []string{"to"}}}},
	}
	rc := map[string]notifier.ReceiverConfig{"r": r}

	disp := notifier.NewDispatcher(notifier.InitializeStore(rc, false, nil), nil)

	var buf bytes.Buffer
	err := renderReceivers(&buf, tmpl, disp)
	assert.NoError(t, err)
	assert.Equal(t, buf.String(), "Never;Never;Never;")
}

func TestRenderHistory(t *testing.T) {
	t.Parallel()

	tmpl := loadTestTemplate(t, "history", `{{define "history"}}{{range .Events}}{{.Type}}:{{.Details}};{{end}}{{end}}`)

	hist := history.NewRingStore(10)
	_ = hist.RecordEvent(context.Background(), history.Event{
		Timestamp:   time.Now().Add(3 * time.Second),
		Type:        "HeartbeatReceived",
		HeartbeatID: "hb1",
		Method:      "GET",
		Source:      "127.0.0.1",
		UserAgent:   "Go-http-client",
		Notification: &notifier.NotificationData{
			ID:          "r1",
			Name:        "r1",
			Description: "desc-1",
			LastBump:    time.Now(),
		},
	})

	_ = hist.RecordEvent(context.Background(), history.Event{
		Timestamp:   time.Now(),
		Type:        "HeartbeatReceived",
		HeartbeatID: "hb1",
		Method:      "GET",
		Source:      "127.0.0.1",
		UserAgent:   "Go-http-client",
		PrevState:   "missing",
		NewState:    "received",
	})

	var buf bytes.Buffer
	err := renderHistory(&buf, tmpl, hist)
	assert.NoError(t, err)
	assert.Equal(t, "HeartbeatReceived:Notification Sent;HeartbeatReceived:missing → received;", buf.String())
}
