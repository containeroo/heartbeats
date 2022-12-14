package handlers

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/containeroo/heartbeats/internal"
	"github.com/gorilla/mux"
)

// PingHelp is the handler for the /ping endpoint
func PingHelp(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	outputFormat := GetOutputFormat(req)

	n := rand.Int() % len(internal.HeartbeatsServer.Heartbeats) // pick a random heartbeat
	usage := struct {
		Status string `json:"status"`
		Usage  string `json:"usage"`
	}{
		Status: "nok",
		Usage:  fmt.Sprintf("You must specify the name or the uuid of the wanted heartbeat in the URL.\nExamples: %s/ping/%s", internal.HeartbeatsServer.Server.SiteRoot, internal.HeartbeatsServer.Heartbeats[n].Name),
	}

	WriteOutput(w, http.StatusNotFound, outputFormat, &usage, "{{ .Usage }}")
}

// Ping is the handler for the /ping/<heartbeat> endpoint
func Ping(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	outputFormat := GetOutputFormat(req)

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

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

	if heartbeat.Enabled != nil && !*heartbeat.Enabled {
		WriteOutput(w, http.StatusServiceUnavailable, outputFormat, &ResponseStatus{Status: "nok", Error: "heartbeat is disabled"}, "Status: {{ .Status }}\nError: {{  .Error }}")
		return
	}

	details := map[string]string{
		"proto":     req.Proto,
		"clientIP":  req.RemoteAddr,
		"method":    req.Method,
		"userAgent": req.UserAgent(),
	}

	heartbeat.GotPing(details)

	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}

// PingFail is the handler for the /ping/<heartbeat>/fail endpoint
func PingFail(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	outputFormat := GetOutputFormat(req)

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	var heartbeat *internal.Heartbeat
	var err error

	if IsValidUUID(heartbeatName) {
		heartbeat, err = internal.HeartbeatsServer.GetHeartbeatByUUID(heartbeatName)
	} else {
		heartbeat, err = internal.HeartbeatsServer.GetHeartbeatByName(heartbeatName)
	}
	if err != nil {
		WriteOutput(w, http.StatusServiceUnavailable, outputFormat, &ResponseStatus{Status: "nok", Error: "heartbeat not found"}, "Status: {{ .Status }}\nError: {{  .Error }}")
		return
	}

	if heartbeat.Enabled != nil && !*heartbeat.Enabled {
		WriteOutput(w, http.StatusServiceUnavailable, outputFormat, &ResponseStatus{Status: "nok", Error: "heartbeat is disabled"}, "Status: {{ .Status }}\nError: {{  .Error }}")
		return
	}

	details := map[string]string{
		"proto":     req.Proto,
		"clientIP":  req.RemoteAddr,
		"method":    req.Method,
		"userAgent": req.UserAgent(),
	}

	heartbeat.GotPingFail(details)

	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}
