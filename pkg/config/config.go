package config

import (
	"fmt"
	"heartbeats/pkg/heartbeat"
	"heartbeats/pkg/history"
	"heartbeats/pkg/notify"
	"os"

	"gopkg.in/yaml.v3"
)

// App is the global configuration instance.
var App = &Config{
	HeartbeatStore:    heartbeat.NewStore(),
	NotificationStore: notify.NewStore(),
}

// HistoryStore is the global HistoryStore.
var HistoryStore = history.NewStore()

// Cache configuration structure.
type Cache struct {
	MaxSize int `yaml:"maxSize"` // Maximum size of the cache
	Reduce  int `yaml:"reduce"`  // Amount to reduce when max size is exceeded
}

// Server configuration structure.
type Server struct {
	SiteRoot      string `yaml:"siteRoot"`      // Site root
	ListenAddress string `yaml:"listenAddress"` // Address on which the application listens
}

// Config holds the entire application configuration.
type Config struct {
	Version           string           `yaml:"version"`
	Verbose           bool             `yaml:"verbose"`
	Path              string           `yaml:"path"`
	Server            Server           `yaml:"server"`
	Cache             Cache            `yaml:"cache"`
	HeartbeatStore    *heartbeat.Store `yaml:"heartbeats"`
	NotificationStore *notify.Store    `yaml:"notifications"`
}

// Read reads the configuration from the file specified in the Config struct.
func (c *Config) Read() error {
	content, err := os.ReadFile(c.Path)
	if err != nil {
		return fmt.Errorf("failed to read config file. %w", err)
	}

	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(content, &rawConfig); err != nil {
		return fmt.Errorf("failed to unmarshal raw config. %w", err)
	}

	if err := c.processNotifications(rawConfig["notifications"]); err != nil {
		return err
	}

	if err := c.processHeartbeats(rawConfig["heartbeats"]); err != nil {
		return err
	}

	return nil
}

// processNotifications handles the unmarshaling and processing of notification configurations.
func (c *Config) processNotifications(rawNotifications interface{}) error {
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

		if err := c.NotificationStore.Add(name, &notification); err != nil {
			return fmt.Errorf("failed to add notification '%s'. %w", name, err)
		}

		if err := c.updateSlackNotification(name, &notification); err != nil {
			return err
		}
	}

	return nil
}

// updateSlackNotification updates the Slack notification with a default color template if not set.
func (c *Config) updateSlackNotification(name string, notification *notify.Notification) error {
	if notification.Type == "slack" && notification.SlackConfig.ColorTemplate == "" {
		notification.SlackConfig.ColorTemplate = `{{ if eq .Status "ok" }}good{{ else }}danger{{ end }}`
		if err := c.NotificationStore.Update(name, notification); err != nil {
			return fmt.Errorf("failed to update notification '%s'. %w", notification.Name, err)
		}
	}
	return nil
}

// processHeartbeats handles the unmarshaling and processing of heartbeat configurations.
func (c *Config) processHeartbeats(rawHeartbeats interface{}) error {
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

		if err := c.HeartbeatStore.Add(name, &hb); err != nil {
			return fmt.Errorf("failed to add heartbeat '%s'. %w", hb.Name, err)
		}

		historyInstance, err := history.NewHistory(c.Cache.MaxSize, c.Cache.Reduce)
		if err != nil {
			return fmt.Errorf("failed to add heartbeat '%s'. %w", hb.Name, err)
		}

		if err := HistoryStore.Add(name, historyInstance); err != nil {
			return fmt.Errorf("failed to create heartbeat history for '%s'. %w", name, err)
		}
	}

	return nil
}
