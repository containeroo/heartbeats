package heartbeat

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalYAML_HeartbeatConfigMap(t *testing.T) {
	t.Parallel()

	t.Run("invalid Heartbeats", func(t *testing.T) {
		t.Parallel()

		input := `
invalid:
  description: "Ping Google"
  interval: 10s
  grace: five
  receivers: ["email", "slack"]
`
		var cfg HeartbeatConfigMap
		err := yaml.NewDecoder(strings.NewReader(input)).Decode(&cfg)
		assert.Error(t, err)
		assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 5: cannot unmarshal !!str `five` into time.Duration")
	})

	t.Run("injects ID into HeartbeatConfig", func(t *testing.T) {
		t.Parallel()

		input := `
ping:
  description: "Ping Google"
  interval: 10s
  grace: 5s
  receivers: ["email", "slack"]
healthcheck:
  description: "Healthcheck API"
  interval: 30s
  grace: 15s
  receivers: ["pagerduty"]
`

		var cfg HeartbeatConfigMap
		err := yaml.NewDecoder(strings.NewReader(input)).Decode(&cfg)
		assert.NoError(t, err)

		assert.Len(t, cfg, 2)

		ping := cfg["ping"]
		assert.Equal(t, "ping", ping.ID)
		assert.Equal(t, "Ping Google", ping.Description)
		assert.Equal(t, 10*time.Second, ping.Interval)
		assert.Equal(t, 5*time.Second, ping.Grace)
		assert.Equal(t, []string{"email", "slack"}, ping.Receivers)

		hc := cfg["healthcheck"]
		assert.Equal(t, "healthcheck", hc.ID)
		assert.Equal(t, "Healthcheck API", hc.Description)
		assert.Equal(t, 30*time.Second, hc.Interval)
		assert.Equal(t, 15*time.Second, hc.Grace)
		assert.Equal(t, []string{"pagerduty"}, hc.Receivers)
	})
}
