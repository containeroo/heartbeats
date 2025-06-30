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
	// LastStatus reports the most recent status of each heartbeat (0 = DOWN, 1 = UP).
	LastStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heartbeats_heartbeat_last_status",
			Help: "Most recent status of each heartbeat (0 = DOWN, 1 = UP)",
		},
		[]string{"heartbeat"},
	)

	// ReceivedTotal counts the total number of received heartbeats per ID.
	ReceivedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heartbeats_heartbeat_received_total",
			Help: "Total number of received heartbeats per ID",
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
	reg.MustRegister(LastStatus)
	reg.MustRegister(ReceivedTotal)

	// Register history store metrics
	reg.MustRegister(history.NewHistoryMetrics(store))

	return &Metrics{
		Registry: reg,
	}
}
