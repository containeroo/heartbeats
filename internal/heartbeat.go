package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Heartbeat is a struct for a heartbeat
type Heartbeat struct {
	Name          string        `mapstructure:"name"`
	Description   string        `mapstructure:"description"`
	Interval      time.Duration `mapstructure:"interval"`
	Grace         time.Duration `mapstructure:"grace"`
	LastPing      time.Time     `mapstructure:"lastPing"`
	Status        string        `mapstructure:"status"`
	Notifications []string      `mapstructure:"notifications"`
	IntervalTimer *Timer        `mapstructure:"-,omitempty" deepcopier:"skip"`
	GraceTimer    *Timer        `mapstructure:"-,omitempty" deepcopier:"skip"`
}

// TimeAgo returns a string with the time since the last ping
func (h *Heartbeat) TimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return CalculateAgo.Format(t)
}

// GotPing starts the timer for the heartbeat (heartbeatName)
func (h *Heartbeat) GotPing() {

	// Timer is expired, grace is running but not completed
	if h.IntervalTimer != nil && h.IntervalTimer.Completed && h.GraceTimer != nil && !h.GraceTimer.Completed {
		log.Tracef("Timer is expired, grace is running but not completed")
		h.GraceTimer.Cancel()
		log.Infof("%s got ping. Stop grace timer", h.Name)
	}

	// Heartbeat is running and not expired
	if h.IntervalTimer != nil && !h.IntervalTimer.Completed {
		log.Tracef("Heartbeat is running and not expired")
		h.IntervalTimer.Reset(h.Interval)
		log.Infof("%s got ping. Reset timer with interval %s", h.Name, h.Interval)
	}

	// This is the first ping or after a grace period has expired
	// IntervalTimer is nil when never started
	// IntervalTimer.Completed is true when expired
	// GraceTimer is nil when never started
	// GraceTimer.Completed is true when expired
	if (h.IntervalTimer == nil || h.IntervalTimer.Completed) && (h.GraceTimer == nil || h.GraceTimer.Completed) {
		log.Tracef("First ping or ping after a grace period has expired")
		h.IntervalTimer = NewTimer(h.Interval, func() {
			log.Infof("%s Timer is expired. Start grace timer with duration %s", h.Name, h.Grace)
			h.Status = "GRACE"
			h.GraceTimer = NewTimer(h.Grace,
				NotificationFunc(h.Name, GRACE))
		})
		log.Infof("%s Start timer with interval %s", h.Name, h.Interval)
	}

	// inform the user that the heartbeat is running again
	if h.Status != "" && h.Status != "OK" {
		go NotificationFunc(h.Name, OK)()
	}

	h.LastPing = time.Now()
	h.Status = "OK"
	PromMetrics.TotalHeartbeats.With(prometheus.Labels{"heartbeat": h.Name}).Inc()
	PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": h.Name}).Set(1)
}

func (h *Heartbeat) GotPingFail() {

	// cancel grace timer if running
	if h.GraceTimer != nil && !h.GraceTimer.Completed {
		h.GraceTimer.Cancel()
	}

	// cancel interval timer if running
	if h.IntervalTimer != nil && !h.IntervalTimer.Completed {
		h.IntervalTimer.Cancel()
	}

	go NotificationFunc(h.Name, FAILED)()

	h.LastPing = time.Now()
	h.Status = "FAIL"
	PromMetrics.TotalHeartbeats.With(prometheus.Labels{"heartbeat": h.Name}).Inc()
	PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": h.Name}).Set(0)
}

// GetServiceByType returns notification settings by type
func (h *Heartbeats) GetServiceByName(notificationType string) (*Notification, error) {
	for i, notification := range h.Notifications.Services {
		if strings.EqualFold(notification.Name, notificationType) {
			return &h.Notifications.Services[i], nil
		}
	}
	return nil, fmt.Errorf("Notification settings for type «%s» not found", notificationType)
}

// GetHeartbeatByName search heartbeat in HeartbeatsConfig.Heartbeats by name and returns it
func (h *Heartbeats) GetHeartbeatByName(name string) (*Heartbeat, error) {
	for i, heartbeat := range h.Heartbeats {
		if strings.EqualFold(heartbeat.Name, name) {
			return &h.Heartbeats[i], nil
		}
	}
	return nil, fmt.Errorf("Heartbeat with name %s not found", name)
}
