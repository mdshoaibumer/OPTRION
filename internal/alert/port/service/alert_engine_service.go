package service

import (
	"context"

	"github.com/optrion/optrion/internal/alert/app/event"
)

type AlertEngineService interface {
	ProcessEvent(ctx context.Context, evt event.IncidentEvent) error
	GetAlertByID(ctx context.Context, id string) (interface{}, error)
	ListAlerts(ctx context.Context, tenantID string) ([]interface{}, error)
}
