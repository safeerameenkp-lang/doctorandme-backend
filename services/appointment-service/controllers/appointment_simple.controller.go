package controllers

import (
	"appointment-service/config"
	"appointment-service/models"
	"appointment-service/utils"
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Global follow-up manager instance
var followUpManager *utils.FollowUpManager

func init() {
	// Initialize follow-up manager after DB is ready
	// This will be set properly when the server starts
}

// =====================================================
// SIMPLIFIED APPOINTMENT BOOKING
// For clinic-specific patients only
// =====================================================

// SimpleAppointmentInput - Simplified input with only clinic_patient_id
type SimpleAppointmentInput struct {
	ClinicPatientID  string  `json:"clinic_patient_id" binding:"required,uuid"`
	DoctorID         string  `json:"doctor_id" binding:"required,uuid"`
	ClinicID         string  `json:"clinic_id" binding:"required,uuid"`
	DepartmentID     *string `json:"department_id"`
	IndividualSlotID *string `json:"individual_slot_id"`
	AppointmentDate  string  `json:"appointment_date" binding:"required"`
	AppointmentTime  string  `json:"appointment_time" binding:"required"`
	ConsultationType string  `json:"consultation_type" binding:"required,oneof=clinic_visit video_consultation follow-up-via-clinic follow-up-via-video"`
	IsFollowUp       bool    `json:"is_follow_up"` // Auto-set based on consultation_type
	Reason           *string `json:"reason"`
	Notes            *string `json:"notes"`
	PaymentMethod    *string `json:"payment_method" binding:"omitempty,oneof=pay_now pay_later way_off"` // Optional for follow-ups
	PaymentType      *string `json:"payment_type" binding:"omitempty,oneof=cash card upi"`
	BookingMode      *string `json:"booking_mode" binding:"omitempty,oneof=slot walk_in"`
}

// RescheduleSimpleAppointmentInput - Input for rescheduling simple appointments based on UI
type RescheduleSimpleAppointmentInput struct {
	DepartmentID     *string `json:"department_id"`
	DoctorID         string  `json:"doctor_id" binding:"required,uuid"`
	ClinicID         string  `json:"clinic_id" binding:"required,uuid"`
	IndividualSlotID *string `json:"individual_slot_id"`
	AppointmentDate  string  `json:"appointment_date" binding:"required"`
	AppointmentTime  string  `json:"appointment_time" binding:"required"`
	Reason           *string `json:"reason"`
	Notes            *string `json:"notes"`
}

// CreateSimpleAppointment - Simplified appointment creation
// POST /appointments/simple
func CreateSimpleAppointment(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	var input SimpleAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	if input.ConsultationType == "follow-up-via-clinic" || input.ConsultationType == "follow-up-via-video" {
		input.IsFollowUp = true
	}

	// Step 1: Parse dates early to fail fast
	appointmentDate, err := time.Parse("2006-01-02", input.AppointmentDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}
	appointmentTime, err := time.Parse("2006-01-02 15:04:05", input.AppointmentTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time format. Use YYYY-MM-DD HH:MM:SS"})
		return
	}

	// Step 2: Merged Validation & Data Fetching (Reduces 5+ queries into 1)
	var (
		patientID, patientClinicID, patientName                           sql.NullString
		doctorID, doctorCode, doctorFirst, doctorLast                     sql.NullString
		consultFee, followupFee                                           *float64
		clinicCode, deptName                                              sql.NullString
		slotClinicID, slotStatus                                          sql.NullString
		slotAvailableCount                                                sql.NullInt64
		activeFollowupID, activeFollowupStatus, activeFollowupLogicStatus sql.NullString
		activeFollowupIsFree                                              sql.NullBool
		activeFollowupValidFrom, activeFollowupValidUntil                 *time.Time
		activeFollowupSourceID, activeFollowupRenewedBy                   sql.NullString
		activeFollowupCreatedAt, activeFollowupUpdatedAt                  *time.Time
		hasPreviousAppointment                                            bool
	)

	err = config.DB.QueryRowContext(ctx, `
		SELECT 
			p.id, p.clinic_id, p.first_name || ' ' || p.last_name as p_name,
			d.id, d.doctor_code, u.first_name, u.last_name,
			COALESCE(cdl.consultation_fee_offline, d.consultation_fee),
			COALESCE(cdl.follow_up_fee, d.follow_up_fee),
			c.clinic_code, dept.name,
			s.clinic_id, s.status, s.available_count,
			f.id, f.status, f.is_free, f.valid_from, f.valid_until, f.source_appointment_id, f.renewed_by_appointment_id,
            f.follow_up_logic_status, f.created_at, f.updated_at,
			EXISTS(SELECT 1 FROM appointments WHERE clinic_patient_id = $1 AND doctor_id = $2 AND status IN ('completed', 'confirmed')) as has_prev
		FROM (SELECT 1) dummy
		LEFT JOIN clinic_patients p ON p.id = $1 AND p.is_active = true
		LEFT JOIN doctors d ON d.id = $2 AND d.is_active = true
		LEFT JOIN users u ON u.id = d.user_id
		LEFT JOIN clinics c ON c.id = d.clinic_id
		LEFT JOIN clinic_doctor_links cdl ON cdl.doctor_id = d.id AND cdl.clinic_id = $3
		LEFT JOIN departments dept ON dept.id = $4
		LEFT JOIN doctor_individual_slots s ON s.id = $5
		LEFT JOIN (
			SELECT * FROM follow_ups 
			WHERE clinic_patient_id = $1 AND doctor_id = $2 
			AND (department_id = $4 OR (department_id IS NULL AND $4 IS NULL))
			AND status = 'active' AND valid_until >= CURRENT_DATE
			ORDER BY created_at DESC LIMIT 1
		) f ON true
	`, input.ClinicPatientID, input.DoctorID, input.ClinicID, input.DepartmentID, input.IndividualSlotID).Scan(
		&patientID, &patientClinicID, &patientName,
		&doctorID, &doctorCode, &doctorFirst, &doctorLast,
		&consultFee, &followupFee,
		&clinicCode, &deptName,
		&slotClinicID, &slotStatus, &slotAvailableCount,
		&activeFollowupID, &activeFollowupStatus, &activeFollowupIsFree, &activeFollowupValidFrom, &activeFollowupValidUntil, &activeFollowupSourceID, &activeFollowupRenewedBy,
		&activeFollowupLogicStatus, &activeFollowupCreatedAt, &activeFollowupUpdatedAt,
		&hasPreviousAppointment,
	)

	if err != nil || !patientID.Valid {
		c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found or inactive"})
		return
	}
	if !doctorID.Valid {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found or inactive"})
		return
	}
	if patientClinicID.String != input.ClinicID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Patient belongs to different clinic"})
		return
	}

	// Step 3: Business Logic Validation
	isFreeFollowUp := false
	if input.IsFollowUp {
		if activeFollowupID.Valid && activeFollowupIsFree.Bool {
			isFreeFollowUp = true
		} else if !hasPreviousAppointment {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not eligible for follow-up", "message": "No previous appointment found with this doctor"})
			return
		}
	}

	// Payment validation
	if !input.IsFollowUp || (input.IsFollowUp && !isFreeFollowUp) {
		if input.PaymentMethod == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payment method required", "message": map[bool]string{true: "This follow-up requires payment", false: "Please specify payment_method"}[input.IsFollowUp]})
			return
		}
		if *input.PaymentMethod == "pay_now" && (input.PaymentType == nil || *input.PaymentType == "") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payment type required", "message": "You must provide payment_type (cash, card, or upi)"})
			return
		}
	}

	// Slot validation
	bookingMode := "slot"
	if input.BookingMode != nil {
		bookingMode = *input.BookingMode
	}
	if bookingMode == "walk_in" {
		if input.IndividualSlotID != nil && *input.IndividualSlotID != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking mode", "message": "individual_slot_id must be null for walk_in mode"})
			return
		}
	} else {
		if !slotStatus.Valid {
			c.JSON(http.StatusNotFound, gin.H{"error": "Slot not found"})
			return
		}
		if slotClinicID.Valid && slotClinicID.String != input.ClinicID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Slot belongs to different clinic"})
			return
		}
		if slotAvailableCount.Int64 <= 0 || slotStatus.String != "available" {
			c.JSON(http.StatusConflict, gin.H{"error": "Slot not available", "message": "This slot is fully booked."})
			return
		}
	}

	// Ensure doctor has a valid code (using names fetched if needed)
	validDoctorCode, _ := utils.GetOrGenerateDoctorCode(doctorID.String)
	doctorCode.String = validDoctorCode

	// Fee calculation
	feeAmount := 0.0
	if input.IsFollowUp && isFreeFollowUp {
		feeAmount = 0.0
	} else if input.ConsultationType == "follow_up" && followupFee != nil {
		feeAmount = *followupFee
	} else if consultFee != nil {
		feeAmount = *consultFee
	}

	// Step 4: Transactional Updates
	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Generate identifiers within tx
	bookingNumber, _ := utils.GenerateBookingNumberWithTx(tx, &doctorCode.String, clinicCode.String, appointmentTime)
	tokenNumber, _ := utils.GenerateTokenNumberWithTx(tx, doctorID.String, input.ClinicID, input.DepartmentID, doctorCode.String)

	paymentStatus := "pending"
	var paymentMode *string
	if input.IsFollowUp && isFreeFollowUp {
		paymentStatus = "waived"
	} else if input.PaymentMethod != nil {
		switch *input.PaymentMethod {
		case "pay_now":
			paymentStatus = "paid"
			paymentMode = input.PaymentType
		case "pay_later":
			paymentStatus = "pending"
		case "way_off":
			paymentStatus = "waived"
		}
	}

	var appointment models.Appointment
	err = tx.QueryRowContext(ctx, `
		INSERT INTO appointments (
			clinic_patient_id, clinic_id, doctor_id, department_id, booking_number, token_number,
			appointment_date, appointment_time, duration_minutes, consultation_type,
			reason, notes, fee_amount, payment_mode, payment_status, status, individual_slot_id, booking_mode
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 5, $9, $10, $11, $12, $13, $14, 'confirmed', $15, $16)
		RETURNING id, clinic_patient_id, clinic_id, doctor_id, booking_number, token_number,
		          appointment_date, appointment_time, duration_minutes, consultation_type,
		          reason, notes, status, fee_amount, payment_status, payment_mode, booking_mode, created_at
	`, input.ClinicPatientID, input.ClinicID, input.DoctorID, input.DepartmentID, bookingNumber, tokenNumber,
		appointmentDate.Format("2006-01-02"), appointmentTime, input.ConsultationType,
		input.Reason, input.Notes, feeAmount, paymentMode, paymentStatus, input.IndividualSlotID, bookingMode).Scan(
		&appointment.ID, &appointment.ClinicPatientID, &appointment.ClinicID, &appointment.DoctorID,
		&appointment.BookingNumber, &appointment.TokenNumber, &appointment.AppointmentDate,
		&appointment.AppointmentTime, &appointment.DurationMinutes, &appointment.ConsultationType,
		&appointment.Reason, &appointment.Notes, &appointment.Status, &appointment.FeeAmount,
		&appointment.PaymentStatus, &appointment.PaymentMode, &appointment.BookingMode, &appointment.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create appointment", "details": err.Error()})
		return
	}

	// Update Slot
	if input.IndividualSlotID != nil && *input.IndividualSlotID != "" {
		_, err = tx.ExecContext(ctx, `
			UPDATE doctor_individual_slots
			SET available_count = available_count - 1,
			    is_booked = (available_count - 1 <= 0),
			    status = CASE WHEN available_count - 1 <= 0 THEN 'booked' ELSE status END,
			    booked_appointment_id = CASE WHEN available_count - 1 <= 0 THEN $1 ELSE booked_appointment_id END,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $2 AND available_count > 0 AND status = 'available'
		`, appointment.ID, *input.IndividualSlotID)
	}

	// Handle Follow-up Tracking
	var followUpID *string
	newPatientFollowupStatus := ""

	if input.ConsultationType == "clinic_visit" || input.ConsultationType == "video_consultation" {
		// New follow-up eligibility
		newPatientFollowupStatus = "active"
		// Mark existing ones as renewed
		tx.ExecContext(ctx, `
			UPDATE follow_ups SET status = 'renewed', renewed_at = CURRENT_TIMESTAMP, renewed_by_appointment_id = $1, follow_up_logic_status = 'renewed', updated_at = CURRENT_TIMESTAMP
			WHERE clinic_patient_id = $2 AND clinic_id = $3 AND doctor_id = $4 AND status IN ('active', 'expired')
			AND (department_id = $5 OR (department_id IS NULL AND $5 IS NULL))
		`, appointment.ID, input.ClinicPatientID, input.ClinicID, input.DoctorID, input.DepartmentID)

		// Insert new follow-up
		err = tx.QueryRowContext(ctx, `
			INSERT INTO follow_ups (
				clinic_patient_id, clinic_id, doctor_id, department_id, source_appointment_id, status, is_free, valid_from, valid_until, follow_up_logic_status, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, 'active', true, $6, $7, 'new', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			RETURNING id
		`, input.ClinicPatientID, input.ClinicID, input.DoctorID, input.DepartmentID, appointment.ID, appointmentDate, appointmentDate.AddDate(0, 0, 5)).Scan(&followUpID)

		// Update patient info
		tx.ExecContext(ctx, `
			UPDATE clinic_patients SET current_followup_status = $1, last_appointment_id = $2, last_followup_id = $3, updated_at = CURRENT_TIMESTAMP
			WHERE id = $4
		`, "active", appointment.ID, followUpID, input.ClinicPatientID)
	} else if input.IsFollowUp && isFreeFollowUp {
		// Use free follow-up
		tx.ExecContext(ctx, `
			UPDATE follow_ups SET status = 'used', used_at = CURRENT_TIMESTAMP, used_appointment_id = $1, follow_up_logic_status = 'used', updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`, appointment.ID, activeFollowupID.String)

		tx.ExecContext(ctx, `
			UPDATE clinic_patients SET current_followup_status = 'used', last_appointment_id = $1, last_followup_id = $2, updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
		`, appointment.ID, activeFollowupID.String, input.ClinicPatientID)

		followUpID = &activeFollowupID.String
		newPatientFollowupStatus = "used"
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Step 5: Build Response
	response := gin.H{
		"message":     "Appointment created successfully",
		"appointment": appointment,
	}

	// Efficiently build follow-up response if needed
	if newPatientFollowupStatus == "active" && followUpID != nil {
		response["follow_up"] = gin.H{
			"id": *followUpID, "patient_name": patientName.String, "doctor_name": "Dr. " + doctorFirst.String + " " + doctorLast.String,
			"department_name": deptName.String, "is_free": true, "valid_until": appointmentDate.AddDate(0, 0, 5).Format(time.RFC3339),
			"days_remaining": 5, "status": "active",
		}
		response["followup_granted"] = true
		response["followup_message"] = "Free follow-up eligibility granted (valid for 5 days)"
	} else if input.IsFollowUp {
		response["is_free_followup"] = isFreeFollowUp
		response["follow_up_info"] = gin.H{
			"is_followup": true, "is_free": isFreeFollowUp, "follow_up_status": "used",
			"message": map[bool]string{true: "This is a FREE follow-up", false: "This is a PAID follow-up"}[isFreeFollowUp],
		}
	}

	c.JSON(http.StatusCreated, response)
}

// RescheduleSimpleAppointment - Reschedule an existing appointment
// POST /appointments/:id/reschedule-simple
func RescheduleSimpleAppointment(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	appointmentID := c.Param("id")

	var input RescheduleSimpleAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	if input.IndividualSlotID == nil || *input.IndividualSlotID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "individual_slot_id is required for rescheduling simple appointments"})
		return
	}

	// Step 1: Consolidated Pre-fetch Query
	// Fetches existing appointment, patient clinic validation, and new slot status in ONE optimized query.
	var (
		existing                                                                         models.Appointment
		existingClinicPatientID, existingConsultationType                                string
		existingFeeAmount                                                                *float64
		existingPaymentStatus, existingPaymentMode, existingToken, existingBookingNumber sql.NullString

		clinicPatientClinicID sql.NullString

		slotClinicID, slotStart, slotStatus sql.NullString
		slotMaxPatients, slotAvailableCount sql.NullInt64
	)

	err := config.DB.QueryRowContext(ctx, `
		SELECT 
			a.id, a.clinic_patient_id, a.clinic_id, a.doctor_id, a.department_id, 
			a.consultation_type, a.fee_amount, a.payment_status, a.payment_mode,
			a.appointment_date, a.appointment_time, a.individual_slot_id, a.token_number, a.booking_number,
			cp.clinic_id as patient_clinic_id,
			s.clinic_id as slot_clinic_id, s.slot_start, s.status, s.max_patients, s.available_count
		FROM appointments a
		LEFT JOIN clinic_patients cp ON cp.id = a.clinic_patient_id AND cp.is_active = true
		LEFT JOIN doctor_individual_slots s ON s.id = $2
		WHERE a.id = $1 AND a.status IN ('confirmed', 'pending')
	`, appointmentID, *input.IndividualSlotID).Scan(
		&existing.ID, &existingClinicPatientID, &existing.ClinicID, &existing.DoctorID, &existing.DepartmentID,
		&existingConsultationType, &existingFeeAmount, &existingPaymentStatus, &existingPaymentMode,
		&existing.AppointmentDate, &existing.AppointmentTime, &existing.IndividualSlotID, &existingToken, &existingBookingNumber,
		&clinicPatientClinicID,
		&slotClinicID, &slotStart, &slotStatus, &slotMaxPatients, &slotAvailableCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found or cannot be rescheduled"})
		} else {
			log.Printf("ERROR: RescheduleSimpleAppointment pre-fetch failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}
	existing.BookingNumber = existingBookingNumber.String
	if existingToken.Valid {
		existing.TokenNumber = &existingToken.String
	}

	// Step 2: In-memory Validations
	if !clinicPatientClinicID.Valid || clinicPatientClinicID.String != input.ClinicID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Patient belongs to different clinic or not found"})
		return
	}

	if !slotClinicID.Valid || slotClinicID.String != input.ClinicID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slot belongs to different clinic or not found"})
		return
	}

	if slotAvailableCount.Int64 <= 0 || slotStatus.String != "available" {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Slot not available",
			"message": "This slot is fully booked. Please select another slot.",
		})
		return
	}

	// Date & Time parsing using package-cached location
	appointmentDate, err := time.ParseInLocation("2006-01-02", input.AppointmentDate, locIST)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
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

	// Step 3: Handle Doctor Change & Fees
	var newDoctor models.DoctorInfo
	var clinicCode string
	newFeeAmount := *existingFeeAmount

	if input.DoctorID != existing.DoctorID {
		err = config.DB.QueryRowContext(ctx, `
			SELECT d.id, d.doctor_code, u.first_name, u.last_name,
			       COALESCE(cdl.consultation_fee_offline, d.consultation_fee),
			       COALESCE(cdl.follow_up_fee, d.follow_up_fee),
				   c.clinic_code
			FROM doctors d
			JOIN users u ON u.id = d.user_id
			JOIN clinics c ON c.id = $1
			LEFT JOIN clinic_doctor_links cdl ON cdl.doctor_id = d.id AND cdl.clinic_id = $1
			WHERE d.id = $2 AND d.is_active = true
		`, input.ClinicID, input.DoctorID).Scan(
			&newDoctor.ID, &newDoctor.DoctorCode, &newDoctor.FirstName, &newDoctor.LastName,
			&newDoctor.ConsultationFee, &newDoctor.FollowUpFee, &clinicCode,
		)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "New doctor not found or inactive"})
			return
		}

		if existingConsultationType == "follow_up" && newDoctor.FollowUpFee != nil {
			newFeeAmount = *newDoctor.FollowUpFee
		} else if newDoctor.ConsultationFee != nil {
			newFeeAmount = *newDoctor.ConsultationFee
		}

		validCode, _ := utils.GetOrGenerateDoctorCode(newDoctor.ID)
		newDoctor.DoctorCode = &validCode
	}

	// Step 4: Transactional Execution
	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Handle sequence generation inside transaction if doctor/date changed
	var bookingNumber, tokenNumber string
	if input.DoctorID != existing.DoctorID || input.AppointmentDate != *existing.AppointmentDate {
		// New clinic code fetch if not already done
		if clinicCode == "" {
			err = tx.QueryRowContext(ctx, "SELECT clinic_code FROM clinics WHERE id = $1", input.ClinicID).Scan(&clinicCode)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clinic metadata"})
				return
			}
			validCode, _ := utils.GetOrGenerateDoctorCode(existing.DoctorID)
			newDoctor.DoctorCode = &validCode
		}

		bookingNumber, err = utils.GenerateBookingNumberWithTx(tx, newDoctor.DoctorCode, clinicCode, appointmentTime)
		if err != nil {
			bookingNumber = "BN" + time.Now().Format("20060102150405")
		}

		tokenNumber, err = utils.GenerateTokenNumberWithTx(tx, input.DoctorID, input.ClinicID, input.DepartmentID, *newDoctor.DoctorCode)
		if err != nil {
			tokenNumber = "T01"
		}
	} else {
		bookingNumber = existing.BookingNumber
		if existing.TokenNumber != nil {
			tokenNumber = *existing.TokenNumber
		} else {
			tokenNumber = "T01"
		}
	}

	// Free old slot
	if existing.IndividualSlotID != nil {
		_, err = tx.ExecContext(ctx, `
			UPDATE doctor_individual_slots
			SET available_count = LEAST(available_count + 1, max_patients),
			    is_booked = CASE WHEN (available_count + 1) >= max_patients THEN false ELSE is_booked END,
			    status = CASE WHEN (available_count + 1) >= max_patients THEN 'available' ELSE status END,
			    booked_appointment_id = CASE WHEN (available_count + 1) >= max_patients THEN NULL ELSE booked_appointment_id END,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`, *existing.IndividualSlotID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to release previous slot"})
			return
		}
	}

	// Update appointment
	_, err = tx.ExecContext(ctx, `
		UPDATE appointments SET
			doctor_id = $1, department_id = $2, individual_slot_id = $3, appointment_date = $4, appointment_time = $5,
			booking_number = $6, token_number = $7, fee_amount = $8, reason = $9, notes = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $11
	`, input.DoctorID, input.DepartmentID, input.IndividualSlotID, input.AppointmentDate, appointmentTime,
		bookingNumber, tokenNumber, newFeeAmount, input.Reason, input.Notes, appointmentID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update appointment record"})
		return
	}

	// Book new slot
	res, err := tx.ExecContext(ctx, `
		UPDATE doctor_individual_slots
		SET available_count = available_count - 1,
		    is_booked = (available_count - 1 <= 0),
		    status = CASE WHEN available_count - 1 <= 0 THEN 'booked' ELSE status END,
		    booked_appointment_id = CASE WHEN available_count - 1 <= 0 THEN $1 ELSE booked_appointment_id END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND available_count > 0 AND status = 'available'
	`, appointmentID, input.IndividualSlotID)

	if err != nil || func() bool { a, _ := res.RowsAffected(); return a == 0 }() {
		c.JSON(http.StatusConflict, gin.H{"error": "Target slot just got fully booked"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize reschedule"})
		return
	}

	// Final Response - Minimal Fetch
	var updated models.Appointment
	config.DB.QueryRowContext(ctx, `
		SELECT id, clinic_patient_id, clinic_id, doctor_id, department_id, booking_number, token_number,
		       appointment_date, appointment_time, duration_minutes, consultation_type,
		       reason, notes, status, fee_amount, payment_status, payment_mode, created_at
		FROM appointments WHERE id = $1
	`, appointmentID).Scan(
		&updated.ID, &updated.ClinicPatientID, &updated.ClinicID, &updated.DoctorID, &updated.DepartmentID,
		&updated.BookingNumber, &updated.TokenNumber, &updated.AppointmentDate, &updated.AppointmentTime,
		&updated.DurationMinutes, &updated.ConsultationType, &updated.Reason, &updated.Notes, &updated.Status,
		&updated.FeeAmount, &updated.PaymentStatus, &updated.PaymentMode, &updated.CreatedAt,
	)

	response := gin.H{
		"message":     "Appointment rescheduled successfully",
		"appointment": updated,
	}
	if existing.IndividualSlotID != nil {
		response["slot_re_enabled"] = gin.H{
			"old_slot_id": *existing.IndividualSlotID,
			"message":     "Previous slot has been made available again",
		}
	}
	c.JSON(http.StatusOK, response)
}
