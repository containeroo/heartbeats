package metrics

import (
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/prometheus/client_golang/prometheus"
)

// heartbeats_heartbeat_last_status
const (
	DOWN float64 = 0
	UP   float64 = 1
)

// heartbeats_receiver_last_status
const (
	SUCCESS float64 = 0
	ERROR   float64 = 1
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

	// ReceiverErrorStatus reports the status of the last notification attempt (1 = ERROR, 0 = SUCCESS)
	ReceiverErrorStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heartbeats_receiver_last_status",
			Help: "Reports the status of the last notification attempt (1 = ERROR, 0 = SUCCESS)",
		},
		[]string{"receiver", "type", "target"},
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
	reg.MustRegister(ReceiverErrorStatus)

	// Register history store metrics
	reg.MustRegister(history.NewHistoryMetrics(store))

	return &Metrics{
		Registry: reg,
	}
}
