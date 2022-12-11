package internal

import "fmt"

type Docs struct {
	SiteRoot  string     `json:"siteRoot"`
	Endpoints []Endpoint `json:"endpoints"`
	Examples  []Example  `json:"examples"`
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

func (d *Docs) Init() {
	d.SiteRoot = HeartbeatsServer.Server.SiteRoot
	d.endpoints()
	d.examples()

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
		Code:             fmt.Sprintf("GET|POST %s/ping/heartbeat1", HeartbeatsServer.Server.SiteRoot),
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
		},
	}

	examples = append(examples, example)
	example = Example{
		Title: "Send a failed heartbeat",
		Code:  fmt.Sprintf("GET|POST %s/ping/heartbeat1/fail", HeartbeatsServer.Server.SiteRoot),
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
		},
	}
	examples = append(examples, example)

	example = Example{
		Title: "Get the current status of a heartbeat",
		Code:  fmt.Sprintf("GET|POST %s/status/heartbeat1", HeartbeatsServer.Server.SiteRoot),
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
		Code:  fmt.Sprintf("GET|POST %s/history/heartbeat1", HeartbeatsServer.Server.SiteRoot),
		Description: "Get the history of a heartbeat. " +
			"Use this if you want to get the history of a heartbeat.",
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