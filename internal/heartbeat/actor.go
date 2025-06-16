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
		// prepare active channels
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
			// handle heartbeat or manual failure
			switch ev {
			case common.EventReceive:
				a.onReceive()
			case common.EventFail:
				a.onFail()
			}

		case <-checkCh:
			// missed heartbeat interval: enter grace
			if a.State == common.HeartbeatStateActive {
				a.onCheckTimeout()
			}

		case <-graceCh:
			// grace expired: declare missing
			if a.State == common.HeartbeatStateGrace {
				a.onGraceTimeout()
			}
		}
	}
}

// onReceive handles a heartbeat ping.
func (a *Actor) onReceive() {
	now := time.Now()
	prev := a.State

	// stop any active timers
	stopTimer(&a.checkTimer)
	stopTimer(&a.graceTimer)

	// if recovering from missing, send recovery notice
	if prev == common.HeartbeatStateMissing {
		a.dispatcher.Dispatch(a.ctx, notifier.NotificationData{
			ID:          a.ID,
			Description: a.ID,
			LastBump:    now,
			Status:      common.HeartbeatStateRecovered.String(),
			Receivers:   a.Receivers,
		})
	}

	// move to active and restart interval timer
	newState := common.HeartbeatStateActive
	a.State = newState
	a.recordStateChange(prev, newState)
	a.LastBump = now
	a.checkTimer = time.NewTimer(a.Interval)

	metrics.TotalHeartbeats.With(prometheus.Labels{"heartbeat": a.ID}).Inc()
	metrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": a.ID}).Set(metrics.UP)
}

// onFail handles a manual failure event.
func (a *Actor) onFail() {
	now := time.Now()
	prev := a.State

	// stop any active timers
	stopTimer(&a.checkTimer)
	stopTimer(&a.graceTimer)

	// send immediate failure notification
	a.dispatcher.Dispatch(a.ctx, notifier.NotificationData{
		ID:          a.ID,
		Description: a.ID,
		LastBump:    now,
		Status:      common.HeartbeatStateMissing.String(),
		Receivers:   a.Receivers,
	})

	// mark as failed and update status
	newState := common.HeartbeatStateFailed
	a.State = newState
	a.recordStateChange(prev, newState)

	metrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": a.ID}).Set(metrics.DOWN)
}

// onCheckTimeout transitions from Active → Grace.
func (a *Actor) onCheckTimeout() {
	prev := a.State
	newState := common.HeartbeatStateGrace
	a.State = newState
	a.recordStateChange(prev, newState)

	a.graceTimer = time.NewTimer(a.Grace)
}

// onGraceTimeout transitions from Grace → Missing.
func (a *Actor) onGraceTimeout() {
	prev := a.State
	newState := common.HeartbeatStateMissing
	a.State = newState
	a.recordStateChange(prev, newState)

	a.dispatcher.Dispatch(a.ctx, notifier.NotificationData{
		ID:          a.ID,
		Description: a.ID,
		LastBump:    a.LastBump,
		Status:      common.HeartbeatStateMissing.String(),
		Receivers:   a.Receivers,
	})
	metrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": a.ID}).Set(metrics.DOWN)
}
