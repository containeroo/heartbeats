package metrics

import (
	"net/http"

	"github.com/containeroo/heartbeats/internal/history"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// Registry holds Prometheus metrics for the app.
type Registry struct {
	registry           *prometheus.Registry
	lastStatus         *prometheus.GaugeVec
	receivedTotal      *prometheus.CounterVec
	receiverLastStatus *prometheus.GaugeVec
}

// New builds a new Prometheus metrics registry.
func New(store history.Store) *Registry {
	lastStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "heartbeats_heartbeat_last_status",
			Help: "Most recent status of each heartbeat (0 = DOWN, 1 = UP)",
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
	reg.MustRegister(lastStatus, receivedTotal, receiverLastStatus)
	reg.MustRegister(history.NewHistoryMetrics(store))

	return &Registry{
		registry:           reg,
		lastStatus:         lastStatus,
		receivedTotal:      receivedTotal,
		receiverLastStatus: receiverLastStatus,
	}
}

// SetHeartbeatStatus updates the last status gauge for a heartbeat.
func (r *Registry) SetHeartbeatStatus(id string, status float64) {
	r.lastStatus.WithLabelValues(id).Set(status)
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
