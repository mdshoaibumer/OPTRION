package service

import (
	"context"
	"github.com/optrion/optrion/internal/alert/app/event"

	"github.com/google/uuid"
)

type AlertEngineService interface {
	ProcessEvent(ctx context.Context, evt event.IncidentEvent) error
	GetAlertByID(ctx context.Context, id uuid.UUID) (interface{}, error)
	ListAlerts(ctx context.Context, tenantID uuid.UUID) ([]interface{}, error)
}
