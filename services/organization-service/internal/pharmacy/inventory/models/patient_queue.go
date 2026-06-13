package models

import "time"

type QueueStatus string

const (
	QueuePending    QueueStatus = "PENDING"
	QueueInProgress QueueStatus = "IN_PROGRESS"
	QueueReady      QueueStatus = "READY"
	QueueCompleted  QueueStatus = "COMPLETED"
)

type PatientQueue struct {
	ID             string      `json:"id"`
	PharmacyID     string      `json:"pharmacy_id"`
	PrescriptionID string      `json:"prescription_id"`
	Status         QueueStatus `json:"status"`
	CreatedAt      time.Time   `json:"created_at"`
}
