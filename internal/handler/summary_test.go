package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat/service"
	"github.com/stretchr/testify/require"
)

type fakeSummaryService struct {
	hb []service.HeartbeatSummary
	rc []service.ReceiverSummary
}

func (f *fakeSummaryService) HeartbeatSummaries() []service.HeartbeatSummary {
	return f.hb
}

func (f *fakeSummaryService) ReceiverSummaries() []service.ReceiverSummary {
	return f.rc
}

func (f *fakeSummaryService) Update(id string, payload string, now time.Time) error {
	return nil
}

func (f *fakeSummaryService) StatusAll() []service.Status {
	return nil
}

func (f *fakeSummaryService) StatusByID(id string) (service.Status, error) {
	return service.Status{}, nil
}

func TestHeartbeatsSumHandler(t *testing.T) {
	resp := []service.HeartbeatSummary{
		{ID: "a", Status: "ok"},
	}
	api := &API{}
	api.SetService(&fakeSummaryService{hb: resp})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/heartbeats", nil)

	api.HeartbeatsSum().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var payload []service.HeartbeatSummary
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &payload))
	require.Equal(t, resp, payload)
}

func TestReceiversSumHandler(t *testing.T) {
	resp := []service.ReceiverSummary{
		{ID: "ops", Type: "webhook"},
	}
	api := &API{}
	api.SetService(&fakeSummaryService{rc: resp})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/receivers", nil)

	api.ReceiversSum().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var payload []service.ReceiverSummary
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &payload))
	require.Equal(t, resp, payload)
}
