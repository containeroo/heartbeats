package render

import (
	"time"

	"github.com/containeroo/heartbeats/internal/notify/types"
)

// RenderData is the template context for notifications.
type RenderData struct {
	HeartbeatID string         // Heartbeat identifier.
	Title       string         // Heartbeat title.
	Status      string         // Status identifier.
	Subject     string         // Rendered subject/title.
	Payload     string         // Raw last payload.
	Timestamp   time.Time      // Event time.
	Interval    time.Duration  // Expected heartbeat interval.
	LateAfter   time.Duration  // Late window duration.
	Receiver    string         // Receiver name.
	Vars        map[string]any // Receiver variables.
	Since       time.Duration  // Time since last heartbeat.
}

// newRenderData builds a RenderData snapshot for templates.
func NewRenderData(n types.Payload, receiver string, vars map[string]any, subject string) RenderData {
	return RenderData{
		HeartbeatID: n.HeartbeatID,
		Title:       n.Title,
		Status:      n.Status,
		Subject:     subject,
		Payload:     n.Payload,
		Timestamp:   n.Timestamp,
		Interval:    n.Interval,
		LateAfter:   n.LateAfter,
		Receiver:    receiver,
		Vars:        vars,
		Since:       n.Since,
	}
}
