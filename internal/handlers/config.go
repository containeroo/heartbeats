package handlers

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/containeroo/heartbeats/internal"
	log "github.com/sirupsen/logrus"
)

// Config returns the configuration of the heartbeats server
func Config(w http.ResponseWriter, req *http.Request) {
	LogRequest(req)

	if outputFormat := req.URL.Query().Get("output"); outputFormat != "" {
		WriteOutput(w, http.StatusOK, outputFormat, &internal.HeartbeatsServer, "{{ . }}")
		return
	}

	templs := []string{
		"web/templates/base.html",
		"web/templates/navbar.html",
		"web/templates/config.html",
		"web/templates/footer.html",
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

	config, err := FormatOutput("yaml", "{{ . }}", &internal.HeartbeatsServer)
	if err != nil {
		log.Errorf("Error formatting output: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(fmt.Sprintf("cannot format output. %s", err.Error())))
		if err != nil {
			log.Errorf("Error writing response: %s", err.Error())
		}
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", &config); err != nil {
		log.Errorf("Error executing template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(fmt.Sprintf("cannot execute template. %s", err.Error())))
		if err != nil {
			log.Errorf("Error writing response: %s", err.Error())
		}

		return
	}

}
