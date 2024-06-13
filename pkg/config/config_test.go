package config

import (
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/notify"
	"heartbeats/pkg/notify/notifier"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// Sample configuration YAML for testing.
const sampleConfig = `
version: "1.0.0"
verbose: true
path: "./config.yaml"
server:
  siteRoot: "http://localhost:8080"
  listenAddress: "localhost:8080"
cache:
  maxSize: 100
  reduce: 10
notifications:
  slack:
    type: "slack"
    slack_config:
      channel: "general"
heartbeats:
  heartbeat1:
    name: "heartbeat1"
    interval: "1m"
    grace: "1m"
    notifications:
      - slack
`

func writeSampleConfig(t *testing.T, content string) string {
	file, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	return file.Name()
}

func TestConfig_Read(t *testing.T) {
	App.NotificationStore = notify.NewStore()
	HistoryStore = history.NewStore()

	tempFile := writeSampleConfig(t, sampleConfig)
	defer os.Remove(tempFile)

	App.Path = tempFile

	err := App.Read()
	assert.NoError(t, err, "Expected no error when reading the config file")

	notification := App.NotificationStore.Get("slack")
	assert.NotNil(t, notification, "Expected slack notification to be present")
	assert.Equal(t, `{{ if eq .Status "ok" }}good{{ else }}danger{{ end }}`, notification.SlackConfig.ColorTemplate)

	heartbeat := App.HeartbeatStore.Get("heartbeat1")
	assert.NotNil(t, heartbeat, "Expected heartbeat1 to be present")
	assert.Equal(t, "heartbeat1", heartbeat.Name)
}

func TestProcessNotifications(t *testing.T) {
	App.NotificationStore = notify.NewStore()
	HistoryStore = history.NewStore()

	var rawConfig map[string]interface{}
	err := yaml.Unmarshal([]byte(sampleConfig), &rawConfig)
	assert.NoError(t, err)
	err = App.processNotifications(rawConfig["notifications"])
	assert.NoError(t, err, "Expected no error when processing notifications")

	notification := App.NotificationStore.Get("slack")
	assert.NotNil(t, notification, "Expected slack notification to be present")
	assert.Equal(t, "slack", notification.Type)
	assert.Equal(t, `{{ if eq .Status "ok" }}good{{ else }}danger{{ end }}`, notification.SlackConfig.ColorTemplate)
}

func TestProcessHeartbeats(t *testing.T) {
	App.HeartbeatStore = heartbeat.NewStore()
	HistoryStore = history.NewStore()

	var rawConfig map[string]interface{}
	err := yaml.Unmarshal([]byte(sampleConfig), &rawConfig)
	assert.NoError(t, err)

	err = App.processHeartbeats(rawConfig["heartbeats"])
	assert.NoError(t, err, "Expected no error when processing heartbeats")

	heartbeat := App.HeartbeatStore.Get("heartbeat1")
	assert.NotNil(t, heartbeat, "Expected heartbeat1 to be present")
	assert.Equal(t, "heartbeat1", heartbeat.Name)
}

func TestUpdateSlackNotification(t *testing.T) {
	notification := &notify.Notification{
		Type: "slack",
		SlackConfig: &notifier.SlackConfig{
			Channel: "general",
		},
	}

	err := App.updateSlackNotification("slack", notification)
	assert.NoError(t, err, "Expected no error when updating slack notification")
	assert.Equal(t, `{{ if eq .Status "ok" }}good{{ else }}danger{{ end }}`, notification.SlackConfig.ColorTemplate)
}
