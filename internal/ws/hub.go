package ws

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/containeroo/heartbeats/internal/heartbeat/service"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
)

var wsjsonWrite = wsjson.Write

// Providers supply snapshot data for websocket clients.
type Providers struct {
	Heartbeats    func() []service.HeartbeatSummary
	Receivers     func() []service.ReceiverSummary
	History       func() []history.Event
	HeartbeatByID func(string) (service.HeartbeatSummary, bool)
	ReceiverByKey func(string, string, string) (service.ReceiverSummary, bool)
}

type message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// Hub manages websocket clients and broadcasts events.
type Hub struct {
	logger    *slog.Logger
	providers Providers
	mu        sync.Mutex
	clients   map[*websocket.Conn]struct{}
}

// NewHub constructs a websocket hub.
func NewHub(logger *slog.Logger, providers Providers) *Hub {
	return &Hub{
		logger:    logger,
		providers: providers,
		clients:   make(map[*websocket.Conn]struct{}),
	}
}

// Start subscribes to history updates and broadcasts them.
func (h *Hub) Start(ctx context.Context, subscribe func(int) (<-chan history.Event, func())) {
	if subscribe == nil {
		return
	}
	stream, cancel := subscribe(256)
	if stream == nil {
		return
	}
	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-stream:
				if !ok {
					return
				}
				h.PublishEvent(ev)
			}
		}
	}()
}

// Handle upgrades the connection and keeps it registered until closed.
func (h *Hub) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		logging.AccessLogger(h.logger).Error("ws accept failed", "event", "ws_accept_failed", "err", err)
		return
	}

	h.add(conn)
	defer h.remove(conn)

	h.sendSnapshot(r.Context(), conn)

	ctx := r.Context()
	for {
		if _, _, err := conn.Read(ctx); err != nil {
			return
		}
	}
}

// PublishEvent broadcasts a history event and derived updates.
func (h *Hub) PublishEvent(ev history.Event) {
	h.publishHistory(ev)
	h.publishHeartbeatUpdate(ev)
	h.publishReceiverUpdate(ev)
}

// publishHistory broadcasts a history event.
func (h *Hub) publishHistory(ev history.Event) {
	h.publish("history", ev)
}

// publish broadcasts a message to all connected clients.
func (h *Hub) publish(kind string, data any) {
	msg := message{
		Type: kind,
		Data: data,
	}

	conns := h.snapshot()
	for _, conn := range conns {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := wsjsonWrite(ctx, conn, msg)
		cancel()
		if err != nil {
			h.remove(conn)
			_ = conn.Close(websocket.StatusNormalClosure, "")
		}
	}
}

// sendSnapshot sends a snapshot of the current state to the client.
func (h *Hub) sendSnapshot(ctx context.Context, conn *websocket.Conn) {
	if h.providers.Heartbeats == nil && h.providers.Receivers == nil && h.providers.History == nil {
		return
	}

	snapshot := struct {
		Heartbeats []service.HeartbeatSummary `json:"heartbeats,omitempty"`
		Receivers  []service.ReceiverSummary  `json:"receivers,omitempty"`
		History    []history.Event            `json:"history,omitempty"`
	}{
		Heartbeats: h.safeHeartbeats(),
		Receivers:  h.safeReceivers(),
		History:    h.safeHistory(),
	}

	if err := wsjsonWrite(ctx, conn, message{Type: "snapshot", Data: snapshot}); err != nil {
		h.remove(conn)
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}
}

// publishHeartbeatUpdate broadcasts a heartbeat update.
func (h *Hub) publishHeartbeatUpdate(ev history.Event) {
	if ev.HeartbeatID == "" {
		return
	}
	if h.providers.HeartbeatByID == nil {
		return
	}

	hb, ok := h.providers.HeartbeatByID(ev.HeartbeatID)
	if !ok {
		hb = service.HeartbeatSummary{
			ID:         ev.HeartbeatID,
			HasHistory: true,
		}
	}
	h.publish("heartbeat", hb)
}

// publishReceiverUpdate broadcasts a receiver update.
func (h *Hub) publishReceiverUpdate(ev history.Event) {
	if ev.Receiver == "" || ev.TargetType == "" {
		return
	}
	if h.providers.ReceiverByKey == nil {
		return
	}

	target, _ := ev.Fields["target"].(string)
	receiver, ok := h.providers.ReceiverByKey(ev.Receiver, ev.TargetType, target)
	if !ok {
		receiver = service.ReceiverSummary{
			ID:          ev.Receiver,
			Type:        ev.TargetType,
			Destination: target,
		}
	}
	h.publish("receiver", receiver)
}

// safeHeartbeats returns the heartbeats from the providers.
func (h *Hub) safeHeartbeats() []service.HeartbeatSummary {
	if h.providers.Heartbeats == nil {
		return nil
	}
	return h.providers.Heartbeats()
}

// safeReceivers returns the receivers from the providers.
func (h *Hub) safeReceivers() []service.ReceiverSummary {
	if h.providers.Receivers == nil {
		return nil
	}
	return h.providers.Receivers()
}

// safeHistory returns the history from the providers.
func (h *Hub) safeHistory() []history.Event {
	if h.providers.History == nil {
		return nil
	}
	return h.providers.History()
}

// add registers a client connection.
func (h *Hub) add(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

// remove unregisters a client connection.
func (h *Hub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

// snapshot returns a snapshot of all connected clients.
func (h *Hub) snapshot() []*websocket.Conn {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := make([]*websocket.Conn, 0, len(h.clients))
	for conn := range h.clients {
		conns = append(conns, conn)
	}
	return conns
}
