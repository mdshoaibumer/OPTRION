package collector

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
)

// BackendConfig holds configuration for backend monitoring.
type BackendConfig struct {
	TargetURL        string
	Timeout          time.Duration
	MaxResponseBytes int64 // Maximum response body size to read (default: 1MB)
	SkipSSRFCheck    bool  // Only set to true for internal/test collectors
}

// defaultMaxResponseBytes is the maximum response body size (1MB).
// Prevents memory exhaustion from malicious or misconfigured targets.
const defaultMaxResponseBytes int64 = 1 * 1024 * 1024

// BackendCollector monitors a backend HTTP service.
type BackendCollector struct {
	tenantID    string
	componentID string
	config      BackendConfig
	client      *http.Client
	lastSuccess time.Time
	startTime   time.Time
}

// NewBackendCollector creates a new backend HTTP collector.
// Returns an error if the target URL fails SSRF validation.
func NewBackendCollector(tenantID, componentID string, cfg BackendConfig) (*BackendCollector, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	// Validate URL to prevent SSRF (skip for explicitly allowed URLs)
	if !cfg.SkipSSRFCheck {
		if err := ValidateTargetURL(cfg.TargetURL); err != nil {
			return nil, fmt.Errorf("SSRF protection: %w", err)
		}
	}

	return &BackendCollector{
		tenantID:    tenantID,
		componentID: componentID,
		config:      cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:           10,
				IdleConnTimeout:        30 * time.Second,
				DisableKeepAlives:      true,
				MaxResponseHeaderBytes: 256 * 1024, // 256KB max headers
			},
			// Prevent open redirects to internal resources (SSRF via redirect)
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return fmt.Errorf("too many redirects")
				}
				// Validate redirect target against SSRF rules
				if !cfg.SkipSSRFCheck {
					if err := ValidateTargetURL(req.URL.String()); err != nil {
						return fmt.Errorf("redirect blocked by SSRF protection: %w", err)
					}
				}
				return nil
			},
		},
		startTime: time.Now(),
	}, nil
}

func (c *BackendCollector) Type() domain.CollectorType { return domain.CollectorBackend }
func (c *BackendCollector) ComponentID() string        { return c.componentID }
func (c *BackendCollector) TenantID() string           { return c.tenantID }

func (c *BackendCollector) Collect(ctx context.Context) (*port.CollectorResult, error) {
	result := &port.CollectorResult{
		ComponentID:   c.componentID,
		CollectorType: domain.CollectorBackend,
		Metrics:       make([]port.MetricReading, 0, 5),
	}

	// Measure response time and availability
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.TargetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		// Backend is down
		result.Metrics = append(result.Metrics,
			port.MetricReading{MetricType: domain.MetricAvailability, Value: 0, Labels: map[string]string{"error": err.Error()}},
			port.MetricReading{MetricType: domain.MetricResponseTime, Value: responseTime.Seconds() * 1000, Labels: map[string]string{"unit": "ms"}},
		)
		result.Error = err
		return result, nil
	}
	defer resp.Body.Close()

	// Limit response body reads to prevent memory exhaustion from large/malicious responses
	maxBytes := c.config.MaxResponseBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxResponseBytes
	}
	limitedReader := io.LimitReader(resp.Body, maxBytes)
	io.Copy(io.Discard, limitedReader) //nolint:errcheck

	// Availability: 1 = up, 0 = down
	available := 1.0
	if resp.StatusCode >= 500 {
		available = 0
	}

	// Error rate: based on status code
	errorRate := 0.0
	if resp.StatusCode >= 400 {
		errorRate = 1.0
	}

	// Update uptime tracking
	if available == 1 {
		c.lastSuccess = time.Now()
	}

	uptimeSeconds := time.Since(c.startTime).Seconds()

	result.Metrics = append(result.Metrics,
		port.MetricReading{
			MetricType: domain.MetricAvailability,
			Value:      available,
			Labels:     map[string]string{"status_code": fmt.Sprintf("%d", resp.StatusCode)},
		},
		port.MetricReading{
			MetricType: domain.MetricResponseTime,
			Value:      responseTime.Seconds() * 1000, // milliseconds
			Labels:     map[string]string{"unit": "ms"},
		},
		port.MetricReading{
			MetricType: domain.MetricErrorRate,
			Value:      errorRate,
			Labels:     map[string]string{"status_code": fmt.Sprintf("%d", resp.StatusCode)},
		},
		port.MetricReading{
			MetricType: domain.MetricThroughput,
			Value:      1, // requests completed per collection cycle
			Labels:     map[string]string{},
		},
		port.MetricReading{
			MetricType: domain.MetricUptime,
			Value:      uptimeSeconds,
			Labels:     map[string]string{"unit": "seconds"},
		},
	)

	return result, nil
}
