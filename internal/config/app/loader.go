package app

import (
	"fmt"
	"os"

	"github.com/optrion/optrion/internal/config/domain"
	"gopkg.in/yaml.v3"
)

// ConfigLoader loads YAML configuration files.
type ConfigLoader struct {
	filePath string
}

// NewConfigLoader creates a new ConfigLoader.
func NewConfigLoader(filePath string) *ConfigLoader {
	return &ConfigLoader{
		filePath: filePath,
	}
}

// Load reads and parses the YAML configuration file.
func (cl *ConfigLoader) Load() (*domain.Config, error) {
	// Read file
	data, err := os.ReadFile(cl.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config domain.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Set defaults
	if !config.Monitoring.Enabled {
		config.Monitoring = domain.DefaultMonitoringConfig()
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// LoadFromString parses YAML from a string.
func (cl *ConfigLoader) LoadFromString(yamlContent string) (*domain.Config, error) {
	var config domain.Config
	if err := yaml.Unmarshal([]byte(yamlContent), &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Set defaults
	if !config.Monitoring.Enabled {
		config.Monitoring = domain.DefaultMonitoringConfig()
	}

	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// GenerateTemplate generates a sample optrion.yaml template.
func GenerateTemplate() string {
	return `# OPTRION Configuration File
# This file configures automatic registration and monitoring for your application

tenant:
  name: "GymFlow"
  slug: "gymflow"
  plan: "starter"  # free, starter, professional, enterprise

product:
  name: "Backend API"
  slug: "backend"
  description: "Main backend service for GymFlow"
  version: "1.0.0"

environment:
  name: "Production"
  tier: "production"  # development, staging, production

# List of components to monitor
components:
  - name: "PostgreSQL Database"
    kind: "database"
    description: "Main application database"
    endpoint: "postgresql://user:pass@postgres:5432/gymflow"
    port: 5432

  - name: "Redis Cache"
    kind: "cache"
    description: "Session and cache storage"
    endpoint: "redis://redis:6379"
    port: 6379

  - name: "Backend Service"
    kind: "api"
    description: "Main REST API service"
    endpoint: "http://localhost:3000"
    port: 3000
    settings:
      health_check_path: "/health"
      timeout_seconds: 5

monitoring:
  enabled: true
  interval: 30  # seconds
  health_check_path: "/health"
  metrics_collectors:
    - "http"
    - "postgres"
    - "redis"
  settings:
    enable_deep_diagnostics: true
    max_metric_history: 1000
`
}
