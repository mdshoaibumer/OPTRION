package provider

import (
	"context"
)

// AIProvider abstracts LLM providers for root cause analysis.
type AIProvider interface {
	Analyze(ctx context.Context, context []byte) ([]byte, error)
	Name() string
}

// Supported providers: Gemini, OpenAI, Anthropic, Ollama
