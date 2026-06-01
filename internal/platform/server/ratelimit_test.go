package server

import (
	"context"
	"testing"
)

func TestInMemoryRateLimiter_AllowsInitialRequests(t *testing.T) {
	rl := NewInMemoryRateLimiter(10)

	allowed, err := rl.Allow(context.Background(), "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("first request should be allowed")
	}
}

func TestInMemoryRateLimiter_ExhaustsBurst(t *testing.T) {
	rl := NewInMemoryRateLimiter(5) // 5 rps, burst of 10

	key := "exhaust-test"
	// Exhaust the burst (5 * 2 = 10 tokens)
	for i := 0; i < 10; i++ {
		allowed, err := rl.Allow(context.Background(), key)
		if err != nil {
			t.Fatalf("unexpected error on request %d: %v", i, err)
		}
		if !allowed {
			t.Fatalf("request %d should have been allowed (within burst)", i)
		}
	}

	// Next request should be rate limited
	allowed, err := rl.Allow(context.Background(), key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("request should be denied after burst exhausted")
	}
}

func TestInMemoryRateLimiter_IndependentKeys(t *testing.T) {
	rl := NewInMemoryRateLimiter(2) // 2 rps, burst of 4

	// Exhaust key1
	for i := 0; i < 4; i++ {
		rl.Allow(context.Background(), "key1")
	}

	// key2 should still be allowed
	allowed, err := rl.Allow(context.Background(), "key2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("key2 should not be affected by key1 exhaustion")
	}
}

func TestInMemoryRateLimiter_DefaultsToHundred(t *testing.T) {
	rl := NewInMemoryRateLimiter(0) // Invalid, should default to 100

	if rl.rate != 100 {
		t.Fatalf("expected default rate of 100, got %d", rl.rate)
	}
}
