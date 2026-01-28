package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestHeartbeatStateValue(t *testing.T) {
	tests := map[string]float64{
		"ok":        HeartbeatOK,
		"late":      HeartbeatLate,
		"missing":   HeartbeatMissing,
		"recovered": HeartbeatRecovered,
		"never":     HeartbeatNever,
		"unknown":   HeartbeatNever,
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			require.Equal(t, expected, heartbeatStateValue(input))
		})
	}
}

func TestRegistryMetrics(t *testing.T) {
	t.Parallel()
	reg := NewRegistry()
	t.Run("SetHeartbeatState", func(t *testing.T) {
		t.Parallel()
		reg.SetHeartbeatState("api", "missing")
		require.Equal(t, HeartbeatMissing, testutil.ToFloat64(reg.lastState.WithLabelValues("api")))
	})

	t.Run("IncHeartbeatReceived", func(t *testing.T) {
		t.Parallel()
		reg.IncHeartbeatReceived("api")
		require.Equal(t, float64(1), testutil.ToFloat64(reg.receivedTotal.WithLabelValues("api")))
	})
	t.Run("SetReceiverStatus", func(t *testing.T) {
		t.Parallel()
		reg.SetReceiverStatus("ops", "webhook", "https://example", ERROR)
		require.Equal(t, ERROR, testutil.ToFloat64(reg.receiverLastStatus.WithLabelValues("ops", "webhook", "https://example")))
	})
}

func TestRegistryMetricsHandler(t *testing.T) {
	reg := NewRegistry()
	reg.SetHeartbeatState("api", "ok")
	reg.IncHeartbeatReceived("api")
	reg.SetReceiverStatus("ops", "webhook", "https://example", SUCCESS)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	reg.Metrics().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Result().StatusCode)
	body := rec.Body.String()
	require.Contains(t, body, "heartbeats_heartbeat_last_state")
	require.Contains(t, body, "heartbeats_heartbeat_received_total")
	require.Contains(t, body, "heartbeats_receiver_last_status")
}
