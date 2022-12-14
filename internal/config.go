package internal

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containeroo/heartbeats/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var HeartbeatsServer Heartbeats
var ConfigCopy Heartbeats
var StaticFS embed.FS

// NotifyConfig holds the configuration for the notifications
type Notifications struct {
	Defaults Defaults  `mapstructure:"defaults"`
	Services []Service `mapstructure:"services"`
}

// Details details holds defaults for notifications
type Defaults struct {
	Subject      string `mapstructure:"subject"`
	Message      string `mapstructure:"message"`
	SendResolved *bool  `mapstructure:"sendResolved"`
}

// Cache holds the configuration for the cache
type Cache struct {
	MaxSize int `mapstructure:"maxSize"`
	Reduce  int `mapstructure:"reduce"`
}

// Server is the holds HTTP server settings
type Server struct {
	Hostname string `mapstructure:"hostname"`
	Port     int    `mapstructure:"port"`
	SiteRoot string `mapstructure:"siteRoot"`
}

// Config config holds general configuration
type Config struct {
	Path    string `mapstructure:"path"`
	Logging string `mapstructure:"logging"`
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

	if err := ProcessHeartbeatsSettings(); err != nil {
		return err
	}

	// check if reduce is bigger than MaxSize
	if HeartbeatsServer.Cache.Reduce > HeartbeatsServer.Cache.MaxSize {
		return fmt.Errorf("reduce is bigger than maxSize")
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
		t := true
		HeartbeatsServer.Notifications.Defaults.SendResolved = &t
		log.Tracef("Default «sendResolved» not explicitly enabled. Defaulting to true")
	}

	if HeartbeatsServer.Notifications.Defaults.Message == "" {
		HeartbeatsServer.Notifications.Defaults.Message = "Heartbeat «{{.Name}}» is {{.Status}}"
		log.Tracef("Default «message» not explicitly set. Defaulting to «%s»", HeartbeatsServer.Notifications.Defaults.Message)
	}

	for _, service := range HeartbeatsServer.Notifications.Services {
		url := os.ExpandEnv(service.Shoutrrr)
		_, err := utils.FormatTemplate(url, &heartbeat) //test if there is an error when expanding any template variables
		if err != nil {
			return fmt.Errorf("Could not format shoutrrr url «%s» for «%s». %s", service.Shoutrrr, service.Name, err)
		}

		message := utils.CheckDefault(service.Message, HeartbeatsServer.Notifications.Defaults.Message)
		message = os.ExpandEnv(message)                    // expand any environment variables
		_, err = utils.FormatTemplate(message, &heartbeat) // expand any template variables
		if err != nil {
			return fmt.Errorf("Could not format message «%s» for «%s». %s", service.Message, service.Name, err)
		}

		if service.Enabled == nil {
			t := true
			service.Enabled = &t // must be set, otherwise the debug message below will fail because of a nil pointer
			log.Tracef("service «%s» not explicitly enabled. Defaulting to true", service.Name)
		}

		if service.SendResolved == nil {
			log.Tracef("service «%s» not explicitly set «sendResolved». Using value from defaults: %t", service.Name, *HeartbeatsServer.Notifications.Defaults.SendResolved)
		}

		log.Debugf("service «%s» is enabled: %t", service.Name, *service.Enabled)
	}
	return nil
}

// ProcessHeartbeatsSettings checks if all heartbeats are valid and can be parsed
func ProcessHeartbeatsSettings() error {
	supportedServices := []string{"discord", "smtp", "gotify", "googlechat", "ifttt", "join", "mattermost", "matrix", "pushover", "rocketchat", "slack", "teams", "telegram", "zulip", "webhook"}
	for i, h := range HeartbeatsServer.Heartbeats {
		// reset HeartbeatsServer.Heartbeats[i].NotificationsMap to avoid duplicates
		HeartbeatsServer.Heartbeats[i].NotificationsMap = []NotificationsMap{}

		for _, n := range h.Notifications {
			s, err := HeartbeatsServer.GetServiceByName(n)
			if err != nil {
				return err
			}
			// extract everything befor :// from the shoutrrr url
			shoutrrrType := s.Shoutrrr[:strings.Index(s.Shoutrrr, "://")]

			if !utils.IsInListOfStrings(supportedServices, shoutrrrType) {
				return fmt.Errorf("service «%s» is not supported", shoutrrrType)
			}

			if s.Enabled == nil {
				trueVar := true
				s.Enabled = &trueVar
			}

			n := NotificationsMap{
				Name:    n,
				Type:    shoutrrrType,
				Enabled: *s.Enabled,
			}
			HeartbeatsServer.Heartbeats[i].NotificationsMap = append(HeartbeatsServer.Heartbeats[i].NotificationsMap, n)
		}
	}
	return nil
}
