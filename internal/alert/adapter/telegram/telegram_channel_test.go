package telegram_test

import (
	"context"
	"testing"
	"time"

	"github.com/optrion/optrion/internal/alert/adapter/telegram"
	"github.com/optrion/optrion/internal/alert/domain/alertchannel"
)

func TestTelegramSender_Send_NetworkError(t *testing.T) {
	// Use a very short retry delay to speed up the test
	sender := telegram.NewTelegramSender(30, 1*time.Millisecond)

	channel := &alertchannel.AlertChannel{
		ID:       "ch-1",
		TenantID: "t-1",
		Type:     alertchannel.ChannelTypeTelegram,
		Config: map[string]string{
			"bot_token": "fake-token",
			"chat_id":   "12345",
		},
	}

	// Use a context with tight timeout to prevent waiting for real API
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := sender.Send(ctx, channel, "Test alert message")
	// Should fail because it can't connect (fake token, network timeout)
	if err == nil {
		t.Log("unexpectedly succeeded - may be in environment with outbound access")
	} else {
		// Verify error message is sanitized (no bot token leakage)
		errMsg := err.Error()
		if contains(errMsg, "fake-token") {
			t.Error("error message should not contain bot token")
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestTelegramSender_Send_MissingConfig(t *testing.T) {
	sender := telegram.NewTelegramSender(30, 100*time.Millisecond)

	tests := []struct {
		name   string
		config map[string]string
	}{
		{"missing bot_token", map[string]string{"chat_id": "123"}},
		{"missing chat_id", map[string]string{"bot_token": "abc"}},
		{"empty config", map[string]string{}},
		{"nil config", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel := &alertchannel.AlertChannel{
				ID:     "ch-1",
				Config: tt.config,
			}

			ctx := context.Background()
			_, err := sender.Send(ctx, channel, "test")
			if err == nil {
				t.Error("expected error for missing config")
			}
		})
	}
}

func TestTelegramSender_Send_ContextCancellation(t *testing.T) {
	sender := telegram.NewTelegramSender(30, 100*time.Millisecond)

	channel := &alertchannel.AlertChannel{
		ID: "ch-1",
		Config: map[string]string{
			"bot_token": "test-token",
			"chat_id":   "12345",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := sender.Send(ctx, channel, "test")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
