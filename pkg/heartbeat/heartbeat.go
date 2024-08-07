package heartbeat

import (
	"context"
	"fmt"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/metrics"
	"heartbeats/pkg/notify"
	"heartbeats/pkg/timer"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Heartbeat represents the configuration and state of a monitoring heartbeat.
type Heartbeat struct {
	Name          string       `yaml:"name"`
	Enabled       *bool        `yaml:"enabled,omitempty"`
	Description   string       `yaml:"description,omitempty"`
	LastPing      time.Time    `yaml:"lastPing,omitempty"`
	Interval      *timer.Timer `yaml:"interval,omitempty" json:"-"`
	Grace         *timer.Timer `yaml:"grace,omitempty" json:"-"`
	Notifications []string     `yaml:"notifications,omitempty"`
	Status        string       `yaml:"status,omitempty"`
	SendResolve   *bool        `yaml:"sendResolve,omitempty"`
	mu            sync.Mutex   `mapstructure:"-" yaml:"-"`
}

// Store manages the storage and retrieval of heartbeats.
type Store struct {
	heartbeats map[string]*Heartbeat
	mu         sync.RWMutex
}

// MarshalYAML implements the yaml.Marshaler interface.
func (s *Store) MarshalYAML() (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.heartbeats, nil
}

// NewStore creates a new HeartbeatStore.
func NewStore() *Store {
	return &Store{
		heartbeats: make(map[string]*Heartbeat),
	}
}

// GetAll returns all heartbeats.
func (s *Store) GetAll() map[string]*Heartbeat {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.heartbeats
}

// Get returns a single heartbeat.
func (s *Store) Get(name string) *Heartbeat {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.heartbeats[name]
}

// Add adds a heartbeat to the HeartbeatStore.
func (s *Store) Add(name string, heartbeat *Heartbeat) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if heartbeat.Interval == nil {
		return fmt.Errorf("interval timer is required")
	}

	if heartbeat.Grace == nil {
		return fmt.Errorf("grace timer is required")
	}

	if exists := s.heartbeats[name]; exists != nil {
		return fmt.Errorf("heartbeat '%s' already exists", name)
	}

	heartbeat.Name = name
	s.heartbeats[name] = heartbeat

	return nil
}

// Delete removes a heartbeat by name.
func (s *Store) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.heartbeats, name)
}

// StartInterval initializes and starts the interval timer for the heartbeat.
func (h *Heartbeat) StartInterval(ctx context.Context, log logger.Logger, notificationStore *notify.Store, hi *history.History) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.updateStatus(ctx, log, StatusOK, notificationStore, hi)
	h.log(log, logger.InfoLevel, hi, EventInterval, fmt.Sprintf("start interval timer %s", h.Interval.Interval))

	h.StopTimers() // Stop all timers before starting new ones

	h.Interval.RunTimer(ctx, func() {
		h.log(log, logger.DebugLevel, hi, EventInterval, fmt.Sprintf("interval timer %s elapsed", h.Interval.Interval))
		h.StartGrace(ctx, log, notificationStore, hi)
	})

	metrics.TotalHeartbeats.With(prometheus.Labels{"heartbeat": h.Name}).Inc()
	metrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": h.Name}).Set(metrics.UP)
}

// StartGrace initializes and starts the grace timer for the heartbeat.
func (h *Heartbeat) StartGrace(ctx context.Context, log logger.Logger, notificationStore *notify.Store, hi *history.History) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Status = StatusGrace.String() // only update status
	h.log(log, logger.InfoLevel, hi, EventGrace, fmt.Sprintf("start grace timer %s", h.Grace.Interval))

	h.StopTimers()

	h.Grace.RunTimer(ctx, func() {
		h.handleGraceTimerElapsed(ctx, log, notificationStore, hi)
	})
}

// handleGraceTimerElapsed is a helper method to handle the logic when the grace timer elapses.
func (h *Heartbeat) handleGraceTimerElapsed(ctx context.Context, log logger.Logger, notificationStore *notify.Store, hi *history.History) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Status = StatusNOK.String() // only update status
	h.log(log, logger.DebugLevel, hi, EventGrace, fmt.Sprintf("grace timer %s elapsed", h.Grace.Interval))

	h.SendNotifications(ctx, log, notificationStore, hi, false)

	metrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": h.Name}).Set(metrics.DOWN)
}

// StopTimers stops both the interval and grace timers.
func (h *Heartbeat) StopTimers() {
	h.Interval.StopTimer()
	h.Grace.StopTimer()
}

// SendNotifications sends notifications based on the current status of the heartbeat.
func (h *Heartbeat) SendNotifications(ctx context.Context, log logger.Logger, notificationStore *notify.Store, hi *history.History, isResolved bool) {
	for _, n := range h.Notifications {
		notification := notificationStore.Get(n)
		if notification == nil {
			log.Debugf("%s Notification '%s' not found", EventSend, n)
			continue
		}

		if notification.Enabled != nil && !*notification.Enabled {
			log.Debugf("%s ignore '%s' (%s) because it is not enabled.", h.Name, notification.Name, notification.Type)
			continue
		}

		if err := notification.Send(ctx, h, isResolved, notify.DefaultFormatter); err != nil {
			h.log(log, logger.ErrorLevel, hi, EventSend, fmt.Sprintf("%s error sending notification '%s' (%s). %v", h.Name, notification.Name, notification.Type, err))
			continue
		}
		h.log(log, logger.InfoLevel, hi, EventSend, fmt.Sprintf("successfully send notification to %s (%s)", notification.Name, notification.Type))
	}
}

// UpdateStatus updates the heartbeat's status and triggers notifications if needed.
func (h *Heartbeat) updateStatus(ctx context.Context, log logger.Logger, newStatus Status, notificationStore *notify.Store, hi *history.History) {
	currentStatus := h.Status
	h.Status = newStatus.String()
	h.LastPing = time.Now()

	if currentStatus == StatusNOK.String() && newStatus == StatusOK && (h.SendResolve == nil || *h.SendResolve) {
		log.Debugf("%s switched status from 'nok' to 'ok'", h.Name)
		h.SendNotifications(ctx, log, notificationStore, hi, true)
	}
}

// log writes a message to the log and history.
func (h *Heartbeat) log(logger logger.Logger, level logger.Level, hi *history.History, eventType Event, msg string) {
	logMsg := fmt.Sprintf("%s %s", h.Name, msg)
	logger.Write(level, logMsg)

	hi.Add(history.Event(eventType), msg, nil)
}
