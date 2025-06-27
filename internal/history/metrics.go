package history

import (
	"github.com/prometheus/client_golang/prometheus"
)

// historyMetrics collects metrics related to the history store.
type historyMetrics struct {
	store Store
	size  *prometheus.Desc // Metric describing byte size of the event history
}

// NewHistoryMetrics returns a Prometheus collector that tracks the size of the history store.
func NewHistoryMetrics(store Store) prometheus.Collector {
	return &historyMetrics{
		store: store,
		size: prometheus.NewDesc(
			"heartbeats_history_byte_size",               // Metric name
			"Current size of the history store in bytes", // Help text
			nil, nil,
		),
	}
}

// Describe sends the descriptor for the history size metric.
func (h *historyMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- h.size
}

// Collect gathers the current size of the history store and sends it as a gauge metric.
func (h *historyMetrics) Collect(ch chan<- prometheus.Metric) {
	size := h.store.ByteSize()
	ch <- prometheus.MustNewConstMetric(h.size, prometheus.GaugeValue, float64(size))
}
