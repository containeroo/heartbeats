package notifier

import (
	"context"
	"errors"
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
		store := NewReceiverStore()
		store.Register("r1", n)
		hist := history.NewRingStore(10)
		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))

		// Buffer size = 1 for test
		dispatcher := NewDispatcher(store, logger, hist, 1, 1*time.Millisecond, 1)

		// Start dispatcher loop
		ctx := t.Context()
		go dispatcher.Run(ctx)

		data := NotificationData{
			Receivers: []string{"r1"},
			Message:   "hello",
		}

		// Send data via mailbox
		dispatcher.Mailbox() <- data

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

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		store := NewReceiverStore()
		hist := history.NewRingStore(10)
		dispatcher := NewDispatcher(store, logger, hist, 1, 1*time.Millisecond, 1)

		// Start dispatcher loop
		ctx := t.Context()
		go dispatcher.Run(ctx)

		data := NotificationData{
			Receivers: []string{"nonexistent"},
			Message:   "should warn",
		}

		// Send data via mailbox
		dispatcher.Mailbox() <- data

		// Let it settle
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("record error", func(t *testing.T) {
		t.Parallel()

		var buf strings.Builder
		logger := slog.New(slog.NewTextHandler(&buf, nil))

		n := &MockNotifier{}
		store := NewReceiverStore()
		store.Register("r1", n)

		mockHist := history.MockStore{
			RecordEventFunc: func(ctx context.Context, e history.Event) error {
				return errors.New("fail!")
			},
		}

		dispatcher := NewDispatcher(store, logger, &mockHist, 1, 1*time.Millisecond, 1)

		// Start dispatcher loop
		ctx := t.Context()
		go dispatcher.Run(ctx)

		data := NotificationData{
			Receivers: []string{"r1"},
			Message:   "hello",
		}

		// Send data via mailbox
		dispatcher.Mailbox() <- data

		// Let it settle
		time.Sleep(10 * time.Millisecond)

		assert.Contains(t, buf.String(), "level=ERROR msg=\"failed to record state change\" err=fail!\n")
	})
}

func TestDispatcher_ListAndGet(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))

	store := NewReceiverStore()
	n1 := &MockNotifier{}
	n2 := &MockNotifier{}

	store.Register("a", n1)
	store.Register("a", n2)
	store.Register("b", n1)

	hist := history.NewRingStore(10)

	// Added buffer size (doesn't matter for this test since we don't use the mailbox)
	dispatcher := NewDispatcher(store, logger, hist, 1, 1*time.Millisecond, 1)

	t.Run("lists all receivers", func(t *testing.T) {
		t.Parallel()

		list := dispatcher.List()
		assert.Len(t, list, 2)
		assert.Len(t, list["a"], 2)
		assert.Len(t, list["b"], 1)
	})

	t.Run("gets all notifiers for a receiver", func(t *testing.T) {
		t.Parallel()

		get := dispatcher.Get("a")
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

	store := NewReceiverStore()
	store.Register("receiver1", &MockNotifier{
		NotifyFunc: func(ctx context.Context, data NotificationData) error {
			return fmt.Errorf("fail!")
		},
	})

	var logBuf strings.Builder
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	dispatcher := NewDispatcher(
		store,
		logger,
		&mockHistory{},
		1,
		1*time.Millisecond,
		1, // buffer size
	)

	// Start dispatcher loop
	ctx := t.Context()
	go dispatcher.Run(ctx)

	// Push notification
	data := NotificationData{Receivers: []string{"receiver1"}}
	dispatcher.Mailbox() <- data

	// Allow goroutine to process
	time.Sleep(10 * time.Millisecond)

	assert.Contains(t, logBuf.String(), "notification error")
}
