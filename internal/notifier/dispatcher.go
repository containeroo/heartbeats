package notifier

import (
	"context"
	"log/slog"
	"sync"
)

// Dispatcher sends alerts to all Notifiers in the ReceiverStore.
type Dispatcher struct {
	store  *ReceiverStore // maps receiver IDs → []Notifier
	logger *slog.Logger   // logs any notify errors
	mu     sync.Mutex     // protects states
}

// NewDispatcher returns a Dispatcher backed by the given store.
func NewDispatcher(store *ReceiverStore, logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		store:  store,
		logger: logger,
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
			n := n // capture
			go func() {
				d.mu.Lock()
				defer d.mu.Unlock()
				err := n.Notify(ctx, data)
				if err != nil {
					d.logger.Error("notification error", "receiver", rid, "error", err)
				}
			}()
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
