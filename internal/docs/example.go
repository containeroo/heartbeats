package docs

import "fmt"

// ResponseCode is the struct for the response codes
type ResponseCode struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// Example is the struct for the examples
type Example struct {
	Title            string         `json:"title"`
	Code             string         `json:"code"`
	Description      string         `json:"description"`
	QueryParameters  string         `json:"queryParameters"`
	QueryDescription string         `json:"queryDescription"`
	ResponseCodes    []ResponseCode `json:"responseCodes"`
}

// exmaples returns the examples documentation
func (d *Docs) examples() {
	examples := []Example{}
	example := Example{
		Title:            "Send a heartbeat",
		Code:             fmt.Sprintf("GET|POST %s/ping/{heartbeat}", d.SiteRoot),
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
		Code:  fmt.Sprintf("GET|POST %s/ping/{heartbeat}/fail", d.SiteRoot),
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
		Code:  fmt.Sprintf("GET|POST %s/status/{heartbeat}", d.SiteRoot),
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
		Code:  fmt.Sprintf("GET|POST %s/status", d.SiteRoot),
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
		Code:  fmt.Sprintf("GET|POST %s/history/{heartbeat}", d.SiteRoot),
		Description: "Get the history of a heartbeat. " +
			"Use this if you want to get the history of a heartbeat." +
			fmt.Sprintf("The maximum number of entries is %d and is defined with the parameter <code>--max-size|-m</code>.", d.Cache.MaxSize) +
			fmt.Sprintf("If the maximum number of entries is reached the last %d entry will be removed. This is defined with the parameter <code>--reduce|-r</code>.", d.Cache.Reduce),
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
		Code:  fmt.Sprintf("GET|POST %s/history", d.SiteRoot),
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
		Code:  fmt.Sprintf("GET|POST %s/config", d.SiteRoot),
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
		Code:  fmt.Sprintf("GET|POST %s/metrics", d.SiteRoot),
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
		Code:  fmt.Sprintf("GET|POST %s/health", d.SiteRoot),
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
		Code:  fmt.Sprintf("GET|POST %s/version", d.SiteRoot),
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
