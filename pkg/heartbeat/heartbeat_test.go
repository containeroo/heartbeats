package heartbeat

import (
	"context"
	"heartbeats/pkg/history"
	"heartbeats/pkg/logger"
	"heartbeats/pkg/notify"
	"heartbeats/pkg/timer"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	store := NewStore()

	interval := time.Second * 3
	grace := time.Second * 5
	tm := timer.Timer{Interval: &interval}
	gr := timer.Timer{Interval: &grace}

	h := &Heartbeat{
		Name:     "test",
		Interval: &tm,
		Grace:    &gr,
	}

	t.Run("Add", func(t *testing.T) {
		err := store.Add("test", h)
		assert.NoError(t, err, "Expected no error when adding a heartbeat")
	})

	t.Run("Add duplicate", func(t *testing.T) {
		err := store.Add("test", h)
		assert.Error(t, err, "Expected error when adding a duplicate heartbeat")
	})

	t.Run("Get All", func(t *testing.T) {
		all := store.GetAll()
		assert.Equal(t, 1, len(all), "Expected one heartbeat in store")
	})

	t.Run("Get", func(t *testing.T) {
		retrieved := store.Get("test")
		assert.NotNil(t, retrieved, "Expected to retrieve the added heartbeat")
	})

	t.Run("Delete", func(t *testing.T) {
		store.Delete("test")
		retrieved := store.Get("test")
		assert.Nil(t, retrieved, "Expected heartbeat to be deleted")
	})
}

func TestHeartbeatTimers(t *testing.T) {
	log := logger.NewLogger(true)
	hist, err := history.NewHistory(20, 20)
	assert.NoError(t, err)

	ns := notify.NewStore()

	interval := time.Second * 2
	grace := time.Second * 2
	tm := timer.Timer{Interval: &interval}
	gr := timer.Timer{Interval: &grace}

	h := &Heartbeat{
		Name:     "test",
		Interval: &tm,
		Grace:    &gr,
	}

	ctx := context.Background()

	t.Run("StartInterval", func(t *testing.T) {
		h.StartInterval(ctx, log, ns, hist)
		assert.Equal(t, StatusOK.String(), h.Status, "Expected status to be OK after starting interval")
	})

	t.Run("Multiple StartTimer with sleep", func(t *testing.T) {
		h.StartInterval(ctx, log, ns, hist)
		assert.Equal(t, StatusOK.String(), h.Status, "Expected status to be OK after starting interval")
		time.Sleep(1 * time.Second)
		h.StartInterval(ctx, log, ns, hist)
		assert.Equal(t, StatusOK.String(), h.Status, "Expected status to be OK after starting interval")
		time.Sleep(1 * time.Second)
		h.StartInterval(ctx, log, ns, hist)
		assert.Equal(t, StatusOK.String(), h.Status, "Expected status to be OK after starting interval")
	})

	t.Run("Multiple StartTimer without sleep", func(t *testing.T) {
		h.StartInterval(ctx, log, ns, hist)
		assert.Equal(t, StatusOK.String(), h.Status, "Expected status to be OK after starting interval")
		h.StartInterval(ctx, log, ns, hist)
		assert.Equal(t, StatusOK.String(), h.Status, "Expected status to be OK after starting interval")
		h.StartInterval(ctx, log, ns, hist)
		assert.Equal(t, StatusOK.String(), h.Status, "Expected status to be OK after starting interval")
	})

	t.Run("GraceAfterInterval", func(t *testing.T) {
		h.StartInterval(ctx, log, ns, hist)
		time.Sleep(3 * time.Second) // wait for the interval to elapse
		assert.Equal(t, StatusGrace.String(), h.Status, "Expected status to be GRACE after interval elapsed")
	})

	t.Run("StartGrace", func(t *testing.T) {
		h.StartGrace(ctx, log, ns, hist)
		assert.Equal(t, StatusGrace.String(), h.Status, "Expected status to be GRACE after starting grace")
	})

	t.Run("GraceToNOK", func(t *testing.T) {
		h.StartGrace(ctx, log, ns, hist)
		time.Sleep(5 * time.Second) // wait for the grace to elapse
		assert.Equal(t, StatusNOK.String(), h.Status, "Expected status to be NOK after grace elapsed")
	})

	t.Run("StopTimer", func(t *testing.T) {
		h.StopTimers()
		assert.Nil(t, h.Interval.Timer, "Expected interval timer to be stopped")
		assert.Nil(t, h.Grace.Timer, "Expected grace timer to be stopped")
	})
}

func TestHeartbeatUpdateStatus(t *testing.T) {
	log := logger.NewLogger(true)
	hist, err := history.NewHistory(10, 2)
	assert.NoError(t, err)

	ns := notify.NewStore()

	interval := time.Second * 2
	grace := time.Second * 2
	tm := timer.Timer{Interval: &interval}
	gr := timer.Timer{Interval: &grace}

	h := &Heartbeat{
		Name:     "test",
		Interval: &tm,
		Grace:    &gr,
	}

	ctx := context.Background()

	// Test UpdateStatus
	t.Run("UpdateStatus", func(t *testing.T) {
		h.updateStatus(ctx, log, StatusOK, ns, hist)
		assert.Equal(t, StatusOK.String(), h.Status, "Expected status to be updated to OK")
		assert.False(t, h.LastPing.IsZero(), "Expected LastPing to be updated")
	})
}

func TestSendNotifications(t *testing.T) {
	log := logger.NewLogger(true)
	hist, err := history.NewHistory(10, 2)
	assert.NoError(t, err)

	ns := notify.NewStore()

	interval := time.Second * 2
	grace := time.Second * 2
	tm := timer.Timer{Interval: &interval}
	gr := timer.Timer{Interval: &grace}

	h := &Heartbeat{
		Name:          "test",
		Interval:      &tm,
		Grace:         &gr,
		Notifications: []string{"test-notification"},
	}

	ctx := context.Background()

	notification := &notify.Notification{
		Name:    "test-notification",
		Enabled: boolPtr(true),
	}

	_ = ns.Add("test-notification", notification)

	t.Run("SendNotifications", func(t *testing.T) {
		h.SendNotifications(ctx, log, ns, hist, false)
		// assert.NotEmpty(t, hist.GetAllEntries(), "Expected notifications to be sent and logged in history")
	})
}

func boolPtr(b bool) *bool {
	return &b
}
