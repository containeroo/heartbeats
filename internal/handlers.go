package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Status struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func HandlerHome(w http.ResponseWriter, req *http.Request) {
	text := fmt.Sprintf("Heartbeats is running.\nVersion: %s", HeartbeatsServer.Version)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(text))
}

// HandlerPing is the handler for the /ping endpoint
func HandlerPing(w http.ResponseWriter, req *http.Request) {
	outputType := req.URL.Query().Get("output")
	if outputType == "" {
		outputType = "txt"
	}

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	heartbeat, err := GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, outputType, http.StatusBadRequest, &Status{Status: "nok", Error: err.Error()})
		return
	}

	heartbeat.GotPing()

	WriteOutput(w, outputType, http.StatusOK, &Status{Status: "ok", Error: ""})
}

func HandlerStatus(w http.ResponseWriter, req *http.Request) {
	outputType := req.URL.Query().Get("output")
	if outputType == "" {
		outputType = "txt"
	}

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	heartbeat, err := GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, outputType, http.StatusBadRequest, &Status{Status: "nok", Error: err.Error()})
		return
	}

	e := ""
	if heartbeat.Status != "" && heartbeat.Status != "ok" {
		e = fmt.Sprintf("Last ping was %s is not ok", heartbeat.LastPing)
	}

	WriteOutput(w, outputType, http.StatusOK, &Status{Status: heartbeat.Status, Error: e})
}

// HeartbeatsServer is the handler for the / endpoint
func HandlerHealthz(w http.ResponseWriter, req *http.Request) {
	outputType := req.URL.Query().Get("output")
	if outputType == "" {
		outputType = "txt"
	}
	WriteOutput(w, outputType, http.StatusOK, &Status{Status: "ok", Error: ""})
}

// WriteOutput writes the output to the response writer
func WriteOutput(w http.ResponseWriter, outputFormat string, StatusCode int, status *Status) {
	switch outputFormat {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(StatusCode)
		var b bytes.Buffer
		jsonEncoder := json.NewEncoder(&b)
		jsonEncoder.SetIndent("", "  ")
		if err := jsonEncoder.Encode(&status); err != nil {
			log.Errorf("Error marshalling yaml: %s", err.Error())
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("cannot marshal yaml. %s", err.Error())))
			return
		}
		w.WriteHeader(StatusCode)
		w.Write(b.Bytes())

	case "yaml", "yml":
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2) // this is what you're looking for
		if err := yamlEncoder.Encode(&status); err != nil {
			log.Errorf("Error marshalling yaml: %s", err.Error())
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("cannot marshal yaml. %s", err.Error())))
			return
		}
		w.WriteHeader(StatusCode)
		w.Write(b.Bytes())

	case "txt", "text":
		w.WriteHeader(StatusCode)
		w.Write([]byte(status.Status))

	default:
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid output format"))
		return
	}
}
