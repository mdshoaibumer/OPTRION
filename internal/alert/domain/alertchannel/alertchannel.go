package alertchannel

import (
	"time"
)

// AlertChannel represents a notification channel (e.g., Telegram, Email).
type AlertChannel struct {
	ID        string
	TenantID  string
	Type      ChannelType
	Name      string
	Config    map[string]string // Channel-specific config (e.g., bot token)
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy string
	UpdatedBy string
}

type ChannelType string

const (
	ChannelTypeTelegram ChannelType = "telegram"
	ChannelTypeEmail    ChannelType = "email"
	// Add more as needed
)
