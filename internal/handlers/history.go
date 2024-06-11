package handlers

import (
	"embed"
	"fmt"
	"heartbeats/internal/config"
	"heartbeats/internal/history"
	"heartbeats/internal/logger"
	"html/template"
	"net/http"
)

// History handles the /history/{id} endpoint
func History(logger logger.Logger, staticFS embed.FS) http.Handler {
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
			Name    string
			Entries []history.HistoryEntry
		}{
			Name:    heartbeatName,
			Entries: config.HistoryStore.Get(heartbeatName).GetAllEntries(),
		}

		if err := tmpl.ExecuteTemplate(w, "history", data); err != nil {
			logger.Errorf("Failed to execute template. %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}
