package server

import (
	"context"
	"fmt"
	"io/fs"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/containeroo/heartbeats/internal"
	"github.com/containeroo/heartbeats/internal/handlers"
	"github.com/containeroo/heartbeats/internal/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

// NewRouter generates the router used in the HTTP Server
func NewRouter() *mux.Router {
	// Create router and define routes and return that router
	router := mux.NewRouter()

	reg := prometheus.NewRegistry() // Create a non-global registry
	metrics.PromMetrics = *metrics.NewMetrics(reg)

	// handler for embed static files
	fsys := fs.FS(internal.StaticFS)
	contentStatic, _ := fs.Sub(fsys, "web/static")
	fs := http.FileServer(http.FS(contentStatic))
	s := http.StripPrefix("/static/", fs)
	router.PathPrefix("/static/").Handler(s)
	http.Handle("/", router)

	//register handlers
	router.HandleFunc("/", handlers.Home)
	router.HandleFunc("/config", handlers.Config)
	router.HandleFunc("/version", handlers.Version)
	router.HandleFunc("/healthz", handlers.Healthz).Methods("GET", "POST")
	router.HandleFunc("/ping", handlers.PingHelp).Methods("GET", "POST").Methods("GET", "POST")
	router.HandleFunc("/ping/{heartbeat:[a-zA-Z0-9 _-]+}/fail", handlers.PingFail).Methods("GET", "POST")
	router.HandleFunc("/ping/{heartbeat:[a-zA-Z0-9 _-]+}", handlers.Ping).Methods("GET", "POST")
	router.HandleFunc("/history", handlers.History)
	router.HandleFunc("/history/{heartbeat:[a-zA-Z0-9 _-]+}", handlers.History)
	router.HandleFunc("/status", handlers.Status).Methods("GET", "POST")
	router.HandleFunc("/status/{heartbeat:[a-zA-Z0-9 _-]+}", handlers.Status).Methods("GET", "POST")
	router.HandleFunc("/dashboard", handlers.Dashboard)
	router.HandleFunc("/docs", handlers.Docs)
	router.HandleFunc("/docs/{chapter:[a-zA-Z0-9 _-]+}", handlers.Chapter)
	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})).Methods("GET", "POST")
	router.NotFoundHandler = http.HandlerFunc(handlers.NotFound)

	return router
}

// Run will run the HTTP Server
func RunServer(hostname string, port int) {

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
		Addr:         fmt.Sprintf("%s:%d", hostname, port),
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
