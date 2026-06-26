package sender

import (
	"context"
	"log/slog"
	"time"

	kit "github.com/containeroo/notifykit/notify"

	htypes "github.com/containeroo/heartbeats/internal/heartbeat/types"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notify"
	"github.com/containeroo/heartbeats/internal/runner"
)

type HeartbeatSender struct {
	Heartbeat *htypes.Heartbeat
	Notifier  kit.Notifier
	History   history.Recorder
	Logger    *slog.Logger
	Metrics   *metrics.Registry
}

// Late handles a late heartbeat event.
func (s *HeartbeatSender) Late(now time.Time, since time.Duration, payload string) {
	if !s.Heartbeat.AlertOnLate {
		return
	}
	s.enqueue(notify.NewEvent(
		s.Heartbeat.ID,
		s.Heartbeat.Title,
		htypes.StatusLate.String(),
		payload,
		since,
		now,
		s.Heartbeat.Config.Interval,
		s.Heartbeat.Config.LateAfter,
		s.Heartbeat.ReceiverIDs,
	))
}

// Missing handles a missing heartbeat event.
func (s *HeartbeatSender) Missing(now time.Time, since time.Duration, payload string) {
	s.enqueue(notify.NewEvent(
		s.Heartbeat.ID,
		s.Heartbeat.Title,
		htypes.StatusMissing.String(),
		payload,
		since,
		now,
		s.Heartbeat.Config.Interval,
		s.Heartbeat.Config.LateAfter,
		s.Heartbeat.ReceiverIDs,
	))
}

// Recovered handles a recovery event.
func (s *HeartbeatSender) Recovered(now time.Time, payload string) {
	if !s.Heartbeat.AlertOnRecovery {
		return
	}
	s.enqueue(notify.NewEvent(
		s.Heartbeat.ID,
		s.Heartbeat.Title,
		htypes.StatusRecovered.String(),
		payload,
		0,
		now,
		s.Heartbeat.Config.Interval,
		s.Heartbeat.Config.LateAfter,
		s.Heartbeat.ReceiverIDs,
	))
}

// Transition handles a stage transition event.
func (s *HeartbeatSender) Transition(now time.Time, from runner.Stage, to runner.Stage, since time.Duration) {
	s.Logger.Info("Runner stage transitioned",
		"event", logging.EventStageTransition.String(),
		"from", from.String(),
		"to", to.String(),
		"since", since.Truncate(time.Millisecond).String(),
	)
	s.History.Add(history.Event{
		Time:        now,
		Type:        history.EventHeartbeatTransition.String(),
		HeartbeatID: s.Heartbeat.ID,
		Status:      to.String(),
		Fields: map[string]any{
			"from":  from.String(),
			"to":    to.String(),
			"since": since.String(),
		},
	})
	s.Metrics.SetHeartbeatState(s.Heartbeat.ID, to.String())
}

// enqueue sends a notification to notifykit.
func (s *HeartbeatSender) enqueue(n *notify.Event) {
	if s == nil || s.Notifier == nil || n == nil {
		return
	}
	id, err := s.Notifier.Enqueue(context.Background(), n)
	if err != nil {
		s.Logger.Error("Notification queue failed",
			"event", logging.EventNotificationDeliveryFailed.String(),
			"heartbeat", n.Heartbeat,
			"status", n.StatusValue,
			"err", err,
		)
		return
	}
	if id == "" {
		s.Logger.Info("Notification skipped",
			"event", logging.EventNotificationMissing.String(),
			"heartbeat", n.Heartbeat,
			"status", n.StatusValue,
		)
		return
	}
	s.Logger.Info("Notification queued",
		"event", logging.EventNotificationDelivered.String(),
		"queue_id", id,
		"heartbeat", n.Heartbeat,
		"status", n.StatusValue,
	)
}
