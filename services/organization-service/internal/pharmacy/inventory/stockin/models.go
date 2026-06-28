package stockin

import (
	"time"

	"github.com/google/uuid"
)

// Purchase represents the header of a stock-in transaction (Invoice Level)
type Purchase struct {
	ID            uuid.UUID `json:"id"`
	PharmacyID    uuid.UUID `json:"pharmacy_id"`
	SupplierID    uuid.UUID `json:"supplier_id"`
	SupplierName  string    `json:"supplier_name"`
	InvoiceNo     string    `json:"invoice_no"`
	PurchaseDate  time.Time `json:"purchase_date"`
	ReceivedBy    string    `json:"received_by"`
	
	GrandTotal    float64   `json:"grand_total"`
	PaidAmount    float64   `json:"paid_amount"`
	DueAmount     float64   `json:"due_amount"`
	PaymentStatus string    `json:"payment_status"`
	
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PurchaseItem represents an individual medicine entry (Batch Level)
type PurchaseItem struct {
	ID                   uuid.UUID `json:"id"`
	PurchaseID           uuid.UUID `json:"purchase_id"`
	PharmacyID           uuid.UUID `json:"pharmacy_id"`
	MedicineID           uuid.UUID `json:"medicine_id"`
	MedicineName         string    `json:"medicine_name"`
	MedicineBrand        string    `json:"medicine_brand"`
	
	// Batch & Expiry
	BatchNo              string    `json:"batch_no"`
	MfgDate              time.Time `json:"mfg_date"`
	ExpiryDate           time.Time `json:"expiry_date"`
	RackNo               string    `json:"rack_no"`
	
	// Quantity Details
	UnitMode             string    `json:"unit_mode"`
	UnitsPerMode         int       `json:"units_per_mode"`
	ReceivedQty          int       `json:"received_qty"`
	BonusQty             int       `json:"bonus_qty"`
	TotalQtyUnits        int       `json:"total_qty_units"`
	BaseUnit             string    `json:"base_unit"`
	
	// Pricing Details (Per Mode)
	PurchasePricePerMode float64   `json:"purchase_price_per_mode"`
	MRPPerMode           float64   `json:"mrp_per_mode"`
	
	// Tax Details
	// Tax Details (Rates Only)
	CGSTRate             float64   `json:"cgst_rate"`
	SGSTRate             float64   `json:"sgst_rate"`
	TotalTaxPercentage   float64   `json:"total_tax_percentage"`

	// Discount/Billing Control Tiers
	RetailDiscountPercentage  float64 `json:"retail_discount_percentage"`
	StaffDiscountPercentage   float64 `json:"staff_discount_percentage"`
	SpecialDiscountPercentage float64 `json:"special_discount_percentage"`
	MaxDiscountPercentage     float64 `json:"max_discount_percentage"`
	
	// Calculated Costs (Per Unit)
	CostPricePerMode     float64   `json:"cost_price_per_mode"`
	CostPricePerUnit     float64   `json:"cost_price_per_unit"`
	ItemTotalAmount      float64   `json:"item_total_amount"`
	
	CreatedAt            time.Time `json:"created_at"`
}

// CreatePurchaseRequest is the payload from the UI
type CreatePurchaseRequest struct {
	SupplierID    uuid.UUID             `json:"supplier_id" validate:"required"`
	InvoiceNo     string                `json:"invoice_no" validate:"required,max=100"`
	PurchaseDate  time.Time             `json:"purchase_date" validate:"required"`
	ReceivedBy    string                `json:"received_by" validate:"required,max=255"`
	GrandTotal    float64               `json:"grand_total" validate:"required,gte=0"`
	PaidAmount    float64               `json:"paid_amount" validate:"gte=0"`
	Notes         string                `json:"notes" validate:"max=500"`
	Items         []CreatePurchaseItem  `json:"items" validate:"required,min=1,dive"`
}

// CreatePurchaseItem includes the item-level specs
type CreatePurchaseItem struct {
	MedicineID           uuid.UUID `json:"medicine_id" validate:"required"`
	BatchNo              string    `json:"batch_no" validate:"required,max=100"`
	MfgDate              time.Time `json:"mfg_date" validate:"required"`
	ExpiryDate           time.Time `json:"expiry_date" validate:"required"`
	RackNo               string    `json:"rack_no" validate:"max=50"`
	
	UnitMode             string    `json:"unit_mode" validate:"required,max=50"`
	UnitsPerMode         int       `json:"units_per_mode" validate:"required,gt=0"`
	ReceivedQty          int       `json:"received_qty" validate:"required,gt=0"`
	BonusQty             int       `json:"bonus_qty" validate:"gte=0"`
	
	PurchasePricePerMode float64   `json:"purchase_price_per_mode" validate:"required,gt=0"`
	MRPPerMode           float64   `json:"mrp_per_mode" validate:"required,gt=0"`
	ItemTotalAmount      float64   `json:"item_total_amount" validate:"required,gte=0"`
	
	// Billing Controls (Entered during stock-in)
	RetailDiscountPercentage  float64 `json:"retail_discount_percentage" validate:"gte=0,lte=100"`
	StaffDiscountPercentage   float64 `json:"staff_discount_percentage" validate:"gte=0,lte=100"`
	SpecialDiscountPercentage float64 `json:"special_discount_percentage" validate:"gte=0,lte=100"`
	MaxDiscountPercentage     float64 `json:"max_discount_percentage" validate:"gte=0,lte=100"`

	// Tax Details (Rates entered or auto-populated during stock-in)
	CGSTRate                  float64 `json:"cgst_rate" validate:"gte=0"`
	SGSTRate                  float64 `json:"sgst_rate" validate:"gte=0"`
	TotalTaxPercentage        float64 `json:"total_tax_percentage" validate:"gte=0"`
}

// UpdateStockInPaymentRequest is the payload to edit only the paid amount
type UpdateStockInPaymentRequest struct {
	PaidAmount float64 `json:"paid_amount" validate:"gte=0"`
}

// StockInStats represents financial summary for stock-in
type StockInStats struct {
	TotalAmount float64 `json:"total_amount"`
	PaidAmount  float64 `json:"paid_amount"`
	DueAmount   float64 `json:"due_amount"`
}

type StockInAuditLog struct {
	ID            uuid.UUID `json:"id"`
	PharmacyID    uuid.UUID `json:"pharmacy_id"`
	StockInID     uuid.UUID `json:"stock_in_id"`
	ActionType    string    `json:"action_type"` // CREATE / UPDATE
	ChangedBy     uuid.UUID `json:"changed_by"`
	ChangedByName string    `json:"changed_by_name"`
	ChangedAt     time.Time `json:"changed_at"`
}
