package sales

import (
	"organization-service/internal/pharmacy/sales/clients"
	"time"

	"github.com/google/uuid"
)

type SaleStatus string

const (
	StatusDraft      SaleStatus = "DRAFT"
	StatusPending    SaleStatus = "PENDING"
	StatusCompleted  SaleStatus = "COMPLETED"
	StatusCancelled  SaleStatus = "CANCELLED"
	StatusDispatched SaleStatus = "DISPATCHED"
)

type SaleType string

const (
	TypeInternalRx SaleType = "INTERNAL_RX"
	TypeWalkIn     SaleType = "WALK_IN"
)

type Patient struct {
	ID           uuid.UUID `json:"id"`
	PharmacyID   uuid.UUID `json:"pharmacy_id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Gender       string    `json:"gender"`
	Age          int       `json:"age"`
	Address      string    `json:"address,omitempty"`
	IsRecurring  bool      `json:"is_recurring"`
	DueAmount    float64   `json:"due_amount"`
	CreditAmount float64   `json:"credit_amount"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Sale struct {
	ID                  uuid.UUID                 `json:"id"`
	PharmacyID          uuid.UUID                 `json:"pharmacy_id"`
	SaleType            SaleType                  `json:"sale_type"`
	PrescriptionID      string                    `json:"prescription_id"`
	PatientID           *uuid.UUID                `json:"patient_id,omitempty"`
	CustomerName        string                    `json:"customer_name,omitempty"`
	CustomerPhone       string                    `json:"customer_phone,omitempty"`
	CustomerAge         int                       `json:"customer_age,omitempty"`
	CustomerGender      string                    `json:"customer_gender,omitempty"`
	CustomerAddress     string                    `json:"customer_address,omitempty"`
	Status              SaleStatus                `json:"status"`
	GrossAmount         float64                   `json:"gross_amount"`
	TotalAmount         float64                   `json:"total_amount"`
	TotalDiscount       float64                   `json:"total_discount"`
	TotalTax            float64                   `json:"total_tax"`
	IsRecurring         bool                      `json:"is_recurring"`
	DaysSupply          int                       `json:"days_supply"`
	NextRefillDate      *time.Time                `json:"next_refill_date,omitempty"`
	AppliedCredit       float64                   `json:"applied_credit"`
	AppliedDue          float64                   `json:"applied_due"`
	GeneratedCredit     float64                   `json:"generated_credit"`
	GeneratedDue        float64                   `json:"generated_due"`
	InvoiceNumber       string                    `json:"invoice_number,omitempty"`
	CreatedAt           time.Time                 `json:"created_at"`
	UpdatedAt           time.Time                 `json:"updated_at"`
	Prescription        *clients.PrescriptionData `json:"prescription,omitempty"`
	Items               []SaleItem                `json:"items,omitempty"`
	PaymentMode         string                    `json:"payment_mode,omitempty"`
	HandledBy           string                    `json:"handled_by,omitempty"`
	PatientDueAmount    float64                   `json:"patient_due_amount"`
	PatientCreditAmount float64                   `json:"patient_credit_amount"`
	CollectedAmount     float64                   `json:"collected_amount"`
}

type SaleItem struct {
	ID                 uuid.UUID `json:"id"`
	SaleID             uuid.UUID `json:"sale_id"`
	ProductID          uuid.UUID `json:"product_id"`
	MedicineName       string    `json:"medicine_name"`
	MedicineBrand      string    `json:"medicine_brand"`
	BatchID            uuid.UUID `json:"batch_id"`
	BatchNo            string    `json:"batch_no,omitempty"`
	Quantity           int       `json:"quantity"`
	AvailableStock     int       `json:"available_stock,omitempty"`
	ExpiryDate         time.Time `json:"expiry_date,omitempty"`
	MRP                float64   `json:"mrp"`
	Price              float64   `json:"unit_price"` // This is the Unit Price from Batch
	DiscountPercentage float64   `json:"discount_percentage"`
	TaxPercentage      float64   `json:"tax_percentage"`
	Subtotal           float64   `json:"subtotal"`
	RetailDiscPerc     float64   `json:"retail_disc_perc"`
	StaffDiscPerc      float64   `json:"staff_disc_perc"`
	SpecialDiscPerc    float64   `json:"special_disc_perc"`
	MaxDiscPerc        float64   `json:"max_disc_perc"`
	ReservationID      string    `json:"reservation_id"`
	RackNo             string    `json:"rack_no"`
	ReturnedQuantity   int       `json:"returned_quantity"`
	CreatedAt          time.Time `json:"created_at"`
}

type PaymentMode string

const (
	PayModeCash   PaymentMode = "CASH"
	PayModeUPI    PaymentMode = "UPI"
	PayModeCard   PaymentMode = "CARD"
	PayModeCredit PaymentMode = "CREDIT"
)

type TransactionType string

const (
	TxTypePayment TransactionType = "PAYMENT"
	TxTypeRefund  TransactionType = "REFUND"
)

type Payment struct {
	ID              uuid.UUID       `json:"id"`
	SaleID          uuid.UUID       `json:"sale_id"`
	ReturnID        *uuid.UUID      `json:"return_id,omitempty"`
	TransactionType TransactionType `json:"transaction_type"`
	Mode            PaymentMode     `json:"mode"`
	Amount          float64         `json:"amount"`
	CreatedAt       time.Time       `json:"created_at"`
}

type SaleReturnStatus string

const (
	ReturnStatusPending   SaleReturnStatus = "PENDING"
	ReturnStatusCompleted SaleReturnStatus = "COMPLETED"
	ReturnStatusCancelled SaleReturnStatus = "CANCELLED"
)

type SaleReturn struct {
	ID            uuid.UUID        `json:"id"`
	PharmacyID    uuid.UUID        `json:"pharmacy_id"`
	SaleID        uuid.UUID        `json:"sale_id"`
	InvoiceNumber string           `json:"invoice_number"` // Original Invoice No
	ReturnNumber  string           `json:"return_number"`  // Unique Return No
	Status        SaleReturnStatus `json:"status"`
	TotalRefund   float64          `json:"total_refund"`
	Reason        string           `json:"reason,omitempty"`
	HandledBy     string           `json:"handled_by,omitempty"`
	RefundMode    PaymentMode      `json:"refund_mode"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	Items         []SaleReturnItem `json:"items,omitempty"`
}

type SaleReturnItem struct {
	ID           uuid.UUID `json:"id"`
	ReturnID     uuid.UUID `json:"return_id"`
	SaleItemID   uuid.UUID `json:"sale_item_id"` // Link to original sale item
	ProductID    uuid.UUID `json:"product_id"`
	MedicineName string    `json:"medicine_name"`
	BatchID      uuid.UUID `json:"batch_id"`
	BatchNo      string    `json:"batch_no"`
	Quantity     int       `json:"quantity"` // Quantity being returned
	RefundAmount float64   `json:"refund_amount"`
	Condition    string    `json:"condition"` // e.g., "SELLABLE", "DAMAGED"
	CreatedAt    time.Time `json:"created_at"`
}

// Request/Response Structs

type CreateReturnRequest struct {
	SaleID     uuid.UUID          `json:"sale_id" validate:"required"`
	Reason     string             `json:"reason" validate:"required"`
	RefundMode PaymentMode        `json:"refund_mode" validate:"required,oneof=CASH UPI CARD CREDIT"`
	Items      []CreateReturnItem `json:"items" validate:"required,min=1,dive"`
}

type CreateReturnItem struct {
	SaleItemID uuid.UUID `json:"sale_item_id" validate:"required"`
	Quantity   int       `json:"quantity" validate:"required,min=1"`
	Condition  string    `json:"condition" validate:"required,oneof=SELLABLE DAMAGED"`
}

type CreateDraftRequest struct {
	SaleType       SaleType `json:"sale_type" validate:"required,oneof=INTERNAL_RX"`
	PrescriptionID string   `json:"prescription_id" validate:"required"`
}

type AddItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1"`
}

type UpdateItemRequest struct {
	Quantity           int      `json:"quantity" validate:"required,min=1"`
	DiscountPercentage *float64 `json:"discount_percentage,omitempty"`
}

type FinalizeSaleRequest struct {
	PaymentMode  PaymentMode `json:"payment_mode" validate:"required,oneof=CASH UPI CARD"`
	AmountPaid   float64     `json:"amount_paid" validate:"required,min=0"`
	IsRecurring  bool        `json:"is_recurring"`
	DaysSupply   int         `json:"days_supply"`
	WalletAction string      `json:"wallet_action,omitempty"`
	WalletAmount float64     `json:"wallet_amount,omitempty"`
}

type SalesStats struct {
	DailySales        float64   `json:"daily_sales"`
	SalesVolume       int       `json:"sales_volume"`
	NewPatients       int       `json:"new_patients"`
	TotalPatients     int       `json:"total_patients"`
	RecurringPatients int       `json:"recurring_patients"`
	RecurringSales    float64   `json:"recurring_sales"`
	TrendAmounts      []float64 `json:"trend_amounts"`
	TrendDates        []string  `json:"trend_dates"`
}
type PatientStats struct {
	TotalPatients     int `json:"total_patients"`
	NewPatientsToday  int `json:"new_patients_today"`
	RecurringPatients int `json:"recurring_patients"`
}

type PatientPurchase struct {
	ID            uuid.UUID `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	TotalAmount   float64   `json:"total_amount"`
	ItemCount     int       `json:"item_count"`
	BilledDate    time.Time `json:"billed_date"`
	DoneBy        string    `json:"done_by"`
}

type RecurringRefillReportItem struct {
	SaleID         uuid.UUID  `json:"sale_id"`
	PatientName    string     `json:"patient_name"`
	InvoiceNumber  string     `json:"invoice_number"`
	LastRefillDate time.Time  `json:"last_refill_date"`
	DaysSupply     int        `json:"days_supply"`
	NextRefillDate *time.Time `json:"next_refill_date"`
}

