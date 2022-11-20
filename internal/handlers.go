package internal

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/ulule/deepcopier"
)

// StResponseStatus represents the response
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

func (h *HeartbeatStatus) TimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return CalculateAgo.Format(t)
}

// HandlerHome is the handler for the / endpoint
func HandlerHome(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, req.URL.RawQuery)

	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}
	msg := struct{ Message string }{Message: fmt.Sprintf("Welcome to the Heartbeat Server.\nVersion: %s", HeartbeatsServer.Version)}
	WriteOutput(w, http.StatusOK, outputFormat, msg, "Message: {{ .Message }}")
}

// HandlerPing is the handler for the /ping/<heartbeat> endpoint
func HandlerPing(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, req.URL.RawQuery)

	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}

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
	log.Tracef("%s %s%s", req.Method, req.RequestURI, req.URL.RawQuery)

	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}

	n := rand.Int() % len(HeartbeatsServer.Heartbeats)
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
	log.Tracef("%s %s%s", req.Method, req.RequestURI, req.URL.RawQuery)

	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	var txtFormat string = `Name: {{ .Name }}
Status: {{ if .Status }}{{ .Status }}{{ else }}-{{ end }}
LastPing: {{ .TimeAgo .LastPing }}`

	if heartbeatName == "" {
		var h []HeartbeatStatus
		for _, heartbeat := range HeartbeatsServer.Heartbeats {
			s := &HeartbeatStatus{
				Name:     heartbeat.Name,
				Status:   heartbeat.Status,
				LastPing: heartbeat.LastPing,
			}
			h = append(h, *s)
		}

		textTmpl := fmt.Sprintf("%s%s\n%s", "{{ range . }}", txtFormat, "{{end}}")
		WriteOutput(w, http.StatusOK, outputFormat, &h, textTmpl)
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
	log.Tracef("%s %s%s", req.Method, req.RequestURI, req.URL.RawQuery)

	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}
	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}

// HandlerConfig is the handler for the /config endpoint
func HandlerConfig(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, req.URL.RawQuery)

	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}

	// copy the config to avoid exposing the password
	HeartbeatsServerCopy := HeartbeatsConfig{}
	deepcopier.Copy(HeartbeatsServer).To(&HeartbeatsServerCopy)

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

// WriteOutput writes the output to the response writer
func WriteOutput(w http.ResponseWriter, statusCode int, outputFormat string, output interface{}, textTemplate string) {
	formatOutput, err := FormatOutput(outputFormat, textTemplate, output)
	if err != nil {
		w.WriteHeader(statusCode)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Errorf("Cannot write response: %s", err)
		}
		return
	}
	w.WriteHeader(statusCode)
	_, err = w.Write([]byte(formatOutput))
	if err != nil {
		log.Errorf("Cannot write response: %s", err)
	}
	log.Tracef("Server respond with: %d %s", statusCode, formatOutput)
}
