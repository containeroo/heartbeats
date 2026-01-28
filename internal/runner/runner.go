package runner

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Stage represents the heartbeat state relative to thresholds.
type Stage int

const (
	// StageNever indicates the heartbeat is never seen.
	StageNever Stage = iota
	// StageOK indicates the heartbeat is within thresholds.
	StageOK
	// StageLate indicates the heartbeat is beyond the late threshold.
	StageLate
	// StageMissing indicates the heartbeat is beyond the missing threshold.
	StageMissing
)

// String returns the stage identifier.
func (s Stage) String() string {
	switch s {
	case StageNever:
		return "never"
	case StageOK:
		return "ok"
	case StageLate:
		return "late"
	case StageMissing:
		return "missing"
	default:
		return "unknown"
	}
}

// Config controls the runner timing and alert behavior.
type Config struct {
	LateAfter       time.Duration // Late window duration after check interval.
	CheckInterval   time.Duration // Polling interval for checks.
	AlertOnRecovery bool          // Whether to emit recovery alerts.
	AlertOnLate     bool          // Whether to emit late alerts.
}

// Sender delivers alerts for runner state transitions.
type Sender interface {
	Late(now time.Time, since time.Duration, payload string)
	Missing(now time.Time, since time.Duration, payload string)
	Recovered(now time.Time, payload string)
	Transition(now time.Time, from Stage, to Stage, since time.Duration)
}

// State stores the last seen heartbeat and alert metadata.
type State struct {
	mu          sync.RWMutex       // Guards all fields.
	lastSeen    time.Time          // Timestamp of last heartbeat.
	lastPayload string             // Body of last heartbeat payload.
	stage       Stage              // Current stage.
	mailbox     chan HeartbeatType // Notifies when a heartbeat arrives.
}

// Snapshot captures a consistent view of State.
type Snapshot struct {
	LastSeen    time.Time // Timestamp of last heartbeat.
	LastPayload string    // Body of last heartbeat payload.
	Stage       Stage     // Current stage.
}

// NewState initializes a State in the OK stage.
func NewState() *State {
	return &State{
		stage:   StageNever,
		mailbox: make(chan HeartbeatType, 1),
	}
}

// UpdateSeen records a heartbeat payload and notifies the runner.
func (s *State) UpdateSeen(now time.Time, payload string) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	s.mu.Lock()
	s.lastSeen = now
	s.lastPayload = payload
	s.mu.Unlock()
	return s.receiveSeen()
}

// Snapshot returns a copy of the current state.
func (s *State) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Snapshot{
		LastSeen:    s.lastSeen,
		LastPayload: s.lastPayload,
		Stage:       s.stage,
	}
}

// MarkMissing updates the state to missing and records the alert time.
func (s *State) MarkMissing(now time.Time) {
	s.mu.Lock()
	s.stage = StageMissing
	s.mu.Unlock()
}

// MarkOK resets the stage to OK.
func (s *State) MarkOK() {
	s.mu.Lock()
	s.stage = StageOK
	s.mu.Unlock()
}

// MarkLate sets the stage to late.
func (s *State) MarkLate() {
	s.mu.Lock()
	s.stage = StageLate
	s.mu.Unlock()
}

// Mailbox returns the heartbeat event channel.
func (s *State) Mailbox() <-chan HeartbeatType { return s.mailbox }

// HeartbeatType represents a heartbeat event.
type HeartbeatType int

const HeartbeatReceive HeartbeatType = iota // HeartbeatReceive represents a received heartbeat.

// receiveSeen enqueues a received heartbeat event.
func (s *State) receiveSeen() bool {
	select {
	case s.mailbox <- HeartbeatReceive:
		return true
	default:
		return false
	}
}

// enterLate transitions to late and arms the late window timer.
func (s *State) enterLate(
	timer *stageTimer,
	sender Sender,
	snap Snapshot,
	now time.Time,
	since time.Duration,
	lateAfter time.Duration,
	alertOnLate bool,
) {
	sender.Transition(now, snap.Stage, StageLate, since)
	s.MarkLate()
	if alertOnLate {
		sender.Late(now, since, snap.LastPayload)
	}
	timer.Reset(lateAfter)
}

// enterMissing transitions to missing, sends alert, and stops timers.
func (s *State) enterMissing(
	timer *stageTimer,
	sender Sender,
	snap Snapshot,
	now time.Time,
	since time.Duration,
) {
	sender.Transition(now, snap.Stage, StageMissing, since)
	s.MarkMissing(now)
	sender.Missing(now, since, snap.LastPayload)
	timer.Stop()
}

// onReceive handles a heartbeat receive event.
func (s *State) onReceive(
	timer *stageTimer,
	sender Sender,
	interval time.Duration,
	alertOnRecovery bool,
	now time.Time,
) {
	snap := s.Snapshot()
	prev := snap.Stage
	since := now.Sub(snap.LastSeen)
	sender.Transition(now, prev, StageOK, since)
	s.MarkOK()
	timer.Reset(interval)
	if prev == StageMissing && alertOnRecovery {
		sender.Recovered(now, snap.LastPayload)
	}
}

// Run executes the periodic runner loop until ctx is canceled.
func Run(ctx context.Context, state *State, cfg Config, sender Sender, logger *slog.Logger) {
	var timer stageTimer

	for {
		select {
		case <-ctx.Done():
			// shutdown requested
			return
		case ev := <-state.Mailbox():
			// handle heartbeat, manual failure and test
			switch ev {
			case HeartbeatReceive:
				now := time.Now().UTC()
				state.onReceive(&timer, sender, cfg.CheckInterval, cfg.AlertOnRecovery, now)
			}

		case <-timer.C():
			// timer fired
			now := time.Now().UTC()
			snap := state.Snapshot()
			if snap.LastSeen.IsZero() {
				continue
			}
			since := now.Sub(snap.LastSeen)
			switch snap.Stage {
			case StageOK: // state was ok, change it late
				state.enterLate(&timer, sender, snap, now, since, cfg.LateAfter, cfg.AlertOnLate)
			case StageLate: // state was late, change it missing
				state.enterMissing(&timer, sender, snap, now, since)
			}
		}
	}
}
