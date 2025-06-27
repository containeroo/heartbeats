package metrics

import (
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/prometheus/client_golang/prometheus"
)

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
func NewMetrics(store history.Store) *Metrics {
	reg := prometheus.NewRegistry()

	// Register core metrics
	reg.MustRegister(HeartbeatStatus)
	reg.MustRegister(TotalHeartbeats)

	// Register history store metrics
	reg.MustRegister(history.NewHistoryMetrics(store))

	return &Metrics{
		Registry: reg,
	}
}
