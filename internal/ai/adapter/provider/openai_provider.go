package provider

import (
	"context"
)

type OpenAIProvider struct{}

func (o *OpenAIProvider) Analyze(ctx context.Context, context []byte) ([]byte, error) {
	// TODO: Implement OpenAI API call
	return nil, nil
}

func (o *OpenAIProvider) Name() string {
	return "OpenAI"
}
