package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// FollowUpManager handles all follow-up related operations
type FollowUpManager struct {
	DB *sql.DB
}

// FollowUpRecord represents a follow-up eligibility record
type FollowUpRecord struct {
	ID                     string
	ClinicPatientID        string
	ClinicID               string
	DoctorID               string
	DepartmentID           *string
	SourceAppointmentID    string
	Status                 string // active, used, expired, renewed
	IsFree                 bool
	ValidFrom              time.Time
	ValidUntil             time.Time
	UsedAt                 *time.Time
	UsedAppointmentID      *string
	RenewedAt              *time.Time
	RenewedByAppointmentID *string
	FollowUpLogicStatus    string // new, expired, used, renewed
	LogicNotes             *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// CreateFollowUp creates a new follow-up eligibility record
// Called when a regular appointment (clinic_visit or video_consultation) is created
func (fm *FollowUpManager) CreateFollowUp(clinicPatientID, clinicID, doctorID string, departmentID *string, appointmentID string, appointmentDate time.Time) error {
	log.Printf("🔄 CreateFollowUp called: Patient=%s, Doctor=%s, Dept=%v, AppointmentID=%s, Date=%s",
		clinicPatientID, doctorID, departmentID, appointmentID, appointmentDate.Format("2006-01-02"))

	validFrom := appointmentDate
	validUntil := appointmentDate.AddDate(0, 0, 5) // 5 days validity

	log.Printf("📅 Follow-up validity: From=%s, Until=%s", validFrom.Format("2006-01-02"), validUntil.Format("2006-01-02"))

	// First, check if there's an existing active or expired follow-up for this doctor+department
	// If yes, mark it as "renewed"
	err := fm.RenewExistingFollowUps(clinicPatientID, clinicID, doctorID, departmentID, appointmentID)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to renew existing follow-ups: %v", err)
		// Don't fail - continue creating new follow-up
	}

	// Create new follow-up record
	logicNotes := fmt.Sprintf("Patient gets one free follow-up valid for 5 days. If used, it expires. If not used after 5 days, it also expires. Subsequent regular appointments with same doctor+department generate new free follow-up.")

	_, err = fm.DB.Exec(`
		INSERT INTO follow_ups (
			clinic_patient_id, clinic_id, doctor_id, department_id,
			source_appointment_id, status, is_free, valid_from, valid_until,
			follow_up_logic_status, logic_notes,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, 'active', true, $6, $7, 'new', $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, clinicPatientID, clinicID, doctorID, departmentID, appointmentID, validFrom, validUntil, logicNotes)

	if err != nil {
		log.Printf("❌ Failed to create follow-up record: %v", err)
		return fmt.Errorf("failed to create follow-up record: %w", err)
	}

	log.Printf("✅ Created follow-up eligibility: Patient=%s, Doctor=%s, Valid until=%s",
		clinicPatientID, doctorID, validUntil.Format("2006-01-02"))

	return nil
}

// RenewExistingFollowUps marks existing follow-ups as "renewed" for this doctor+department combination
func (fm *FollowUpManager) RenewExistingFollowUps(clinicPatientID, clinicID, doctorID string, departmentID *string, newAppointmentID string) error {
	log.Printf("🔄 RenewExistingFollowUps called: Patient=%s, Doctor=%s, Dept=%v, NewAppointment=%s",
		clinicPatientID, doctorID, departmentID, newAppointmentID)

	query := `
		UPDATE follow_ups
		SET status = 'renewed',
		    renewed_at = CURRENT_TIMESTAMP,
		    renewed_by_appointment_id = $1,
		    follow_up_logic_status = 'renewed',
		    logic_notes = 'Old follow-up renewed by new regular appointment',
		    updated_at = CURRENT_TIMESTAMP
		WHERE clinic_patient_id = $2
		  AND clinic_id = $3
		  AND doctor_id = $4
		  AND status IN ('active', 'expired')
	`

	args := []interface{}{newAppointmentID, clinicPatientID, clinicID, doctorID}

	// Add department filter if provided
	if departmentID != nil {
		query += ` AND department_id = $5`
		args = append(args, *departmentID)
	} else {
		query += ` AND department_id IS NULL`
	}

	log.Printf("🔄 Executing renewal query: %s with args: %v", query, args)

	result, err := fm.DB.Exec(query, args...)
	if err != nil {
		log.Printf("❌ Failed to renew existing follow-ups: %v", err)
		return fmt.Errorf("failed to renew existing follow-ups: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("✅ Renewed %d existing follow-up(s) for Patient=%s, Doctor=%s",
			rowsAffected, clinicPatientID, doctorID)
	} else {
		log.Printf("ℹ️ No existing follow-ups to renew for Patient=%s, Doctor=%s",
			clinicPatientID, doctorID)
	}

	return nil
}

// MarkFollowUpAsUsed marks a follow-up as used when a follow-up appointment is created
func (fm *FollowUpManager) MarkFollowUpAsUsed(clinicPatientID, clinicID, doctorID string, departmentID *string, followUpAppointmentID string) error {
	log.Printf("🔧 MarkFollowUpAsUsed called with: Patient=%s, Clinic=%s, Doctor=%s, Dept=%v, Appointment=%s",
		clinicPatientID, clinicID, doctorID, departmentID, followUpAppointmentID)

	// First, get the follow-up ID that will be marked as used
	var followUpID string
	getQuery := `
		SELECT id
		FROM follow_ups
		WHERE clinic_patient_id = $1
		  AND clinic_id = $2
		  AND doctor_id = $3
		  AND status = 'active'
		  AND is_free = true
		  AND valid_until >= CURRENT_DATE
	`

	getArgs := []interface{}{clinicPatientID, clinicID, doctorID}

	// Add department filter if provided
	if departmentID != nil {
		getQuery += ` AND department_id = $4`
		getArgs = append(getArgs, *departmentID)
	} else {
		getQuery += ` AND department_id IS NULL`
	}

	getQuery += ` ORDER BY created_at DESC LIMIT 1`

	err := fm.DB.QueryRow(getQuery, getArgs...).Scan(&followUpID)
	if err != nil {
		log.Printf("⚠️ No active free follow-up found: %v", err)
		return fmt.Errorf("no active free follow-up found: %w", err)
	}

	log.Printf("✅ Found follow-up to mark as used: %s", followUpID)

	// Now update the follow-up status
	updateQuery := `
		UPDATE follow_ups
		SET status = 'used',
		    used_at = CURRENT_TIMESTAMP,
		    used_appointment_id = $1,
		    follow_up_logic_status = 'used',
		    logic_notes = 'Free follow-up was used. Patient can book follow-up again but next one is PAID.',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := fm.DB.Exec(updateQuery, followUpAppointmentID, followUpID)
	if err != nil {
		log.Printf("❌ Failed to update follow-up: %v", err)
		return fmt.Errorf("failed to mark follow-up as used: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("✅ Marked follow-up as used: FollowUpID=%s, AppointmentID=%s",
			followUpID, followUpAppointmentID)

		// ✅ ALSO UPDATE clinic_patient status to 'used'
		_, err = fm.DB.Exec(`
			UPDATE clinic_patients
			SET current_followup_status = 'used',
			    last_appointment_id = $1,
			    last_followup_id = $2,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
			  AND clinic_id = $4
		`, followUpAppointmentID, followUpID, clinicPatientID, clinicID)

		if err != nil {
			log.Printf("⚠️ Warning: Failed to update clinic_patient status to 'used': %v", err)
		} else {
			log.Printf("✅ Updated clinic_patient status to 'used' for patient=%s", clinicPatientID)
		}
	}

	return nil
}

// GetActiveFollowUp gets the active follow-up for a patient with a specific doctor+department
func (fm *FollowUpManager) GetActiveFollowUp(clinicPatientID, clinicID, doctorID string, departmentID *string) (*FollowUpRecord, error) {
	log.Printf("🔍 GetActiveFollowUp: Patient=%s, Clinic=%s, Doctor=%s, Dept=%v",
		clinicPatientID, clinicID, doctorID, departmentID)

	query := `
		SELECT id, clinic_patient_id, clinic_id, doctor_id, department_id,
		       source_appointment_id, status, is_free, valid_from, valid_until,
		       used_at, used_appointment_id, renewed_at, renewed_by_appointment_id,
		       follow_up_logic_status, logic_notes,
		       created_at, updated_at
		FROM follow_ups
		WHERE clinic_patient_id = $1
		  AND clinic_id = $2
		  AND doctor_id = $3
		  AND status = 'active'
		  AND valid_until >= CURRENT_DATE
	`

	args := []interface{}{clinicPatientID, clinicID, doctorID}

	// Add department filter if provided
	if departmentID != nil {
		query += ` AND department_id = $4`
		args = append(args, *departmentID)
	} else {
		query += ` AND department_id IS NULL`
	}

	query += ` ORDER BY created_at DESC LIMIT 1`

	log.Printf("🔍 Executing query: %s with args: %v", query, args)

	var record FollowUpRecord
	err := fm.DB.QueryRow(query, args...).Scan(
		&record.ID, &record.ClinicPatientID, &record.ClinicID, &record.DoctorID, &record.DepartmentID,
		&record.SourceAppointmentID, &record.Status, &record.IsFree, &record.ValidFrom, &record.ValidUntil,
		&record.UsedAt, &record.UsedAppointmentID, &record.RenewedAt, &record.RenewedByAppointmentID,
		&record.FollowUpLogicStatus, &record.LogicNotes,
		&record.CreatedAt, &record.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		log.Printf("⚠️ No active follow-up found")
		return nil, nil // No active follow-up found (not an error)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get active follow-up: %w", err)
	}

	return &record, nil
}

// CheckFollowUpEligibility checks if a patient is eligible for follow-up with a doctor+department
// Returns: (isFree bool, isEligible bool, message string, error)
func (fm *FollowUpManager) CheckFollowUpEligibility(clinicPatientID, clinicID, doctorID string, departmentID *string) (bool, bool, string, error) {
	// ✅ FIRST: Auto-expire any old follow-ups that have passed their valid_until date
	fm.ExpireOldFollowUps()

	// Check if there's an active free follow-up
	log.Printf("🔍 CheckFollowUpEligibility: Patient=%s, Clinic=%s, Doctor=%s, Dept=%v",
		clinicPatientID, clinicID, doctorID, departmentID)

	// Get active follow-up (only returns if valid_until >= CURRENT_DATE)
	activeFollowUp, err := fm.GetActiveFollowUp(clinicPatientID, clinicID, doctorID, departmentID)
	if err != nil {
		log.Printf("❌ GetActiveFollowUp error: %v", err)
		return false, false, "", err
	}

	log.Printf("🔍 GetActiveFollowUp result: %+v", activeFollowUp)

	// ✅ Found active free follow-up
	if activeFollowUp != nil && activeFollowUp.IsFree {
		daysRemaining := int(time.Until(activeFollowUp.ValidUntil).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}
		log.Printf("✅ Found active free follow-up: %d days remaining", daysRemaining)
		return true, true, fmt.Sprintf("Free follow-up available (%d days remaining)", daysRemaining), nil
	}

	// ✅ Check if patient has any expired or used follow-ups for this doctor
	checkQuery := `
		SELECT status, is_free, valid_until
		FROM follow_ups
		WHERE clinic_patient_id = $1
		  AND clinic_id = $2
		  AND doctor_id = $3
	`
	checkArgs := []interface{}{clinicPatientID, clinicID, doctorID}

	if departmentID != nil {
		checkQuery += ` AND department_id = $4`
		checkArgs = append(checkArgs, *departmentID)
	} else {
		checkQuery += ` AND department_id IS NULL`
	}

	checkQuery += ` ORDER BY created_at DESC LIMIT 1`

	var followUpStatus string
	var isFree bool
	var validUntil time.Time

	err = fm.DB.QueryRow(checkQuery, checkArgs...).Scan(&followUpStatus, &isFree, &validUntil)

	if err == nil {
		// Found a follow-up record (even if expired/used)
		if followUpStatus == "expired" {
			log.Printf("⏰ Follow-up expired on %s", validUntil.Format("2006-01-02"))
			return false, true, "Free follow-up expired. This follow-up requires payment.", nil
		}
		if followUpStatus == "used" {
			log.Printf("✅ Follow-up already used")
			return false, true, "Free follow-up already used. This follow-up requires payment.", nil
		}
		// If status is "active" but not returned by GetActiveFollowUp, it might be expired but not marked
		if followUpStatus == "active" && validUntil.Before(time.Now()) {
			log.Printf("⏰ Follow-up active but expired (should be marked as expired)")
			return false, true, "Free follow-up expired. This follow-up requires payment.", nil
		}
	}

	// Check if patient has ANY appointment with this doctor+department (even if expired)
	// This determines if they can book a PAID follow-up
	query := `
		SELECT EXISTS(
			SELECT 1 FROM appointments
			WHERE clinic_patient_id = $1
			  AND clinic_id = $2
			  AND doctor_id = $3
			  AND consultation_type IN ('clinic_visit', 'video_consultation')
			  AND status IN ('completed', 'confirmed')
	`

	args := []interface{}{clinicPatientID, clinicID, doctorID}

	if departmentID != nil {
		query += ` AND department_id = $4`
		args = append(args, *departmentID)
	}

	query += `)`

	var hasPreviousAppointment bool
	err = fm.DB.QueryRow(query, args...).Scan(&hasPreviousAppointment)
	if err != nil {
		return false, false, "", fmt.Errorf("failed to check previous appointments: %w", err)
	}

	if hasPreviousAppointment {
		return false, true, "Follow-up available (payment required)", nil
	}

	return false, false, "No previous appointment found with this doctor", nil
}

// ExpireOldFollowUps marks follow-ups as expired if they're past their validity date
// This should be called periodically (e.g., daily cron job) or on-demand
// ExpireOldFollowUps automatically expires follow-ups that are past their valid_until date
// and updates the clinic_patient status accordingly
func (fm *FollowUpManager) ExpireOldFollowUps() (int64, error) {
	return fm.ExpireOldFollowUpsContext(context.Background())
}

// ExpireOldFollowUpsContext marks follow-ups as expired if they're past their validity date
func (fm *FollowUpManager) ExpireOldFollowUpsContext(ctx context.Context) (int64, error) {
	tx, err := fm.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 1. Update clinic_patients status for those whose LATEST follow-up is expiring
	patientUpdateRes, err := tx.ExecContext(ctx, `
		UPDATE clinic_patients cp
		SET current_followup_status = 'expired',
		    last_followup_id = f.id,
		    updated_at = CURRENT_TIMESTAMP
		FROM (
			SELECT DISTINCT ON (clinic_patient_id) id, clinic_patient_id
			FROM follow_ups
			WHERE status = 'active' AND valid_until < CURRENT_DATE
			ORDER BY clinic_patient_id, created_at DESC
		) f
		WHERE cp.id = f.clinic_patient_id
	`)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to bulk update clinic_patients status: %v", err)
	}

	// 2. Update all active but past-due follow_ups to 'expired'
	res, err := tx.ExecContext(ctx, `
		UPDATE follow_ups
		SET status = 'expired',
		    follow_up_logic_status = 'expired',
		    logic_notes = 'Follow-up expired after 5 days. Patient can book follow-up again but next one is PAID.',
		    updated_at = CURRENT_TIMESTAMP
		WHERE status = 'active' AND valid_until < CURRENT_DATE
	`)

	if err != nil {
		return 0, fmt.Errorf("failed to expire follow-ups: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	expiredCount, _ := res.RowsAffected()
	if expiredCount > 0 {
		patientCount, _ := patientUpdateRes.RowsAffected()
		log.Printf("⏰ Expired %d follow-up(s) and updated %d patient(s) status", expiredCount, patientCount)
	}

	return expiredCount, nil
}

// GetAllActiveFollowUps gets all active follow-ups for a patient
func (fm *FollowUpManager) GetAllActiveFollowUps(clinicPatientID, clinicID string) ([]FollowUpRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := fm.DB.QueryContext(ctx, `
		SELECT id, clinic_patient_id, clinic_id, doctor_id, department_id,
		       source_appointment_id, status, is_free, valid_from, valid_until,
		       used_at, used_appointment_id, renewed_at, renewed_by_appointment_id,
		       follow_up_logic_status, logic_notes,
		       created_at, updated_at
		FROM follow_ups
		WHERE clinic_patient_id = $1
		  AND clinic_id = $2
		  AND status = 'active'
		  AND valid_until >= CURRENT_DATE
		ORDER BY valid_until ASC
	`, clinicPatientID, clinicID)

	if err != nil {
		return nil, fmt.Errorf("failed to get active follow-ups: %w", err)
	}
	defer rows.Close()

	records := make([]FollowUpRecord, 0, 5)
	for rows.Next() {
		var record FollowUpRecord
		err := rows.Scan(
			&record.ID, &record.ClinicPatientID, &record.ClinicID, &record.DoctorID, &record.DepartmentID,
			&record.SourceAppointmentID, &record.Status, &record.IsFree, &record.ValidFrom, &record.ValidUntil,
			&record.UsedAt, &record.UsedAppointmentID, &record.RenewedAt, &record.RenewedByAppointmentID,
			&record.FollowUpLogicStatus, &record.LogicNotes,
			&record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			continue
		}
		records = append(records, record)
	}

	return records, nil
}
