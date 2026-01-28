package dispatch

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync/atomic"
	"time"

	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notify/types"
)

var eventCounter uint64

// Manager owns notification delivery infrastructure.
type Manager struct {
	store      *Store
	mailbox    chan string
	dispatcher *Dispatcher
	registry   map[string]map[string]*types.Receiver
	heartbeats map[string]types.HeartbeatMeta
}

// NewManager constructs a notification manager with a buffered mailbox.
func NewManager(
	logger *slog.Logger,
	historyStore history.Recorder,
	metricsReg *metrics.Registry,
) *Manager {
	store := NewStore()
	mailbox := make(chan string, 128)
	delivery := NewDelivery(logger, historyStore, metricsReg)
	registry := make(map[string]map[string]*types.Receiver)
	heartbeats := make(map[string]types.HeartbeatMeta)
	dispatcher := NewDispatcher(store, mailbox, delivery, registry, heartbeats, logger)
	return &Manager{
		store:      store,
		mailbox:    mailbox,
		dispatcher: dispatcher,
		registry:   registry,
		heartbeats: heartbeats,
	}
}

// Enqueue stores the notification and returns its id.
func (m *Manager) Enqueue(n types.Notification) string {
	id := nextEventID()
	m.store.Put(id, n)
	m.mailbox <- id
	return id
}

// Register stores receivers for a heartbeat id.
func (m *Manager) Register(heartbeatID string, receivers map[string]*types.Receiver) {
	if heartbeatID == "" || receivers == nil {
		return
	}
	m.registry[heartbeatID] = receivers
}

// RegisterHeartbeat stores metadata for a heartbeat id.
func (m *Manager) RegisterHeartbeat(heartbeatID string, meta types.HeartbeatMeta) {
	if heartbeatID == "" {
		return
	}
	m.heartbeats[heartbeatID] = meta
}

// ResetRegistries clears the receiver and heartbeat registries.
func (m *Manager) ResetRegistries() {
	if m == nil {
		return
	}
	for key := range m.registry {
		delete(m.registry, key)
	}
	for key := range m.heartbeats {
		delete(m.heartbeats, key)
	}
}

// Receivers returns the unique receivers registered across all heartbeats.
func (m *Manager) Receivers() []*types.Receiver {
	if m == nil {
		return nil
	}
	seen := make(map[string]*types.Receiver)
	for _, perHeartbeat := range m.registry {
		for name, rcv := range perHeartbeat {
			if rcv == nil {
				continue
			}
			if _, ok := seen[name]; !ok {
				seen[name] = rcv
			}
		}
	}
	out := make([]*types.Receiver, 0, len(seen))
	for _, rcv := range seen {
		out = append(out, rcv)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

// nextEventID generates a unique event id.
func nextEventID() string {
	seq := atomic.AddUint64(&eventCounter, 1)
	return fmt.Sprintf("%d-%d", time.Now().UTC().UnixNano(), seq)
}

// Start begins delivery processing.
func (m *Manager) Start(ctx context.Context) {
	if m.dispatcher == nil {
		return
	}
	go m.dispatcher.Start(ctx)
}
