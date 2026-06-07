// Package main provides a test server launcher for Playwright E2E tests.
// It starts embedded PostgreSQL + miniredis, runs migrations, and starts the Optrion server.
// Usage: go run ./test/e2e-playwright/testserver
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/optrion/optrion/internal/app"
	"github.com/optrion/optrion/migrations"
	"github.com/optrion/optrion/test/testutil"
)

func main() {
	if !testutil.EmbeddedPostgresSupported() {
		log.Fatal("Embedded PostgreSQL is not supported on this OS/architecture")
	}

	log.Println("[playwright-server] Starting embedded PostgreSQL...")
	pg, err := testutil.StartEmbeddedPostgres()
	if err != nil {
		log.Fatalf("Failed to start embedded PostgreSQL: %v", err)
	}
	pg.SetEnv()
	log.Printf("[playwright-server] PostgreSQL running on port %d", pg.Port)

	log.Println("[playwright-server] Starting miniredis...")
	mr, err := testutil.StartMiniRedis()
	if err != nil {
		_ = pg.Stop() //nolint:errcheck // best-effort cleanup on startup failure
		log.Fatalf("Failed to start miniredis: %v", err)
	}
	mr.SetEnv()
	log.Printf("[playwright-server] Redis running on %s", os.Getenv("REDIS_ADDR"))

	// Ensure HTTP port is set
	if os.Getenv("HTTP_PORT") == "" {
		os.Setenv("HTTP_PORT", "8080")
	}
	os.Setenv("HTTP_HOST", "0.0.0.0")
	os.Setenv("APP_ENV", "development")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("LOG_FORMAT", "text")

	// Set up application
	app.Migrations = migrations.FS

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("[playwright-server] Initializing application container...")
	container, err := app.NewContainer(ctx)
	if err != nil {
		mr.Stop()
		_ = pg.Stop() //nolint:errcheck // best-effort cleanup on startup failure
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Start HTTP server
	errCh := make(chan error, 1)
	go func() {
		errCh <- container.Server.Start()
	}()

	port := os.Getenv("HTTP_PORT")
	log.Printf("[playwright-server] ✓ Optrion API server running on http://localhost:%s", port)
	log.Println("[playwright-server] Ready for Playwright tests. Press Ctrl+C to stop.")
	fmt.Printf("\n  BASE_URL=http://localhost:%s\n\n", port)

	// Start health monitoring if available
	if container.Scheduler != nil {
		container.Scheduler.Start(ctx)
	}

	// Wait for shutdown
	select {
	case <-ctx.Done():
		log.Println("[playwright-server] Shutdown signal received...")
	case err := <-errCh:
		if err != nil {
			log.Printf("[playwright-server] Server error: %v", err)
		}
	}

	// Cleanup
	mr.Stop()
	if err := pg.Stop(); err != nil {
		log.Printf("[playwright-server] Warning: failed to stop embedded postgres: %v", err)
	}
	log.Println("[playwright-server] Stopped.")
}
