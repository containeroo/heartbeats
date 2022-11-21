package internal

import (
	"context"
	"fmt"

	"github.com/containeroo/heartbeats/internal/notifications"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/nikoksr/notify"

	log "github.com/sirupsen/logrus"
)

// SendDetails holds prepared subject & message
type SendDetails struct {
	Subject string
	Message string
}

// NotificationFunc returns a function that can be used to send notifications
func NotificationFunc(heartbeatName string, success bool) func() {
	return func() {
		var status string
		if success {
			status = "OK"
			PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(1)
			log.Infof("%s got Ping. Status is now «OK»", heartbeatName)
		} else {
			status = "NOK"
			PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(0)
			log.Warnf("%s Grace is expired. Sending notification(s)", heartbeatName)
		}
		heartbeat, err := GetHeartbeatByName(heartbeatName)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		heartbeat.Status = status

		for _, notification := range heartbeat.Notifications {
			notificationService, err := HeartbeatsServer.GetServiceByName(notification)
			if err != nil {
				log.Errorf("%s Could not get notification service «%s».", heartbeatName, notification)
				continue
			}
			var serviceName, subject, message string
			var notifier notify.Notifier
			var enabled, sendResolve *bool

			switch service := notificationService.(type) {
			case notifications.SlackSettings:
				enabled = service.Enabled
				serviceName = service.Name
				subject = service.Subject
				message = service.Message
				notifier = service.Notifier
				sendResolve = service.SendResolve

			case notifications.MailSettings:
				enabled = service.Enabled
				serviceName = service.Name
				subject = service.Subject
				message = service.Message
				notifier = service.Notifier
				sendResolve = service.SendResolve

			case notifications.MsteamsSettings:
				enabled = service.Enabled
				serviceName = service.Name
				subject = service.Subject
				message = service.Message
				notifier = service.Notifier
				sendResolve = service.SendResolve

			default:
				log.Errorf("%s Unknown notification type: %v", heartbeatName, notification)
			}

			msg, err := PrepareSend(subject, message, &HeartbeatsServer.Notifications.Defaults, &heartbeat)
			if err != nil {
				log.Errorf("%s Could not prepare message: %s", heartbeatName, err)
				continue
			}

			if enabled == nil || !*enabled {
				log.Debugf("%s Skip «%s» because it is disabled", heartbeatName, serviceName)
				continue
			}

			// check if we should send a resolve notification
			var r *bool
			if sendResolve == nil {
				r = HeartbeatsServer.Notifications.Defaults.SendResolve
			} else {
				r = sendResolve
			}

			if success && !*r {
				log.Debugf("%s Skip «%s» because it is «sendResolve» is disabled", heartbeatName, serviceName)
				continue
			}

			if err := Send(heartbeatName, serviceName, notifier, msg.Subject, msg.Message); err != nil {
				log.Errorf("%s Could not send notification to «%s». %s", heartbeatName, serviceName, err)
				continue
			}
		}
	}
}

// PrepareSend prepares subject & message to be sent
func PrepareSend(subject string, message string, defaults *Defaults, values interface{}) (SendDetails, error) {

	subject = CheckDefault(subject, defaults.Subject)
	subject, err := FormatTemplate(subject, &values)
	if err != nil {
		return SendDetails{}, fmt.Errorf("Could not substitute subject: %s", err)
	}

	message = CheckDefault(message, defaults.Message)
	message, err = FormatTemplate(message, &values)
	if err != nil {
		return SendDetails{}, fmt.Errorf("Could not substitute message: %s", err)
	}

	return SendDetails{Subject: subject, Message: message}, nil
}

// CheckDefault checks if value is empty and returns default value
func CheckDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// Send sends a notification
func Send(heartbeatName string, serviceName string, notifier notify.Notifier, subject string, message string) error {
	if err := notifier.Send(context.Background(), subject, message); err != nil {
		return err
	}

	if log.GetLevel() == log.DebugLevel {
		log.Debugf("%s Sent notification to «%s» with subject «%s» and message «%s»", heartbeatName, serviceName, subject, message)
	} else {
		log.Infof("%s Sent notification to «%s»", heartbeatName, serviceName)
	}

	return nil
}
