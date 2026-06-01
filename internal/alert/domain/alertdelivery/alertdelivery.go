package alertdelivery

import (
	"time"
)

// AlertDelivery tracks the delivery status of an alert to a channel.
type AlertDelivery struct {
	ID        string
	TenantID  string
	AlertID   string
	ChannelID string
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
