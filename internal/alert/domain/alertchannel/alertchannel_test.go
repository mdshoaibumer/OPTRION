package alertchannel_test

import (
	"testing"

	"github.com/optrion/optrion/internal/alert/domain/alertchannel"
)

func TestAlertChannel_ChannelTypes(t *testing.T) {
	tests := []struct {
		name     string
		channel  alertchannel.AlertChannel
		wantType alertchannel.ChannelType
	}{
		{
			name: "telegram channel",
			channel: alertchannel.AlertChannel{
				ID:       "ch-1",
				TenantID: "t-1",
				Type:     alertchannel.ChannelTypeTelegram,
				Name:     "Ops Telegram",
				Config:   map[string]string{"bot_token": "abc", "chat_id": "123"},
				Enabled:  true,
			},
			wantType: alertchannel.ChannelTypeTelegram,
		},
		{
			name: "email channel",
			channel: alertchannel.AlertChannel{
				ID:       "ch-2",
				TenantID: "t-1",
				Type:     alertchannel.ChannelTypeEmail,
				Name:     "Ops Email",
				Config:   map[string]string{"to": "team@example.com"},
				Enabled:  true,
			},
			wantType: alertchannel.ChannelTypeEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.channel.Type != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, tt.channel.Type)
			}
			if tt.channel.ID == "" {
				t.Error("channel ID should not be empty")
			}
		})
	}
}

func TestAlertChannel_ConfigAccess(t *testing.T) {
	ch := alertchannel.AlertChannel{
		Config: map[string]string{
			"bot_token": "secret-token",
			"chat_id":   "12345",
		},
	}

	if ch.Config["bot_token"] != "secret-token" {
		t.Errorf("expected bot_token, got %s", ch.Config["bot_token"])
	}
	if ch.Config["chat_id"] != "12345" {
		t.Errorf("expected chat_id, got %s", ch.Config["chat_id"])
	}
	if ch.Config["nonexistent"] != "" {
		t.Error("nonexistent key should return empty string")
	}
}

func TestAlertChannel_Disabled(t *testing.T) {
	ch := alertchannel.AlertChannel{
		Enabled: false,
	}
	if ch.Enabled {
		t.Error("expected channel to be disabled")
	}
}
