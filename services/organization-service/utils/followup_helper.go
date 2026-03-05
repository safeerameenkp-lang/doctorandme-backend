package utils

import (
	"database/sql"
	"fmt"
	"time"
)

// FollowUpHelper provides helper functions for querying follow-up data
// This is READ-ONLY - creation/updates happen in appointment-service
type FollowUpHelper struct {
	DB *sql.DB
}

// ActiveFollowUpSummary represents an active follow-up for display
type ActiveFollowUpSummary struct {
	FollowUpID      string
	DoctorID        string
	DoctorName      string
	DepartmentID    *string
	DepartmentName  *string
	AppointmentID   string
	AppointmentDate string
	ValidUntil      string
	DaysRemaining   int
	IsFree          bool
	Note            string
}

// ExpiredFollowUpSummary represents an expired follow-up
type ExpiredFollowUpSummary struct {
	DoctorID       string
	DoctorName     string
	DepartmentID   *string
	DepartmentName *string
	ExpiredOn      string
	Note           string
}

// GetActiveFollowUps gets all active follow-ups for a patient
func (fh *FollowUpHelper) GetActiveFollowUps(clinicPatientID, clinicID string) ([]ActiveFollowUpSummary, error) {
	rows, err := fh.DB.Query(`
		SELECT 
			f.id,
			f.doctor_id,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
			f.department_id,
			dept.name as department_name,
			f.source_appointment_id,
			f.valid_from,
			f.valid_until,
			f.is_free
		FROM follow_ups f
		JOIN doctors d ON d.id = f.doctor_id
		JOIN users u ON u.id = d.user_id
		LEFT JOIN departments dept ON dept.id = f.department_id
		WHERE f.clinic_patient_id = $1
		  AND f.clinic_id = $2
		  AND f.status = 'active'
		  AND f.valid_until >= CURRENT_DATE
		ORDER BY f.valid_until ASC
	`, clinicPatientID, clinicID)

	if err != nil {
		return nil, fmt.Errorf("failed to get active follow-ups: %w", err)
	}
	defer rows.Close()

	var summaries []ActiveFollowUpSummary
	for rows.Next() {
		var summary ActiveFollowUpSummary
		var validFrom, validUntil time.Time

		err := rows.Scan(
			&summary.FollowUpID,
			&summary.DoctorID,
			&summary.DoctorName,
			&summary.DepartmentID,
			&summary.DepartmentName,
			&summary.AppointmentID,
			&validFrom,
			&validUntil,
			&summary.IsFree,
		)

		if err != nil {
			continue
		}

		summary.AppointmentDate = validFrom.Format("2006-01-02")
		summary.ValidUntil = validUntil.Format("2006-01-02")
		summary.DaysRemaining = int(time.Until(validUntil).Hours() / 24)

		deptName := "General"
		if summary.DepartmentName != nil {
			deptName = *summary.DepartmentName
		}

		if summary.IsFree {
			summary.Note = fmt.Sprintf("Eligible for FREE follow-up with %s (%s)", summary.DoctorName, deptName)
		} else {
			summary.Note = fmt.Sprintf("Follow-up available with %s (%s) - payment required", summary.DoctorName, deptName)
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetExpiredFollowUps gets expired follow-ups that need renewal
func (fh *FollowUpHelper) GetExpiredFollowUps(clinicPatientID, clinicID string) ([]ExpiredFollowUpSummary, error) {
	// Get the most recent expired follow-up per doctor+department combo
	rows, err := fh.DB.Query(`
		WITH ranked_expired AS (
			SELECT 
				f.doctor_id,
				COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
				f.department_id,
				dept.name as department_name,
				f.valid_until,
				ROW_NUMBER() OVER (PARTITION BY f.doctor_id, f.department_id ORDER BY f.valid_until DESC) as rn
			FROM follow_ups f
			JOIN doctors d ON d.id = f.doctor_id
			JOIN users u ON u.id = d.user_id
			LEFT JOIN departments dept ON dept.id = f.department_id
			WHERE f.clinic_patient_id = $1
			  AND f.clinic_id = $2
			  AND f.status = 'expired'
		)
		SELECT doctor_id, doctor_name, department_id, department_name, valid_until
		FROM ranked_expired
		WHERE rn = 1
		ORDER BY valid_until DESC
	`, clinicPatientID, clinicID)

	if err != nil {
		return nil, fmt.Errorf("failed to get expired follow-ups: %w", err)
	}
	defer rows.Close()

	var summaries []ExpiredFollowUpSummary
	for rows.Next() {
		var summary ExpiredFollowUpSummary
		var validUntil time.Time

		err := rows.Scan(
			&summary.DoctorID,
			&summary.DoctorName,
			&summary.DepartmentID,
			&summary.DepartmentName,
			&validUntil,
		)

		if err != nil {
			continue
		}

		summary.ExpiredOn = validUntil.Format("2006-01-02")

		deptName := "General"
		if summary.DepartmentName != nil {
			deptName = *summary.DepartmentName
		}

		summary.Note = fmt.Sprintf("Follow-up expired — book a new regular appointment with %s (%s) to restart your free follow-up", summary.DoctorName, deptName)

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// CheckFollowUpEligibility checks if patient has active follow-up with specific doctor+department
func (fh *FollowUpHelper) CheckFollowUpEligibility(clinicPatientID, clinicID, doctorID string, departmentID *string) (bool, bool, string, error) {
	query := `
		SELECT id, is_free, valid_until
		FROM follow_ups
		WHERE clinic_patient_id = $1
		  AND clinic_id = $2
		  AND doctor_id = $3
		  AND status = 'active'
		  AND valid_until >= CURRENT_DATE
	`

	args := []interface{}{clinicPatientID, clinicID, doctorID}

	if departmentID != nil {
		query += ` AND department_id = $4`
		args = append(args, *departmentID)
	} else {
		query += ` AND department_id IS NULL`
	}

	query += ` ORDER BY created_at DESC LIMIT 1`

	var followUpID string
	var isFree bool
	var validUntil time.Time

	err := fh.DB.QueryRow(query, args...).Scan(&followUpID, &isFree, &validUntil)

	if err == sql.ErrNoRows {
		// No active follow-up - check if patient has any previous appointment
		checkQuery := `
			SELECT EXISTS(
				SELECT 1 FROM appointments
				WHERE clinic_patient_id = $1
				  AND clinic_id = $2
				  AND doctor_id = $3
				  AND consultation_type IN ('clinic_visit', 'video_consultation')
				  AND status IN ('completed', 'confirmed')
		`

		checkArgs := []interface{}{clinicPatientID, clinicID, doctorID}

		if departmentID != nil {
			checkQuery += ` AND department_id = $4`
			checkArgs = append(checkArgs, *departmentID)
		}

		checkQuery += `)`

		var hasPrevious bool
		err = fh.DB.QueryRow(checkQuery, checkArgs...).Scan(&hasPrevious)
		if err != nil {
			return false, false, "", fmt.Errorf("failed to check previous appointments: %w", err)
		}

		if hasPrevious {
			return false, true, "Follow-up available (payment required)", nil
		}

		return false, false, "No previous appointment found", nil
	}

	if err != nil {
		return false, false, "", fmt.Errorf("failed to check follow-up eligibility: %w", err)
	}

	// Active follow-up found
	daysRemaining := int(time.Until(validUntil).Hours() / 24)
	if isFree {
		return true, true, fmt.Sprintf("Free follow-up available (%d days remaining)", daysRemaining), nil
	}

	return false, true, "Follow-up available (payment required)", nil
}

