package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"organization-service/config"
	"organization-service/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"organization-service/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// =====================================================
// CLINIC-SPECIFIC PATIENT MANAGEMENT
// Patients are isolated per clinic (no global users)
// =====================================================

// Helper to securely extract clinic ID entirely from the JWT Context mapping, strictly enforcing multi-tenant isolation.
func extractClinicIDFromContext(c *gin.Context) string {
	if clinicID := c.GetString("clinic_id"); clinicID != "" {
		return clinicID
	}
	// Fallback for staff/admin tokens where roles are loaded
	if clinicIDs := c.GetStringSlice("clinic_ids"); len(clinicIDs) > 0 {
		return clinicIDs[0]
	}
	return ""
}

// Local helper to parse integers from query strings with default values
func getQueryInt(c *gin.Context, key string, defaultValue int) int {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

// Helper functions to safely convert pointers to values

// CreateClinicPatientInput for clinic-specific patient creation
// Only name and phone required, all other fields optional
type CreateClinicPatientInput struct {
	ClinicID       string  `json:"clinic_id"` // Optional: use if not in JWT (for super_admins)
	FirstName      string  `json:"first_name" binding:"required,max=100"`
	LastName       *string `json:"last_name" binding:"omitempty,max=100"`
	Phone          string  `json:"phone" binding:"required,max=20"`
	Email          *string `json:"email" binding:"omitempty,email"`
	DateOfBirth    *string `json:"date_of_birth"`
	Age            *int    `json:"age" binding:"omitempty,min=0,max=150"`
	Gender         *string `json:"gender" binding:"omitempty,max=20"`
	Address1       *string `json:"address1" binding:"omitempty,max=200"`
	Address2       *string `json:"address2" binding:"omitempty,max=200"`
	District       *string `json:"district" binding:"omitempty,max=100"`
	State          *string `json:"state" binding:"omitempty,max=100"`
	MOID           *string `json:"mo_id" binding:"omitempty,max=50"`
	MedicalHistory *string `json:"medical_history"`
	Allergies      *string `json:"allergies"`
	BloodGroup     *string `json:"blood_group" binding:"omitempty,max=10"`
	SmokingStatus  *string `json:"smoking_status" binding:"omitempty,max=20"`
	AlcoholUse     *string `json:"alcohol_use" binding:"omitempty,max=20"`
	HeightCm       *int    `json:"height_cm" binding:"omitempty,min=0,max=300"`
	WeightKg       *int    `json:"weight_kg" binding:"omitempty,min=0,max=500"`
}

// UpdateClinicPatientInput for updating clinic patient
type UpdateClinicPatientInput struct {
	FirstName      *string `json:"first_name" binding:"omitempty,max=100"`
	LastName       *string `json:"last_name" binding:"omitempty,max=100"`
	Phone          *string `json:"phone" binding:"omitempty,max=20"`
	Email          *string `json:"email" binding:"omitempty,email"`
	DateOfBirth    *string `json:"date_of_birth"`
	Age            *int    `json:"age" binding:"omitempty,min=0,max=150"`
	Gender         *string `json:"gender" binding:"omitempty,max=20"`
	Address1       *string `json:"address1" binding:"omitempty,max=200"`
	Address2       *string `json:"address2" binding:"omitempty,max=200"`
	District       *string `json:"district" binding:"omitempty,max=100"`
	State          *string `json:"state" binding:"omitempty,max=100"`
	MOID           *string `json:"mo_id" binding:"omitempty,max=50"`
	MedicalHistory *string `json:"medical_history"`
	Allergies      *string `json:"allergies"`
	BloodGroup     *string `json:"blood_group" binding:"omitempty,max=10"`
	SmokingStatus  *string `json:"smoking_status" binding:"omitempty,max=20"`
	AlcoholUse     *string `json:"alcohol_use" binding:"omitempty,max=20"`
	HeightCm       *int    `json:"height_cm" binding:"omitempty,min=0,max=300"`
	WeightKg       *int    `json:"weight_kg" binding:"omitempty,min=0,max=500"`
	IsActive       *bool   `json:"is_active"`
}

// Helper functions to safely convert pointers to values
func ptrToStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func ptrToInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func ptrToBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

// 📌 AppointmentHistoryItem - Compact history item for list views
type AppointmentHistoryItem struct {
	ID                  string    `json:"appointment_id"`
	DoctorID            string    `json:"doctor_id"`
	DoctorName          string    `json:"doctor_name"`
	DepartmentID        string    `json:"department_id"`
	Department          string    `json:"department"`
	AppointmentType     string    `json:"appointment_type"` // clinic_visit, video_consultation
	AppointmentDate     string    `json:"appointment_date"`
	DaysSince           int       `json:"days_since"`
	ValidityDays        int       `json:"validity_days"`        // Always 5
	RemainingDays       int       `json:"remaining_days"`       // Days left for free follow-up
	Status              string    `json:"status"`               // active, expired, future
	FollowUpEligible    bool      `json:"follow_up_eligible"`   // Can book follow-up?
	FollowUpStatus      string    `json:"follow_up_status"`     // active, expired, used
	RenewalStatus       string    `json:"renewal_status"`       // valid, waiting, renewed
	FreeFollowUpUsed    bool      `json:"free_follow_up_used"`  // Already used?
	NextFollowUpExpiry  string    `json:"next_followup_expiry"` // When follow-up expires
	Note                string    `json:"note"`                 // Human-readable status note
	appointmentDateTime time.Time `json:"-"`                    // Internal use only (not exported)
}

// ✅ EligibleFollowUp - Structure for active follow-up eligibility
type EligibleFollowUp struct {
	AppointmentID      string `json:"appointment_id"`
	DoctorID           string `json:"doctor_id"`
	DoctorName         string `json:"doctor_name"`
	DepartmentID       string `json:"department_id"`
	Department         string `json:"department"`
	AppointmentDate    string `json:"appointment_date"`
	RemainingDays      int    `json:"remaining_days"`
	NextFollowUpExpiry string `json:"next_followup_expiry"` // When eligibility expires
	Note               string `json:"note"`                 // Human-readable note
}

// ✅ ExpiredFollowUp - Expired follow-ups that need renewal
type ExpiredFollowUp struct {
	AppointmentID string `json:"appointment_id"`
	DoctorID      string `json:"doctor_id"`
	DoctorName    string `json:"doctor_name"`
	DepartmentID  string `json:"department_id"`
	Department    string `json:"department"`
	ExpiredOn     string `json:"expired_on"` // When it expired (5 days after appointment)
	Note          string `json:"note"`       // Human-readable message
}

// LastAppointmentInfo represents the patient's last appointment details
type LastAppointmentInfo struct {
	ID           string `json:"id"`
	DoctorID     string `json:"doctor_id"`
	DoctorName   string `json:"doctor_name"`
	DepartmentID string `json:"department_id"`
	Department   string `json:"department"`
	Date         string `json:"date"`
	Status       string `json:"status"`
	DaysSince    int    `json:"days_since"`
}

// FollowUpEligibility represents follow-up appointment eligibility
type FollowUpEligibility struct {
	Eligible      bool   `json:"eligible"`
	IsFree        bool   `json:"is_free"`
	Reason        string `json:"reason"`
	DaysRemaining int    `json:"days_remaining"`
	Message       string `json:"message"`
}

// ClinicPatientResponse represents clinic patient data
type ClinicPatientResponse struct {
	ID              string    `json:"id"`
	ClinicID        string    `json:"clinic_id"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Phone           string    `json:"phone"`
	Email           string    `json:"email"`
	DateOfBirth     string    `json:"date_of_birth"`
	Age             int       `json:"age"`
	Gender          string    `json:"gender"`
	Address1        string    `json:"address1"`
	Address2        string    `json:"address2"`
	District        string    `json:"district"`
	State           string    `json:"state"`
	MOID            string    `json:"mo_id"`
	MedicalHistory  string    `json:"medical_history"`
	Allergies       string    `json:"allergies"`
	BloodGroup      string    `json:"blood_group"`
	SmokingStatus   string    `json:"smoking_status"`
	AlcoholUse      string    `json:"alcohol_use"`
	HeightCm        int       `json:"height_cm"`
	WeightKg        int       `json:"weight_kg"`
	IsActive        bool      `json:"is_active"`
	GlobalPatientID string    `json:"global_patient_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// ✅ NEW: Status fields from clinic_patients table
	CurrentFollowupStatus string `json:"current_followup_status"`
	LastAppointmentID     string `json:"last_appointment_id"`
	LastFollowupID        string `json:"last_followup_id"`

	// Appointments array (full details)
	Appointments []AppointmentDetail `json:"appointments,omitempty"`

	// Follow-ups array (full details)
	FollowUps []FollowUpDetail `json:"follow_ups,omitempty"`

	// Legacy fields (kept for backward compatibility)
	LastAppointment     *LastAppointmentInfo     `json:"last_appointment,omitempty"`
	FollowUpEligibility *FollowUpEligibility     `json:"follow_up_eligibility,omitempty"`
	TotalAppointments   int                      `json:"total_appointments"`
	AppointmentHistory  []AppointmentHistoryItem `json:"appointment_history,omitempty"`
	EligibleFollowUps   []EligibleFollowUp       `json:"eligible_follow_ups,omitempty"`
	ExpiredFollowUps    []ExpiredFollowUp        `json:"expired_followups,omitempty"`
}

// AppointmentDetail - Full appointment details
type AppointmentDetail struct {
	AppointmentID    string  `json:"appointment_id"`
	DoctorID         string  `json:"doctor_id"`
	DepartmentID     string  `json:"department_id"`
	AppointmentTime  string  `json:"appointment_time"`
	SlotType         string  `json:"slot_type"`
	ConsultationType string  `json:"consultation_type"`
	Status           string  `json:"status"`
	FeeAmount        float64 `json:"fee_amount"`
	PaymentStatus    string  `json:"payment_status"`
	PaymentMode      string  `json:"payment_mode"`
	IsPriority       bool    `json:"is_priority"`
	CreatedAt        string  `json:"created_at"`
}

// FollowUpDetail - Full follow-up details
type FollowUpDetail struct {
	FollowUpID             string `json:"follow_up_id"`
	SourceAppointmentID    string `json:"source_appointment_id"`
	DoctorID               string `json:"doctor_id"`
	DepartmentID           string `json:"department_id"`
	Status                 string `json:"status"`
	IsFree                 bool   `json:"is_free"`
	ValidFrom              string `json:"valid_from"`
	ValidUntil             string `json:"valid_until"`
	UsedAppointmentID      string `json:"used_appointment_id"`
	RenewedByAppointmentID string `json:"renewed_by_appointment_id"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

// CreateClinicPatient - Create clinic-specific patient (no global user)
// POST /clinic-specific-patients
func CreateClinicPatient(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var input CreateClinicPatientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Priority: 1. JWT context, 2. Request Body (useful for Super Admins or missing tenant context)
	clinicID := extractClinicIDFromContext(c)
	if clinicID == "" {
		clinicID = input.ClinicID
	}

	if clinicID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "Valid clinic context is required to register a patient",
		})
		return
	}

	var clinicExists bool
	var clinicCode string
	var existingPatientID sql.NullString
	err := config.DB.QueryRowContext(ctx, `
		SELECT 
			EXISTS(SELECT 1 FROM clinics WHERE id = $1 AND is_active = true) as clinic_exists,
			COALESCE((SELECT clinic_code FROM clinics WHERE id = $1), '') as clinic_code,
			(SELECT id FROM clinic_patients 
			 WHERE clinic_id = $1 AND phone = $2 
			   AND LOWER(first_name) = LOWER($3) 
			   AND LOWER(COALESCE(last_name, '')) = LOWER($4) 
			   AND is_active = true LIMIT 1) as existing_patient_id
	`, clinicID, input.Phone, input.FirstName, ptrToStr(input.LastName)).Scan(&clinicExists, &clinicCode, &existingPatientID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to validate context")
		return
	}

	if !clinicExists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Clinic not found",
			"message": "Clinic not found or is inactive",
		})
		return
	}

	if existingPatientID.Valid {
		var patient ClinicPatientResponse
		err := config.DB.QueryRowContext(ctx, `
			SELECT id, clinic_id, first_name, last_name, phone, 
			       COALESCE(email, ''), COALESCE(date_of_birth::text, ''), COALESCE(age, 0), COALESCE(gender, ''),
			       COALESCE(address1, ''), COALESCE(address2, ''), COALESCE(district, ''), COALESCE(state, ''),
			       COALESCE(mo_id, ''), COALESCE(medical_history, ''), COALESCE(allergies, ''), 
			       COALESCE(blood_group, ''), COALESCE(smoking_status, ''), COALESCE(alcohol_use, ''), 
			       COALESCE(height_cm, 0), COALESCE(weight_kg, 0), 
			       is_active, COALESCE(global_patient_id::text, ''),
			       COALESCE(current_followup_status, ''), COALESCE(last_appointment_id::text, ''), COALESCE(last_followup_id::text, ''),
			       created_at, updated_at
			FROM clinic_patients
			WHERE id = $1
		`, existingPatientID.String).Scan(
			&patient.ID, &patient.ClinicID, &patient.FirstName, &patient.LastName,
			&patient.Phone, &patient.Email, &patient.DateOfBirth, &patient.Age, &patient.Gender,
			&patient.Address1, &patient.Address2, &patient.District, &patient.State,
			&patient.MOID, &patient.MedicalHistory, &patient.Allergies, &patient.BloodGroup,
			&patient.SmokingStatus, &patient.AlcoholUse, &patient.HeightCm, &patient.WeightKg,
			&patient.IsActive, &patient.GlobalPatientID,
			&patient.CurrentFollowupStatus, &patient.LastAppointmentID, &patient.LastFollowupID,
			&patient.CreatedAt, &patient.UpdatedAt,
		)

		if err == nil {
			populateAppointmentHistory(ctx, &patient, config.DB, "", "")
			populateFullAppointmentHistory(ctx, &patient, config.DB)
			c.JSON(http.StatusOK, gin.H{
				"message": "Patient already exists in this clinic. Returning existing patient.",
				"patient": patient,
			})
			return
		}
	}

	// Start transaction for safe sequential MO ID generation
	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		middleware.SendDatabaseError(c, "Transaction initialization failed")
		return
	}
	defer tx.Rollback()

	// Advisory lock on the string hash of clinicID to serialize inserts per clinic globally
	_, _ = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock( hashtext($1) )`, clinicID)

	// Auto-generate MO ID if not provided
	var generatedMOID string
	if input.MOID == nil || *input.MOID == "" {
		if clinicCode == "" {
			middleware.SendDatabaseError(c, "Clinic code not found")
			return
		}

		// Get the highest sequential number for this clinic's MO IDs
		var maxNumber int
		err = tx.QueryRowContext(ctx, `
			SELECT COALESCE(MAX(
				CASE 
					WHEN mo_id ~ ('^' || $1 || '[0-9]+$') 
					THEN CAST(SUBSTRING(mo_id FROM LENGTH($1) + 1) AS INTEGER)
					ELSE 0
				END
			), 0) as max_num
			FROM clinic_patients 
			WHERE clinic_id = $2
		`, clinicCode, clinicID).Scan(&maxNumber)

		if err != nil {
			middleware.SendDatabaseError(c, "Failed to generate MO ID")
			return
		}

		// Generate next MO ID: {clinic_code}{sequential_number}
		generatedMOID = fmt.Sprintf("%s%04d", clinicCode, maxNumber+1)
		input.MOID = &generatedMOID
	}

	// Check if Mo ID already exists for THIS clinic
	var moIDExists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM clinic_patients 
			WHERE clinic_id = $1 AND mo_id = $2 AND is_active = true
		)
	`, clinicID, *input.MOID).Scan(&moIDExists)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check Mo ID")
		return
	}

	if moIDExists {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Mo ID exists in this clinic",
			"message": "A patient with this Mo ID already exists in your clinic",
		})
		return
	}

	// Create clinic-specific patient
	var patientID string
	var createdAt, updatedAt time.Time

	err = tx.QueryRowContext(ctx, `
		INSERT INTO clinic_patients (
			clinic_id, first_name, last_name, phone, email, date_of_birth, age, gender,
			address1, address2, district, state, mo_id, medical_history, allergies, 
			blood_group, smoking_status, alcohol_use, height_cm, weight_kg, 
			is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`, clinicID, input.FirstName, input.LastName, input.Phone, input.Email,
		input.DateOfBirth, input.Age, input.Gender, input.Address1, input.Address2,
		input.District, input.State, input.MOID, input.MedicalHistory, input.Allergies,
		input.BloodGroup, input.SmokingStatus, input.AlcoholUse, input.HeightCm, input.WeightKg).Scan(&patientID, &createdAt, &updatedAt)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to create patient")
		return
	}

	if err = tx.Commit(); err != nil {
		middleware.SendDatabaseError(c, "Failed to complete transaction")
		return
	}

	// Build response
	response := ClinicPatientResponse{
		ID:             patientID,
		ClinicID:       clinicID,
		FirstName:      input.FirstName,
		LastName:       ptrToStr(input.LastName),
		Phone:          input.Phone,
		Email:          ptrToStr(input.Email),
		DateOfBirth:    ptrToStr(input.DateOfBirth),
		Age:            ptrToInt(input.Age),
		Gender:         ptrToStr(input.Gender),
		Address1:       ptrToStr(input.Address1),
		Address2:       ptrToStr(input.Address2),
		District:       ptrToStr(input.District),
		State:          ptrToStr(input.State),
		MOID:           ptrToStr(input.MOID),
		MedicalHistory: ptrToStr(input.MedicalHistory),
		Allergies:      ptrToStr(input.Allergies),
		BloodGroup:     ptrToStr(input.BloodGroup),
		SmokingStatus:  ptrToStr(input.SmokingStatus),
		AlcoholUse:     ptrToStr(input.AlcoholUse),
		HeightCm:       ptrToInt(input.HeightCm),
		WeightKg:       ptrToInt(input.WeightKg),
		IsActive:       true,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Patient created successfully for this clinic",
		"patient": response,
	})
}

// ListClinicPatients - List all patients for a specific clinic
// GET /clinic-specific-patients
func ListClinicPatients(c *gin.Context) {
	clinicID := extractClinicIDFromContext(c)
	search := c.Query("search")
	onlyActive := c.DefaultQuery("only_active", "true")

	// Pagination parameters
	limit := getQueryInt(c, "limit", 50)
	offset := getQueryInt(c, "offset", 0)

	// ✅ NEW: Optional parameters to check follow-up eligibility for specific doctor+department
	doctorID := c.Query("doctor_id")
	departmentID := c.Query("department_id")

	if clinicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "clinic_id is required from context securely",
		})
		return
	}

	if _, err := uuid.Parse(clinicID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid clinic_id format",
		})
		return
	}

	// Build query with new status fields and optimized total count subquery to eliminate N+1
	query := `
		SELECT cp.id, cp.clinic_id, cp.first_name, cp.last_name, cp.phone, 
		       COALESCE(cp.email, ''), COALESCE(cp.date_of_birth::text, ''), COALESCE(cp.age, 0), COALESCE(cp.gender, ''),
		       COALESCE(cp.address1, ''), COALESCE(cp.address2, ''), COALESCE(cp.district, ''), COALESCE(cp.state, ''), 
		       COALESCE(cp.mo_id, ''), COALESCE(cp.medical_history, ''), COALESCE(cp.allergies, ''), 
		       COALESCE(cp.blood_group, ''), COALESCE(cp.smoking_status, ''), COALESCE(cp.alcohol_use, ''), 
		       COALESCE(cp.height_cm, 0), COALESCE(cp.weight_kg, 0), 
		       cp.is_active, COALESCE(cp.global_patient_id::text, ''),
		       COALESCE(cp.current_followup_status, ''), COALESCE(cp.last_appointment_id::text, ''), COALESCE(cp.last_followup_id::text, ''),
		       cp.created_at, cp.updated_at,
		       (SELECT COUNT(*) FROM appointments WHERE clinic_patient_id = cp.id AND is_active = true) as total_appt_count
		FROM clinic_patients cp
		WHERE cp.clinic_id = $1
	`
	args := []interface{}{clinicID}
	argIndex := 2

	if onlyActive == "true" {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, true)
		argIndex++
	}

	if search != "" {
		query += fmt.Sprintf(` AND (
			LOWER(first_name) LIKE LOWER($%d) OR 
			LOWER(last_name) LIKE LOWER($%d) OR 
			LOWER(phone) LIKE LOWER($%d) OR 
			LOWER(mo_id) LIKE LOWER($%d) OR
			LOWER(address1) LIKE LOWER($%d) OR
			LOWER(district) LIKE LOWER($%d) OR
			LOWER(state) LIKE LOWER($%d)
		)`, argIndex, argIndex, argIndex, argIndex, argIndex, argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++ // ← MUST increment so LIMIT/OFFSET get correct $N
	}

	query += " ORDER BY cp.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Fetch Patients within context bound
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch patients")
		return
	}
	defer rows.Close()

	// Pre-allocate to prevent high GC sweeps
	patients := make([]ClinicPatientResponse, 0, cap(args))
	for rows.Next() {
		var patient ClinicPatientResponse
		err := rows.Scan(
			&patient.ID, &patient.ClinicID, &patient.FirstName, &patient.LastName,
			&patient.Phone, &patient.Email, &patient.DateOfBirth, &patient.Age, &patient.Gender,
			&patient.Address1, &patient.Address2, &patient.District, &patient.State,
			&patient.MOID, &patient.MedicalHistory, &patient.Allergies, &patient.BloodGroup,
			&patient.SmokingStatus, &patient.AlcoholUse, &patient.HeightCm, &patient.WeightKg,
			&patient.IsActive, &patient.GlobalPatientID,
			&patient.CurrentFollowupStatus, &patient.LastAppointmentID, &patient.LastFollowupID,
			&patient.CreatedAt, &patient.UpdatedAt,
			&patient.TotalAppointments,
		)
		if err != nil {
			continue
		}

		patients = append(patients, patient)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse records"})
		return
	}

	// ✅ PRODUCTION OPTIMIZATION: Replace N*4 sequential blocking queries with high-performance Bulk Population
	if len(patients) > 0 {
		populatePatientsBulk(ctx, patients, config.DB, clinicID, doctorID, departmentID)
	}

	c.JSON(http.StatusOK, gin.H{
		"clinic_id": clinicID,
		"total":     len(patients),
		"patients":  patients,
	})
}

// ✅ populatePatientsBulk - High-performance Bulk Relationship Populator
// Replaces the N+1 problem by fetching all related data in constant time queries
func populatePatientsBulk(ctx context.Context, patients []ClinicPatientResponse, db *sql.DB, clinicID, doctorID, departmentID string) {
	if len(patients) == 0 {
		return
	}

	patientIDs := make([]string, len(patients))
	patientMap := make(map[string]*ClinicPatientResponse, len(patients))
	for i := range patients {
		patientIDs[i] = patients[i].ID
		patientMap[patients[i].ID] = &patients[i]
		// Initialize slices to avoid nil in JSON
		patients[i].Appointments = make([]AppointmentDetail, 0)
		patients[i].FollowUps = make([]FollowUpDetail, 0)
		patients[i].AppointmentHistory = make([]AppointmentHistoryItem, 0)
	}

	var wg sync.WaitGroup
	wg.Add(4)

	// 1. Bulk Load Full Appointments
	go func() {
		defer wg.Done()
		rows, err := db.QueryContext(ctx, `
			SELECT 
				id, clinic_patient_id, doctor_id, COALESCE(department_id::text, ''), appointment_time,
				consultation_type, status, COALESCE(fee_amount, 0), payment_status, COALESCE(payment_mode, ''),
				is_priority, created_at
			FROM appointments 
			WHERE clinic_patient_id = ANY($1) AND clinic_id = $2
			ORDER BY appointment_time DESC`, pq.Array(patientIDs), clinicID)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			var appt AppointmentDetail
			var pID string
			var apptTime, createdAt time.Time
			err := rows.Scan(
				&appt.AppointmentID, &pID, &appt.DoctorID, &appt.DepartmentID, &apptTime,
				&appt.ConsultationType, &appt.Status, &appt.FeeAmount, &appt.PaymentStatus,
				&appt.PaymentMode, &appt.IsPriority, &createdAt,
			)
			if err == nil {
				appt.AppointmentTime = apptTime.Format(time.RFC3339)
				appt.CreatedAt = createdAt.Format(time.RFC3339)
				appt.SlotType = mapConsultationTypeToSlotType(appt.ConsultationType)
				if p, ok := patientMap[pID]; ok {
					p.Appointments = append(p.Appointments, appt)
				}
			}
		}
	}()

	// 2. Bulk Load Follow-Ups
	go func() {
		defer wg.Done()
		rows, err := db.QueryContext(ctx, `
			SELECT 
				id, clinic_patient_id, source_appointment_id, doctor_id, COALESCE(department_id::text, ''),
				status, is_free, valid_from, valid_until, COALESCE(used_appointment_id::text, ''),
				COALESCE(renewed_by_appointment_id::text, ''), created_at, updated_at
			FROM follow_ups
			WHERE clinic_patient_id = ANY($1) AND clinic_id = $2
			ORDER BY created_at DESC`, pq.Array(patientIDs), clinicID)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			var fu FollowUpDetail
			var pID string
			var validFrom, validUntil, createdAt, updatedAt time.Time
			err := rows.Scan(
				&fu.FollowUpID, &pID, &fu.SourceAppointmentID, &fu.DoctorID, &fu.DepartmentID,
				&fu.Status, &fu.IsFree, &validFrom, &validUntil, &fu.UsedAppointmentID,
				&fu.RenewedByAppointmentID, &createdAt, &updatedAt,
			)
			if err == nil {
				fu.ValidFrom = validFrom.Format("2006-01-02")
				fu.ValidUntil = validUntil.Format("2006-01-02")
				fu.CreatedAt = createdAt.Format(time.RFC3339)
				fu.UpdatedAt = updatedAt.Format(time.RFC3339)

				if p, ok := patientMap[pID]; ok {
					p.FollowUps = append(p.FollowUps, fu)
					// Populate Eligible/Expired based on follow_ups table logic
					if fu.Status == "active" && validUntil.After(time.Now()) {
						p.EligibleFollowUps = append(p.EligibleFollowUps, EligibleFollowUp{
							AppointmentID:      fu.SourceAppointmentID,
							DoctorID:           fu.DoctorID,
							DoctorName:         "", // Will be populated in detailed view if needed
							DepartmentID:       fu.DepartmentID,
							Department:         "",
							AppointmentDate:    fu.ValidFrom,
							RemainingDays:      int(time.Until(validUntil).Hours() / 24),
							NextFollowUpExpiry: fu.ValidUntil,
							Note:               "Eligible for free follow-up",
						})
					} else if fu.Status == "expired" {
						p.ExpiredFollowUps = append(p.ExpiredFollowUps, ExpiredFollowUp{
							DoctorID:  fu.DoctorID,
							ExpiredOn: fu.ValidUntil,
							Note:      "Follow-up expired",
						})
					}
				}
			}
		}
	}()

	// 3. Populate Last Appointment Info (Legacy compatibility)
	go func() {
		defer wg.Done()
		rows, err := db.QueryContext(ctx, `
			SELECT DISTINCT ON (a.clinic_patient_id)
				a.id, a.clinic_patient_id, a.doctor_id, COALESCE(u.first_name || ' ' || u.last_name, u.first_name),
				COALESCE(a.department_id::text, ''), COALESCE(dept.name, ''), a.appointment_date, a.status
			FROM appointments a
			JOIN doctors d ON d.id = a.doctor_id
			JOIN users u ON u.id = d.user_id
			LEFT JOIN departments dept ON dept.id = a.department_id
			WHERE a.clinic_patient_id = ANY($1) AND a.clinic_id = $2
			  AND a.status IN ('completed', 'confirmed')
			  AND a.consultation_type NOT IN ('follow-up-via-clinic', 'follow-up-via-video')
			ORDER BY a.clinic_patient_id, a.appointment_date DESC, a.appointment_time DESC`, pq.Array(patientIDs), clinicID)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			var appt LastAppointmentInfo
			var pID string
			var apptDate time.Time
			err := rows.Scan(
				&appt.ID, &pID, &appt.DoctorID, &appt.DoctorName,
				&appt.DepartmentID, &appt.Department, &apptDate, &appt.Status,
			)
			if err == nil {
				appt.Date = apptDate.Format("2006-01-02")
				appt.DaysSince = int(time.Since(apptDate).Hours() / 24)
				if p, ok := patientMap[pID]; ok {
					p.LastAppointment = &appt
				}
			}
		}
	}()

	// 4. Bulk Load Appointment History (Last 10 per patient)
	go func() {
		defer wg.Done()
		rows, err := db.QueryContext(ctx, `
			SELECT p.id as patient_id, h.*
			FROM UNNEST($1::uuid[]) p(id)
			CROSS JOIN LATERAL (
				SELECT 
					a.id, a.doctor_id, COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
					COALESCE(a.department_id::text, ''), COALESCE(dept.name, '') as department, a.consultation_type, a.appointment_date
				FROM appointments a
				JOIN doctors d ON d.id = a.doctor_id
				JOIN users u ON u.id = d.user_id
				LEFT JOIN departments dept ON dept.id = a.department_id
				WHERE a.clinic_patient_id = p.id AND a.clinic_id = $2
				  AND a.consultation_type IN ('clinic_visit', 'video_consultation')
				  AND a.status IN ('completed', 'confirmed')
				  AND a.is_active = true
				ORDER BY a.appointment_date DESC, a.appointment_time DESC
				LIMIT 10
			) h`, pq.Array(patientIDs), clinicID)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			var pID string
			var item AppointmentHistoryItem
			var appointmentDate time.Time
			err := rows.Scan(
				&pID, &item.ID, &item.DoctorID, &item.DoctorName,
				&item.DepartmentID, &item.Department, &item.AppointmentType, &appointmentDate,
			)
			if err == nil {
				item.AppointmentDate = appointmentDate.Format("2006-01-02")
				item.DaysSince = int(time.Since(appointmentDate).Hours() / 24)
				item.ValidityDays = 5
				if p, ok := patientMap[pID]; ok {
					p.AppointmentHistory = append(p.AppointmentHistory, item)
				}
			}
		}
	}()

	wg.Wait()
}

// GetClinicPatient - Get single clinic patient
// GET /clinic-specific-patients/:id?doctor_id=xxx&department_id=xxx
func GetClinicPatient(c *gin.Context) {
	patientID := c.Param("id")
	clinicID := extractClinicIDFromContext(c)

	// ✅ NEW: Optional parameters to check follow-up eligibility for specific doctor+department
	doctorID := c.Query("doctor_id")
	departmentID := c.Query("department_id")

	if _, err := uuid.Parse(patientID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid patient_id format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var patient ClinicPatientResponse
	err := config.DB.QueryRowContext(ctx, `
		SELECT cp.id, cp.clinic_id, cp.first_name, cp.last_name, cp.phone, 
		       COALESCE(cp.email, ''), COALESCE(cp.date_of_birth::text, ''), COALESCE(cp.age, 0), COALESCE(cp.gender, ''),
		       COALESCE(cp.address1, ''), COALESCE(cp.address2, ''), COALESCE(cp.district, ''), COALESCE(cp.state, ''), 
		       COALESCE(cp.mo_id, ''), COALESCE(cp.medical_history, ''), COALESCE(cp.allergies, ''), 
		       COALESCE(cp.blood_group, ''), COALESCE(cp.smoking_status, ''), COALESCE(cp.alcohol_use, ''), 
		       COALESCE(cp.height_cm, 0), COALESCE(cp.weight_kg, 0), 
		       cp.is_active, COALESCE(cp.global_patient_id::text, ''),
		       COALESCE(cp.current_followup_status, ''), COALESCE(cp.last_appointment_id::text, ''), COALESCE(cp.last_followup_id::text, ''),
		       cp.created_at, cp.updated_at,
		       (SELECT COUNT(*) FROM appointments WHERE clinic_patient_id = cp.id AND is_active = true) as total_appt_count
		FROM clinic_patients cp
		WHERE cp.id = $1 AND cp.clinic_id = $2
	`, patientID, clinicID).Scan(
		&patient.ID, &patient.ClinicID, &patient.FirstName, &patient.LastName,
		&patient.Phone, &patient.Email, &patient.DateOfBirth, &patient.Age, &patient.Gender,
		&patient.Address1, &patient.Address2, &patient.District, &patient.State,
		&patient.MOID, &patient.MedicalHistory, &patient.Allergies, &patient.BloodGroup,
		&patient.SmokingStatus, &patient.AlcoholUse, &patient.HeightCm, &patient.WeightKg,
		&patient.IsActive, &patient.GlobalPatientID,
		&patient.CurrentFollowupStatus, &patient.LastAppointmentID, &patient.LastFollowupID,
		&patient.CreatedAt, &patient.UpdatedAt,
		&patient.TotalAppointments,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Patient not found",
			})
			return
		}
		middleware.SendDatabaseError(c, "Failed to fetch patient")
		return
	}

	// ✅ CONCURRENCY OPTIMIZATION: Populate appointment history and follow-up eligibility in parallel
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		populateAppointmentHistory(ctx, &patient, config.DB, doctorID, departmentID)
	}()

	go func() {
		defer wg.Done()
		populateFullAppointmentHistory(ctx, &patient, config.DB)
	}()

	wg.Wait()

	c.JSON(http.StatusOK, gin.H{
		"patient": patient,
	})
}

// UpdateClinicPatient - Update clinic patient
// PUT /clinic-specific-patients/:id
func UpdateClinicPatient(c *gin.Context) {
	patientID := c.Param("id")
	clinicID := extractClinicIDFromContext(c)
	var input UpdateClinicPatientInput

	if _, err := uuid.Parse(patientID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid patient_id format",
		})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if input.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, *input.FirstName)
		argIndex++
	}

	if input.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, *input.LastName)
		argIndex++
	}

	if input.Phone != nil {
		setParts = append(setParts, fmt.Sprintf("phone = $%d", argIndex))
		args = append(args, *input.Phone)
		argIndex++
	}

	if input.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *input.Email)
		argIndex++
	}

	if input.DateOfBirth != nil {
		setParts = append(setParts, fmt.Sprintf("date_of_birth = $%d", argIndex))
		args = append(args, *input.DateOfBirth)
		argIndex++
	}

	if input.Age != nil {
		setParts = append(setParts, fmt.Sprintf("age = $%d", argIndex))
		args = append(args, *input.Age)
		argIndex++
	}

	if input.Gender != nil {
		setParts = append(setParts, fmt.Sprintf("gender = $%d", argIndex))
		args = append(args, *input.Gender)
		argIndex++
	}

	if input.Address1 != nil {
		setParts = append(setParts, fmt.Sprintf("address1 = $%d", argIndex))
		args = append(args, *input.Address1)
		argIndex++
	}

	if input.Address2 != nil {
		setParts = append(setParts, fmt.Sprintf("address2 = $%d", argIndex))
		args = append(args, *input.Address2)
		argIndex++
	}

	if input.District != nil {
		setParts = append(setParts, fmt.Sprintf("district = $%d", argIndex))
		args = append(args, *input.District)
		argIndex++
	}

	if input.State != nil {
		setParts = append(setParts, fmt.Sprintf("state = $%d", argIndex))
		args = append(args, *input.State)
		argIndex++
	}

	if input.MOID != nil {
		setParts = append(setParts, fmt.Sprintf("mo_id = $%d", argIndex))
		args = append(args, *input.MOID)
		argIndex++
	}

	if input.MedicalHistory != nil {
		setParts = append(setParts, fmt.Sprintf("medical_history = $%d", argIndex))
		args = append(args, *input.MedicalHistory)
		argIndex++
	}

	if input.Allergies != nil {
		setParts = append(setParts, fmt.Sprintf("allergies = $%d", argIndex))
		args = append(args, *input.Allergies)
		argIndex++
	}

	if input.BloodGroup != nil {
		setParts = append(setParts, fmt.Sprintf("blood_group = $%d", argIndex))
		args = append(args, *input.BloodGroup)
		argIndex++
	}

	if input.SmokingStatus != nil {
		setParts = append(setParts, fmt.Sprintf("smoking_status = $%d", argIndex))
		args = append(args, *input.SmokingStatus)
		argIndex++
	}

	if input.AlcoholUse != nil {
		setParts = append(setParts, fmt.Sprintf("alcohol_use = $%d", argIndex))
		args = append(args, *input.AlcoholUse)
		argIndex++
	}

	if input.HeightCm != nil {
		setParts = append(setParts, fmt.Sprintf("height_cm = $%d", argIndex))
		args = append(args, *input.HeightCm)
		argIndex++
	}

	if input.WeightKg != nil {
		setParts = append(setParts, fmt.Sprintf("weight_kg = $%d", argIndex))
		args = append(args, *input.WeightKg)
		argIndex++
	}

	if input.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *input.IsActive)
		argIndex++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No fields to update",
		})
		return
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add patient_id
	args = append(args, patientID)

	// Add clinic_id
	args = append(args, clinicID)

	// Build UPDATE query
	query := fmt.Sprintf(`
		UPDATE clinic_patients 
		SET %s
		WHERE id = $%d AND clinic_id = $%d
		RETURNING id, clinic_id, first_name, last_name, phone, 
		          COALESCE(email, ''), COALESCE(date_of_birth::text, ''), COALESCE(age, 0), COALESCE(gender, ''),
		          COALESCE(address1, ''), COALESCE(address2, ''), COALESCE(district, ''), COALESCE(state, ''), 
		          COALESCE(mo_id, ''), COALESCE(medical_history, ''), COALESCE(allergies, ''), 
		          COALESCE(blood_group, ''), COALESCE(smoking_status, ''), COALESCE(alcohol_use, ''), 
		          COALESCE(height_cm, 0), COALESCE(weight_kg, 0), 
		          is_active, COALESCE(global_patient_id::text, ''),
		          COALESCE(current_followup_status, ''), COALESCE(last_appointment_id::text, ''), COALESCE(last_followup_id::text, ''),
		          created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex, argIndex+1)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var patient ClinicPatientResponse
	err := config.DB.QueryRowContext(ctx, query, args...).Scan(
		&patient.ID, &patient.ClinicID, &patient.FirstName, &patient.LastName,
		&patient.Phone, &patient.Email, &patient.DateOfBirth, &patient.Age, &patient.Gender,
		&patient.Address1, &patient.Address2, &patient.District, &patient.State,
		&patient.MOID, &patient.MedicalHistory, &patient.Allergies, &patient.BloodGroup,
		&patient.SmokingStatus, &patient.AlcoholUse, &patient.HeightCm, &patient.WeightKg,
		&patient.IsActive, &patient.GlobalPatientID,
		&patient.CurrentFollowupStatus, &patient.LastAppointmentID, &patient.LastFollowupID,
		&patient.CreatedAt, &patient.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Patient not found",
			})
			return
		}
		middleware.SendDatabaseError(c, "Failed to update patient")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Patient updated successfully",
		"patient": patient,
	})
}

// DeleteClinicPatient - Soft delete clinic patient
// DELETE /clinic-specific-patients/:id
func DeleteClinicPatient(c *gin.Context) {
	patientID := c.Param("id")
	clinicID := extractClinicIDFromContext(c)

	if _, err := uuid.Parse(patientID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid patient_id format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, err := config.DB.ExecContext(ctx, `
		UPDATE clinic_patients 
		SET is_active = false, updated_at = $1
		WHERE id = $2 AND clinic_id = $3
	`, time.Now(), patientID, clinicID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to delete patient")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check deletion result")
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Patient not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Patient deleted successfully",
	})
}

// ✅ Helper function to populate appointment history and follow-up eligibility
// Requires context for DB propagation bounds
func populateAppointmentHistory(ctx context.Context, patient *ClinicPatientResponse, db *sql.DB, doctorID, departmentID string) {
	// Query last appointment with all details
	var lastAppt LastAppointmentInfo
	var appointmentDate time.Time

	// Build query based on whether we're checking for a specific doctor+department
	query := `
		SELECT 
			a.id,
			a.doctor_id,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
			a.department_id,
			dept.name as department,
			a.appointment_date,
			a.status
		FROM appointments a
		JOIN doctors d ON d.id = a.doctor_id
		JOIN users u ON u.id = d.user_id
		LEFT JOIN departments dept ON dept.id = a.department_id
		WHERE a.clinic_patient_id = $1
		  AND a.clinic_id = $2
		  AND a.status IN ('completed', 'confirmed')
		  AND a.consultation_type NOT IN ('follow-up-via-clinic', 'follow-up-via-video')`

	args := []interface{}{patient.ID, patient.ClinicID}
	argIndex := 3

	// ✅ If checking for specific doctor+department, filter by them
	if doctorID != "" {
		query += fmt.Sprintf(" AND a.doctor_id = $%d", argIndex)
		args = append(args, doctorID)
		argIndex++
	}

	if departmentID != "" {
		query += fmt.Sprintf(" AND a.department_id = $%d", argIndex)
		args = append(args, departmentID)
		argIndex++
	}

	// ✅ OPTIMIZATION: If we have a cached last_appointment_id and no specific doctor/dept filters, use it directly
	if patient.LastAppointmentID != "" && doctorID == "" && departmentID == "" {
		query = `
			SELECT 
				a.id,
				a.doctor_id,
				COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
				a.department_id,
				dept.name as department,
				a.appointment_date,
				a.status
			FROM appointments a
			JOIN doctors d ON d.id = a.doctor_id
			JOIN users u ON u.id = d.user_id
			LEFT JOIN departments dept ON dept.id = a.department_id
			WHERE a.id = $1`
		args = []interface{}{patient.LastAppointmentID}
	} else {
		query += " ORDER BY a.appointment_date DESC, a.appointment_time DESC LIMIT 1"
	}

	err := db.QueryRowContext(ctx, query, args...).Scan(
		&lastAppt.ID,
		&lastAppt.DoctorID,
		&lastAppt.DoctorName,
		&lastAppt.DepartmentID,
		&lastAppt.Department,
		&appointmentDate,
		&lastAppt.Status,
	)

	if err == nil {
		// Calculate days since last appointment
		daysSince := int(time.Since(appointmentDate).Hours() / 24)
		lastAppt.DaysSince = daysSince
		lastAppt.Date = appointmentDate.Format("2006-01-02")
		patient.LastAppointment = &lastAppt

		// ✅ NEW: Use follow-up helper to check eligibility from follow_ups table
		followUpHelper := &utils.FollowUpHelper{DB: db}
		eligibility := &FollowUpEligibility{}

		// If doctor+department specified, check for that combination
		var deptID *string
		if departmentID != "" {
			deptID = &departmentID
		} else if lastAppt.DepartmentID != "" {
			deptID = &lastAppt.DepartmentID
		}

		checkDoctorID := doctorID
		if checkDoctorID == "" {
			checkDoctorID = lastAppt.DoctorID
		}

		isFree, isEligible, message, err := followUpHelper.CheckFollowUpEligibility(
			patient.ID,
			patient.ClinicID,
			checkDoctorID,
			deptID,
		)

		if err == nil {
			eligibility.Eligible = isEligible
			eligibility.IsFree = isFree
			eligibility.Message = message

			if isFree && isEligible {
				// Calculate days remaining from follow_ups table
				if daysSince <= 5 {
					eligibility.DaysRemaining = 5 - daysSince
				}
			}
		} else {
			// Fallback to default message if query fails
			eligibility.Eligible = false
			eligibility.IsFree = false
			eligibility.Reason = "Could not check follow-up eligibility"
		}

		patient.FollowUpEligibility = eligibility
	} else {
		// No previous appointment - not eligible for any follow-up
		patient.LastAppointment = nil
		patient.FollowUpEligibility = &FollowUpEligibility{
			Eligible: false,
			IsFree:   false,
			Reason:   "No previous appointment found",
		}
	}

	// Only count if not already populated from main query (optimization)
	if patient.TotalAppointments == 0 {
		var totalCount int
		err = db.QueryRowContext(ctx, `
			SELECT COUNT(*) 
			FROM appointments 
			WHERE clinic_patient_id = $1 AND clinic_id = $2 AND is_active = true
		`, patient.ID, patient.ClinicID).Scan(&totalCount)

		if err == nil {
			patient.TotalAppointments = totalCount
		}
	}
}

// ✅ NEW: Populate full appointment history with follow-up validity using follow_ups table
func populateFullAppointmentHistory(ctx context.Context, patient *ClinicPatientResponse, db *sql.DB) {
	// Use the new follow-up helper to get clean data from follow_ups table
	followUpHelper := &utils.FollowUpHelper{DB: db}

	// Get all active follow-ups
	activeFollowUps, err := followUpHelper.GetActiveFollowUps(patient.ID, patient.ClinicID)
	if err == nil && len(activeFollowUps) > 0 {
		// Convert to EligibleFollowUp format
		for _, active := range activeFollowUps {
			eligible := EligibleFollowUp{
				AppointmentID:      active.AppointmentID,
				DoctorID:           active.DoctorID,
				DoctorName:         active.DoctorName,
				DepartmentID:       ptrToStr(active.DepartmentID),
				Department:         ptrToStr(active.DepartmentName),
				AppointmentDate:    active.AppointmentDate,
				RemainingDays:      active.DaysRemaining,
				NextFollowUpExpiry: active.ValidUntil,
				Note:               active.Note,
			}
			patient.EligibleFollowUps = append(patient.EligibleFollowUps, eligible)
		}
	}

	// Get expired follow-ups
	expiredFollowUps, err := followUpHelper.GetExpiredFollowUps(patient.ID, patient.ClinicID)
	if err == nil && len(expiredFollowUps) > 0 {
		// Convert to ExpiredFollowUp format
		for _, expired := range expiredFollowUps {
			expiredItem := ExpiredFollowUp{
				AppointmentID: "", // Not tracking specific appointment for expired
				DoctorID:      expired.DoctorID,
				DoctorName:    expired.DoctorName,
				DepartmentID:  ptrToStr(expired.DepartmentID),
				Department:    ptrToStr(expired.DepartmentName),
				ExpiredOn:     expired.ExpiredOn,
				Note:          expired.Note,
			}
			patient.ExpiredFollowUps = append(patient.ExpiredFollowUps, expiredItem)
		}
	}

	// ✅ SIMPLIFIED: Get basic appointment history (optional - for display purposes only)
	// The follow_ups table already populated EligibleFollowUps and ExpiredFollowUps above
	rows, err := db.QueryContext(ctx, `
		SELECT 
			a.id,
			a.doctor_id,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
			COALESCE(a.department_id::text, ''),
			COALESCE(dept.name, '') as department,
			a.consultation_type,
			a.appointment_date
		FROM appointments a
		JOIN doctors d ON d.id = a.doctor_id
		JOIN users u ON u.id = d.user_id
		LEFT JOIN departments dept ON dept.id = a.department_id
		WHERE a.clinic_patient_id = $1
		  AND a.clinic_id = $2
		  AND a.consultation_type IN ('clinic_visit', 'video_consultation')
		  AND a.status IN ('completed', 'confirmed')
		ORDER BY a.appointment_date DESC, a.appointment_time DESC
		LIMIT 10
	`, patient.ID, patient.ClinicID)

	if err != nil {
		return
	}
	defer rows.Close()

	var history []AppointmentHistoryItem

	// Simple appointment history without complex follow-up logic
	for rows.Next() {
		var item AppointmentHistoryItem
		var appointmentDate time.Time

		err := rows.Scan(
			&item.ID,
			&item.DoctorID,
			&item.DoctorName,
			&item.DepartmentID,
			&item.Department,
			&item.AppointmentType,
			&appointmentDate,
		)

		if err != nil {
			continue
		}

		item.AppointmentDate = appointmentDate.Format("2006-01-02")
		item.DaysSince = int(time.Since(appointmentDate).Hours() / 24)
		item.ValidityDays = 5

		// Simple status determination
		if item.DaysSince < 0 {
			item.Status = "future"
			item.Note = fmt.Sprintf("Scheduled appointment with %s", item.DoctorName)
		} else if item.DaysSince <= 5 {
			item.Status = "active"
			item.Note = fmt.Sprintf("Recent appointment with %s", item.DoctorName)
		} else {
			item.Status = "expired"
			item.Note = fmt.Sprintf("Past appointment with %s", item.DoctorName)
		}

		history = append(history, item)
	}

	patient.AppointmentHistory = history
	// Note: EligibleFollowUps and ExpiredFollowUps already populated from follow_ups table above
}

// ✅ populateAppointmentsArray - Populate full appointments array
func populateAppointmentsArray(ctx context.Context, patient *ClinicPatientResponse, db *sql.DB) {
	rows, err := db.QueryContext(ctx, `
		SELECT 
			a.id,
			a.doctor_id,
			COALESCE(a.department_id::text, ''),
			a.appointment_time,
			a.consultation_type,
			a.status,
			COALESCE(a.fee_amount, 0),
			a.payment_status,
			COALESCE(a.payment_mode, ''),
			a.is_priority,
			a.created_at
		FROM appointments a
		WHERE a.clinic_patient_id = $1
		  AND a.clinic_id = $2
		ORDER BY a.appointment_time DESC
	`, patient.ID, patient.ClinicID)

	if err != nil {
		return
	}
	defer rows.Close()

	appointments := make([]AppointmentDetail, 0, 10) // Fast Pre-allocation
	for rows.Next() {
		var appt AppointmentDetail
		var appointmentTime time.Time
		var createdAt time.Time

		err := rows.Scan(
			&appt.AppointmentID,
			&appt.DoctorID,
			&appt.DepartmentID,
			&appointmentTime,
			&appt.ConsultationType,
			&appt.Status,
			&appt.FeeAmount,
			&appt.PaymentStatus,
			&appt.PaymentMode,
			&appt.IsPriority,
			&createdAt,
		)

		if err != nil {
			continue
		}

		appt.AppointmentTime = appointmentTime.Format(time.RFC3339)
		appt.CreatedAt = createdAt.Format(time.RFC3339)
		appt.SlotType = mapConsultationTypeToSlotType(appt.ConsultationType)

		appointments = append(appointments, appt)
	}

	patient.Appointments = appointments
}

// ✅ populateFollowUpsArray - Populate full follow-ups array
func populateFollowUpsArray(ctx context.Context, patient *ClinicPatientResponse, db *sql.DB) {
	rows, err := db.QueryContext(ctx, `
		SELECT 
			id,
			source_appointment_id,
			doctor_id,
			COALESCE(department_id::text, ''),
			status,
			is_free,
			valid_from,
			valid_until,
			COALESCE(used_appointment_id::text, ''),
			COALESCE(renewed_by_appointment_id::text, ''),
			created_at,
			updated_at
		FROM follow_ups
		WHERE clinic_patient_id = $1
		  AND clinic_id = $2
		ORDER BY created_at DESC
	`, patient.ID, patient.ClinicID)

	if err != nil {
		return
	}
	defer rows.Close()

	followUps := make([]FollowUpDetail, 0, 10) // Fast pre-allocation
	for rows.Next() {
		var fu FollowUpDetail
		var validFrom, validUntil time.Time
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&fu.FollowUpID,
			&fu.SourceAppointmentID,
			&fu.DoctorID,
			&fu.DepartmentID,
			&fu.Status,
			&fu.IsFree,
			&validFrom,
			&validUntil,
			&fu.UsedAppointmentID,
			&fu.RenewedByAppointmentID,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			continue
		}

		fu.ValidFrom = validFrom.Format("2006-01-02")
		fu.ValidUntil = validUntil.Format("2006-01-02")
		fu.CreatedAt = createdAt.Format(time.RFC3339)
		fu.UpdatedAt = updatedAt.Format(time.RFC3339)

		followUps = append(followUps, fu)
	}

	patient.FollowUps = followUps
}

// Helper function to map consultation type to slot type
func mapConsultationTypeToSlotType(consultationType string) string {
	switch consultationType {
	case "clinic_visit":
		return "clinic_visit"
	case "video_consultation":
		return "video_consultation"
	case "follow-up-via-clinic":
		return "clinic_followup"
	case "follow-up-via-video":
		return "video_followup"
	default:
		return consultationType
	}
}
