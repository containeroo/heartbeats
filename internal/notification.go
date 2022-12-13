package internal

import (
	"os"

	"github.com/containeroo/heartbeats/internal/cache"
	"github.com/containrrr/shoutrrr"
	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"
)

type Notification struct {
	Name         string `mapstructure:"name"`
	Enabled      *bool  `mapstructure:"enabled,omitempty"`
	SendResolved *bool  `mapstructure:"sendResolved,omitempty"`
	Message      string `mapstructure:"message,omitempty"`
	Shoutrrr     string `mapstructure:"shoutrrr"`
}

// Enum for Status
type Status int16

const (
	OK Status = iota
	GRACE
	FAILED
)

func (s Status) String() string {
	return [...]string{"OK", "NOK", "FAILED"}[s]
}

// Notify returns a function that can be used to send notifications
func Notify(heartbeatName string, status Status) func() {
	return func() {
		var msg string
		var history cache.History
		heartbeat, err := HeartbeatsServer.GetHeartbeatByName(heartbeatName)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		switch status {
		case OK:
			PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(1)
			msg = "got ping. Status is now «OK»"
			history = cache.History{
				Event:   cache.EventPing,
				Message: msg,
			}
			log.Infof(msg)
		case GRACE:
			PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(0)
			msg = "Grace is expired. Sending notification(s)"
			history = cache.History{
				Event:   cache.EventGrace,
				Message: msg,
			}
			log.Warnf(msg)
		case FAILED:
			PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(0)
			msg = "got 'failed' ping. Sending notification(s)"
			history = cache.History{
				Event:   cache.EventFailed,
				Message: msg,
			}
			log.Warnf("%s %s", heartbeatName, msg)
		}

		heartbeat.Status = status.String()
		cache.Local.Add(heartbeatName, history)

		for name, service := range heartbeat.NotificationsMap {
			notificationService, err := HeartbeatsServer.GetServiceByName(service)
			if err != nil {
				log.Errorf("%s Could not get notification service «%s».", heartbeatName, name)
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

			url := os.ExpandEnv(notificationService.Shoutrrr) // expand any environment variables
			url, err = FormatTemplate(url, heartbeat)         // expand any template variables
			if err != nil {
				log.Errorf("%s Could not format shoutrrr url «%s» for «%s». %s", heartbeatName, notificationService.Shoutrrr, notificationService.Name, err)
				continue
			}

			message := CheckDefault(notificationService.Message, HeartbeatsServer.Notifications.Defaults.Message)
			message = os.ExpandEnv(message)                   // expand any environment variables
			message, err = FormatTemplate(message, heartbeat) // expand any template variables
			if err != nil {
				log.Errorf("%s Could not format message «%s» for «%s». %s", heartbeatName, notificationService.Message, notificationService.Name, err)
			}

			if err := shoutrrr.Send(url, message); err != nil {
				log.Errorf("%s Could not send notification to «%s». %s", heartbeatName, notificationService.Name, err)
				continue
			}
			cache.Local.Add(heartbeatName, cache.History{
				Event:   cache.EventSend,
				Message: "Sent notification to " + notificationService.Name,
			})
		}
	}
}
