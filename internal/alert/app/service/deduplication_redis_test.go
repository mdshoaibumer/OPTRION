package service_test

import (
	"testing"
	"time"

	"github.com/optrion/optrion/internal/alert/app/service"
)

func TestDeduplicationService_ShouldSuppress_FirstCall(t *testing.T) {
	ds := service.NewDeduplicationService(5 * time.Minute)

	// First call should NOT be suppressed
	if ds.ShouldSuppress("key1") {
		t.Error("first call should not be suppressed")
	}
}

func TestDeduplicationService_ShouldSuppress_DuplicateWithinWindow(t *testing.T) {
	ds := service.NewDeduplicationService(5 * time.Minute)

	// First call
	ds.ShouldSuppress("key1")

	// Second call with same key should be suppressed
	if !ds.ShouldSuppress("key1") {
		t.Error("duplicate within window should be suppressed")
	}
}

func TestDeduplicationService_ShouldSuppress_DifferentKeys(t *testing.T) {
	ds := service.NewDeduplicationService(5 * time.Minute)

	ds.ShouldSuppress("key1")

	// Different key should NOT be suppressed
	if ds.ShouldSuppress("key2") {
		t.Error("different key should not be suppressed")
	}
}

func TestDeduplicationService_ShouldSuppress_AfterWindowExpires(t *testing.T) {
	// Use a very short window for testing
	ds := service.NewDeduplicationService(1 * time.Millisecond)

	ds.ShouldSuppress("key1")

	// Wait for window to expire
	time.Sleep(5 * time.Millisecond)

	// Should no longer be suppressed
	if ds.ShouldSuppress("key1") {
		t.Error("key should not be suppressed after window expires")
	}
}

func TestDeduplicationService_ShouldSuppress_ConcurrentAccess(t *testing.T) {
	ds := service.NewDeduplicationService(5 * time.Minute)

	// Simulate concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := "concurrent-key"
			ds.ShouldSuppress(key)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// After all goroutines, key should be suppressed
	if !ds.ShouldSuppress("concurrent-key") {
		t.Error("key should be suppressed after concurrent writes")
	}
}

func TestDeduplicationService_WithRedis_NilClient(t *testing.T) {
	ds := service.NewDeduplicationService(5 * time.Minute)
	// WithRedis(nil) should still work using in-memory fallback
	ds.WithRedis(nil)

	if ds.ShouldSuppress("key1") {
		t.Error("first call should not be suppressed")
	}
	if !ds.ShouldSuppress("key1") {
		t.Error("duplicate should be suppressed with in-memory fallback")
	}
}
