package provider

import (
	"context"
	"fmt"
	"log/slog"
)

// FallbackProvider wraps multiple AI providers and tries them in order.
// If the primary provider fails, it falls back to the next provider in the chain.
type FallbackProvider struct {
	providers []AIProvider
	logger    *slog.Logger
}

// NewFallbackProvider creates a provider that tries multiple providers in order.
// At least one provider must be supplied.
func NewFallbackProvider(logger *slog.Logger, providers ...AIProvider) (*FallbackProvider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("at least one AI provider is required")
	}
	return &FallbackProvider{
		providers: providers,
		logger:    logger,
	}, nil
}

// Analyze tries each provider in order until one succeeds.
func (fp *FallbackProvider) Analyze(ctx context.Context, analysisContext []byte) ([]byte, error) {
	var lastErr error

	for _, p := range fp.providers {
		result, err := p.Analyze(ctx, analysisContext)
		if err == nil {
			return result, nil
		}

		lastErr = err
		fp.logger.Warn("AI provider failed, trying next",
			"provider", p.Name(),
			"error", err,
		)
	}

	return nil, fmt.Errorf("all AI providers failed, last error: %w", lastErr)
}

// Name returns a description of the fallback chain.
func (fp *FallbackProvider) Name() string {
	if len(fp.providers) > 0 {
		return fmt.Sprintf("fallback(%s+%d)", fp.providers[0].Name(), len(fp.providers)-1)
	}
	return "fallback(empty)"
}
