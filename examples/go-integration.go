//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/optrion/optrion/sdk"
)

// ExampleGoIntegration demonstrates how to integrate a Go application with OPTRION.
//
// This example shows:
// 1. Loading OPTRION configuration
// 2. Registering with the OPTRION platform
// 3. Starting metrics collection
// 4. Exposing health endpoints
// 5. Graceful shutdown

func main() {
	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Create context with cancellation
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	// Initialize OPTRION SDK
	if err := initializeOPTRION(ctx, logger); err != nil {
		logger.ErrorContext(ctx, "failed to initialize OPTRION", "error", err)
		os.Exit(1)
	}

	// Start HTTP server
	if err := runServer(ctx, logger); err != nil && err != http.ErrServerClosed {
		logger.ErrorContext(ctx, "server error", "error", err)
		os.Exit(1)
	}

	logger.InfoContext(ctx, "application shutting down")
}

// initializeOPTRION sets up OPTRION monitoring for the application.
func initializeOPTRION(ctx context.Context, logger *slog.Logger) error {
	// Create OPTRION SDK client
	client, err := sdk.NewClient(sdk.Config{
		// OPTRION server endpoint
		Endpoint: getEnv("OPTRION_ENDPOINT", "http://localhost:8080"),

		// API key from registration
		APIKey: getEnvRequired("OPTRION_API_KEY"),

		// Application identifiers
		TenantID:      getEnvRequired("OPTRION_TENANT_ID"),
		ProductID:     getEnvRequired("OPTRION_PRODUCT_ID"),
		EnvironmentID: getEnvRequired("OPTRION_ENVIRONMENT_ID"),

		// Metrics collection settings
		MetricsInterval: 30 * time.Second,
		HealthCheckPath: "/health",

		// Collectors to enable
		Collectors: []string{
			"runtime",
			"memory",
			"cpu",
			"disk",
		},

		Logger: logger,
	})
	if err != nil {
		return fmt.Errorf("failed to create OPTRION client: %w", err)
	}

	// Register application with OPTRION
	logger.InfoContext(ctx, "registering with OPTRION")
	if err := client.Register(ctx); err != nil {
		return fmt.Errorf("failed to register with OPTRION: %w", err)
	}

	// Start metrics collection
	logger.InfoContext(ctx, "starting OPTRION metrics collection")
	if err := client.StartMonitoring(ctx); err != nil {
		return fmt.Errorf("failed to start metrics collection: %w", err)
	}

	// Setup graceful shutdown
	go func() {
		<-ctx.Done()
		logger.Info("stopping OPTRION metrics collection")
		client.StopMonitoring()
	}()

	return nil
}

// runServer starts the HTTP server with OPTRION integration.
func runServer(ctx context.Context, logger *slog.Logger) error {
	mux := http.NewServeMux()

	// Health check endpoint - used by OPTRION
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"memory_alloc": %d,
			"memory_total": %d,
			"goroutines": %d,
			"gc_runs": %d
		}`, m.Alloc, m.TotalAlloc, runtime.NumGoroutine(), m.NumGC)
	})

	// API endpoints
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"name":"John Doe"}]`))
	})

	// Server
	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	logger.InfoContext(ctx, "starting HTTP server", "addr", server.Addr)

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)

	case err := <-errCh:
		return err
	}
}

// getEnv gets an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvRequired gets a required environment variable.
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		fmt.Fprintf(os.Stderr, "error: required environment variable %s is not set\n", key)
		os.Exit(1)
	}
	return value
}

// Example environment setup:
//
// # Initialize OPTRION configuration
// $ optrion-cli init
//
// # Register with OPTRION server
// $ optrion-cli register --config optrion.yaml --server http://localhost:8080
//
// # Set required environment variables
// $ export OPTRION_ENDPOINT="http://localhost:8080"
// $ export OPTRION_API_KEY="optrion_key_xxxxx"
// $ export OPTRION_TENANT_ID="tenant-uuid"
// $ export OPTRION_PRODUCT_ID="product-uuid"
// $ export OPTRION_ENVIRONMENT_ID="env-uuid"
//
// # Run the application
// $ go run examples/go-integration.go
//
// # In another terminal, verify integration
// $ optrion-cli verify \
//     --config optrion.yaml \
//     --server http://localhost:8080 \
//     --api-key optrion_key_xxxxx
