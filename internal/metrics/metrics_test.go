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
	HeartbeatStatus.With(prometheus.Labels{"heartbeat": "test_heartbeat"})
	TotalHeartbeats.With(prometheus.Labels{"heartbeat": "test_heartbeat"})

	gatherers := prometheus.Gatherers{
		PromMetrics.Registry,
	}

	mfs, err := gatherers.Gather()
	assert.NoError(t, err, "Expected no error while gathering metrics")

	var foundHeartbeatStatus, foundTotalHeartbeats, foundHistorySize bool

	for _, mf := range mfs {
		if *mf.Name == "heartbeats_heartbeat_last_status" {
			foundHeartbeatStatus = true
		}
		if *mf.Name == "heartbeats_heartbeats_total" {
			foundTotalHeartbeats = true
		}
		if *mf.Name == "heartbeats_history_byte_size" {
			foundHistorySize = true
		}
	}

	assert.True(t, foundHeartbeatStatus, "Expected to find heartbeats_heartbeat_last_status metric")
	assert.True(t, foundTotalHeartbeats, "Expected to find heartbeats_heartbeats_total metric")
	assert.True(t, foundHistorySize, "Expected to find heartbeats_history_byte_size metric")
}

func TestHeartbeatStatusMetric(t *testing.T) {
	t.Parallel()

	hist := history.NewRingStore(10)
	promMetrics := NewMetrics(hist)

	HeartbeatStatus.With(prometheus.Labels{"heartbeat": "test_heartbeat"}).Set(UP)

	expected := `
	# HELP heartbeats_heartbeat_last_status Total number of heartbeats
	# TYPE heartbeats_heartbeat_last_status gauge
	heartbeats_heartbeat_last_status{heartbeat="test_heartbeat"} 1
	`
	err := testutil.GatherAndCompare(promMetrics.Registry, strings.NewReader(expected), "heartbeats_heartbeat_last_status")
	assert.NoError(t, err, "Expected no error while gathering and comparing metrics")
}

func TestTotalHeartbeatsMetric(t *testing.T) {
	t.Parallel()

	hist := history.NewRingStore(10)
	promMetrics := NewMetrics(hist)

	TotalHeartbeats.With(prometheus.Labels{"heartbeat": "test_heartbeat"}).Inc()

	expected := `
	# HELP heartbeats_heartbeats_total The total number of heartbeats
	# TYPE heartbeats_heartbeats_total counter
	heartbeats_heartbeats_total{heartbeat="test_heartbeat"} 1
	`
	err := testutil.GatherAndCompare(promMetrics.Registry, strings.NewReader(expected), "heartbeats_heartbeats_total")
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
		err := hist.RecordEvent(ctx, ev)
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

	assert.Greater(t, got, float64(1.898883e+05), "ByteSize should be reasonably large")
	assert.Less(t, got, float64(1.898883e+07), "ByteSize should be within expected upper bound")
}
