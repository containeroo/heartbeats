package handlers

import (
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"text/template"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/containeroo/heartbeats/internal/view"
)

// PartialHandler serves HTML snippets for dashboard sections: heartbeats, receivers, history.
func PartialHandler(
	webFS fs.FS,
	siteRoot string,
	mgr *heartbeat.Manager,
	hist history.Store,
	disp *notifier.Dispatcher,
	logger *slog.Logger,
) http.HandlerFunc {
	tmpl := template.Must(
		template.New("partials").
			Funcs(notifier.FuncMap()).
			ParseFS(webFS,
				"web/templates/heartbeats.html",
				"web/templates/receivers.html",
				"web/templates/history.html",
			),
	)

	bumpURL := strings.TrimSuffix(siteRoot, "/") + "/bump/" // remove trailing slash

	return func(w http.ResponseWriter, r *http.Request) {
		section := path.Base(r.URL.Path)
		var err error

		switch section {
		case "heartbeats":
			err = view.RenderHeartbeats(w, tmpl, bumpURL, mgr, hist)
		case "receivers":
			err = view.RenderReceivers(w, tmpl, disp)
		case "history":
			err = view.RenderHistory(w, tmpl, hist)
		default:
			http.NotFound(w, r)
			return
		}

		if err != nil {
			logger.Error("render "+section+" partial", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}
