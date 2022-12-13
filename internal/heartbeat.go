package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/containeroo/heartbeats/internal/ago"
	"github.com/containeroo/heartbeats/internal/cache"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/timer"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type NotificationsMap struct {
	Name    string `mapstructure:"name"`
	Type    string `mapstructure:"type"`
	Enabled bool   `mapstructure:"enabled"`
}

// Heartbeat is a struct for a heartbeat
type Heartbeat struct {
	Name             string             `mapstructure:"name"`
	Enabled          *bool              `mapstructure:"enabled"`
	Description      string             `mapstructure:"description"`
	Interval         time.Duration      `mapstructure:"interval"`
	Grace            time.Duration      `mapstructure:"grace"`
	LastPing         time.Time          `mapstructure:"lastPing"`
	Status           string             `mapstructure:"status"`
	Notifications    []string           `mapstructure:"notifications"`
	NotificationsMap []NotificationsMap `mapstructure:",-,omitempty"`
	IntervalTimer    *timer.Timer       `mapstructure:"intervalTimer,omitempty"`
	GraceTimer       *timer.Timer       `mapstructure:"graceTimer,omitempty"`
}

// TimeAgo returns a string with the time since the last ping
func (h *Heartbeat) TimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return ago.Calculate.Format(t)
}

// GotPing starts the timer for the heartbeat (heartbeatName)
func (h *Heartbeat) GotPing() {
	var msg string
	// Timer is expired, grace is running but not completed
	if h.IntervalTimer != nil && h.IntervalTimer.Completed && h.GraceTimer != nil && !h.GraceTimer.Completed {
		log.Trace("Timer is expired, grace is running but not completed")
		h.GraceTimer.Cancel()
		msg = "got ping. Stop grace timer"
	}

	// Heartbeat is running and not expired
	if h.IntervalTimer != nil && !h.IntervalTimer.Completed {
		log.Trace("Heartbeat is running and not expired")
		h.IntervalTimer.Reset(h.Interval)
		msg = fmt.Sprintf("got ping. Reset timer with interval %s", h.Interval)
	}

	// This is the first ping or after a grace period has expired
	// IntervalTimer is nil when never started
	// IntervalTimer.Completed is true when expired
	// GraceTimer is nil when never started
	// GraceTimer.Completed is true when expired
	if (h.IntervalTimer == nil || h.IntervalTimer.Completed) && (h.GraceTimer == nil || h.GraceTimer.Completed) {
		log.Tracef("First ping or ping after a grace period has expired")
		h.IntervalTimer = timer.NewTimer(h.Interval, func() {
			msg := fmt.Sprintf("Timer is expired. Start grace timer with duration %s", h.Grace)
			log.Infof("%s %s", h.Name, msg)
			cache.Local.Add(h.Name, cache.History{
				Event:   cache.EventGrace,
				Message: fmt.Sprintf("Timer is expired. Start grace timer with duration %s", h.Grace),
			})
			h.Status = "GRACE"
			h.GraceTimer = timer.NewTimer(h.Grace,
				Notify(h.Name, GRACE))
		})
		msg = fmt.Sprintf("Start timer with interval %s", h.Interval)
	}

	// inform the user that the heartbeat is running again
	if h.Status != "" && h.Status != "OK" {
		go Notify(h.Name, OK)()
	}

	h.LastPing = time.Now()
	h.Status = "OK"
	metrics.PromMetrics.TotalHeartbeats.With(prometheus.Labels{"heartbeat": h.Name}).Inc()
	metrics.PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": h.Name}).Set(1)

	log.Infof("%s %s", h.Name, msg)
	cache.Local.Add(h.Name, cache.History{
		Event:   cache.EventPing,
		Message: msg,
	})
}

// GotPingFail stops the timer for the heartbeat
func (h *Heartbeat) GotPingFail() {

	// cancel grace timer if running
	if h.GraceTimer != nil && !h.GraceTimer.Completed {
		h.GraceTimer.Cancel()
	}

	// cancel interval timer if running
	if h.IntervalTimer != nil && !h.IntervalTimer.Completed {
		h.IntervalTimer.Cancel()
	}

	go Notify(h.Name, FAILED)()

	h.LastPing = time.Now()
	h.Status = "FAIL"
	metrics.PromMetrics.TotalHeartbeats.With(prometheus.Labels{"heartbeat": h.Name}).Inc()
	metrics.PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": h.Name}).Set(0)

	log.Infof("%s got '/fail' ping", h.Name)
	cache.Local.Add(h.Name, cache.History{
		Event:   cache.EventFailed,
		Message: "got '/fail' ping",
	})
}

// GetServiceByType returns notification settings by type
func (h *Heartbeats) GetServiceByName(notificationType string) (*Service, error) {
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
