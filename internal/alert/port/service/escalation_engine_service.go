package service

import (
	"context"

	"github.com/google/uuid"
)

type EscalationEngineService interface {
	StartEscalation(ctx context.Context, alertID uuid.UUID) error
}
