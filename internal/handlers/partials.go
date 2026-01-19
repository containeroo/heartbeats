package handlers

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"text/template"

	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/containeroo/heartbeats/internal/view"
)

// PartialHandler serves HTML snippets for dashboard sections: heartbeats, receivers, history.
func (a *API) PartialHandler(
	siteRoot string,
) http.HandlerFunc {
	tmpl := template.Must(
		template.New("partials").
			Funcs(notifier.FuncMap()).
			ParseFS(a.webFS,
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
			err = view.RenderHeartbeats(w, tmpl, bumpURL, a.mgr, a.hist)
		case "receivers":
			err = view.RenderReceivers(w, tmpl, a.disp)
		case "history":
			err = view.RenderHistory(w, tmpl, a.hist)
		default:
			a.respondJSON(w, http.StatusNotFound, errorResponse{Error: fmt.Sprintf("unknown partial %q", section)})
			return
		}

		if err != nil {
			a.Logger.Error("render "+section+" partial", "error", err)
			a.respondJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		}
	}
}
