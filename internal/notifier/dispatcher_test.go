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
		dispatcher := NewDispatcher(store, slog.Default(), hist, 1, 1)

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

		store := newReceiverStore()
		logger := slog.Default()

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

	store := newReceiverStore()
	n1 := &MockNotifier{}
	n2 := &MockNotifier{}

	store.addNotifier("a", n1)
	store.addNotifier("a", n2)
	store.addNotifier("b", n1)

	hist := history.NewRingStore(10)
	d := NewDispatcher(store, slog.Default(), hist, 1, 1)

	list := d.List()
	assert.Len(t, list, 2)
	assert.Len(t, list["a"], 2)
	assert.Len(t, list["b"], 1)

	get := d.Get("a")
	assert.Len(t, get, 2)
	assert.Equal(t, get[0], n1)
	assert.Equal(t, get[1], n2)
}

func TestDispatcher_LogsErrorFromNotifier(t *testing.T) {
	store := newReceiverStore()
	store.addNotifier("receiver1", &MockNotifier{
		NotifyFunc: func(ctx context.Context, data NotificationData) error {
			return fmt.Errorf("fail!")
		},
	})

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))

	d := &Dispatcher{store: store, logger: logger}
	data := NotificationData{Receivers: []string{"receiver1"}}

	d.Dispatch(context.Background(), data)
	time.Sleep(10 * time.Millisecond) // allow goroutine to finish
}
