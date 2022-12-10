package internal

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

var StaticFs embed.FS

// Server is the holds HTTP server settings
type Server struct {
	Hostname string `mapstructure:"hostname"`
	Port     int    `mapstructure:"port"`
	SiteRoot string `mapstructure:"siteRoot"`
}

// NewRouter generates the router used in the HTTP Server
func NewRouter() *mux.Router {
	// Create router and define routes and return that router
	router := mux.NewRouter()

	reg := prometheus.NewRegistry() // Create a non-global registry
	PromMetrics = *NewMetrics(reg)

	// handler for embed static files
	fsys := fs.FS(StaticFs)
	contentStatic, _ := fs.Sub(fsys, "web/static")
	fs := http.FileServer(http.FS(contentStatic))
	s := http.StripPrefix("/static/", fs)
	router.PathPrefix("/static/").Handler(s)
	http.Handle("/", router)

	//register handlers
	router.HandleFunc("/", HandlerHome)
	router.HandleFunc("/config", HandlerConfig)
	router.HandleFunc("/healthz", HandlerHealthz)
	router.HandleFunc("/ping", HandlerPingHelp)
	router.HandleFunc("/ping/{heartbeat:[a-zA-Z0-9 _-]+}/fail", HandlerPingFail).Methods("GET", "POST")
	router.HandleFunc("/ping/{heartbeat:[a-zA-Z0-9 _-]+}", HandlerPing).Methods("GET", "POST")
	router.HandleFunc("/history", HandlerHistory)
	router.HandleFunc("/history/{heartbeat:[a-zA-Z0-9 _-]+}", HandlerHistory)
	router.HandleFunc("/status", HandlerStatus)
	router.HandleFunc("/status/{heartbeat:[a-zA-Z0-9 _-]+}", HandlerStatus)
	router.HandleFunc("/dashboard", HandlerDashboard)
	router.HandleFunc("/docs", HandlerDocs)
	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	return router
}

// Run will run the HTTP Server
func (h *Heartbeats) Run() {
	// Set up a channel to listen to for interrupt signals
	var runChan = make(chan os.Signal, 1)

	// Set up a context to allow for graceful server shutdowns in the event
	// of an OS interrupt (defers the cancel just in case)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	// Define server options
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", h.Server.Hostname, h.Server.Port),
		Handler:      NewRouter(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	// Handle ctrl+c/ctrl+x interrupt
	signal.Notify(runChan, os.Interrupt, syscall.SIGTSTP)

	// Alert the user that the server is starting
	log.Infof("Server is starting on %s", server.Addr)

	// Run the server on a new goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				// Normal interrupt operation, ignore
			} else {
				log.Fatalf("Server failed to start due to err: %v", err)
			}
		}
	}()

	// Block on this channel listeninf for those previously defined syscalls assign
	// to variable so we can let the user know why the server is shutting down
	interrupt := <-runChan

	// If we get one of the pre-prescribed syscalls, gracefully terminate the server
	// while alerting the user
	log.Infof("Server is shutting down due to %+v", interrupt)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server was unable to gracefully shutdown due to err: %+v", err)
	}
}
