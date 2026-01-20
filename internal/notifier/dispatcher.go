package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/metrics"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
)

// Dispatcher handles queued notifications via mailbox.
type Dispatcher struct {
	store   *ReceiverStore           // maps receiver IDs → []Notifier
	storeMu sync.RWMutex             // guards store
	history *servicehistory.Recorder // stores notifications
	logger  *slog.Logger             // logs any notify errors
	retries int                      // number of retries for notifications
	delay   time.Duration            // delay between retries
	mailbox chan NotificationData    // channel for receiving notifications
	metrics *metrics.Registry        // metrics registry
}

// NewDispatcher returns a Dispatcher backed by the given store.
func NewDispatcher(
	store *ReceiverStore,
	logger *slog.Logger,
	hist *servicehistory.Recorder,
	retries int,
	delay time.Duration,
	bufferSize int,
	metricsReg *metrics.Registry,
) *Dispatcher {
	return &Dispatcher{
		store:   store,
		logger:  logger,
		history: hist,
		retries: retries,
		delay:   delay,
		mailbox: make(chan NotificationData, bufferSize),
		metrics: metricsReg,
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
	d.storeMu.RLock()
	store := d.store
	d.storeMu.RUnlock()

	for _, rid := range data.Receivers {
		notifiers := store.getNotifiers(rid)
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
	d.storeMu.RLock()
	defer d.storeMu.RUnlock()
	return d.store.notifiers
}

// Get returns notifiers for a receiver.
func (d *Dispatcher) Get(id string) []Notifier {
	d.storeMu.RLock()
	defer d.storeMu.RUnlock()
	return d.store.notifiers[id]
}

// UpdateStore swaps the receiver store used by the dispatcher.
func (d *Dispatcher) UpdateStore(store *ReceiverStore) {
	d.storeMu.Lock()
	d.store = store
	d.storeMu.Unlock()
}

// sendWithRetry retries a notification and records its outcome.
func (d *Dispatcher) sendWithRetry(ctx context.Context, receiverID string, n Notifier, data NotificationData) {
	factory := servicehistory.NewFactory()
	payload := servicehistory.NotificationPayload{
		Receiver: receiverID,
		Type:     n.Type(),
		Target:   n.Target(),
	}

	eventType := servicehistory.EventTypeNotificationSent
	receiverStatus := metrics.SUCCESS

	formatted := data
	if f, ok := n.(Formatter); ok {
		var err error
		formatted, err = f.Format(data)
		if err != nil {
			payload.Error = err.Error()

			logging.SystemLogger(d.logger, nil).Error("notification format error",
				"receiver", receiverID,
				"type", n.Type(),
				"target", n.Target(),
				"error", err,
			)

			d.metrics.SetReceiverStatus(receiverID, n.Type(), n.Target(), metrics.ERROR)

			ev := factory.NotificationFailed(data.ID, payload.Receiver, payload.Type, payload.Target, payload.Error)
			if err := d.history.Append(ctx, ev); err != nil {
				logging.SystemLogger(d.logger, nil).Error("failed to record state change", "err", err)
			}
			return
		}
	}

	if err := d.retryNotify(ctx, n, formatted, receiverID); err != nil {
		payload.Error = err.Error()
		eventType = servicehistory.EventTypeNotificationFailed

		receiverStatus = metrics.ERROR

		logging.SystemLogger(d.logger, nil).Error("notification error",
			"receiver", receiverID,
			"type", n.Type(),
			"target", n.Target(),
			"error", err,
		)
	}

	d.metrics.SetReceiverStatus(receiverID, n.Type(), n.Target(), receiverStatus)

	var ev servicehistory.Event
	if eventType == servicehistory.EventTypeNotificationFailed {
		ev = factory.NotificationFailed(data.ID, payload.Receiver, payload.Type, payload.Target, payload.Error)
	} else {
		ev = factory.NotificationSent(data.ID, payload.Receiver, payload.Type, payload.Target)
	}
	if err := d.history.Append(ctx, ev); err != nil {
		logging.SystemLogger(d.logger, nil).Error("failed to record state change", "err", err)
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

		logging.SystemLogger(d.logger, nil).Debug("retrying", "attempt", i+1, "receiver", receiverID)
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
