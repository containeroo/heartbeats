package internal

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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

	templs := []string{"web/templates/base.html", "web/templates/navbar.html", "web/templates/heartbeats.html", "web/templates/footer.html"}

	tmpl, err := template.ParseFiles(templs...)
	if err != nil {
		log.Errorf("Error parsing template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot parse template. %s", err.Error())))
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", &HeartbeatsServer); err != nil {
		log.Errorf("Error executing template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot execute template. %s", err.Error())))
		return
	}
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

// HandlerPing is the handler for the /ping/<heartbeat> endpoint
func HandlerPing(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	heartbeat, err := HeartbeatsServer.GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusNotFound, outputFormat, err.Error(), "{{ .Status }}")
		return
	}

	heartbeat.GotPing()

	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}

// HandlerPingFail is the handler for the /ping/<heartbeat>/fail endpoint
func HandlerPingFail(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	heartbeat, err := HeartbeatsServer.GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusNotFound, outputFormat, err.Error(), "{{ .Status }}")
		return
	}

	heartbeat.GotPingFail()

	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
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

	heartbeat, err := HeartbeatsServer.GetHeartbeatByName(heartbeatName)
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

// OutputHistory outputs history in wanted format
func OutputHistory(w http.ResponseWriter, req *http.Request, outputFormat string, heartbeatName string) {
	if heartbeatName == "" {
		WriteOutput(w, http.StatusOK, outputFormat, &HistoryCache.History, "{{ . }}")
		return
	}

	heartbeat, err := HeartbeatsServer.GetHeartbeatByName(heartbeatName)
	if err != nil {
		WriteOutput(w, http.StatusNotFound, outputFormat, ResponseStatus{Status: "nok", Error: err.Error()}, "Status: {{ .Status }} Error: {{  .Error }}")
		return
	}
	histories, ok := HistoryCache.History[heartbeat.Name]
	if !ok {
		WriteOutput(w, http.StatusNotFound, outputFormat, ResponseStatus{Status: "nok", Error: "No history found"}, "Status: {{ .Status }} Error: {{  .Error }}")
		return
	}

	WriteOutput(w, http.StatusOK, outputFormat, &histories, "{{ . }}")
}

// HandlerHistory is the handler for the /history endpoint
func HandlerHistory(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	vars := mux.Vars(req)
	heartbeatName := vars["heartbeat"]

	if outputFormat := req.URL.Query().Get("output"); outputFormat != "" {
		OutputHistory(w, req, outputFormat, heartbeatName)
		return
	}

	var templs []string

	if heartbeatName == "" {
		templs = []string{"web/templates/base.html", "web/templates/navbar.html", "web/templates/history_all.html", "web/templates/footer.html"}
	} else {
		templs = []string{"web/templates/base.html", "web/templates/navbar.html", "web/templates/history.html", "web/templates/footer.html"}
	}

	tmpl, err := template.ParseFiles(templs...)
	if err != nil {
		log.Errorf("Error parsing template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot parse template. %s", err.Error())))
		return
	}

	if heartbeatName == "" {
		err = tmpl.ExecuteTemplate(w, "base", HistoryCache)
	} else {
		history, err := HistoryCache.Get(heartbeatName)
		if err != nil {
			log.Warnf("Error getting history: %s", err.Error())
		}
		h := struct {
			Name    string
			History *[]History
		}{
			Name:    heartbeatName,
			History: &history,
		}
		err = tmpl.ExecuteTemplate(w, "base", h)
	}
	if err != nil {
		log.Errorf("Error executing template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot execute template. %s", err.Error())))
		return
	}
}

// HeartbeatsServer is the handler for the /healthz endpoint
func HandlerHealthz(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	WriteOutput(w, http.StatusOK, outputFormat, &ResponseStatus{Status: "ok", Error: ""}, "{{ .Status }}")
}

func HandlerConfig(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	if outputFormat := req.URL.Query().Get("output"); outputFormat != "" {
		WriteOutput(w, http.StatusOK, outputFormat, &HeartbeatsServer, "{{ . }}")
		return
	}

	templs := []string{"web/templates/base.html", "web/templates/navbar.html", "web/templates/config.html", "web/templates/footer.html"}

	tmpl, err := template.ParseFiles(templs...)
	if err != nil {
		log.Errorf("Error parsing template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot parse template. %s", err.Error())))
		return
	}

	o, err := FormatOutput("txt", "{{ . }}", &HeartbeatsServer)

	if err != nil {
		log.Errorf("Error formatting output: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot format output. %s", err.Error())))
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base", o); err != nil {
		log.Errorf("Error executing template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("cannot execute template. %s", err.Error())))
		return
	}
}

// HandlerDashboard is the handler for the /dashboard endpoint
func HandlerDashboard(w http.ResponseWriter, req *http.Request) {
	log.Tracef("%s %s%s", req.Method, req.RequestURI, strings.TrimSpace(req.URL.RawQuery))

	outputFormat := GetOutputFormat(req)

	textTmpl := `<html>
	<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="refresh" content="5" />
<style>
* {
  box-sizing: border-box;
}

body {
  font-family: Arial, Helvetica, sans-serif;
}

/* Float four columns side by side */
.column {
  float: left;
  width: 25%;
  padding: 0 10px;
}

/* Remove extra left and right margins, due to padding */
.row {margin: 0 -5px;}

/* Clear floats after the columns */
.row:after {
  content: "";
  display: table;
  clear: both;
}

/* Responsive columns */
@media screen and (max-width: 600px) {
  .column {
    width: 100%;
    display: block;
    margin-bottom: 20px;
  }
}

/* Style the counter cards */
.card {
  box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.2);
  padding: 16px;
  text-align: center;
}

.status-ok {
	background-color: #4CAF50;
}
.status-nok {
	background-color: #f44336;
}
.status-grace {
	background-color: #ff9800;
}
.status-never {
	background-color: #f1f1f1;
}

</style>
</head>
<body>

<h2>Heartbeats Dashboard</h2>
<p>Resize the browser window to see the effect.</p>

<div class="row">

	{{ range . }}
  <div class="column">
    <div class="card {{ if eq .Status "OK" }}status-ok{{else if eq .Status "GRACE"}}status-grace{{else if eq .Status "NOK" }}status-nok{{ else }}status-never{{end}}"">
      <h3>{{ .Name }}</h3>
      <p>Last Ping: {{ if .LastPing.IsZero }}never{{ else }}{{ .LastPing.Format "2006-01-02 15:04:05" }}{{ end }}</p>
    </div>
  </div>
	{{ end }}

</div>

</body>
</html>
`

	WriteOutput(w, http.StatusOK, outputFormat, &HeartbeatsServer.Heartbeats, textTmpl)
}
