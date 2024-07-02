package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/paulcockrell/greenlight/internal/data"
)

const version = "1.0.0"

type config struct {
	env string
	db  struct {
		dsn          string
		maxIdleTime  string
		maxOpenConns int
		maxIdleConns int
	}
	port int
}

type application struct {
	logger *log.Logger
	models data.Models
	config config
}

func main() {
	// Declare instance of the config struct
	var cfg config

	// Read the value of the port and env command-line flags into the config struct
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Read database configurations settings into the config struct
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle connections")

	flag.Parse()

	// Initialize a new logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Create DB connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// Defer call to db.Close() so that the connection pool is close before the main()
	// function exits
	defer db.Close()
	logger.Printf("database connection pool established")

	// Declare an instance of the application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	// Declare a HTTP server with some sensible timeout settings
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start the HTTP server
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext to establish a new connection to the database, passing in the
	// context we created above as a parameter. If the connection couldn't be
	// established successfsully withing the 5 second deadline, the n this will return an
	// error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
