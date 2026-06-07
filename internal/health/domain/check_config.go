package domain

import (
	"fmt"
	"time"

	"github.com/optrion/optrion/internal/shared/id"
)

// HealthCheckConfig defines per-component health check configuration.
type HealthCheckConfig struct {
	ID            string
	TenantID      string
	ComponentID   string
	CheckInterval time.Duration
	Timeout       time.Duration
	Retries       int
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewHealthCheckConfig creates a new health check configuration with defaults.
func NewHealthCheckConfig(tenantID, componentID string) *HealthCheckConfig {
	now := time.Now().UTC()
	return &HealthCheckConfig{
		ID:            id.New(),
		TenantID:      tenantID,
		ComponentID:   componentID,
		CheckInterval: 60 * time.Second,
		Timeout:       10 * time.Second,
		Retries:       3,
		Enabled:       true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Validate checks that the configuration is valid.
func (c *HealthCheckConfig) Validate() error {
	if c.CheckInterval < 5*time.Second {
		return fmt.Errorf("check_interval must be at least 5 seconds")
	}
	if c.CheckInterval > 1*time.Hour {
		return fmt.Errorf("check_interval must be at most 1 hour")
	}
	if c.Timeout < 1*time.Second {
		return fmt.Errorf("timeout must be at least 1 second")
	}
	if c.Timeout > 30*time.Second {
		return fmt.Errorf("timeout must be at most 30 seconds")
	}
	if c.Timeout >= c.CheckInterval {
		return fmt.Errorf("timeout must be less than check_interval")
	}
	if c.Retries < 0 || c.Retries > 10 {
		return fmt.Errorf("retries must be between 0 and 10")
	}
	return nil
}

// Update modifies the configuration.
func (c *HealthCheckConfig) Update(interval, timeout time.Duration, retries int, enabled bool) {
	if interval > 0 {
		c.CheckInterval = interval
	}
	if timeout > 0 {
		c.Timeout = timeout
	}
	c.Retries = retries
	c.Enabled = enabled
	c.UpdatedAt = time.Now().UTC()
}
