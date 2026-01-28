package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/containeroo/heartbeats/internal/resolve"
)

// Load reads and validates a YAML configuration file.
func Load(path string) (*Config, error) {
	return LoadWithOptions(path, LoadOptions{})
}

// LoadOptions controls config loading behavior.
type LoadOptions struct {
	StrictEnv bool // Whether unresolved env vars should error.
}

// LoadWithOptions reads and validates a YAML configuration file.
func LoadWithOptions(path string, opts LoadOptions) (*Config, error) {
	if path == "" {
		return nil, errors.New("config path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	data, err = resolve.ExpandEnv(data, resolve.Options{Strict: opts.StrictEnv})
	if err != nil {
		return nil, fmt.Errorf("resolve env: %w", err)
	}
	cfg := &Config{}
	if err := yamlUnmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate checks for missing or invalid config values.
func (c *Config) Validate() error {
	if err := validateReceivers(c.Receivers); err != nil {
		return err
	}
	if err := validateHeartbeats(c.Heartbeats, c.Receivers); err != nil {
		return err
	}
	return validateHistory(c.History)
}

// validateReceivers validates a map of receiver configurations.
func validateReceivers(receivers map[string]ReceiverConfig) error {
	if len(receivers) == 0 {
		return errors.New("at least one receiver is required")
	}
	for name, rcv := range receivers {
		if err := validateReceiver(name, rcv); err != nil {
			return err
		}
	}
	return nil
}

// validateReceiver validates a single receiver configuration.
func validateReceiver(name string, rcv ReceiverConfig) error {
	if len(rcv.Webhooks) == 0 && len(rcv.Emails) == 0 {
		return fmt.Errorf("receiver %q must configure webhooks and/or emails", name)
	}
	for idx, webhook := range rcv.Webhooks {
		if webhook.URL == "" {
			return fmt.Errorf("receiver %q webhook[%d] url is required", name, idx)
		}
	}
	for idx, email := range rcv.Emails {
		if email.Host == "" || email.From == "" || len(email.To) == 0 {
			return fmt.Errorf("receiver %q email[%d] host/from/to are required", name, idx)
		}
	}
	return nil
}

// validateHeartbeats validates a map of heartbeat configurations.
func validateHeartbeats(heartbeats map[string]HeartbeatConfig, receivers map[string]ReceiverConfig) error {
	if len(heartbeats) == 0 {
		return errors.New("at least one heartbeat is required")
	}
	for id, hb := range heartbeats {
		if err := validateHeartbeat(id, hb, receivers); err != nil {
			return err
		}
	}
	return nil
}

// validateHeartbeat validates a single heartbeat configuration.
func validateHeartbeat(id string, hb HeartbeatConfig, receivers map[string]ReceiverConfig) error {
	if hb.Interval <= 0 {
		return fmt.Errorf("heartbeat %q interval must be > 0", id)
	}
	if hb.LateAfter <= 0 {
		return fmt.Errorf("heartbeat %q late_after must be > 0", id)
	}
	if len(hb.Receivers) == 0 {
		return fmt.Errorf("heartbeat %q must have at least one receiver", id)
	}
	for _, r := range hb.Receivers {
		if _, ok := receivers[r]; !ok {
			return fmt.Errorf("heartbeat %q references unknown receiver %q", id, r)
		}
	}
	return nil
}

// validateHistory validates the history configuration.
func validateHistory(cfg HistoryConfig) error {
	if cfg.Size < 0 {
		return errors.New("history size must be >= 0")
	}
	if cfg.Buffer < 0 {
		return errors.New("history buffer must be >= 0")
	}
	return nil
}
