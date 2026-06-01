package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/optrion/optrion/internal/recommendation/domain/recommendation"
)

// RecommendationPostgresRepository persists recommendations.
type RecommendationPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewRecommendationRepository(pool *pgxpool.Pool) *RecommendationPostgresRepository {
	return &RecommendationPostgresRepository{pool: pool}
}

func (r *RecommendationPostgresRepository) Create(ctx context.Context, rec *recommendation.Recommendation) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO recommendations (id, tenant_id, incident_id, report_id, category, priority, title, description, confidence, risk_level, evidence_ids, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		rec.ID, rec.TenantID, rec.IncidentID, rec.ReportID, rec.Category,
		rec.Priority, rec.Title, rec.Description, rec.Confidence, rec.RiskLevel,
		rec.EvidenceIDs, rec.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert recommendation: %w", err)
	}
	return nil
}

func (r *RecommendationPostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*recommendation.Recommendation, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, tenant_id, incident_id, report_id, category, priority, title, description, confidence, risk_level, evidence_ids, created_at
		 FROM recommendations WHERE id = $1`, id)

	var rec recommendation.Recommendation
	err := row.Scan(&rec.ID, &rec.TenantID, &rec.IncidentID, &rec.ReportID, &rec.Category,
		&rec.Priority, &rec.Title, &rec.Description, &rec.Confidence, &rec.RiskLevel,
		&rec.EvidenceIDs, &rec.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find recommendation: %w", err)
	}
	return &rec, nil
}

func (r *RecommendationPostgresRepository) ListByIncident(ctx context.Context, incidentID uuid.UUID) ([]*recommendation.Recommendation, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, tenant_id, incident_id, report_id, category, priority, title, description, confidence, risk_level, evidence_ids, created_at
		 FROM recommendations WHERE incident_id = $1 ORDER BY created_at DESC`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("list recommendations: %w", err)
	}
	defer rows.Close()

	var results []*recommendation.Recommendation
	for rows.Next() {
		var rec recommendation.Recommendation
		if err := rows.Scan(&rec.ID, &rec.TenantID, &rec.IncidentID, &rec.ReportID, &rec.Category,
			&rec.Priority, &rec.Title, &rec.Description, &rec.Confidence, &rec.RiskLevel,
			&rec.EvidenceIDs, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan recommendation: %w", err)
		}
		results = append(results, &rec)
	}
	return results, nil
}
