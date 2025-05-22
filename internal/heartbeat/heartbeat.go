package heartbeat

import (
	"time"

	"gopkg.in/yaml.v3"
)

// HeartbeatConfig contains heartbeat settings.
type HeartbeatConfig struct {
	ID          string        `json:"id"`
	Description string        `json:"description"`
	Interval    time.Duration `json:"interval"`
	Grace       time.Duration `json:"grace"`
	Receivers   []string      `json:"receivers"`
}

// HeartbeatConfigMap injects the map key as the ID field into each HeartbeatConfig during YAML unmarshalling.
type HeartbeatConfigMap map[string]HeartbeatConfig

// UnmarshalYAML parses the map and assigns each key as the ID field of its HeartbeatConfig.
func (hcm *HeartbeatConfigMap) UnmarshalYAML(value *yaml.Node) error {
	raw := make(map[string]HeartbeatConfig)
	if err := value.Decode(&raw); err != nil {
		return err
	}

	*hcm = make(HeartbeatConfigMap, len(raw))
	for id, cfg := range raw {
		cfg.ID = id
		(*hcm)[id] = cfg
	}
	return nil
}
