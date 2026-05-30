package notificationtemplate

import (
	"time"

	"github.com/google/uuid"
)

// NotificationTemplate defines the format and variables for alert notifications.
type NotificationTemplate struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description string
	Template    string // Markdown with variables
	Variables   []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}
