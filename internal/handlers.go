package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"
)

// State represents the response
type Status struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// State represents a heardbeat state
type State struct {
	Name     string     `json:"name"`
	Status   string     `json:"status"`
	LastPing *time.Time `json:"lastPing"`
}

// HandlerHome is the handler for the / endpoint
func HandlerHome(w http.ResponseWriter, req *http.Request) {
	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}
	msg := struct{ Message string }{Message: fmt.Sprintf("Welcome to the Heartbeat Server.\nVersion: %s", HeartbeatsServer.Version)}
	WriteOutput(w, http.StatusOK, outputFormat, msg, `Message: {{ .Message }}`)
}

// HandlerPing is the handler for the /ping endpoint
func HandlerPing(w http.ResponseWriter, req *http.Request) {
	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	textTemplate := `{{ .Status }}`

	heartbeat, err := GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusBadRequest, outputFormat, err.Error(), textTemplate)
		return
	}

	heartbeat.GotPing()

	WriteOutput(w, http.StatusOK, outputFormat, &Status{Status: "ok", Error: ""}, textTemplate)
}

// HandlerState is the handler for the /status endpoint
func HandlerStatus(w http.ResponseWriter, req *http.Request) {
	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	if heartbeatName == "" {
		var h []State
		for _, heartbeat := range HeartbeatsServer.Heartbeats {
			s := State{
				Name:     heartbeat.Name,
				Status:   heartbeat.Status,
				LastPing: &heartbeat.LastPing,
			}
			h = append(h, s)
		}

		textTmpl := `{{ range . }}Name: {{ .Name }}
Status: {{ .Status }}
LastPing: {{  .LastPing }}
{{ end }}`
		WriteOutput(w, http.StatusNotFound, outputFormat, h, textTmpl)
		return
	}

	heartbeat, err := GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusNotFound, outputFormat, Status{Status: "nok", Error: err.Error()}, `Status: {{ .Status }} Error: {{  .Error }}`)
		return
	}

	state := &State{
		Name:     heartbeat.Name,
		Status:   heartbeat.Status,
		LastPing: &heartbeat.LastPing,
	}

	txtFormat := `Name: {{ .Name }}
Status: {{ if .Status }}{{ .Status }}{{ else }}-{{ end }}
LastPing: {{if .LastPing.IsZero }}never{{ else }}{{ .LastPing }}{{ end }}`

	WriteOutput(w, http.StatusOK, outputFormat, state, txtFormat)
}

// HeartbeatsServer is the handler for the /healthz endpoint
func HandlerHealthz(w http.ResponseWriter, req *http.Request) {
	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}
	WriteOutput(w, http.StatusOK, outputFormat, &Status{Status: "ok", Error: ""}, `{{ .Status }}`)
}

// WriteOutput writes the output to the response writer
func WriteOutput(w http.ResponseWriter, StatusCode int, outputFormat string, output interface{}, textTemplate string) {
	o, err := FormatOutput(w, outputFormat, output, textTemplate)
	if err != nil {
		w.WriteHeader(StatusCode)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(StatusCode)
	w.Write([]byte(o))
	return
}

// FormatOutput formats the output according to outputFormat
func FormatOutput(w http.ResponseWriter, outputFormat string, output interface{}, textTemplate string) (string, error) {
	switch outputFormat {
	case "json":
		var b bytes.Buffer
		jsonEncoder := json.NewEncoder(&b)
		jsonEncoder.SetIndent("", "  ")
		if err := jsonEncoder.Encode(&output); err != nil {
			return "", err
		}
		return b.String(), nil

	case "yaml", "yml":
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2) // this is what you're looking for
		if err := yamlEncoder.Encode(&output); err != nil {
			return "", err
		}
		return b.String(), nil

	case "txt", "text":
		txt, err := FormatTextOutput(w, outputFormat, textTemplate, output)
		if err != nil {
			return "", fmt.Errorf("Error formatting output")
		}
		return fmt.Sprintf("%+v", txt), nil

	default:
		return "", fmt.Errorf("Output format %s not supported", outputFormat)
	}
}

// FormatTextOutput format given output as text
func FormatTextOutput(w http.ResponseWriter, outputFormat string, textTemplate string, intr interface{}) (string, error) {
	tmpl, err := template.New("status").Parse(textTemplate)
	if err != nil {
		return "", fmt.Errorf("Error executing template: %s", err.Error())
	}
	b := bytes.NewBufferString("")
	err = tmpl.Execute(b, &intr)
	if err != nil {
		return "", fmt.Errorf("Error executing template: %s", err.Error())
	}

	return b.String(), nil
}
