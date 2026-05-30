package provider

import (
	"context"
)

type OllamaProvider struct{}

func (o *OllamaProvider) Analyze(ctx context.Context, context []byte) ([]byte, error) {
	// TODO: Implement Ollama API call
	return nil, nil
}

func (o *OllamaProvider) Name() string {
	return "Ollama"
}
