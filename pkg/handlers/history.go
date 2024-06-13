package handlers

import (
	"fmt"
	"heartbeats/pkg/config"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"html/template"
	"io/fs"
	"net/http"
)

// History handles the /history/{id} endpoint
func History(logger logger.Logger, staticFS fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		heartbeatName := r.PathValue("id")
		logger.Debugf("%s /history/%s %s %s", r.Method, heartbeatName, r.RemoteAddr, r.UserAgent())

		h := config.App.HeartbeatStore.Get(heartbeatName)
		if h == nil {
			errMsg := fmt.Sprintf("Heartbeat '%s' not found", heartbeatName)
			logger.Warn(errMsg)
			http.Error(w, errMsg, http.StatusNotFound)
			return
		}

		fmap := template.FuncMap{
			"formatTime": formatTime,
		}

		tmpl, err := template.New("history").
			Funcs(fmap).
			ParseFS(
				staticFS,
				"web/templates/history.html",
				"web/templates/footer.html",
			)
		if err != nil {
			logger.Errorf("Failed to parse template. %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := struct {
			Version string
			Name    string
			Entries []history.HistoryEntry
		}{
			Version: config.App.Version,
			Name:    heartbeatName,
			Entries: config.HistoryStore.Get(heartbeatName).GetAllEntries(),
		}

		if err := tmpl.ExecuteTemplate(w, "history", data); err != nil {
			logger.Errorf("Failed to execute template. %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}
