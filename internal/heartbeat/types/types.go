package types

import (
	kit "github.com/containeroo/notifykit/notify"

	"github.com/containeroo/heartbeats/internal/config"
	"github.com/containeroo/heartbeats/internal/runner"
)

// Heartbeat holds runtime state for a configured heartbeat.
type Heartbeat struct {
	ID              string
	Title           string
	Config          config.HeartbeatConfig
	Receivers       []string
	ReceiverIDs     []kit.ReceiverID
	State           *runner.State
	AlertOnRecovery bool
	AlertOnLate     bool
}
