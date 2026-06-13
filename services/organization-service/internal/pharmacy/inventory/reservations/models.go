package reservations

import (
	"time"

	"github.com/google/uuid"
)

type ReservationStatus string

const (
	StatusPending   ReservationStatus = "PENDING"
	StatusConfirmed ReservationStatus = "CONFIRMED"
	StatusCancelled ReservationStatus = "CANCELLED"
	StatusExpired   ReservationStatus = "EXPIRED"
)

type Reservation struct {
	ID        uuid.UUID         `json:"id"`
	PharmacyID uuid.UUID        `json:"pharmacy_id"`
	ProductID uuid.UUID         `json:"product_id"`
	BatchID   uuid.UUID         `json:"batch_id"`
	Quantity  int               `json:"quantity"`
	Status    ReservationStatus `json:"status"`
	ExpiresAt time.Time         `json:"expires_at"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type CreateReservationRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	BatchID   uuid.UUID `json:"batch_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1"`
}

type UpdateReservationRequest struct {
	Quantity int `json:"quantity" validate:"required,min=1"`
}
