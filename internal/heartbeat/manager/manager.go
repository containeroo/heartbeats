package manager

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"sync"

	kit "github.com/containeroo/notifykit/notify"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/heartbeat/sender"
	htypes "github.com/containeroo/heartbeats/internal/heartbeat/types"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notify"
	"github.com/containeroo/heartbeats/internal/runner"
	"github.com/containeroo/heartbeats/internal/utils"
)

// Manager owns the configured heartbeats and notification mailbox.
type Manager struct {
	mu         sync.RWMutex
	heartbeats map[string]*htypes.Heartbeat
	cancels    map[string]context.CancelFunc
	notifier   kit.Notifier
	history    history.Recorder
	logger     *slog.Logger
	metrics    *metrics.Registry
	routes     notify.ReceiverRoutes
}

// NewManager builds a Manager from config.
func NewManager(
	cfg *config.Config,
	notifier kit.Notifier,
	routes notify.ReceiverRoutes,
	historyStore history.Recorder,
	metricsReg *metrics.Registry,
	logger *slog.Logger,
) (*Manager, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	heartbeatMap := buildHeartbeatMap(cfg, routes, nil)

	return &Manager{
		heartbeats: heartbeatMap,
		cancels:    make(map[string]context.CancelFunc),
		notifier:   notifier,
		history:    historyStore,
		logger:     logger,
		metrics:    metricsReg,
		routes:     routes,
	}, nil
}

// StartAll launches runner loops for all configured heartbeats.
func (m *Manager) StartAll(ctx context.Context) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancels == nil {
		m.cancels = make(map[string]context.CancelFunc)
	}
	for _, hb := range m.heartbeats {
		m.startHeartbeat(ctx, hb)
	}
}

// StopAll stops all heartbeat runner loops.
func (m *Manager) StopAll() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, cancel := range m.cancels {
		if cancel != nil {
			cancel()
		}
		delete(m.cancels, id)
	}
}

// Get returns a heartbeat by id.
func (m *Manager) Get(id string) (*htypes.Heartbeat, bool) {
	if m == nil {
		return nil, false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	hb, ok := m.heartbeats[id]
	return hb, ok
}

// All returns all configured heartbeats.
func (m *Manager) All() []*htypes.Heartbeat {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*htypes.Heartbeat, 0, len(m.heartbeats))
	for _, hb := range m.heartbeats {
		out = append(out, hb)
	}
	return out
}

// ReloadResult reports heartbeat changes after a reload.
type ReloadResult struct {
	Added   int
	Updated int
	Removed int
}

// Reload rebuilds heartbeat runners and receiver routes.
func (m *Manager) Reload(ctx context.Context, cfg *config.Config, routes notify.ReceiverRoutes) (ReloadResult, error) {
	if m == nil {
		return ReloadResult{}, errors.New("manager is nil")
	}
	if cfg == nil {
		return ReloadResult{}, errors.New("config is nil")
	}

	m.mu.RLock()
	oldSnapshot := make(map[string]*htypes.Heartbeat, len(m.heartbeats))
	oldStates := make(map[string]*runner.State, len(m.heartbeats))
	for id, hb := range m.heartbeats {
		oldSnapshot[id] = hb
		if hb != nil && hb.State != nil {
			oldStates[id] = hb.State
		}
	}
	m.mu.RUnlock()

	nextMap := buildHeartbeatMap(cfg, routes, oldStates)
	result := diffHeartbeatSets(oldSnapshot, nextMap)

	m.StopAll()
	m.mu.Lock()
	m.heartbeats = nextMap
	m.routes = routes
	m.mu.Unlock()
	m.StartAll(ctx)

	return result, nil
}

// startHeartbeat starts a single heartbeat runner.
func (m *Manager) startHeartbeat(ctx context.Context, hb *htypes.Heartbeat) {
	if hb == nil {
		return
	}
	if cancel, ok := m.cancels[hb.ID]; ok {
		if cancel != nil {
			cancel()
		}
	}
	heartbeatCtx, cancel := context.WithCancel(ctx)
	m.cancels[hb.ID] = cancel

	sender := &sender.HeartbeatSender{
		Heartbeat: hb,
		Notifier:  m.notifier,
		History:   m.history,
		Logger:    m.logger,
		Metrics:   m.metrics,
	}
	go runner.Run(heartbeatCtx, hb.State, runner.Config{
		LateAfter:       hb.Config.LateAfter,
		CheckInterval:   hb.Config.Interval,
		AlertOnRecovery: hb.AlertOnRecovery,
		AlertOnLate:     hb.AlertOnLate,
	}, sender, m.logger)

	m.logger.Debug("Started heartbeat runner",
		"event", logging.EventHeartbeatStarted.String(),
		"heartbeat", hb.ID,
		"interval", hb.Config.Interval.String(),
		"late_after", hb.Config.LateAfter.String(),
	)
}

// buildHeartbeatMap builds a map of heartbeats from config.
func buildHeartbeatMap(
	cfg *config.Config,
	routes notify.ReceiverRoutes,
	states map[string]*runner.State,
) map[string]*htypes.Heartbeat {
	heartbeatMap := make(map[string]*htypes.Heartbeat, len(cfg.Heartbeats))
	for id, sc := range cfg.Heartbeats {
		title := strings.TrimSpace(sc.Title)
		if title == "" {
			title = id
		}

		state := runner.NewState()
		if states != nil {
			if existing := states[id]; existing != nil {
				state = existing
			}
		}
		heartbeatMap[id] = &htypes.Heartbeat{
			ID:              id,
			Title:           title,
			Config:          sc,
			Receivers:       append([]string(nil), sc.Receivers...),
			ReceiverIDs:     routes.ReceiverIDs(id),
			State:           state,
			AlertOnLate:     *utils.DefaultIfZero(sc.AlertOnLate, utils.ToPtr(false)),
			AlertOnRecovery: *utils.DefaultIfZero(sc.AlertOnRecovery, utils.ToPtr(true)),
		}
	}

	return heartbeatMap
}

// diffHeartbeatSets compares two heartbeat sets.
func diffHeartbeatSets(
	oldSet map[string]*htypes.Heartbeat,
	newSet map[string]*htypes.Heartbeat,
) ReloadResult {
	var result ReloadResult
	for id, hb := range newSet {
		prev, ok := oldSet[id]
		if !ok || prev == nil {
			result.Added++
			continue
		}
		if heartbeatChanged(prev, hb) {
			result.Updated++
		}
	}
	for id := range oldSet {
		if _, ok := newSet[id]; !ok {
			result.Removed++
		}
	}
	return result
}

// heartbeatChanged returns true if the heartbeats differ.
func heartbeatChanged(prev, next *htypes.Heartbeat) bool {
	if prev == nil || next == nil {
		return true
	}
	if prev.Title != next.Title {
		return true
	}
	if prev.AlertOnLate != next.AlertOnLate || prev.AlertOnRecovery != next.AlertOnRecovery {
		return true
	}
	if !reflect.DeepEqual(prev.Config, next.Config) {
		return true
	}
	if !reflect.DeepEqual(prev.Receivers, next.Receivers) {
		return true
	}
	if !reflect.DeepEqual(prev.ReceiverIDs, next.ReceiverIDs) {
		return true
	}
	return false
}
