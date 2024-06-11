package config

import (
	"fmt"
	"heartbeats/internal/heartbeat"
	"heartbeats/internal/history"
	"heartbeats/internal/notify"
	"heartbeats/internal/timer"
	"reflect"
	"time"

	"github.com/spf13/viper"
)

// App is the global configuration instance.
var App = &Config{
	HeartbeatStore:    heartbeat.NewStore(),
	NotificationStore: notify.NewStore(),
}

var HistoryStore = history.NewStore()

// Cache configuration structure.
type Cache struct {
	MaxSize int `yaml:"maxSize"` // Maximum size of the cache
	Reduce  int `yaml:"reduce"`  // Amount to reduce when max size is exceeded
}

// Server configuration structure.
type Server struct {
	SiteRoot      string // Site root
	ListenAddress string // Address on which the application listens
}

// Config holds the entire application configuration.
type Config struct {
	Version           string
	Verbose           bool
	Path              string
	Server            Server
	Cache             Cache
	HeartbeatStore    *heartbeat.Store `yaml:"heartbeats"`
	NotificationStore *notify.Store    `yaml:"notifications"`
}

// Read reads the configuration file and initializes the stores.
func (c *Config) Read() error {
	viper.SetConfigFile(c.Path)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file. %w", err)
	}

	notifications := make(map[string]*notify.Notification)
	if err := viper.UnmarshalKey("notifications", &notifications); err != nil {
		return fmt.Errorf("failed to unmarshal notifications. %w", err)
	}

	for name, n := range notifications {
		if err := c.NotificationStore.Add(name, n); err != nil {
			return fmt.Errorf("failed to add notification '%s'. %w", n.Name, err)
		}

		notification := c.NotificationStore.Get(name)
		if notification.Type == "slack" && notification.SlackConfig.ColorTemplate == "" {
			notification.SlackConfig.ColorTemplate = `{{ if eq .Status "ok" }}good{{ else }}danger{{ end }}`
			if err := c.NotificationStore.Update(name, notification); err != nil {
				return fmt.Errorf("failed to update notification '%s'. %w", notification.Name, err)
			}
		}
	}

	heartbeats := make(map[string]*heartbeat.Heartbeat)
	if err := viper.UnmarshalKey("heartbeats", &heartbeats, viper.DecodeHook(decodeHookHeartbeats)); err != nil {
		return fmt.Errorf("failed to unmarshal heartbeats. %w", err)
	}

	for name, h := range heartbeats {
		if err := c.HeartbeatStore.Add(name, h); err != nil {
			return fmt.Errorf("failed to add heartbeat '%s'. %w", h.Name, err)
		}
		historyInstance := history.NewHistory(c.Cache.MaxSize, c.Cache.Reduce)
		if err := HistoryStore.Add(name, historyInstance); err != nil {
			return fmt.Errorf("failed to create heartbeat history for '%s'. %w", name, err)
		}
	}

	return nil
}

// decodeHookHeartbeats is a custom decode hook for the 'heartbeats' configuration section.
func decodeHookHeartbeats(_, to reflect.Type, data interface{}) (interface{}, error) {
	if data == nil {
		return nil, nil // No data to process
	}

	heartbeats, ok := data.(map[string]interface{})
	if !ok {
		return data, nil // Not the correct type, no error but no processing
	}

	result := make(map[string]heartbeat.Heartbeat)

	for name, hb := range heartbeats {
		h, ok := hb.(map[string]interface{})
		if !ok {
			return data, fmt.Errorf("failed to assert heartbeat as map[string]interface{}")
		}

		description, _ := h["description"].(string)

		interval, err := parseTimer(h, "interval")
		if err != nil {
			return nil, fmt.Errorf("failed to parse interval for heartbeat '%s': %w", name, err)
		}

		grace, err := parseTimer(h, "grace")
		if err != nil {
			return nil, fmt.Errorf("failed to parse grace for heartbeat '%s': %w", name, err)
		}

		var sendResolve *bool
		if sr, ok := h["sendresolve"].(bool); ok { // All lowercase because of viper
			sendResolve = &sr
		}

		notifications, err := parseNotifications(h)
		if err != nil {
			return nil, fmt.Errorf("failed to parse notifications for heartbeat '%s': %w", name, err)
		}

		result[name] = heartbeat.Heartbeat{
			Name:          name,
			Description:   description,
			Grace:         grace,
			Interval:      interval,
			SendResolve:   sendResolve,
			Notifications: notifications,
		}
	}

	return result, nil
}

// parseTimer parses a timer configuration from the heartbeat configuration.
func parseTimer(h map[string]interface{}, key string) (*timer.Timer, error) {
	t := &timer.Timer{}
	if value, ok := h[key].(string); ok {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse '%s' as duration: %w", key, err)
		}
		t.Interval = &duration
	}
	return t, nil
}

// parseNotifications parses the notifications configuration from the heartbeat configuration.
func parseNotifications(h map[string]interface{}) ([]string, error) {
	var notifications []string
	if notificationList, ok := h["notifications"].([]interface{}); ok {
		for _, notificationName := range notificationList {
			name, isString := notificationName.(string)
			if !isString {
				return nil, fmt.Errorf("notification name is not a string")
			}
			notifications = append(notifications, name)
		}
	}
	return notifications, nil
}
