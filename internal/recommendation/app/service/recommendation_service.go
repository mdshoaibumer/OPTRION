package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/optrion/optrion/internal/ai/adapter/provider"
	"github.com/optrion/optrion/internal/recommendation/domain/recommendation"
	"github.com/optrion/optrion/internal/recommendation/port/repository"
)

// RecommendationService generates and manages recommendations.
type RecommendationService struct {
	provider provider.AIProvider
	recs     repository.RecommendationRepository
	context  IncidentContextProvider
	logger   *slog.Logger
}

// IncidentContextProvider provides incident data for recommendation generation.
type IncidentContextProvider interface {
	GetIncidentContext(ctx context.Context, incidentID uuid.UUID) ([]byte, uuid.UUID, error)
}

// NewRecommendationService creates a new recommendation service.
func NewRecommendationService(
	p provider.AIProvider,
	recs repository.RecommendationRepository,
	ctxProvider IncidentContextProvider,
	logger *slog.Logger,
) *RecommendationService {
	return &RecommendationService{
		provider: p,
		recs:     recs,
		context:  ctxProvider,
		logger:   logger,
	}
}

// Recommend generates recommendations for an incident using AI.
func (s *RecommendationService) Recommend(ctx context.Context, incidentID uuid.UUID) error {
	snapshot, tenantID, err := s.context.GetIncidentContext(ctx, incidentID)
	if err != nil {
		return fmt.Errorf("getting incident context: %w", err)
	}

	prompt := fmt.Sprintf(`You are an expert SRE. Based on the following incident context, provide actionable recommendations.

INCIDENT CONTEXT:
%s

Respond ONLY with valid JSON in this exact format:
{
  "recommendations": [
    {
      "category": "Database|Redis|Infrastructure|Application|Network|Deployment|Capacity|Security",
      "priority": "Critical|High|Medium|Low",
      "title": "Short actionable title",
      "description": "Detailed explanation with steps",
      "confidence": 0.85,
      "risk_level": "Low|Medium|High"
    }
  ]
}

Rules:
- Provide 1-5 recommendations max
- confidence must be between 0.0 and 1.0
- Be specific and actionable
- Do NOT recommend dangerous actions (DROP DATABASE, rm -rf, etc.)`, string(snapshot))

	output, err := s.provider.Analyze(ctx, []byte(prompt))
	if err != nil {
		return fmt.Errorf("AI provider recommendation: %w", err)
	}

	var result struct {
		Recommendations []struct {
			Category    string  `json:"category"`
			Priority    string  `json:"priority"`
			Title       string  `json:"title"`
			Description string  `json:"description"`
			Confidence  float64 `json:"confidence"`
			RiskLevel   string  `json:"risk_level"`
		} `json:"recommendations"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("parsing AI recommendation output: %w", err)
	}

	reportID := uuid.New()
	for _, r := range result.Recommendations {
		conf := r.Confidence
		if conf < 0 {
			conf = 0
		}
		if conf > 1 {
			conf = 1
		}

		rec := &recommendation.Recommendation{
			ID:          uuid.New(),
			TenantID:    tenantID,
			IncidentID:  incidentID,
			ReportID:    reportID,
			Category:    r.Category,
			Priority:    r.Priority,
			Title:       r.Title,
			Description: r.Description,
			Confidence:  conf,
			RiskLevel:   r.RiskLevel,
			EvidenceIDs: []uuid.UUID{},
			CreatedAt:   time.Now(),
		}
		if err := s.recs.Create(ctx, rec); err != nil {
			s.logger.ErrorContext(ctx, "failed to store recommendation", "error", err)
			continue
		}
	}

	s.logger.InfoContext(ctx, "recommendations generated",
		"incident_id", incidentID,
		"count", len(result.Recommendations),
	)

	return nil
}

// GetRecommendationsByIncident returns all recommendations for an incident.
func (s *RecommendationService) GetRecommendationsByIncident(ctx context.Context, incidentID uuid.UUID) ([]*recommendation.Recommendation, error) {
	return s.recs.ListByIncident(ctx, incidentID)
}
