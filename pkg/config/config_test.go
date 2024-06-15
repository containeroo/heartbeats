package config

import (
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/notify"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestRead(t *testing.T) {
	sampleConfig := `
heartbeats:
  test_heartbeat:
    enabled: true
    interval: 1m
    grace: 1m
    notifications: ["test_notification"]

notifications:
  test_notification:
    slack_config:
      channel: general

`
	tmpFile, err := os.CreateTemp("", "config.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(sampleConfig))
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	heartbeatsStore := heartbeat.NewStore()
	notificationStore := notify.NewStore()
	historyStore := history.NewStore()

	maxSize := 110
	reduce := 25

	t.Run("Read config file and process notifications and heartbeats", func(t *testing.T) {
		err := Read(tmpFile.Name(), heartbeatsStore, notificationStore, historyStore, maxSize, reduce)
		assert.NoError(t, err)

		// Verify heartbeats
		hb := heartbeatsStore.Get("test_heartbeat")
		assert.NotNil(t, hb)
		assert.Equal(t, "test_heartbeat", hb.Name)
		assert.NotNil(t, hb.Interval)
		assert.NotNil(t, hb.Grace)
		assert.Len(t, hb.Notifications, 1)
		assert.Equal(t, "test_notification", hb.Notifications[0])

		// Verify notifications
		n := notificationStore.Get("test_notification")
		assert.NotNil(t, n)
		assert.Equal(t, "test_notification", n.Name)
		assert.Equal(t, "slack", n.Type)

		// Verify history
		h := historyStore.Get("test_heartbeat")
		assert.NotNil(t, h)
		assert.Equal(t, 110, maxSize)
		assert.Equal(t, 25, reduce)
	})
}

func TestProcessNotifications(t *testing.T) {
	sampleConfig := `
heartbeats:
  test_heartbeat:
    enabled: true
    interval: 1m
    grace: 1m
    notifications: ["test_notification"]

notifications:
  test_notification:
    slack_config:
      channel: general

`

	notificationStore := notify.NewStore()

	t.Run("Process valid notifications", func(t *testing.T) {
		var rawConfig map[string]interface{}
		err := yaml.Unmarshal([]byte(sampleConfig), &rawConfig)
		assert.NoError(t, err)

		err = processNotifications(rawConfig["notifications"], notificationStore)
		assert.NoError(t, err)

		// Verify notification
		n := notificationStore.Get("test_notification")
		assert.NotNil(t, n)
		assert.Equal(t, "test_notification", n.Name)
		assert.Equal(t, "slack", n.Type)
	})
}

func TestProcessHeartbeats(t *testing.T) {
	sampleConfig := `
heartbeats:
  test_heartbeat:
    enabled: true
    interval: 1m
    grace: 1m
    notifications: ["test_notification"]

notifications:
  test_notification:
    slack_config:
      channel: general

`

	heartbeatsStore := heartbeat.NewStore()
	historyStore := history.NewStore()

	maxSize := 120
	reduce := 25

	t.Run("Process valid heartbeats", func(t *testing.T) {
		var rawConfig map[string]interface{}
		err := yaml.Unmarshal([]byte(sampleConfig), &rawConfig)
		assert.NoError(t, err)

		err = processHeartbeats(rawConfig["heartbeats"], heartbeatsStore, historyStore, maxSize, reduce)
		assert.NoError(t, err)

		// Verify heartbeat
		hb := heartbeatsStore.Get("test_heartbeat")
		assert.NotNil(t, hb)
		assert.Equal(t, "test_heartbeat", hb.Name)
		assert.NotNil(t, hb.Interval)
		assert.NotNil(t, hb.Grace)
		assert.Len(t, hb.Notifications, 1)
		assert.Equal(t, "test_notification", hb.Notifications[0])

		// Verify history
		h := historyStore.Get("test_heartbeat")
		assert.NotNil(t, h)
		assert.Equal(t, 120, maxSize)
		assert.Equal(t, 25, reduce)
	})
}
