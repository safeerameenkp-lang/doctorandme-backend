package prescriptions

import (
	"time"

	"github.com/google/uuid"
)

type SalesHistoryStats struct {
	TotalSales     int     `json:"total_sales"`
	TotalAmount    float64 `json:"total_amount"`
	PendingSales   int     `json:"pending_sales"`
	CompletedSales int     `json:"completed_sales"`
}

type Prescription struct {
	ID           string             `json:"id"`
	PharmacyID   uuid.UUID          `json:"pharmacy_id"`
	TokenNo      string             `json:"token_no"`
	PatientName  string             `json:"patient_name"`
	PatientPhone string             `json:"patient_phone"`
	DoctorName   string             `json:"doctor_name"`
	Date         time.Time          `json:"date"`
	Status         string             `json:"status"` // PENDING, DISPENSED, CANCELLED
	Items          []PrescriptionItem `json:"items"`
	TotalMedicines int                `json:"total_medicines"`
	BillAmount     *float64           `json:"bill_amount"`
	PaymentMethod  *string            `json:"payment_method"`
	HandledByName  *string            `json:"handled_by_name"`
	LatestSaleID   *uuid.UUID         `json:"latest_sale_id"`
	InvoiceNumber  *string            `json:"invoice_number"`
}

type PrescriptionItem struct {
	ID             uuid.UUID `json:"id"`
	PrescriptionID string    `json:"prescription_id"`
	ProductID      uuid.UUID `json:"product_id"`
	MedicineName   string    `json:"medicine_name"`
	MedicineBrand  string    `json:"medicine_brand"`
	Quantity       int       `json:"quantity"`
	DurationDays   int       `json:"duration_days"`
	DosagePerDay   float64   `json:"dosage_per_day"`
	Morning        float64   `json:"morning"`
	Noon           float64   `json:"noon"`
	Night          float64   `json:"night"`
	Instructions   string    `json:"instructions"`
}

type CreatePrescriptionRequest struct {
	TokenNo      string                   `json:"token_no"`
	PatientName  string                   `json:"patient_name" validate:"required"`
	PatientPhone string                   `json:"patient_phone"`
	DoctorName   string                   `json:"doctor_name" validate:"required"`
	Items        []CreatePrescriptionItem `json:"items" validate:"required,min=1"`
}

type CreatePrescriptionItem struct {
	ProductID     uuid.UUID `json:"product_id" validate:"required"`
	MedicineName  string    `json:"medicine_name" validate:"required"`
	MedicineBrand string    `json:"medicine_brand"`
	Quantity      int       `json:"quantity"` // Can be manually passed or auto-calculated
	DurationDays int       `json:"duration_days"`
	DosagePerDay float64   `json:"dosage_per_day"`
	Morning      float64   `json:"morning"`
	Noon         float64   `json:"noon"`
	Night        float64   `json:"night"`
	Instructions string    `json:"instructions"`
}
