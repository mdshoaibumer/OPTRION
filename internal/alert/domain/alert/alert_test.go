package alert

import (
	"testing"
	"time"
)

func TestAlertStatus_Values(t *testing.T) {
	statuses := []AlertStatus{
		AlertStatusPending,
		AlertStatusSent,
		AlertStatusDelivered,
		AlertStatusFailed,
		AlertStatusSuppressed,
	}

	for _, s := range statuses {
		if s == "" {
			t.Fatal("alert status should not be empty")
		}
	}
}

func TestAlert_Struct(t *testing.T) {
	a := Alert{
		ID:         "alert-1",
		TenantID:   "tenant-1",
		RuleID:     "rule-1",
		IncidentID: "incident-1",
		Severity:   "critical",
		Status:     AlertStatusPending,
		Message:    "Database connection pool exhausted",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if a.ID == "" {
		t.Fatal("alert ID should not be empty")
	}
	if a.Status != AlertStatusPending {
		t.Fatalf("expected status pending, got %s", a.Status)
	}
}
