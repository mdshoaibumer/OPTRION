package validation

// OutputValidator validates AI provider output for structure and content.
type OutputValidator interface {
	Validate(output []byte) error
}
