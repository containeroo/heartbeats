package docs

// Heartbeat is the struct for the heartbeats
type Heartbeat struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// heartbeats returns the heartbeats documentation
func (d *Docs) heartbeats() {
	heartbeats := []Heartbeat{}

	heartbeat := Heartbeat{
		Key: "enabled",
		Description: `Whether the heartbeat is enabled or not. This is used to calculate the next expected heartbeat.
The enabled must be a boolean. Default is "true".`,
		Example: "true",
	}
	heartbeats = append(heartbeats, heartbeat)

	heartbeat = Heartbeat{
		Key: "name",
		Description: `The name of the heartbeat. This is used to identify the heartbeat in the URL and in the status page.
The name must be unique and can only contain letters, numbers, dashes and underscores.`,
		Example: "my-heartbeat",
	}
	heartbeats = append(heartbeats, heartbeat)

	heartbeat = Heartbeat{
		Key: "uuid",
		Description: `The UUID of the heartbeat. This is an alternative to the name. This is used to identify the heartbeat in the URL and in the status page.
The UUID must be unique and can only contain letters, numbers, dashes and underscores.`,
		Example: "9e22b12b-a9c0-4820-8e54-1b9e226ff45f",
	}
	heartbeats = append(heartbeats, heartbeat)

	heartbeat = Heartbeat{
		Key: "description",
		Description: `The description of the heartbeat. This is used to identify the heartbeat in the status page.
The description can be any string.`,
		Example: "My heartbeat",
	}
	heartbeats = append(heartbeats, heartbeat)

	heartbeat = Heartbeat{
		Key: "interval",
		Description: `The interval in seconds between each heartbeat. This is used to calculate the next expected heartbeat.
The interval must be a positive duration.`,
		Example: "60",
	}
	heartbeats = append(heartbeats, heartbeat)

	heartbeat = Heartbeat{
		Key: "grace",
		Description: `The grace period in seconds before a heartbeat is considered failed. Sometimes a ping can be considered
valid if it has a certain amount of \"delay\". The grace must be a positive duration.`,
		Example: "120",
	}
	heartbeats = append(heartbeats, heartbeat)

	heartbeat = Heartbeat{
		Key:         "notifications",
		Description: "The notification services to use when the heartbeat interval and grace period has passed. The notifications must be a list of strings. The strings must be match with the name of the notification service. See <a href=\"/docs/notifications\">here</a> for more information.",
		Example:     "- mail-provider-x<br>- slack",
	}
	heartbeats = append(heartbeats, heartbeat)

	d.Heartbeats = heartbeats
}
