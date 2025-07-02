package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/metrics"
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
func (d *Dispatcher) Mailbox() chan<- NotificationData { return d.mailbox }

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
	payload := history.NotificationPayload{
		Receiver: receiverID,
		Type:     n.Type(),
		Target:   n.Target(),
	}

	eventType := history.EventTypeNotificationSent
	receiverStatus := metrics.SUCCESS

	if err := d.retryNotify(ctx, n, data, receiverID); err != nil {
		payload.Error = err.Error()
		eventType = history.EventTypeNotificationFailed

		receiverStatus = metrics.ERROR

		d.logger.Error("notification error",
			"receiver", receiverID,
			"type", n.Type(),
			"target", n.Target(),
			"error", err,
		)
	}

	metrics.ReceiverErrorStatus.
		WithLabelValues(receiverID, n.Type(), n.Target()).
		Set(receiverStatus)

	ev := history.MustNewEvent(eventType, data.ID, payload)
	if err := d.history.Append(ctx, ev); err != nil {
		d.logger.Error("failed to record state change", "err", err)
	}
}

// retryNotify retries sending notification.
func (d *Dispatcher) retryNotify(
	ctx context.Context,
	n Notifier,
	data NotificationData,
	receiverID string,
) error {
	var err error
	for i := 0; i < d.retries; i++ {
		err = n.Notify(ctx, data)
		if err == nil {
			return nil // success
		}

		d.logger.Debug("retrying", "attempt", i+1, "receiver", receiverID)
		// Wait unless it's the last attempt
		if i < d.retries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err() // context cancelled
			case <-time.After(d.delay): // wait before next retry
			}
		}
	}
	return fmt.Errorf("notification failed after %d retries: %w", d.retries, err)
}
