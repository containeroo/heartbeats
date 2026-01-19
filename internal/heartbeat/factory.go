package heartbeat

import (
	"fmt"
	"log/slog"

	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/notifier"
	servicehistory "github.com/containeroo/heartbeats/internal/service/history"
)

// ActorFactory builds actors from heartbeat configs.
type ActorFactory interface {
	Build(cfg HeartbeatConfig) (*Actor, error)
}

// DefaultActorFactory builds actors using the shared dependencies.
type DefaultActorFactory struct {
	Logger     *slog.Logger
	History    *servicehistory.Recorder
	Metrics    *metrics.Registry
	DispatchCh chan<- notifier.NotificationData
}

// Build constructs a new Actor without starting it.
func (f DefaultActorFactory) Build(cfg HeartbeatConfig) (*Actor, error) {
	if f.Metrics == nil {
		return nil, fmt.Errorf("metrics registry is required")
	}
	return NewActorFromConfig(ActorConfig{
		ID:          cfg.ID,
		Description: cfg.Description,
		Interval:    cfg.Interval,
		Grace:       cfg.Grace,
		Receivers:   cfg.Receivers,
		Logger:      f.Logger,
		History:     f.History,
		DispatchCh:  f.DispatchCh,
		Metrics:     f.Metrics,
	}), nil
}
