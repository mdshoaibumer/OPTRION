package server

import (
	"testing"
	"time"
)

func TestAuthFailureTracker_NotLockedInitially(t *testing.T) {
	tracker := NewAuthFailureTracker(5, 15*time.Minute)

	if tracker.IsLocked("192.168.1.1") {
		t.Fatal("new IP should not be locked")
	}
}

func TestAuthFailureTracker_LockAfterMaxAttempts(t *testing.T) {
	tracker := NewAuthFailureTracker(3, 15*time.Minute)

	ip := "192.168.1.100"
	tracker.RecordFailure(ip)
	tracker.RecordFailure(ip)
	locked := tracker.RecordFailure(ip) // 3rd attempt should lock

	if !locked {
		t.Fatal("should be locked after 3 failures")
	}
	if !tracker.IsLocked(ip) {
		t.Fatal("IP should be locked")
	}
}

func TestAuthFailureTracker_NotLockedBeforeMax(t *testing.T) {
	tracker := NewAuthFailureTracker(5, 15*time.Minute)

	ip := "10.0.0.1"
	tracker.RecordFailure(ip)
	tracker.RecordFailure(ip)
	tracker.RecordFailure(ip)

	if tracker.IsLocked(ip) {
		t.Fatal("should not be locked before reaching max attempts")
	}
}

func TestAuthFailureTracker_SuccessClearsRecord(t *testing.T) {
	tracker := NewAuthFailureTracker(3, 15*time.Minute)

	ip := "172.16.0.1"
	tracker.RecordFailure(ip)
	tracker.RecordFailure(ip)
	tracker.RecordSuccess(ip) // Success should clear

	// Should be able to fail again without immediate lockout
	locked := tracker.RecordFailure(ip)
	if locked {
		t.Fatal("should not be locked after success cleared the record")
	}
}

func TestAuthFailureTracker_DifferentIPsIndependent(t *testing.T) {
	tracker := NewAuthFailureTracker(2, 15*time.Minute)

	ip1 := "1.1.1.1"
	ip2 := "2.2.2.2"

	tracker.RecordFailure(ip1)
	tracker.RecordFailure(ip1) // ip1 locked

	if tracker.IsLocked(ip2) {
		t.Fatal("ip2 should not be affected by ip1 lockout")
	}
}

func TestAuthFailureTracker_LockoutExpires(t *testing.T) {
	tracker := NewAuthFailureTracker(2, 1*time.Millisecond) // Very short lockout

	ip := "3.3.3.3"
	tracker.RecordFailure(ip)
	tracker.RecordFailure(ip)

	if !tracker.IsLocked(ip) {
		t.Fatal("should be locked immediately")
	}

	time.Sleep(5 * time.Millisecond)

	if tracker.IsLocked(ip) {
		t.Fatal("lockout should have expired")
	}
}
