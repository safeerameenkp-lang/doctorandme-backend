package batches

import (
	"time"

	"github.com/google/uuid"
)

// Batch represents the live inventory state of a specific medicine batch
type Batch struct {
	ID            uuid.UUID `json:"id"`
	PharmacyID    uuid.UUID `json:"pharmacy_id"`
	MedicineID    uuid.UUID `json:"medicine_id"`
	MedicineName  string    `json:"medicine_name,omitempty"`
	MedicineBrand string    `json:"medicine_brand,omitempty"`
	BatchNo       string    `json:"batch_no"`
	MfgDate       time.Time `json:"mfg_date"`
	ExpiryDate    time.Time `json:"expiry_date"`
	RackNo        string    `json:"rack_no,omitempty"`

	QuantityAvailable int `json:"quantity_available"`

	CostPrice float64 `json:"cost_price"`
	MRP       float64 `json:"mrp"`
	UnitPrice float64 `json:"unit_price"`

	CGSTRate           float64 `json:"cgst_rate"`
	SGSTRate           float64 `json:"sgst_rate"`
	TotalTaxPercentage float64 `json:"total_tax_percentage"`

	RetailDiscPerc  float64 `json:"retail_disc_perc"`
	StaffDiscPerc   float64 `json:"staff_disc_perc"`
	SpecialDiscPerc float64 `json:"special_disc_perc"`
	MaxDiscPerc     float64 `json:"max_disc_perc"`

	SupplierID   *uuid.UUID `json:"supplier_id,omitempty"`
	SupplierName string     `json:"supplier_name,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// UpdateBatchDTO creates/updates a batch from a Purchase Item
type UpdateBatchDTO struct {
	PharmacyID    uuid.UUID
	MedicineID    uuid.UUID
	BatchNo       string
	MfgDate       time.Time
	ExpiryDate    time.Time
	RackNo        string
	QuantityToAdd int

	CostPrice float64
	MRP       float64
	UnitPrice float64

	CGSTRate           float64
	SGSTRate           float64
	TotalTaxPercentage float64

	RetailDiscPerc  float64
	StaffDiscPerc   float64
	SpecialDiscPerc float64
	MaxDiscPerc     float64

	SupplierID uuid.UUID

	// Added for Ledger Tracking
	TransactionType string
	ReferenceType   string
	ReferenceID     *uuid.UUID
	PerformedBy     *uuid.UUID
	Notes           string
}

type BatchStats struct {
	TotalStocks       int     `json:"total_stocks"`
	TotalStockValue   float64 `json:"total_stock_value"`
	OutOfStock        int     `json:"out_of_stock"`
	ExpiredStock      int     `json:"expired_stock"`
	ExpiringSoon      int     `json:"expiring_soon"`
	ExpiringSoonValue float64 `json:"expiring_soon_value"`
	HighRiskCount     int     `json:"high_risk_count"`
}

type BatchAuditLog struct {
	ID            uuid.UUID `json:"id"`
	PharmacyID    uuid.UUID `json:"pharmacy_id"`
	BatchID       uuid.UUID `json:"batch_id"`
	ActionType    string    `json:"action_type"` // CREATE, UPDATE, STOCK_IN, STOCK_ADJUSTMENT
	ChangedBy     uuid.UUID `json:"changed_by"`
	ChangedByName string    `json:"changed_by_name"`
	Notes         string    `json:"notes"`
	ChangedAt     time.Time `json:"changed_at"`
}

type EditBatchRequest struct {
	RackNo             string  `json:"rack_no"`
	MRP                float64 `json:"mrp"`
	UnitPrice          float64 `json:"unit_price"`
	CGSTRate           float64 `json:"cgst_rate"`
	SGSTRate           float64 `json:"sgst_rate"`
	TotalTaxPercentage float64 `json:"total_tax_percentage"`
	RetailDiscPerc     float64 `json:"retail_disc_perc"`
	StaffDiscPerc      float64 `json:"staff_disc_perc"`
	SpecialDiscPerc    float64 `json:"special_disc_perc"`
	MaxDiscPerc        float64 `json:"max_disc_perc"`
}

type ReturnItemRequest struct {
	BatchID  uuid.UUID `json:"batch_id" validate:"required"`
	Quantity int       `json:"quantity" validate:"required,min=1"`
	Reason   string    `json:"reason"`
}

type BatchReturnRequest struct {
	Items []ReturnItemRequest `json:"items" validate:"required,min=1,dive"`
}
