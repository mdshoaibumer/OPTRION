package service

import (
	"testing"
	"time"
)

func TestDeduplicationService_FirstCallNotSuppressed(t *testing.T) {
	dedup := NewDeduplicationService(5 * time.Minute)

	if dedup.ShouldSuppress("key1") {
		t.Fatal("first call should not be suppressed")
	}
}

func TestDeduplicationService_SecondCallSuppressed(t *testing.T) {
	dedup := NewDeduplicationService(5 * time.Minute)

	dedup.ShouldSuppress("key1")
	if !dedup.ShouldSuppress("key1") {
		t.Fatal("duplicate within window should be suppressed")
	}
}

func TestDeduplicationService_DifferentKeysNotSuppressed(t *testing.T) {
	dedup := NewDeduplicationService(5 * time.Minute)

	dedup.ShouldSuppress("key1")
	if dedup.ShouldSuppress("key2") {
		t.Fatal("different key should not be suppressed")
	}
}

func TestDeduplicationService_ExpiredWindowNotSuppressed(t *testing.T) {
	dedup := NewDeduplicationService(1 * time.Millisecond)

	dedup.ShouldSuppress("key1")
	time.Sleep(5 * time.Millisecond)

	if dedup.ShouldSuppress("key1") {
		t.Fatal("expired key should not be suppressed")
	}
}
