package handlers

import (
	"embed"
	"heartbeats/internal/config"
	"heartbeats/internal/history"
	"heartbeats/internal/logger"
	"html/template"
	"net/http"

	"github.com/Masterminds/sprig"
)

// History handles the /history endpoint
func History(logger logger.Logger, staticFS embed.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		heartbeatName := r.PathValue("id")
		logger.Debugf("%s /history/%s %s %s", r.Method, heartbeatName, r.RemoteAddr, r.UserAgent())

		h := config.App.HeartbeatStore.Get(heartbeatName)
		if h == nil {
			logger.Warnf("Heartbeat «%s» not found", heartbeatName)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("heartbeat not found")) // Make linter happy
			return
		}

		fmap := sprig.TxtFuncMap()
		fmap["formatTime"] = formatTime

		tmpl, err := template.New("heartbeat").
			Funcs(fmap).
			ParseFS(
				staticFS,
				"web/templates/history.html",
				"web/templates/footer.html",
			)
		if err != nil {
			logger.Errorf("Failed to parse template: %v", err)
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
			logger.Errorf("Failed to execute template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
