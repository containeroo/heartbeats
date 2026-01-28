package history

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/containeroo/heartbeats/internal/logging"
)

const defaultHistoryBuffer = 256

// AsyncRecorder records history events asynchronously.
type AsyncRecorder struct {
	logger   *slog.Logger
	recorder Recorder
	ch       chan Event
	mu       sync.Mutex
	subs     map[int]chan Event
	nextSub  int
}

// NewAsyncRecorder wraps a Recorder with a buffered async queue.
func NewAsyncRecorder(recorder Recorder, logger *slog.Logger, buffer int) *AsyncRecorder {
	if buffer <= 0 {
		buffer = defaultHistoryBuffer
	}
	return &AsyncRecorder{
		logger:   logger,
		recorder: recorder,
		ch:       make(chan Event, buffer),
	}
}

// Start begins draining the async queue until ctx is canceled.
func (a *AsyncRecorder) Start(ctx context.Context) {
	if a == nil || a.recorder == nil {
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-a.ch:
				a.recorder.Add(e)
			}
		}
	}()
}

// Add enqueues an event without blocking.
func (a *AsyncRecorder) Add(e Event) {
	if a == nil || a.recorder == nil {
		return
	}
	if e.Time.IsZero() {
		e.Time = time.Now().UTC()
	}
	select {
	case a.ch <- e:
	default:
		a.logger.Debug("History buffer full; dropping event",
			"event", logging.EventHistoryDropped.String(),
			"type", e.Type,
		)
	}
	a.broadcast(e)
}

// List returns a snapshot of recorded events.
func (a *AsyncRecorder) List() []Event {
	if a == nil || a.recorder == nil {
		return nil
	}
	return a.recorder.List()
}

// ListByID returns recorded events filtered by heartbeat id.
func (a *AsyncRecorder) ListByID(heartbeatID string) []Event {
	if a == nil || a.recorder == nil {
		return nil
	}
	return a.recorder.ListByID(heartbeatID)
}

// Subscribe registers a buffered event stream.
func (a *AsyncRecorder) Subscribe(buffer int) (<-chan Event, func()) {
	if a == nil {
		return nil, func() {}
	}
	if buffer <= 0 {
		buffer = defaultHistoryBuffer
	}
	ch := make(chan Event, buffer)
	a.mu.Lock()
	if a.subs == nil {
		a.subs = make(map[int]chan Event)
	}
	id := a.nextSub
	a.nextSub++
	a.subs[id] = ch
	a.mu.Unlock()

	cancel := func() {
		a.mu.Lock()
		if sub, ok := a.subs[id]; ok {
			delete(a.subs, id)
			close(sub)
		}
		a.mu.Unlock()
	}
	return ch, cancel
}

func (a *AsyncRecorder) broadcast(e Event) {
	a.mu.Lock()
	if len(a.subs) == 0 {
		a.mu.Unlock()
		return
	}
	subs := make([]chan Event, 0, len(a.subs))
	for _, ch := range a.subs {
		subs = append(subs, ch)
	}
	a.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- e:
		default:
		}
	}
}
