package service

import (
	"sync"
	"time"
)

// DeduplicationService suppresses duplicate alerts within a time window.
type DeduplicationService struct {
	mu     sync.Mutex
	cache  map[string]time.Time
	window time.Duration
}

func NewDeduplicationService(window time.Duration) *DeduplicationService {
	return &DeduplicationService{
		cache:  make(map[string]time.Time),
		window: window,
	}
}

// ShouldSuppress returns true if an alert with the same key should be suppressed.
func (ds *DeduplicationService) ShouldSuppress(key string) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	now := time.Now()
	if t, ok := ds.cache[key]; ok && now.Sub(t) < ds.window {
		return true
	}
	ds.cache[key] = now
	return false
}
