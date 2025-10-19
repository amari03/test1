package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// Declare the http.Server.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// This channel will receive any errors returned by the graceful shutdown process.
	shutdownError := make(chan error)

	// Start a background goroutine to listen for shutdown signals.
	go func() {
		// A quit channel which carries os.Signal values.
		quit := make(chan os.Signal, 1)

		// Listen for incoming SIGINT and SIGTERM signals and relay them to the quit channel.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This will block until a signal is received.
		s := <-quit

		// Log a message to say that the signal has been caught.
		app.logger.Info("shutting down server", "signal", s.String())

		// Create a context with a 30-second timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Call Shutdown() on the server, passing in the context.
		// This will cause ListenAndServe() to immediately return a http.ErrServerClosed error.
		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	// Calling ListenAndServe() blocks until an error is received.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Block until we receive the return value from Shutdown().
	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}