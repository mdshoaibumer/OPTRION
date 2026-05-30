package validation

// RecommendationValidator validates recommendations for safety and hallucination control.
type RecommendationValidator interface {
	Validate(recommendation string, evidence []string) error
}
