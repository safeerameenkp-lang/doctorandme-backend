package ledger

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	Purchase       TransactionType = "PURCHASE"
	Sale           TransactionType = "SALE"
	SaleReturn     TransactionType = "SALE_RETURN"
	PurchaseReturn TransactionType = "PURCHASE_RETURN"
	Adjustment     TransactionType = "ADJUSTMENT"
)

// StockLedger represents an entry in the audit log
type StockLedger struct {
	ID              uuid.UUID       `json:"id"`
	PharmacyID      uuid.UUID       `json:"pharmacy_id"`
	MedicineID      uuid.UUID       `json:"medicine_id"`
	BatchID         uuid.UUID       `json:"batch_id"`
	TransactionType TransactionType `json:"transaction_type"`
	QuantityChange  int             `json:"quantity_change"`
	BalanceAfter    int             `json:"balance_after"`
	ReferenceType   *string         `json:"reference_type,omitempty"`
	ReferenceID     *uuid.UUID      `json:"reference_id,omitempty"`
	PerformedBy     *uuid.UUID      `json:"performed_by,omitempty"`
	Notes           *string         `json:"notes,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// RecordMovementDTO is the input for recording a stock change
type RecordMovementDTO struct {
	PharmacyID      uuid.UUID
	MedicineID      uuid.UUID
	BatchID         uuid.UUID
	TransactionType TransactionType
	QuantityChange  int
	ReferenceType   *string
	ReferenceID     *uuid.UUID
	PerformedBy     *uuid.UUID
	Notes           *string
}

type Repository interface {
	RecordMovement(ctx context.Context, dto RecordMovementDTO) (int, error)
	GetByBatch(ctx context.Context, pharmacyID, batchID uuid.UUID) ([]StockLedger, error)
	GetByMedicine(ctx context.Context, pharmacyID, medicineID uuid.UUID) ([]StockLedger, error)
}

type Service interface {
	Record(ctx context.Context, dto RecordMovementDTO) (int, error)
}
