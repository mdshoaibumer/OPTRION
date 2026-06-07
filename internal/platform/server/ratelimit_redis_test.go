package server

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisRateLimiter_AllowsInitialRequests(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	rl := NewRedisRateLimiter(client, 10)

	allowed, err := rl.Allow(context.Background(), "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("first request should be allowed")
	}
}

func TestRedisRateLimiter_ExhaustsLimit(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	rl := NewRedisRateLimiter(client, 5)

	key := "exhaust-test"
	for i := 0; i < 5; i++ {
		allowed, err := rl.Allow(context.Background(), key)
		if err != nil {
			t.Fatalf("unexpected error on request %d: %v", i, err)
		}
		if !allowed {
			t.Fatalf("request %d should have been allowed (within limit)", i)
		}
	}

	// Next request should be rate limited
	allowed, err := rl.Allow(context.Background(), key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("request should be denied after limit exhausted")
	}
}

func TestRedisRateLimiter_IndependentKeys(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	rl := NewRedisRateLimiter(client, 2)

	// Exhaust key1
	for i := 0; i < 2; i++ {
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

func TestRedisRateLimiter_DefaultsToHundred(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	rl := NewRedisRateLimiter(client, 0)
	if rl.rate != 100 {
		t.Fatalf("expected default rate of 100, got %d", rl.rate)
	}
}
