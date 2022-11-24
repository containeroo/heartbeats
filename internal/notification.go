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
			log.Infof("%s got Ping. Status is now «OK»", heartbeatName)
		case GRACE:
			PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(0)
			log.Warnf("%s Grace is expired. Sending notification(s)", heartbeatName)
		case FAILED:
			PromMetrics.HeartbeatStatus.With(prometheus.Labels{"heartbeat": heartbeatName}).Set(0)
			log.Warnf("%s got \"failed\" ping. Sending notification(s)", heartbeatName)
		}
		heartbeat, err := GetHeartbeatByName(heartbeatName)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		heartbeat.Status = status.String()

		for _, notification := range heartbeat.Notifications {
			notificationService, err := HeartbeatsServer.GetServiceByName(notification)
			if err != nil {
				log.Errorf("%s Could not get notification service «%s».", heartbeatName, notification)
				continue
			}
			var serviceName, subject, message string
			var notifier notify.Notifier
			var enabled, sendResolved *bool

			switch service := notificationService.(type) {
			case notifications.SlackSettings:
				enabled = service.Enabled
				serviceName = service.Name
				subject = service.Subject
				message = service.Message
				notifier = service.Notifier
				sendResolved = service.SendResolved

			case notifications.MailSettings:
				enabled = service.Enabled
				serviceName = service.Name
				subject = service.Subject
				message = service.Message
				notifier = service.Notifier
				sendResolved = service.SendResolved

			case notifications.MsteamsSettings:
				enabled = service.Enabled
				serviceName = service.Name
				subject = service.Subject
				message = service.Message
				notifier = service.Notifier
				sendResolved = service.SendResolved

			default:
				log.Errorf("%s Unknown notification type: %v", heartbeatName, notification)
			}

			msg, err := PrepareSend(subject, message, &HeartbeatsServer.Notifications.Defaults, &heartbeat)
			if err != nil {
				log.Errorf("%s Could not prepare message: %s", heartbeatName, err)
				continue
			}

			if !*enabled {
				log.Debugf("%s Skip «%s» because it is disabled", heartbeatName, serviceName)
				continue
			}

			if status == OK && !*sendResolved && !*HeartbeatsServer.Notifications.Defaults.SendResolved {
				log.Debugf("%s Skip «%s» because it is «sendResolved» is disabled", heartbeatName, serviceName)
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
