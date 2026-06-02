package provider

import (
	"context"
	"fmt"

	"github.com/optrion/optrion/internal/platform/circuitbreaker"
)

// ResilientProvider wraps an AIProvider with a circuit breaker for fault tolerance.
// When the underlying provider fails repeatedly, the circuit opens and rejects
// requests immediately — preventing cascade failures and API cost accumulation.
type ResilientProvider struct {
	inner   AIProvider
	breaker *circuitbreaker.CircuitBreaker
}

// NewResilientProvider wraps an AI provider with circuit breaker protection.
func NewResilientProvider(inner AIProvider, cfg circuitbreaker.Config) *ResilientProvider {
	return &ResilientProvider{
		inner:   inner,
		breaker: circuitbreaker.New(cfg),
	}
}

// Analyze delegates to the inner provider through the circuit breaker.
func (rp *ResilientProvider) Analyze(ctx context.Context, analysisContext []byte) ([]byte, error) {
	var result []byte

	err := rp.breaker.Execute(func() error {
		var innerErr error
		result, innerErr = rp.inner.Analyze(ctx, analysisContext)
		return innerErr
	})

	if err != nil {
		if err == circuitbreaker.ErrCircuitOpen {
			return nil, fmt.Errorf("%s: circuit breaker open — provider is unavailable, will retry after cooldown", rp.inner.Name())
		}
		return nil, err
	}

	return result, nil
}

// Name returns the inner provider name with resilient wrapper indication.
func (rp *ResilientProvider) Name() string {
	return rp.inner.Name()
}

// State returns the current circuit breaker state for observability.
func (rp *ResilientProvider) State() string {
	return rp.breaker.State().String()
}

// Reset manually resets the circuit breaker (useful for admin operations).
func (rp *ResilientProvider) Reset() {
	rp.breaker.Reset()
}
