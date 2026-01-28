package dispatch

import (
	"context"
	"log/slog"

	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/notify/types"
)

// Dispatcher dequeues notifications and delivers them.
type Dispatcher struct {
	store      *Store
	mailbox    <-chan string
	delivery   types.Delivery
	registry   map[string]map[string]*types.Receiver
	heartbeats map[string]types.HeartbeatMeta
	logger     *slog.Logger
}

// NewDispatcher constructs a Dispatcher with the provided mailbox.
func NewDispatcher(
	store *Store,
	mailbox <-chan string,
	delivery types.Delivery,
	registry map[string]map[string]*types.Receiver,
	heartbeats map[string]types.HeartbeatMeta,
	logger *slog.Logger,
) *Dispatcher {
	return &Dispatcher{
		store:      store,
		mailbox:    mailbox,
		delivery:   delivery,
		registry:   registry,
		heartbeats: heartbeats,
		logger:     logger,
	}
}

// Start begins processing notifications until ctx is canceled.
func (d *Dispatcher) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case id, ok := <-d.mailbox:
			if !ok {
				return
			}
			d.dispatch(ctx, id)
		}
	}
}

// dispatch delivers a notification.
func (d *Dispatcher) dispatch(ctx context.Context, id string) {
	n, ok := d.store.Get(id)
	if !ok {
		d.logger.Warn("Notification not found",
			"event", logging.EventNotificationMissing.String(),
			"id", id)
		return
	}
	d.store.Delete(id)
	meta, ok := d.heartbeats[n.HeartbeatID()]
	if !ok {
		d.logger.Warn("Heartbeat metadata not found",
			"event", logging.EventHeartbeatMetadataMissing.String(),
			"id", id,
			"heartbeat_id", n.HeartbeatID())
		return
	}
	payload := types.Payload{
		HeartbeatID: n.HeartbeatID(),
		Title:       meta.Title,
		Status:      n.Status(),
		Payload:     n.Payload(),
		Timestamp:   n.Timestamp(),
		Interval:    meta.Interval,
		LateAfter:   meta.LateAfter,
		Since:       n.Since(),
	}
	receiverNames := n.ReceiverNames()
	receiverMap := d.registry[payload.HeartbeatID]
	receivers := make([]*types.Receiver, 0, len(receiverNames))

	for _, name := range receiverNames {
		rcv, ok := receiverMap[name]
		if !ok {
			d.logger.Warn("Receiver not found",
				"event", logging.EventReceiverMissing.String(),
				"id", id,
				"receiver", name)
			continue
		}
		receivers = append(receivers, rcv)
	}
	if len(receivers) == 0 {
		d.logger.Warn("No receivers resolved",
			"event", logging.EventReceiverEmpty.String(),
			"id", id,
			"heartbeat_id", payload.HeartbeatID)
		return
	}

	if err := d.delivery.Dispatch(ctx, payload, receivers); err != nil {
		d.logger.Error("Notification delivery failed",
			"event", logging.EventNotificationDeliveryFailed.String(),
			"id", id,
			"error", err)
		return
	}
	d.logger.Info("Notification delivered",
		"event", logging.EventNotificationDelivered.String(),
		"id", id)
}
