package metrics

import (
	"fmt"
	"strings"
	"testing"

	"github.com/containeroo/heartbeats/internal/history"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	t.Parallel()

	hist := history.NewRingStore(10)
	PromMetrics := NewMetrics(hist)
	assert.NotNil(t, PromMetrics.Registry, "Registry should not be nil")

	// Initialize metrics
	LastStatus.With(prometheus.Labels{"heartbeat": "test_heartbeat"})
	ReceivedTotal.With(prometheus.Labels{"heartbeat": "test_heartbeat"})

	gatherers := prometheus.Gatherers{
		PromMetrics.Registry,
	}

	mfs, err := gatherers.Gather()
	assert.NoError(t, err, "Expected no error while gathering metrics")

	var foundLastStatus, foundReceivedTotal, foundHistorySize bool

	for _, mf := range mfs {
		if *mf.Name == "heartbeats_heartbeat_last_status" {
			foundLastStatus = true
		}
		if *mf.Name == "heartbeats_heartbeat_received_total" {
			foundReceivedTotal = true
		}
		if *mf.Name == "heartbeats_history_byte_size" {
			foundHistorySize = true
		}
	}

	assert.True(t, foundLastStatus, "Expected to find heartbeats_heartbeat_last_status metric")
	assert.True(t, foundReceivedTotal, "Expected to find heartbeats_heartbeats_total metric")
	assert.True(t, foundHistorySize, "Expected to find heartbeats_history_byte_size metric")
}

func TestLastStatusMetric(t *testing.T) {
	t.Parallel()

	hist := history.NewRingStore(10)
	promMetrics := NewMetrics(hist)

	LastStatus.With(prometheus.Labels{"heartbeat": "test_heartbeat"}).Set(UP)

	expected := `
	# HELP heartbeats_heartbeat_last_status Most recent status of each heartbeat (0 = DOWN, 1 = UP)
	# TYPE heartbeats_heartbeat_last_status gauge
	heartbeats_heartbeat_last_status{heartbeat="test_heartbeat"} 1
	`
	err := testutil.GatherAndCompare(promMetrics.Registry, strings.NewReader(expected), "heartbeats_heartbeat_last_status")
	assert.NoError(t, err, "Expected no error while gathering and comparing metrics")
}

func TestReceivedTotalMetric(t *testing.T) {
	t.Parallel()

	hist := history.NewRingStore(10)
	promMetrics := NewMetrics(hist)

	ReceivedTotal.With(prometheus.Labels{"heartbeat": "test_heartbeat"}).Inc()

	expected := `
  # HELP heartbeats_heartbeat_received_total Total number of received heartbeats per ID
  # TYPE heartbeats_heartbeat_received_total counter
  heartbeats_heartbeat_received_total{heartbeat="test_heartbeat"} 1

	`
	err := testutil.GatherAndCompare(promMetrics.Registry, strings.NewReader(expected), "heartbeats_heartbeat_received_total")
	assert.NoError(t, err, "Expected no error while gathering and comparing metrics")
}

func TestHistorySizeMetric(t *testing.T) {
	t.Parallel()

	size := 10_000
	hist := history.NewRingStore(size)
	promMetrics := NewMetrics(hist)
	ctx := t.Context()

	payload := history.RequestMetadataPayload{
		Source:    "http://localhost:9090",
		Method:    "GET",
		UserAgent: "go-test",
	}

	for range size {
		ev := history.MustNewEvent(history.EventTypeHeartbeatReceived, "test_heartbeat", payload)
		err := hist.Append(ctx, ev)
		assert.NoError(t, err)
	}

	got := float64(hist.ByteSize())

	expected := fmt.Sprintf(`
  # HELP heartbeats_history_byte_size Current size of the history store in bytes
  # TYPE heartbeats_history_byte_size gauge
  heartbeats_history_byte_size %f
  `, got)

	err := testutil.GatherAndCompare(promMetrics.Registry, strings.NewReader(expected), "heartbeats_history_byte_size")
	assert.NoError(t, err, "Expected no error while gathering and comparing metrics")

	assert.Equal(t, float64(1.83e+06), got)
}
