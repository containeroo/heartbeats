package config

import (
	"fmt"
	"os"

	"github.com/containeroo/heartbeats/internal/heartbeat"
	"github.com/containeroo/heartbeats/internal/notifier"
	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration.
type Config struct {
	Receivers  map[string]notifier.ReceiverConfig `yaml:"receivers"`  // Receivers is the map of receiver IDs to their configurations.
	Heartbeats heartbeat.HeartbeatConfigMap       `yaml:"heartbeats"` // Heartbeats is the map of heartbeat IDs to their configurations.
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate checks that each heartbeat references an existing receiver and
// validates and resolves secrets in receiver configurations.
func (c *Config) Validate() error {
	for receiverID, rc := range c.Receivers {
		for i := range rc.SlackConfigs {
			if err := rc.SlackConfigs[i].Resolve(); err != nil {
				return fmt.Errorf("receiver %q slack config error: %w", receiverID, err)
			}
			if err := rc.SlackConfigs[i].Validate(); err != nil {
				return fmt.Errorf("receiver %q slack config error: %w", receiverID, err)
			}
		}
		for i := range rc.EmailConfigs {
			if err := rc.EmailConfigs[i].Resolve(); err != nil {
				return fmt.Errorf("receiver %q email config error: %w", receiverID, err)
			}
			if err := rc.EmailConfigs[i].Validate(); err != nil {
				return fmt.Errorf("receiver %q email config error: %w", receiverID, err)
			}
		}
		for i := range rc.MSTeamsConfigs {
			if err := rc.MSTeamsConfigs[i].Resolve(); err != nil {
				return fmt.Errorf("receiver %q MSTeams config error: %w", receiverID, err)
			}
			if err := rc.MSTeamsConfigs[i].Validate(); err != nil {
				return fmt.Errorf("receiver %q MSTeams config error: %w", receiverID, err)
			}
		}
		c.Receivers[receiverID] = rc // Write back updated receiver config.
	}

	// Validate that each heartbeat references valid receivers.
	for hbName, hb := range c.Heartbeats {
		for _, receiverName := range hb.Receivers {
			if _, ok := c.Receivers[receiverName]; !ok {
				return fmt.Errorf("heartbeat %q references unknown receiver %q", hbName, receiverName)
			}
		}
	}
	return nil
}
