package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/optrion/optrion/internal/health/port"
)

// CollectorSchedule pairs a collector with its polling interval.
type CollectorSchedule struct {
	Collector port.HealthCollector
	Interval  time.Duration
}

// ResultHandler is called when a collector produces results.
type ResultHandler func(ctx context.Context, result *port.CollectorResult)

// Scheduler manages periodic collection of health metrics.
type Scheduler struct {
	schedules []CollectorSchedule
	handler   ResultHandler
	logger    *slog.Logger
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.Mutex
	running   bool
}

// NewScheduler creates a new metric collection scheduler.
func NewScheduler(handler ResultHandler, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		schedules: make([]CollectorSchedule, 0),
		handler:   handler,
		logger:    logger,
	}
}

// Register adds a collector with its interval to the scheduler.
func (s *Scheduler) Register(collector port.HealthCollector, interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.schedules = append(s.schedules, CollectorSchedule{
		Collector: collector,
		Interval:  interval,
	})

	s.logger.Info("collector registered",
		"type", collector.Type(),
		"component_id", collector.ComponentID(),
		"interval", interval.String(),
	)
}

// Start begins all scheduled collection loops.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true

	childCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.mu.Unlock()

	s.logger.Info("scheduler starting", "collectors", len(s.schedules))

	for _, schedule := range s.schedules {
		s.wg.Add(1)
		go s.runCollector(childCtx, schedule)
	}
}

// Stop halts all collection loops and waits for them to finish.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.cancel()
	s.mu.Unlock()

	s.wg.Wait()
	s.logger.Info("scheduler stopped")
}

// runCollector runs a single collector on its interval.
func (s *Scheduler) runCollector(ctx context.Context, schedule CollectorSchedule) {
	defer s.wg.Done()

	collector := schedule.Collector
	ticker := time.NewTicker(schedule.Interval)
	defer ticker.Stop()

	// Run immediately on start
	s.collect(ctx, collector)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.collect(ctx, collector)
		}
	}
}

// collect executes a single collection and passes results to the handler.
func (s *Scheduler) collect(ctx context.Context, collector port.HealthCollector) {
	collectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := collector.Collect(collectCtx)
	if err != nil {
		s.logger.Error("collector failed",
			"type", collector.Type(),
			"component_id", collector.ComponentID(),
			"error", err,
		)
		return
	}

	s.handler(ctx, result)
}

// CollectorCount returns the number of registered collectors.
func (s *Scheduler) CollectorCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.schedules)
}
