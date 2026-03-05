package utils

import (
	"appointment-service/config"
	"appointment-service/models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// CheckDoctorAvailability checks if a doctor is available at a specific time with context
func CheckDoctorAvailability(ctx context.Context, doctorID string, appointmentTime time.Time, durationMinutes *int) (bool, error) {
	// Set default duration
	duration := 12
	if durationMinutes != nil {
		duration = *durationMinutes
	}

	endTime := appointmentTime.Add(time.Duration(duration) * time.Minute)
	appointmentDate := appointmentTime.Format("2006-01-02")
	startTimeOnly := appointmentTime.Format("15:04:05")
	endTimeOnly := endTime.Format("15:04:05")

	// Check for overlapping appointments using index-friendly filters
	var count int
	err := config.DB.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM appointments
        WHERE doctor_id = $1 
		AND appointment_date = $2
        AND status IN ('booked', 'arrived', 'in_consultation', 'confirmed')
        AND (
            (appointment_time <= $3 AND (appointment_time + INTERVAL '1 minute' * duration_minutes) > $3)
            OR (appointment_time < $4 AND (appointment_time + INTERVAL '1 minute' * duration_minutes) >= $4)
        )
    `, doctorID, appointmentDate, startTimeOnly, endTimeOnly).Scan(&count)

	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// CalculateAppointmentFee calculates the fee based on consultation type and doctor rules
func CalculateAppointmentFee(doctor models.DoctorInfo, consultationType string, patientID string) *float64 {
	switch consultationType {
	case "new":
		return doctor.ConsultationFee
	case "followup":
		// Check if this is within the free follow-up period
		var lastAppointment time.Time
		err := config.DB.QueryRow(`
            SELECT MAX(appointment_time) FROM appointments
            WHERE patient_id = $1 AND doctor_id = $2 AND status = 'completed'
        `, patientID, doctor.ID).Scan(&lastAppointment)

		if err == nil && !lastAppointment.IsZero() {
			daysSinceLastVisit := int(time.Since(lastAppointment).Hours() / 24)
			if doctor.FollowUpDays != nil && daysSinceLastVisit <= *doctor.FollowUpDays {
				// Free follow-up
				return nil
			}
		}
		return doctor.FollowUpFee
	case "walkin", "emergency":
		// Emergency and walk-in appointments might have different pricing
		if doctor.ConsultationFee != nil {
			// Add 20% premium for emergency/walk-in
			fee := *doctor.ConsultationFee * 1.2
			return &fee
		}
		return doctor.ConsultationFee
	default:
		return doctor.ConsultationFee
	}
}

// GenerateBookingNumberWithTx generates a unique booking number within a transaction
func GenerateBookingNumberWithTx(tx *sql.Tx, doctorCode *string, clinicCode string, appointmentTime time.Time) (string, error) {
	code := clinicCode
	if doctorCode != nil && *doctorCode != "" {
		code = *doctorCode
	}

	dateStr := appointmentTime.Format("20060102")

	// Get next serial number for the day
	var serialNo int
	err := tx.QueryRow(`
        SELECT COALESCE(MAX(CAST(SUBSTRING(booking_number FROM '.*-.*-(.*)$') AS INTEGER)), 0) + 1
        FROM appointments
        WHERE booking_number LIKE $1
    `, fmt.Sprintf("%s-%s-%%", code, dateStr)).Scan(&serialNo)

	if err != nil {
		return "", err
	}

	bookingNumber := fmt.Sprintf("%s-%s-%04d", code, dateStr, serialNo)
	return bookingNumber, nil
}

// GenerateBookingNumber generates a unique booking number
func GenerateBookingNumber(doctorCode *string, appointmentTime time.Time) (string, error) {
	// Get clinic code from doctor
	var clinicCode string
	err := config.DB.QueryRow(`
        SELECT c.clinic_code FROM clinics c
        JOIN doctors d ON d.clinic_id = c.id
        WHERE d.id = (SELECT id FROM doctors WHERE doctor_code = $1 LIMIT 1)
    `, doctorCode).Scan(&clinicCode)
	if err != nil {
		return "", err
	}

	// Use clinic code if doctor code is not available
	code := clinicCode
	if doctorCode != nil && *doctorCode != "" {
		code = *doctorCode
	}

	dateStr := appointmentTime.Format("20060102")

	// Get next serial number for the day
	var serialNo int
	err = config.DB.QueryRow(`
        SELECT COALESCE(MAX(CAST(SUBSTRING(booking_number FROM '.*-.*-(.*)$') AS INTEGER)), 0) + 1
        FROM appointments
        WHERE booking_number LIKE $1
    `, fmt.Sprintf("%s-%s-%%", code, dateStr)).Scan(&serialNo)
	if err != nil {
		return "", err
	}

	bookingNumber := fmt.Sprintf("%s-%s-%04d", code, dateStr, serialNo)
	return bookingNumber, nil
}

// GenerateTimeSlots generates available time slots for a doctor on a specific date
func GenerateTimeSlots(targetDate time.Time, startTime, endTime time.Time, slotDuration int, doctorID string) []models.TimeSlot {
	var slots []models.TimeSlot

	// Combine date with time
	startDateTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
		startTime.Hour(), startTime.Minute(), startTime.Second(), 0, targetDate.Location())
	endDateTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
		endTime.Hour(), endTime.Minute(), endTime.Second(), 0, targetDate.Location())

	current := startDateTime
	for current.Before(endDateTime) {
		slotEnd := current.Add(time.Duration(slotDuration) * time.Minute)

		// Check if this slot is booked
		var appointmentID *string
		err := config.DB.QueryRow(`
            SELECT id FROM appointments
            WHERE doctor_id = $1 
            AND appointment_time = $2
            AND status NOT IN ('cancelled', 'no_show')
        `, doctorID, current).Scan(&appointmentID)

		isBooked := err == nil && appointmentID != nil

		slot := models.TimeSlot{
			StartTime:     current,
			EndTime:       slotEnd,
			IsAvailable:   !isBooked,
			IsBooked:      isBooked,
			AppointmentID: appointmentID,
		}

		slots = append(slots, slot)
		current = slotEnd
	}

	return slots
}

// GenerateTokenNumber generates and returns the next token number for a doctor on a specific date at a specific clinic
// Token numbers are unique per doctor, per clinic, per department, per day
// Tokens reset to 1 at the start of each new day
// Returns a formatted string like 'RA-01'
func GenerateTokenNumber(doctorID, clinicID string, departmentID *string, appointmentDate time.Time) (string, error) {
	// 1. Get/Generate doctor code for identifier
	doctorCode, err := GetOrGenerateDoctorCode(doctorID)
	if err != nil {
		log.Printf("⚠️ GenerateTokenNumber: failed to get doctor code: %v. Using DR.", err)
		doctorCode = "DR"
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	token, err := GenerateTokenNumberWithTx(tx, doctorID, clinicID, departmentID, doctorCode)
	if err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return token, nil
}

// GenerateTokenNumberWithTx generates the next token number using an existing transaction
func GenerateTokenNumberWithTx(tx *sql.Tx, doctorID, clinicID string, departmentID *string, doctorCode string) (string, error) {
	// GLOBAL TOKENS: We use a fixed date to keep the sequence increasing indefinitely
	dateStr := "0001-01-01"

	// Handle department_id for unique sequence
	dummyUUID := "00000000-0000-0000-0000-000000000000"
	var deptIDInput interface{}
	deptIDInput = nil
	deptIDStr := dummyUUID

	if departmentID != nil && *departmentID != "" && *departmentID != "null" {
		deptIDStr = *departmentID
		deptIDInput = *departmentID
	}

	var serialNumber int

	// Try to get existing token record for this doctor, department, and fixed date
	err := tx.QueryRow(`
        SELECT current_token 
        FROM doctor_tokens 
        WHERE doctor_id = $1 
        AND COALESCE(department_id, '00000000-0000-0000-0000-000000000000') = $2 
        AND token_date = $3
        FOR UPDATE
    `, doctorID, deptIDStr, dateStr).Scan(&serialNumber)

	if err != nil {
		if err == sql.ErrNoRows {
			// No token record exists, create one starting with token 1
			_, err = tx.Exec(`
                INSERT INTO doctor_tokens (doctor_id, clinic_id, department_id, token_date, current_token)
                VALUES ($1, $2, $3, $4, 1)
            `, doctorID, clinicID, deptIDInput, dateStr)

			if err != nil {
				return "", fmt.Errorf("failed to create token record: %v", err)
			}
			serialNumber = 1
		} else {
			return "", fmt.Errorf("failed to query token: %v", err)
		}
	} else {
		// Record exists, increment it
		serialNumber++

		_, err = tx.Exec(`
            UPDATE doctor_tokens 
            SET current_token = $1, clinic_id = $2, updated_at = CURRENT_TIMESTAMP
            WHERE doctor_id = $3 
            AND COALESCE(department_id, '00000000-0000-0000-0000-000000000000') = $4 
            AND token_date = $5
        `, serialNumber, clinicID, doctorID, deptIDStr, dateStr)

		if err != nil {
			return "", fmt.Errorf("failed to update token: %v", err)
		}
	}

	// Format final token (e.g. RA-01)
	formattedToken := fmt.Sprintf("%s-%02d", doctorCode, serialNumber)
	return formattedToken, nil
}

// GetOrGenerateDoctorCode ensures a doctor has a name-based short code
func GetOrGenerateDoctorCode(doctorID string) (string, error) {
	var doctorCode sql.NullString
	var firstName, lastName string
	var departmentName sql.NullString

	err := config.DB.QueryRow(`
		SELECT d.doctor_code, u.first_name, u.last_name, dept.name
		FROM doctors d
		JOIN users u ON d.user_id = u.id
		LEFT JOIN departments dept ON d.department_id = dept.id
		WHERE d.id = $1
	`, doctorID).Scan(&doctorCode, &firstName, &lastName, &departmentName)

	if err != nil {
		return "DR", err
	}

	// If code already exists and is not numeric, use it
	if doctorCode.Valid && doctorCode.String != "" {
		isNumeric := true
		for _, char := range doctorCode.String {
			if (char < '0' || char > '9') && char != '-' {
				isNumeric = false
				break
			}
		}
		if !isNumeric {
			return doctorCode.String, nil
		}
	}

	// Generate initials based on name
	newCode := ""
	f := strings.TrimSpace(strings.ToUpper(firstName))
	l := strings.TrimSpace(strings.ToUpper(lastName))
	d := ""
	if departmentName.Valid {
		d = strings.TrimSpace(strings.ToUpper(departmentName.String))
	}

	// Strategy 1: First Initial + Last Initial (RA)
	if len(f) > 0 && len(l) > 0 {
		newCode = string(f[0]) + string(l[0])
	} else if len(f) >= 2 {
		newCode = f[0:2]
	} else {
		newCode = "DR"
	}

	// Strategy 2: If RA exists and it's a different doctor, try First Initial + Dept Initial (RC)
	if isDoctorCodeTaken(newCode, doctorID) && len(f) > 0 && len(d) > 0 {
		newCode = string(f[0]) + string(d[0])
	}

	// Strategy 3: Default uniqueness handler (RA1, RA2...)
	uniqueCode := ensureUniqueDoctorCode(newCode, doctorID)

	// Update DB for future use
	_, _ = config.DB.Exec(`UPDATE doctors SET doctor_code = $1 WHERE id = $2`, uniqueCode, doctorID)

	return uniqueCode, nil
}

func isDoctorCodeTaken(code, doctorID string) bool {
	var exists bool
	err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM doctors WHERE doctor_code = $1 AND id != $2)`, code, doctorID).Scan(&exists)
	return err == nil && exists
}

func ensureUniqueDoctorCode(baseCode, doctorID string) string {
	code := baseCode
	counter := 1

	for {
		var exists bool
		err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM doctors WHERE doctor_code = $1 AND id != $2)`, code, doctorID).Scan(&exists)
		if err != nil || !exists {
			return code
		}

		// Try taking another letter from name or append a number
		if counter < 10 {
			code = fmt.Sprintf("%s%d", baseCode, counter)
		} else {
			// Random fallback if too many collisions
			code = fmt.Sprintf("%s%d", baseCode, time.Now().Unix()%100)
			return code
		}
		counter++
	}
}

// GetAppointmentReports generates various appointment reports
func GetAppointmentReports(reportType string, startDate, endDate time.Time, doctorID *string) ([]models.AppointmentReport, error) {
	var query string
	var args []interface{}

	switch reportType {
	case "daily_collection":
		query = `
            SELECT 
                appointment_date as date,
                a.doctor_id,
                CONCAT(du.first_name, ' ', du.last_name) as doctor_name,
                COUNT(*) as total_bookings,
                COUNT(CASE WHEN a.status = 'completed' THEN 1 END) as completed,
                COUNT(CASE WHEN a.status = 'no_show' THEN 1 END) as no_show,
                COUNT(CASE WHEN a.status = 'cancelled' THEN 1 END) as cancelled,
                COALESCE(SUM(CASE WHEN a.payment_status = 'paid' THEN a.fee_amount ELSE 0 END), 0) as total_revenue,
                COALESCE(SUM(CASE WHEN a.payment_status = 'paid' AND a.payment_mode = 'cash' THEN a.fee_amount ELSE 0 END), 0) as cash_revenue,
                COALESCE(SUM(CASE WHEN a.payment_status = 'paid' AND a.payment_mode = 'card' THEN a.fee_amount ELSE 0 END), 0) as card_revenue,
                COALESCE(SUM(CASE WHEN a.payment_status = 'paid' AND a.payment_mode = 'upi' THEN a.fee_amount ELSE 0 END), 0) as upi_revenue,
                COALESCE(SUM(CASE WHEN a.payment_status = 'pending' THEN a.fee_amount ELSE 0 END), 0) as pending_revenue
            FROM appointments a
            JOIN doctors d ON d.id = a.doctor_id
            JOIN users du ON du.id = d.user_id
            WHERE a.appointment_date BETWEEN $1 AND $2
        `
		args = []interface{}{startDate, endDate}

		if doctorID != nil {
			query += " AND a.doctor_id = $3"
			args = append(args, *doctorID)
		}

		query += " GROUP BY a.appointment_date, a.doctor_id, du.first_name, du.last_name ORDER BY date DESC"
	}

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []models.AppointmentReport
	for rows.Next() {
		var report models.AppointmentReport
		err := rows.Scan(
			&report.Date, &report.DoctorID, &report.DoctorName,
			&report.TotalBookings, &report.Completed, &report.NoShow,
			&report.Cancelled, &report.TotalRevenue, &report.CashRevenue,
			&report.CardRevenue, &report.UPIRevenue, &report.PendingRevenue,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}
