package config

import (
	"fmt"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/notify"
	"os"

	"gopkg.in/yaml.v3"
)

// Read reads the configuration from the file specified in the Config struct.
func Read(path string, heartbeatsStore *heartbeat.Store, notificationStore *notify.Store, historyStore *history.Store, maxSize int, reduce int) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file. %w", err)
	}

	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(content, &rawConfig); err != nil {
		return fmt.Errorf("failed to unmarshal raw config. %w", err)
	}

	if err := processNotifications(rawConfig["notifications"], notificationStore); err != nil {
		return err
	}

	if err := processHeartbeats(rawConfig["heartbeats"], heartbeatsStore, historyStore, maxSize, reduce); err != nil {
		return err
	}

	return nil
}

// processNotifications handles the unmarshaling and processing of notification configurations.
func processNotifications(rawNotifications interface{}, notificationStore *notify.Store) error {
	notifications, ok := rawNotifications.(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to unmarshal notifications")
	}

	for name, rawNotification := range notifications {
		notificationBytes, err := yaml.Marshal(rawNotification)
		if err != nil {
			return fmt.Errorf("failed to marshal notification '%s'. %w", name, err)
		}

		var notification notify.Notification
		if err := yaml.Unmarshal(notificationBytes, &notification); err != nil {
			return fmt.Errorf("failed to unmarshal notification '%s'. %w", name, err)
		}

		if err := notificationStore.Add(name, &notification); err != nil {
			return fmt.Errorf("failed to add notification '%s'. %w", name, err)
		}

		if err := updateSlackNotification(name, &notification, notificationStore); err != nil {
			return err
		}
	}

	return nil
}

// updateSlackNotification updates the Slack notification with a default color template if not set.
func updateSlackNotification(name string, notification *notify.Notification, notificationStore *notify.Store) error {
	if notification.Type == "slack" && notification.SlackConfig.ColorTemplate == "" {
		notification.SlackConfig.ColorTemplate = `{{ if eq .Status "ok" }}good{{ else }}danger{{ end }}`
		if err := notificationStore.Update(name, notification); err != nil {
			return fmt.Errorf("failed to update notification '%s'. %w", notification.Name, err)
		}
	}
	return nil
}

// processHeartbeats handles the unmarshaling and processing of heartbeat configurations.
func processHeartbeats(rawHeartbeats interface{}, heartbeatStore *heartbeat.Store, historyStore *history.Store, maxSize, reduce int) error {
	heartbeats, ok := rawHeartbeats.(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to unmarshal heartbeats")
	}

	for name, rawHeartbeat := range heartbeats {
		heartbeatBytes, err := yaml.Marshal(rawHeartbeat)
		if err != nil {
			return fmt.Errorf("failed to marshal heartbeat '%s'. %w", name, err)
		}

		var hb heartbeat.Heartbeat
		if err := yaml.Unmarshal(heartbeatBytes, &hb); err != nil {
			return fmt.Errorf("failed to unmarshal heartbeat '%s'. %w", name, err)
		}

		if err := heartbeatStore.Add(name, &hb); err != nil {
			return fmt.Errorf("failed to add heartbeat '%s'. %w", hb.Name, err)
		}

		historyInstance, err := history.NewHistory(maxSize, reduce)
		if err != nil {
			return fmt.Errorf("failed to create history for heartbeat '%s'. %w", name, err)
		}

		if err := historyStore.Add(name, historyInstance); err != nil {
			return fmt.Errorf("failed to create heartbeat history for '%s'. %w", name, err)
		}
	}

	return nil
}
