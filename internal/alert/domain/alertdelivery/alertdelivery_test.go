package alertdelivery_test

import (
	"testing"
	"time"

	"github.com/optrion/optrion/internal/alert/domain/alertdelivery"
)

func TestAlertDelivery_StatusTransitions(t *testing.T) {
	delivery := alertdelivery.AlertDelivery{
		ID:        "d-1",
		TenantID:  "t-1",
		AlertID:   "a-1",
		ChannelID: "ch-1",
		Status:    alertdelivery.DeliveryStatusPending,
		Attempts:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if delivery.Status != alertdelivery.DeliveryStatusPending {
		t.Errorf("expected pending, got %s", delivery.Status)
	}

	// Simulate retry
	delivery.Status = alertdelivery.DeliveryStatusRetrying
	delivery.Attempts++
	if delivery.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", delivery.Attempts)
	}

	// Simulate success
	delivery.Status = alertdelivery.DeliveryStatusDelivered
	if delivery.Status != alertdelivery.DeliveryStatusDelivered {
		t.Errorf("expected delivered, got %s", delivery.Status)
	}
}

func TestAlertDelivery_FailureTracking(t *testing.T) {
	delivery := alertdelivery.AlertDelivery{
		ID:       "d-1",
		Status:   alertdelivery.DeliveryStatusPending,
		Attempts: 0,
	}

	// Simulate 3 failed attempts
	for i := 0; i < 3; i++ {
		delivery.Attempts++
		delivery.Status = alertdelivery.DeliveryStatusRetrying
		delivery.LastError = "connection timeout"
	}

	// Final failure
	delivery.Status = alertdelivery.DeliveryStatusFailed

	if delivery.Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", delivery.Attempts)
	}
	if delivery.Status != alertdelivery.DeliveryStatusFailed {
		t.Errorf("expected failed, got %s", delivery.Status)
	}
	if delivery.LastError != "connection timeout" {
		t.Errorf("expected last error, got %s", delivery.LastError)
	}
}

func TestDeliveryStatus_Constants(t *testing.T) {
	statuses := []alertdelivery.DeliveryStatus{
		alertdelivery.DeliveryStatusPending,
		alertdelivery.DeliveryStatusSent,
		alertdelivery.DeliveryStatusDelivered,
		alertdelivery.DeliveryStatusFailed,
		alertdelivery.DeliveryStatusRetrying,
	}

	expected := []string{"pending", "sent", "delivered", "failed", "retrying"}
	for i, s := range statuses {
		if string(s) != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], s)
		}
	}
}

func TestDeliveryHistory(t *testing.T) {
	history := []alertdelivery.DeliveryHistory{
		{Timestamp: time.Now().Add(-2 * time.Minute), Status: alertdelivery.DeliveryStatusPending, Error: ""},
		{Timestamp: time.Now().Add(-1 * time.Minute), Status: alertdelivery.DeliveryStatusRetrying, Error: "timeout"},
		{Timestamp: time.Now(), Status: alertdelivery.DeliveryStatusDelivered, Error: ""},
	}

	if len(history) != 3 {
		t.Fatalf("expected 3 history entries, got %d", len(history))
	}
	if history[1].Error != "timeout" {
		t.Errorf("expected timeout error, got %s", history[1].Error)
	}
}
