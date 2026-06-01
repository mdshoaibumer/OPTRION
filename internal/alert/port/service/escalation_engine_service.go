package service

import (
	"context"
)

type EscalationEngineService interface {
	StartEscalation(ctx context.Context, alertID string) error
}
