package bump

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notifier"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
	"github.com/stretchr/testify/assert"
)

func setupManager(t *testing.T, hist history.Store, hbName string) (*heartbeat.Manager, *servicehistory.Recorder) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	store := notifier.NewReceiverStore()
	recorder := servicehistory.NewRecorder(hist)
	disp := notifier.NewDispatcher(store, logger, recorder, 1, 1, 10, nil)

	cfg := heartbeat.HeartbeatConfigMap{
		hbName: {
			ID:          hbName,
			Description: "desc",
			Interval:    time.Second,
			Grace:       time.Second,
			Receivers:   []string{"r1"},
		},
	}
	return heartbeat.NewManagerFromHeartbeatMap(context.Background(), cfg, disp.Mailbox(), recorder, logger, nil), recorder
}

func findEventByType(t *testing.T, events []history.Event, h history.EventType) *history.Event {
	t.Helper()

	for i := range events {
		if events[i].Type == h {
			return &events[i]
		}
	}
	return nil
}

func TestReceive(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	mgr, recorder := setupManager(t, hist, "hb1")

	err := Receive(context.Background(), mgr, recorder, logger, "hb1", "1.2.3.4:5678", "GET", "Go-test")
	assert.NoError(t, err)

	events := hist.ListByID("hb1")
	ev := findEventByType(t, events, history.EventTypeHeartbeatReceived)
	if assert.NotNil(t, ev) {
		var meta history.RequestMetadataPayload
		assert.NoError(t, json.Unmarshal(ev.RawPayload, &meta))
		assert.Equal(t, "GET", meta.Method)
		assert.Equal(t, "1.2.3.4:5678", meta.Source)
		assert.Equal(t, "Go-test", meta.UserAgent)
	}
}

func TestReceiveUnknownHeartbeat(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	mgr, recorder := setupManager(t, hist, "hb1")

	err := Receive(context.Background(), mgr, recorder, logger, "missing", "1.2.3.4:5678", "GET", "Go-test")
	assert.ErrorIs(t, err, ErrUnknownHeartbeat)
}

func TestReceiveAppendError(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	expectedErr := errors.New("append failed")
	mockStore := &history.MockStore{
		RecordEventFunc: func(ctx context.Context, e history.Event) error {
			return expectedErr
		},
	}
	mgr, recorder := setupManager(t, mockStore, "hb1")

	err := Receive(context.Background(), mgr, recorder, logger, "hb1", "1.2.3.4:5678", "GET", "Go-test")
	assert.ErrorIs(t, err, expectedErr)
}

func TestFail(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	mgr, recorder := setupManager(t, hist, "hb1")

	err := Fail(context.Background(), mgr, recorder, logger, "hb1", "1.2.3.4:5678", "POST", "Go-test")
	assert.NoError(t, err)

	events := hist.ListByID("hb1")
	ev := findEventByType(t, events, history.EventTypeHeartbeatFailed)
	if assert.NotNil(t, ev) {
		var meta history.RequestMetadataPayload
		assert.NoError(t, json.Unmarshal(ev.RawPayload, &meta))
		assert.Equal(t, "POST", meta.Method)
		assert.Equal(t, "1.2.3.4:5678", meta.Source)
		assert.Equal(t, "Go-test", meta.UserAgent)
	}
}

func TestFailUnknownHeartbeat(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	hist := history.NewRingStore(10)
	mgr, recorder := setupManager(t, hist, "hb1")

	err := Fail(context.Background(), mgr, recorder, logger, "missing", "1.2.3.4:5678", "POST", "Go-test")
	assert.ErrorIs(t, err, ErrUnknownHeartbeat)
}

func TestFailAppendError(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	expectedErr := errors.New("append failed")
	mockStore := &history.MockStore{
		RecordEventFunc: func(ctx context.Context, e history.Event) error {
			return expectedErr
		},
	}
	mgr, recorder := setupManager(t, mockStore, "hb1")

	err := Fail(context.Background(), mgr, recorder, logger, "hb1", "1.2.3.4:5678", "POST", "Go-test")
	assert.ErrorIs(t, err, expectedErr)
}
