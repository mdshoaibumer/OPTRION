package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/optrion/optrion/internal/ai/domain/aianalysis"
	"github.com/optrion/optrion/internal/ai/domain/aicontext"
	"github.com/optrion/optrion/internal/ai/domain/rootcausereport"
)

// AIAnalysisPostgresRepository persists AI analyses.
type AIAnalysisPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewAIAnalysisRepository(pool *pgxpool.Pool) *AIAnalysisPostgresRepository {
	return &AIAnalysisPostgresRepository{pool: pool}
}

func (r *AIAnalysisPostgresRepository) Create(ctx context.Context, a *aianalysis.AIAnalysis) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ai_analyses (id, tenant_id, incident_id, context_id, report_id, provider, requested_at, completed_at, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (id) DO UPDATE SET status = $9, completed_at = $8`,
		a.ID, a.TenantID, a.IncidentID, a.ContextID, a.ReportID,
		a.Provider, a.RequestedAt, nilTime(a.CompletedAt), a.Status, a.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert ai_analysis: %w", err)
	}
	return nil
}

func (r *AIAnalysisPostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*aianalysis.AIAnalysis, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, incident_id, context_id, report_id, provider, requested_at, completed_at, status, created_at
		 FROM ai_analyses WHERE id = $1`, id)

	var a aianalysis.AIAnalysis
	err := row.Scan(&a.ID, &a.TenantID, &a.IncidentID, &a.ContextID, &a.ReportID,
		&a.Provider, &a.RequestedAt, &a.CompletedAt, &a.Status, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find ai_analysis: %w", err)
	}
	return &a, nil
}

func (r *AIAnalysisPostgresRepository) ListByIncident(ctx context.Context, incidentID uuid.UUID) ([]*aianalysis.AIAnalysis, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, incident_id, context_id, report_id, provider, requested_at, completed_at, status, created_at
		 FROM ai_analyses WHERE incident_id = $1 ORDER BY created_at DESC`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("list ai_analyses: %w", err)
	}
	defer rows.Close()

	var results []*aianalysis.AIAnalysis
	for rows.Next() {
		var a aianalysis.AIAnalysis
		if err := rows.Scan(&a.ID, &a.TenantID, &a.IncidentID, &a.ContextID, &a.ReportID,
			&a.Provider, &a.RequestedAt, &a.CompletedAt, &a.Status, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan ai_analysis: %w", err)
		}
		results = append(results, &a)
	}
	return results, nil
}

// AIContextPostgresRepository persists AI context snapshots.
type AIContextPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewAIContextRepository(pool *pgxpool.Pool) *AIContextPostgresRepository {
	return &AIContextPostgresRepository{pool: pool}
}

func (r *AIContextPostgresRepository) Create(ctx context.Context, c *aicontext.AIContext) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ai_context_snapshots (id, tenant_id, incident_id, snapshot, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		c.ID, c.TenantID, c.IncidentID, c.Snapshot, c.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert ai_context: %w", err)
	}
	return nil
}

func (r *AIContextPostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*aicontext.AIContext, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, incident_id, snapshot, created_at
		 FROM ai_context_snapshots WHERE id = $1`, id)

	var c aicontext.AIContext
	err := row.Scan(&c.ID, &c.TenantID, &c.IncidentID, &c.Snapshot, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find ai_context: %w", err)
	}
	return &c, nil
}

// RootCauseReportPostgresRepository persists root cause reports.
type RootCauseReportPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewRootCauseReportRepository(pool *pgxpool.Pool) *RootCauseReportPostgresRepository {
	return &RootCauseReportPostgresRepository{pool: pool}
}

func (r *RootCauseReportPostgresRepository) Create(ctx context.Context, report *rootcausereport.RootCauseReport) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ai_reports (id, tenant_id, incident_id, likely_cause, affected_components, confidence, investigation_hints, raw_output, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		report.ID, report.TenantID, report.IncidentID, report.LikelyCause,
		report.AffectedComponents, report.Confidence, report.InvestigationHints,
		report.RawOutput, report.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert ai_report: %w", err)
	}
	return nil
}

func (r *RootCauseReportPostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*rootcausereport.RootCauseReport, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, incident_id, likely_cause, affected_components, confidence, investigation_hints, raw_output, created_at
		 FROM ai_reports WHERE id = $1`, id)

	var report rootcausereport.RootCauseReport
	err := row.Scan(&report.ID, &report.TenantID, &report.IncidentID, &report.LikelyCause,
		&report.AffectedComponents, &report.Confidence, &report.InvestigationHints,
		&report.RawOutput, &report.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find ai_report: %w", err)
	}
	return &report, nil
}

func (r *RootCauseReportPostgresRepository) ListByIncident(ctx context.Context, incidentID uuid.UUID) ([]*rootcausereport.RootCauseReport, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, incident_id, likely_cause, affected_components, confidence, investigation_hints, raw_output, created_at
		 FROM ai_reports WHERE incident_id = $1 ORDER BY created_at DESC`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("list ai_reports: %w", err)
	}
	defer rows.Close()

	var results []*rootcausereport.RootCauseReport
	for rows.Next() {
		var report rootcausereport.RootCauseReport
		if err := rows.Scan(&report.ID, &report.TenantID, &report.IncidentID, &report.LikelyCause,
			&report.AffectedComponents, &report.Confidence, &report.InvestigationHints,
			&report.RawOutput, &report.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan ai_report: %w", err)
		}
		results = append(results, &report)
	}
	return results, nil
}

// nilTime helper for nullable time fields.
func nilTime(t interface{}) interface{} {
	return t
}
