package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/containeroo/heartbeats/internal/notifications"

	"github.com/mitchellh/mapstructure"
	"github.com/nikoksr/notify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var HeartbeatsServer HeartbeatsConfig

const (
	EnvPrefix = "env:"
)

// Config config holds general configuration
type Config struct {
	Path         string `mapstructure:"path"`
	PrintVersion bool   `mapstructure:"printVersion"`
	Logging      string `mapstructure:"logging"`
}

// Details details holds defaults for notifications
type Defaults struct {
	Subject     string `mapstructure:"subject" default:"Heartbeat"`
	Message     string `mapstructure:"message" default:"Heartbeat is missing"`
	SendResolve *bool  `mapstructure:"sendResolve" default:"true"`
}

// NotifyConfig holds the configuration for the notifications
type Notifications struct {
	Defaults Defaults      `mapstructure:"defaults"`
	Services []interface{} `mapstructure:"services"`
}

// HeartbeatsConfig is the main configuration struct
type HeartbeatsConfig struct {
	Version       string        `mapstructure:"version"`
	Config        Config        `mapstructure:"config"`
	Server        Server        `mapstructure:"server"`
	Heartbeats    []Heartbeat   `mapstructure:"heartbeats"`
	Notifications Notifications `mapstructure:"notifications"`
}

// ReadConfigFile reads the notifications config file
func ReadConfigFile(configPath string) error {
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
	if HeartbeatsServer.Heartbeats != nil {
		previousHeartbeats = HeartbeatsServer.Heartbeats
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
		return fmt.Errorf("error while processing notification services: %s", err)
	}

	if err := CheckSendDetails(); err != nil {
		return err
	}

	ResetTimerIfRunning(&previousHeartbeats)

	return nil
}

// ResetTimerIfRunning resets existing timers if they are running
func ResetTimerIfRunning(previousHeartbeats *[]Heartbeat) {
	for _, currentHeartbeat := range HeartbeatsServer.Heartbeats {

		var previousHeartbeat *Heartbeat
		var err error
		for _, p := range *previousHeartbeats {
			previousHeartbeat, err = GetHeartbeatByName(p.Name)
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

// ProcessNotificationSettings processes the list with notifications
func ProcessServiceSettings() error {
	for i, service := range HeartbeatsServer.Notifications.Services {
		var serviceType string

		// this is needed because the type of the service is not known when config is read
		switch service.(type) {
		case map[string]interface{}:
			s, ok := service.(map[string]interface{})["type"].(string)
			if !ok {
				return fmt.Errorf("type of service %s is not set", service.(map[string]interface{})["name"])
			}
			serviceType = s
		case notifications.SlackSettings:
			serviceType = "slack"
		case notifications.MailSettings:
			serviceType = "mail"
		case notifications.MsteamsSettings:
			serviceType = "msteams"
		default:
			return fmt.Errorf("invalid service type in notifications config file")
		}

		// now the type is known and the service can be processed
		switch serviceType {
		case "slack":
			var result notifications.SlackSettings
			if err := mapstructure.Decode(service, &result); err != nil {
				return err
			}

			for name, value := range SubstituteFieldsWithEnv(EnvPrefix, result) {
				reflect.ValueOf(&result).Elem().FieldByName(name).Set(value)
			}

			svc, err := notifications.GenerateSlackService(result.OauthToken, result.Channels)
			if err != nil {
				return fmt.Errorf("error while generating slack service: %s", err)
			}
			result.Notifier = notify.New()
			result.Notifier.UseServices(svc)

			HeartbeatsServer.Notifications.Services[i] = result

			log.Debugf("Slack service «%s» is enabled: %t", result.Name, result.Enabled)

		case "mail":
			var result notifications.MailSettings
			if err := mapstructure.Decode(service, &result); err != nil {
				return err
			}

			for name, value := range SubstituteFieldsWithEnv(EnvPrefix, result) {
				reflect.ValueOf(&result).Elem().FieldByName(name).Set(value)
			}

			svc, err := notifications.GenerateMailService(result.SenderAddress, result.SmtpHostAddr, result.SmtpHostPort, result.SmtpAuthUser, result.SmtpAuthPassword, result.ReceiverAddresses)
			if err != nil {
				return fmt.Errorf("error while generating mail service: %s", err)
			}
			result.Notifier = notify.New()
			result.Notifier.UseServices(svc)

			HeartbeatsServer.Notifications.Services[i] = result

			log.Debugf("Mail service «%s» is enabled: %t", result.Name, result.Enabled)

		case "msteams":
			var result notifications.MsteamsSettings
			if err := mapstructure.Decode(service, &result); err != nil {
				return err
			}

			for name, value := range SubstituteFieldsWithEnv(EnvPrefix, result) {
				reflect.ValueOf(&result).Elem().FieldByName(name).Set(value)
			}

			svc, err := notifications.GenerateMsteamsService(result.WebHooks)
			if err != nil {
				return fmt.Errorf("error while generating MS Teams service: %s", err)
			}
			result.Notifier = notify.New()
			result.Notifier.UseServices(svc)

			HeartbeatsServer.Notifications.Services[i] = result

			log.Debugf("MS Teams service «%s» is enabled: %t", result.Name, result.Enabled)

		default:
			return fmt.Errorf("Unknown notification service type")
		}
	}
	return nil
}

// CheckSendDetails checks if the send details are set and parsing is possible
func CheckSendDetails() error {
	var heartbeat Heartbeat

	// check defaults
	if HeartbeatsServer.Notifications.Defaults.Subject == "" {
		return fmt.Errorf("default subject is not set")
	}
	if _, err := FormatTemplate(HeartbeatsServer.Notifications.Defaults.Subject, &heartbeat); err != nil {
		return err
	}

	if _, err := FormatTemplate(HeartbeatsServer.Notifications.Defaults.Message, &heartbeat); err != nil {
		return err
	}

	if HeartbeatsServer.Notifications.Defaults.Message == "" {
		log.Warnf("default message is not set")
	}

	var name, subject, message string

	for _, notification := range HeartbeatsServer.Notifications.Services {
		switch settings := notification.(type) {
		case notifications.SlackSettings:
			name = settings.Name
			subject = settings.Subject
			message = settings.Message
		case notifications.MailSettings:
			name = settings.Name
			subject = settings.Subject
			message = settings.Message
		case notifications.MsteamsSettings:
			name = settings.Name
			subject = settings.Subject
			message = settings.Message
		default:
			return fmt.Errorf("invalid service type in notifications config file")
		}

		if _, err := FormatTemplate(subject, &heartbeat); err != nil {
			log.Warnf("service «%s». Cannot evaluate subject: %s", name, err)
		}
		if _, err := FormatTemplate(message, &heartbeat); err != nil {
			log.Warnf("service «%s». Cannot evaluate message: %s", name, err)
		}
	}
	return nil
}

// SubstituteFieldsWithEnv searches in env for the given key and replaces the value with the value from env
func SubstituteFieldsWithEnv(prefix string, a any) map[string]reflect.Value {
	result := make(map[string]reflect.Value)

	r := reflect.TypeOf(a)
	for i := 0; i < r.NumField(); i++ {
		field := r.Field(i)
		// get field value
		value := reflect.ValueOf(a).FieldByName(field.Name)
		if !strings.HasPrefix(value.String(), prefix) {
			continue
		}

		envValue := os.Getenv(value.String()[len(prefix):])
		if envValue != "" {
			reflectedValue := reflect.ValueOf(envValue)
			result[field.Name] = reflectedValue
		}
	}
	return result
}
