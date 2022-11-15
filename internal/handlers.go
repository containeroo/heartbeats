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
	Name     string `json:"name"`
	Status   string `json:"status"`
	LastPing string `json:"lastPing"`
}

// HandlerHome is the handler for the / endpoint
func HandlerHome(w http.ResponseWriter, req *http.Request) {
	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}

	o := struct{ Message string }{Message: fmt.Sprintf("Welcome to the Heartbeat Server.\nVersion: %s", HeartbeatsServer.Version)}
	if outputFormat == "txt" || outputFormat == "text" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(o.Message))
		return
	}

	WriteOutput(w, http.StatusOK, outputFormat, o)
}

// HandlerPing is the handler for the /ping endpoint
func HandlerPing(w http.ResponseWriter, req *http.Request) {
	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	heartbeat, err := GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusBadRequest, outputFormat, err.Error())
		return
	}

	heartbeat.GotPing()

	if outputFormat == "txt" || outputFormat == "text" {
		w.Write([]byte("ok"))
		return
	}

	WriteOutput(w, http.StatusOK, outputFormat, &Status{Status: "ok", Error: ""})
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
			var lastPing string
			if heartbeat.LastPing.IsZero() {
				lastPing = "never"
			} else {
				lastPing = heartbeat.LastPing.Format(time.RFC1123)
			}
			s := State{
				Name:     heartbeat.Name,
				Status:   heartbeat.Status,
				LastPing: lastPing,
			}
			h = append(h, s)
		}

		if outputFormat == "txt" || outputFormat == "text" {
			state, err := GetStateAsText(w, outputFormat, &h)
			if err != nil {
				WriteOutput(w, http.StatusInternalServerError, outputFormat, err.Error())
				return
			}
			w.Write([]byte(state))
			return
		}

		WriteOutput(w, http.StatusNotFound, outputFormat, &h)
		return
	}

	heartbeat, err := GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusNotFound, outputFormat, &Status{Status: "nok", Error: err.Error()})
		return
	}

	state := &State{
		Name:     heartbeat.Name,
		Status:   heartbeat.Status,
		LastPing: heartbeat.LastPing.Format(time.RFC1123),
	}
	WriteOutput(w, http.StatusOK, outputFormat, state)
}

// HeartbeatsServer is the handler for the /healthz endpoint
func HandlerHealthz(w http.ResponseWriter, req *http.Request) {
	outputFormat := req.URL.Query().Get("output")
	if outputFormat == "" {
		outputFormat = "txt"
	}
	WriteOutput(w, http.StatusOK, outputFormat, &Status{Status: "ok", Error: ""})
}

// GetStateAsText writes the output to the response writer
func GetStateAsText(w http.ResponseWriter, outputFormat string, state *[]State) (string, error) {
	s := `
{{ range . }}
Name: {{ .Name }}
Status: {{ .Status }}
LastPing: {{  .LastPing }}
{{ end }}
`
	tmpl, err := template.New("status").Parse(s)
	if err != nil {
		return "", fmt.Errorf("Error executing template: %s", err.Error())
	}
	b := bytes.NewBufferString("")
	err = tmpl.Execute(b, &state)
	if err != nil {
		return "", fmt.Errorf("Error executing template: %s", err.Error())
	}

	return b.String(), nil
}

// FormatOutput formats the output according to outputFormat
func FormatOutput(w http.ResponseWriter, outputFormat string, output interface{}) (string, error) {
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
		return fmt.Sprintf("%v", output), nil

	default:
		return "", fmt.Errorf("Output format %s not supported", outputFormat)
	}
}

// WriteOutput writes the output to the response writer
func WriteOutput(w http.ResponseWriter, StatusCode int, outputFormat string, status interface{}) {
	output, err := FormatOutput(w, outputFormat, status)
	if err != nil {
		w.WriteHeader(StatusCode)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(StatusCode)
	w.Write([]byte(output))
	return
}
