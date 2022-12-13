package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/containeroo/heartbeats/internal"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// PingHelp is the handler for the /ping endpoint
func PingHelp(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	n := rand.Int() % len(internal.HeartbeatsServer.Heartbeats) // pick a random heartbeat
	usage := struct {
		Status string `json:"status"`
		Usage  string `json:"usage"`
	}{
		Status: "ok",
		Usage:  fmt.Sprintf("You must specify the name of the wanted heartbeat in the URL.\nExample: %s/ping/%s", internal.HeartbeatsServer.Server.SiteRoot, internal.HeartbeatsServer.Heartbeats[n].Name),
	}

	WriteOutput(w, http.StatusOK, outputFormat, &usage, "{{ .Usage }}")
}

// Ping is the handler for the /ping/<heartbeat> endpoint
func Ping(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	heartbeat, err := internal.HeartbeatsServer.GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusNotFound, outputFormat, &ResponseStatus{Status: "nok", Error: err.Error()}, "Status: {{ .Status }}\nError: {{  .Error }}")
		return
	}

	if heartbeat.Enabled != nil && *heartbeat.Enabled == false {
		WriteOutput(w, http.StatusServiceUnavailable, outputFormat, &ResponseStatus{Status: "nok", Error: "heartbeat is disabled"}, "Status: {{ .Status }}\nError: {{  .Error }}")
		return
	}

	heartbeat.GotPing()

	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}

// PingFail is the handler for the /ping/<heartbeat>/fail endpoint
func PingFail(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	heartbeat, err := internal.HeartbeatsServer.GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusServiceUnavailable, outputFormat, &ResponseStatus{Status: "nok", Error: "heartbeat not found"}, "Status: {{ .Status }}\nError: {{  .Error }}")
		return
	}

	if !*heartbeat.Enabled {
		WriteOutput(w, http.StatusServiceUnavailable, outputFormat, &ResponseStatus{Status: "nok", Error: "heartbeat is disabled"}, "Status: {{ .Status }}\nError: {{  .Error }}")
		return
	}

	heartbeat.GotPingFail()

	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}
