package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/optrion/optrion/internal/ai/domain/aianalysis"
	"github.com/optrion/optrion/internal/ai/domain/aicontext"
	"github.com/optrion/optrion/internal/ai/domain/rootcausereport"
)

// --- Mocks ---

type mockAIProvider struct {
	name      string
	output    []byte
	err       error
	callCount int
}

func (m *mockAIProvider) Analyze(_ context.Context, _ []byte) ([]byte, error) {
	m.callCount++
	return m.output, m.err
}

func (m *mockAIProvider) Name() string {
	if m.name == "" {
		return "MockProvider"
	}
	return m.name
}

type mockAnalysisRepo struct {
	created []*aianalysis.AIAnalysis
	listed  []*aianalysis.AIAnalysis
	err     error
}

func (m *mockAnalysisRepo) Create(_ context.Context, a *aianalysis.AIAnalysis) error {
	m.created = append(m.created, a)
	return m.err
}

func (m *mockAnalysisRepo) FindByID(_ context.Context, id uuid.UUID) (*aianalysis.AIAnalysis, error) {
	for _, a := range m.created {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockAnalysisRepo) ListByIncident(_ context.Context, _ uuid.UUID) ([]*aianalysis.AIAnalysis, error) {
	return m.listed, m.err
}

type mockContextRepo struct {
	created []*aicontext.AIContext
	err     error
}

func (m *mockContextRepo) Create(_ context.Context, c *aicontext.AIContext) error {
	m.created = append(m.created, c)
	return m.err
}

func (m *mockContextRepo) FindByID(_ context.Context, id uuid.UUID) (*aicontext.AIContext, error) {
	for _, c := range m.created {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, errors.New("not found")
}

type mockReportRepo struct {
	created []*rootcausereport.RootCauseReport
	listed  []*rootcausereport.RootCauseReport
	err     error
}

func (m *mockReportRepo) Create(_ context.Context, r *rootcausereport.RootCauseReport) error {
	m.created = append(m.created, r)
	return m.err
}

func (m *mockReportRepo) FindByID(_ context.Context, id uuid.UUID) (*rootcausereport.RootCauseReport, error) {
	for _, r := range m.created {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockReportRepo) ListByIncident(_ context.Context, _ uuid.UUID) ([]*rootcausereport.RootCauseReport, error) {
	return m.listed, m.err
}

type mockContextProvider struct {
	snapshot []byte
	tenantID uuid.UUID
	err      error
}

func (m *mockContextProvider) GetIncidentContext(_ context.Context, _ uuid.UUID) ([]byte, uuid.UUID, error) {
	return m.snapshot, m.tenantID, m.err
}

// --- Tests ---

func newTestService(
	provider *mockAIProvider,
	analyses *mockAnalysisRepo,
	contexts *mockContextRepo,
	reports *mockReportRepo,
	ctxProvider *mockContextProvider,
) *RootCauseService {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewRootCauseService(provider, analyses, contexts, reports, ctxProvider, logger)
}

func TestRootCauseService_Analyze_Success(t *testing.T) {
	incidentID := uuid.New()
	tenantID := uuid.New()

	provider := &mockAIProvider{
		output: []byte(`{"root_cause": "Memory leak in worker", "affected_components": ["worker-1"], "confidence": 0.87, "investigation_hints": ["Check heap"]}`),
	}
	analyses := &mockAnalysisRepo{}
	contexts := &mockContextRepo{}
	reports := &mockReportRepo{}
	ctxProvider := &mockContextProvider{
		snapshot: []byte(`{"incident": "test"}`),
		tenantID: tenantID,
	}

	svc := newTestService(provider, analyses, contexts, reports, ctxProvider)

	err := svc.Analyze(context.Background(), incidentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if provider.callCount != 1 {
		t.Fatalf("expected 1 provider call, got %d", provider.callCount)
	}
	if len(contexts.created) != 1 {
		t.Fatalf("expected 1 context created, got %d", len(contexts.created))
	}
	if len(analyses.created) != 1 {
		t.Fatalf("expected 1 analysis created, got %d", len(analyses.created))
	}
	if len(reports.created) != 1 {
		t.Fatalf("expected 1 report created, got %d", len(reports.created))
	}

	report := reports.created[0]
	if report.LikelyCause != "Memory leak in worker" {
		t.Errorf("expected root cause 'Memory leak in worker', got %q", report.LikelyCause)
	}
	if report.Confidence != 0.87 {
		t.Errorf("expected confidence 0.87, got %f", report.Confidence)
	}
	if len(report.AffectedComponents) != 1 || report.AffectedComponents[0] != "worker-1" {
		t.Errorf("unexpected affected components: %v", report.AffectedComponents)
	}
}

func TestRootCauseService_Analyze_ProviderError(t *testing.T) {
	provider := &mockAIProvider{err: errors.New("rate limited")}
	analyses := &mockAnalysisRepo{}
	contexts := &mockContextRepo{}
	reports := &mockReportRepo{}
	ctxProvider := &mockContextProvider{
		snapshot: []byte(`{"incident": "test"}`),
		tenantID: uuid.New(),
	}

	svc := newTestService(provider, analyses, contexts, reports, ctxProvider)

	err := svc.Analyze(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error from provider failure")
	}
	if len(reports.created) != 0 {
		t.Fatal("expected no report created on provider failure")
	}
}

func TestRootCauseService_Analyze_ContextProviderError(t *testing.T) {
	provider := &mockAIProvider{}
	analyses := &mockAnalysisRepo{}
	contexts := &mockContextRepo{}
	reports := &mockReportRepo{}
	ctxProvider := &mockContextProvider{err: errors.New("incident not found")}

	svc := newTestService(provider, analyses, contexts, reports, ctxProvider)

	err := svc.Analyze(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error from context provider failure")
	}
	if provider.callCount != 0 {
		t.Fatal("provider should not have been called")
	}
}

func TestRootCauseService_Analyze_InvalidJSONFromProvider(t *testing.T) {
	provider := &mockAIProvider{output: []byte("not valid json")}
	analyses := &mockAnalysisRepo{}
	contexts := &mockContextRepo{}
	reports := &mockReportRepo{}
	ctxProvider := &mockContextProvider{
		snapshot: []byte(`{"incident": "test"}`),
		tenantID: uuid.New(),
	}

	svc := newTestService(provider, analyses, contexts, reports, ctxProvider)

	err := svc.Analyze(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for invalid JSON output")
	}
	if len(reports.created) != 0 {
		t.Fatal("expected no report on bad JSON")
	}
}

func TestRootCauseService_Analyze_ConfidenceClamping(t *testing.T) {
	// Confidence > 1 should be clamped to 1
	provider := &mockAIProvider{
		output: []byte(`{"root_cause": "test", "affected_components": [], "confidence": 1.5, "investigation_hints": []}`),
	}
	analyses := &mockAnalysisRepo{}
	contexts := &mockContextRepo{}
	reports := &mockReportRepo{}
	ctxProvider := &mockContextProvider{
		snapshot: []byte(`{}`),
		tenantID: uuid.New(),
	}

	svc := newTestService(provider, analyses, contexts, reports, ctxProvider)

	err := svc.Analyze(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reports.created[0].Confidence != 1.0 {
		t.Errorf("expected confidence clamped to 1.0, got %f", reports.created[0].Confidence)
	}
}

func TestRootCauseService_Analyze_NegativeConfidenceClamping(t *testing.T) {
	provider := &mockAIProvider{
		output: []byte(`{"root_cause": "test", "affected_components": [], "confidence": -0.5, "investigation_hints": []}`),
	}
	analyses := &mockAnalysisRepo{}
	contexts := &mockContextRepo{}
	reports := &mockReportRepo{}
	ctxProvider := &mockContextProvider{
		snapshot: []byte(`{}`),
		tenantID: uuid.New(),
	}

	svc := newTestService(provider, analyses, contexts, reports, ctxProvider)

	err := svc.Analyze(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reports.created[0].Confidence != 0.0 {
		t.Errorf("expected confidence clamped to 0.0, got %f", reports.created[0].Confidence)
	}
}

func TestRootCauseService_GetReportsByIncident(t *testing.T) {
	incidentID := uuid.New()
	expected := []*rootcausereport.RootCauseReport{
		{ID: uuid.New(), IncidentID: incidentID, LikelyCause: "cause1"},
		{ID: uuid.New(), IncidentID: incidentID, LikelyCause: "cause2"},
	}
	reports := &mockReportRepo{listed: expected}
	svc := newTestService(&mockAIProvider{}, &mockAnalysisRepo{}, &mockContextRepo{}, reports, &mockContextProvider{})

	result, err := svc.GetReportsByIncident(context.Background(), incidentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(result))
	}
}

func TestRootCauseService_GetAnalysesByIncident(t *testing.T) {
	incidentID := uuid.New()
	expected := []*aianalysis.AIAnalysis{
		{ID: uuid.New(), IncidentID: incidentID, Status: "completed"},
	}
	analyses := &mockAnalysisRepo{listed: expected}
	svc := newTestService(&mockAIProvider{}, analyses, &mockContextRepo{}, &mockReportRepo{}, &mockContextProvider{})

	result, err := svc.GetAnalysesByIncident(context.Background(), incidentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 analysis, got %d", len(result))
	}
	if result[0].Status != "completed" {
		t.Errorf("expected status 'completed', got %q", result[0].Status)
	}
}
