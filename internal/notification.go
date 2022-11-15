package internal

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/gi8lino/heartbeats/internal/notifications"

	"github.com/nikoksr/notify"

	log "github.com/sirupsen/logrus"
)

type SendDetails struct {
	Subject string
	Message string
}

// NotificationFunc returns a function that can be used to send notifications
func NotificationFunc(heartbeatName string, success bool) func() {
	return func() {
		var status string
		if success {
			log.Infof("%s got Ping. Status is now «OK»", heartbeatName)
			status = "OK"
		} else {
			log.Warnf("%s Grace is expired. Sending notification(s)", heartbeatName)
			status = "NOK"
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
			var enabled bool

			switch notificationService.(type) {
			case notifications.SlackSettings:
				service := notificationService.(notifications.SlackSettings)
				serviceName = service.Name
				subject = service.Subject
				message = service.Message
				notifier = service.Notifier
				enabled = service.Enabled

			case notifications.MailSettings:
				service := notificationService.(notifications.MailSettings)
				serviceName = service.Name
				subject = service.Subject
				message = service.Message
				notifier = service.Notifier
				enabled = service.Enabled

			default:
				log.Errorf("%s Unknown notification type: %v", heartbeatName, notification)
			}

			msg, err := PrepareSend(subject, message, &HeartbeatsServer.Notifications.Defaults, &heartbeat)
			if err != nil {
				log.Errorf("%s Could not prepare message: %s", heartbeatName, err)
				continue
			}

			if !enabled {
				log.Debugf("%s Skip «%s» because it is disabled", heartbeatName, serviceName)
				continue
			}
			if err := Send(heartbeatName, serviceName, notifier, msg.Subject, msg.Message); err != nil {
				log.Errorf("%s Could not send notification to «%s»", heartbeatName, serviceName, err)
				continue
			}
		}
	}
}

// GetServiceByType returns notification settings by type
func (h *HeartbeatsConfig) GetServiceByName(notificationType string) (interface{}, error) {
	for _, notification := range h.Notifications.Services {
		switch notification.(type) {
		case notifications.SlackSettings:
			service := notification.(notifications.SlackSettings)
			if strings.EqualFold(service.Name, notificationType) {
				return service, nil
			}
		case notifications.MailSettings:
			service := notification.(notifications.MailSettings)
			if strings.EqualFold(service.Name, notificationType) {
				return service, nil
			}
		default:
			return nil, fmt.Errorf("Unknown notification type")
		}
	}
	return nil, fmt.Errorf("Notification settings for type «%s» not found", notificationType)
}

// PrepareSend prepares subject & message to be sent
func PrepareSend(subject string, message string, defaults *Defaults, values interface{}) (SendDetails, error) {

	subject = CheckDefault(subject, defaults.Subject)
	subject, err := Substitute("subject", subject, &values)
	if err != nil {
		return SendDetails{}, fmt.Errorf("Could not substitute subject: %s", err)
	}

	message = CheckDefault(message, defaults.Message)
	message, err = Substitute("message", message, &values)
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

// Substitute substitutes values in tmpl
func Substitute(title, tmpl string, values interface{}) (string, error) {
	t, err := template.New(title).Parse(tmpl)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	if err := t.Execute(buf, &values); err != nil {
		return "", err
	}

	return buf.String(), nil
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
