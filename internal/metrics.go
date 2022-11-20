package internal

import (
	"github.com/prometheus/client_golang/prometheus"
)

var PromMetrics Metrics

type Metrics struct {
	HeartbeatStatus *prometheus.GaugeVec
	TotalHeartbeats *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		HeartbeatStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "heartbeats_heartbeat_last_status",
				Help: "Total number of heartbeats",
			},
			[]string{"heartbeat"},
		),
		TotalHeartbeats: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "heartbeats_heartbeats_total",
				Help: "The total number of heartbeats",
			},
			[]string{"heartbeat"},
		),
	}
	reg.MustRegister(m.TotalHeartbeats)
	reg.MustRegister(m.HeartbeatStatus)
	return m
}
