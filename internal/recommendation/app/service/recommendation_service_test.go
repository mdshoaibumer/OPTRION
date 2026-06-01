package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/optrion/optrion/internal/recommendation/domain/recommendation"
)

// --- Mocks ---

type mockRecAIProvider struct {
	output    []byte
	err       error
	callCount int
}

func (m *mockRecAIProvider) Analyze(_ context.Context, _ []byte) ([]byte, error) {
	m.callCount++
	return m.output, m.err
}

func (m *mockRecAIProvider) Name() string { return "MockProvider" }

type mockRecRepo struct {
	created []*recommendation.Recommendation
	listed  []*recommendation.Recommendation
	err     error
}

func (m *mockRecRepo) Create(_ context.Context, r *recommendation.Recommendation) error {
	m.created = append(m.created, r)
	return m.err
}

func (m *mockRecRepo) FindByID(_ context.Context, id uuid.UUID) (*recommendation.Recommendation, error) {
	for _, r := range m.created {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockRecRepo) ListByIncident(_ context.Context, _ uuid.UUID) ([]*recommendation.Recommendation, error) {
	return m.listed, m.err
}

type mockRecContextProvider struct {
	snapshot []byte
	tenantID uuid.UUID
	err      error
}

func (m *mockRecContextProvider) GetIncidentContext(_ context.Context, _ uuid.UUID) ([]byte, uuid.UUID, error) {
	return m.snapshot, m.tenantID, m.err
}

// --- Tests ---

func newTestRecService(
	provider *mockRecAIProvider,
	recs *mockRecRepo,
	ctxProvider *mockRecContextProvider,
) *RecommendationService {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewRecommendationService(provider, recs, ctxProvider, logger)
}

func TestRecommendationService_Recommend_Success(t *testing.T) {
	incidentID := uuid.New()
	tenantID := uuid.New()

	provider := &mockRecAIProvider{
		output: []byte(`{"recommendations": [{"category": "Database", "priority": "High", "title": "Increase pool size", "description": "Scale connection pool from 25 to 50", "confidence": 0.91, "risk_level": "Low"}, {"category": "Infrastructure", "priority": "Medium", "title": "Add read replica", "description": "Deploy read replica for query offload", "confidence": 0.75, "risk_level": "Medium"}]}`),
	}
	recs := &mockRecRepo{}
	ctxProvider := &mockRecContextProvider{
		snapshot: []byte(`{"incident": "test data"}`),
		tenantID: tenantID,
	}

	svc := newTestRecService(provider, recs, ctxProvider)
	err := svc.Recommend(context.Background(), incidentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if provider.callCount != 1 {
		t.Fatalf("expected 1 provider call, got %d", provider.callCount)
	}
	if len(recs.created) != 2 {
		t.Fatalf("expected 2 recommendations created, got %d", len(recs.created))
	}

	rec1 := recs.created[0]
	if rec1.Category != "Database" {
		t.Errorf("expected category 'Database', got %q", rec1.Category)
	}
	if rec1.Priority != "High" {
		t.Errorf("expected priority 'High', got %q", rec1.Priority)
	}
	if rec1.Title != "Increase pool size" {
		t.Errorf("expected title 'Increase pool size', got %q", rec1.Title)
	}
	if rec1.Confidence != 0.91 {
		t.Errorf("expected confidence 0.91, got %f", rec1.Confidence)
	}
	if rec1.TenantID != tenantID {
		t.Error("expected tenantID to match")
	}
	if rec1.IncidentID != incidentID {
		t.Error("expected incidentID to match")
	}
}

func TestRecommendationService_Recommend_ProviderError(t *testing.T) {
	provider := &mockRecAIProvider{err: errors.New("connection timeout")}
	recs := &mockRecRepo{}
	ctxProvider := &mockRecContextProvider{
		snapshot: []byte(`{}`),
		tenantID: uuid.New(),
	}

	svc := newTestRecService(provider, recs, ctxProvider)
	err := svc.Recommend(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error from provider failure")
	}
	if len(recs.created) != 0 {
		t.Fatal("expected no recommendations on provider failure")
	}
}

func TestRecommendationService_Recommend_ContextError(t *testing.T) {
	provider := &mockRecAIProvider{}
	recs := &mockRecRepo{}
	ctxProvider := &mockRecContextProvider{err: errors.New("incident not found")}

	svc := newTestRecService(provider, recs, ctxProvider)
	err := svc.Recommend(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error from context provider failure")
	}
	if provider.callCount != 0 {
		t.Fatal("provider should not be called on context error")
	}
}

func TestRecommendationService_Recommend_InvalidJSON(t *testing.T) {
	provider := &mockRecAIProvider{output: []byte("not json")}
	recs := &mockRecRepo{}
	ctxProvider := &mockRecContextProvider{
		snapshot: []byte(`{}`),
		tenantID: uuid.New(),
	}

	svc := newTestRecService(provider, recs, ctxProvider)
	err := svc.Recommend(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestRecommendationService_Recommend_ConfidenceClamping(t *testing.T) {
	provider := &mockRecAIProvider{
		output: []byte(`{"recommendations": [{"category": "App", "priority": "Low", "title": "test", "description": "test", "confidence": 2.5, "risk_level": "Low"}]}`),
	}
	recs := &mockRecRepo{}
	ctxProvider := &mockRecContextProvider{
		snapshot: []byte(`{}`),
		tenantID: uuid.New(),
	}

	svc := newTestRecService(provider, recs, ctxProvider)
	err := svc.Recommend(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recs.created[0].Confidence != 1.0 {
		t.Errorf("expected confidence clamped to 1.0, got %f", recs.created[0].Confidence)
	}
}

func TestRecommendationService_Recommend_RepoErrorContinues(t *testing.T) {
	provider := &mockRecAIProvider{
		output: []byte(`{"recommendations": [{"category": "App", "priority": "Low", "title": "test1", "description": "d1", "confidence": 0.5, "risk_level": "Low"}, {"category": "Net", "priority": "High", "title": "test2", "description": "d2", "confidence": 0.8, "risk_level": "High"}]}`),
	}
	recs := &mockRecRepo{err: errors.New("db write error")}
	ctxProvider := &mockRecContextProvider{
		snapshot: []byte(`{}`),
		tenantID: uuid.New(),
	}

	svc := newTestRecService(provider, recs, ctxProvider)
	// Should not return error - continues despite individual repo failures
	err := svc.Recommend(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecommendationService_GetRecommendationsByIncident(t *testing.T) {
	incidentID := uuid.New()
	expected := []*recommendation.Recommendation{
		{ID: uuid.New(), IncidentID: incidentID, Title: "Scale up"},
		{ID: uuid.New(), IncidentID: incidentID, Title: "Add monitoring"},
	}
	recs := &mockRecRepo{listed: expected}
	svc := newTestRecService(&mockRecAIProvider{}, recs, &mockRecContextProvider{})

	result, err := svc.GetRecommendationsByIncident(context.Background(), incidentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 recommendations, got %d", len(result))
	}
}
