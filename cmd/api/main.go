package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

type config struct {
	env  string
	port int
}

type application struct {
	logger *log.Logger
	config config
}

func main() {
	// Declare instance of the config struct
	var cfg config

	// Read the value of the port and env command-line flags into the config struct.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Initialize a new logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Declare an instance of the application struct
	app := &application{
		config: cfg,
		logger: logger,
	}

	// Declare a new servemux and add a /v1/healthcheck route
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// Declare a HTTP server with some sensible timeout settings
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}
