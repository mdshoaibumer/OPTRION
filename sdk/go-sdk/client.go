package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Config holds SDK configuration.
type Config struct {
	// Server endpoint
	Endpoint string

	// API key for authentication
	APIKey string

	// Application metadata
	TenantID      string
	ProductID     string
	EnvironmentID string

	// Monitoring settings
	MetricsInterval time.Duration
	HealthCheckPath string

	// Component collectors to enable
	Collectors []string

	// Logger
	Logger *slog.Logger

	// Auto-discovery enabled
	AutoDiscovery bool
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Endpoint:        "http://localhost:8080",
		MetricsInterval: 30 * time.Second,
		HealthCheckPath: "/health",
		Collectors: []string{
			"runtime",
			"disk",
			"memory",
			"cpu",
		},
		Logger:        slog.Default(),
		AutoDiscovery: true,
	}
}

// Client is the main SDK client for OPTRION.
type Client struct {
	config           Config
	httpClient       *http.Client
	logger           *slog.Logger
	stopCh           chan struct{}
	customCollectors map[string]MetricCollector
	mu               sync.RWMutex
}

// NewClient creates a new OPTRION SDK client.
func NewClient(config *Config) (*Client, error) {
	if config.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("api key is required")
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	return &Client{
		config:           *config,
		httpClient:       &http.Client{Timeout: 10 * time.Second},
		logger:           config.Logger,
		stopCh:           make(chan struct{}),
		customCollectors: make(map[string]MetricCollector),
	}, nil
}

// registrationRequest is the payload sent to the server during registration.
type registrationRequest struct {
	TenantID      string `json:"tenant_id"`
	ProductID     string `json:"product_id"`
	EnvironmentID string `json:"environment_id"`
	Hostname      string `json:"hostname"`
	GoVersion     string `json:"go_version"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
}

// metricsPayload is the payload sent to the server with collected metrics.
type metricsPayload struct {
	TenantID      string                 `json:"tenant_id"`
	ProductID     string                 `json:"product_id"`
	EnvironmentID string                 `json:"environment_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Metrics       map[string]interface{} `json:"metrics"`
}

// Register registers the application with OPTRION platform.
// This should be called once during application startup.
func (c *Client) Register(ctx context.Context) error {
	c.logger.InfoContext(ctx, "registering with OPTRION",
		"endpoint", c.config.Endpoint,
		"product", c.config.ProductID,
	)

	payload := registrationRequest{
		TenantID:      c.config.TenantID,
		ProductID:     c.config.ProductID,
		EnvironmentID: c.config.EnvironmentID,
		GoVersion:     runtime.Version(),
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
	}

	if err := c.doPost(ctx, "/api/v1/register", payload); err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	c.logger.InfoContext(ctx, "successfully registered with OPTRION")
	return nil
}

// StartMonitoring begins collecting and sending metrics to OPTRION.
// This should be called after Register().
func (c *Client) StartMonitoring(ctx context.Context) error {
	c.logger.InfoContext(ctx, "starting metrics collection",
		"interval", c.config.MetricsInterval,
	)

	// Start metrics collection goroutine
	go c.metricsCollectionLoop(ctx)

	c.logger.InfoContext(ctx, "metrics collection started")
	return nil
}

// StopMonitoring stops collecting and sending metrics.
func (c *Client) StopMonitoring() {
	close(c.stopCh)
	c.logger.Info("metrics collection stopped")
}

// metricsCollectionLoop runs the periodic metrics collection.
func (c *Client) metricsCollectionLoop(ctx context.Context) {
	ticker := time.NewTicker(c.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			if err := c.collectAndSendMetrics(ctx); err != nil {
				c.logger.WarnContext(ctx, "failed to collect metrics", "error", err)
			}
		}
	}
}

// collectAndSendMetrics collects runtime metrics and sends them to OPTRION.
func (c *Client) collectAndSendMetrics(ctx context.Context) error {
	metrics := c.collectMetrics()

	// Collect from custom collectors
	c.mu.RLock()
	for name, collector := range c.customCollectors {
		customMetrics, err := collector.Collect(ctx)
		if err != nil {
			c.logger.WarnContext(ctx, "custom collector failed", "name", name, "error", err)
			continue
		}
		for k, v := range customMetrics {
			metrics[fmt.Sprintf("%s.%s", name, k)] = v
		}
	}
	c.mu.RUnlock()

	payload := metricsPayload{
		TenantID:      c.config.TenantID,
		ProductID:     c.config.ProductID,
		EnvironmentID: c.config.EnvironmentID,
		Timestamp:     time.Now().UTC(),
		Metrics:       metrics,
	}

	if err := c.doPost(ctx, "/api/v1/metrics", payload); err != nil {
		return fmt.Errorf("sending metrics: %w", err)
	}

	c.logger.DebugContext(ctx, "metrics collected and sent",
		"metric_count", len(metrics),
	)

	return nil
}

// collectMetrics gathers runtime metrics from the application.
func (c *Client) collectMetrics() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"timestamp":      time.Now().UTC(),
		"memory_alloc":   m.Alloc,
		"memory_total":   m.TotalAlloc,
		"memory_sys":     m.Sys,
		"goroutines":     runtime.NumGoroutine(),
		"gc_runs":        m.NumGC,
		"gc_pause_total": m.PauseTotalNs,
		"heap_objects":   m.HeapObjects,
		"heap_inuse":     m.HeapInuse,
	}
}

// HealthStatus represents the health status of the application.
type HealthStatus struct {
	Status     string                 `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	Components map[string]interface{} `json:"components"`
	Metrics    map[string]interface{} `json:"metrics"`
}

// GetHealth returns the current health status.
func (c *Client) GetHealth(_ context.Context) (*HealthStatus, error) {
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	return &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Components: map[string]interface{}{
			"memory": m.Alloc,
			"cpu":    runtime.NumGoroutine(),
		},
		Metrics: c.collectMetrics(),
	}, nil
}

// RegisterMetricCollector registers a custom metric collector.
func (c *Client) RegisterMetricCollector(name string, collector MetricCollector) error {
	if name == "" {
		return fmt.Errorf("collector name is required")
	}
	if collector == nil {
		return fmt.Errorf("collector cannot be nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.customCollectors[name] = collector
	c.logger.InfoContext(context.Background(), "metric collector registered", "name", name)
	return nil
}

// MetricCollector defines the interface for custom metric collectors.
type MetricCollector interface {
	// Collect returns the current metrics.
	Collect(ctx context.Context) (map[string]interface{}, error)

	// Name returns the collector name.
	Name() string
}

// doPost sends a POST request with JSON body to the OPTRION server.
func (c *Client) doPost(ctx context.Context, path string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.Endpoint+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("User-Agent", "optrion-go-sdk/1.0.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Discard body to allow connection reuse
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}
