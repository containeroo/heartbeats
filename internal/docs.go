package internal

import "fmt"

var Documentation Docs
var chapters []string = []string{"config", "api", "endpoints", "heartbeats"}

type Docs struct {
	SiteRoot   string          `json:"siteRoot"`
	Endpoints  []Endpoint      `json:"endpoints"`
	Examples   []Example       `json:"examples"`
	Heartbeats []HeartbeatDocs `json:"heartbeats"`
}

type Endpoint struct {
	Path        string   `json:"path"`
	Methods     []string `json:"method"`
	Description string   `json:"description"`
}

type ResponseCode struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type Example struct {
	Title            string         `json:"title"`
	Code             string         `json:"code"`
	Description      string         `json:"description"`
	QueryParameters  string         `json:"queryParameters"`
	QueryDescription string         `json:"queryDescription"`
	ResponseCodes    []ResponseCode `json:"responseCodes"`
}

type HeartbeatDocs struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

func NewDocumentation(siteRoot string) *Docs {
	d := Docs{
		SiteRoot: siteRoot,
	}
	d.endpoints()
	d.examples()
	d.heartbeat()

	return &d
}

func (d *Docs) heartbeat() {
	heartbeats := []HeartbeatDocs{}

	heartbeat := HeartbeatDocs{
		Key: "enabled",
		Description: `Whether the heartbeat is enabled or not. This is used to calculate the next expected heartbeat.
The enabled must be a boolean. Default is "true".`,
		Example: "true",
	}
	heartbeats = append(heartbeats, heartbeat)

	heartbeat = HeartbeatDocs{
		Key: "name",
		Description: `The name of the heartbeat. This is used to identify the heartbeat in the URL and in the status page.
The name must be unique and can only contain letters, numbers, dashes and underscores.`,
		Example: "my-heartbeat",
	}
	heartbeats = append(heartbeats, heartbeat)

	heartbeat = HeartbeatDocs{
		Key: "description",
		Description: `The description of the heartbeat. This is used to identify the heartbeat in the status page.
The description can be any string.`,
		Example: "My heartbeat",
	}
	heartbeats = append(heartbeats, heartbeat)

	heartbeat = HeartbeatDocs{
		Key: "interval",
		Description: `The interval in seconds between each heartbeat. This is used to calculate the next expected heartbeat.
The interval must be a positive integer.`,
		Example: "60",
	}
	heartbeats = append(heartbeats, heartbeat)

	heartbeat = HeartbeatDocs{
		Key:         "grace",
		Description: "The grace period in seconds before a heartbeat is considered failed. Sometimes a ping can be considered valid if it has a certain amount of \"delay\". The grace must be a positive integer.",
		Example:     "120",
	}
	heartbeats = append(heartbeats, heartbeat)

	d.Heartbeats = heartbeats
}

func (d *Docs) endpoints() {

	endpoints := []Endpoint{}

	endpoint := Endpoint{
		Path:        "/",
		Methods:     []string{"GET"},
		Description: "Home page",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/config",
		Methods:     []string{"GET"},
		Description: "Get the current configuration",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/ping/{heartbeat}",
		Methods:     []string{"GET", "POST"},
		Description: "Reset timer for given heartbeat",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/ping/{heartbeat}/fail",
		Methods:     []string{"GET", "POST"},
		Description: "Set heartbeat to failed",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/status",
		Methods:     []string{"GET"},
		Description: "Get the current status of all heartbeats",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/status/{heartbeat}",
		Methods:     []string{"GET", "POST"},
		Description: "Get the current status of a heartbeat",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/history",
		Methods:     []string{"GET", "POST"},
		Description: "Get the history of all heartbeats",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/history/{heartbeat}",
		Methods:     []string{"GET", "POST"},
		Description: "Get the history of a heartbeat",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/dashboard",
		Methods:     []string{"GET"},
		Description: "Get the dashboard",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/docs",
		Methods:     []string{"GET"},
		Description: "Get the documentation",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/docs/{chapter}",
		Methods:     []string{"GET"},
		Description: "Get the documentation for a chapter",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/metrics",
		Methods:     []string{"GET", "POST"},
		Description: "Get prometheus metrics",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/health",
		Methods:     []string{"GET", "POST"},
		Description: "Get health status",
	}
	endpoints = append(endpoints, endpoint)

	endpoint = Endpoint{
		Path:        "/version",
		Methods:     []string{"GET"},
		Description: "Get version information",
	}
	endpoints = append(endpoints, endpoint)

	d.Endpoints = endpoints
}

func (d *Docs) examples() {
	examples := []Example{}
	example := Example{
		Title:            "Send a heartbeat",
		Code:             fmt.Sprintf("GET|POST %s/ping/{heartbeat}", HeartbeatsServer.Server.SiteRoot),
		Description:      "Sends a \"alive\" message.",
		QueryParameters:  "output=json|yaml|yml|text|txt",
		QueryDescription: "Format response in one of the passed format. If no specific format is passed the response will be <code>text</code>.",
		ResponseCodes: []ResponseCode{
			{
				Code:        "200",
				Description: "OK",
			},
			{
				Code:        "404",
				Description: "heartbeat not found",
			},
			{
				Code:        "503",
				Description: "heartbeat is disabled",
			},
		},
	}
	examples = append(examples, example)

	example = Example{
		Title: "Send a failed heartbeat",
		Code:  fmt.Sprintf("GET|POST %s/ping/{heartbeat}/fail", HeartbeatsServer.Server.SiteRoot),
		Description: "Sends a \"failed\" message. This will set the heartbeat to failed and will not reset the timer. " +
			"Use this if you want to set a heartbeat to failed without resetting the timer.",
		QueryParameters:  "output=json|yaml|yml|text|txt",
		QueryDescription: "Format response in one of the passed format. If no specific format is passed the response will be <code>text</code>.",
		ResponseCodes: []ResponseCode{
			{
				Code:        "200",
				Description: "OK",
			},
			{
				Code:        "404",
				Description: "heartbeat not found",
			},
			{
				Code:        "503",
				Description: "heartbeat is disabled",
			},
		},
	}
	examples = append(examples, example)

	example = Example{
		Title: "Get the current status of a heartbeat",
		Code:  fmt.Sprintf("GET|POST %s/status/{heartbeat}", HeartbeatsServer.Server.SiteRoot),
		Description: "Get the current status of a heartbeat. " +
			"Use this if you want to get the current status of a heartbeat.",
		QueryParameters:  "output=json|yaml|yml|text|txt",
		QueryDescription: "Format response in one of the passed format. If no specific format is passed the response will be <code>text</code>.",
		ResponseCodes: []ResponseCode{
			{
				Code:        "200",
				Description: "OK",
			},
			{
				Code:        "404",
				Description: "heartbeat not found",
			},
		},
	}
	examples = append(examples, example)

	example = Example{
		Title: "Get the current status of all heartbeats",
		Code:  fmt.Sprintf("GET|POST %s/status", HeartbeatsServer.Server.SiteRoot),
		Description: "Get the current status of all heartbeats. " +
			"Use this if you want to get the current status of all heartbeats.",
		QueryParameters:  "output=json|yaml|yml|text|txt",
		QueryDescription: "Format response in one of the passed format. If no specific format is passed the response will be <code>text</code>.",
		ResponseCodes: []ResponseCode{
			{
				Code:        "200",
				Description: "OK",
			},
			{
				Code:        "404",
				Description: "heartbeat not found",
			},
		},
	}
	examples = append(examples, example)

	example = Example{
		Title: "Get the history of a heartbeat",
		Code:  fmt.Sprintf("GET|POST %s/history/{heartbeat}", HeartbeatsServer.Server.SiteRoot),
		Description: "Get the history of a heartbeat. " +
			"Use this if you want to get the history of a heartbeat." +
			fmt.Sprintf("The maximum number of entries is %d and is defined with the parameter <code>--max-size|-m</code>.", HeartbeatsServer.Cache.MaxSize) +
			fmt.Sprintf("If the maximum number of entries is reached the last %d entry will be removed. This is defined with the parameter <code>--reduce|-r</code>.", HeartbeatsServer.Cache.Reduce),
		QueryParameters:  "output=json|yaml|yml|text|txt",
		QueryDescription: "Format response in one of the passed format. If no specific format is passed the response will be <code>text</code>.",
		ResponseCodes: []ResponseCode{
			{
				Code:        "200",
				Description: "OK",
			},
			{
				Code:        "404",
				Description: "heartbeat not found",
			},
		},
	}
	examples = append(examples, example)

	example = Example{
		Title: "Get the history of all heartbeats",
		Code:  fmt.Sprintf("GET|POST %s/history", HeartbeatsServer.Server.SiteRoot),
		Description: "Get the history of all heartbeats. " +
			"Use this if you want to get the history of all heartbeats.",
		QueryParameters:  "output=json|yaml|yml|text|txt",
		QueryDescription: "Format response in one of the passed format. If no specific format is passed the response will be <code>text</code>.",
		ResponseCodes: []ResponseCode{
			{
				Code:        "200",
				Description: "OK",
			},
			{
				Code:        "404",
				Description: "heartbeat not found",
			},
		},
	}
	examples = append(examples, example)

	example = Example{
		Title: "Show the current configuration",
		Code:  fmt.Sprintf("GET|POST %s/config", HeartbeatsServer.Server.SiteRoot),
		Description: "Show the current configuration. " +
			"Use this if you want to see the current configuration.",
		QueryParameters:  "output=json|yaml|yml|text|txt",
		QueryDescription: "Format response in one of the passed format. If no specific format is passed the response will be <code>text</code>.",
		ResponseCodes: []ResponseCode{
			{
				Code:        "200",
				Description: "OK",
			},
			{
				Code:        "404",
				Description: "heartbeat not found",
			},
		},
	}
	examples = append(examples, example)

	example = Example{
		Title: "Show Prometheus metrics",
		Code:  fmt.Sprintf("GET|POST %s/metrics", HeartbeatsServer.Server.SiteRoot),
		Description: "Show Prometheus metrics. " +
			"Use this if you want to see the Prometheus metrics.",
		QueryParameters:  "output=json|yaml|yml|text|txt",
		QueryDescription: "Format response in one of the passed format. If no specific format is passed the response will be <code>text</code>.",
		ResponseCodes: []ResponseCode{
			{
				Code:        "200",
				Description: "OK",
			},
			{
				Code:        "404",
				Description: "heartbeat not found",
			},
		},
	}
	examples = append(examples, example)

	example = Example{
		Title: "Show Heartbeats health",
		Code:  fmt.Sprintf("GET|POST %s/health", HeartbeatsServer.Server.SiteRoot),
		Description: "Show Heartbeats health. " +
			"Use this if you want to see the Heartbeats health.",
		QueryParameters:  "output=json|yaml|yml|text|txt",
		QueryDescription: "Format response in one of the passed format. If no specific format is passed the response will be <code>text</code>.",
		ResponseCodes: []ResponseCode{
			{
				Code:        "200",
				Description: "OK",
			},
			{
				Code:        "404",
				Description: "heartbeat not found",
			},
		},
	}
	examples = append(examples, example)

	example = Example{
		Title: "Show Heartbeats version",
		Code:  fmt.Sprintf("GET|POST %s/version", HeartbeatsServer.Server.SiteRoot),
		Description: "Show Heartbeats version. " +
			"Use this if you want to see the Heartbeats version.",
		QueryParameters:  "output=json|yaml|yml|text|txt",
		QueryDescription: "Format response in one of the passed format. If no specific format is passed the response will be <code>text</code>.",
		ResponseCodes: []ResponseCode{
			{
				Code:        "200",
				Description: "OK",
			},
		},
	}
	examples = append(examples, example)

	d.Examples = examples

}
