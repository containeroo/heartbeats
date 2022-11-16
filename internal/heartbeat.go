package internal

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// GotPing starts the timer for the heartbeat (heartbeatName)
func (h *Heartbeat) GotPing() {

	// Grace is running and timer is not expired
	if h.GraceTimer != nil && !h.GraceTimer.Completed {
		log.Infof("%s got ping. Stop grace timer", h.Name)
		h.GraceTimer.Cancel()
	}

	// Heartbeat is running and not expired, so reset timer
	if h.IntervalTimer != nil && !h.IntervalTimer.Completed {
		log.Infof("%s got ping. Reset timer with interval %s", h.Name, h.Interval)
		h.IntervalTimer.Reset(h.Interval)

		if h.Status == "NOK" {
			NotificationFunc(h.Name, true)() // notify user that heartbeat turned OK
		}
	}

	// Timer is not running, so start timer
	// This is the first ping or after a grace period has expired
	if h.IntervalTimer == nil || h.IntervalTimer.Completed {
		log.Infof("%s Start timer with interval %s", h.Name, h.Interval)
		h.IntervalTimer = NewTimer(h.Interval, func() {
			log.Infof("%s Timer is expired. Start grace timer with duration %s", h.Name, h.Grace)
			h.Status = "GRACE"
			h.GraceTimer = NewTimer(h.Grace,
				NotificationFunc(h.Name, false))
		})
	}
	h.LastPing = time.Now()
	h.Status = "OK"
}

// GetHeartbeatByName search heartbeat in HeartbeatsConfig.Heartbeats by name
func GetHeartbeatByName(name string) (*Heartbeat, error) {
	for i, heartbeat := range HeartbeatsServer.Heartbeats {
		// compare lowercase strings
		if strings.EqualFold(heartbeat.Name, name) {
			return &HeartbeatsServer.Heartbeats[i], nil
		}
	}
	return nil, fmt.Errorf("Heartbeat with name %s not found", name)
}
