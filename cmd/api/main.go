package main

import (
    "context"
    "database/sql"
    "flag"
    "log/slog"
    "os"
    "time"

    "github.com/amari03/test1/internal/data"
    _ "github.com/lib/pq"
)

const appVersion = "1.0.0"

type config struct {
    port int
    env  string
    db   struct {
        dsn string
    }
}

type application struct {
    config config
    logger *slog.Logger
    models data.Models
}

func main() {
    var cfg config

    flag.IntVar(&cfg.port, "port", 4000, "API server port")
    flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
    flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("CRABOO_DB_DSN"), "PostgreSQL DSN")
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