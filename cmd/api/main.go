package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/hisamcode/try-greenlight/internal/data"
	"github.com/hisamcode/try-greenlight/internal/jsonlog"
	"github.com/hisamcode/try-greenlight/internal/mailer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	host string
	port string
	env  string
	db   struct {
		dsn          string
		maxOpenConns string // int
		maxIdleConns string // int
		maxIdleTime  string
	}
	limiter struct {
		rps     string // float64
		burst   string // int
		enabled string // bool
	}
	smtp struct {
		host     string
		port     string //int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins string // []string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	flag.StringVar(&cfg.host, "host", os.Getenv("HOST"), "host IP server")
	flag.StringVar(&cfg.port, "port", os.Getenv("PORT"), "API server port")
	flag.StringVar(&cfg.env, "env", os.Getenv("ENVIRONMENT"), "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.StringVar(&cfg.db.maxOpenConns, "db-max-open-conns", os.Getenv("MAX_OPEN_CONNS"), "PostgreSQL max open connections")
	flag.StringVar(&cfg.db.maxIdleConns, "db-max-idle-conns", os.Getenv("MAX_IDLE_CONNS"), "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", os.Getenv("MAX_IDLE_TIME"), "PostgreSQL max connection idle time")

	flag.StringVar(&cfg.limiter.rps, "limiter-rps", os.Getenv("LIMITER_RPS"), "Rate limiter maximum requests per second")
	flag.StringVar(&cfg.limiter.burst, "limiter-burst", os.Getenv("LIMITER_BURST"), "Rate limiter maximum burst")
	flag.StringVar(&cfg.limiter.enabled, "limiter-enabled", os.Getenv("LIMITER_ENABLED"), "Enable rate limiter (true|false)")

	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP HOST")
	flag.StringVar(&cfg.smtp.port, "smtp-port", os.Getenv("SMTP_PORT"), "SMTP PORT")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP USERNAME")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP PASSWORD")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", os.Getenv("SMTP_SENDER"), "SMTP SENDER")

	flag.StringVar(&cfg.cors.trustedOrigins, "cors-trusted-origins", os.Getenv("TRUSTED_CORS_ORIGINS"), "Trusted CORS origins(space seperated)")

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version: \t%s\n", version)
		os.Exit(0)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	mailerPort, err := strconv.Atoi(cfg.smtp.port)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	expvar.NewString("version").Set(version)
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, mailerPort, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
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

	// Set the maximum number of open (in-use + idle) connections in the pool. Note that
	// passing a value less than or equal to 0 will mean there is no limit.
	maxOpenConns, err := strconv.Atoi(cfg.db.maxOpenConns)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxOpenConns)

	// Set the maximum number of idle connections in the pool. Again, passing a value
	// less than or equal to 0 will mean there is no limit.
	maxIdleConns, err := strconv.Atoi(cfg.db.maxIdleConns)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
