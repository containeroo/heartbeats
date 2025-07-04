package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"text/template"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/stretchr/testify/assert"
)

func loadTestTemplate(t *testing.T, name string, content string) *template.Template {
	t.Helper()

	fs := fstest.MapFS{
		"web/templates/" + name + ".html": &fstest.MapFile{Data: []byte(content)},
	}
	tmpl, err := template.New(name).
		Funcs(notifier.FuncMap()).
		ParseFS(fs, "web/templates/"+name+".html")

	assert.NoError(t, err)

	return tmpl
}

func encodePayload(t *testing.T, payload any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	return data
}

func TestPartialHandler(t *testing.T) {
	t.Parallel()

	webFS := fstest.MapFS{
		"web/templates/heartbeats.html": &fstest.MapFile{Data: []byte(`{{define "heartbeats"}}HEARTBEATS{{end}}`)},
		"web/templates/receivers.html":  &fstest.MapFile{Data: []byte(`{{define "receivers"}}RECEIVERS{{end}}`)},
		"web/templates/history.html":    &fstest.MapFile{Data: []byte(`{{define "history"}}HISTORY{{end}}`)},
	}

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	_ = hist.Append(context.Background(), history.MustNewEvent(
		history.EventTypeHeartbeatReceived,
		"hb1",
		history.RequestMetadataPayload{
			Method:    "GET",
			Source:    "127.0.0.1",
			UserAgent: "Go-http-client",
		},
	))
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, logger, hist, 1, 1, 10)

	mgr := heartbeat.NewManagerFromHeartbeatMap(context.Background(), map[string]heartbeat.HeartbeatConfig{
		"hb1": {
			Description: "desc",
			Interval:    10 * time.Second,
			Grace:       5 * time.Second,
			Receivers:   []string{"r1"},
		},
	}, disp.Mailbox(), hist, logger)

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/partials/invalid", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(webFS, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "404 page not found\n", rr.Body.String())
	})

	t.Run("heartbeats", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/partials/heartbeats", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(webFS, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "HEARTBEATS")
	})

	t.Run("Receivers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/partials/receivers", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(webFS, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "RECEIVERS")
	})

	t.Run("History", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/partials/history", nil)
		rr := httptest.NewRecorder()

		handler := PartialHandler(webFS, "http://localhost", mgr, hist, disp, logger)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "HISTORY")
	})
}

func TestRenderHeartbeats(t *testing.T) {
	t.Parallel()

	tmpl := loadTestTemplate(t, "heartbeats", `{{define "heartbeats"}}{{range .Heartbeats}}{{.ID}}:{{.Status}};{{end}}{{end}}`)

	hist := history.NewRingStore(10)
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	store := notifier.InitializeStore(nil, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, nil, hist, 1, 1, 10)
	mgr := heartbeat.NewManagerFromHeartbeatMap(context.Background(), map[string]heartbeat.HeartbeatConfig{
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
	}, disp.Mailbox(), hist, nil)

	var buf bytes.Buffer
	a := mgr.Get("b")
	a.LastBump = time.Now()

	err := renderHeartbeats(&buf, tmpl, "http://localhost", mgr, hist)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "a:idle;b:idle")
}

func TestRenderReceivers(t *testing.T) {
	t.Parallel()

	tmpl := loadTestTemplate(t, "receivers", `{{define "receivers"}}{{range .Receivers}}{{.Type}};{{end}}{{end}}`)

	r := notifier.ReceiverConfig{
		SlackConfigs:   []notifier.SlackConfig{{Channel: "channel"}},
		MSTeamsConfigs: []notifier.MSTeamsConfig{{WebhookURL: "example.com"}},
		EmailConfigs:   []notifier.EmailConfig{{EmailDetails: notifier.EmailDetails{To: []string{"to"}}}},
	}
	rc := map[string]notifier.ReceiverConfig{"r": r}

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	store := notifier.InitializeStore(rc, false, "0.0.0", logger)
	disp := notifier.NewDispatcher(store, nil, nil, 1, 1, 10)

	var buf bytes.Buffer
	err := renderReceivers(&buf, tmpl, disp)
	assert.NoError(t, err)
	assert.Equal(t, buf.String(), "slack;email;msteams;")
}

func TestRenderHistory(t *testing.T) {
	t.Parallel()

	tmpl := loadTestTemplate(t, "history", `{{define "history"}}{{range .Events}}{{.Type}}:{{.Details}};{{end}}{{end}}`)

	hist := history.NewRingStore(10)

	// Record state change event
	_ = hist.Append(context.Background(), history.MustNewEvent(
		history.EventTypeStateChanged,
		"hb1",
		history.StateChangePayload{
			From: "missing",
			To:   "received",
		},
	))

	// Record notification event
	_ = hist.Append(context.Background(), history.MustNewEvent(
		history.EventTypeNotificationSent,
		"hb1",
		history.NotificationPayload{
			Receiver: "r1",
			Type:     "mock-type",
			Target:   "mock-target",
		},
	))

	var buf bytes.Buffer
	err := renderHistory(&buf, tmpl, hist)

	assert.NoError(t, err)
	assert.Equal(t,
		"NotificationSent:Notification sent to \"r1\" via mock-type (mock-target);StateChanged:missing → received;",
		buf.String(),
	)

	t.Run("notification failed", func(t *testing.T) {
		t.Parallel()

		hist := history.NewRingStore(10)
		_ = hist.Append(context.Background(), history.MustNewEvent(
			history.EventTypeNotificationFailed,
			"hb2",
			history.NotificationPayload{
				Receiver: "r2",
				Type:     "mock-type",
				Target:   "mock-target",
				Error:    "fail!",
			},
		))

		var buf bytes.Buffer
		err := renderHistory(&buf, tmpl, hist)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), `Notification to "r2" via mock-type (mock-target) failed: fail!`)
	})

	t.Run("invalid notification payload", func(t *testing.T) {
		t.Parallel()

		hist := history.NewRingStore(10)
		ev := history.Event{
			Timestamp:   time.Now(),
			Type:        history.EventTypeNotificationSent,
			HeartbeatID: "hb3",
			RawPayload:  json.RawMessage(`{invalid-json}`), // deliberately broken
		}
		_ = hist.Append(context.Background(), ev)

		var buf bytes.Buffer
		err := renderHistory(&buf, tmpl, hist)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Invalid notification payload")
	})

	t.Run("invalid state change", func(t *testing.T) {
		t.Parallel()

		hist := history.NewRingStore(10)

		// Force an invalid payload: a raw string is not a struct
		ev := history.Event{
			Timestamp:   time.Now(),
			Type:        history.EventTypeStateChanged,
			HeartbeatID: "hb4",
			RawPayload:  json.RawMessage(`"not a valid state change"`),
		}
		_ = hist.Append(context.Background(), ev)

		var buf bytes.Buffer
		err := renderHistory(&buf, tmpl, hist)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Invalid state change payload")
		t.Log(buf.String())
	})

	t.Run("invalid request metadata", func(t *testing.T) {
		t.Parallel()

		hist := history.NewRingStore(10)

		// deliberately use a string literal instead of a struct
		ev := history.Event{
			Timestamp:   time.Now(),
			Type:        history.EventTypeHeartbeatReceived,
			HeartbeatID: "hb5",
			RawPayload:  json.RawMessage(`"this is not a struct"`), // will not decode into struct
		}
		_ = hist.Append(context.Background(), ev)

		var buf bytes.Buffer
		err := renderHistory(&buf, tmpl, hist)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Invalid request metadata")
		t.Log(buf.String())
	})

	t.Run("unknown event type", func(t *testing.T) {
		t.Parallel()

		hist := history.NewRingStore(10)
		ev := history.MustNewEvent(history.EventType("999"), "hb6", nil)
		ev.RawPayload = nil
		_ = hist.Append(context.Background(), ev)

		var buf bytes.Buffer
		err := renderHistory(&buf, tmpl, hist)
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Unknown event type")
	})
}
