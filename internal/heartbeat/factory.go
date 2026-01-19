package heartbeat

import (
	"log/slog"

	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notifier"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
)

// ActorFactory builds actors from heartbeat configs.
type ActorFactory interface {
	Build(cfg HeartbeatConfig) (*Actor, error)
}

// ActorDeps bundles shared dependencies for actors.
type ActorDeps struct {
	Logger     *slog.Logger
	History    *servicehistory.Recorder
	Metrics    *metrics.Registry
	DispatchCh chan<- notifier.NotificationData
}

// DefaultActorFactory builds actors using the shared dependencies.
type DefaultActorFactory struct {
	Deps ActorDeps
}

// Build constructs a new Actor without starting it.
func (f DefaultActorFactory) Build(cfg HeartbeatConfig) (*Actor, error) {
	return NewActorFromConfig(ActorConfig{
		ID:          cfg.ID,
		Description: cfg.Description,
		Interval:    cfg.Interval,
		Grace:       cfg.Grace,
		Receivers:   cfg.Receivers,
		Logger:      f.Deps.Logger,
		History:     f.Deps.History,
		DispatchCh:  f.Deps.DispatchCh,
		Metrics:     f.Deps.Metrics,
	}), nil
}
