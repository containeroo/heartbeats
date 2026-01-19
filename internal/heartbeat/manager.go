package heartbeat

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/containeroo/heartbeats/internal/utils"
)

// Manager routes HTTP pings to Actors.
type Manager struct {
	actors  map[string]*managedActor
	logger  *slog.Logger
	baseCtx context.Context
	factory ActorFactory
	started bool
	mu      sync.RWMutex
}

type managedActor struct {
	actor   *Actor
	cfg     HeartbeatConfig
	started bool
	cancel  context.CancelFunc
}

// NewManagerFromHeartbeatMap creates a Manager from heartbeat config without starting actors.
func NewManagerFromHeartbeatMap(
	ctx context.Context,
	heartbeatConfigs HeartbeatConfigMap,
	logger *slog.Logger,
	factory ActorFactory,
) (*Manager, error) {
	if factory == nil {
		return nil, fmt.Errorf("manager requires an actor factory")
	}
	m := &Manager{
		actors:  make(map[string]*managedActor, len(heartbeatConfigs)),
		logger:  logger,
		baseCtx: ctx,
		factory: factory,
	}
	for id, hb := range heartbeatConfigs {
		hb.ID = utils.DefaultIfZero(hb.ID, id)

		actor, err := factory.Build(hb)
		if err != nil {
			return nil, fmt.Errorf("build heartbeat %q: %w", id, err)
		}

		m.actors[id] = &managedActor{actor: actor, cfg: hb}
	}
	return m, nil
}

// StartAll starts all actors that are not running yet.
func (m *Manager) StartAll() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	started := 0
	for _, ma := range m.actors {
		if ma.started {
			continue
		}
		ctx, cancel := context.WithCancel(m.baseCtx)
		ma.cancel = cancel
		ma.started = true
		go ma.actor.Run(ctx)
		started++
		m.logger.Debug("started heartbeat", "id", ma.actor.ID)
	}
	if started > 0 {
		m.started = true
	}
	return started
}

// List returns all configured heartbeats.
func (m *Manager) List() map[string]*Actor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*Actor, len(m.actors))
	for id, ma := range m.actors {
		result[id] = ma.actor
	}
	return result
}

// Get returns one heartbeat’s info by ID.
func (m *Manager) Get(id string) *Actor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ma, ok := m.actors[id]
	if !ok {
		return nil
	}
	return ma.actor
}

// Reconcile updates the manager with a new config set without rebuilding everything.
func (m *Manager) Reconcile(heartbeatConfigs HeartbeatConfigMap) (ReconcileResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var res ReconcileResult

	for id, ma := range m.actors {
		if _, ok := heartbeatConfigs[id]; !ok {
			m.stopManaged(ma)
			delete(m.actors, id)
			res.Removed++
		}
	}

	for id, hb := range heartbeatConfigs {
		if hb.ID == "" {
			hb.ID = id
		}
		if ma, ok := m.actors[id]; ok {
			if sameConfig(ma.actor, hb) {
				ma.cfg = hb
				continue
			}

			actor, err := m.factory.Build(hb)
			if err != nil {
				return res, fmt.Errorf("build heartbeat %q: %w", id, err)
			}

			m.stopManaged(ma)
			newActor := &managedActor{actor: actor, cfg: hb}
			if m.started {
				ctx, cancel := context.WithCancel(m.baseCtx)
				newActor.cancel = cancel
				newActor.started = true
				go actor.Run(ctx)
			}
			m.actors[id] = newActor
			res.Updated++
			continue
		}

		actor, err := m.factory.Build(hb)
		if err != nil {
			return res, fmt.Errorf("build heartbeat %q: %w", id, err)
		}

		newActor := &managedActor{actor: actor, cfg: hb}
		if m.started {
			ctx, cancel := context.WithCancel(m.baseCtx)
			newActor.cancel = cancel
			newActor.started = true
			go actor.Run(ctx)
		}
		m.actors[id] = newActor
		res.Added++
	}

	return res, nil
}

// ReconcileResult reports how many actors were added, updated, or removed.
type ReconcileResult struct {
	Added   int
	Updated int
	Removed int
}

// Receive notifies the Actor with a heartbeat "receive" event.
func (m *Manager) Receive(id string) error {
	m.mu.RLock()
	a, ok := m.actors[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("heartbeat ID %q not found", id)
	}
	a.actor.Mailbox() <- EventReceive
	return nil
}

// Fail notifies the Actor with a heartbeat "fail" event.
func (m *Manager) Fail(id string) error {
	m.mu.RLock()
	a, ok := m.actors[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("heartbeat ID %q not found", id)
	}
	a.actor.Mailbox() <- EventFail
	return nil
}

// Test sends a test notification event to the Actor.
func (m *Manager) Test(id string) error {
	m.mu.RLock()
	a, ok := m.actors[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("heartbeat ID %q not found", id)
	}
	a.actor.Mailbox() <- EventTest
	return nil
}

func (m *Manager) stopManaged(ma *managedActor) {
	if ma == nil || !ma.started || ma.cancel == nil {
		return
	}
	ma.cancel()
	ma.cancel = nil
	ma.started = false
}

func sameConfig(a *Actor, cfg HeartbeatConfig) bool {
	if a == nil {
		return false
	}
	if a.ID != cfg.ID {
		return false
	}
	if a.Description != cfg.Description {
		return false
	}
	if a.Interval != cfg.Interval {
		return false
	}
	if a.Grace != cfg.Grace {
		return false
	}
	return slices.Equal(a.Receivers, cfg.Receivers)
}
