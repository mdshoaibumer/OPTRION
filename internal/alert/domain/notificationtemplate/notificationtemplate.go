package notificationtemplate

import (
	"time"
)

// NotificationTemplate defines the format and variables for alert notifications.
type NotificationTemplate struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	Template    string // Markdown with variables
	Variables   []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   string
	UpdatedBy   string
}
