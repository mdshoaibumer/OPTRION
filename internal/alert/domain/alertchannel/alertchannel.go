package alertchannel

import (
	"time"

	"github.com/google/uuid"
)

// AlertChannel represents a notification channel (e.g., Telegram, Email).
type AlertChannel struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Type      ChannelType
	Name      string
	Config    map[string]string // Channel-specific config (e.g., bot token)
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

type ChannelType string

const (
	ChannelTypeTelegram ChannelType = "telegram"
	ChannelTypeEmail    ChannelType = "email"
	// Add more as needed
)
