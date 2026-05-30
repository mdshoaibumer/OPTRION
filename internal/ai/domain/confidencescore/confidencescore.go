package confidencescore

// ConfidenceScore represents the AI's confidence in its analysis.
type ConfidenceScore struct {
	Value float64 // 0.0 - 1.0
	Label string  // e.g., Low, Medium, High
}
