package notifier

import "time"

// NotificationData is the payload for alerts.
type NotificationData struct {
	ID          string    `json:"id"`          // heartbeat ID
	Name        string    `json:"name"`        // human-friendly name
	Description string    `json:"description"` // heartbeat description
	LastBump    time.Time `json:"lastPing"`    // time of last ping
	Status      string    `json:"status"`      // current status
	Receivers   []string  `json:"receivers"`   // list of receiver IDs
	Title       string    `json:"title"`       // rendered notification title
	Message     string    `json:"message"`     // rendered notification body
}
