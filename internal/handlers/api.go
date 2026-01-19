package handlers

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notifier"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
)

// API bundles shared handler dependencies and runtime configuration.
type API struct {
	Version string // Version is the build version string.
	Commit  string // Commit is the build commit SHA.
	webFS   fs.FS
	Logger  *slog.Logger
	mgr     *heartbeat.Manager
	hist    history.Store
	rec     *servicehistory.Recorder
	disp    *notifier.Dispatcher
	metrics *metrics.Registry
}

// NewAPI builds a handler container with shared dependencies.
func NewAPI(
	version, commit string,
	webFS fs.FS,
	logger *slog.Logger,
	mgr *heartbeat.Manager,
	hist history.Store,
	rec *servicehistory.Recorder,
	disp *notifier.Dispatcher,
) *API {
	return &API{
		Version: version,
		Commit:  commit,
		webFS:   webFS,
		Logger:  logger,
		mgr:     mgr,
		hist:    hist,
		rec:     rec,
		disp:    disp,
		metrics: metrics.New(hist),
	}
}

// respondJSON writes a JSON response.
func (a *API) respondJSON(w http.ResponseWriter, status int, v any) {
	if err := encode(w, status, v); err != nil {
		a.Logger.Error("encode response failed", "err", err)
	}
}

// statusResponse is the standard success payload.
type statusResponse struct {
	Status string `json:"status"`
}

// errorResponse is the standard error payload.
type errorResponse struct {
	Error string `json:"error"`
}
