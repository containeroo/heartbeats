package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/internal/history"
)

// Dispatcher sends alerts to all Notifiers in the ReceiverStore.
type Dispatcher struct {
	store   *ReceiverStore // maps receiver IDs → []Notifier
	history history.Store  // stores notifications
	logger  *slog.Logger   // logs any notify errors
	retries int            // number of retries for notifications
	delay   time.Duration  // delay between retries
}

// NewDispatcher returns a Dispatcher backed by the given store.
func NewDispatcher(store *ReceiverStore, logger *slog.Logger, hist history.Store, retries int, delay time.Duration) *Dispatcher {
	return &Dispatcher{
		store:   store,
		logger:  logger,
		history: hist,
		retries: retries,
		delay:   delay,
	}
}

// Dispatch looks up each receiver’s Notifiers and fires them in parallel,
// recording the outcome in d.states.
func (d *Dispatcher) Dispatch(ctx context.Context, data NotificationData) {
	for _, rid := range data.Receivers {
		notifiers := d.store.getNotifiers(rid)
		if len(notifiers) == 0 {
			d.logger.Warn("no notifiers for receiver", "receiver", rid)
			continue
		}

		for _, n := range notifiers {
			// capture loop vars
			receiverID := rid
			notifier := n
			go d.sendWithRetry(ctx, receiverID, notifier, data)
		}
	}
}

// List returns all configured notifier.
func (d *Dispatcher) List() map[string][]Notifier {
	return d.store.notifiers
}

// Get returns one notifier’s info by ID.
func (d *Dispatcher) Get(id string) []Notifier {
	return d.store.notifiers[id]
}

// sendWithRetry retries a notification and records its outcome in the event history.
func (d *Dispatcher) sendWithRetry(ctx context.Context, receiverID string, n Notifier, data NotificationData) {
	info := NotificationInfo{Receiver: receiverID}
	eventType := history.EventTypeNotificationSent

	err := retryNotify(ctx, n, data, d.retries, d.delay, receiverID, d.logger)
	if err != nil {
		d.logger.Error("notification error", "receiver", receiverID, "error", err)
		info.Error = err
		eventType = history.EventTypeNotificationFailed
	}

	_ = d.history.RecordEvent(ctx, history.Event{
		Timestamp:   time.Now(),
		Type:        eventType,
		NewState:    common.HeartbeatStateRecovered.String(),
		HeartbeatID: data.ID,
		Payload:     info,
	})
}

// retryNotify tries sending a notification up to `retries` times with delay between attempts.
func retryNotify(
	ctx context.Context,
	notifier Notifier,
	data NotificationData,
	retries int,
	delay time.Duration,
	receiverID string,
	logger *slog.Logger,
) error {
	var err error
	for i := range retries {
		err = notifier.Notify(ctx, data)
		if err == nil {
			return nil // success
		}

		logger.Debug("retrying", "attempt", i+1, "receiver", receiverID)

		// Wait unless it's the last attempt
		if i < retries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err() // context cancelled
			case <-time.After(delay): // wait before next retry
			}
		}
	}
	return fmt.Errorf("notification failed after %d retries: %w", retries, err)
}
