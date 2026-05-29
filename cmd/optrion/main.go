package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/optrion/optrion/internal/app"
	"github.com/optrion/optrion/migrations"
)

func main() {
	app.Migrations = migrations.FS

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Root context — cancelled on OS signal
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Build application container (Config → Logger → DB → Redis → Server)
	container, err := app.NewContainer(ctx)
	if err != nil {
		return fmt.Errorf("initializing application: %w", err)
	}

	// Start HTTP server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- container.Server.Start()
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		container.Logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	container.Shutdown(shutdownCtx)
	return nil
}
