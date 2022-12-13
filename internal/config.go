package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var HeartbeatsServer Heartbeats

// Cache holds the configuration for the cache
type Cache struct {
	MaxSize int `mapstructure:"max_size"`
	Reduce  int `mapstructure:"reduce"`
}

// Config config holds general configuration
type Config struct {
	Path    string `mapstructure:"path"`
	Logging string `mapstructure:"logging"`
}

// Details details holds defaults for notifications
type Defaults struct {
	Subject      string `mapstructure:"subject"`
	Message      string `mapstructure:"message"`
	SendResolved *bool  `mapstructure:"sendResolved"`
}

// NotifyConfig holds the configuration for the notifications
type Notifications struct {
	Defaults Defaults       `mapstructure:"defaults"`
	Services []Notification `mapstructure:"services"`
}

// Heartbeats is the main configuration struct
type Heartbeats struct {
	Version       string        `mapstructure:"version"`
	Config        Config        `mapstructure:"config"`
	Server        Server        `mapstructure:"server"`
	Cache         Cache         `mapstructure:"cache"`
	Heartbeats    []Heartbeat   `mapstructure:"heartbeats"`
	Notifications Notifications `mapstructure:"notifications"`
}

// ReadConfigFile reads the notifications config file
// configPath is the path to the config file
// init is true if the config file is read on startup. this will skip the comparison of the previous config
func ReadConfigFile(configPath string, init bool) error {
	parentDir := filepath.Dir(configPath)
	absolutePath, err := filepath.Abs(parentDir)
	if err != nil {
		return err
	}
	fileExtension := filepath.Ext(configPath)
	fileExtensionWithoutDot := strings.TrimPrefix(fileExtension, ".")

	viper.AddConfigPath(absolutePath) // necessary for notify
	viper.SetConfigFile(configPath)
	viper.SetConfigType(fileExtensionWithoutDot)

	var previousHeartbeats []Heartbeat
	if !init {
		if HeartbeatsServer.Heartbeats != nil {
			previousHeartbeats = HeartbeatsServer.Heartbeats
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&HeartbeatsServer); err != nil {
		return err
	}

	if len(HeartbeatsServer.Heartbeats) == 0 {
		return fmt.Errorf("no heartbeats configured")
	}

	if len(HeartbeatsServer.Notifications.Services) == 0 {
		return fmt.Errorf("no notifications configured")
	}

	if err := ProcessServiceSettings(); err != nil {
		return err
	}

	if err := ProecessHeartbeatsSettings(); err != nil {
		return err
	}

	if !init {
		ResetTimerIfRunning(&previousHeartbeats)
	}

	return nil
}

// ResetTimerIfRunning resets existing timers if they are running
func ResetTimerIfRunning(previousHeartbeats *[]Heartbeat) {
	for _, currentHeartbeat := range HeartbeatsServer.Heartbeats {

		var previousHeartbeat *Heartbeat
		var err error
		for _, p := range *previousHeartbeats {
			previousHeartbeat, err = HeartbeatsServer.GetHeartbeatByName(p.Name)
			if err != nil {
				log.Errorf("%s not found in previous heartbeats. %s", currentHeartbeat.Name, err.Error())
				continue
			}
		}
		if previousHeartbeat == nil {
			log.Errorf("%s not found in previous heartbeats", currentHeartbeat.Name)
			continue
		}

		// Heartbeat Interval is running
		if (currentHeartbeat.IntervalTimer != nil && !currentHeartbeat.IntervalTimer.Completed) && (currentHeartbeat.Interval != previousHeartbeat.Interval) {
			currentHeartbeat.IntervalTimer.Reset(currentHeartbeat.Interval)
			log.Debugf("%s Interval timer reset to %s", currentHeartbeat.Name, currentHeartbeat.Interval)
			return
		}

		// Heartbeat Grace is running
		if (currentHeartbeat.GraceTimer != nil && !currentHeartbeat.GraceTimer.Completed) && (currentHeartbeat.Grace != previousHeartbeat.Grace) {
			currentHeartbeat.GraceTimer.Reset(currentHeartbeat.Grace)
			log.Debugf("%s Grace timer reset to %s", currentHeartbeat.Name, currentHeartbeat.Grace)
		}
	}
}

// ProcessNotificationSettings checks if all services are valid and can be parsed
func ProcessServiceSettings() error {
	var heartbeat Heartbeat

	if HeartbeatsServer.Notifications.Defaults.SendResolved == nil {
		HeartbeatsServer.Notifications.Defaults.SendResolved = new(bool) // must be set, otherwise the debug message will fail because of a nil pointer
		log.Tracef("Default «sendResolved» not explicitly enabled. Defaulting to true")
	}

	if HeartbeatsServer.Notifications.Defaults.Message == "" {
		HeartbeatsServer.Notifications.Defaults.Message = "Heartbeat «{{.Name}}» is {{.Status}}"
		log.Tracef("Default «message» not explicitly set. Defaulting to «%s»", HeartbeatsServer.Notifications.Defaults.Message)
	}

	for _, service := range HeartbeatsServer.Notifications.Services {
		url := os.ExpandEnv(service.Shoutrrr)       // expand any environment variables
		url, err := FormatTemplate(url, &heartbeat) // expand any template variables
		if err != nil {
			return fmt.Errorf("Could not format shoutrrr url «%s» for «%s». %s", service.Shoutrrr, service.Name, err)
		}

		message := CheckDefault(service.Message, HeartbeatsServer.Notifications.Defaults.Message)
		message = os.ExpandEnv(message)                    // expand any environment variables
		message, err = FormatTemplate(message, &heartbeat) // expand any template variables
		if err != nil {
			return fmt.Errorf("Could not format message «%s» for «%s». %s", service.Message, service.Name, err)
		}

		if service.Enabled == nil {
			service.Enabled = new(bool) // must be set, otherwise the debug message below will fail because of a nil pointer
			log.Tracef("service «%s» not explicitly enabled. Defaulting to true", service.Name)
		}

		if service.SendResolved == nil {
			log.Tracef("service «%s» not explicitly set «sendResolved». Using value from defaults: %t", service.Name, *HeartbeatsServer.Notifications.Defaults.SendResolved)
		}

		log.Debugf("service «%s» is enabled: %t", service.Name, *service.Enabled)
	}
	return nil
}

func ProecessHeartbeatsSettings() error {
	for i, h := range HeartbeatsServer.Heartbeats {
		m := make(map[string]string)

		for _, n := range h.Notifications {
			s, err := HeartbeatsServer.GetServiceByName(n)
			if err != nil {
				return err
			}
			// extract everything befor :// from the shoutrrr url
			shoutrrrType := s.Shoutrrr[:strings.Index(s.Shoutrrr, "://")]

			switch shoutrrrType {
			case "discord":
				m[n] = "discord"
			case "smtp":
				m[n] = "email"
			case "gotify":
				m[n] = "gotify"
			case "googlechat":
				m[n] = "googlechat"
			case "ifttt":
				m[n] = "ifttt"
			case "join":
				m[n] = "join"
			case "mattermost":
				m[n] = "mattermost"
			case "matrix":
				m[n] = "matrix"
			case "opsgenie":
				m[n] = "opsgenie"
			case "pushbullet":
				m[n] = "pushbullet"
			case "pushover":
				m[n] = "pushover"
			case "rocketchat":
				m[n] = "rocketchat"
			case "slack":
				m[n] = "slack"
			case "teams":
				m[n] = "msteams"
			case "telegram":
				m[n] = "telegram"
			case "zulip":
				m[n] = "zulip"
			// TODO: webhook
			default:
				return fmt.Errorf("invalid service type in notifications config file")
			}
			HeartbeatsServer.Heartbeats[i].NotificationsMap = m
		}
	}
	return nil
}
