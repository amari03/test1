package main

import (
    "context"
    "database/sql"
    "flag"
    "log/slog"
    "os"
    "time"
    "sync"

    "github.com/amari03/test1/internal/data"
    "github.com/amari03/test1/internal/mailer"
    _ "github.com/lib/pq"
)

const appVersion = "1.0.0"

type config struct {
    port int
    env  string
    db   struct {
        dsn string
    }
    smtp struct { // Add smtp settings
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
    config config
    logger *slog.Logger
    models data.Models
    mailer mailer.Mailer
    wg      sync.WaitGroup
}

func main() {
    var cfg config

    flag.IntVar(&cfg.port, "port", 4000, "API server port")
    flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
    flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("CRABOO_DB_DSN"), "PostgreSQL DSN")

    // Add flags for SMTP settings.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "a317105c18d1b8", "your-mailtrap-username", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "13c0df1625c6e6", "your-mailtrap-password", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Your App <no-reply@yourapp.com>", "SMTP sender")

    flag.Parse()

    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    db, err := openDB(cfg)
    if err != nil {
        logger.Error(err.Error())
        os.Exit(1)
    }
    defer db.Close()
    logger.Info("database connection pool established")

    app := &application{
        config: cfg,
        logger: logger,
        models: data.NewModels(db),
        mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
    }

    err=app.serve()
    if err != nil{
        logger.Error(err.Error())
        os.Exit(1)
    }
}

func openDB(cfg config) (*sql.DB, error) {
    db, err := sql.Open("postgres", cfg.db.dsn)
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err = db.PingContext(ctx)
    if err != nil {
        return nil, err
    }

    return db, nil
}