package internal

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/ulule/deepcopier"
)

// ResponseStatus represents a response
type ResponseStatus struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// HeartbeatStatus represents a heartbeat status
type HeartbeatStatus struct {
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	LastPing time.Time `json:"lastPing"`
}

// TimeAgo returns a string representing the time since the given time
func (h *HeartbeatStatus) TimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return CalculateAgo.Format(t)
}

// HandlerHome is the handler for the / endpoint
func HandlerHome(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	html := `<!DOCTYPE html>
<html>
<head>
<title>Heartbeats</title>
</head>
<body>
<h1>Heartbeats</h1>
<p><a href="/config">Config</a></p>
<p><a href="/metrics">Metrics</a></p>
<h2>Status</h2>
<ul>
<li><a href="/status">All</a></li>
{{ range . }}
<li><a href="/status/{{.Name}}">{{.Name}}</a></li>
{{end}}
</ul>
</body>
</html>
`

	s, err := FormatTemplate(html, HeartbeatsServer.Heartbeats)
	if err != nil {
		WriteOutput(w, http.StatusOK, outputFormat, ResponseStatus{Status: "nok", Error: err.Error()}, "Status: {{ .Status }} Error: {{  .Error }}")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(s)); err != nil {
		log.Errorf("Error writing response: %s", err)
	}
}

// HandlerPing is the handler for the /ping/<heartbeat> endpoint
func HandlerPing(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	heartbeat, err := GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusNotFound, outputFormat, err.Error(), "{{ .Status }}")
		return
	}

	heartbeat.GotPing()

	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}

// HandlerPingHelp is the handler for the /ping endpoint
func HandlerPingHelp(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	n := rand.Int() % len(HeartbeatsServer.Heartbeats) // pick a random heartbeat
	usage := struct {
		Status string `json:"status"`
		Usage  string `json:"usage"`
	}{
		Status: "ok",
		Usage:  fmt.Sprintf("You must specify the name of the wanted heartbeat in the URL.\nExample: %s/ping/%s", HeartbeatsServer.Server.SiteRoot, HeartbeatsServer.Heartbeats[n].Name),
	}

	WriteOutput(w, http.StatusOK, outputFormat, &usage, "{{ .Usage }}")
}

// HandlerState is the handler for the /status endpoint
func HandlerStatus(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	var txtFormat string = `Name: {{ .Name }}
Status: {{ if .Status }}{{ .Status }}{{ else }}-{{ end }}
LastPing: {{ .TimeAgo .LastPing }}`

	if heartbeatName == "" {
		textTmpl := fmt.Sprintf("%s%s\n%s", "{{ range . }}", txtFormat, "{{end}}")
		WriteOutput(w, http.StatusOK, outputFormat, &HeartbeatsServer.Heartbeats, textTmpl)
		return
	}

	heartbeat, err := GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusNotFound, outputFormat, ResponseStatus{Status: "nok", Error: err.Error()}, "Status: {{ .Status }} Error: {{  .Error }}")
		return
	}

	heartbeatStates := &HeartbeatStatus{
		Name:     heartbeat.Name,
		Status:   heartbeat.Status,
		LastPing: heartbeat.LastPing,
	}

	WriteOutput(w, http.StatusOK, outputFormat, &heartbeatStates, txtFormat)
}

// HeartbeatsServer is the handler for the /healthz endpoint
func HandlerHealthz(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}

// HandlerConfig is the handler for the /config endpoint
func HandlerConfig(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	// copy the config to avoid exposing the password
	HeartbeatsServerCopy := HeartbeatsConfig{}
	if err := deepcopier.Copy(HeartbeatsServer).To(&HeartbeatsServerCopy); err != nil {
		WriteOutput(w, http.StatusInternalServerError, outputFormat, err.Error(), "{{ .Status }}")
		return
	}

	r, err := RedactServices(HeartbeatsServerCopy.Notifications.Services)
	if err != nil {
		WriteOutput(w, http.StatusInternalServerError, outputFormat, err.Error(), "{{ .Status }}")
		return
	}

	HeartbeatsServerCopy.Notifications.Services = r

	if outputFormat == "txt" || outputFormat == "text" {
		outputFormat = "yaml" // switch to yaml for better output
	}

	WriteOutput(w, http.StatusOK, outputFormat, &HeartbeatsServerCopy, "{{ . }}")
}
