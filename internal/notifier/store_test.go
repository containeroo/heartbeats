package notifier

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReceiverStore_addAndGet(t *testing.T) {
	t.Parallel()

	t.Run("add and retrieve mock notifier", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		mock := &MockNotifier{
			TypeName: "mock",
			Sent:     time.Now(),
			lastErr:  nil,
		}

		store.addNotifier("r1", mock)

		result := store.getNotifiers("r1")
		assert.Len(t, result, 1)
		assert.Equal(t, mock, result[0])
	})
}

func TestInitializeStore(t *testing.T) {
	t.Parallel()

	t.Run("empty config returns empty store", func(t *testing.T) {
		t.Parallel()

		var logBuffer strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		store := InitializeStore(nil, false, logger)
		assert.Empty(t, store.getNotifiers("any"))
	})

	t.Run("single receiver with all notifier types", func(t *testing.T) {
		t.Parallel()

		var logBuffer strings.Builder
		logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

		cfg := map[string]ReceiverConfig{
			"r1": {
				SlackConfigs: []SlackConfig{
					{Channel: "#general"},
				},
				EmailConfigs: []EmailConfig{
					{EmailDetails: EmailDetails{To: []string{"user@example.com"}}},
				},
				MSTeamsConfigs: []MSTeamsConfig{
					{WebhookURL: "https://teams.example.com"},
				},
			},
		}

		store := InitializeStore(cfg, false, logger)

		notifiers := store.getNotifiers("r1")
		assert.Len(t, notifiers, 3)

		assert.IsType(t, &SlackConfig{}, notifiers[0])
		assert.IsType(t, &EmailConfig{}, notifiers[1])
		assert.IsType(t, &MSTeamsConfig{}, notifiers[2])
	})

	t.Run("multiple receivers are isolated", func(t *testing.T) {
		t.Parallel()

		cfg := map[string]ReceiverConfig{
			"r1": {
				SlackConfigs: []SlackConfig{{Channel: "#a"}},
			},
			"r2": {
				EmailConfigs: []EmailConfig{{EmailDetails: EmailDetails{To: []string{"x"}}}},
			},
		}

		store := InitializeStore(cfg, false, slog.Default())

		r1 := store.getNotifiers("r1")
		r2 := store.getNotifiers("r2")

		assert.Len(t, r1, 1)
		assert.Len(t, r2, 1)
		assert.IsType(t, &SlackConfig{}, r1[0])
		assert.IsType(t, &EmailConfig{}, r2[0])
	})
}
