package provider

import (
	"context"
)

type AnthropicProvider struct{}

func (a *AnthropicProvider) Analyze(ctx context.Context, context []byte) ([]byte, error) {
	// TODO: Implement Anthropic API call
	return nil, nil
}

func (a *AnthropicProvider) Name() string {
	return "Anthropic"
}
