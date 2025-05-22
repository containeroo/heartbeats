package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	PromMetrics = NewMetrics()
	assert.NotNil(t, PromMetrics.Registry, "Registry should not be nil")

	// Initialize metrics
	HeartbeatStatus.With(prometheus.Labels{"heartbeat": "test_heartbeat"})
	TotalHeartbeats.With(prometheus.Labels{"heartbeat": "test_heartbeat"})

	gatherers := prometheus.Gatherers{
		PromMetrics.Registry,
	}

	mfs, err := gatherers.Gather()
	assert.NoError(t, err, "Expected no error while gathering metrics")

	foundHeartbeatStatus := false
	foundTotalHeartbeats := false

	for _, mf := range mfs {
		if *mf.Name == "heartbeats_heartbeat_last_status" {
			foundHeartbeatStatus = true
		}
		if *mf.Name == "heartbeats_heartbeats_total" {
			foundTotalHeartbeats = true
		}
	}

	assert.True(t, foundHeartbeatStatus, "Expected to find heartbeats_heartbeat_last_status metric")
	assert.True(t, foundTotalHeartbeats, "Expected to find heartbeats_heartbeats_total metric")
}

func TestHeartbeatStatusMetric(t *testing.T) {
	PromMetrics = NewMetrics()

	HeartbeatStatus.With(prometheus.Labels{"heartbeat": "test_heartbeat"}).Set(UP)

	expected := `
	# HELP heartbeats_heartbeat_last_status Total number of heartbeats
	# TYPE heartbeats_heartbeat_last_status gauge
	heartbeats_heartbeat_last_status{heartbeat="test_heartbeat"} 1
	`
	err := testutil.GatherAndCompare(PromMetrics.Registry, strings.NewReader(expected), "heartbeats_heartbeat_last_status")
	assert.NoError(t, err, "Expected no error while gathering and comparing metrics")
}

func TestTotalHeartbeatsMetric(t *testing.T) {
	PromMetrics = NewMetrics()

	TotalHeartbeats.With(prometheus.Labels{"heartbeat": "test_heartbeat"}).Inc()

	expected := `
	# HELP heartbeats_heartbeats_total The total number of heartbeats
	# TYPE heartbeats_heartbeats_total counter
	heartbeats_heartbeats_total{heartbeat="test_heartbeat"} 1
	`
	err := testutil.GatherAndCompare(PromMetrics.Registry, strings.NewReader(expected), "heartbeats_heartbeats_total")
	assert.NoError(t, err, "Expected no error while gathering and comparing metrics")
}
