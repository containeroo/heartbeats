package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat/service"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeService struct {
	updateErr error
	updated   []string
}

func (f *fakeService) HeartbeatSummaries() []service.HeartbeatSummary { return nil }
func (f *fakeService) ReceiverSummaries() []service.ReceiverSummary   { return nil }
func (f *fakeService) StatusAll() []service.Status                    { return nil }
func (f *fakeService) StatusByID(id string) (service.Status, error)   { return service.Status{}, nil }

func (f *fakeService) Update(id string, payload string, now time.Time) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updated = append(f.updated, id)
	return nil
}

func newHeartbeatAPI(svc ServiceProvider, metricsReg *metrics.Registry) *API {
	api := NewAPI(
		"test",
		"test",
		"http://example.com",
		slog.New(slog.NewTextHandler(&strings.Builder{}, nil)),
	)
	api.SetService(svc)
	api.SetMetrics(metricsReg)
	return api
}

func bump(t *testing.T, api *API, id string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", "/api/heartbeat/"+id, strings.NewReader("payload"))
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	api.Heartbeat().ServeHTTP(rec, req)
	return rec
}

func TestHeartbeatHandler(t *testing.T) {
	t.Parallel()

	t.Run("accepted bump increments received counter", func(t *testing.T) {
		t.Parallel()

		svc := &fakeService{}
		metricsReg := metrics.NewRegistry()
		api := newHeartbeatAPI(svc, metricsReg)

		rec := bump(t, api, "api")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, []string{"api"}, svc.updated)

		metricsRec := httptest.NewRecorder()
		api.Metrics().ServeHTTP(metricsRec, httptest.NewRequest("GET", "/metrics", nil))
		assert.Contains(t,
			metricsRec.Body.String(),
			`heartbeats_heartbeat_received_total{heartbeat="api"} 1`,
			"Expected an accepted bump to increment 'heartbeats_heartbeat_received_total'",
		)
	})

	t.Run("rejected bump does not increment received counter", func(t *testing.T) {
		t.Parallel()

		svc := &fakeService{updateErr: errors.New("heartbeat \"api\" not found")}
		metricsReg := metrics.NewRegistry()
		api := newHeartbeatAPI(svc, metricsReg)

		rec := bump(t, api, "api")
		require.Equal(t, http.StatusNotFound, rec.Code)

		metricsRec := httptest.NewRecorder()
		api.Metrics().ServeHTTP(metricsRec, httptest.NewRequest("GET", "/metrics", nil))
		assert.NotContains(t,
			metricsRec.Body.String(),
			"heartbeats_heartbeat_received_total{",
			"Expected a rejected bump to leave 'heartbeats_heartbeat_received_total' untouched",
		)
	})

	t.Run("missing id is rejected", func(t *testing.T) {
		t.Parallel()

		svc := &fakeService{}
		api := newHeartbeatAPI(svc, metrics.NewRegistry())

		rec := bump(t, api, "")
		require.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Empty(t, svc.updated)
	})
}
