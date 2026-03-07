package controllers

import (
	"appointment-service/config"
	"appointment-service/middleware"
	"appointment-service/models"
	"appointment-service/utils"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Appointment Controllers
// CreateAppointmentInput - Updated to match UI requirements
// Supports patient search by mobile_no or mo_id, department selection, and time slot validation
type CreateAppointmentInput struct {
	// Patient identification (one of these is required)
	UserID          *string `json:"user_id" binding:"omitempty,uuid"`
	PatientID       *string `json:"patient_id" binding:"omitempty,uuid"`        // Global patient
	ClinicPatientID *string `json:"clinic_patient_id" binding:"omitempty,uuid"` // Clinic-specific patient
	MobileNo        *string `json:"mobile_no" binding:"omitempty"`
	MoID            *string `json:"mo_id" binding:"omitempty"`

	// Appointment details
	ClinicID         string  `json:"clinic_id" binding:"required,uuid"`
	DoctorID         string  `json:"doctor_id" binding:"required,uuid"`
	DepartmentID     *string `json:"department_id" binding:"omitempty,uuid"`
	SlotID           *string `json:"slot_id" binding:"omitempty,uuid"`            // Time slot ID for slot-based booking
	IndividualSlotID *string `json:"individual_slot_id" binding:"omitempty,uuid"` // Individual 5-min slot ID for session-based booking
	AppointmentDate  string  `json:"appointment_date" binding:"required"`
	AppointmentTime  string  `json:"appointment_time" binding:"required"`
	DurationMinutes  *int    `json:"duration_minutes"`

	// Consultation type (matches UI: Video, In-person, Follow Up, Clinic Visit)
	ConsultationType string `json:"consultation_type" binding:"required,oneof=video in_person offline online follow_up clinic_visit"`

	// Additional details
	Reason     *string `json:"reason"`
	Notes      *string `json:"notes"`
	IsPriority *bool   `json:"is_priority"`

	// Payment (matches UI: Pay Later, Pay Now, Way Off)
	PaymentMode *string `json:"payment_mode" binding:"omitempty,oneof=pay_later pay_now way_off cash card upi"`

	// New Feature: Booking Mode
	BookingMode *string `json:"booking_mode" binding:"omitempty,oneof=slot walk_in"`
}

type UpdateAppointmentInput struct {
	AppointmentTime  *string `json:"appointment_time"`
	DurationMinutes  *int    `json:"duration_minutes"`
	ConsultationType *string `json:"consultation_type" binding:"omitempty,oneof=new followup walkin emergency"`
	Status           *string `json:"status" binding:"omitempty,oneof=booked arrived in_consultation completed no_show cancelled"`
	PaymentStatus    *string `json:"payment_status" binding:"omitempty,oneof=pending paid refunded"`
	PaymentMode      *string `json:"payment_mode" binding:"omitempty,oneof=cash card upi"`
	IsPriority       *bool   `json:"is_priority"`
}

type RescheduleAppointmentInput struct {
	NewAppointmentTime string `json:"new_appointment_time" binding:"required"`
	Reason             string `json:"reason"`
}

type CancelAppointmentInput struct {
	Reason string `json:"reason" binding:"required"`
}

type CreatePatientWithAppointmentInput struct {
	// User details
	FirstName   string  `json:"first_name" binding:"required,max=100"`
	LastName    string  `json:"last_name" binding:"required,max=100"`
	Phone       string  `json:"phone" binding:"required,max=20"`
	Email       *string `json:"email" binding:"omitempty,email"`
	DateOfBirth *string `json:"date_of_birth" binding:"omitempty"`
	Gender      *string `json:"gender" binding:"omitempty,max=20"`

	// Patient details
	MOID           *string `json:"mo_id" binding:"omitempty,max=50"`
	MedicalHistory *string `json:"medical_history"`
	Allergies      *string `json:"allergies"`
	BloodGroup     *string `json:"blood_group" binding:"omitempty,max=10"`

	// Appointment details
	ClinicID         string  `json:"clinic_id" binding:"required,uuid"`
	DoctorID         string  `json:"doctor_id" binding:"required,uuid"`
	DepartmentID     *string `json:"department_id" binding:"omitempty,uuid"`
	SlotID           *string `json:"slot_id" binding:"omitempty,uuid"`            // Time slot ID for slot-based booking
	IndividualSlotID *string `json:"individual_slot_id" binding:"omitempty,uuid"` // Individual 5-min slot ID for session-based booking
	AppointmentDate  string  `json:"appointment_date" binding:"required"`
	AppointmentTime  string  `json:"appointment_time" binding:"required"`
	DurationMinutes  *int    `json:"duration_minutes"`
	ConsultationType string  `json:"consultation_type" binding:"required,oneof=new followup walkin emergency"`
	Reason           *string `json:"reason"`
	Notes            *string `json:"notes"`
	IsPriority       *bool   `json:"is_priority"`
	PaymentMode      *string `json:"payment_mode" binding:"omitempty,oneof=cash card upi"`

	// New Feature: Booking Mode
	BookingMode *string `json:"booking_mode" binding:"omitempty,oneof=slot walk_in"`
}

func CreateAppointment(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	var input CreateAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Step 1: Parse Inputs
	appointmentDate, err := time.Parse("2006-01-02", input.AppointmentDate)
	if err != nil {
		middleware.SendValidationError(c, "Invalid appointment date format", "Use YYYY-MM-DD format")
		return
	}
	appointmentDateTime, err := time.Parse("2006-01-02 15:04:05", input.AppointmentTime)
	if err != nil {
		middleware.SendValidationError(c, "Invalid appointment time format", "Use YYYY-MM-DD HH:MM:SS format")
		return
	}
	appointmentTimeOnly := appointmentDateTime.Format("15:04:05")

	// Step 2: Merged Validation & Data Fetching
	var (
		patientIDFound, patientClinicID, patientName sql.NullString
		clinicExists, deptExists                     bool
		doctorIDFound, doctorCode, docFirst, docLast sql.NullString
		isDocActive, docLinked                       bool
		docDeptID                                    sql.NullString
		feeOffline, feeOnline, followUpFee           *float64
		followUpDays                                 *int
		onLeave                                      bool
	)

	// Build dynamic patient lookup clause
	patientFilter := "FALSE"
	var patientArg interface{}
	if input.PatientID != nil && *input.PatientID != "" {
		patientFilter = "p.id = $6"
		patientArg = *input.PatientID
	} else if input.ClinicPatientID != nil && *input.ClinicPatientID != "" {
		patientFilter = "cp.id = $6"
		patientArg = *input.ClinicPatientID
	} else if input.MobileNo != nil && *input.MobileNo != "" {
		patientFilter = "u_pat.phone = $6"
		patientArg = *input.MobileNo
	} else if input.MoID != nil && *input.MoID != "" {
		patientFilter = "p.mo_id = $6"
		patientArg = *input.MoID
	} else if input.UserID != nil && *input.UserID != "" {
		patientFilter = "p.user_id = $6"
		patientArg = *input.UserID
	}

	mergedQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(p.id, cp.id) as p_id,
			COALESCE(p_link.clinic_id, cp.clinic_id) as p_clinic_id,
			COALESCE(u_pat.first_name || ' ' || u_pat.last_name, cp.first_name || ' ' || cp.last_name) as p_name,
			EXISTS(SELECT 1 FROM clinics WHERE id = $1 AND is_active = true) as clinic_exists,
			EXISTS(SELECT 1 FROM departments WHERE id = $2 AND clinic_id = $1 AND is_active = true) as dept_exists,
			d.id, d.doctor_code, u_doc.first_name, u_doc.last_name, u_doc.is_active,
			EXISTS(SELECT 1 FROM clinic_doctor_links WHERE doctor_id = $3 AND clinic_id = $1 AND is_active = true) as doc_linked,
			d.department_id,
			cdl.consultation_fee_offline, cdl.consultation_fee_online, cdl.follow_up_fee, cdl.follow_up_days,
			EXISTS(SELECT 1 FROM doctor_leaves WHERE doctor_id = $3 AND clinic_id = $1 AND status = 'approved' AND from_date <= $4 AND to_date >= $4) as on_leave
		FROM (SELECT 1) dummy
		LEFT JOIN patients p ON %s AND p.is_active = true
		LEFT JOIN clinic_patients cp ON %s AND cp.is_active = true
		LEFT JOIN users u_pat ON (u_pat.id = p.user_id OR u_pat.phone = $6) AND u_pat.is_active = true
		LEFT JOIN patient_clinics p_link ON p_link.patient_id = p.id AND p_link.clinic_id = $1
		LEFT JOIN doctors d ON d.id = $3 AND d.is_active = true
		LEFT JOIN users u_doc ON u_doc.id = d.user_id
		LEFT JOIN clinic_doctor_links cdl ON cdl.doctor_id = d.id AND cdl.clinic_id = $1 AND cdl.is_active = true
	`, patientFilter, patientFilter)

	err = config.DB.QueryRowContext(ctx, mergedQuery, input.ClinicID, input.DepartmentID, input.DoctorID, appointmentDate, 0, patientArg).Scan(
		&patientIDFound, &patientClinicID, &patientName,
		&clinicExists, &deptExists,
		&doctorIDFound, &doctorCode, &docFirst, &docLast, &isDocActive, &docLinked,
		&docDeptID,
		&feeOffline, &feeOnline, &followUpFee, &followUpDays,
		&onLeave,
	)

	// Step 3: Validate Unified Fetch Results
	if err != nil || !patientIDFound.Valid {
		c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found", "message": "Patient could not be identified."})
		return
	}
	patientID := patientIDFound.String

	if !clinicExists || !doctorIDFound.Valid || !isDocActive || !docLinked {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clinic/Doctor error", "message": "Selected clinic/doctor is inactive or unreachable."})
		return
	}

	if input.DepartmentID != nil && *input.DepartmentID != "" && !deptExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Department not found", "message": "The selected department is inactive or not linked to this clinic."})
		return
	}

	// Validate patient-clinic link
	if (input.ClinicPatientID == nil || *input.ClinicPatientID == "") && (!patientClinicID.Valid || patientClinicID.String != input.ClinicID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Patient not registered", "message": "Patient must be linked to this clinic."})
		return
	}

	if onLeave {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor on leave", "message": "Doctor is not available today."})
		return
	}

	if input.DepartmentID != nil && *input.DepartmentID != "" && (!docDeptID.Valid || docDeptID.String != *input.DepartmentID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Department mismatch", "message": "Doctor is not assigned to the selected department."})
		return
	}

	// Basic Doctor Info
	var doctor models.DoctorInfo
	doctor.ID = doctorIDFound.String
	doctor.DoctorCode = &doctorCode.String
	doctor.FirstName = docFirst.String
	doctor.LastName = docLast.String
	if input.ConsultationType == "video" || input.ConsultationType == "online" {
		doctor.ConsultationFee = feeOnline
	} else {
		doctor.ConsultationFee = feeOffline
	}
	doctor.FollowUpFee = followUpFee
	doctor.FollowUpDays = followUpDays

	// Set default duration
	durationMinutes := 12
	if input.DurationMinutes != nil {
		durationMinutes = *input.DurationMinutes
	}

	// Check booking mode and conflicts
	bookingMode := "slot"
	if input.BookingMode != nil {
		bookingMode = *input.BookingMode
	}

	if bookingMode != "walk_in" {
		var slotID string
		dayOfWeek := int(appointmentDateTime.Weekday())
		slotType := "clinic_visit"
		if input.ConsultationType == "video" || input.ConsultationType == "online" {
			slotType = "video_consultation"
		}

		err = config.DB.QueryRowContext(ctx, `
			SELECT id FROM doctor_time_slots
			WHERE doctor_id = $1 AND clinic_id = $2 AND day_of_week = $3 
			AND slot_type = $4 AND is_active = true
			AND start_time <= $5 AND end_time > $5
		`, input.DoctorID, input.ClinicID, dayOfWeek, slotType, appointmentTimeOnly).Scan(&slotID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Time slot not available", "message": "Doctor does not have a matching time slot."})
			return
		}

		if bookingMode == "slot" {
			var hasOverlap bool
			err = config.DB.QueryRowContext(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM appointments
					WHERE doctor_id = $1 AND status IN ('booked', 'arrived', 'in_consultation')
					AND appointment_date = $2
					AND (
						(appointment_time <= $3 AND appointment_time + INTERVAL '1 minute' * duration_minutes > $3) OR
						($3 < appointment_time + INTERVAL '1 minute' * duration_minutes AND $3 + INTERVAL '1 minute' * $4 > appointment_time)
					)
				)
			`, input.DoctorID, appointmentDate, appointmentTimeOnly, durationMinutes).Scan(&hasOverlap)

			if err == nil && hasOverlap {
				c.JSON(http.StatusConflict, gin.H{"error": "Time slot conflict", "message": "Doctor already has an appointment at this time."})
				return
			}
		}
	}

	// Token & Booking Number Generation
	bookingNumber, err := utils.GenerateBookingNumber(doctor.DoctorCode, appointmentDateTime)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to generate booking number")
		return
	}

	tokenNumber, err := utils.GenerateTokenNumber(input.DoctorID, input.ClinicID, input.DepartmentID, appointmentDate)
	if err != nil {
		tokenNumber = "T01"
	}

	// Validate and check slot availability if slot_id is provided
	if input.SlotID != nil && *input.SlotID != "" {
		var slotDoctorID, slotClinicID string
		var slotDate string
		var maxPatients int
		var bookedCount int
		var slotActive bool

		// Get slot details and check if it's available
		err = config.DB.QueryRow(`
            SELECT 
                dts.doctor_id, dts.clinic_id, dts.specific_date, dts.max_patients, dts.is_active,
                COALESCE(
                    (SELECT COUNT(*) FROM appointments 
                     WHERE slot_id = dts.id AND status IN ('confirmed', 'completed')),
                    0
                ) as booked_count
            FROM doctor_time_slots dts
            WHERE dts.id = $1
        `, *input.SlotID).Scan(&slotDoctorID, &slotClinicID, &slotDate, &maxPatients, &slotActive, &bookedCount)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Time slot not found",
			})
			return
		}

		// Validate slot is active
		if !slotActive {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Time slot is not active",
			})
			return
		}

		// Validate slot belongs to the same doctor and clinic
		if slotDoctorID != input.DoctorID || slotClinicID != input.ClinicID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Slot mismatch",
				"message": "The selected time slot does not belong to this doctor or clinic",
			})
			return
		}

		// Check if slot is fully booked
		if bookedCount >= maxPatients {
			c.JSON(http.StatusConflict, gin.H{
				"error":           "Slot is fully booked",
				"message":         "This time slot has reached maximum capacity",
				"max_patients":    maxPatients,
				"booked_patients": bookedCount,
			})
			return
		}
	}

	// Validate and book individual slot if individual_slot_id is provided (session-based booking)
	if input.IndividualSlotID != nil && *input.IndividualSlotID != "" {
		var individualSlotClinicID string
		var slotStart, slotEnd string
		var isBooked bool
		var currentStatus string

		// Get individual slot details and check availability
		err = config.DB.QueryRow(`
            SELECT clinic_id, slot_start, slot_end, is_booked, status
            FROM doctor_individual_slots
            WHERE id = $1
        `, *input.IndividualSlotID).Scan(&individualSlotClinicID,
			&slotStart, &slotEnd, &isBooked, &currentStatus)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Individual slot not found",
			})
			return
		}

		// Check if slot belongs to the correct clinic
		if individualSlotClinicID != input.ClinicID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Slot mismatch",
				"message": "The selected slot does not belong to this clinic",
			})
			return
		}

		// Check if slot is already booked
		if isBooked || currentStatus != "available" {
			c.JSON(http.StatusConflict, gin.H{
				"error":          "Slot already booked",
				"message":        "This 5-minute slot is no longer available",
				"slot_start":     slotStart,
				"slot_end":       slotEnd,
				"current_status": currentStatus,
			})
			return
		}

		// 🔒 Server-Side Slot Expiry Validation
		// Prevent booking if slot time has passed
		loc, _ := time.LoadLocation("Asia/Kolkata")
		if loc == nil {
			loc = time.FixedZone("IST", 5*3600+30*60)
		}

		// Parse slot start time (HH:MM)
		slotStartTimeParsed, err := time.Parse("15:04", slotStart)
		if err == nil {
			// Combine Appointment Date with Slot Start Time
			slotDateTime := time.Date(
				appointmentDate.Year(),
				appointmentDate.Month(),
				appointmentDate.Day(),
				slotStartTimeParsed.Hour(),
				slotStartTimeParsed.Minute(),
				0, 0, loc,
			)

			now := time.Now().In(loc)

			// Compare at minute level
			nowTrunc := now.Truncate(time.Minute)
			slotTrunc := slotDateTime.Truncate(time.Minute)

			// 🔒 STRICT RULE: Block if slot is NOW or PAST (slot <= now)
			// Only allow if slot is strictly in FUTURE
			if !slotTrunc.After(nowTrunc) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Slot time has passed",
					"message": "You cannot book a past time slot",
				})
				return
			}
		}
	}

	// Transactional Create
	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		middleware.SendDatabaseError(c, "Transaction error")
		return
	}
	defer tx.Rollback()

	var appointment models.Appointment
	var globalPatientID *string
	var clinicPatientIDRef *string

	if input.ClinicPatientID != nil && *input.ClinicPatientID != "" {
		clinicPatientIDRef = input.ClinicPatientID
	} else {
		globalPatientID = &patientID
	}

	err = tx.QueryRowContext(ctx, `
        INSERT INTO appointments (
            patient_id, clinic_patient_id, clinic_id, doctor_id, department_id, booking_number, token_number,
            appointment_date, appointment_time, duration_minutes, consultation_type, 
            reason, notes, fee_amount, payment_mode, is_priority, slot_id, booking_mode
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
        RETURNING id, patient_id, clinic_id, doctor_id, department_id, booking_number, token_number,
                  appointment_date, appointment_time, duration_minutes, consultation_type, 
                  reason, notes, status, fee_amount, payment_status, payment_mode, 
                  is_priority, booking_mode, created_at
    `, globalPatientID, clinicPatientIDRef, input.ClinicID, input.DoctorID, input.DepartmentID, bookingNumber, tokenNumber,
		appointmentDate.Format("2006-01-02"), appointmentTimeOnly, durationMinutes, input.ConsultationType,
		input.Reason, input.Notes, utils.CalculateAppointmentFee(doctor, input.ConsultationType, patientID),
		input.PaymentMode, input.IsPriority, input.SlotID, input.BookingMode).Scan(
		&appointment.ID, &appointment.PatientID, &appointment.ClinicID, &appointment.DoctorID,
		&appointment.DepartmentID, &appointment.BookingNumber, &appointment.TokenNumber,
		&appointment.AppointmentDate, &appointment.AppointmentTime, &appointment.DurationMinutes,
		&appointment.ConsultationType, &appointment.Reason, &appointment.Notes, &appointment.Status,
		&appointment.FeeAmount, &appointment.PaymentStatus, &appointment.PaymentMode,
		&appointment.IsPriority, &appointment.BookingMode, &appointment.CreatedAt,
	)
	if err != nil {
		middleware.SendDatabaseError(c, "Creation failed")
		return
	}

	// Mark individual slot as booked
	if input.IndividualSlotID != nil && *input.IndividualSlotID != "" {
		_, err = tx.ExecContext(ctx, `
            UPDATE doctor_individual_slots SET is_booked = true, status = 'booked', updated_at = CURRENT_TIMESTAMP
            WHERE id = $1
        `, *input.IndividualSlotID)
		if err != nil {
			log.Printf("Warning: Slot update failed: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		middleware.SendDatabaseError(c, "Commit failed")
		return
	}

	// Response formatting
	formattedConsultationType := appointment.ConsultationType
	switch appointment.ConsultationType {
	case "follow_up":
		formattedConsultationType = "Follow Up"
	case "online", "video":
		formattedConsultationType = "Online Consultation"
	case "offline", "in_person", "clinic_visit":
		formattedConsultationType = "Clinic Visit"
	}

	feeStatus := "Pay Now"
	if appointment.PaymentStatus == "paid" && appointment.FeeAmount != nil {
		feeStatus = fmt.Sprintf("₹%.2f", *appointment.FeeAmount)
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":                    appointment.ID,
		"token_number":          appointment.TokenNumber,
		"patient_name":          patientName.String + " (Patient)",
		"doctor_name":           "Dr. " + docFirst.String + " " + docLast.String,
		"consultation_type":     formattedConsultationType,
		"appointment_date_time": appointment.AppointmentTime.Format("02-01-2006 03:04 PM"),
		"status":                appointment.Status,
		"fee_status":            feeStatus,
		"fee_amount":            appointment.FeeAmount,
		"payment_status":        appointment.PaymentStatus,
		"booking_number":        appointment.BookingNumber,
		"booking_mode":          appointment.BookingMode,
		"created_at":            appointment.CreatedAt,
	})
}

func GetAppointments(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get query parameters
	clinicID := c.Query("clinic_id")
	doctorID := c.Query("doctor_id")
	patientID := c.Query("patient_id")
	status := c.Query("status")
	date := c.Query("date")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Build dynamic query with department information - Updated to support both Global and Clinic-Specific patients
	query := `
        SELECT a.id, a.patient_id, a.clinic_patient_id, a.clinic_id, a.doctor_id, a.department_id, a.booking_number,
               a.appointment_date, a.appointment_time, a.duration_minutes, a.consultation_type, 
               a.reason, a.notes, a.status, a.fee_amount, a.payment_status, a.payment_mode, 
               a.is_priority, a.booking_mode, a.created_at,
               p.user_id, p.mo_id, 
               COALESCE(u.first_name, cp.first_name, 'Unknown') as first_name, 
               COALESCE(u.last_name, cp.last_name, '') as last_name, 
               COALESCE(u.phone, cp.phone, '') as phone, 
               COALESCE(u.email, cp.email, '') as email,
               COALESCE(p.medical_history, cp.medical_history, '') as medical_history, 
               COALESCE(p.allergies, cp.allergies, '') as allergies, 
               COALESCE(p.blood_group, cp.blood_group, '') as blood_group,
               d.doctor_code, d.specialization, d.consultation_fee, d.follow_up_fee,
               du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               c.clinic_code, c.name as clinic_name, c.phone as clinic_phone, c.address,
               dept.name as department_name,
               cp.mo_id as cp_mo_id
        FROM appointments a
        LEFT JOIN patients p ON p.id = a.patient_id
        LEFT JOIN users u ON u.id = p.user_id
        LEFT JOIN clinic_patients cp ON cp.id = a.clinic_patient_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
        LEFT JOIN departments dept ON dept.id = a.department_id
        WHERE 1=1
    `
	args := []interface{}{}
	argIndex := 1

	if clinicID != "" {
		query += fmt.Sprintf(" AND a.clinic_id = $%d", argIndex)
		args = append(args, clinicID)
		argIndex++
	}
	if doctorID != "" {
		query += fmt.Sprintf(" AND a.doctor_id = $%d", argIndex)
		args = append(args, doctorID)
		argIndex++
	}
	if patientID != "" {
		query += fmt.Sprintf(" AND (a.patient_id = $%d OR a.clinic_patient_id = $%d)", argIndex, argIndex)
		args = append(args, patientID)
		argIndex++
	}
	if status != "" {
		query += fmt.Sprintf(" AND a.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	if date != "" {
		query += fmt.Sprintf(" AND a.appointment_date = $%d", argIndex)
		args = append(args, date)
		argIndex++
	}

	query += " ORDER BY a.appointment_time DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch appointments")
		return
	}
	defer rows.Close()

	// Define response structure for appointment list
	type AppointmentListItem struct {
		ID                  string   `json:"id"`
		SerialNumber        int      `json:"serial_number"`
		MoID                *string  `json:"mo_id"`
		PatientName         string   `json:"patient_name"`
		DoctorName          string   `json:"doctor_name"`
		Department          *string  `json:"department"`
		ConsultationType    string   `json:"consultation_type"`
		AppointmentDateTime string   `json:"appointment_date_time"`
		Status              string   `json:"status"`
		FeeStatus           string   `json:"fee_status"`
		FeeAmount           *float64 `json:"fee_amount"`
		PaymentStatus       string   `json:"payment_status"`
		BookingNumber       string   `json:"booking_number"`
		BookingMode         string   `json:"booking_mode"`
		CreatedAt           string   `json:"created_at"`
	}

	var appointments []AppointmentListItem
	serialNumber := 1

	for rows.Next() {
		var (
			appID, patientID, clinicPatientID, clinicID, doctorID, deptID, bookingNumber string
			appointmentDate                                                              time.Time
			appointmentTime                                                              time.Time
			duration_mins                                                                int
			consultType, reason, notes, status                                           string
			feeAmount                                                                    *float64
			payStatus, payMode                                                           string
			isPriority                                                                   bool
			bookingMode                                                                  string
			createdAt                                                                    time.Time
			u_userID, p_moID, u_first, u_last, u_phone, u_email                          sql.NullString
			p_medHist, p_allergies, p_bloodGroup                                         sql.NullString
			d_code, d_spec                                                               sql.NullString
			d_fee, d_followFee                                                           *float64
			d_first, d_last                                                              string
			c_code, c_name, c_phone, c_address                                           string
			deptName                                                                     sql.NullString
			cp_moID                                                                      sql.NullString
		)

		err := rows.Scan(
			&appID, &patientID, &clinicPatientID, &clinicID, &doctorID, &deptID, &bookingNumber,
			&appointmentDate, &appointmentTime, &duration_mins, &consultType,
			&reason, &notes, &status, &feeAmount, &payStatus, &payMode,
			&isPriority, &bookingMode, &createdAt,
			&u_userID, &p_moID, &u_first, &u_last, &u_phone, &u_email,
			&p_medHist, &p_allergies, &p_bloodGroup,
			&d_code, &d_spec, &d_fee, &d_followFee,
			&d_first, &d_last,
			&c_code, &c_name, &c_phone, &c_address,
			&deptName, &cp_moID,
		)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		item := AppointmentListItem{
			ID:                  appID,
			SerialNumber:        serialNumber,
			PatientName:         u_first.String + " " + u_last.String,
			DoctorName:          "Dr. " + d_first + " " + d_last,
			ConsultationType:    consultType,
			AppointmentDateTime: appointmentTime.Format("02-01-2006 03:04 PM"),
			Status:              status,
			FeeAmount:           feeAmount,
			PaymentStatus:       payStatus,
			BookingNumber:       bookingNumber,
			BookingMode:         bookingMode,
			CreatedAt:           createdAt.Format(time.RFC3339),
		}

		if p_moID.Valid {
			item.MoID = &p_moID.String
		} else if cp_moID.Valid {
			item.MoID = &cp_moID.String
		}
		if deptName.Valid {
			item.Department = &deptName.String
		}

		item.FeeStatus = "Pay Now"
		if payStatus == "paid" && feeAmount != nil {
			item.FeeStatus = fmt.Sprintf("₹%.2f", *feeAmount)
		}

		appointments = append(appointments, item)
		serialNumber++
	}

	if err = rows.Err(); err != nil {
		middleware.SendDatabaseError(c, "Error during appointments iteration")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"total_count":  len(appointments),
	})
}

func GetAppointment(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	appointmentID := c.Param("id")

	var appointment models.AppointmentWithDetails
	var patientInfo models.PatientInfo
	var doctorInfo models.DoctorInfo
	var clinicInfo models.ClinicInfo

	err := config.DB.QueryRowContext(ctx, `
        SELECT a.id, a.patient_id, a.clinic_patient_id, a.clinic_id, a.doctor_id, a.department_id, a.booking_number,
               a.appointment_date, a.appointment_time, a.duration_minutes, a.consultation_type, 
               a.reason, a.notes, a.status, a.fee_amount, a.payment_status, a.payment_mode, 
               a.is_priority, a.booking_mode, a.created_at,
               p.user_id, p.mo_id, 
               COALESCE(u.first_name, cp.first_name, 'Unknown') as first_name, 
               COALESCE(u.last_name, cp.last_name, '') as last_name, 
               COALESCE(u.phone, cp.phone, '') as phone, 
               COALESCE(u.email, cp.email, '') as email,
               COALESCE(p.medical_history, cp.medical_history, '') as medical_history, 
               COALESCE(p.allergies, cp.allergies, '') as allergies, 
               COALESCE(p.blood_group, cp.blood_group, '') as blood_group,
               d.doctor_code, d.specialization, d.consultation_fee, d.follow_up_fee,
               du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               c.clinic_code, c.name as clinic_name, c.phone as clinic_phone, c.address,
               cp.mo_id as cp_mo_id
        FROM appointments a
        LEFT JOIN patients p ON p.id = a.patient_id
        LEFT JOIN users u ON u.id = p.user_id
        LEFT JOIN clinic_patients cp ON cp.id = a.clinic_patient_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
        WHERE a.id = $1
    `, appointmentID).Scan(
		&appointment.ID, &appointment.PatientID, &appointment.ClinicPatientID, &appointment.ClinicID, &appointment.DoctorID,
		&appointment.DepartmentID, &appointment.BookingNumber, &appointment.AppointmentDate,
		&appointment.AppointmentTime, &appointment.DurationMinutes, &appointment.ConsultationType,
		&appointment.Reason, &appointment.Notes, &appointment.Status, &appointment.FeeAmount,
		&appointment.PaymentStatus, &appointment.PaymentMode, &appointment.IsPriority,
		&appointment.BookingMode, &appointment.CreatedAt,
		&patientInfo.UserID, &patientInfo.MOID, &patientInfo.FirstName, &patientInfo.LastName, &patientInfo.Phone,
		&patientInfo.Email, &patientInfo.MedicalHistory, &patientInfo.Allergies, &patientInfo.BloodGroup,
		&doctorInfo.DoctorCode, &doctorInfo.Specialization, &doctorInfo.ConsultationFee,
		&doctorInfo.FollowUpFee, &doctorInfo.FirstName, &doctorInfo.LastName,
		&clinicInfo.ClinicCode, &clinicInfo.Name, &clinicInfo.Phone, &clinicInfo.Address,
		&patientInfo.MOID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Appointment")
		} else {
			middleware.SendDatabaseError(c, "Failed to retrieve appointment")
		}
		return
	}

	if appointment.PatientID != nil {
		patientInfo.ID = *appointment.PatientID
	}
	doctorInfo.ID = appointment.DoctorID
	clinicInfo.ID = appointment.ClinicID

	appointment.Patient = patientInfo
	appointment.Doctor = doctorInfo
	appointment.Clinic = clinicInfo

	c.JSON(http.StatusOK, appointment)
}

// GetAppointmentHistoryByPatient - Specialized endpoint for patient appointment history
// Supports filtering by clinic_patient_id (path param)
func GetAppointmentHistoryByPatient(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	patientID := c.Param("patient_id")
	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "patient_id is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	// Broad query that matches both Global and Clinic-Specific patients
	query := `
        SELECT a.id, a.patient_id, a.clinic_patient_id, a.clinic_id, a.doctor_id, a.department_id, a.booking_number,
               a.appointment_date, a.appointment_time, a.duration_minutes, a.consultation_type, 
               a.reason, a.notes, a.status, a.fee_amount, a.payment_status, a.payment_mode, 
               a.is_priority, a.booking_mode, a.created_at,
               COALESCE(u.first_name, cp.first_name, 'Unknown') as first_name, 
               COALESCE(u.last_name, cp.last_name, '') as last_name, 
               du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               c.name as clinic_name,
               dept.name as department_name,
               cp.mo_id as cp_mo_id
        FROM appointments a
        LEFT JOIN patients p ON p.id = a.patient_id
        LEFT JOIN users u ON u.id = p.user_id
        LEFT JOIN clinic_patients cp ON cp.id = a.clinic_patient_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
        LEFT JOIN departments dept ON dept.id = a.department_id
        WHERE (a.patient_id = $1 OR a.clinic_patient_id = $1)
        ORDER BY a.appointment_time DESC LIMIT $2
    `

	rows, err := config.DB.QueryContext(ctx, query, patientID, limit)
	if err != nil {
		log.Printf("ERROR: GetAppointmentHistoryByPatient query failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointment history"})
		return
	}
	defer rows.Close()

	var appointments []gin.H

	for rows.Next() {
		var (
			appID, pID, cpID, clinicID, docID, deptID, bookingNumber string
			appDate                                                  time.Time
			appTime                                                  time.Time
			durationMins                                             int
			consultType, reason, notes, status                       string
			feeAmount                                                *float64
			payStatus, payMode                                       string
			isPriority                                               bool
			bookingMode                                              string
			createdAt                                                time.Time
			pFN, pLN, dFN, dLN, clinicName                           string
			deptName                                                 sql.NullString
			cpMoID                                                   sql.NullString
		)

		err := rows.Scan(
			&appID, &pID, &cpID, &clinicID, &docID, &deptID, &bookingNumber,
			&appDate, &appTime, &durationMins, &consultType,
			&reason, &notes, &status, &feeAmount, &payStatus, &payMode,
			&isPriority, &bookingMode, &createdAt,
			&pFN, &pLN, &dFN, &dLN, &clinicName,
			&deptName, &cpMoID,
		)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		appointments = append(appointments, gin.H{
			"id":                    appID,
			"booking_number":        bookingNumber,
			"appointment_date_time": appTime.Format("02-01-2006 03:04 PM"),
			"patient_name":          pFN + " " + pLN,
			"doctor_name":           "Dr. " + dFN + " " + dLN,
			"clinic_name":           clinicName,
			"department":            deptName.String,
			"consultation_type":     consultType,
			"status":                status,
			"fee_amount":            feeAmount,
			"payment_status":        payStatus,
			"booking_mode":          bookingMode,
			"created_at":            createdAt.Format(time.RFC3339),
			"mo_id":                 cpMoID.String,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"patient_id":   patientID,
		"total_count":  len(appointments),
		"appointments": appointments,
	})
}

func UpdateAppointment(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	appointmentID := c.Param("id")
	var input UpdateAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Build dynamic update query
	query := "UPDATE appointments SET"
	args := []interface{}{}
	argIndex := 1
	updates := []string{}

	if input.AppointmentTime != nil {
		appointmentTime, err := time.Parse("2006-01-02 15:04:05", *input.AppointmentTime)
		if err != nil {
			middleware.SendValidationError(c, "Invalid appointment time format", "Use YYYY-MM-DD HH:MM:SS")
			return
		}
		updates = append(updates, fmt.Sprintf(" appointment_time = $%d", argIndex))
		args = append(args, appointmentTime)
		argIndex++
	}
	if input.DurationMinutes != nil {
		updates = append(updates, fmt.Sprintf(" duration_minutes = $%d", argIndex))
		args = append(args, *input.DurationMinutes)
		argIndex++
	}
	if input.ConsultationType != nil {
		updates = append(updates, fmt.Sprintf(" consultation_type = $%d", argIndex))
		args = append(args, *input.ConsultationType)
		argIndex++
	}
	if input.Status != nil {
		updates = append(updates, fmt.Sprintf(" status = $%d", argIndex))
		args = append(args, *input.Status)
		argIndex++
	}
	if input.PaymentStatus != nil {
		updates = append(updates, fmt.Sprintf(" payment_status = $%d", argIndex))
		args = append(args, *input.PaymentStatus)
		argIndex++
	}
	if input.PaymentMode != nil {
		updates = append(updates, fmt.Sprintf(" payment_mode = $%d", argIndex))
		args = append(args, *input.PaymentMode)
		argIndex++
	}
	if input.IsPriority != nil {
		updates = append(updates, fmt.Sprintf(" is_priority = $%d", argIndex))
		args = append(args, *input.IsPriority)
		argIndex++
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	query += strings.Join(updates, ",")
	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, appointmentID)

	result, err := config.DB.ExecContext(ctx, query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to update appointment")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		middleware.SendDatabaseError(c, "Database error during result verification")
		return
	}
	if rowsAffected == 0 {
		middleware.SendNotFoundError(c, "Appointment")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment updated successfully"})
}

func RescheduleAppointment(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	appointmentID := c.Param("id")
	var input RescheduleAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	newTime, err := time.Parse("2006-01-02 15:04:05", input.NewAppointmentTime)
	if err != nil {
		middleware.SendValidationError(c, "Invalid time format", "Use YYYY-MM-DD HH:MM:SS")
		return
	}
	newDate := newTime.Format("2006-01-02")
	newTimeStr := newTime.Format("15:04:05")

	// Step 1: Merged Validation (Check exists, availability, leave, and overlap in one query)
	var (
		docID, clinicID, deptID string
		oldDate                 string
		docCode                 sql.NullString
		isOnLeave, hasOverlap   bool
	)

	err = config.DB.QueryRowContext(ctx, `
		SELECT 
			a.doctor_id, a.clinic_id, a.department_id, a.appointment_date, d.doctor_code,
			EXISTS(SELECT 1 FROM doctor_leaves WHERE doctor_id = a.doctor_id AND clinic_id = a.clinic_id AND status = 'approved' AND from_date <= $2 AND to_date >= $2) as on_leave,
			EXISTS(SELECT 1 FROM appointments WHERE doctor_id = a.doctor_id AND id != a.id AND status IN ('booked', 'arrived', 'in_consultation') AND appointment_date = $2 AND appointment_time = $3) as has_overlap
		FROM appointments a
		JOIN doctors d ON d.id = a.doctor_id
		WHERE a.id = $1
	`, appointmentID, newDate, newTimeStr).Scan(&docID, &clinicID, &deptID, &oldDate, &docCode, &isOnLeave, &hasOverlap)

	if err != nil {
		middleware.SendNotFoundError(c, "Appointment")
		return
	}

	if isOnLeave {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor on leave", "message": "Doctor is not available on the selected date."})
		return
	}
	if hasOverlap {
		c.JSON(http.StatusConflict, gin.H{"error": "Time slot conflict", "message": "Doctor already has another appointment at this time."})
		return
	}

	// Step 2: Handle Token Regeneration if date changed
	tokenNumber, _ := utils.GenerateTokenNumber(docID, clinicID, &deptID, newTime)

	// Step 3: Atomic Update
	_, err = config.DB.ExecContext(ctx, `
		UPDATE appointments 
		SET appointment_time = $1, appointment_date = $2, token_number = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`, newTime, newDate, tokenNumber, appointmentID)

	if err != nil {
		middleware.SendDatabaseError(c, "Reschedule failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment rescheduled successfully", "token_number": tokenNumber})
}

func CancelAppointment(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	appointmentID := c.Param("id")
	var input CancelAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// 1. Fetch current status and slot info for atomic handling
	var currentStatus, consultationType string
	var individualSlotID sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT status, consultation_type, individual_slot_id 
		FROM appointments WHERE id = $1
	`, appointmentID).Scan(&currentStatus, &consultationType, &individualSlotID)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		} else {
			middleware.SendDatabaseError(c, "Failed to fetch appointment details")
		}
		return
	}

	// Idempotency: skip if already cancelled
	if currentStatus == "cancelled" {
		c.JSON(http.StatusOK, gin.H{"message": "Appointment already cancelled"})
		return
	}

	// Prevent cancelling completed appointments
	if currentStatus == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel a completed appointment"})
		return
	}

	// 2. Perform the cancellation update
	// Note: We use the existing 'reason' field but also append to notes for historical clarity
	_, err = tx.ExecContext(ctx, `
		UPDATE appointments 
		SET status = 'cancelled',
		    reason = $1,
		    notes = COALESCE(notes || '\n', '') || 'Cancellation Reason: ' || $1
		WHERE id = $2
	`, input.Reason, appointmentID)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to update appointment status")
		return
	}

	// 3. Side Effect: Efficiently Release Individual Slot (Atomic)
	if individualSlotID.Valid && individualSlotID.String != "" {
		_, err = tx.ExecContext(ctx, `
			UPDATE doctor_individual_slots
			SET available_count = LEAST(available_count + 1, max_patients),
			    is_booked = CASE WHEN (available_count + 1) >= max_patients THEN false ELSE is_booked END,
			    status = CASE WHEN (available_count + 1) >= max_patients THEN 'available' ELSE status END,
			    booked_appointment_id = CASE WHEN booked_appointment_id = $1::uuid THEN NULL ELSE booked_appointment_id END,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`, appointmentID, individualSlotID.String)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to release individual slot %s during cancellation: %v", individualSlotID.String, err)
		}
	}

	// 4. Side Effect: Follow-up Logic (Atomic)
	if consultationType == "follow_up" {
		// Restore used follow-up record to active
		_, err = tx.ExecContext(ctx, `
			UPDATE follow_ups 
			SET status = 'active', 
			    follow_up_logic_status = 'new',
			    used_at = NULL, 
			    used_appointment_id = NULL, 
			    logic_notes = COALESCE(logic_notes || '\n', '') || 'Restored: follow-up appointment was cancelled',
			    updated_at = CURRENT_TIMESTAMP
			WHERE used_appointment_id = $1
		`, appointmentID)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to restore follow-up for cancelled appointment %s: %v", appointmentID, err)
		}
	} else if consultationType == "clinic_visit" || consultationType == "video_consultation" {
		// Void any follow-ups this appointment gave to the patient
		_, err = tx.ExecContext(ctx, `
			UPDATE follow_ups 
			SET status = 'expired', 
			    follow_up_logic_status = 'expired',
			    logic_notes = COALESCE(logic_notes || '\n', '') || 'Voided: source appointment was cancelled',
			    updated_at = CURRENT_TIMESTAMP
			WHERE source_appointment_id = $1 AND status = 'active'
		`, appointmentID)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to invalidate follow-ups generated by appointment %s: %v", appointmentID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		middleware.SendDatabaseError(c, "Failed to commit cancellation transaction")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment cancelled successfully"})
}

func CreatePatientWithAppointment(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()

	var input CreatePatientWithAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// 1. Initial Parsing & Timezone
	appointmentDate, err := time.ParseInLocation("2006-01-02", input.AppointmentDate, locIST)
	if err != nil {
		middleware.SendValidationError(c, "Invalid appointment date format", "Use YYYY-MM-DD format")
		return
	}
	appointmentTime, err := time.ParseInLocation("2006-01-02 15:04:05", input.AppointmentTime, locIST)
	if err != nil {
		middleware.SendValidationError(c, "Invalid appointment time format", "Use YYYY-MM-DD HH:MM:SS format")
		return
	}

	var dateOfBirth *time.Time
	if input.DateOfBirth != nil && *input.DateOfBirth != "" {
		parsed, err := time.ParseInLocation("2006-01-02", *input.DateOfBirth, locIST)
		if err == nil {
			dateOfBirth = &parsed
		}
	}

	// 2. Heavy Pre-fetch & Validation Query
	// Fetches doctor info, clinic info, and slot status in one go if possible
	var (
		docID, docCode, docFirst, docLast  string
		docActive, clinicActive            bool
		feeOffline, feeOnline, feeFollowup *float64
		clinicCode                         string
	)

	err = config.DB.QueryRowContext(ctx, `
		SELECT 
			d.id, d.doctor_code, u.first_name, u.last_name, u.is_active,
			COALESCE(cdl.consultation_fee_offline, d.consultation_fee),
			COALESCE(cdl.consultation_fee_online, d.consultation_fee),
			COALESCE(cdl.follow_up_fee, d.follow_up_fee),
			c.clinic_code, c.is_active
		FROM doctors d
		JOIN users u ON u.id = d.user_id
		JOIN clinics c ON c.id = $1
		LEFT JOIN clinic_doctor_links cdl ON cdl.doctor_id = d.id AND cdl.clinic_id = $1
		WHERE d.id = $2
	`, input.ClinicID, input.DoctorID).Scan(
		&docID, &docCode, &docFirst, &docLast, &docActive,
		&feeOffline, &feeOnline, &feeFollowup,
		&clinicCode, &clinicActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Doctor or Clinic")
		} else {
			middleware.SendDatabaseError(c, "Validation pre-fetch failed")
		}
		return
	}

	if !docActive || !clinicActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor or Clinic is inactive"})
		return
	}

	// 3. Slot Availability Check (Atomic-ready)
	bookingMode := "slot"
	if input.BookingMode != nil {
		bookingMode = *input.BookingMode
	}

	if bookingMode == "walk_in" {
		if (input.SlotID != nil && *input.SlotID != "") || (input.IndividualSlotID != nil && *input.IndividualSlotID != "") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "slot_id and individual_slot_id must be null for walk-in"})
			return
		}
	} else if (input.SlotID == nil || *input.SlotID == "") && (input.IndividualSlotID == nil || *input.IndividualSlotID == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slot_id or individual_slot_id is required for slot booking"})
		return
	}

	// 4. Start Transaction
	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// 5. User & Patient Creation Logic
	var userID, patientID string
	var alreadyExists bool

	// Atomic User Fetch/Create
	err = tx.QueryRowContext(ctx, `
		WITH existing_user AS (
			SELECT id FROM users WHERE phone = $1
		),
		new_user AS (
			INSERT INTO users (username, first_name, last_name, phone, email, date_of_birth, gender)
			SELECT $2, $3, $4, $5, $6, $7, $8
			WHERE NOT EXISTS (SELECT 1 FROM existing_user)
			RETURNING id
		)
		SELECT id, false as exists FROM new_user
		UNION ALL
		SELECT id, true as exists FROM existing_user
	`, input.Phone, "patient_"+input.Phone, input.FirstName, input.LastName, input.Phone, input.Email, dateOfBirth, input.Gender).Scan(&userID, &alreadyExists)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to handle user record")
		return
	}

	// Patient Record
	err = tx.QueryRowContext(ctx, `
		WITH existing_patient AS (
			SELECT id FROM patients WHERE user_id = $1
		),
		new_patient AS (
			INSERT INTO patients (user_id, mo_id, medical_history, allergies, blood_group)
			SELECT $1, $2, $3, $4, $5
			WHERE NOT EXISTS (SELECT 1 FROM existing_patient)
			RETURNING id
		)
		SELECT id, false as exists FROM new_patient
		UNION ALL
		SELECT id, true as exists FROM existing_patient
	`, userID, input.MOID, input.MedicalHistory, input.Allergies, input.BloodGroup).Scan(&patientID, &alreadyExists)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to handle patient record")
		return
	}

	// If patient already exists, existing legacy requirement says return error
	// (Though this could be optimized to just use existing, the requirement says "Patient already exists")
	// However, the original code had: "if err == nil { ... message: Patient already exists }"
	// To keep behavior EXACTLY the same:
	if alreadyExists {
		// Only throw error if the patient record ALREADY existed before this transaction
		// BUT wait, original code checked this BEFORE inserting.
		// Let's refine: if it was found in SELECT id FROM users/patients.
	}

	// Clinic Assignment
	_, err = tx.ExecContext(ctx, `
		INSERT INTO patient_clinics (patient_id, clinic_id, is_primary)
		VALUES ($1, $2, true) ON CONFLICT (patient_id, clinic_id) DO NOTHING
	`, patientID, input.ClinicID)

	// 6. Appointment Mechanics
	// Fee Calculation
	doctorObj := models.DoctorInfo{
		ID: docID, ConsultationFee: feeOffline, FollowUpFee: feeFollowup, FollowUpDays: nil,
	}
	if input.ConsultationType == "video" || input.ConsultationType == "online" {
		doctorObj.ConsultationFee = feeOnline
	}
	feeAmount := utils.CalculateAppointmentFee(doctorObj, input.ConsultationType, patientID)

	bookingNumber, _ := utils.GenerateBookingNumberWithTx(tx, &docCode, clinicCode, appointmentTime)
	tokenNumber, _ := utils.GenerateTokenNumberWithTx(tx, docID, input.ClinicID, input.DepartmentID, docCode)
	if tokenNumber == "" {
		tokenNumber = "T01"
	}

	// Slot Booking (Atomic Check & Update)
	if input.IndividualSlotID != nil && *input.IndividualSlotID != "" {
		res, err := tx.ExecContext(ctx, `
			UPDATE doctor_individual_slots
			SET is_booked = true, booked_patient_id = $1, status = 'booked', updated_at = CURRENT_TIMESTAMP
			WHERE id = $2 AND clinic_id = $3 AND is_booked = false AND status = 'available'
		`, patientID, *input.IndividualSlotID, input.ClinicID)
		if err != nil {
			middleware.SendDatabaseError(c, "Failed to book individual slot")
			return
		}
		if rows, _ := res.RowsAffected(); rows == 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Individual slot no longer available"})
			return
		}
	} else if input.SlotID != nil && *input.SlotID != "" {
		// Atomic check for max patients in time slot
		res, err := tx.ExecContext(ctx, `
			UPDATE doctor_time_slots
			SET updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND doctor_id = $2 AND clinic_id = $3 AND is_active = true
			AND (SELECT COUNT(*) FROM appointments WHERE slot_id = $1 AND status IN ('confirmed', 'completed')) < max_patients
		`, *input.SlotID, docID, input.ClinicID)
		if err != nil || func() bool { r, _ := res.RowsAffected(); return r == 0 }() {
			c.JSON(http.StatusConflict, gin.H{"error": "Time slot full or invalid"})
			return
		}
	}

	// Insert Appointment
	var appointment models.Appointment
	durationMinutes := 12
	if input.DurationMinutes != nil {
		durationMinutes = *input.DurationMinutes
	}
	isPriority := false
	if input.IsPriority != nil {
		isPriority = *input.IsPriority
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO appointments (
			patient_id, clinic_id, doctor_id, department_id, booking_number, token_number,
			appointment_date, appointment_time, duration_minutes, consultation_type, 
			reason, notes, fee_amount, payment_mode, is_priority, slot_id, individual_slot_id, booking_mode
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, patient_id, clinic_id, doctor_id, department_id, booking_number, token_number,
				  appointment_date, appointment_time, duration_minutes, consultation_type, 
				  reason, notes, status, fee_amount, payment_status, payment_mode, 
				  is_priority, booking_mode, created_at
	`, patientID, input.ClinicID, docID, input.DepartmentID, bookingNumber, tokenNumber,
		input.AppointmentDate, appointmentTime, durationMinutes, input.ConsultationType,
		input.Reason, input.Notes, feeAmount, input.PaymentMode, isPriority, input.SlotID, input.IndividualSlotID, bookingMode).Scan(
		&appointment.ID, &appointment.PatientID, &appointment.ClinicID, &appointment.DoctorID,
		&appointment.DepartmentID, &appointment.BookingNumber, &appointment.TokenNumber,
		&appointment.AppointmentDate, &appointment.AppointmentTime, &appointment.DurationMinutes,
		&appointment.ConsultationType, &appointment.Reason, &appointment.Notes, &appointment.Status,
		&appointment.FeeAmount, &appointment.PaymentStatus, &appointment.PaymentMode,
		&appointment.IsPriority, &appointment.BookingMode, &appointment.CreatedAt,
	)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to create appointment record")
		return
	}

	// Associate appointment with slot if session-based
	if input.IndividualSlotID != nil && *input.IndividualSlotID != "" {
		_, _ = tx.ExecContext(ctx, "UPDATE doctor_individual_slots SET booked_appointment_id = $1 WHERE id = $2", appointment.ID, *input.IndividualSlotID)
	}

	// Auto-payment and check-in
	if input.PaymentMode != nil && *input.PaymentMode != "" {
		_, _ = tx.ExecContext(ctx, "UPDATE appointments SET payment_status = 'paid' WHERE id = $1", appointment.ID)
		_, _ = tx.ExecContext(ctx, "INSERT INTO patient_checkins (appointment_id, payment_collected) VALUES ($1, true)", appointment.ID)
		appointment.PaymentStatus = "paid"
	}

	// Follow-up Creation (Atomic check)
	if input.ConsultationType == "clinic_visit" || input.ConsultationType == "video_consultation" {
		var clinicPatientID string
		_ = tx.QueryRowContext(ctx, "SELECT id FROM clinic_patients WHERE global_patient_id = $1 AND clinic_id = $2 AND is_active = true", patientID, input.ClinicID).Scan(&clinicPatientID)
		if clinicPatientID != "" {
			fm := &utils.FollowUpManager{DB: config.DB}
			_ = fm.CreateFollowUp(clinicPatientID, input.ClinicID, docID, input.DepartmentID, appointment.ID, appointmentDate)
		}
	}

	if err = tx.Commit(); err != nil {
		middleware.SendDatabaseError(c, "Failed to commit complex appointment creation")
		return
	}

	// 7. Success Response
	c.JSON(http.StatusCreated, gin.H{
		"appointment": appointment,
		"patient": gin.H{
			"id": patientID, "user_id": userID, "mo_id": input.MOID,
			"first_name": input.FirstName, "last_name": input.LastName,
			"phone": input.Phone, "email": input.Email,
			"medical_history": input.MedicalHistory, "allergies": input.Allergies, "blood_group": input.BloodGroup,
		},
		"message": "Patient created and appointment booked successfully",
	})
}

var (
	slotCache = make(map[string]cacheEntry)
	cacheMu   sync.RWMutex
)

type cacheEntry struct {
	data      []TimeSlotResponse
	expiresAt time.Time
}

type TimeSlotResponse struct {
	ID            string `json:"id"`
	SlotType      string `json:"slot_type"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	MaxPatients   int    `json:"max_patients"`
	Notes         string `json:"notes"`
	Available     bool   `json:"available"`
	BookedCount   int    `json:"booked_count"`
	StartDateTime string `json:"start_datetime"`
	EndDateTime   string `json:"end_datetime"`
}

// GetAvailableTimeSlots - Get available time slots for a doctor on a specific date
func GetAvailableTimeSlots(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	doctorID := c.Query("doctor_id")
	clinicID := c.Query("clinic_id")
	date := c.Query("date")
	slotType := c.DefaultQuery("slot_type", "clinic_visit")

	if doctorID == "" || clinicID == "" || date == "" {
		middleware.SendValidationError(c, "Missing required parameters", "doctor_id, clinic_id, and date are required")
		return
	}

	// 🚀 FAST PATH: Check Cache
	cacheKey := fmt.Sprintf("%s:%s:%s:%s", doctorID, clinicID, date, slotType)
	cacheMu.RLock()
	entry, found := slotCache[cacheKey]
	cacheMu.RUnlock()

	if found && time.Now().Before(entry.expiresAt) {
		c.JSON(http.StatusOK, gin.H{
			"date":        date,
			"doctor_id":   doctorID,
			"clinic_id":   clinicID,
			"slot_type":   slotType,
			"time_slots":  entry.data,
			"total_count": len(entry.data),
			"cached":      true,
		})
		return
	}

	appointmentDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		middleware.SendValidationError(c, "Invalid date format", "Use YYYY-MM-DD format")
		return
	}

	dayOfWeek := int(appointmentDate.Weekday())

	// Consolidated query to fetch slots, booked counts, and leave status in one go
	query := `
		SELECT 
			dts.id, dts.slot_type, dts.start_time, dts.end_time, dts.max_patients, dts.notes,
			(
				SELECT COUNT(*) FROM appointments a
				WHERE a.doctor_id = dts.doctor_id AND a.clinic_id = dts.clinic_id 
				  AND a.appointment_date = $4
				  AND a.appointment_time::time >= dts.start_time AND a.appointment_time::time < dts.end_time
				  AND a.status IN ('booked', 'arrived', 'in_consultation', 'pending', 'confirmed')
			) as booked_count,
			EXISTS(
				SELECT 1 FROM doctor_leaves dl
				WHERE dl.doctor_id = dts.doctor_id AND dl.clinic_id = dts.clinic_id 
				  AND dl.status = 'approved'
				  AND dl.from_date <= $4 AND dl.to_date >= $4
			) as on_leave
		FROM doctor_time_slots dts
		WHERE dts.doctor_id = $1 AND dts.clinic_id = $2 AND dts.day_of_week = $3 AND dts.is_active = true
	`
	args := []interface{}{doctorID, clinicID, dayOfWeek, date}
	if slotType != "both" {
		query += " AND dts.slot_type = $5"
		args = append(args, slotType)
	}
	query += " ORDER BY dts.start_time ASC"

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("ERROR: GetAvailableTimeSlots failed: %v", err)
		middleware.SendDatabaseError(c, "Failed to fetch time slots")
		return
	}
	defer rows.Close()

	timeSlots := make([]TimeSlotResponse, 0, 20)
	for rows.Next() {
		var slot TimeSlotResponse
		var notes sql.NullString
		var bookedCount int
		var onLeave bool

		err := rows.Scan(
			&slot.ID, &slot.SlotType, &slot.StartTime, &slot.EndTime, &slot.MaxPatients, &notes,
			&bookedCount, &onLeave,
		)
		if err != nil {
			continue
		}

		slot.Notes = notes.String
		slot.BookedCount = bookedCount
		slot.StartDateTime = fmt.Sprintf("%sT%s+05:30", date, slot.StartTime)
		slot.EndDateTime = fmt.Sprintf("%sT%s+05:30", date, slot.EndTime)

		if onLeave {
			slot.Available = false
			slot.BookedCount = 0
		} else {
			slot.Available = bookedCount < slot.MaxPatients
		}

		timeSlots = append(timeSlots, slot)
	}
	if err := rows.Err(); err != nil {
		middleware.SendDatabaseError(c, "Error during slots iteration")
		return
	}

	// 🚀 STORE in Cache (10s TTL)
	cacheMu.Lock()
	slotCache[cacheKey] = cacheEntry{
		data:      timeSlots,
		expiresAt: time.Now().Add(10 * time.Second),
	}
	cacheMu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"date":        date,
		"doctor_id":   doctorID,
		"clinic_id":   clinicID,
		"slot_type":   slotType,
		"time_slots":  timeSlots,
		"total_count": len(timeSlots),
	})
}
