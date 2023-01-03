package internal

import (
	"fmt"
	"os"

	"github.com/containeroo/heartbeats/internal/cache"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/containeroo/heartbeats/internal/utils"
	"github.com/containrrr/shoutrrr"
	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"
)

// Service represents a notification service
type Service struct {
	Name         string `mapstructure:"name"`
	Enabled      *bool  `mapstructure:"enabled,omitempty"`
	SendResolved *bool  `mapstructure:"sendResolved,omitempty"`
	Message      string `mapstructure:"message,omitempty"`
	Shoutrrr     string `mapstructure:"shoutrrr"`
}

// Enum for Status
type Status int16

// Enum values for Status
const (
	OK Status = iota
	GRACE
	FAILED
)

// String returns the string representation of the Status
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
			metrics.PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(1)
			msg = "got ping. Status is now «OK»"
			history = cache.History{
				Event:   cache.EventPing,
				Message: msg,
			}
			log.Infof(msg)
		case GRACE:
			metrics.PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(0)
			msg = "Grace is expired. Sending notification(s)"
			history = cache.History{
				Event:   cache.EventGrace,
				Message: msg,
			}
			log.Warnf(msg)
		case FAILED:
			metrics.PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(0)
			msg = "got 'failed' ping. Sending notification(s)"
			history = cache.History{
				Event:   cache.EventFailed,
				Message: msg,
			}
			log.Warnf("%s %s", heartbeatName, msg)
		}

		heartbeat.Status = status.String()
		cache.Local.Add(heartbeatName, history)

		for _, notification := range heartbeat.NotificationsMap {
			notificationService, err := HeartbeatsServer.GetServiceByName(notification.Name)
			if err != nil {
				log.Errorf("%s Could not get %s-notification service «%s». %s", heartbeatName, notification.Type, notificationService.Name, err.Error())
				continue
			}

			// check if service.enabled is set and if enabled is set to false
			if !*notificationService.Enabled {
				log.Debugf("%s Skip %s-notification «%s» because it is disabled", heartbeatName, notification.Type, notificationService.Name)
				continue
			}

			// check if service.sendResolved is set and if sendResolved is set to false
			if status == OK && notificationService.SendResolved != nil && !*notificationService.SendResolved && !*HeartbeatsServer.Notifications.Defaults.SendResolved {
				log.Debugf("%s Skip %s-notification «%s» because «sendResolved» is disabled", heartbeatName, notification.Type, notificationService.Name)
				continue
			}

			url := os.ExpandEnv(notificationService.Shoutrrr) // expand any environment variables
			url, err = utils.FormatTemplate(url, heartbeat)   // expand any template variables
			if err != nil {
				log.Errorf("%s Could not format shoutrrr url «%s» for «%s». %s", heartbeatName, notificationService.Shoutrrr, notificationService.Name, err)
				continue
			}

			message := utils.CheckDefault(notificationService.Message, HeartbeatsServer.Notifications.Defaults.Message)
			message = os.ExpandEnv(message)                         // expand any environment variables
			message, err = utils.FormatTemplate(message, heartbeat) // expand any template variables
			if err != nil {
				log.Errorf("%s Could not format message «%s» for «%s». %s", heartbeatName, notificationService.Message, notificationService.Name, err)
			}

			if err := shoutrrr.Send(url, message); err != nil {
				log.Errorf("%s Could not send %s-notification to «%s». %s", heartbeatName, notification.Type, notificationService.Name, err)
				continue
			}
			msg := fmt.Sprintf("Sent %s-notification to «%s»", notification.Type, notificationService.Name)
			log.Debug(msg)
			cache.Local.Add(heartbeatName, cache.History{
				Event:   cache.EventSend,
				Message: msg,
			})
		}
	}
}
