package manager

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"reflect"
	"strings"
	"sync"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/heartbeat/sender"
	htypes "github.com/containeroo/heartbeats/internal/heartbeat/types"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/metrics"
	ntypes "github.com/containeroo/heartbeats/internal/notify/types"
	"github.com/containeroo/heartbeats/internal/runner"
	"github.com/containeroo/heartbeats/internal/templates"
	"github.com/containeroo/heartbeats/internal/utils"
)

var builtinWebhookTemplates = map[string]string{
	"default": "templates/default.tmpl",
	"slack":   "templates/slack.tmpl",
}

var builtinEmailTemplates = map[string]string{
	"default": "templates/email.tmpl",
	"email":   "templates/email.tmpl",
}

// Manager owns the configured heartbeats and notification mailbox.
type Manager struct {
	mu         sync.RWMutex
	heartbeats map[string]*htypes.Heartbeat
	cancels    map[string]context.CancelFunc
	notifier   ntypes.Notifier
	history    history.Recorder
	logger     *slog.Logger
	metrics    *metrics.Registry
}

// NewManager builds a Manager from config and templates.
func NewManager(
	cfg *config.Config,
	templateFS fs.FS,
	notifier ntypes.Notifier,
	historyStore history.Recorder,
	metricsReg *metrics.Registry,
	logger *slog.Logger,
) (*Manager, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	heartbeatMap, err := buildHeartbeatMap(cfg, templateFS, notifier, logger, nil)
	if err != nil {
		return nil, err
	}

	return &Manager{
		heartbeats: heartbeatMap,
		cancels:    make(map[string]context.CancelFunc),
		notifier:   notifier,
		history:    historyStore,
		logger:     logger,
		metrics:    metricsReg,
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

// Reload rebuilds heartbeat runners and receiver registrations.
func (m *Manager) Reload(ctx context.Context, cfg *config.Config, templateFS fs.FS) (ReloadResult, error) {
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

	if resetter, ok := m.notifier.(interface{ ResetRegistries() }); ok {
		resetter.ResetRegistries()
	}

	nextMap, err := buildHeartbeatMap(cfg, templateFS, m.notifier, m.logger, oldStates)
	if err != nil {
		return ReloadResult{}, err
	}

	result := diffHeartbeatSets(oldSnapshot, nextMap)

	m.StopAll()
	m.mu.Lock()
	m.heartbeats = nextMap
	m.mu.Unlock()
	m.StartAll(ctx)

	return result, nil
}

// startHeartbeat
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
	templateFS fs.FS,
	notifier ntypes.Notifier,
	logger *slog.Logger,
	states map[string]*runner.State,
) (map[string]*htypes.Heartbeat, error) {
	defaultWebhookTmpl, err := templates.LoadDefault(templateFS)
	if err != nil {
		return nil, fmt.Errorf("load default template: %w", err)
	}
	defaultTitleTmpl, err := templates.LoadStringFromFS(templateFS, "templates/title.tmpl")
	if err != nil {
		return nil, fmt.Errorf("load default title template: %w", err)
	}
	defaultEmailTmpl, err := templates.LoadStringFromFS(templateFS, "templates/email.tmpl")
	if err != nil {
		return nil, fmt.Errorf("load default email template: %w", err)
	}

	heartbeatMap := make(map[string]*htypes.Heartbeat, len(cfg.Heartbeats))
	for id, sc := range cfg.Heartbeats {
		title := strings.TrimSpace(sc.Title)
		if title == "" {
			title = id
		}
		receiverMap, receiverNames, err := buildReceivers(
			cfg,
			sc,
			defaultWebhookTmpl,
			defaultTitleTmpl,
			defaultEmailTmpl,
			templateFS,
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("heartbeat %q receivers: %w", id, err)
		}
		if registry, ok := notifier.(ntypes.ReceiverRegistry); ok {
			registry.Register(id, receiverMap)
		}
		if registry, ok := notifier.(ntypes.HeartbeatRegistry); ok {
			registry.RegisterHeartbeat(id, ntypes.HeartbeatMeta{
				Title:     title,
				Interval:  sc.Interval,
				LateAfter: sc.LateAfter,
			})
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
			Receivers:       receiverNames,
			State:           state,
			AlertOnLate:     *utils.DefaultIfZero(sc.AlertOnLate, utils.ToPtr(false)),
			AlertOnRecovery: *utils.DefaultIfZero(sc.AlertOnRecovery, utils.ToPtr(true)),
		}
	}

	return heartbeatMap, nil
}

// diffHeartbeatSets compares two heartbeat sets and returns
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
	return false
}
