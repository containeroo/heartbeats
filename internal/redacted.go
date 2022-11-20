package internal

import (
	"fmt"
	"reflect"

	"github.com/gi8lino/heartbeats/internal/notifications"
	log "github.com/sirupsen/logrus"
)

// RedactNotifications redacts the fields with the tag "redacted" in the notifications struct
func RedactServices(services []interface{}) ([]interface{}, error) {
	for i, service := range services {
		redactedLabels := []string{"redacted"}

		switch service.(type) {
		case notifications.SlackSettings:
			svc := service.(notifications.SlackSettings)
			svc.Notifier = nil
			for _, name := range GetFieldsByLabel(redactedLabels, svc) {
				reflect.ValueOf(&svc).Elem().FieldByName(name).SetString("<REDACTED>")
			}
			services[i] = svc

		case notifications.MailSettings:
			svc := service.(notifications.MailSettings)
			svc.Notifier = nil
			for _, name := range GetFieldsByLabel(redactedLabels, svc) {
				reflect.ValueOf(&svc).Elem().FieldByName(name).SetString("<REDACTED>")
			}
			services[i] = svc
		default:
			return nil, fmt.Errorf("Unknown notification type")
		}
	}
	return services, nil
}

// GetRedactedFields search in a struct for fields with the tag "redacted" and returns a list with the field names
func GetFieldsByLabel(labels []string, a any) []string {
	var fields []string
	r := reflect.TypeOf(a)
	for i := 0; i < r.NumField(); i++ {
		field := r.Field(i)
		for _, label := range labels {
			t := field.Tag.Get(label)
			if t != "" && t == "true" {
				log.Debugf("Field '%s' is %s", field.Name, label)
				fields = append(fields, field.Name)
			}
		}
	}
	return fields
}
