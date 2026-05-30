package recommendationpriority

// RecommendationPriority defines the priority of a recommendation.
type RecommendationPriority string

const (
	PriorityCritical RecommendationPriority = "Critical"
	PriorityHigh     RecommendationPriority = "High"
	PriorityMedium   RecommendationPriority = "Medium"
	PriorityLow      RecommendationPriority = "Low"
)
