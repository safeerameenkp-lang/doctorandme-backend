package patient

// PatientResponse represents the API output format for a patient
type PatientResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	MOID           string `json:"mo_id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	DateOfBirth    string `json:"date_of_birth"`
	Gender         string `json:"gender"`
	MedicalHistory string `json:"medical_history"`
	Allergies      string `json:"allergies"`
	BloodGroup     string `json:"blood_group"`
	IsActive       bool   `json:"is_active"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	ClinicName     string `json:"clinic_name,omitempty"` // Included in list endpoint
}

// CreatePatientInput represents the incoming payload
type CreatePatientInput struct {
	// User information
	FirstName   string  `json:"first_name" binding:"required,min=2,max=50"`
	LastName    string  `json:"last_name" binding:"required,min=2,max=50"`
	Phone       string  `json:"phone" binding:"required,min=10,max=15"`
	Email       *string `json:"email" binding:"omitempty,email"`
	DateOfBirth *string `json:"date_of_birth" binding:"omitempty"`
	Gender      *string `json:"gender" binding:"omitempty,oneof=male female other"`

	// Patient-specific information
	MOID           *string `json:"mo_id" binding:"omitempty,min=3,max=20"`
	MedicalHistory *string `json:"medical_history"`
	Allergies      *string `json:"allergies"`
	BloodGroup     *string `json:"blood_group" binding:"omitempty,oneof=A+ A- B+ B- AB+ AB- O+ O-"`

	// Clinic assignment (optional for Super Admin, required for Clinic Admin)
	ClinicID *string `json:"clinic_id" binding:"omitempty,uuid"`
}

// UpdatePatientInput represents the payload for updating a patient
type UpdatePatientInput struct {
	FirstName      *string `json:"first_name" binding:"omitempty,min=2,max=50"`
	LastName       *string `json:"last_name" binding:"omitempty,min=2,max=50"`
	Phone          *string `json:"phone" binding:"omitempty,min=10,max=15"`
	Email          *string `json:"email" binding:"omitempty,email"`
	DateOfBirth    *string `json:"date_of_birth" binding:"omitempty"`
	Gender         *string `json:"gender" binding:"omitempty,oneof=male female other"`
	MOID           *string `json:"mo_id" binding:"omitempty,min=3,max=20"`
	MedicalHistory *string `json:"medical_history"`
	Allergies      *string `json:"allergies"`
	BloodGroup     *string `json:"blood_group" binding:"omitempty,oneof=A+ A- B+ B- AB+ AB- O+ O-"`
	IsActive       *bool   `json:"is_active"`
}

// AssignClinicInput represents the assignment payload
type AssignClinicInput struct {
	ClinicID string `json:"clinic_id" binding:"required,uuid"`
}
