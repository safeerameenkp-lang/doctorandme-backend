package supplier

import (
	"time"

	"github.com/google/uuid"
)

type BankDetails struct {
	BankName      string `json:"bank_name"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	IFSCCode      string `json:"ifsc_code"` // Or SWIFT
}

type CreditTerms struct {
	CreditPeriodDays int     `json:"credit_period_days"`
	CreditLimit      float64 `json:"credit_limit"`
}

type Supplier struct {
	ID            uuid.UUID   `json:"id"`
	PharmacyID    uuid.UUID   `json:"pharmacy_id"`
	Name          string      `json:"name"`
	SupplierType  string      `json:"supplier_type"`
	ContactPerson string      `json:"contact_person"`
	ContactNumber string      `json:"contact_number"`
	Website       string      `json:"website"`
	Email         string      `json:"email"`
	Address       string      `json:"address"`
	State         string      `json:"state"`
	Pincode       string      `json:"pincode"`
	GSTNumber     string      `json:"gst_number"`
	PANNumber     string      `json:"pan_number"`
	LicenseNumber string      `json:"license_number"`
	BankDetails   BankDetails `json:"bank_details"`
	CreditTerms   CreditTerms `json:"credit_terms"`
	IsActive      bool        `json:"is_active"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	CreatedBy     uuid.UUID   `json:"created_by"`
	UpdatedBy     uuid.UUID   `json:"updated_by"`
}

type CreateSupplierRequest struct {
	Name          string      `json:"name"`
	SupplierType  string      `json:"supplier_type"`
	ContactPerson string      `json:"contact_person"`
	ContactNumber string      `json:"contact_number"`
	Website       string      `json:"website"`
	Email         string      `json:"email"`
	Address       string      `json:"address"`
	State         string      `json:"state"`
	Pincode       string      `json:"pincode"`
	GSTNumber     string      `json:"gst_number"`
	PANNumber     string      `json:"pan_number"`
	LicenseNumber string      `json:"license_number"`
	BankDetails   BankDetails `json:"bank_details"`
	CreditTerms   CreditTerms `json:"credit_terms"`
}

type UpdateSupplierRequest struct {
	Name          *string      `json:"name"`
	SupplierType  *string      `json:"supplier_type"`
	ContactPerson *string      `json:"contact_person"`
	ContactNumber *string      `json:"contact_number"`
	Website       *string      `json:"website"`
	Email         *string      `json:"email"`
	Address       *string      `json:"address"`
	State         *string      `json:"state"`
	Pincode       *string      `json:"pincode"`
	GSTNumber     *string      `json:"gst_number"`
	PANNumber     *string      `json:"pan_number"`
	LicenseNumber *string      `json:"license_number"`
	BankDetails   *BankDetails `json:"bank_details"`
	CreditTerms   *CreditTerms `json:"credit_terms"`
	IsActive      *bool        `json:"is_active"`
}

type SupplierAuditLog struct {
	ID            uuid.UUID `json:"id"`
	PharmacyID    uuid.UUID `json:"pharmacy_id"`
	SupplierID    uuid.UUID `json:"supplier_id"`
	ActionType    string    `json:"action_type"`
	ChangedBy     uuid.UUID `json:"changed_by"`
	ChangedByName string    `json:"changed_by_name"`
	ChangedAt     time.Time `json:"changed_at"`
}

type SupplierStats struct {
	TotalSuppliers  int `json:"total_suppliers"`
	ActiveSuppliers int `json:"active_suppliers"`
	CreditSuppliers int `json:"credit_suppliers"`
	GSTPending      int `json:"gst_pending"`
}

type PaginationMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type APIResponse struct {
	Success bool            `json:"success"`
	Data    interface{}     `json:"data,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
	Error   string          `json:"error,omitempty"`
}
