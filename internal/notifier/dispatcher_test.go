package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/history"
	"github.com/stretchr/testify/assert"
)

func TestDispatcher_Dispatch(t *testing.T) {
	t.Parallel()

	t.Run("calls notifier for each receiver", func(t *testing.T) {
		t.Parallel()

		n := &MockNotifier{}
		store := newReceiverStore()
		store.addNotifier("r1", n)
		hist := history.NewRingStore(10)

		var logBuffer strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		dispatcher := NewDispatcher(store, logger, hist, 1, 1)

		data := NotificationData{
			Receivers: []string{"r1"},
			Message:   "hello",
		}

		ctx := context.Background()
		dispatcher.Dispatch(ctx, data)

		// wait for goroutines to finish
		assert.Eventually(t, func() bool {
			n.mu.Lock()
			defer n.mu.Unlock()
			return n.called
		}, time.Second, 10*time.Millisecond)

		n.mu.Lock()
		defer n.mu.Unlock()

		assert.Equal(t, "hello", n.last.Message)
	})

	t.Run("logs and skips unknown receiver", func(t *testing.T) {
		t.Parallel()

		var logBuffer strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		store := newReceiverStore()
		hist := history.NewRingStore(10)
		dispatcher := NewDispatcher(store, logger, hist, 1, 1)

		data := NotificationData{
			Receivers: []string{"nonexistent"},
			Message:   "should warn",
		}

		dispatcher.Dispatch(context.Background(), data)
	})
}

func TestDispatcher_ListAndGet(t *testing.T) {
	t.Parallel()

	var logBuffer strings.Builder
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

	store := newReceiverStore()
	n1 := &MockNotifier{}
	n2 := &MockNotifier{}

	store.addNotifier("a", n1)
	store.addNotifier("a", n2)
	store.addNotifier("b", n1)

	hist := history.NewRingStore(10)
	d := NewDispatcher(store, logger, hist, 1, 1)

	t.Run("lists all receivers", func(t *testing.T) {
		t.Parallel()

		list := d.List()
		assert.Len(t, list, 2)
		assert.Len(t, list["a"], 2)
		assert.Len(t, list["b"], 1)
	})
	t.Run("Gets all notifiers for a receiver", func(t *testing.T) {
		t.Parallel()

		get := d.Get("a")
		assert.Len(t, get, 2)
		assert.Equal(t, get[0], n1)
		assert.Equal(t, get[1], n2)
	})
}

type mockHistory struct{}

func (m *mockHistory) RecordEvent(ctx context.Context, e history.Event) error { return nil }
func (m *mockHistory) GetEvents() []history.Event                             { return nil }
func (m *mockHistory) GetEventsByID(id string) []history.Event                { return nil }

func TestDispatcher_LogsErrorFromNotifier(t *testing.T) {
	t.Parallel()

	store := newReceiverStore()
	store.addNotifier("receiver1", &MockNotifier{
		NotifyFunc: func(ctx context.Context, data NotificationData) error {
			return fmt.Errorf("fail!")
		},
	})

	var logBuffer strings.Builder
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

	d := &Dispatcher{
		store:   store,
		logger:  logger,
		retries: 1,
		delay:   0,
		history: &mockHistory{},
	}

	data := NotificationData{Receivers: []string{"receiver1"}}
	d.Dispatch(context.Background(), data)
	time.Sleep(10 * time.Millisecond) // allow goroutine to finish

	assert.Contains(t, logBuffer.String(), "notification error")
}
