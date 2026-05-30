package recommendationcategory

// RecommendationCategory defines the type of recommendation.
type RecommendationCategory string

const (
	CategoryDatabase       RecommendationCategory = "Database"
	CategoryRedis          RecommendationCategory = "Redis"
	CategoryInfrastructure RecommendationCategory = "Infrastructure"
	CategoryApplication    RecommendationCategory = "Application"
	CategoryNetwork        RecommendationCategory = "Network"
	CategoryDeployment     RecommendationCategory = "Deployment"
	CategoryCapacity       RecommendationCategory = "Capacity"
	CategorySecurity       RecommendationCategory = "Security"
)
