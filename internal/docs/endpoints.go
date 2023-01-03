package docs

// Endpoint is the struct for the endpoints
type Endpoint struct {
	Path        string   `json:"path"`
	Methods     []string `json:"method"`
	Description string   `json:"description"`
}

// endpoints returns the endpoints documentation
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
