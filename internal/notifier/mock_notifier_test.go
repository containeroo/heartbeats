package notifier

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockNotifier(t *testing.T) {
	t.Parallel()

	t.Run("Notify sets called and stores last input", func(t *testing.T) {
		t.Parallel()

		mock := &MockNotifier{}
		input := NotificationData{ID: "abc"}

		err := mock.Notify(context.Background(), input)
		assert.NoError(t, err)

		assert.True(t, mock.called)
		assert.Equal(t, input, mock.last)
	})

	t.Run("Notify uses custom NotifyFunc if set", func(t *testing.T) {
		t.Parallel()

		mock := &MockNotifier{
			NotifyFunc: func(ctx context.Context, data NotificationData) error {
				return fmt.Errorf("fail!")
			},
		}

		err := mock.Notify(context.Background(), NotificationData{})
		assert.EqualError(t, err, "fail!")
	})

	t.Run("Format returns original data if no override", func(t *testing.T) {
		t.Parallel()

		mock := &MockNotifier{}
		in := NotificationData{ID: "x"}
		out, err := mock.Format(in)

		assert.NoError(t, err)
		assert.Equal(t, in, out)
	})

	t.Run("Format uses FormatFunc if set", func(t *testing.T) {
		t.Parallel()

		mock := &MockNotifier{
			FormatFunc: func(data NotificationData) (NotificationData, error) {
				data.Title = "set"
				return data, nil
			},
		}
		out, err := mock.Format(NotificationData{ID: "x"})

		assert.NoError(t, err)
		assert.Equal(t, "set", out.Title)
	})

	t.Run("Type returns custom type name", func(t *testing.T) {
		t.Parallel()

		mock := &MockNotifier{TypeName: "test-mock"}

		assert.Equal(t, "test-mock", mock.Type())
	})

	t.Run("Type returns default when unset", func(t *testing.T) {
		t.Parallel()

		mock := &MockNotifier{}

		assert.Equal(t, "mock", mock.Type())
	})

	t.Run("LastSent and LastErr return configured values", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		mock := &MockNotifier{
			Sent:    now,
			lastErr: fmt.Errorf("some error"),
		}

		assert.Equal(t, now, mock.LastSent())
		assert.EqualError(t, mock.LastErr(), "some error")
	})
}

func TestMockValidator(t *testing.T) {
	t.Parallel()

	mock := &MockNotifier{}

	assert.Nil(t, mock.Validate())
}

func TestMockResolver(t *testing.T) {
	t.Parallel()

	mock := &MockNotifier{}

	assert.Nil(t, mock.Resolve())
}
