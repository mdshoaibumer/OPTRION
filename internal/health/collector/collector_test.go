package collector_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/optrion/optrion/internal/health/collector"
	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
)

func TestBackendCollector_HealthyEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK")) //nolint:errcheck
	}))
	defer srv.Close()

	c := collector.NewBackendCollector("tenant-1", "comp-1", collector.BackendConfig{
		TargetURL: srv.URL,
		Timeout:   5 * time.Second,
	})

	if c.Type() != domain.CollectorBackend {
		t.Errorf("expected type %s, got %s", domain.CollectorBackend, c.Type())
	}
	if c.ComponentID() != "comp-1" {
		t.Errorf("expected component_id comp-1, got %s", c.ComponentID())
	}
	if c.TenantID() != "tenant-1" {
		t.Errorf("expected tenant_id tenant-1, got %s", c.TenantID())
	}

	result, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != nil {
		t.Fatalf("unexpected result error: %v", result.Error)
	}

	metrics := metricMap(result)

	// Check availability
	if v, ok := metrics[domain.MetricAvailability]; !ok || v != 1 {
		t.Errorf("expected availability=1, got %v", v)
	}

	// Check response time is reasonable (< 1 second for local test server)
	if v, ok := metrics[domain.MetricResponseTime]; !ok || v > 1000 {
		t.Errorf("expected response_time < 1000ms, got %v", v)
	}

	// Check no error rate
	if v, ok := metrics[domain.MetricErrorRate]; !ok || v != 0 {
		t.Errorf("expected error_rate=0, got %v", v)
	}

	// Check throughput
	if v, ok := metrics[domain.MetricThroughput]; !ok || v != 1 {
		t.Errorf("expected throughput=1, got %v", v)
	}

	// Check uptime exists
	if _, ok := metrics[domain.MetricUptime]; !ok {
		t.Error("expected uptime metric to exist")
	}
}

func TestBackendCollector_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := collector.NewBackendCollector("tenant-1", "comp-1", collector.BackendConfig{
		TargetURL: srv.URL,
		Timeout:   5 * time.Second,
	})

	result, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	metrics := metricMap(result)

	// 500 → not available
	if v := metrics[domain.MetricAvailability]; v != 0 {
		t.Errorf("expected availability=0 for 500, got %v", v)
	}

	// 500 → error rate = 1
	if v := metrics[domain.MetricErrorRate]; v != 1 {
		t.Errorf("expected error_rate=1 for 500, got %v", v)
	}
}

func TestBackendCollector_ClientError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := collector.NewBackendCollector("tenant-1", "comp-1", collector.BackendConfig{
		TargetURL: srv.URL,
		Timeout:   5 * time.Second,
	})

	result, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	metrics := metricMap(result)

	// 404 → still available (not 5xx)
	if v := metrics[domain.MetricAvailability]; v != 1 {
		t.Errorf("expected availability=1 for 404, got %v", v)
	}

	// 404 → error rate = 1 (>=400)
	if v := metrics[domain.MetricErrorRate]; v != 1 {
		t.Errorf("expected error_rate=1 for 404, got %v", v)
	}
}

func TestBackendCollector_Unreachable(t *testing.T) {
	c := collector.NewBackendCollector("tenant-1", "comp-1", collector.BackendConfig{
		TargetURL: "http://127.0.0.1:1", // Non-routable port
		Timeout:   1 * time.Second,
	})

	result, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error (should be in result): %v", err)
	}

	if result.Error == nil {
		t.Fatal("expected result.Error for unreachable host")
	}

	metrics := metricMap(result)
	if v := metrics[domain.MetricAvailability]; v != 0 {
		t.Errorf("expected availability=0 for unreachable, got %v", v)
	}
}

func TestBackendCollector_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // slow handler
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := collector.NewBackendCollector("tenant-1", "comp-1", collector.BackendConfig{
		TargetURL: srv.URL,
		Timeout:   10 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := c.Collect(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Error == nil {
		t.Fatal("expected result.Error for cancelled context")
	}

	metrics := metricMap(result)
	if v := metrics[domain.MetricAvailability]; v != 0 {
		t.Errorf("expected availability=0 for timeout, got %v", v)
	}
}

// --- Server Collector Tests ---

func TestServerCollector_Collect(t *testing.T) {
	c := collector.NewServerCollector("tenant-1", "comp-server")

	if c.Type() != domain.CollectorServer {
		t.Errorf("expected type %s, got %s", domain.CollectorServer, c.Type())
	}

	result, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ComponentID != "comp-server" {
		t.Errorf("expected component_id comp-server, got %s", result.ComponentID)
	}

	metrics := metricMap(result)

	// All server metrics should be present
	expectedMetrics := []domain.MetricType{
		domain.MetricCPU, domain.MetricRAM, domain.MetricDisk,
		domain.MetricLoadAverage, domain.MetricNetwork,
	}

	for _, mt := range expectedMetrics {
		if _, ok := metrics[mt]; !ok {
			t.Errorf("expected metric %s to be present", mt)
		}
	}

	// RAM should be between 0 and 100
	if ram := metrics[domain.MetricRAM]; ram < 0 || ram > 100 {
		t.Errorf("expected RAM 0-100, got %v", ram)
	}
}

// --- Helper ---

func metricMap(result *port.CollectorResult) map[domain.MetricType]float64 {
	m := make(map[domain.MetricType]float64)
	for _, reading := range result.Metrics {
		m[reading.MetricType] = reading.Value
	}
	return m
}
