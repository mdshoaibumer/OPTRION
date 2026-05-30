package anomaly_test

import (
	"testing"

	"github.com/optrion/optrion/internal/health/anomaly"
	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
)

func TestDetector_NoAnomalyBeforeMinSamples(t *testing.T) {
	d := anomaly.NewDetector(3.0, 60, 10)

	// Feed 9 normal readings — below minSamples threshold
	for i := 0; i < 9; i++ {
		result := &port.CollectorResult{
			ComponentID:   "comp-1",
			CollectorType: domain.CollectorBackend,
			Metrics: []port.MetricReading{
				{MetricType: domain.MetricResponseTime, Value: 100},
			},
		}
		detections := d.Analyze(result)
		if len(detections) > 0 {
			t.Fatalf("iteration %d: expected no anomalies before minSamples, got %d", i, len(detections))
		}
	}
}

func TestDetector_NoAnomalyForNormalValues(t *testing.T) {
	d := anomaly.NewDetector(3.0, 60, 10)

	// Feed 20 stable readings
	for i := 0; i < 20; i++ {
		result := &port.CollectorResult{
			ComponentID:   "comp-1",
			CollectorType: domain.CollectorBackend,
			Metrics: []port.MetricReading{
				{MetricType: domain.MetricResponseTime, Value: 100},
			},
		}
		detections := d.Analyze(result)
		if len(detections) > 0 {
			t.Fatalf("iteration %d: expected no anomalies for stable values, got %d", i, len(detections))
		}
	}
}

func TestDetector_DetectsSpike(t *testing.T) {
	d := anomaly.NewDetector(3.0, 60, 10)

	// Build baseline with consistent values (small variance)
	for i := 0; i < 15; i++ {
		result := &port.CollectorResult{
			ComponentID:   "comp-1",
			CollectorType: domain.CollectorBackend,
			Metrics: []port.MetricReading{
				{MetricType: domain.MetricResponseTime, Value: 100 + float64(i%3)}, // 100-102 range
			},
		}
		d.Analyze(result)
	}

	// Now send a massive spike
	spikeResult := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricResponseTime, Value: 500},
		},
	}
	detections := d.Analyze(spikeResult)
	if len(detections) == 0 {
		t.Fatal("expected anomaly for massive spike")
	}
	if !detections[0].IsAnomaly {
		t.Error("expected IsAnomaly=true")
	}
	if detections[0].ActualValue != 500 {
		t.Errorf("expected actual value 500, got %f", detections[0].ActualValue)
	}
}

func TestDetector_SeverityClassification(t *testing.T) {
	d := anomaly.NewDetector(3.0, 60, 10)

	// Build baseline: very tight variance around 100
	for i := 0; i < 20; i++ {
		result := &port.CollectorResult{
			ComponentID:   "comp-1",
			CollectorType: domain.CollectorBackend,
			Metrics: []port.MetricReading{
				{MetricType: domain.MetricResponseTime, Value: 100},
			},
		}
		d.Analyze(result)
	}

	// With zero stddev, no anomalies can be detected (division by zero guarded)
	// Let's use a detector with slight variance
	d2 := anomaly.NewDetector(3.0, 60, 10)
	for i := 0; i < 20; i++ {
		val := 100.0
		if i%2 == 0 {
			val = 101.0
		}
		result := &port.CollectorResult{
			ComponentID:   "comp-2",
			CollectorType: domain.CollectorBackend,
			Metrics: []port.MetricReading{
				{MetricType: domain.MetricResponseTime, Value: val},
			},
		}
		d2.Analyze(result)
	}

	// The stddev is ~0.5 (alternating 100/101). Mean ~100.5
	// A value of 105 → z = (105-100.5)/0.5 = 9.0 → critical (>5σ)
	spikeResult := &port.CollectorResult{
		ComponentID:   "comp-2",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricResponseTime, Value: 105},
		},
	}
	detections := d2.Analyze(spikeResult)
	if len(detections) == 0 {
		t.Fatal("expected anomaly detection")
	}
	if detections[0].Severity != domain.SeverityCritical {
		t.Errorf("expected critical severity for z=9, got %s", detections[0].Severity)
	}
}

func TestDetector_MultipleMetrics(t *testing.T) {
	d := anomaly.NewDetector(3.0, 60, 10)

	// Build baseline with two metrics
	for i := 0; i < 15; i++ {
		result := &port.CollectorResult{
			ComponentID:   "comp-1",
			CollectorType: domain.CollectorServer,
			Metrics: []port.MetricReading{
				{MetricType: domain.MetricCPU, Value: 50 + float64(i%3)},
				{MetricType: domain.MetricRAM, Value: 60 + float64(i%2)},
			},
		}
		d.Analyze(result)
	}

	// Spike only in CPU
	spikeResult := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorServer,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricCPU, Value: 200},
			{MetricType: domain.MetricRAM, Value: 60},
		},
	}
	detections := d.Analyze(spikeResult)
	if len(detections) == 0 {
		t.Fatal("expected at least one anomaly for CPU spike")
	}

	// Check that it's the CPU anomaly, not RAM
	foundCPU := false
	for _, det := range detections {
		if det.ActualValue == 200 {
			foundCPU = true
		}
	}
	if !foundCPU {
		t.Error("expected CPU anomaly to be detected")
	}
}

func TestDetector_DefaultParameters(t *testing.T) {
	// Test that zero/negative params get defaults
	d := anomaly.NewDetector(0, 0, 0)

	// Should not panic; should use defaults (3.0, 60, 10)
	result := &port.CollectorResult{
		ComponentID:   "comp-1",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricResponseTime, Value: 100},
		},
	}
	detections := d.Analyze(result)
	if len(detections) != 0 {
		t.Error("expected no anomalies on first sample")
	}
}

func TestDetector_IsolatesComponents(t *testing.T) {
	d := anomaly.NewDetector(3.0, 60, 10)

	// Build baseline for comp-1
	for i := 0; i < 15; i++ {
		result := &port.CollectorResult{
			ComponentID:   "comp-1",
			CollectorType: domain.CollectorBackend,
			Metrics: []port.MetricReading{
				{MetricType: domain.MetricResponseTime, Value: 100 + float64(i%2)},
			},
		}
		d.Analyze(result)
	}

	// comp-2 sends value 200 — should NOT be anomaly since comp-2 has no baseline yet
	result := &port.CollectorResult{
		ComponentID:   "comp-2",
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricResponseTime, Value: 200},
		},
	}
	detections := d.Analyze(result)
	if len(detections) > 0 {
		t.Error("expected no anomalies for component with insufficient samples")
	}
}
