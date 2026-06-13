package notification

import "time"

type NotificationType string

const (
	NotificationSMS  NotificationType = "SMS"
	NotificationPush NotificationType = "PUSH"
)

type Notification struct {
	ID         string           `json:"id"`
	PharmacyID string           `json:"pharmacy_id"`
	UserID     string           `json:"user_id"`
	Type       NotificationType `json:"type"`
	Message    string           `json:"message"`
	SentAt     time.Time        `json:"sent_at"`
}
