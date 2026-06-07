package retention

import (
	"testing"
	"time"
)

func TestDefaultPolicy(t *testing.T) {
	policy := DefaultPolicy()

	if policy.MetricSnapshots != 90*24*time.Hour {
		t.Fatalf("expected 90 days for metric snapshots, got %v", policy.MetricSnapshots)
	}
	if policy.ResolvedIncidents != 180*24*time.Hour {
		t.Fatalf("expected 180 days for resolved incidents, got %v", policy.ResolvedIncidents)
	}
	if policy.AuditEvents != 365*24*time.Hour {
		t.Fatalf("expected 365 days for audit events, got %v", policy.AuditEvents)
	}
	if policy.AlertHistory != 90*24*time.Hour {
		t.Fatalf("expected 90 days for alert history, got %v", policy.AlertHistory)
	}
	if policy.AIAnalyses != 90*24*time.Hour {
		t.Fatalf("expected 90 days for AI analyses, got %v", policy.AIAnalyses)
	}
}

func TestCleanupJob_Creation(t *testing.T) {
	policy := DefaultPolicy()
	job := NewCleanupJob(nil, policy, nil)

	if job == nil {
		t.Fatal("expected non-nil cleanup job")
	}
	if job.policy.MetricSnapshots != policy.MetricSnapshots {
		t.Fatal("policy mismatch")
	}
}

func TestCleanupJob_Stop(t *testing.T) {
	policy := DefaultPolicy()
	job := NewCleanupJob(nil, policy, nil)

	// Should not panic on Stop
	job.Stop()
}
