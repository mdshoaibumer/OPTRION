package server

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter implements distributed rate limiting using Redis.
// Uses the sliding window counter algorithm for accurate rate limiting
// across multiple application instances.
type RedisRateLimiter struct {
	client  *redis.Client
	rate    int           // requests per second
	window  time.Duration // sliding window duration
	counter atomic.Int64  // unique member suffix
}

// NewRedisRateLimiter creates a new Redis-backed rate limiter.
func NewRedisRateLimiter(client *redis.Client, rps int) *RedisRateLimiter {
	if rps <= 0 {
		rps = 100
	}
	return &RedisRateLimiter{
		client: client,
		rate:   rps,
		window: time.Second,
	}
}

// Allow checks if a request should be allowed using a Redis sliding window counter.
func (rl *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := fmt.Sprintf("ratelimit:%s", key)
	now := time.Now().UnixMilli()
	windowStart := now - rl.window.Milliseconds()
	seq := rl.counter.Add(1)

	pipe := rl.client.Pipeline()

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart))

	// Count current window entries
	countCmd := pipe.ZCard(ctx, redisKey)

	// Add current request with unique member
	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d:%d", now, seq),
	})

	// Set expiry on the key
	pipe.Expire(ctx, redisKey, rl.window*2)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("redis rate limiter error: %w", err)
	}

	count := countCmd.Val()
	return count < int64(rl.rate), nil
}
