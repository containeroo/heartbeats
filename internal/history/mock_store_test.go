package history

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockStore_RecordEvent(t *testing.T) {
	t.Parallel()

	t.Run("calls custom RecordEventFunc", func(t *testing.T) {
		t.Parallel()

		called := false
		mock := &MockStore{
			RecordEventFunc: func(ctx context.Context, e Event) error {
				called = true
				return nil
			},
		}

		err := mock.RecordEvent(context.Background(), Event{})
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("returns nil when RecordEventFunc is nil", func(t *testing.T) {
		t.Parallel()

		mock := &MockStore{}
		err := mock.RecordEvent(context.Background(), Event{})
		assert.NoError(t, err)
	})
}

func TestMockStore_GetEvents(t *testing.T) {
	t.Parallel()

	t.Run("calls custom GetEventsFunc", func(t *testing.T) {
		t.Parallel()

		want := []Event{{Timestamp: time.Now()}}
		mock := &MockStore{
			GetEventsFunc: func() []Event {
				return want
			},
		}

		got := mock.GetEvents()
		assert.Equal(t, want, got)
	})

	t.Run("returns nil when GetEventsFunc is nil", func(t *testing.T) {
		t.Parallel()

		mock := &MockStore{}
		got := mock.GetEvents()
		assert.Nil(t, got)
	})
}

func TestMockStore_GetEventsByID(t *testing.T) {
	t.Parallel()

	t.Run("calls custom GetEventsByIDFunc", func(t *testing.T) {
		t.Parallel()

		want := []Event{{HeartbeatID: "abc"}}
		mock := &MockStore{
			GetEventsByIDFunc: func(id string) []Event {
				if id == "abc" {
					return want
				}
				return nil
			},
		}

		got := mock.GetEventsByID("abc")
		assert.Equal(t, want, got)
	})

	t.Run("returns nil when GetEventsByIDFunc is nil", func(t *testing.T) {
		t.Parallel()

		mock := &MockStore{}
		got := mock.GetEventsByID("xyz")
		assert.Nil(t, got)
	})
}

func TestMockStore_ByteSize(t *testing.T) {
	t.Parallel()

	t.Run("calls custom ByteSizeFunc", func(t *testing.T) {
		t.Parallel()
		mock := &MockStore{
			ByteSizeFunc: func() int {
				return 1234
			},
		}
		assert.Equal(t, 1234, mock.ByteSize())
	})

	t.Run("returns 0 when ByteSizeFunc is nil", func(t *testing.T) {
		t.Parallel()
		mock := &MockStore{}
		assert.Equal(t, int(0), mock.ByteSize())
	})
}
