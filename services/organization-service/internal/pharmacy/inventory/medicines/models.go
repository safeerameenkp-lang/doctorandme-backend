package medicines

import (
	"time"

	"github.com/google/uuid"
)

type Medicine struct {
	ID                uuid.UUID `json:"id"`
	PharmacyID        uuid.UUID `json:"pharmacy_id"`
	CreatedBy         uuid.UUID `json:"created_by"`
	CreatedByName     string    `json:"created_by_name"`
	UpdatedBy         uuid.UUID `json:"updated_by"`
	UpdatedByName     string    `json:"updated_by_name"`
	Name              string    `json:"name"`
	BrandName         string    `json:"brand_name"`
	DosageForm        string    `json:"dosage_form"`
	Category          string    `json:"category"`
	Manufacturer      string    `json:"manufacturer"`
	MfgLicense        string    `json:"mfg_license"`
	SupplierID        uuid.UUID `json:"supplier_id"` // Non-pointer since it's NOT NULL
	HSNCode           string    `json:"hsn_code"`
	ScheduleType      string    `json:"schedule_type"`
	IsRxRequired      bool      `json:"is_rx_required"`
	UnitType          string    `json:"unit_type"`
	Barcode           string    `json:"barcode"`
	StorageCondition  string    `json:"storage_condition"`
	CGSTRate          float64   `json:"cgst_rate"`
	SGSTRate          float64   `json:"sgst_rate"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CreateMedicineRequest struct {
	Name              string  `json:"name" validate:"required,min=2,max=255"`
	BrandName         string  `json:"brand_name" validate:"max=255"`
	DosageForm        string  `json:"dosage_form" validate:"required,max=100"`
	Category          string  `json:"category" validate:"required,max=100"`
	Manufacturer      string  `json:"manufacturer" validate:"required,max=255"`
	MfgLicense        string  `json:"mfg_license" validate:"max=255"`
	SupplierID        uuid.UUID `json:"supplier_id" validate:"required"`
	HSNCode           string  `json:"hsn_code" validate:"required,max=20"`
	ScheduleType      string  `json:"schedule_type" validate:"required,max=20"`
	IsRxRequired      bool    `json:"is_rx_required"`
	UnitType          string  `json:"unit_type" validate:"required,max=50"`
	Barcode           string  `json:"barcode" validate:"max=100"`
	StorageCondition  string  `json:"storage_condition" validate:"max=100"`
	CGSTRate          float64 `json:"cgst_rate" validate:"gte=0"`
	SGSTRate          float64 `json:"sgst_rate" validate:"gte=0"`
}

type UpdateMedicineRequest struct {
	Name              *string    `json:"name" validate:"omitempty,min=2,max=255"`
	BrandName         *string    `json:"brand_name" validate:"omitempty,max=255"`
	DosageForm        *string    `json:"dosage_form" validate:"omitempty,max=100"`
	Category          *string    `json:"category" validate:"omitempty,max=100"`
	Manufacturer      *string    `json:"manufacturer" validate:"omitempty,max=255"`
	MfgLicense        *string    `json:"mfg_license" validate:"omitempty,max=255"`
	SupplierID        *uuid.UUID `json:"supplier_id"`
	HSNCode           *string    `json:"hsn_code" validate:"omitempty,max=20"`
	ScheduleType      *string    `json:"schedule_type" validate:"omitempty,max=20"`
	IsRxRequired      *bool      `json:"is_rx_required"`
	UnitType          *string    `json:"unit_type" validate:"omitempty,max=50"`
	Barcode           *string    `json:"barcode" validate:"omitempty,max=100"`
	StorageCondition  *string    `json:"storage_condition" validate:"omitempty,max=100"`
	CGSTRate          *float64   `json:"cgst_rate" validate:"omitempty,gte=0"`
	SGSTRate          *float64   `json:"sgst_rate" validate:"omitempty,gte=0"`
	IsActive          *bool      `json:"is_active"`
}

type MedicineStats struct {
	TotalMedicines      int `json:"total_medicines"`
	ActiveMedicines     int `json:"active_medicines"`
	RestrictedMedicines int `json:"restricted_medicines"`
	ZeroGstMedicines    int `json:"zero_gst_medicines"`
}

type MedicineAuditLog struct {
	ID          uuid.UUID `json:"id"`
	PharmacyID  uuid.UUID `json:"pharmacy_id"`
	MedicineID  uuid.UUID `json:"medicine_id"`
	ActionType  string    `json:"action_type"` // CREATE / UPDATE / ACTIVATE / DEACTIVATE
	ChangedBy     uuid.UUID `json:"changed_by"`
	ChangedByName string    `json:"changed_by_name"`
	ChangedAt     time.Time `json:"changed_at"`
}

