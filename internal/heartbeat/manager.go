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

// NewManagerFromHeartbeatMap creates a Manager from heartbeat config and launches all actors.
func NewManagerFromHeartbeatMap(
	ctx context.Context,
	heartbeatConfigs HeartbeatConfigMap,
	dispatchCh chan<- notifier.NotificationData,
	hist history.Store,
	logger *slog.Logger,
) *Manager {
	m := &Manager{actors: make(map[string]*Actor, len(heartbeatConfigs)), logger: logger}
	for id, hb := range heartbeatConfigs {
		actor := NewActorFromConfig(ActorConfig{
			Ctx:         ctx,
			ID:          id,
			Description: hb.Description,
			Interval:    hb.Interval,
			Grace:       hb.Grace,
			Receivers:   hb.Receivers,
			Logger:      logger,
			History:     hist,
			DispatchCh:  dispatchCh,
		})
		m.actors[id] = actor
		go actor.Run(ctx)
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
