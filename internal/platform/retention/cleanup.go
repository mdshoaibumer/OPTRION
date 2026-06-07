package retention

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RetentionPolicy defines how long data should be kept.
type RetentionPolicy struct {
	MetricSnapshots   time.Duration // How long to keep metric snapshots
	ResolvedIncidents time.Duration // How long to keep resolved/closed incidents
	AuditEvents       time.Duration // How long to keep audit events
	AlertHistory      time.Duration // How long to keep delivered alerts
	AIAnalyses        time.Duration // How long to keep AI analysis results
}

// DefaultPolicy returns a sensible default retention policy.
func DefaultPolicy() RetentionPolicy {
	return RetentionPolicy{
		MetricSnapshots:   90 * 24 * time.Hour,  // 90 days
		ResolvedIncidents: 180 * 24 * time.Hour, // 180 days
		AuditEvents:       365 * 24 * time.Hour, // 1 year
		AlertHistory:      90 * 24 * time.Hour,  // 90 days
		AIAnalyses:        90 * 24 * time.Hour,  // 90 days
	}
}

// CleanupJob runs periodic data retention cleanup.
type CleanupJob struct {
	pool   *pgxpool.Pool
	policy RetentionPolicy
	logger *slog.Logger
	stopCh chan struct{}
}

// NewCleanupJob creates a new data retention cleanup job.
func NewCleanupJob(pool *pgxpool.Pool, policy RetentionPolicy, logger *slog.Logger) *CleanupJob {
	return &CleanupJob{
		pool:   pool,
		policy: policy,
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

// Start begins the periodic cleanup job.
func (j *CleanupJob) Start(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		interval = 6 * time.Hour // Run every 6 hours by default
	}

	j.logger.Info("data retention cleanup job started",
		"interval", interval,
		"metric_retention", j.policy.MetricSnapshots,
		"incident_retention", j.policy.ResolvedIncidents,
	)

	go func() {
		// Run immediately on startup
		j.runCleanup(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-j.stopCh:
				return
			case <-ticker.C:
				j.runCleanup(ctx)
			}
		}
	}()
}

// Stop halts the cleanup job.
func (j *CleanupJob) Stop() {
	close(j.stopCh)
}

// runCleanup executes all retention cleanup queries.
func (j *CleanupJob) runCleanup(ctx context.Context) {
	j.logger.Info("running data retention cleanup")

	totalDeleted := int64(0)

	// 1. Clean old metric snapshots
	if n, err := j.cleanMetricSnapshots(ctx); err != nil {
		j.logger.Error("failed to clean metric snapshots", "error", err)
	} else {
		totalDeleted += n
	}

	// 2. Clean old resolved incidents
	if n, err := j.cleanResolvedIncidents(ctx); err != nil {
		j.logger.Error("failed to clean resolved incidents", "error", err)
	} else {
		totalDeleted += n
	}

	// 3. Clean old audit events
	if n, err := j.cleanAuditEvents(ctx); err != nil {
		j.logger.Error("failed to clean audit events", "error", err)
	} else {
		totalDeleted += n
	}

	// 4. Clean old alert deliveries
	if n, err := j.cleanAlertHistory(ctx); err != nil {
		j.logger.Error("failed to clean alert history", "error", err)
	} else {
		totalDeleted += n
	}

	// 5. Clean old AI analyses
	if n, err := j.cleanAIAnalyses(ctx); err != nil {
		j.logger.Error("failed to clean AI analyses", "error", err)
	} else {
		totalDeleted += n
	}

	// 6. Clean expired API key grace periods
	if n, err := j.cleanExpiredGracePeriodKeys(ctx); err != nil {
		j.logger.Error("failed to clean expired grace period keys", "error", err)
	} else {
		totalDeleted += n
	}

	j.logger.Info("data retention cleanup completed", "total_deleted", totalDeleted)
}

func (j *CleanupJob) cleanMetricSnapshots(ctx context.Context) (int64, error) {
	cutoff := time.Now().UTC().Add(-j.policy.MetricSnapshots)
	result, err := j.pool.Exec(ctx,
		`DELETE FROM metric_snapshots WHERE collected_at < $1`, cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("deleting old metric snapshots: %w", err)
	}
	n := result.RowsAffected()
	if n > 0 {
		j.logger.Info("cleaned metric snapshots", "deleted", n, "cutoff", cutoff)
	}
	return n, nil
}

func (j *CleanupJob) cleanResolvedIncidents(ctx context.Context) (int64, error) {
	cutoff := time.Now().UTC().Add(-j.policy.ResolvedIncidents)
	// Only delete closed incidents (terminal state)
	result, err := j.pool.Exec(ctx,
		`DELETE FROM incidents WHERE status = 'closed' AND closed_at < $1`, cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("deleting old resolved incidents: %w", err)
	}
	n := result.RowsAffected()
	if n > 0 {
		j.logger.Info("cleaned closed incidents", "deleted", n, "cutoff", cutoff)
	}
	return n, nil
}

func (j *CleanupJob) cleanAuditEvents(ctx context.Context) (int64, error) {
	cutoff := time.Now().UTC().Add(-j.policy.AuditEvents)
	result, err := j.pool.Exec(ctx,
		`DELETE FROM audit_events WHERE occurred_at < $1`, cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("deleting old audit events: %w", err)
	}
	n := result.RowsAffected()
	if n > 0 {
		j.logger.Info("cleaned audit events", "deleted", n, "cutoff", cutoff)
	}
	return n, nil
}

func (j *CleanupJob) cleanAlertHistory(ctx context.Context) (int64, error) {
	cutoff := time.Now().UTC().Add(-j.policy.AlertHistory)
	result, err := j.pool.Exec(ctx,
		`DELETE FROM alert_deliveries WHERE delivered_at < $1`, cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("deleting old alert deliveries: %w", err)
	}
	n := result.RowsAffected()
	if n > 0 {
		j.logger.Info("cleaned alert deliveries", "deleted", n, "cutoff", cutoff)
	}
	return n, nil
}

func (j *CleanupJob) cleanAIAnalyses(ctx context.Context) (int64, error) {
	cutoff := time.Now().UTC().Add(-j.policy.AIAnalyses)
	result, err := j.pool.Exec(ctx,
		`DELETE FROM ai_analyses WHERE created_at < $1`, cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("deleting old AI analyses: %w", err)
	}
	n := result.RowsAffected()
	if n > 0 {
		j.logger.Info("cleaned AI analyses", "deleted", n, "cutoff", cutoff)
	}
	return n, nil
}

func (j *CleanupJob) cleanExpiredGracePeriodKeys(ctx context.Context) (int64, error) {
	now := time.Now().UTC()
	result, err := j.pool.Exec(ctx,
		`UPDATE api_keys SET status = 'revoked', updated_at = $1
		 WHERE grace_expires_at IS NOT NULL AND grace_expires_at < $1 AND status = 'active'`,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("revoking expired grace period keys: %w", err)
	}
	n := result.RowsAffected()
	if n > 0 {
		j.logger.Info("revoked expired grace period keys", "revoked", n)
	}
	return n, nil
}
