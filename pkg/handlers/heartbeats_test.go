package handlers

import (
	"heartbeats/pkg/heartbeat"
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

	heartbeatStore := heartbeat.NewStore()
	notificationStore := notify.NewStore()
	version := "1.0.0"
	siteRoot := "http://localhost:8080"

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

	err := heartbeatStore.Add("test", h)
	assert.NoError(t, err)

	ns := &notify.Notification{
		Name:       "test",
		Type:       "email",
		Enabled:    new(bool),
		MailConfig: &notifier.MailConfig{},
	}
	*ns.Enabled = false

	err = notificationStore.Add("test", ns)
	assert.NoError(t, err)

	aferoFS := setupAferoFSForHeartbeats()

	mux := http.NewServeMux()
	mux.Handle("/", Heartbeats(log, aferoToCustomAferoFS(aferoFS), version, siteRoot, heartbeatStore, notificationStore))

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
		mux.Handle("/", Heartbeats(log, aferoToCustomAferoFS(invalidFS), version, siteRoot, heartbeatStore, notificationStore))
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code, "Expected status code 500")
		assert.Contains(t, rec.Body.String(), "Internal Server Error", "Expected internal server error message")
	})
}

func TestCreateHeartbeatList(t *testing.T) {
	heartbeatStore := heartbeat.NewStore()
	notificationStore := notify.NewStore()

	interval := time.Minute
	grace := time.Minute

	h := &heartbeat.Heartbeat{
		Name:          "test",
		Enabled:       new(bool),
		Interval:      &timer.Timer{Interval: &interval},
		Grace:         &timer.Timer{Interval: &grace},
		LastPing:      time.Now(),
		Notifications: []string{"test"},
	}
	*h.Enabled = true

	err := heartbeatStore.Add("test", h)
	assert.NoError(t, err)

	ns := &notify.Notification{
		Name:       "test",
		Type:       "email",
		Enabled:    new(bool),
		MailConfig: &notifier.MailConfig{},
	}
	*ns.Enabled = true

	err = notificationStore.Add("test", ns)
	assert.NoError(t, err)

	t.Run("CreateHeartbeatList", func(t *testing.T) {
		heartbeatList := createHeartbeatList(heartbeatStore, notificationStore)
		assert.Len(t, heartbeatList, 1, "Expected one heartbeat in list")
		assert.Equal(t, "test", heartbeatList[0].Name, "Expected heartbeat name to be 'test'")
		assert.Equal(t, "email", heartbeatList[0].Notifications[0].Type, "Expected notification type to be 'email'")
		assert.Equal(t, true, heartbeatList[0].Notifications[0].Enabled, "Expected notification to be enabled")
		assert.Equal(t, interval, *heartbeatList[0].Interval.Interval, "Expected interval duration to match")
		assert.Equal(t, grace, *heartbeatList[0].Grace.Interval, "Expected grace duration to match")
	})
}
