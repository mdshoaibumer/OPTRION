package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/optrion/optrion/internal/incident/app"
)

// DefaultIncidentContextProvider builds context snapshots from incident data.
type DefaultIncidentContextProvider struct {
	incidentService *app.IncidentService
}

// NewIncidentContextProvider creates a new provider that queries incident data.
func NewIncidentContextProvider(svc *app.IncidentService) *DefaultIncidentContextProvider {
	return &DefaultIncidentContextProvider{incidentService: svc}
}

// GetIncidentContext builds a JSON context snapshot for AI analysis.
func (p *DefaultIncidentContextProvider) GetIncidentContext(ctx context.Context, incidentID uuid.UUID) ([]byte, uuid.UUID, error) {
	incident, err := p.incidentService.GetIncident(ctx, incidentID.String())
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("getting incident: %w", err)
	}
	if incident == nil {
		return nil, uuid.Nil, fmt.Errorf("incident %s not found", incidentID)
	}

	tenantID, _ := uuid.Parse(incident.TenantID)

	// Get timeline for additional context
	timeline, _ := p.incidentService.GetTimeline(ctx, incidentID.String(), 50)

	// Get comments
	comments, _ := p.incidentService.GetComments(ctx, incidentID.String())

	contextData := map[string]interface{}{
		"incident": map[string]interface{}{
			"id":           incident.ID,
			"title":        incident.Title,
			"severity":     incident.Severity,
			"status":       incident.Status,
			"component_id": incident.ComponentID,
			"tenant_id":    incident.TenantID,
			"created_at":   incident.CreatedAt,
			"description":  incident.Description,
		},
		"timeline_events": len(timeline),
		"comments":        len(comments),
	}

	if len(timeline) > 0 {
		var events []map[string]interface{}
		for _, e := range timeline {
			events = append(events, map[string]interface{}{
				"type":        string(e.EntryType),
				"title":       e.Title,
				"occurred_at": e.OccurredAt,
			})
		}
		contextData["timeline"] = events
	}

	snapshot, err := json.Marshal(contextData)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("marshaling context: %w", err)
	}

	return snapshot, tenantID, nil
}
