package models

// CreateClinicInput defines the input structure for creating a clinic
type CreateClinicInput struct {
	OrganizationID string  `json:"organization_id" form:"organization_id" binding:"required,uuid"`
	UserID         string  `json:"user_id" form:"user_id" binding:"required,uuid"`
	ClinicCode     *string `json:"clinic_code" form:"clinic_code" binding:"omitempty,min=2,max=20"` // Optional, auto-generated if empty
	Name           string  `json:"name" form:"name" binding:"required,min=2,max=255"`
	ClinicType     string  `json:"clinic_type" form:"clinic_type" binding:"required,max=50"`
	Email          *string `json:"email" form:"email" binding:"omitempty,email"`
	Phone          *string `json:"phone" form:"phone" binding:"omitempty,len=10"`
	Address        *string `json:"address" form:"address" binding:"omitempty,max=500"`
	LicenseNumber  *string `json:"license_number" form:"license_number" binding:"omitempty,max=100"`
}

// CreateClinicWithAdminInput defines the input structure for creating a clinic with a new admin
type CreateClinicWithAdminInput struct {
	OrganizationID string  `json:"organization_id" form:"organization_id" binding:"required,uuid"`
	ClinicCode     *string `json:"clinic_code" form:"clinic_code" binding:"omitempty,min=2,max=20"` // Optional, auto-generated if empty
	Name           string  `json:"name" form:"name" binding:"required,min=2,max=255"`
	ClinicType     string  `json:"clinic_type" form:"clinic_type" binding:"required,max=50"`
	Email          *string `json:"email" form:"email" binding:"omitempty,email"`
	Phone          *string `json:"phone" form:"phone" binding:"omitempty,len=10"`
	Address        *string `json:"address" form:"address" binding:"omitempty,max=500"`
	LicenseNumber  *string `json:"license_number" form:"license_number" binding:"omitempty,max=100"`
	// Admin details
	AdminFirstName string `json:"admin_first_name" form:"admin_first_name" binding:"max=50"`
	AdminLastName  string `json:"admin_last_name" form:"admin_last_name" binding:"max=50"`
	AdminEmail     string `json:"admin_email" form:"admin_email" binding:"required,email"`
	AdminUsername  string `json:"admin_username" form:"admin_username" binding:"required,min=3,max=30"`
	AdminPhone     string `json:"admin_phone" form:"admin_phone" binding:"omitempty,len=10"`
	AdminPassword  string `json:"admin_password" form:"admin_password" binding:"required,min=8"`
}

// UpdateClinicInput defines fields for updating a clinic
type UpdateClinicInput struct {
	ClinicCode    *string `json:"clinic_code" form:"clinic_code" binding:"omitempty,min=2,max=20"`
	Name          *string `json:"name" form:"name" binding:"omitempty,min=2,max=255"`
	ClinicType    *string `json:"clinic_type" form:"clinic_type" binding:"omitempty,max=50"`
	Email         *string `json:"email" form:"email" binding:"omitempty,email"`
	Phone         *string `json:"phone" form:"phone" binding:"omitempty,len=10"`
	Address       *string `json:"address" form:"address" binding:"omitempty,max=500"`
	LicenseNumber *string `json:"license_number" form:"license_number" binding:"omitempty,max=100"`
	IsActive      *bool   `json:"is_active" form:"is_active"`
	Logo          *string // handled separately or not at all for now
}
