package domain

import "fmt"

// Config represents the optrion.yaml configuration structure.
type Config struct {
	Tenant      TenantConfig      `yaml:"tenant"`
	Product     ProductConfig     `yaml:"product"`
	Environment EnvironmentConfig `yaml:"environment"`
	Components  []ComponentConfig `yaml:"components"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
}

// TenantConfig represents tenant settings in YAML.
type TenantConfig struct {
	Name string `yaml:"name"`
	Slug string `yaml:"slug"`
	Plan string `yaml:"plan"` // free, starter, professional, enterprise
}

// ProductConfig represents product settings in YAML.
type ProductConfig struct {
	Name        string `yaml:"name"`
	Slug        string `yaml:"slug"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

// EnvironmentConfig represents environment settings in YAML.
type EnvironmentConfig struct {
	Name string `yaml:"name"`
	Tier string `yaml:"tier"` // development, staging, production
}

// ComponentConfig represents component settings in YAML.
type ComponentConfig struct {
	Name        string                 `yaml:"name"`
	Kind        string                 `yaml:"kind"` // database, cache, api, web, queue, storage, service, external
	Description string                 `yaml:"description"`
	Endpoint    string                 `yaml:"endpoint"`
	Port        int                    `yaml:"port"`
	Settings    map[string]interface{} `yaml:"settings"`
}

// MonitoringConfig represents monitoring settings in YAML.
type MonitoringConfig struct {
	Enabled           bool                   `yaml:"enabled"`
	Interval          int                    `yaml:"interval"` // seconds
	HealthCheckPath   string                 `yaml:"health_check_path"`
	MetricsCollectors []string               `yaml:"metrics_collectors"`
	Settings          map[string]interface{} `yaml:"settings"`
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Tenant.Name == "" {
		return fmt.Errorf("tenant.name is required")
	}
	if c.Tenant.Slug == "" {
		return fmt.Errorf("tenant.slug is required")
	}
	if c.Tenant.Plan == "" {
		return fmt.Errorf("tenant.plan is required")
	}

	if c.Product.Name == "" {
		return fmt.Errorf("product.name is required")
	}
	if c.Product.Slug == "" {
		return fmt.Errorf("product.slug is required")
	}

	if c.Environment.Name == "" {
		return fmt.Errorf("environment.name is required")
	}
	if c.Environment.Tier == "" {
		return fmt.Errorf("environment.tier is required")
	}

	if len(c.Components) == 0 {
		return fmt.Errorf("at least one component is required")
	}

	for i, comp := range c.Components {
		if comp.Name == "" {
			return fmt.Errorf("components[%d].name is required", i)
		}
		if comp.Kind == "" {
			return fmt.Errorf("components[%d].kind is required", i)
		}
	}

	return nil
}

// DefaultMonitoringConfig returns default monitoring settings.
func DefaultMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		Enabled:         true,
		Interval:        30,
		HealthCheckPath: "/health",
		MetricsCollectors: []string{
			"http",
			"postgres",
			"redis",
		},
		Settings: make(map[string]interface{}),
	}
}
