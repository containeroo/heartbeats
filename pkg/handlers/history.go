package handlers

import (
	"fmt"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/utils"
	"html/template"
	"io/fs"
	"net/http"
)

// History handles the /history/{id} endpoint
func History(logger logger.Logger, staticFS fs.FS, version string, heartbeatStore *heartbeat.Store, historyStore *history.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		heartbeatName := r.PathValue("id")

		clientIP := getClientIP(r)
		logger.Debugf("%s /history/%s %s %s", r.Method, heartbeatName, clientIP, r.UserAgent())

		h := heartbeatStore.Get(heartbeatName)
		if h == nil {
			errMsg := fmt.Sprintf("Heartbeat '%s' not found", heartbeatName)
			logger.Warn(errMsg)
			http.Error(w, errMsg, http.StatusNotFound)
			return
		}

		hi := historyStore.Get(heartbeatName)
		if hi == nil {
			errMsg := fmt.Sprintf("No history found for heartbeat '%s'", heartbeatName)
			logger.Warn(errMsg)
			http.Error(w, errMsg, http.StatusNotFound)
			return
		}

		fmap := template.FuncMap{
			"formatTime": utils.FormatTime,
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
			Version: version,
			Name:    heartbeatName,
			Entries: hi.GetAllEntries(),
		}

		if err := tmpl.ExecuteTemplate(w, "history", data); err != nil {
			logger.Errorf("Failed to execute template. %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}
