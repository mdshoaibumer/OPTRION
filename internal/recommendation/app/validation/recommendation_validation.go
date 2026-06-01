package validation

import (
	"fmt"
	"strings"
)

// RecommendationValidator validates recommendations for safety and hallucination control.
type RecommendationValidator interface {
	Validate(recommendation string, evidence []string) error
}

// DefaultRecommendationValidator is the production implementation of safety validation.
type DefaultRecommendationValidator struct {
	dangerousPatterns       []string
	maxRecommendationLength int
	minEvidenceRequired     int
}

// NewDefaultRecommendationValidator creates a validator with safety defaults.
func NewDefaultRecommendationValidator() *DefaultRecommendationValidator {
	return &DefaultRecommendationValidator{
		dangerousPatterns: []string{
			"DROP DATABASE",
			"DROP TABLE",
			"DELETE FROM",
			"TRUNCATE TABLE",
			"DISABLE SECURITY",
			"DISABLE AUTHENTICATION",
			"DISABLE AUTH",
			"REMOVE FIREWALL",
			"DISABLE FIREWALL",
			"RM -RF",
			"CHMOD 777",
			"FORMAT C:",
			"DESTROY",
			"DELETE ALL",
			"WIPE",
			"PURGE ALL",
			"SHUTDOWN",
			"KILL -9",
			"DISABLE ENCRYPTION",
			"DISABLE SSL",
			"DISABLE TLS",
			"ALLOW ALL",
			"GRANT ALL PRIVILEGES",
			"SET PASSWORD ''",
			"NO PASSWORD",
		},
		maxRecommendationLength: 2000,
		minEvidenceRequired:     1,
	}
}

// Validate checks a recommendation for safety, hallucination, and evidence requirements.
func (v *DefaultRecommendationValidator) Validate(recommendation string, evidence []string) error {
	if strings.TrimSpace(recommendation) == "" {
		return fmt.Errorf("recommendation cannot be empty")
	}

	if len(recommendation) > v.maxRecommendationLength {
		return fmt.Errorf("recommendation exceeds maximum length of %d characters", v.maxRecommendationLength)
	}

	// Require evidence to support the recommendation
	if len(evidence) < v.minEvidenceRequired {
		return fmt.Errorf("recommendation requires at least %d piece(s) of evidence", v.minEvidenceRequired)
	}

	// Check for empty or trivial evidence
	validEvidence := 0
	for _, e := range evidence {
		if len(strings.TrimSpace(e)) > 10 {
			validEvidence++
		}
	}
	if validEvidence < v.minEvidenceRequired {
		return fmt.Errorf("recommendation evidence is insufficient (trivial or empty)")
	}

	// Check for dangerous patterns (patterns are stored pre-uppercased)
	upperRec := strings.ToUpper(recommendation)
	for _, pattern := range v.dangerousPatterns {
		if strings.Contains(upperRec, pattern) {
			return fmt.Errorf("recommendation contains dangerous operation: %s", pattern)
		}
	}

	// Check for hallucination indicators
	if err := v.checkHallucination(recommendation, evidence); err != nil {
		return err
	}

	return nil
}

// checkHallucination detects recommendations that appear fabricated.
func (v *DefaultRecommendationValidator) checkHallucination(recommendation string, evidence []string) error {
	// A recommendation that contradicts its own evidence
	lowerRec := strings.ToLower(recommendation)

	// Check if recommendation references things not in evidence
	specifics := []string{"version", "CVE-", "bug #", "issue #"}
	for _, s := range specifics {
		if strings.Contains(lowerRec, strings.ToLower(s)) {
			// If a specific reference is made, it should appear in at least one evidence item
			found := false
			for _, e := range evidence {
				if strings.Contains(strings.ToLower(e), strings.ToLower(s)) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("recommendation references %q but no evidence supports it", s)
			}
		}
	}

	return nil
}
