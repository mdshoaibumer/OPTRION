package alertdelivery

import (
	"time"

	"github.com/google/uuid"
)

// AlertDelivery tracks the delivery status of an alert to a channel.
type AlertDelivery struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	AlertID   uuid.UUID
	ChannelID uuid.UUID
	Status    DeliveryStatus
	Attempts  int
	LastError string
	History   []DeliveryHistory
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusRetrying  DeliveryStatus = "retrying"
)

type DeliveryHistory struct {
	Timestamp time.Time
	Status    DeliveryStatus
	Error     string
}
