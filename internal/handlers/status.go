package handlers

import (
	"fmt"
	"net/http"

	"github.com/containeroo/heartbeats/internal"
	"github.com/gorilla/mux"
)

// Status is the handler for the /status endpoint
func Status(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	outputFormat := GetOutputFormat(req)

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	var txtFormat string = `Name: {{ .Name }}
Status: {{ if .Status }}{{ .Status }}{{ else }}-{{ end }}
LastPing: {{ .TimeAgo .LastPing }}`

	// if no heartbeat is given, return all heartbeats
	if heartbeatName == "" {
		textTmpl := fmt.Sprintf("%s%s\n%s", "{{ range . }}", txtFormat, "{{end}}")
		WriteOutput(w, http.StatusOK, outputFormat, &internal.HeartbeatsServer.Heartbeats, textTmpl)
		return
	}

	var heartbeat *internal.Heartbeat
	var err error

	if IsValidUUID(heartbeatName) {
		heartbeat, err = internal.HeartbeatsServer.GetHeartbeatByUUID(heartbeatName)
	} else {
		heartbeat, err = internal.HeartbeatsServer.GetHeartbeatByName(heartbeatName)
	}
	if err != nil {
		WriteOutput(w, http.StatusNotFound, outputFormat, &ResponseStatus{Status: "nok", Error: err.Error()}, "Status: {{ .Status }}\nError: {{  .Error }}")
		return
	}

	heartbeatStates := &HeartbeatStatus{
		Name:     heartbeat.Name,
		Status:   heartbeat.Status,
		LastPing: heartbeat.LastPing,
	}

	WriteOutput(w, http.StatusOK, outputFormat, &heartbeatStates, txtFormat)
}
