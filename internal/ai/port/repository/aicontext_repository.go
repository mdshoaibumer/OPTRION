package repository

import (
	"context"
	"github.com/optrion/optrion/internal/ai/domain/aicontext"

	"github.com/google/uuid"
)

type AIContextRepository interface {
	Create(ctx context.Context, c *aicontext.AIContext) error
	FindByID(ctx context.Context, id uuid.UUID) (*aicontext.AIContext, error)
}
