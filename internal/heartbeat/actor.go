package heartbeat

import (
	"context"
	"log/slog"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notifier"
	"github.com/prometheus/client_golang/prometheus"
)

// Actor handles a single heartbeat’s lifecycle.
type Actor struct {
	ctx         context.Context       // context for cancellation and timeouts
	ID          string                // unique heartbeat identifier
	Interval    time.Duration         // expected interval between pings
	Description string                // human‐friendly description of this heartbeat
	Grace       time.Duration         // grace period before triggering an alert
	Receivers   []string              // list of receiver IDs to notify upon alerts
	mailbox     chan common.EventType // incoming event channel for this actor
	logger      *slog.Logger          // structured logger scoped to this actor
	hist        history.Store         // history store for recording events
	dispatcher  *notifier.Dispatcher  // dispatches notifications to receivers
	LastBump    time.Time             // timestamp of the last received heartbeat
	checkTimer  *time.Timer           // timer waiting for the next heartbeat
	graceTimer  *time.Timer           // timer for the grace period countdown
	State       common.HeartbeatState // current state (idle, active, grace, missing)
}

// NewActor constructs a new Actor, wiring in history recording and notification dispatch.
func NewActor(
	ctx context.Context,
	id string,
	description string,
	interval, grace time.Duration,
	receivers []string,
	logger *slog.Logger,
	hist history.Store,
	dispatcher *notifier.Dispatcher,
) *Actor {
	return &Actor{
		ctx:         ctx,
		ID:          id,
		Description: description,
		Interval:    interval,
		Grace:       grace,
		Receivers:   receivers,
		mailbox:     make(chan common.EventType, 1),
		logger:      logger,
		hist:        hist,
		dispatcher:  dispatcher,
		State:       common.HeartbeatStateIdle,
	}
}

// Mailbox returns the channel on which this actor receives events.
func (a *Actor) Mailbox() chan<- common.EventType { return a.mailbox }

// Run starts the actor’s event loop, listening for pings, failures, and timer ticks.
func (a *Actor) Run(ctx context.Context) {
	for {
		// prepare channels from timers if set
		var checkCh, graceCh <-chan time.Time
		if a.checkTimer != nil {
			checkCh = a.checkTimer.C
		}
		if a.graceTimer != nil {
			graceCh = a.graceTimer.C
		}

		select {
		case <-ctx.Done():
			// shutdown requested
			return

		case ev := <-a.mailbox:
			// incoming EventReceive or EventFail
			switch ev {
			case common.EventReceive:
				a.onReceive()
			case common.EventFail:
				a.onFail()
			}

		case <-checkCh:
			// check timeout expired: start grace period
			if a.State == common.HeartbeatStateActive {
				a.onCheckTimeout()
			}

		case <-graceCh:
			// grace timeout expired: trigger missing alert
			if a.State == common.HeartbeatStateGrace {
				a.onGraceTimeout()
			}
		}
	}
}

// onReceive handles a heartbeat ping.
// Stops any existing timers, transitions to Active (or Recovered), records history, and starts next check.
func (a *Actor) onReceive() {
	now := time.Now()
	prev := a.State

	// stop prior timers to avoid leaks or stale ticks
	stopTimer(&a.checkTimer)
	stopTimer(&a.graceTimer)

	// if recovering from missing, send a recovered notification
	if prev == common.HeartbeatStateMissing {
		data := notifier.NotificationData{
			ID:          a.ID,
			Description: a.ID,
			LastBump:    now,
			Status:      common.HeartbeatStateRecovered.String(),
			Receivers:   a.Receivers,
		}
		a.dispatcher.Dispatch(a.ctx, data)
		_ = a.hist.RecordEvent(a.ctx, history.Event{
			Timestamp:    now,
			Type:         history.EventTypeNotificationSent,
			HeartbeatID:  a.ID,
			Notification: &data,
		})
	}

	newState := common.HeartbeatStateActive
	// log and record the state change
	a.logger.Info("state change",
		"heartbeat", a.ID,
		"from", prev.String(),
		"to", newState.String(),
	)
	_ = a.hist.RecordEvent(a.ctx, history.Event{
		Timestamp:   now,
		Type:        history.EventTypeStateChanged,
		HeartbeatID: a.ID,
		PrevState:   prev.String(),
		NewState:    newState.String(),
	})

	// update internal state and last bump time, then start the check timer
	a.State = newState
	a.LastBump = now
	a.checkTimer = time.NewTimer(a.Interval)

	// update prometheus metric
	metrics.TotalHeartbeats.With(prometheus.Labels{"heartbeat": a.ID}).Inc()
	metrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": a.ID}).Set(metrics.UP)
}

// onFail handles a manual failure event.
// Cancels timers, sends an immediate missing alert, and records the state change.
func (a *Actor) onFail() {
	now := time.Now()
	prev := a.State

	stopTimer(&a.checkTimer)
	stopTimer(&a.graceTimer)

	// send manual failure notification
	data := notifier.NotificationData{
		ID:          a.ID,
		Description: a.ID,
		LastBump:    now,
		Status:      common.HeartbeatStateMissing.String(),
		Receivers:   a.Receivers,
	}
	a.dispatcher.Dispatch(a.ctx, data)
	_ = a.hist.RecordEvent(a.ctx, history.Event{
		Timestamp:    now,
		Type:         history.EventTypeNotificationSent,
		HeartbeatID:  a.ID,
		Notification: &data,
	})

	newState := common.HeartbeatStateFailed
	a.logger.Info("state change",
		"heartbeat", a.ID,
		"from", prev.String(),
		"to", newState.String(),
	)
	_ = a.hist.RecordEvent(a.ctx, history.Event{
		Timestamp:   now,
		Type:        history.EventTypeStateChanged,
		HeartbeatID: a.ID,
		PrevState:   prev.String(),
		NewState:    newState.String(),
	})

	// update to Failed state (no new timer)
	a.State = newState

	// update prometheus metric
	metrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": a.ID}).Set(metrics.DOWN)
}

// onCheckTimeout transitions from Active → Grace and records the change.
func (a *Actor) onCheckTimeout() {
	now := time.Now()
	prev := a.State
	newState := common.HeartbeatStateGrace

	a.logger.Info("state change",
		"heartbeat", a.ID,
		"from", prev.String(),
		"to", newState.String(),
	)
	_ = a.hist.RecordEvent(a.ctx, history.Event{
		Timestamp:   now,
		Type:        history.EventTypeStateChanged,
		HeartbeatID: a.ID,
		PrevState:   prev.String(),
		NewState:    newState.String(),
	})

	a.State = newState
	// start the grace timer
	a.graceTimer = time.NewTimer(a.Grace)
}

// onGraceTimeout transitions from Grace → Missing, sends an overdue alert, and records both.
func (a *Actor) onGraceTimeout() {
	now := time.Now()
	prev := a.State

	// record the grace-to-missing state change
	newState := common.HeartbeatStateMissing
	a.logger.Info("state change",
		"heartbeat", a.ID,
		"from", prev.String(),
		"to", newState.String(),
	)
	_ = a.hist.RecordEvent(a.ctx, history.Event{
		Timestamp:   now,
		Type:        history.EventTypeStateChanged,
		HeartbeatID: a.ID,
		PrevState:   prev.String(),
		NewState:    newState.String(),
	})
	a.State = newState

	// send overdue notification
	data := notifier.NotificationData{
		ID:          a.ID,
		Description: a.ID,
		LastBump:    a.LastBump,
		Status:      common.HeartbeatStateMissing.String(),
		Receivers:   a.Receivers,
	}
	a.dispatcher.Dispatch(a.ctx, data)
	_ = a.hist.RecordEvent(a.ctx, history.Event{
		Timestamp:    now,
		Type:         history.EventTypeNotificationSent,
		HeartbeatID:  a.ID,
		Notification: &data,
	})
	metrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": a.ID}).Set(metrics.DOWN)
}
