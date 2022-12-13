package handlers

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/containeroo/heartbeats/internal"
	"github.com/containeroo/heartbeats/internal/cache"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// outputHistory outputs history in wanted format
func outputHistory(w http.ResponseWriter, req *http.Request, outputFormat string, heartbeatName string) {
	if heartbeatName == "" {
		WriteOutput(w, http.StatusOK, outputFormat, &cache.Local.History, "{{ . }}")
		return
	}

	heartbeat, err := internal.HeartbeatsServer.GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusNotFound, outputFormat, ResponseStatus{Status: "nok", Error: err.Error()}, "Status: {{ .Status }}\nError: {{  .Error }}")
		return
	}
	histories, ok := cache.Local.History[heartbeat.Name]
	if !ok {
		WriteOutput(w, http.StatusNotFound, outputFormat, ResponseStatus{Status: "nok", Error: "No history found"}, "Status: {{ .Status }}\nError: {{  .Error }}")
		return
	}

	WriteOutput(w, http.StatusOK, outputFormat, &histories, "{{ . }}")
}

// HandlerHistory is the handler for the /history endpoint
func History(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	if outputFormat := req.URL.Query().Get("output"); outputFormat != "" {
		outputHistory(w, req, outputFormat, heartbeatName)
		return
	}

	var templs []string

	if heartbeatName == "" {
		templs = []string{
			"web/templates/base.html",
			"web/templates/navbar.html",
			"web/templates/history_all.html",
			"web/templates/footer.html",
		}
	} else {
		templs = []string{
			"web/templates/base.html",
			"web/templates/navbar.html",
			"web/templates/history.html",
			"web/templates/footer.html",
		}
	}

	tmpl, err := template.ParseFS(internal.StaticFS, templs...)
	if err != nil {
		log.Errorf("Error parsing template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(fmt.Sprintf("cannot parse template. %s", err.Error())))
		if err != nil {
			log.Errorf("Error writing response: %s", err.Error())
		}
		return
	}

	if heartbeatName == "" {
		err = tmpl.ExecuteTemplate(w, "base", cache.Local)
	} else {
		history, err := cache.Local.Get(heartbeatName)
		if err != nil {
			log.Warnf("Error getting history: %s", err.Error())
		}
		h := struct {
			Name    string
			History *[]cache.History
		}{
			Name:    heartbeatName,
			History: &history,
		}
		err = tmpl.ExecuteTemplate(w, "base", h)
	}
	if err != nil {
		log.Errorf("Error executing template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(fmt.Sprintf("cannot execute template. %s", err.Error())))
		if err != nil {
			log.Errorf("Error writing response: %s", err.Error())
		}
		return
	}
}
