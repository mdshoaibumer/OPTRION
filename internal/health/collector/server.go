package collector

import (
	"context"
	"fmt"
	"runtime"

	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
)

// ServerCollector monitors the server's resource usage.
// Uses Go runtime metrics for cross-platform compatibility.
type ServerCollector struct {
	tenantID    string
	componentID string
}

// NewServerCollector creates a new server resource collector.
func NewServerCollector(tenantID, componentID string) *ServerCollector {
	return &ServerCollector{
		tenantID:    tenantID,
		componentID: componentID,
	}
}

func (c *ServerCollector) Type() domain.CollectorType { return domain.CollectorServer }
func (c *ServerCollector) ComponentID() string        { return c.componentID }
func (c *ServerCollector) TenantID() string           { return c.tenantID }

func (c *ServerCollector) Collect(_ context.Context) (*port.CollectorResult, error) {
	result := &port.CollectorResult{
		ComponentID:   c.componentID,
		CollectorType: domain.CollectorServer,
		Metrics:       make([]port.MetricReading, 0, 5),
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 1. CPU: goroutines as proxy for CPU pressure (cross-platform)
	numGoroutines := runtime.NumGoroutine()
	numCPU := runtime.NumCPU()
	// Normalize: goroutines per CPU core (>10 per core = concern)
	cpuPressure := (float64(numGoroutines) / float64(numCPU)) * 10
	if cpuPressure > 100 {
		cpuPressure = 100
	}
	result.Metrics = append(result.Metrics,
		port.MetricReading{
			MetricType: domain.MetricCPU,
			Value:      cpuPressure,
			Labels:     map[string]string{"goroutines": itoa(numGoroutines), "num_cpu": itoa(numCPU), "unit": "percent"},
		},
	)

	// 2. RAM: heap usage as percentage of system memory allocated
	heapUsedMB := float64(memStats.HeapInuse) / (1024 * 1024)
	sysMB := float64(memStats.Sys) / (1024 * 1024)
	ramPct := 0.0
	if sysMB > 0 {
		ramPct = (heapUsedMB / sysMB) * 100
	}
	result.Metrics = append(result.Metrics,
		port.MetricReading{
			MetricType: domain.MetricRAM,
			Value:      ramPct,
			Labels:     map[string]string{"heap_mb": ftoa(heapUsedMB), "sys_mb": ftoa(sysMB), "unit": "percent"},
		},
	)

	// 3. Disk: GC pressure as proxy (number of GC cycles)
	gcPct := float64(memStats.NumGC)
	result.Metrics = append(result.Metrics,
		port.MetricReading{
			MetricType: domain.MetricDisk,
			Value:      0, // Disk is not directly measurable cross-platform without cgo
			Labels:     map[string]string{"gc_cycles": ftoa(gcPct), "note": "disk monitoring requires OS-specific tools"},
		},
	)

	// 4. Load average: use total allocations per second as proxy
	totalAllocs := float64(memStats.TotalAlloc) / (1024 * 1024)
	result.Metrics = append(result.Metrics,
		port.MetricReading{
			MetricType: domain.MetricLoadAverage,
			Value:      float64(numGoroutines),
			Labels:     map[string]string{"total_alloc_mb": ftoa(totalAllocs)},
		},
	)

	// 5. Network: use the number of goroutines as network connection proxy
	result.Metrics = append(result.Metrics,
		port.MetricReading{
			MetricType: domain.MetricNetwork,
			Value:      float64(numGoroutines),
			Labels:     map[string]string{"note": "goroutine count as network activity proxy"},
		},
	)

	return result, nil
}

func itoa(i int) string {
	return ftoa(float64(i))
}

func ftoa(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
