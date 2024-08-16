package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/paulcockrell/greenlight/internal/data"
	"github.com/paulcockrell/greenlight/internal/jsonlog"
	"github.com/paulcockrell/greenlight/internal/mailer"
)

const version = "1.0.0"

type config struct {
	env  string
	smtp struct {
		host     string
		username string
		password string
		sender   string
		port     int
	}
	db struct {
		dsn          string
		maxIdleTime  string
		maxOpenConns int
		maxIdleConns int
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	port int
}

type application struct {
	models data.Models
	logger *jsonlog.Logger
	mailer mailer.Mailer
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

	// Read rate-limiter configuration settings into the config struct
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "ce58e18287850a", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "40571e9f30ab0a", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.paulcockrell.net>", "SMTP sender")

	flag.Parse()

	// Initialize a new logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Create DB connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Defer call to db.Close() so that the connection pool is close before the main()
	// function exits
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	// Declare an instance of the application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
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
