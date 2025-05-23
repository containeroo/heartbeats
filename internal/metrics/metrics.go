package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// PromMetrics holds the global instance of Metrics.
var PromMetrics = NewMetrics()

const (
	DOWN = iota
	UP
)

var (
	// HeartbeatStatus is a gauge metric representing the last status of each heartbeat.
	HeartbeatStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heartbeats_heartbeat_last_status",
			Help: "Total number of heartbeats",
		},
		[]string{"heartbeat"},
	)

	// TotalHeartbeats is a counter metric tracking the total number of heartbeats.
	TotalHeartbeats = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heartbeats_heartbeats_total",
			Help: "The total number of heartbeats",
		},
		[]string{"heartbeat"},
	)
)

// Metrics wraps a Prometheus registry and provides functionality for metric registration.
type Metrics struct {
	Registry *prometheus.Registry
}

// NewMetrics creates and initializes a new Metrics instance with all relevant metrics registered.
func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(HeartbeatStatus)
	reg.MustRegister(TotalHeartbeats)

	return &Metrics{
		Registry: reg,
	}
}
