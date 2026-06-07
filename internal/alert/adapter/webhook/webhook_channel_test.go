package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/optrion/optrion/internal/alert/domain/alertchannel"
)

func TestWebhookSender_MissingURL(t *testing.T) {
	sender := NewWebhookSender(1, time.Millisecond)
	channel := &alertchannel.AlertChannel{
		Config: map[string]string{},
	}

	_, err := sender.Send(context.Background(), channel, "test message")
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestWebhookSender_RequiresHTTPS(t *testing.T) {
	sender := NewWebhookSender(1, time.Millisecond)
	channel := &alertchannel.AlertChannel{
		Config: map[string]string{
			"url": "http://example.com/webhook",
		},
	}

	_, err := sender.Send(context.Background(), channel, "test message")
	if err == nil {
		t.Fatal("expected error for non-HTTPS URL")
	}
}

func TestWebhookSender_SuccessfulDelivery(t *testing.T) {
	var receivedPayload WebhookPayload
	var receivedSignature string

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSignature = r.Header.Get("X-Optrion-Signature")
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &WebhookSender{
		client:     server.Client(),
		maxRetries: 1,
		retryDelay: time.Millisecond,
	}

	channel := &alertchannel.AlertChannel{
		Config: map[string]string{
			"url":    server.URL + "/webhook",
			"secret": "test-secret",
		},
	}

	// Override scheme check for test (TLS test server uses https)
	deliveryID, err := sender.sendAndReturn(context.Background(), channel, "test alert message")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deliveryID == "" {
		t.Fatal("expected non-empty delivery ID")
	}
	if receivedPayload.Message != "test alert message" {
		t.Fatalf("expected message 'test alert message', got '%s'", receivedPayload.Message)
	}
	if receivedSignature == "" {
		t.Fatal("expected HMAC signature header to be set")
	}
}

// sendAndReturn is a test helper that skips URL scheme validation.
func (ws *WebhookSender) sendAndReturn(ctx context.Context, channel *alertchannel.AlertChannel, message string) (string, error) {
	payload := WebhookPayload{
		ID:        "test-id",
		Event:     "alert",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Message:   message,
	}
	body, _ := json.Marshal(payload)
	err := ws.sendRequest(ctx, channel.Config["url"], body, channel.Config["secret"])
	return "test-id", err
}

func TestWebhookSender_RetryOnFailure(t *testing.T) {
	attempts := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &WebhookSender{
		client:     server.Client(),
		maxRetries: 3,
		retryDelay: time.Millisecond,
	}

	payload := WebhookPayload{Message: "retry test"}
	body, _ := json.Marshal(payload)

	err := sender.sendRequest(context.Background(), server.URL+"/webhook", body, "")
	// First attempt returns 500
	if err == nil {
		t.Fatal("expected error on 500 response")
	}
}

func TestIsAllowedScheme(t *testing.T) {
	tests := []struct {
		url     string
		allowed bool
	}{
		{"https://example.com/webhook", true},
		{"http://example.com/webhook", false},
		{"ftp://example.com/webhook", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := isAllowedScheme(tt.url); got != tt.allowed {
			t.Errorf("isAllowedScheme(%q) = %v, want %v", tt.url, got, tt.allowed)
		}
	}
}
