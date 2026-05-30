package scheduler_test

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/optrion/optrion/internal/health/domain"
	"github.com/optrion/optrion/internal/health/port"
	"github.com/optrion/optrion/internal/health/scheduler"
)

// mockCollector implements port.HealthCollector for testing.
type mockCollector struct {
	tenantID    string
	componentID string
	calls       atomic.Int32
}

func (m *mockCollector) Type() domain.CollectorType { return domain.CollectorBackend }
func (m *mockCollector) ComponentID() string        { return m.componentID }
func (m *mockCollector) TenantID() string           { return m.tenantID }

func (m *mockCollector) Collect(_ context.Context) (*port.CollectorResult, error) {
	m.calls.Add(1)
	return &port.CollectorResult{
		ComponentID:   m.componentID,
		CollectorType: domain.CollectorBackend,
		Metrics: []port.MetricReading{
			{MetricType: domain.MetricAvailability, Value: 1},
		},
	}, nil
}

func TestScheduler_Register(t *testing.T) {
	handler := func(_ context.Context, _ *port.CollectorResult) {}
	s := scheduler.NewScheduler(handler, slog.Default())

	mc := &mockCollector{tenantID: "t1", componentID: "c1"}
	s.Register(mc, 30*time.Second)

	if s.CollectorCount() != 1 {
		t.Errorf("expected 1 collector, got %d", s.CollectorCount())
	}
}

func TestScheduler_StartStop(t *testing.T) {
	var mu sync.Mutex
	var results []*port.CollectorResult

	handler := func(_ context.Context, result *port.CollectorResult) {
		mu.Lock()
		results = append(results, result)
		mu.Unlock()
	}

	s := scheduler.NewScheduler(handler, slog.Default())
	mc := &mockCollector{tenantID: "t1", componentID: "c1"}
	s.Register(mc, 50*time.Millisecond)

	ctx := context.Background()
	s.Start(ctx)

	// Wait for at least 2 collections (immediate + 1 tick)
	time.Sleep(120 * time.Millisecond)
	s.Stop()

	mu.Lock()
	count := len(results)
	mu.Unlock()

	if count < 2 {
		t.Errorf("expected at least 2 results, got %d", count)
	}
}

func TestScheduler_ImmediateCollection(t *testing.T) {
	var mu sync.Mutex
	var results []*port.CollectorResult

	handler := func(_ context.Context, result *port.CollectorResult) {
		mu.Lock()
		results = append(results, result)
		mu.Unlock()
	}

	s := scheduler.NewScheduler(handler, slog.Default())
	mc := &mockCollector{tenantID: "t1", componentID: "c1"}
	s.Register(mc, 10*time.Second) // Long interval — shouldn't tick in test

	ctx := context.Background()
	s.Start(ctx)

	// Give time for the immediate collection
	time.Sleep(50 * time.Millisecond)
	s.Stop()

	mu.Lock()
	count := len(results)
	mu.Unlock()

	// Should have at least 1 (the immediate collection)
	if count < 1 {
		t.Errorf("expected at least 1 result from immediate collection, got %d", count)
	}
}

func TestScheduler_MultipleCollectors(t *testing.T) {
	var mu sync.Mutex
	componentResults := make(map[string]int)

	handler := func(_ context.Context, result *port.CollectorResult) {
		mu.Lock()
		componentResults[result.ComponentID]++
		mu.Unlock()
	}

	s := scheduler.NewScheduler(handler, slog.Default())
	s.Register(&mockCollector{tenantID: "t1", componentID: "c1"}, 50*time.Millisecond)
	s.Register(&mockCollector{tenantID: "t1", componentID: "c2"}, 50*time.Millisecond)

	ctx := context.Background()
	s.Start(ctx)
	time.Sleep(120 * time.Millisecond)
	s.Stop()

	mu.Lock()
	c1Count := componentResults["c1"]
	c2Count := componentResults["c2"]
	mu.Unlock()

	if c1Count < 2 {
		t.Errorf("expected at least 2 collections for c1, got %d", c1Count)
	}
	if c2Count < 2 {
		t.Errorf("expected at least 2 collections for c2, got %d", c2Count)
	}
}

func TestScheduler_StopWithoutStart(t *testing.T) {
	handler := func(_ context.Context, _ *port.CollectorResult) {}
	s := scheduler.NewScheduler(handler, slog.Default())

	// Should not panic
	s.Stop()
}

func TestScheduler_DoubleStart(t *testing.T) {
	var mu sync.Mutex
	var results []*port.CollectorResult

	handler := func(_ context.Context, result *port.CollectorResult) {
		mu.Lock()
		results = append(results, result)
		mu.Unlock()
	}

	s := scheduler.NewScheduler(handler, slog.Default())
	mc := &mockCollector{tenantID: "t1", componentID: "c1"}
	s.Register(mc, 10*time.Second)

	ctx := context.Background()
	s.Start(ctx)
	s.Start(ctx) // Should be no-op

	time.Sleep(50 * time.Millisecond)
	s.Stop()

	// Only 1 goroutine should have run (immediate collection only)
	mu.Lock()
	count := len(results)
	mu.Unlock()

	if count != 1 {
		t.Errorf("expected exactly 1 result (double start should be no-op), got %d", count)
	}
}

func TestScheduler_ContextCancellation(t *testing.T) {
	var mu sync.Mutex
	var results []*port.CollectorResult

	handler := func(_ context.Context, result *port.CollectorResult) {
		mu.Lock()
		results = append(results, result)
		mu.Unlock()
	}

	s := scheduler.NewScheduler(handler, slog.Default())
	mc := &mockCollector{tenantID: "t1", componentID: "c1"}
	s.Register(mc, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)

	// Let it run briefly
	time.Sleep(80 * time.Millisecond)

	// Cancel context instead of calling Stop
	cancel()
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	count := len(results)
	mu.Unlock()

	if count < 1 {
		t.Errorf("expected at least 1 result before cancellation, got %d", count)
	}
}
