package provider

import (
	"context"
)

type GeminiProvider struct{}

func (g *GeminiProvider) Analyze(ctx context.Context, context []byte) ([]byte, error) {
	// TODO: Implement Gemini API call
	return nil, nil
}

func (g *GeminiProvider) Name() string {
	return "Gemini"
}
