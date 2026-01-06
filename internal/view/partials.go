package view

import (
	"fmt"
	"io"
	"sort"
	"text/template"
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

// RenderHeartbeats builds HeartbeatView slice, sorts it, and executes the template.
func RenderHeartbeats(
	w io.Writer,
	tmpl *template.Template,
	bumpURL string,
	mgr *heartbeat.Manager,
	hist history.Store,
) error {
	actors := mgr.List()
	views := make([]HeartbeatView, 0, len(actors))
	for id, a := range actors {
		evs := hist.ListByID(id)

		views = append(views, HeartbeatView{
			ID:              id,
			Status:          a.State.String(),
			Description:     a.Description,
			Interval:        a.Interval.String(),
			IntervalSeconds: a.Interval.Seconds(),
			Grace:           a.Grace.String(),
			LastBump:        a.LastBump,
			URL:             bumpURL + id,
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

// RenderReceivers builds ReceiverView slice, sorts it by (ID,Type), and executes the template.
func RenderReceivers(
	w io.Writer,
	tmpl *template.Template,
	disp *notifier.Dispatcher,
) error {
	raw := disp.List()

	views := make([]ReceiverView, 0, len(raw))
	for rid, nots := range raw {
		for _, n := range nots {
			rv := ReceiverView{
				ID:          rid,
				Type:        n.Type(),
				Destination: n.Target(),
				LastSent:    n.LastSent(),
				LastErr:     n.LastErr(),
			}

			views = append(views, rv)
		}
	}

	// Sort alphabetically by ID for consistent ordering
	sort.Slice(views, func(i, j int) bool { return views[i].ID < views[j].ID })

	data := struct{ Receivers []ReceiverView }{Receivers: views}

	return tmpl.ExecuteTemplate(w, "receivers", data)
}

// RenderHistory sorts events newest-first, builds the filter list, and executes the template.
func RenderHistory(
	w io.Writer,
	tmpl *template.Template,
	hist history.Store,
) error {
	raw := hist.List()

	views := make([]HistoryView, 0, len(raw))
	for _, e := range raw {
		var det string

		switch e.Type {
		case history.EventTypeNotificationSent, history.EventTypeNotificationFailed:
			var p history.NotificationPayload
			if err := e.DecodePayload(&p); err == nil {
				if p.Error != "" {
					det = fmt.Sprintf("Notification to %q via %s (%s) failed: %s",
						p.Receiver, p.Type, p.Target, p.Error,
					)
				} else {
					det = fmt.Sprintf("Notification sent to %q via %s (%s)",
						p.Receiver, p.Type, p.Target,
					)
				}
			} else {
				det = "Invalid notification payload"
			}

		case history.EventTypeStateChanged:
			var p history.StateChangePayload
			if err := e.DecodePayload(&p); err == nil {
				det = fmt.Sprintf("%s → %s", p.From, p.To)
			} else {
				det = "Invalid state change payload"
			}

		case history.EventTypeHeartbeatReceived, history.EventTypeHeartbeatFailed:
			var p history.RequestMetadataPayload
			if err := e.DecodePayload(&p); err == nil {
				det = fmt.Sprintf("%s from %s with %q", p.Method, p.Source, p.UserAgent)
			} else {
				det = "Invalid request metadata"
			}

		default:
			det = "Unknown event type"
		}

		views = append(views, HistoryView{
			Timestamp:   e.Timestamp,
			Type:        e.Type.String(),
			HeartbeatID: e.HeartbeatID,
			Details:     det,
		})
	}

	// Newest first
	sort.Slice(views, func(i, j int) bool {
		return views[j].Timestamp.Before(views[i].Timestamp)
	})

	return tmpl.ExecuteTemplate(w, "history", struct{ Events []HistoryView }{Events: views})
}
