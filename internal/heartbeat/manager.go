package heartbeat

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
)

// Manager routes HTTP pings to Actors.
type Manager struct {
	actors map[string]*Actor
	logger *slog.Logger
}

// NewManager constructs and starts all Actors.
func NewManager(
	ctx context.Context,
	cfg map[string]HeartbeatConfig,
	dispatchCh chan<- notifier.NotificationData,
	hist history.Store,
	logger *slog.Logger,
) *Manager {
	m := &Manager{actors: make(map[string]*Actor), logger: logger}
	for id, c := range cfg {
		act := NewActor(
			ctx,
			id,
			c.Description,
			c.Interval,
			c.Grace,
			c.Receivers,
			logger,
			hist,
			dispatchCh,
		)
		m.actors[id] = act
		go act.Run(ctx)
	}
	return m
}

// List returns all configured heartbeats.
func (m *Manager) List() map[string]*Actor { return m.actors }

// Get returns one heartbeat’s info by ID.
func (m *Manager) Get(id string) *Actor { return m.actors[id] }

// HandleReceive pings the Actor or logs unknown ID.
func (m *Manager) HandleReceive(id string) error {
	a, ok := m.actors[id]
	if !ok {
		return fmt.Errorf("unknown heartbeat id %q", id)
	}
	a.Mailbox() <- common.EventReceive
	return nil
}

// HandleFail marks the Actor failed or logs unknown ID.
func (m *Manager) HandleFail(id string) error {
	a, ok := m.actors[id]
	if !ok {
		return fmt.Errorf("unknown heartbeat id %q", id)
	}
	a.Mailbox() <- common.EventFail
	return nil
}
