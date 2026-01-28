package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// heartbeats_receiver_last_status
const (
	SUCCESS float64 = 0
	ERROR   float64 = 1
)

// heartbeats_heartbeat_last_state
const (
	HeartbeatOK        float64 = 0
	HeartbeatLate      float64 = 1
	HeartbeatMissing   float64 = 2
	HeartbeatRecovered float64 = 3
	HeartbeatNever     float64 = -1
)

// Registry holds Prometheus metrics for the app.
type Registry struct {
	registry           *prometheus.Registry
	lastState          *prometheus.GaugeVec
	receivedTotal      *prometheus.CounterVec
	receiverLastStatus *prometheus.GaugeVec
}

// NewRegistry builds a new Prometheus metrics registry.
func NewRegistry() *Registry {
	lastState := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heartbeats_heartbeat_last_state",
			Help: "Most recent state of each heartbeat (0 = ok, 1 = late, 2 = missing, 3 = recovered, -1 = never)",
		},
		[]string{"heartbeat"},
	)
	receivedTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "heartbeats_heartbeat_received_total",
			Help: "Total number of received heartbeats per ID",
		},
		[]string{"heartbeat"},
	)
	receiverLastStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heartbeats_receiver_last_status",
			Help: "Reports the status of the last notification attempt (1 = ERROR, 0 = SUCCESS)",
		},
		[]string{"receiver", "type", "target"},
	)

	reg := prometheus.NewRegistry()
	reg.MustRegister(lastState, receivedTotal, receiverLastStatus)

	return &Registry{
		registry:           reg,
		lastState:          lastState,
		receivedTotal:      receivedTotal,
		receiverLastStatus: receiverLastStatus,
	}
}

// SetHeartbeatState updates the last state gauge for a heartbeat.
func (r *Registry) SetHeartbeatState(id string, state string) {
	r.lastState.WithLabelValues(id).Set(heartbeatStateValue(state))
}

// IncHeartbeatReceived increments the receive counter for a heartbeat.
func (r *Registry) IncHeartbeatReceived(id string) {
	r.receivedTotal.WithLabelValues(id).Inc()
}

// SetReceiverStatus sets the receiver status gauge.
func (r *Registry) SetReceiverStatus(receiver, typ, target string, status float64) {
	r.receiverLastStatus.WithLabelValues(receiver, typ, target).Set(status)
}

// Metrics returns the Prometheus metrics handler.
func (r *Registry) Metrics() http.Handler {
	return promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{})
}

func heartbeatStateValue(state string) float64 {
	switch state {
	case "ok":
		return HeartbeatOK
	case "late":
		return HeartbeatLate
	case "missing":
		return HeartbeatMissing
	case "recovered":
		return HeartbeatRecovered
	case "never":
		return HeartbeatNever
	default:
		return HeartbeatNever
	}
}
