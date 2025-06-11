package handlers

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"sort"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
)

type HeartbeatView struct {
	ID              string
	Status          string
	Description     string
	Interval        string
	IntervalSeconds float64 // for table sorting
	Grace           string
	GraceSeconds    float64 // for table sorting
	LastBump        time.Time
	URL             string // full URL to copy
	Receivers       []string
	HasHistory      bool
}

// ReceiverView for rendering each notifier instance.
type ReceiverView struct {
	ID          string // receiver ID
	Type        string // "slack", "email", etc.
	Destination string // e.g. channel name, email addrs
	LastSent    time.Time
	LastErr     error
}

// HistoryView is what the template actually sees.
// Details is already formatted for display.
type HistoryView struct {
	Timestamp   time.Time
	Type        string // EventType
	HeartbeatID string
	Details     string // e.g. "Notification Sent", "old → new", or blank
}

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

	return func(w http.ResponseWriter, r *http.Request) {
		section := path.Base(r.URL.Path)
		var err error

		switch section {
		case "heartbeats":
			err = renderHeartbeats(w, tmpl, siteRoot, mgr, hist)
		case "receivers":
			err = renderReceivers(w, tmpl, disp)
		case "history":
			err = renderHistory(w, tmpl, hist)
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

// renderHeartbeats builds HeartbeatView slice, sorts it, and executes the template.
func renderHeartbeats(
	w io.Writer,
	tmpl *template.Template,
	siteRoot string,
	mgr *heartbeat.Manager,
	hist history.Store,
) error {
	actors := mgr.List()
	views := make([]HeartbeatView, 0, len(actors))
	for id, a := range actors {
		evs := hist.GetEventsByID(id)

		views = append(views, HeartbeatView{
			ID:              id,
			Status:          a.State.String(),
			Description:     a.Description,
			Interval:        a.Interval.String(),
			IntervalSeconds: a.Interval.Seconds(),
			Grace:           a.Grace.String(),
			LastBump:        a.LastBump,
			URL:             siteRoot + id,
			Receivers:       a.Receivers,
			HasHistory:      len(evs) > 0,
		})
	}
	// Sort alphabetically by ID for consistent ordering
	sort.Slice(views, func(i, j int) bool { return views[i].ID < views[j].ID })

	data := struct {
		Heartbeats []HeartbeatView
	}{
		Heartbeats: views,
	}
	return tmpl.ExecuteTemplate(w, "heartbeats", data)
}

// renderReceivers builds ReceiverView slice, sorts it by (ID,Type), and executes the template.
func renderReceivers(
	w io.Writer,
	tmpl *template.Template,
	disp *notifier.Dispatcher,
) error {
	raw := disp.List()

	views := make([]ReceiverView, 0, len(raw))
	for rid, nots := range raw {
		for _, n := range nots {
			rv := ReceiverView{
				ID:       rid,
				Type:     n.Type(),
				LastSent: n.LastSent(),
				LastErr:  n.LastErr(),
			}
			// Derive the destination field based on concrete Notifier type.
			switch x := n.(type) {
			case *notifier.SlackConfig:
				rv.Destination = x.Channel
			case *notifier.EmailConfig:
				rv.Destination = formatEmailRecipients(x.EmailDetails)
			case *notifier.MSTeamsConfig:
				rv.Destination = masqueradeURL(x.WebhookURL, 4)

			}

			views = append(views, rv)
		}
	}

	// Sort alphabetically by ID for consistent ordering
	sort.Slice(views, func(i, j int) bool { return views[i].ID < views[j].ID })

	data := struct{ Receivers []ReceiverView }{Receivers: views}

	return tmpl.ExecuteTemplate(w, "receivers", data)
}

// renderHistory sorts events newest-first, builds the filter list, and executes the template.
func renderHistory(
	w io.Writer,
	tmpl *template.Template,
	hist history.Store,
) error {
	raw := hist.GetEvents()

	// build our slice of HistoryView
	views := make([]HistoryView, 0, len(raw))
	for _, e := range raw {

		// pick Details
		var det string
		switch {
		case e.Payload != nil:
			det = "Notification Sent"
		case e.PrevState != "":
			det = e.PrevState + " → " + e.NewState
		case e.Type == history.EventTypeHeartbeatReceived, e.Type == history.EventTypeHeartbeatFailed:
			det = fmt.Sprintf("%s from %s with %s", e.Method, e.Source, e.UserAgent)
		}

		views = append(views, HistoryView{
			Timestamp:   e.Timestamp,
			Type:        e.Type.String(),
			HeartbeatID: e.HeartbeatID,
			Details:     det,
		})
	}

	// sort oldest first
	sort.Slice(views, func(i, j int) bool {
		return views[i].Timestamp.Before(views[j].Timestamp)
	})

	// execute template
	data := struct{ Events []HistoryView }{Events: views}

	return tmpl.ExecuteTemplate(w, "history", data)
}
