package domain

import (
	"testing"
	"time"
)

func TestHealthCheckConfig_Defaults(t *testing.T) {
	cfg := NewHealthCheckConfig("tenant-1", "comp-1")

	if cfg.CheckInterval != 60*time.Second {
		t.Fatalf("expected 60s interval, got %v", cfg.CheckInterval)
	}
	if cfg.Timeout != 10*time.Second {
		t.Fatalf("expected 10s timeout, got %v", cfg.Timeout)
	}
	if cfg.Retries != 3 {
		t.Fatalf("expected 3 retries, got %d", cfg.Retries)
	}
	if !cfg.Enabled {
		t.Fatal("expected enabled by default")
	}
}

func TestHealthCheckConfig_ValidateMinInterval(t *testing.T) {
	cfg := NewHealthCheckConfig("t", "c")
	cfg.CheckInterval = 2 * time.Second
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for interval < 5s")
	}
}

func TestHealthCheckConfig_ValidateMaxInterval(t *testing.T) {
	cfg := NewHealthCheckConfig("t", "c")
	cfg.CheckInterval = 2 * time.Hour
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for interval > 1h")
	}
}

func TestHealthCheckConfig_ValidateTimeoutVsInterval(t *testing.T) {
	cfg := NewHealthCheckConfig("t", "c")
	cfg.CheckInterval = 10 * time.Second
	cfg.Timeout = 15 * time.Second
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when timeout >= interval")
	}
}

func TestHealthCheckConfig_ValidateRetries(t *testing.T) {
	cfg := NewHealthCheckConfig("t", "c")
	cfg.Retries = 15
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for retries > 10")
	}
}

func TestHealthCheckConfig_ValidSuccess(t *testing.T) {
	cfg := NewHealthCheckConfig("t", "c")
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestHealthCheckConfig_Update(t *testing.T) {
	cfg := NewHealthCheckConfig("t", "c")
	cfg.Update(30*time.Second, 5*time.Second, 2, false)

	if cfg.CheckInterval != 30*time.Second {
		t.Fatalf("expected 30s interval, got %v", cfg.CheckInterval)
	}
	if cfg.Timeout != 5*time.Second {
		t.Fatalf("expected 5s timeout, got %v", cfg.Timeout)
	}
	if cfg.Retries != 2 {
		t.Fatalf("expected 2 retries, got %d", cfg.Retries)
	}
	if cfg.Enabled {
		t.Fatal("expected disabled")
	}
}
