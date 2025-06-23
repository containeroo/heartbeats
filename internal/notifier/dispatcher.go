package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/heartbeats/internal/common"
	"github.com/containeroo/heartbeats/internal/history"
)

// Dispatcher handles queued notifications via mailbox.
type Dispatcher struct {
	store   *ReceiverStore        // maps receiver IDs → []Notifier
	history history.Store         // stores notifications
	logger  *slog.Logger          // logs any notify errors
	retries int                   // number of retries for notifications
	delay   time.Duration         // delay between retries
	mailbox chan NotificationData // channel for receiving notifications
}

// NewDispatcher returns a Dispatcher backed by the given store.
func NewDispatcher(
	store *ReceiverStore,
	logger *slog.Logger,
	hist history.Store,
	retries int,
	delay time.Duration,
	bufferSize int,
) *Dispatcher {
	return &Dispatcher{
		store:   store,
		logger:  logger,
		history: hist,
		retries: retries,
		delay:   delay,
		mailbox: make(chan NotificationData, bufferSize),
	}
}

// Dispatch looks up each receiver’s Notifiers and fires them in parallel,
// recording the outcome in d.states.
// Run processes NotificationData from the mailbox.
func (d *Dispatcher) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-d.mailbox:
			d.dispatch(ctx, data)
		}
	}
}

// Mailbox returns the channel to send NotificationData.
func (d *Dispatcher) Mailbox() chan<- NotificationData {
	return d.mailbox
}

// dispatch looks up receivers and sends notifications in parallel.
func (d *Dispatcher) dispatch(ctx context.Context, data NotificationData) {
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

// List returns all configured notifiers.
func (d *Dispatcher) List() map[string][]Notifier {
	return d.store.notifiers
}

// Get returns notifiers for a receiver.
func (d *Dispatcher) Get(id string) []Notifier {
	return d.store.notifiers[id]
}

// sendWithRetry retries a notification and records its outcome.
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

// retryNotify retries sending notification.
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
