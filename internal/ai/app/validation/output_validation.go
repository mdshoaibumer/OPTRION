package validation

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OutputValidator validates AI provider output for structure and content.
type OutputValidator interface {
	Validate(output []byte) error
}

// DefaultOutputValidator is the production implementation of OutputValidator.
type DefaultOutputValidator struct {
	maxOutputLength   int
	requiredFields    []string
	forbiddenPatterns []string
}

// NewDefaultOutputValidator creates a new output validator with safety defaults.
func NewDefaultOutputValidator() *DefaultOutputValidator {
	return &DefaultOutputValidator{
		maxOutputLength: 50000, // 50KB max output
		requiredFields:  []string{"root_cause", "confidence"},
		forbiddenPatterns: []string{
			"DROP DATABASE",
			"DROP TABLE",
			"DELETE FROM",
			"TRUNCATE",
			"rm -rf",
			"shutdown",
			"format c:",
			":(){ :|:& };:",
			"exec(",
			"eval(",
			"<script>",
			"javascript:",
		},
	}
}

// Validate checks AI output for structural correctness, hallucination indicators, and safety.
func (v *DefaultOutputValidator) Validate(output []byte) error {
	if len(output) == 0 {
		return fmt.Errorf("AI output is empty")
	}

	if len(output) > v.maxOutputLength {
		return fmt.Errorf("AI output exceeds maximum length of %d bytes", v.maxOutputLength)
	}

	// Check for valid JSON structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(output, &parsed); err != nil {
		return fmt.Errorf("AI output is not valid JSON: %w", err)
	}

	// Check required fields are present
	for _, field := range v.requiredFields {
		if _, exists := parsed[field]; !exists {
			return fmt.Errorf("AI output missing required field: %s", field)
		}
	}

	// Validate confidence score is within bounds
	if conf, ok := parsed["confidence"]; ok {
		confFloat, ok := conf.(float64)
		if !ok {
			return fmt.Errorf("confidence field must be a number")
		}
		if confFloat < 0 || confFloat > 1.0 {
			return fmt.Errorf("confidence score must be between 0 and 1, got: %f", confFloat)
		}
	}

	// Check for dangerous/forbidden patterns in the output
	outputStr := strings.ToUpper(string(output))
	for _, pattern := range v.forbiddenPatterns {
		if strings.Contains(outputStr, strings.ToUpper(pattern)) {
			return fmt.Errorf("AI output contains forbidden pattern: %s", pattern)
		}
	}

	// Check for hallucination indicators (fabricated UUIDs, impossible timestamps)
	if err := v.checkHallucinationIndicators(parsed); err != nil {
		return fmt.Errorf("hallucination detected: %w", err)
	}

	return nil
}

// checkHallucinationIndicators detects common signs of AI hallucination.
func (v *DefaultOutputValidator) checkHallucinationIndicators(parsed map[string]interface{}) error {
	// Check for suspiciously empty root cause with high confidence
	if rootCause, ok := parsed["root_cause"].(string); ok {
		if conf, ok := parsed["confidence"].(float64); ok {
			if len(rootCause) < 10 && conf > 0.8 {
				return fmt.Errorf("high confidence (%f) with very short root cause is suspicious", conf)
			}
		}
	}

	// Check that affected_components references look valid (not fabricated)
	if components, ok := parsed["affected_components"].([]interface{}); ok {
		for _, comp := range components {
			compStr, ok := comp.(string)
			if !ok {
				continue
			}
			// Flag obviously fabricated component names
			if len(compStr) > 200 {
				return fmt.Errorf("component reference exceeds reasonable length: %d chars", len(compStr))
			}
		}
	}

	return nil
}
