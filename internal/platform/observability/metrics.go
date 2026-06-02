package observability

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics provides lightweight application metrics collection.
// Exposes Prometheus-compatible text format on the /metrics endpoint.
// No external dependency required — implements the exposition format directly.
type Metrics struct {
	mu       sync.RWMutex
	counters map[string]*atomic.Int64
	gauges   map[string]*atomic.Int64
	hists    map[string]*Histogram
	logger   *slog.Logger
}

// Histogram tracks value distributions with pre-defined buckets.
type Histogram struct {
	mu      sync.Mutex
	buckets []float64
	counts  []int64
	sum     float64
	count   int64
}

// NewMetrics creates a new metrics collector.
func NewMetrics(logger *slog.Logger) *Metrics {
	return &Metrics{
		counters: make(map[string]*atomic.Int64),
		gauges:   make(map[string]*atomic.Int64),
		hists:    make(map[string]*Histogram),
		logger:   logger,
	}
}

// Counter increments a counter metric.
func (m *Metrics) Counter(name string, delta int64) {
	m.mu.RLock()
	c, exists := m.counters[name]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		c, exists = m.counters[name]
		if !exists {
			c = &atomic.Int64{}
			m.counters[name] = c
		}
		m.mu.Unlock()
	}
	c.Add(delta)
}

// Gauge sets a gauge metric to the given value.
func (m *Metrics) Gauge(name string, value int64) {
	m.mu.RLock()
	g, exists := m.gauges[name]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		g, exists = m.gauges[name]
		if !exists {
			g = &atomic.Int64{}
			m.gauges[name] = g
		}
		m.mu.Unlock()
	}
	g.Store(value)
}

// ObserveLatency records a latency observation to a histogram.
func (m *Metrics) ObserveLatency(name string, duration time.Duration) {
	m.mu.RLock()
	h, exists := m.hists[name]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		h, exists = m.hists[name]
		if !exists {
			// Default latency buckets: 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s
			h = &Histogram{
				buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
				counts:  make([]int64, 12), // buckets + Inf
			}
			m.hists[name] = h
		}
		m.mu.Unlock()
	}

	seconds := duration.Seconds()
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sum += seconds
	h.count++
	for i, boundary := range h.buckets {
		if seconds <= boundary {
			h.counts[i]++
		}
	}
	h.counts[len(h.buckets)]++ // +Inf bucket always incremented
}

// Handler returns an HTTP handler that exposes metrics in Prometheus text format.
func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		m.mu.RLock()
		defer m.mu.RUnlock()

		// Counters
		for name, c := range m.counters {
			fmt.Fprintf(w, "# TYPE %s counter\n", name)
			fmt.Fprintf(w, "%s %d\n", name, c.Load())
		}

		// Gauges
		for name, g := range m.gauges {
			fmt.Fprintf(w, "# TYPE %s gauge\n", name)
			fmt.Fprintf(w, "%s %d\n", name, g.Load())
		}

		// Histograms
		for name, h := range m.hists {
			h.mu.Lock()
			fmt.Fprintf(w, "# TYPE %s histogram\n", name)
			cumulative := int64(0)
			for i, boundary := range h.buckets {
				cumulative += h.counts[i]
				fmt.Fprintf(w, "%s_bucket{le=\"%g\"} %d\n", name, boundary, cumulative)
			}
			cumulative += h.counts[len(h.buckets)]
			fmt.Fprintf(w, "%s_bucket{le=\"+Inf\"} %d\n", name, cumulative)
			fmt.Fprintf(w, "%s_sum %g\n", name, h.sum)
			fmt.Fprintf(w, "%s_count %d\n", name, h.count)
			h.mu.Unlock()
		}
	}
}

// TraceMiddleware adds request tracing with correlation IDs and latency recording.
type TraceMiddleware struct {
	metrics *Metrics
	logger  *slog.Logger
}

// NewTraceMiddleware creates middleware that records request metrics.
func NewTraceMiddleware(metrics *Metrics, logger *slog.Logger) *TraceMiddleware {
	return &TraceMiddleware{metrics: metrics, logger: logger}
}

// Middleware returns an HTTP middleware that captures request-level metrics.
func (t *TraceMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		ww := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		// Record metrics
		t.metrics.Counter("optrion_http_requests_total", 1)
		t.metrics.ObserveLatency("optrion_http_request_duration_seconds", duration)

		if ww.statusCode >= 400 {
			t.metrics.Counter("optrion_http_errors_total", 1)
		}

		// Structured log for request tracing
		t.logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.statusCode,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *statusResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}

// HealthCheckMetrics records health check execution metrics.
type HealthCheckMetrics struct {
	metrics *Metrics
}

// NewHealthCheckMetrics creates health check specific metrics recorder.
func NewHealthCheckMetrics(metrics *Metrics) *HealthCheckMetrics {
	return &HealthCheckMetrics{metrics: metrics}
}

// RecordCheck records a single health check execution.
func (h *HealthCheckMetrics) RecordCheck(tenantID, componentID string, duration time.Duration, success bool) {
	h.metrics.Counter("optrion_health_checks_total", 1)
	h.metrics.ObserveLatency("optrion_health_check_duration_seconds", duration)
	if !success {
		h.metrics.Counter("optrion_health_checks_failed_total", 1)
	}
}

// RecordIncident records incident lifecycle metrics.
func (h *HealthCheckMetrics) RecordIncident(eventType string) {
	h.metrics.Counter("optrion_incidents_total", 1)
}

// RecordEventOutbox records outbox processing metrics.
func (h *HealthCheckMetrics) RecordEventOutbox(status string) {
	h.metrics.Counter("optrion_outbox_events_processed_total", 1)
	if status == "failed" || status == "dead_letter" {
		h.metrics.Counter("optrion_outbox_events_failed_total", 1)
	}
}

// ContextWithTraceID injects a trace ID into the context for distributed tracing.
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext extracts the trace ID from context.
func TraceIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(traceIDKey).(string)
	return v
}

type traceContextKey string

const traceIDKey traceContextKey = "trace_id"
