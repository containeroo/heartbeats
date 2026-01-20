package handler

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/history"
	"github.com/containeroo/heartbeats/internal/logging"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notifier"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
)

// API bundles shared handler dependencies and runtime configuration.
type API struct {
	Version     string // Version is the build version string.
	Commit      string // Commit is the build commit SHA.
	webFS       fs.FS
	SiteRoot    string
	RoutePrefix string
	Debug       bool
	Logger      *slog.Logger
	mgr         *heartbeat.Manager
	hist        history.Store
	rec         *servicehistory.Recorder
	disp        *notifier.Dispatcher
	metrics     *metrics.Registry
	reload      func() error
}

// NewAPI builds a handler container with shared dependencies.
func NewAPI(
	version, commit string,
	webFS fs.FS,
	siteRoot string,
	routePrefix string,
	debug bool,
	logger *slog.Logger,
	mgr *heartbeat.Manager,
	hist history.Store,
	rec *servicehistory.Recorder,
	disp *notifier.Dispatcher,
	metricsReg *metrics.Registry,
	configReloadFn func() error,
) *API {
	return &API{
		Version:     version,
		Commit:      commit,
		webFS:       webFS,
		SiteRoot:    siteRoot,
		RoutePrefix: routePrefix,
		Debug:       debug,
		Logger:      logger,
		mgr:         mgr,
		hist:        hist,
		rec:         rec,
		disp:        disp,
		metrics:     metricsReg,
		reload:      configReloadFn,
	}
}

// respondJSON writes a JSON response.
func (a *API) respondJSON(w http.ResponseWriter, status int, v any) {
	if err := encode(w, status, v); err != nil {
		logging.AccessLogger(a.Logger, nil).Error(
			"encode response failed",
			"event", "encode_response_failed",
			"err", err,
		)
	}
}

// businessLogger returns a logger enriched with request context for business events.
func (a *API) businessLogger(r *http.Request) *slog.Logger {
	return logging.BusinessLogger(a.Logger, r.Context())
}

// logRequestError records a structured error for the current request.
func (a *API) logRequestError(r *http.Request, event, message string, err error) {
	logging.AccessLogger(a.Logger, r.Context()).Error(
		message,
		"event", event,
		"err", err,
	)
}

// statusResponse is the standard success payload.
type statusResponse struct {
	Status string `json:"status"`
}

// errorResponse is the standard error payload.
type errorResponse struct {
	Error string `json:"error"`
}
