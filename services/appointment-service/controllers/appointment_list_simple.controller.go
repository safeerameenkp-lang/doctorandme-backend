package controllers

import (
	"appointment-service/config"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var locIST *time.Location

func init() {
	var err error
	locIST, err = time.LoadLocation("Asia/Kolkata")
	if err != nil {
		// Fallback to Fixed Zone if LoadLocation fails (e.g. missing tzdata)
		locIST = time.FixedZone("IST", 5*3600+30*60)
	}
}

// GetSimpleAppointmentList - Simple appointment list for clinic
// GET /appointments/simple-list?clinic_id=xxx&date=xxx
func GetSimpleAppointmentList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	clinicID := c.Query("clinic_id")
	date := c.Query("date")

	if clinicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clinic_id is required"})
		return
	}

	// Optimized query with specific columns and efficient joins
	query := `
		SELECT 
			a.id,
			a.token_number,
			cp.mo_id,
			cp.phone,
			a.clinic_patient_id,
			COALESCE(cp.first_name || ' ' || cp.last_name, cp.first_name, 'Unknown') as patient_name,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name, 'Unknown Doctor') as doctor_name,
			COALESCE(dept_appt.name, dept_doc.name) as department,
			a.consultation_type,
			a.appointment_date,
			a.appointment_time,
			a.status,
			a.fee_amount,
			a.payment_status,
			a.booking_number,
			a.booking_mode,
			a.created_at
		FROM appointments a
		LEFT JOIN clinic_patients cp ON cp.id = a.clinic_patient_id
		LEFT JOIN doctors d ON d.id = a.doctor_id
		LEFT JOIN users u ON u.id = d.user_id
		LEFT JOIN departments dept_appt ON dept_appt.id = a.department_id
		LEFT JOIN departments dept_doc ON dept_doc.id = d.department_id
		WHERE a.clinic_id = $1
	`

	args := []interface{}{clinicID}
	if date != "" {
		query += " AND a.appointment_date = $2"
		args = append(args, date)
	}

	query += " ORDER BY a.appointment_date DESC, a.appointment_time DESC"

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("ERROR: GetSimpleAppointmentList failed to query database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointments"})
		return
	}
	defer rows.Close()

	// Pre-allocate slice with a reasonable initial capacity to reduce re-allocations
	appointments := make([]AppointmentListItem, 0, 100)

	for rows.Next() {
		var apt AppointmentListItem
		var appointmentDate *string
		var appointmentTime, createdAt time.Time

		err := rows.Scan(
			&apt.ID,
			&apt.TokenNumber,
			&apt.MoID,
			&apt.PatientNumber,
			&apt.ClinicPatientID,
			&apt.PatientName,
			&apt.DoctorName,
			&apt.Department,
			&apt.ConsultationType,
			&appointmentDate,
			&appointmentTime,
			&apt.Status,
			&apt.FeeAmount,
			&apt.PaymentStatus,
			&apt.BookingNumber,
			&apt.BookingMode,
			&createdAt,
		)
		if err != nil {
			log.Printf("ERROR: GetSimpleAppointmentList failed to scan row: %v", err)
			continue
		}

		// Combined formatting logic for better performance
		if appointmentDate != nil {
			apt.AppointmentDateTime = *appointmentDate + " " + appointmentTime.Format("15:04:05")
		} else {
			apt.AppointmentDateTime = appointmentTime.Format("2006-01-02 15:04:05")
		}

		apt.FeeStatus = apt.PaymentStatus
		apt.CreatedAt = createdAt.Format("2006-01-02 15:04:05")

		appointments = append(appointments, apt)
	}

	if err = rows.Err(); err != nil {
		log.Printf("ERROR: GetSimpleAppointmentList rows iteration error: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"clinic_id":    clinicID,
		"date":         date,
		"total":        len(appointments),
		"appointments": appointments,
	})
}

// GetSimpleAppointmentDetails - Get single appointment details
// GET /appointments/simple/:id
// Uses the SAME fields as GetSimpleAppointmentList
func GetSimpleAppointmentDetails(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	appointmentID := c.Param("id")
	if appointmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "appointment_id is required"})
		return
	}

	// Optimized query for single appointment
	query := `
		SELECT 
			a.id,
			a.token_number,
			cp.mo_id,
			cp.phone,
			a.clinic_patient_id,
			COALESCE(cp.first_name || ' ' || cp.last_name, cp.first_name, 'Unknown') as patient_name,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name, 'Unknown Doctor') as doctor_name,
			COALESCE(dept_appt.name, dept_doc.name) as department,
			a.consultation_type,
			a.appointment_date,
			a.appointment_time,
			a.status,
			a.fee_amount,
			a.payment_status,
			a.booking_number,
			a.booking_mode,
			a.created_at
		FROM appointments a
		LEFT JOIN clinic_patients cp ON cp.id = a.clinic_patient_id
		LEFT JOIN doctors d ON d.id = a.doctor_id
		LEFT JOIN users u ON u.id = d.user_id
		LEFT JOIN departments dept_appt ON dept_appt.id = a.department_id
		LEFT JOIN departments dept_doc ON dept_doc.id = d.department_id
		WHERE a.id = $1
	`

	var (
		id, patientName, doctorName, status, bookingNumber, paymentStatus                   string
		tokenNumber                                                                         *string
		moID, patientNumber, clinicPatientID, department, consultationType, appointmentDate *string
		appointmentTime, createdAt                                                          time.Time
		feeAmount                                                                           *float64
		bookingMode                                                                         string
	)

	err := config.DB.QueryRowContext(ctx, query, appointmentID).Scan(
		&id,
		&tokenNumber,
		&moID,
		&patientNumber,
		&clinicPatientID,
		&patientName,
		&doctorName,
		&department,
		&consultationType,
		&appointmentDate,
		&appointmentTime,
		&status,
		&feeAmount,
		&paymentStatus,
		&bookingNumber,
		&bookingMode,
		&createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		} else {
			log.Printf("ERROR: GetSimpleAppointmentDetails failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Format appointment date and time efficiently
	var appointmentDateTime string
	if appointmentDate != nil {
		appointmentDateTime = *appointmentDate + " " + appointmentTime.Format("15:04:05")
	} else {
		appointmentDateTime = appointmentTime.Format("2006-01-02 15:04:05")
	}

	// Build exact response JSON structure
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"appointment": gin.H{
			"id":                    id,
			"token_number":          tokenNumber,
			"mo_id":                 moID,
			"patient_number":        patientNumber,
			"clinic_patient_id":     clinicPatientID,
			"patient_name":          patientName,
			"doctor_name":           doctorName,
			"department":            department,
			"consultation_type":     consultationType,
			"appointment_date_time": appointmentDateTime,
			"status":                status,
			"fee_amount":            feeAmount,
			"payment_status":        paymentStatus,
			"fee_status":            paymentStatus, // Business logic: same as payment_status
			"booking_number":        bookingNumber,
			"booking_mode":          bookingMode,
			"created_at":            createdAt.Format("2006-01-02 15:04:05"),
		},
	})
}

// RescheduleAppointmentDetails - Reschedule appointment with slot selection
// POST /appointments/simple/:id/reschedule
func RescheduleAppointmentDetails(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	appointmentID := c.Param("id")
	if appointmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "appointment_id is required"})
		return
	}

	var input struct {
		DoctorID         string  `json:"doctor_id" binding:"required,uuid"`
		DepartmentID     *string `json:"department_id" binding:"omitempty,uuid"`
		IndividualSlotID string  `json:"individual_slot_id" binding:"required,uuid"`
		ConsultationType *string `json:"consultation_type"`
		AppointmentDate  string  `json:"appointment_date" binding:"required"`
		AppointmentTime  string  `json:"appointment_time" binding:"required"`
		Reason           *string `json:"reason"`
		Notes            *string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Step 1: Merged Validation & Pre-fetch Data
	// Fetches existing appointment, new slot details, and metadata for response in ONE query.
	var (
		existingClinicID, existingDoctorID, existingPatientID                                           string
		existingSlotID                                                                                  sql.NullString
		existingFeeAmount                                                                               *float64
		existingBookingNumber, existingConsultationType                                                 string
		existingTokenNumber                                                                             sql.NullString
		existingStatus, existingPaymentStatus, existingBookingMode, existingMoID, existingPatientNumber sql.NullString
		existingCreatedAt                                                                               time.Time

		slotClinicID, slotStart, slotStatus sql.NullString
		slotAvailableCount                  sql.NullInt64

		patientName, doctorName sql.NullString
		departmentName          sql.NullString
	)

	err := config.DB.QueryRowContext(ctx, `
		SELECT 
			a.clinic_id, a.doctor_id, a.clinic_patient_id, a.individual_slot_id,
			a.fee_amount, a.booking_number, a.consultation_type, a.token_number,
			a.status, a.payment_status, a.booking_mode, cp.mo_id, cp.phone, a.created_at,
			s.clinic_id, s.slot_start, s.status, s.available_count,
			COALESCE(cp.first_name || ' ' || cp.last_name, cp.first_name, 'Unknown'),
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name, 'Unknown Doctor'),
			dept.name
		FROM appointments a
		LEFT JOIN doctor_individual_slots s ON s.id = $2
		LEFT JOIN clinic_patients cp ON cp.id = a.clinic_patient_id
		LEFT JOIN doctors d ON d.id = $3
		LEFT JOIN users u ON u.id = d.user_id
		LEFT JOIN departments dept ON dept.id = $4
		WHERE a.id = $1 AND a.status IN ('scheduled', 'confirmed', 'pending')
	`, appointmentID, input.IndividualSlotID, input.DoctorID, input.DepartmentID).Scan(
		&existingClinicID, &existingDoctorID, &existingPatientID, &existingSlotID,
		&existingFeeAmount, &existingBookingNumber, &existingConsultationType, &existingTokenNumber,
		&existingStatus, &existingPaymentStatus, &existingBookingMode, &existingMoID, &existingPatientNumber, &existingCreatedAt,
		&slotClinicID, &slotStart, &slotStatus, &slotAvailableCount,
		&patientName, &doctorName, &departmentName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found or cannot be rescheduled"})
		} else {
			log.Printf("ERROR: RescheduleAppointmentDetails pre-fetch failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Step 2: In-memory Validation
	if !slotClinicID.Valid {
		c.JSON(http.StatusNotFound, gin.H{"error": "Slot not found"})
		return
	}
	if slotClinicID.String != existingClinicID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slot belongs to different clinic"})
		return
	}

	isSameSlot := existingSlotID.Valid && existingSlotID.String == input.IndividualSlotID
	if !isSameSlot {
		if slotAvailableCount.Int64 <= 0 || slotStatus.String != "available" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Slot not available",
				"message": "This slot is fully booked. Please select another slot.",
			})
			return
		}
	}

	// Past date/time validation using cached IST location
	appointmentDate, err := time.ParseInLocation("2006-01-02", input.AppointmentDate, locIST)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	// Slot date mismatch validation
	if slotTime, err := time.Parse(time.RFC3339, slotStart.String); err == nil {
		slotTimeIST := slotTime.In(locIST)
		if slotTimeIST.Year() > 1 && slotTimeIST.Format("2006-01-02") != input.AppointmentDate {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Slot date mismatch",
				"message": fmt.Sprintf("The selected slot is for %s, but you requested %s", slotTimeIST.Format("2006-01-02"), input.AppointmentDate),
			})
			return
		}
	}

	appointmentTime, err := time.ParseInLocation("2006-01-02 15:04:05", input.AppointmentTime, locIST)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time format. Use YYYY-MM-DD HH:MM:SS"})
		return
	}

	now := time.Now().In(locIST)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, locIST)
	if appointmentDate.Before(today) || (appointmentDate.Equal(today) && appointmentTime.Before(now)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot reschedule to past date or time"})
		return
	}

	// Step 3: Transactional Updates
	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Free old slot
	if existingSlotID.Valid && !isSameSlot {
		_, err = tx.ExecContext(ctx, `
			UPDATE doctor_individual_slots
			SET available_count = available_count + 1, is_booked = false, status = 'available', booked_appointment_id = NULL, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`, existingSlotID.String)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to free old slot"})
			return
		}
	}

	// Update appointment
	newConsultationType := existingConsultationType
	if input.ConsultationType != nil {
		newConsultationType = *input.ConsultationType
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE appointments SET
			doctor_id = $1, department_id = $2, individual_slot_id = $3, appointment_date = $4, appointment_time = $5,
			consultation_type = $6, reason = $7, notes = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9
	`, input.DoctorID, input.DepartmentID, input.IndividualSlotID, input.AppointmentDate, appointmentTime,
		newConsultationType, input.Reason, input.Notes, appointmentID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update appointment"})
		return
	}

	// Book new slot
	if !isSameSlot {
		res, err := tx.ExecContext(ctx, `
			UPDATE doctor_individual_slots
			SET available_count = available_count - 1,
			    is_booked = (available_count - 1 <= 0),
			    status = CASE WHEN available_count - 1 <= 0 THEN 'booked' ELSE status END,
			    booked_appointment_id = CASE WHEN available_count - 1 <= 0 THEN $1 ELSE booked_appointment_id END,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $2 AND available_count > 0 AND status = 'available'
		`, appointmentID, input.IndividualSlotID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to book new slot"})
			return
		}
		if affected, _ := res.RowsAffected(); affected == 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Slot just got booked by another user"})
			return
		}
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes"})
		return
	}

	// Step 4: Final Response (Reconstruct from pre-fetched metadata and input)
	appointmentDateTime := input.AppointmentDate + " " + appointmentTime.Format("15:04:05")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Appointment rescheduled successfully",
		"appointment": gin.H{
			"id":                    appointmentID,
			"token_number":          existingTokenNumber.String,
			"mo_id":                 existingMoID.String,
			"patient_number":        existingPatientNumber.String,
			"clinic_patient_id":     existingPatientID,
			"patient_name":          patientName.String,
			"doctor_name":           doctorName.String,
			"department":            departmentName.String,
			"consultation_type":     newConsultationType,
			"appointment_date_time": appointmentDateTime,
			"status":                existingStatus.String,
			"fee_amount":            existingFeeAmount,
			"payment_status":        existingPaymentStatus.String,
			"fee_status":            existingPaymentStatus.String,
			"booking_number":        existingBookingNumber,
			"booking_mode":          existingBookingMode.String,
			"created_at":            existingCreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}
