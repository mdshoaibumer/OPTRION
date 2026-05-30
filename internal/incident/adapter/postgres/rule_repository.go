package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/optrion/optrion/internal/incident/domain"
	"github.com/optrion/optrion/internal/shared/id"
)

// RuleRepository implements port.IncidentRuleRepository using PostgreSQL.
type RuleRepository struct {
	pool *pgxpool.Pool
}

// NewRuleRepository creates a new PostgreSQL rule repository.
func NewRuleRepository(pool *pgxpool.Pool) *RuleRepository {
	return &RuleRepository{pool: pool}
}

func (r *RuleRepository) Create(ctx context.Context, rule *domain.IncidentRule) error {
	condition, _ := json.Marshal(rule.Condition)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO incident_rules (id, tenant_id, name, description, component_id, collector_type, condition, severity, cooldown_sec, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, rule.ID, rule.TenantID, rule.Name, rule.Description,
		nullableString(rule.ComponentID), rule.CollectorType,
		condition, rule.Severity, int(rule.Cooldown.Seconds()),
		rule.Enabled, rule.CreatedAt, rule.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting rule: %w", err)
	}
	return nil
}

func (r *RuleRepository) GetByID(ctx context.Context, ruleID string) (*domain.IncidentRule, error) {
	var rule domain.IncidentRule
	var componentID *string
	var conditionJSON []byte
	var cooldownSec int

	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, description, component_id, collector_type, condition, severity, cooldown_sec, enabled, created_at, updated_at
		FROM incident_rules WHERE id = $1
	`, ruleID).Scan(
		&rule.ID, &rule.TenantID, &rule.Name, &rule.Description,
		&componentID, &rule.CollectorType, &conditionJSON, &rule.Severity,
		&cooldownSec, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("getting rule: %w", err)
	}

	if componentID != nil {
		rule.ComponentID = *componentID
	}
	json.Unmarshal(conditionJSON, &rule.Condition) //nolint:errcheck
	rule.Cooldown = time.Duration(cooldownSec) * time.Second
	return &rule, nil
}

func (r *RuleRepository) Update(ctx context.Context, rule *domain.IncidentRule) error {
	condition, _ := json.Marshal(rule.Condition)
	_, err := r.pool.Exec(ctx, `
		UPDATE incident_rules SET
			name = $1, description = $2, component_id = $3, collector_type = $4,
			condition = $5, severity = $6, cooldown_sec = $7, enabled = $8, updated_at = $9
		WHERE id = $10
	`, rule.Name, rule.Description, nullableString(rule.ComponentID),
		rule.CollectorType, condition, rule.Severity,
		int(rule.Cooldown.Seconds()), rule.Enabled, rule.UpdatedAt, rule.ID)
	if err != nil {
		return fmt.Errorf("updating rule: %w", err)
	}
	return nil
}

func (r *RuleRepository) ListEnabled(ctx context.Context, tenantID string) ([]*domain.IncidentRule, error) {
	return r.listRules(ctx, `
		SELECT id, tenant_id, name, description, component_id, collector_type, condition, severity, cooldown_sec, enabled, created_at, updated_at
		FROM incident_rules WHERE tenant_id = $1 AND enabled = TRUE
		ORDER BY created_at ASC
	`, tenantID)
}

func (r *RuleRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.IncidentRule, error) {
	return r.listRules(ctx, `
		SELECT id, tenant_id, name, description, component_id, collector_type, condition, severity, cooldown_sec, enabled, created_at, updated_at
		FROM incident_rules WHERE tenant_id = $1
		ORDER BY created_at ASC
	`, tenantID)
}

func (r *RuleRepository) GetLastFiredAt(ctx context.Context, ruleID, componentID string) (*time.Time, error) {
	// Use the incidents table to determine last firing time
	var lastFired time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(occurred_at) FROM incidents
		WHERE rule_id = $1 AND component_id = $2
	`, ruleID, componentID).Scan(&lastFired)
	if err != nil {
		return nil, nil // No previous firing
	}
	if lastFired.IsZero() {
		return nil, nil
	}
	return &lastFired, nil
}

func (r *RuleRepository) RecordFired(ctx context.Context, ruleID, componentID string, firedAt time.Time) error {
	// Firing is implicitly recorded by creating the incident.
	// This is a no-op since we track via incidents table.
	_ = ruleID
	_ = componentID
	_ = firedAt
	return nil
}

func (r *RuleRepository) listRules(ctx context.Context, query string, args ...interface{}) ([]*domain.IncidentRule, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.IncidentRule
	for rows.Next() {
		var rule domain.IncidentRule
		var componentID *string
		var conditionJSON []byte
		var cooldownSec int

		if err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.Name, &rule.Description,
			&componentID, &rule.CollectorType, &conditionJSON, &rule.Severity,
			&cooldownSec, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning rule: %w", err)
		}

		if componentID != nil {
			rule.ComponentID = *componentID
		}
		json.Unmarshal(conditionJSON, &rule.Condition) //nolint:errcheck
		rule.Cooldown = time.Duration(cooldownSec) * time.Second
		rules = append(rules, &rule)
	}
	return rules, rows.Err()
}

// --- Comment Repository ---

// CommentRepository implements port.IncidentCommentRepository.
type CommentRepository struct {
	pool *pgxpool.Pool
}

// NewCommentRepository creates a new PostgreSQL comment repository.
func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

func (r *CommentRepository) Create(ctx context.Context, comment *domain.IncidentComment) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO incident_comments (id, tenant_id, incident_id, author_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, comment.ID, comment.TenantID, comment.IncidentID, comment.AuthorID, comment.Content, comment.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting comment: %w", err)
	}
	return nil
}

func (r *CommentRepository) ListByIncident(ctx context.Context, incidentID string) ([]*domain.IncidentComment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, incident_id, author_id, content, created_at
		FROM incident_comments WHERE incident_id = $1
		ORDER BY created_at ASC
	`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("listing comments: %w", err)
	}
	defer rows.Close()

	var comments []*domain.IncidentComment
	for rows.Next() {
		var c domain.IncidentComment
		if err := rows.Scan(&c.ID, &c.TenantID, &c.IncidentID, &c.AuthorID, &c.Content, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning comment: %w", err)
		}
		comments = append(comments, &c)
	}
	return comments, rows.Err()
}

// --- Timeline Repository ---

// TimelineRepository implements port.IncidentTimelineRepository.
type TimelineRepository struct {
	pool *pgxpool.Pool
}

// NewTimelineRepository creates a new PostgreSQL timeline repository.
func NewTimelineRepository(pool *pgxpool.Pool) *TimelineRepository {
	return &TimelineRepository{pool: pool}
}

func (r *TimelineRepository) Create(ctx context.Context, entry *domain.IncidentTimeline) error {
	if entry.ID == "" {
		entry.ID = id.New()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO incident_timeline (id, tenant_id, incident_id, entry_type, title, details, actor_id, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, entry.ID, entry.TenantID, entry.IncidentID, entry.EntryType, entry.Title, entry.Details, entry.ActorID, entry.OccurredAt)
	if err != nil {
		return fmt.Errorf("inserting timeline entry: %w", err)
	}
	return nil
}

func (r *TimelineRepository) ListByIncident(ctx context.Context, incidentID string, limit int) ([]*domain.IncidentTimeline, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, incident_id, entry_type, title, details, actor_id, occurred_at
		FROM incident_timeline WHERE incident_id = $1
		ORDER BY occurred_at ASC LIMIT $2
	`, incidentID, limit)
	if err != nil {
		return nil, fmt.Errorf("listing timeline: %w", err)
	}
	defer rows.Close()

	return scanTimelineRows(rows)
}

func (r *TimelineRepository) ListByTenant(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]*domain.IncidentTimeline, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, incident_id, entry_type, title, details, actor_id, occurred_at
		FROM incident_timeline
		WHERE tenant_id = $1 AND occurred_at >= $2 AND occurred_at <= $3
		ORDER BY occurred_at DESC LIMIT $4
	`, tenantID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("listing tenant timeline: %w", err)
	}
	defer rows.Close()

	return scanTimelineRows(rows)
}

func scanTimelineRows(rows interface {
	Next() bool
	Scan(...interface{}) error
	Err() error
}) ([]*domain.IncidentTimeline, error) {
	var entries []*domain.IncidentTimeline
	for rows.Next() {
		var e domain.IncidentTimeline
		if err := rows.Scan(&e.ID, &e.TenantID, &e.IncidentID, &e.EntryType, &e.Title, &e.Details, &e.ActorID, &e.OccurredAt); err != nil {
			return nil, fmt.Errorf("scanning timeline entry: %w", err)
		}
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}
