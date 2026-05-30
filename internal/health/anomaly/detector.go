package anomaly

import (
	"fmt"
	"math"
	"sync"

	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
)

// Detector identifies anomalous metric readings by comparing against a rolling baseline.
type Detector struct {
	mu sync.RWMutex
	// baselines tracks rolling average and stddev per (component, metric) pair
	baselines map[string]*baseline
	// thresholdMultiplier defines how many standard deviations from mean triggers an anomaly
	thresholdMultiplier float64
	// minSamples required before anomaly detection activates
	minSamples int
}

// baseline holds statistical information for a metric.
type baseline struct {
	count  int
	sum    float64
	sumSq  float64
	values []float64 // circular buffer of recent values
	maxLen int
}

func newBaseline(maxLen int) *baseline {
	return &baseline{
		values: make([]float64, 0, maxLen),
		maxLen: maxLen,
	}
}

func (b *baseline) add(value float64) {
	if len(b.values) >= b.maxLen {
		// Remove oldest value from running stats
		oldest := b.values[0]
		b.sum -= oldest
		b.sumSq -= oldest * oldest
		b.count--
		b.values = b.values[1:]
	}
	b.values = append(b.values, value)
	b.sum += value
	b.sumSq += value * value
	b.count++
}

func (b *baseline) mean() float64 {
	if b.count == 0 {
		return 0
	}
	return b.sum / float64(b.count)
}

func (b *baseline) stddev() float64 {
	if b.count < 2 {
		return 0
	}
	mean := b.mean()
	variance := (b.sumSq / float64(b.count)) - (mean * mean)
	if variance < 0 {
		variance = 0
	}
	return math.Sqrt(variance)
}

// NewDetector creates a new anomaly detector.
// thresholdMultiplier: number of stddevs to trigger anomaly (default 3.0)
// windowSize: number of recent samples to consider (default 60)
// minSamples: minimum samples before detection starts (default 10)
func NewDetector(thresholdMultiplier float64, windowSize, minSamples int) *Detector {
	if thresholdMultiplier <= 0 {
		thresholdMultiplier = 3.0
	}
	if windowSize <= 0 {
		windowSize = 60
	}
	if minSamples <= 0 {
		minSamples = 10
	}
	return &Detector{
		baselines:           make(map[string]*baseline),
		thresholdMultiplier: thresholdMultiplier,
		minSamples:          minSamples,
	}
}

// DetectionResult holds the outcome of anomaly analysis.
type DetectionResult struct {
	IsAnomaly     bool
	Severity      domain.Severity
	Title         string
	Description   string
	ExpectedValue float64
	ActualValue   float64
}

// Analyze checks a collector result for anomalies.
// Returns a slice of detected anomalies (may be empty).
func (d *Detector) Analyze(result *port.CollectorResult) []DetectionResult {
	d.mu.Lock()
	defer d.mu.Unlock()

	detections := make([]DetectionResult, 0)

	for _, metric := range result.Metrics {
		key := result.ComponentID + ":" + string(metric.MetricType)

		bl, exists := d.baselines[key]
		if !exists {
			bl = newBaseline(60) // Window size
			d.baselines[key] = bl
		}

		// Only detect after enough samples
		if bl.count >= d.minSamples {
			mean := bl.mean()
			stddev := bl.stddev()

			if stddev > 0 {
				zScore := math.Abs(metric.Value-mean) / stddev
				if zScore > d.thresholdMultiplier {
					severity := classifySeverity(zScore)
					detections = append(detections, DetectionResult{
						IsAnomaly:     true,
						Severity:      severity,
						Title:         fmt.Sprintf("%s spike detected", metric.MetricType),
						Description:   fmt.Sprintf("%s deviated %.1f standard deviations from baseline (expected: %.2f, actual: %.2f)", metric.MetricType, zScore, mean, metric.Value),
						ExpectedValue: mean,
						ActualValue:   metric.Value,
					})
				}
			}
		}

		// Always update baseline
		bl.add(metric.Value)
	}

	return detections
}

// classifySeverity determines severity based on how far from the baseline.
func classifySeverity(zScore float64) domain.Severity {
	switch {
	case zScore > 5:
		return domain.SeverityCritical
	case zScore > 4:
		return domain.SeverityHigh
	case zScore > 3:
		return domain.SeverityMedium
	default:
		return domain.SeverityLow
	}
}
