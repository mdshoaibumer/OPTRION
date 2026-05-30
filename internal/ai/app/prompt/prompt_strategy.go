package prompt

// PromptStrategy defines how to generate structured, deterministic prompts for AI providers.
type PromptStrategy interface {
	GeneratePrompt(context []byte) (string, error)
}
