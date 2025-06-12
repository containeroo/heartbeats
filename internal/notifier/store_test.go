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

		store := NewReceiverStore()

		mock := &MockNotifier{
			TypeName: "mock",
			Sent:     time.Now(),
			lastErr:  nil,
		}

		store.Register("r1", mock) // nolint:errcheck

		result := store.getNotifiers("r1")
		assert.Len(t, result, 1)
		assert.Equal(t, mock, result[0])
	})
}

func TestInitializeStore(t *testing.T) {
	t.Parallel()

	t.Run("empty config returns empty store", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		store := InitializeStore(nil, false, "0.0.0", logger)
		assert.Empty(t, store.getNotifiers("any"))
	})

	t.Run("single receiver with all notifier types", func(t *testing.T) {
		t.Parallel()

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
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
				MSTeamsGraphConfig: []MSTeamsGraphConfig{
					{
						Token:     "graph-token",
						TeamID:    "team-id",
						ChannelID: "channel-id",
					},
				},
			},
		}

		store := InitializeStore(cfg, false, "0.0.0", logger)

		notifiers := store.getNotifiers("r1")
		assert.Len(t, notifiers, 4)

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

		logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
		store := InitializeStore(cfg, false, "0.0.0", logger)

		r1 := store.getNotifiers("r1")
		r2 := store.getNotifiers("r2")

		assert.Len(t, r1, 1)
		assert.Len(t, r2, 1)
		assert.IsType(t, &SlackConfig{}, r1[0])
		assert.IsType(t, &EmailConfig{}, r2[0])
	})
}
