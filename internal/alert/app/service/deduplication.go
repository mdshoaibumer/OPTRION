package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// DeduplicationService suppresses duplicate alerts within a time window.
// Uses Redis for persistent state that survives server restarts.
// Falls back to in-memory cache if Redis is unavailable.
type DeduplicationService struct {
	mu     sync.Mutex
	cache  map[string]time.Time
	window time.Duration
	redis  *redis.Client
}

func NewDeduplicationService(window time.Duration) *DeduplicationService {
	return &DeduplicationService{
		cache:  make(map[string]time.Time),
		window: window,
	}
}

// WithRedis enables Redis-backed deduplication for persistence across restarts.
func (ds *DeduplicationService) WithRedis(client *redis.Client) *DeduplicationService {
	ds.redis = client
	return ds
}

// ShouldSuppress returns true if an alert with the same key should be suppressed.
func (ds *DeduplicationService) ShouldSuppress(key string) bool {
	// Try Redis first if available
	if ds.redis != nil {
		return ds.shouldSuppressRedis(key)
	}
	return ds.shouldSuppressMemory(key)
}

func (ds *DeduplicationService) shouldSuppressRedis(key string) bool {
	ctx := context.Background()
	redisKey := fmt.Sprintf("optrion:dedup:%s", key)

	// Try to set the key with NX (only if not exists) and TTL
	set, err := ds.redis.SetNX(ctx, redisKey, "1", ds.window).Result()
	if err != nil {
		// Redis unavailable — fall back to in-memory
		return ds.shouldSuppressMemory(key)
	}

	// If set==true, key was NEW (not suppressed); if false, key already existed (suppress)
	return !set
}

func (ds *DeduplicationService) shouldSuppressMemory(key string) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	now := time.Now()
	if t, ok := ds.cache[key]; ok && now.Sub(t) < ds.window {
		return true
	}
	ds.cache[key] = now
	return false
}
