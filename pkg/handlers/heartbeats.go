package handlers

import (
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/notify"
	"heartbeats/pkg/timer"
	"heartbeats/pkg/utils"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/Masterminds/sprig"
)

// HeartbeatPageData represents the data structure passed to the heartbeat template.
type HeartbeatPageData struct {
	Version    string
	SiteRoot   string
	Heartbeats []*HeartbeatData
}

// HeartbeatData represents individual heartbeat data.
type HeartbeatData struct {
	Name          string
	Status        string
	Interval      *timer.Timer
	Grace         *timer.Timer
	LastPing      time.Time
	Notifications []NotificationState
}

// NotificationState represents the state of a notification.
type NotificationState struct {
	Name    string
	Type    string
	Enabled bool
}

// Heartbeats handles the / endpoint
func Heartbeats(logger logger.Logger, staticFS fs.FS, version, siteRoot string, heartbeatStore *heartbeat.Store, notificationStore *notify.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmap := sprig.TxtFuncMap()
		fmap["isTrue"] = utils.IsTrue
		fmap["isFalse"] = utils.IsFalse
		fmap["formatTime"] = utils.FormatTime

		tmpl, err := template.New("heartbeat").
			Funcs(fmap).
			ParseFS(
				staticFS,
				"web/templates/heartbeats.html",
				"web/templates/footer.html",
			)
		if err != nil {
			logger.Errorf("Failed to parse template. %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		heartbeatList := createHeartbeatList(heartbeatStore, notificationStore)

		data := HeartbeatPageData{
			Version:    version,
			SiteRoot:   siteRoot,
			Heartbeats: heartbeatList,
		}

		if err := tmpl.ExecuteTemplate(w, "heartbeat", data); err != nil {
			logger.Errorf("Failed to execute template. %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// createHeartbeatList formats the heartbeats from the heartbeatStore for the template
func createHeartbeatList(heartbeatStore *heartbeat.Store, notificationStore *notify.Store) []*HeartbeatData {
	var heartbeatDataList []*HeartbeatData
	for _, h := range heartbeatStore.GetAll() {
		var notifications []NotificationState
		for _, notificationName := range h.Notifications {
			n := notificationStore.Get(notificationName)
			if n != nil {
				notifications = append(notifications, NotificationState{
					Name:    n.Name,
					Enabled: *n.Enabled,
					Type:    n.Type,
				})
			}
		}
		heartbeatDataList = append(heartbeatDataList, &HeartbeatData{
			Name:          h.Name,
			Status:        h.Status,
			Interval:      h.Interval,
			Grace:         h.Grace,
			LastPing:      h.LastPing,
			Notifications: notifications,
		})
	}

	return heartbeatDataList
}
