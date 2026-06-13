package stockouts

import (
	"time"

	"github.com/google/uuid"
)

// StockOutType represents the reason for the inventory deduction
type StockOutType string

const (
	TypeDamaged    StockOutType = "DAMAGED"
	TypeExpired    StockOutType = "EXPIRED"
	TypeTransfer   StockOutType = "TRANSFER"
	TypeAdjustment StockOutType = "ADJUSTMENT"
)

// StockOutStatus represents the current state of the transaction
type StockOutStatus string

const (
	StatusCompleted StockOutStatus = "COMPLETED"
	StatusPending   StockOutStatus = "PENDING"
	StatusCancelled StockOutStatus = "CANCELLED"
)

// StockOut represents the header of a stock-out transaction
type StockOut struct {
	ID             uuid.UUID      `json:"id"`
	PharmacyID     uuid.UUID      `json:"pharmacy_id"`
	Status         StockOutStatus `json:"status"`
	Type           StockOutType   `json:"type"`
	Reason          string         `json:"reason"`
	DestinationType string         `json:"destination_type"`
	DestinationName string         `json:"destination_name"`
	DestinationID   *uuid.UUID     `json:"destination_id"` // Optional ID for validation
	TotalLossValue  float64        `json:"total_loss_value"`
	
	CreatedByID    uuid.UUID      `json:"created_by_id"`
	CreatedByName  string         `json:"created_by_name"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// StockOutItem represents an individual line item in the stock-out
type StockOutItem struct {
	ID            uuid.UUID `json:"id"`
	StockOutID    uuid.UUID `json:"stock_out_id"`
	PharmacyID    uuid.UUID `json:"pharmacy_id"`
	
	MedicineID    uuid.UUID `json:"medicine_id"`
	MedicineName  string    `json:"medicine_name"`
	BatchID       uuid.UUID `json:"batch_id"`
	BatchNo       string    `json:"batch_no"`
	ExpiryDate    time.Time `json:"expiry_date"`
	UnitType      string    `json:"unit_type"`
	
	Quantity      int       `json:"quantity"`
	UnitCostPrice float64   `json:"unit_cost_price"`
	TotalLoss     float64   `json:"total_loss"`
	
	CreatedAt     time.Time `json:"created_at"`
}

// CreateStockOutRequest is the payload from the UI
type CreateStockOutRequest struct {
	Type            StockOutType          `json:"type" validate:"required"`
	Reason          string                `json:"reason" validate:"max=500"`
	DestinationType string                `json:"destination_type" validate:"required"`
	DestinationName string                `json:"destination_name" validate:"required"`
	DestinationID   *uuid.UUID            `json:"destination_id"` // Added for supplier validation
	Items           []CreateStockOutItem `json:"items" validate:"required,min=1,dive"`
}

// CreateStockOutItem represents the detail for a single item deduction
type CreateStockOutItem struct {
	BatchID    uuid.UUID `json:"batch_id" validate:"required"`
	MedicineID uuid.UUID `json:"medicine_id" validate:"required"`
	Quantity   int       `json:"quantity" validate:"required,gt=0"`
	ExpiryDate time.Time `json:"expiry_date" validate:"required"`
	UnitType   string    `json:"unit_type" validate:"required"`
}

// StockOutStats represents the summary metrics for the UI cards
type StockOutStats struct {
	TotalEntries   int     `json:"total_entries"`
	TotalLossValue float64 `json:"total_loss_value"`
	DamagedCount   int     `json:"damaged_count"`
	TransferCount  int     `json:"transfer_count"`
}

// StockOutAuditLog provides a history of changes for accountability
type StockOutAuditLog struct {
	ID            uuid.UUID `json:"id"`
	PharmacyID    uuid.UUID `json:"pharmacy_id"`
	StockOutID    uuid.UUID `json:"stock_out_id"`
	ActionType    string    `json:"action_type"` // CREATE, CANCEL
	ChangedBy     uuid.UUID `json:"changed_by"`
	ChangedByName string    `json:"changed_by_name"`
	ChangedAt     time.Time `json:"changed_at"`
}
