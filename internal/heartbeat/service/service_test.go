package service

import (
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/heartbeat/types"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/notify/targets"
	ntypes "github.com/containeroo/heartbeats/internal/notify/types"
	"github.com/containeroo/heartbeats/internal/runner"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	s map[string]*types.Heartbeat
}

func newFakeStore() *fakeStore {
	return &fakeStore{s: map[string]*types.Heartbeat{}}
}

func (f *fakeStore) Get(id string) (*types.Heartbeat, bool) {
	hb, ok := f.s[id]
	return hb, ok
}

func (f *fakeStore) All() []*types.Heartbeat {
	out := make([]*types.Heartbeat, 0, len(f.s))
	for _, hb := range f.s {
		out = append(out, hb)
	}
	return out
}

func newHeartbeat(t *testing.T, id string, interval, lateAfter time.Duration, stage runner.Stage, lastSeen time.Time) *types.Heartbeat {
	t.Helper()
	state := runner.NewState()
	if !lastSeen.IsZero() {
		require.True(t, state.UpdateSeen(lastSeen, "payload"))
		select {
		case <-state.Mailbox():
		default:
		}
	}
	switch stage {
	case runner.StageOK:
		state.MarkOK()
	case runner.StageLate:
		state.MarkLate()
	case runner.StageMissing:
		state.MarkMissing(lastSeen)
	default:
		state.MarkOK()
	}
	return &types.Heartbeat{
		ID:        id,
		Title:     "heartbeat-" + id,
		State:     state,
		Config:    config.HeartbeatConfig{Interval: interval, LateAfter: lateAfter},
		Receivers: []string{"ops"},
	}
}

func newReceiverStore() ReceiverStore {
	return &fakeReceiverStore{
		receivers: []*ntypes.Receiver{
			{
				Name: "ops",
				Targets: []ntypes.Target{
					&targets.WebhookTarget{URL: "https://example.com"},
				},
			},
		},
	}
}

type fakeReceiverStore struct {
	receivers []*ntypes.Receiver
}

func (f *fakeReceiverStore) Receivers() []*ntypes.Receiver { return f.receivers }

func newTestService(t *testing.T) (*Service, *fakeStore, *history.Store) {
	t.Helper()
	store := newFakeStore()
	hist := history.NewStore(10)
	svc := NewService(store, newReceiverStore(), hist)
	return svc, store, hist
}

func TestServiceUpdate(t *testing.T) {
	t.Parallel()

	svc, store, hist := newTestService(t)
	hb := newHeartbeat(t, "api", time.Second, 2*time.Second, runner.StageLate, time.Now())
	store.s["api"] = hb

	require.NoError(t, svc.Update("api", "payload", time.Now()))
	err := svc.Update("api", "payload", time.Now())
	require.Error(t, err)
	require.Contains(t, err.Error(), "heartbeat mailbox full")

	events := hist.List()
	require.Len(t, events, 2)
	require.Equal(t, false, events[1].Fields["enqueued"])
}

func TestServiceStatus(t *testing.T) {
	t.Parallel()

	svc, store, _ := newTestService(t)
	lastSeen := time.Now().Add(-time.Minute)
	store.s["api"] = newHeartbeat(t, "api", 5*time.Second, 10*time.Second, runner.StageOK, lastSeen)

	status, err := svc.StatusByID("api")
	require.NoError(t, err)
	require.Equal(t, "api", status.ID)
	require.Equal(t, runner.StageOK, status.Stage)
	require.Greater(t, status.SinceSeen, time.Duration(0))

	all := svc.StatusAll()
	require.Len(t, all, 1)
}

func TestHeartbeatSummaries(t *testing.T) {
	t.Parallel()

	svc, store, hist := newTestService(t)
	lastSeen := time.Now().Truncate(time.Second)
	hb := newHeartbeat(t, "api", 5*time.Second, 10*time.Second, runner.StageMissing, lastSeen)
	store.s["api"] = hb
	hist.Add(history.Event{HeartbeatID: "api"})

	summaries := svc.HeartbeatSummaries()
	require.Len(t, summaries, 1)
	require.Equal(t, "missing", summaries[0].Status)
	require.Equal(t, lastSeen.UTC().Format(time.RFC3339Nano), summaries[0].LastBump)
}

func TestReceiverSummaries(t *testing.T) {
	t.Parallel()

	svc, _, hist := newTestService(t)
	failed := history.Event{
		Receiver:   "ops",
		TargetType: ntypes.TargetWebhook.String(),
		Fields: map[string]any{
			"target": "https://example.com",
		},
		Type:    history.EventNotificationFailed.String(),
		Message: "boom",
		Time:    time.Now(),
	}
	hist.Add(failed)

	summaries := svc.ReceiverSummaries()
	require.Len(t, summaries, 1)
	require.Equal(t, "ops", summaries[0].ID)
	require.Equal(t, "https://example.com", summaries[0].Destination)
	require.Equal(t, "boom", summaries[0].LastErr)

	receiver, ok := svc.ReceiverSummaryByKey("ops", ntypes.TargetWebhook.String(), "https://example.com")
	require.True(t, ok)
	require.Equal(t, "ops", receiver.ID)
}
