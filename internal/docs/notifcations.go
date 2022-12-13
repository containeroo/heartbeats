package docs

// Notification is the struct for the heartbeats
type Service struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// Notification is the struct for the heartbeats
type Default struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

func (d *Docs) defaults() {
	defaults := []Default{}

	dft := Default{
		Key:         "sendResolved",
		Description: "Whether to send services when a heartbeat is resolved. The sendResolved must be a boolean. Default is \"true\". Can be controlled with services.defaults.sendResolved.",
		Example:     "true",
	}
	defaults = append(defaults, dft)

	dft = Default{
		Key:         "message",
		Description: "The message template to send to the notification service. The message can be any string. Can be controlled with services.defaults.message. Replaces \"bash\" variables ($VAR or ${MY_VAR}) with environment variables. Will be parsed with the heartbeat data if go-template used. See <a href=\"https://golang.org/pkg/text/template/\">here</a> for more information",
		Example:     "The heartbeat {{ .Name }} has failed.",
	}
	defaults = append(defaults, dft)

	d.Defaults = defaults
}

// services returns the services documentation
func (d *Docs) services() {
	services := []Service{}

	service := Service{
		Key:         "name",
		Description: "Name of notification service. This is used to identify the notification service in the Heartbeat.Notification. The name must be unique and can only contain letters, numbers, dashes and underscores.",
		Example:     "mail-provider-x",
	}
	services = append(services, service)

	service = Service{
		Key:         "enabled",
		Description: "Whether the notification service is enabled or not. The enabled must be a boolean. Default is \"true\".",
		Example:     "true",
	}
	services = append(services, service)

	service = Service{
		Key:         "shoutrrr",
		Description: "The Shoutrrr URL to send services to. This is used to send services to the notification service. The Shoutrrr URL must be a valid Shoutrrr URL. See  <a href=\"https://containrrr.github.io/shoutrrr/\">here<a> for more information.",
		Example:     "slack://token@channel",
	}
	services = append(services, service)

	service = Service{
		Key:         "sendResolved",
		Description: "Whether to send services when a heartbeat is resolved. The sendResolved must be a boolean. Default is \"true\". Can be controlled with services.defaults.sendResolved.",
		Example:     "true",
	}
	services = append(services, service)

	service = Service{
		Key:         "message",
		Description: "The message template to send to the notification service. The message can be any string. Can be controlled with services.defaults.message. Replaces \"bash\" variables ($VAR or ${MY_VAR}) with environment variables. Will be parsed with the heartbeat data if go-template used. See <a href=\"https://golang.org/pkg/text/template/\">here</a> for more information",
		Example:     "Heartbeat {{.Name}} is {{.Status}}",
	}
	services = append(services, service)

	d.Services = services
}
