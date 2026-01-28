package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/containeroo/heartbeats/internal/heartbeat/service"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/ws"
)

// ServiceProvider exposes the subset of the heartbeat service used by handlers.
type ServiceProvider interface {
	HeartbeatSummaries() []service.HeartbeatSummary
	ReceiverSummaries() []service.ReceiverSummary
	Update(id string, payload string, now time.Time) error
	StatusAll() []service.Status
	StatusByID(id string) (service.Status, error)
}

type websocketHub interface {
	Handle(http.ResponseWriter, *http.Request)
}

// API bundles shared handler dependencies and runtime configuration.
type API struct {
	Version  string            // Build version string.
	Commit   string            // Build commit string.
	SiteURL  string            // Site root URL.
	Logger   *slog.Logger      // Logger for access/system logs.
	service  ServiceProvider   // Domain service for heartbeat state.
	history  history.Recorder  // In-memory history recorder.
	metrics  *metrics.Registry // Prometheus metrics registry.
	wsHub    websocketHub      // websocket hub.
	reloadFn func() error      // reload function.
}

// NewAPI builds an API container with shared handler dependencies.
func NewAPI(
	version, commit string,
	siteURL string,
	logger *slog.Logger,
) *API {
	return &API{
		Version: version,
		Commit:  commit,
		SiteURL: siteURL,
		Logger:  logger,
	}
}

// SetHistory attaches a history recorder
func (a *API) SetHistory(h history.Recorder) {
	a.history = h
}

// SetMetrics attaches a metrics registry
func (a *API) SetMetrics(m *metrics.Registry) {
	a.metrics = m
}

// SetService attaches a heartbeat service
func (a *API) SetService(s ServiceProvider) {
	a.service = s
}

// HistoryRecorder returns the configured history recorder.
func (a *API) HistoryRecorder() history.Recorder {
	return a.history
}

// SetHub attaches a websocket hub to the API.
func (a *API) SetHub(hub *ws.Hub) {
	a.wsHub = hub
}

// SetReloadFn attaches a reload function to the API.
func (a *API) SetReloadFn(fn func() error) {
	a.reloadFn = fn
}

// respondJSON writes a JSON response and logs failures.
func (a *API) respondJSON(w http.ResponseWriter, status int, v any) {
	if err := encode(w, status, v); err != nil {
		a.Logger.Error("encode response failed", "event", logging.EventEncodeResponseFailed.String(), "err", err)
	}
}

// statusResponse is the standard success payload.
type statusResponse struct {
	Status string `json:"status"` // Status string.
}

// errorResponse is the standard error payload.
type errorResponse struct {
	Error string `json:"error"` // Error message.
}

// encode encodes a value to JSON and writes it to the response.
func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
