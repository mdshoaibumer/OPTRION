package ranking

import (
	"github.com/google/uuid"
)

// RankingStrategy ranks recommendations by impact, confidence, safety, complexity.
type RankingStrategy interface {
	Rank(recommendationIDs []uuid.UUID) ([]uuid.UUID, error)
}
