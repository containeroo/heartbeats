package handlers

import (
	"heartbeats/pkg/config"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/notify"
	"heartbeats/pkg/notify/notifier"
	"heartbeats/pkg/timer"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func setupAferoFSForHeartbeats() afero.Fs {
	aferoFS := afero.NewMemMapFs()

	templateFiles := []string{
		"heartbeats.html",
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

func TestHeartbeatsHandler(t *testing.T) {
	log := logger.NewLogger(true)
	config.App.HeartbeatStore = heartbeat.NewStore()
	config.App.NotificationStore = notify.NewStore()
	config.HistoryStore = history.NewStore()

	h := &heartbeat.Heartbeat{
		Name:          "test",
		Enabled:       new(bool),
		Interval:      &timer.Timer{Interval: new(time.Duration)},
		Grace:         &timer.Timer{Interval: new(time.Duration)},
		Notifications: []string{"test"},
	}
	*h.Enabled = true
	*h.Interval.Interval = time.Minute
	*h.Grace.Interval = time.Minute

	err := config.App.HeartbeatStore.Add("test", h)
	assert.NoError(t, err)

	ns := &notify.Notification{
		Name:       "test",
		Type:       "email",
		Enabled:    new(bool),
		MailConfig: &notifier.MailConfig{},
	}
	*ns.Enabled = false

	err = config.App.NotificationStore.Add("test", ns)
	assert.NoError(t, err)

	aferoFS := setupAferoFSForHeartbeats()

	mux := http.NewServeMux()
	mux.Handle("/", Heartbeats(log, aferoToCustomAferoFS(aferoFS)))

	t.Run("Heartbeat page renders correctly", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200")
		assert.Contains(t, rec.Body.String(), "Heartbeat", "Expected 'Heartbeat' in response body")
		assert.Contains(t, rec.Body.String(), "test", "Expected 'test' in response body")
	})

	// Simulate a template parsing error by using an invalid template path
	t.Run("Template parsing error", func(t *testing.T) {
		invalidFS := afero.NewMemMapFs() // Empty FS to simulate missing templates
		mux := http.NewServeMux()
		mux.Handle("/", Heartbeats(log, aferoToCustomAferoFS(invalidFS)))

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code, "Expected status code 500")
		assert.Contains(t, rec.Body.String(), "Internal Server Error", "Expected internal server error message")
	})
}
