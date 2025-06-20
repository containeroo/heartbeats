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

// transitionDelay adds buffer before state transitions to absorb late pings.
const transitionDelay time.Duration = 1 * time.Second

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
	delayTimer  *time.Timer           // timer for deferring transitions (e.g. active → grace)
	pending     func()                // next transition to run after delay
	State       common.HeartbeatState // current state (idle, active, grace, missing, etc.)
}

// NewActor creates a new heartbeat actor.
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

// Mailbox returns the actor's event channel.
func (a *Actor) Mailbox() chan<- common.EventType { return a.mailbox }

// Run starts the actor loop and handles incoming events and timers.
func (a *Actor) Run(ctx context.Context) {
	for {
		// prepare active channels
		var checkCh, graceCh, delayCh <-chan time.Time
		if a.checkTimer != nil {
			checkCh = a.checkTimer.C
		}
		if a.graceTimer != nil {
			graceCh = a.graceTimer.C
		}
		if a.delayTimer != nil {
			delayCh = a.delayTimer.C
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
			// missed expected ping → defer grace transition
			if a.State == common.HeartbeatStateActive {
				a.setPending(a.onEnterGrace)
			}

		case <-graceCh:
			// grace expired → defer missing transition
			if a.State == common.HeartbeatStateGrace {
				a.setPending(a.onEnterMissing)
			}

		case <-delayCh:
			// apply pending state change after delay
			a.runPending()
		}
	}
}

// onReceive handles an incoming heartbeat ping.
func (a *Actor) onReceive() {
	a.stopAllTimers()

	now := time.Now()
	prev := a.State

	a.pending = nil // clear pending state change

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

	a.State = common.HeartbeatStateActive
	a.recordStateChange(prev, a.State)

	a.LastBump = now
	a.checkTimer = time.NewTimer(a.Interval)

	metrics.TotalHeartbeats.With(prometheus.Labels{"heartbeat": a.ID}).Inc()
	metrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": a.ID}).Set(metrics.UP)
}

// onFail handles a manual failure trigger.
func (a *Actor) onFail() {
	a.stopAllTimers()

	now := time.Now()
	prev := a.State

	a.pending = nil // clear pending state change

	// send immediate failure notification
	a.dispatcher.Dispatch(a.ctx, notifier.NotificationData{
		ID:          a.ID,
		Description: a.ID,
		LastBump:    now,
		Status:      common.HeartbeatStateMissing.String(),
		Receivers:   a.Receivers,
	})

	a.State = common.HeartbeatStateFailed
	a.recordStateChange(prev, a.State)

	metrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": a.ID}).Set(metrics.DOWN)
}

// onEnterGrace transitions to Grace state.
func (a *Actor) onEnterGrace() {
	if a.State != common.HeartbeatStateActive {
		return
	}

	prev := a.State
	a.State = common.HeartbeatStateGrace
	a.recordStateChange(prev, a.State)

	a.graceTimer = time.NewTimer(a.Grace)
}

// onEnterMissing transitions to Missing state and sends alert.
func (a *Actor) onEnterMissing() {
	if a.State != common.HeartbeatStateGrace {
		return
	}
	prev := a.State
	a.State = common.HeartbeatStateMissing
	a.recordStateChange(prev, a.State)

	a.dispatcher.Dispatch(a.ctx, notifier.NotificationData{
		ID:          a.ID,
		Description: a.ID,
		LastBump:    a.LastBump,
		Status:      common.HeartbeatStateMissing.String(),
		Receivers:   a.Receivers,
	})
	metrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": a.ID}).Set(metrics.DOWN)
}

// runPending executes a delayed state change, if set.
func (a *Actor) runPending() {
	stopTimer(&a.delayTimer) // clear pending state change

	if fn := a.pending; fn != nil {
		a.pending = nil
		fn()
	}
}

// setPending defers a state change by transitionDelay.
func (a *Actor) setPending(fn func()) {
	stopTimer(&a.delayTimer) // clear any pending state change
	a.pending = fn
	a.delayTimer = time.NewTimer(transitionDelay)
}
