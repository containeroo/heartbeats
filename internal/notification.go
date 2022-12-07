package internal

import (
	"os"

	"github.com/containrrr/shoutrrr"
	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"
)

type Notification struct {
	Name         string `mapstructure:"name"`
	Enabled      *bool  `mapstructure:"enabled,omitempty"`
	SendResolved *bool  `mapstructure:"sendResolved,omitempty"`
	Message      string `mapstructure:"message,omitempty"`
	URL          string `mapstructure:"shoutrrr"`
}

// Enum for Status with ok, nok and failed
type Status int16

const (
	OK Status = iota
	GRACE
	FAILED
)

func (s Status) String() string {
	return [...]string{"OK", "NOK", "FAILED"}[s]
}

// NotificationFunc returns a function that can be used to send notifications
func NotificationFunc(heartbeatName string, status Status) func() {
	return func() {
		switch status {
		case OK:
			PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(1)
			log.Infof("%s got ping. Status is now «OK»", heartbeatName)
		case GRACE:
			PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(0)
			log.Warnf("%s Grace is expired. Sending notification(s)", heartbeatName)
		case FAILED:
			PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(0)
			log.Warnf("%s got \"failed\" ping. Sending notification(s)", heartbeatName)
		}
		heartbeat, err := HeartbeatsServer.GetHeartbeatByName(heartbeatName)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		heartbeat.Status = status.String()

		for _, service := range heartbeat.Notifications {
			notificationService, err := HeartbeatsServer.GetServiceByName(service)
			if err != nil {
				log.Errorf("%s Could not get notification service «%s».", heartbeatName, service)
				continue
			}

			// check if service.enabled is set and if enabled is set to false
			if notificationService.Enabled != nil && !*notificationService.Enabled {
				log.Debugf("%s Skip «%s» because it is disabled", heartbeatName, notificationService.Name)
				continue
			}

			// check if service.sendResolved is set and if sendResolved is set to false
			if status == OK && notificationService.SendResolved != nil && !*notificationService.SendResolved && !*HeartbeatsServer.Notifications.Defaults.SendResolved {
				log.Debugf("%s Skip «%s» because «sendResolved» is disabled", heartbeatName, notificationService.Name)
				continue
			}

			url := os.ExpandEnv(notificationService.URL) // expand any environment variables
			url, err = FormatTemplate(url, heartbeat)    // expand any template variables
			if err != nil {
				log.Errorf("%s Could not format shoutrrr url «%s» for «%s». %s", heartbeatName, notificationService.URL, notificationService.Name, err)
				continue
			}

			message := CheckDefault(notificationService.Message, HeartbeatsServer.Notifications.Defaults.Message)
			message = os.ExpandEnv(message)                   // expand any environment variables
			message, err = FormatTemplate(message, heartbeat) // expand any template variables
			if err != nil {
				log.Errorf("%s Could not format message «%s» for «%s». %s", heartbeatName, notificationService.Message, notificationService.Name, err)
				continue
			}

			if err := shoutrrr.Send(url, message); err != nil {
				log.Errorf("%s Could not send notification to «%s». %s", heartbeatName, notificationService.Name, err)
				continue
			}
		}
	}
}
