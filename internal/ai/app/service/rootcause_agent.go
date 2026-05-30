package service

import (
	"context"

	"github.com/google/uuid"
)

// RootCauseAgent analyzes incidents and produces root cause reports.
type RootCauseAgent interface {
	Analyze(ctx context.Context, incidentID uuid.UUID) error
}
