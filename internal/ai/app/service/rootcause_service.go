package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/optrion/optrion/internal/ai/adapter/provider"
	"github.com/optrion/optrion/internal/ai/domain/aianalysis"
	"github.com/optrion/optrion/internal/ai/domain/aicontext"
	"github.com/optrion/optrion/internal/ai/domain/rootcausereport"
	"github.com/optrion/optrion/internal/ai/port/repository"
)

// RootCauseService is the concrete implementation of RootCauseAgent.
type RootCauseService struct {
	provider    provider.AIProvider
	analyses    repository.AIAnalysisRepository
	contexts    repository.AIContextRepository
	reports     repository.RootCauseReportRepository
	contextData IncidentContextProvider
	logger      *slog.Logger
}

// IncidentContextProvider provides incident data for AI analysis.
type IncidentContextProvider interface {
	GetIncidentContext(ctx context.Context, incidentID uuid.UUID) ([]byte, uuid.UUID, error)
}

// NewRootCauseService creates a new root cause analysis service.
func NewRootCauseService(
	p provider.AIProvider,
	analyses repository.AIAnalysisRepository,
	contexts repository.AIContextRepository,
	reports repository.RootCauseReportRepository,
	contextData IncidentContextProvider,
	logger *slog.Logger,
) *RootCauseService {
	return &RootCauseService{
		provider:    p,
		analyses:    analyses,
		contexts:    contexts,
		reports:     reports,
		contextData: contextData,
		logger:      logger,
	}
}

// Analyze triggers AI root cause analysis for an incident.
func (s *RootCauseService) Analyze(ctx context.Context, incidentID uuid.UUID) error {
	tenantID := tenantIDFromContext(ctx)

	// Build context snapshot
	snapshot, tenantUUID, err := s.contextData.GetIncidentContext(ctx, incidentID)
	if err != nil {
		return fmt.Errorf("building incident context: %w", err)
	}
	if tenantUUID != uuid.Nil {
		tenantID = tenantUUID
	}

	// Store context snapshot
	ctxID := uuid.New()
	aiCtx := &aicontext.AIContext{
		ID:         ctxID,
		TenantID:   tenantID,
		IncidentID: incidentID,
		Snapshot:   snapshot,
		CreatedAt:  time.Now(),
	}
	if err := s.contexts.Create(ctx, aiCtx); err != nil {
		return fmt.Errorf("storing context snapshot: %w", err)
	}

	// Create analysis record
	reportID := uuid.New()
	analysisID := uuid.New()
	analysis := &aianalysis.AIAnalysis{
		ID:          analysisID,
		TenantID:    tenantID,
		IncidentID:  incidentID,
		ContextID:   ctxID,
		ReportID:    reportID,
		Provider:    s.provider.Name(),
		RequestedAt: time.Now(),
		Status:      "requested",
		CreatedAt:   time.Now(),
	}
	if err := s.analyses.Create(ctx, analysis); err != nil {
		return fmt.Errorf("creating analysis record: %w", err)
	}

	// Call AI provider
	output, err := s.provider.Analyze(ctx, snapshot)
	if err != nil {
		s.logger.ErrorContext(ctx, "AI analysis failed", "incident_id", incidentID, "error", err)
		analysis.Status = "failed"
		analysis.CompletedAt = time.Now()
		_ = s.analyses.Create(ctx, analysis) // update status
		return fmt.Errorf("AI provider analysis: %w", err)
	}

	// Parse response
	var result struct {
		RootCause          string   `json:"root_cause"`
		AffectedComponents []string `json:"affected_components"`
		Confidence         float64  `json:"confidence"`
		InvestigationHints []string `json:"investigation_hints"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return fmt.Errorf("parsing AI output: %w", err)
	}

	// Clamp confidence
	if result.Confidence < 0 {
		result.Confidence = 0
	}
	if result.Confidence > 1 {
		result.Confidence = 1
	}

	// Store report
	report := &rootcausereport.RootCauseReport{
		ID:                 reportID,
		TenantID:           tenantID,
		IncidentID:         incidentID,
		LikelyCause:        result.RootCause,
		AffectedComponents: result.AffectedComponents,
		Confidence:         result.Confidence,
		InvestigationHints: result.InvestigationHints,
		RawOutput:          output,
		CreatedAt:          time.Now(),
	}
	if err := s.reports.Create(ctx, report); err != nil {
		return fmt.Errorf("storing report: %w", err)
	}

	// Mark analysis complete
	analysis.Status = "completed"
	analysis.CompletedAt = time.Now()

	s.logger.InfoContext(ctx, "AI analysis completed",
		"incident_id", incidentID,
		"provider", s.provider.Name(),
		"confidence", result.Confidence,
	)

	return nil
}

// GetAnalysesByIncident returns all analyses for an incident.
func (s *RootCauseService) GetAnalysesByIncident(ctx context.Context, incidentID uuid.UUID) ([]*aianalysis.AIAnalysis, error) {
	return s.analyses.ListByIncident(ctx, incidentID)
}

// GetReport returns a root cause report by ID.
func (s *RootCauseService) GetReport(ctx context.Context, reportID uuid.UUID) (*rootcausereport.RootCauseReport, error) {
	return s.reports.FindByID(ctx, reportID)
}

// GetReportsByIncident returns all reports for an incident.
func (s *RootCauseService) GetReportsByIncident(ctx context.Context, incidentID uuid.UUID) ([]*rootcausereport.RootCauseReport, error) {
	return s.reports.ListByIncident(ctx, incidentID)
}

func tenantIDFromContext(ctx context.Context) uuid.UUID {
	// Extract tenant ID from context if available
	if v := ctx.Value(contextKeyTenantID); v != nil {
		if s, ok := v.(string); ok {
			if id, err := uuid.Parse(s); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

type contextKeyType string

const contextKeyTenantID contextKeyType = "tenant_id"
