package sender

import (
	"log/slog"
	"time"

	htypes "github.com/containeroo/heartbeats/internal/heartbeat/types"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notify/event"
	ntypes "github.com/containeroo/heartbeats/internal/notify/types"
	"github.com/containeroo/heartbeats/internal/runner"
)

type HeartbeatSender struct {
	Heartbeat *htypes.Heartbeat
	Notifier  ntypes.Notifier
	History   history.Recorder
	Logger    *slog.Logger
	Metrics   *metrics.Registry
}

// Late handles a late heartbeat event (suppressed).
func (s *HeartbeatSender) Late(now time.Time, since time.Duration, payload string) {
	if !s.Heartbeat.AlertOnLate {
		return
	}
	s.enqueue(event.NewEvent(
		s.Heartbeat.ID,
		htypes.StatusLate.String(),
		payload,
		since,
		now,
		s.Heartbeat.Receivers,
	))
}

// Missing handles a missing heartbeat event.
func (s *HeartbeatSender) Missing(now time.Time, since time.Duration, payload string) {
	s.enqueue(event.NewEvent(
		s.Heartbeat.ID,
		htypes.StatusMissing.String(),
		payload,
		since,
		now,
		s.Heartbeat.Receivers,
	))
}

// Recovered handles a recovery event.
func (s *HeartbeatSender) Recovered(now time.Time, payload string) {
	if !s.Heartbeat.AlertOnRecovery {
		return
	}
	s.enqueue(event.NewEvent(
		s.Heartbeat.ID,
		htypes.StatusRecovered.String(),
		payload,
		0,
		now,
		s.Heartbeat.Receivers,
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

// enqueue sends a notification to the mailbox.
func (s *HeartbeatSender) enqueue(n ntypes.Notification) { s.Notifier.Enqueue(n) }
