package controllers

import (
    "appointment-service/config"
    "appointment-service/models"
    "appointment-service/utils"
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
    "strconv"
    "strings"
    "time"
    "shared-security"
)

// Appointment Controllers
// CreateAppointmentInput - Can accept either user_id or patient_id
// If user_id is provided, patient record is created automatically if doesn't exist
type CreateAppointmentInput struct {
    UserID           *string `json:"user_id" binding:"omitempty,uuid"`
    PatientID        *string `json:"patient_id" binding:"omitempty,uuid"`
    ClinicID         string  `json:"clinic_id" binding:"required,uuid"`
    DoctorID         string  `json:"doctor_id" binding:"required,uuid"`
    DepartmentID     *string `json:"department_id" binding:"omitempty,uuid"`
    AppointmentDate  string  `json:"appointment_date" binding:"required"`
    AppointmentTime  string  `json:"appointment_time" binding:"required"`
    DurationMinutes  *int    `json:"duration_minutes"`
    ConsultationType string  `json:"consultation_type" binding:"required,oneof=new followup walkin emergency"`
    Reason           *string `json:"reason"`
    Notes            *string `json:"notes"`
    IsPriority       *bool   `json:"is_priority"`
    PaymentMode      *string `json:"payment_mode" binding:"omitempty,oneof=cash card upi"`
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
    AppointmentDate  string  `json:"appointment_date" binding:"required"`
    AppointmentTime  string  `json:"appointment_time" binding:"required"`
    DurationMinutes  *int    `json:"duration_minutes"`
    ConsultationType string  `json:"consultation_type" binding:"required,oneof=new followup walkin emergency"`
    Reason           *string `json:"reason"`
    Notes            *string `json:"notes"`
    IsPriority       *bool   `json:"is_priority"`
    PaymentMode      *string `json:"payment_mode" binding:"omitempty,oneof=cash card upi"`
}

func CreateAppointment(c *gin.Context) {
    var input CreateAppointmentInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Validate that either user_id or patient_id is provided
    if input.UserID == nil && input.PatientID == nil {
        security.SendValidationError(c, "Missing required field", "Either user_id or patient_id must be provided")
        return
    }

    // Parse appointment date
    appointmentDate, err := time.Parse("2006-01-02", input.AppointmentDate)
    if err != nil {
        security.SendValidationError(c, "Invalid appointment date format", "Use YYYY-MM-DD format")
        return
    }

    // Parse appointment time
    appointmentTime, err := time.Parse("2006-01-02 15:04:05", input.AppointmentTime)
    if err != nil {
        security.SendValidationError(c, "Invalid appointment time format", "Use YYYY-MM-DD HH:MM:SS format")
        return
    }

    // Get or create patient record
    var patientID string
    var patientCreated bool = false
    
    if input.PatientID != nil && *input.PatientID != "" {
        // Use provided patient_id and verify it exists
        var patientExists bool
        err = config.DB.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM patients p
                JOIN users u ON u.id = p.user_id
                WHERE p.id = $1 AND p.is_active = true AND u.is_active = true
            )
        `, *input.PatientID).Scan(&patientExists)
        if err != nil {
            security.SendDatabaseError(c, "Database error while checking patient")
            return
        }
        if !patientExists {
            security.SendNotFoundError(c, "patient")
            return
        }
        patientID = *input.PatientID
    } else if input.UserID != nil && *input.UserID != "" {
        // Check if user exists and is active
        var userExists bool
        err = config.DB.QueryRow(`
            SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND is_active = true)
        `, *input.UserID).Scan(&userExists)
        if err != nil {
            security.SendDatabaseError(c, "Database error while checking user")
            return
        }
        if !userExists {
            security.SendNotFoundError(c, "user")
            return
        }

        // Check if patient record already exists for this user
        err = config.DB.QueryRow(`
            SELECT id FROM patients WHERE user_id = $1 AND is_active = true
        `, *input.UserID).Scan(&patientID)
        
        if err != nil {
            // Patient doesn't exist, create new patient record automatically
            // Note: User already has "patient" role assigned during registration
            err = config.DB.QueryRow(`
                INSERT INTO patients (user_id, is_active)
                VALUES ($1, true)
                RETURNING id
            `, *input.UserID).Scan(&patientID)
            
            if err != nil {
                security.SendDatabaseError(c, "Failed to create patient record")
                return
            }
            patientCreated = true
            
            // Assign patient to clinic automatically
            _, err = config.DB.Exec(`
                INSERT INTO patient_clinics (patient_id, clinic_id, is_primary)
                VALUES ($1, $2, true)
                ON CONFLICT (patient_id, clinic_id) DO NOTHING
            `, patientID, input.ClinicID)
            if err != nil {
                // Log error but don't fail appointment creation
                fmt.Printf("Warning: Failed to assign patient to clinic: %v\n", err)
            }
        }
    }

    // Verify clinic exists and is active
    var clinicExists bool
    err = config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM clinics WHERE id = $1 AND is_active = true)
    `, input.ClinicID).Scan(&clinicExists)
    if err != nil {
        security.SendDatabaseError(c, "Database error while checking clinic")
        return
    }
    if !clinicExists {
        security.SendNotFoundError(c, "clinic")
        return
    }

	// Get doctor details (verify existence and get info for fee calculation and booking number)
	var doctor models.DoctorInfo
	var isDoctorActive bool
	err = config.DB.QueryRow(`
		SELECT d.id, d.user_id, d.doctor_code, d.specialization, d.consultation_fee, 
		       d.follow_up_fee, d.follow_up_days, u.first_name, u.last_name, u.is_active
		FROM doctors d
		JOIN users u ON u.id = d.user_id
		WHERE d.id = $1
	`, input.DoctorID).Scan(
		&doctor.ID, &doctor.UserID, &doctor.DoctorCode, &doctor.Specialization,
		&doctor.ConsultationFee, &doctor.FollowUpFee, &doctor.FollowUpDays,
		&doctor.FirstName, &doctor.LastName, &isDoctorActive,
	)
	if err != nil {
		security.SendNotFoundError(c, "doctor")
		return
	}
	if !isDoctorActive { // checking is_active
        security.SendValidationError(c, "Doctor is inactive", "The selected doctor is not active")
        return
    }

    // Set default duration if not provided
    durationMinutes := 12
    if input.DurationMinutes != nil {
        durationMinutes = *input.DurationMinutes
    }

    // Check if doctor is available at the requested time
    isAvailable, err := utils.CheckDoctorAvailability(input.DoctorID, appointmentTime, &durationMinutes)
    if err != nil {
        security.SendDatabaseError(c, "Failed to check doctor availability")
        return
    }
    if !isAvailable {
        security.SendValidationError(c, "Doctor not available", "Doctor is not available at the requested time")
        return
    }

    // Calculate fee based on consultation type
    feeAmount := utils.CalculateAppointmentFee(doctor, input.ConsultationType, patientID)

    // Generate booking number in format: DOCTORCODE-YYYYMMDD-NNNN
    bookingNumber, err := utils.GenerateBookingNumber(doctor.DoctorCode, appointmentTime)
    if err != nil {
        security.SendDatabaseError(c, "Failed to generate booking number")
        return
    }

    // Set default priority
    isPriority := false
    if input.IsPriority != nil {
        isPriority = *input.IsPriority
    }

    // Format appointment date as string
    appointmentDateStr := appointmentDate.Format("2006-01-02")

    // Create appointment
    var appointment models.Appointment
    err = config.DB.QueryRow(`
        INSERT INTO appointments (
            patient_id, clinic_id, doctor_id, department_id, booking_number, 
            appointment_date, appointment_time, duration_minutes, consultation_type, 
            reason, notes, fee_amount, payment_mode, is_priority
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
        RETURNING id, patient_id, clinic_id, doctor_id, department_id, booking_number, 
                  appointment_date, appointment_time, duration_minutes, consultation_type, 
                  reason, notes, status, fee_amount, payment_status, payment_mode, 
                  is_priority, created_at
    `, patientID, input.ClinicID, input.DoctorID, input.DepartmentID, bookingNumber,
        appointmentDateStr, appointmentTime, durationMinutes, input.ConsultationType,
        input.Reason, input.Notes, feeAmount, input.PaymentMode, isPriority).Scan(
        &appointment.ID, &appointment.PatientID, &appointment.ClinicID, &appointment.DoctorID,
        &appointment.DepartmentID, &appointment.BookingNumber, &appointment.AppointmentDate,
        &appointment.AppointmentTime, &appointment.DurationMinutes, &appointment.ConsultationType,
        &appointment.Reason, &appointment.Notes, &appointment.Status, &appointment.FeeAmount,
        &appointment.PaymentStatus, &appointment.PaymentMode, &appointment.IsPriority,
        &appointment.CreatedAt,
    )
    if err != nil {
        security.SendDatabaseError(c, "Failed to create appointment")
        return
    }

    // If payment is made immediately, mark as paid and create check-in
    if input.PaymentMode != nil && *input.PaymentMode != "" {
        _, err = config.DB.Exec(`
            UPDATE appointments SET payment_status = 'paid' WHERE id = $1
        `, appointment.ID)
        if err != nil {
            security.SendDatabaseError(c, "Failed to update payment status")
            return
        }
        appointment.PaymentStatus = "paid"

        // Auto check-in if payment is completed
        _, err = config.DB.Exec(`
            INSERT INTO patient_checkins (appointment_id, payment_collected)
            VALUES ($1, true)
        `, appointment.ID)
        if err != nil {
            // Log error but don't fail the appointment creation
            fmt.Printf("Warning: Failed to create auto check-in: %v\n", err)
        }
    }

    // Build response
    response := gin.H{
        "appointment": appointment,
    }
    
    if patientCreated {
        response["message"] = "Patient record created automatically and appointment booked successfully"
        response["patient_id"] = patientID
    }

    c.JSON(http.StatusCreated, response)
}

func GetAppointments(c *gin.Context) {
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

    // Build dynamic query
    query := `
        SELECT a.id, a.patient_id, a.clinic_id, a.doctor_id, a.department_id, a.booking_number,
               a.appointment_date, a.appointment_time, a.duration_minutes, a.consultation_type, 
               a.reason, a.notes, a.status, a.fee_amount, a.payment_status, a.payment_mode, 
               a.is_priority, a.created_at,
               p.user_id, p.mo_id, u.first_name, u.last_name, u.phone, u.email,
               p.medical_history, p.allergies, p.blood_group,
               d.doctor_code, d.specialization, d.consultation_fee, d.follow_up_fee,
               du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               c.clinic_code, c.name as clinic_name, c.phone as clinic_phone, c.address
        FROM appointments a
        JOIN patients p ON p.id = a.patient_id
        JOIN users u ON u.id = p.user_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
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
        query += fmt.Sprintf(" AND a.patient_id = $%d", argIndex)
        args = append(args, patientID)
        argIndex++
    }
    if status != "" {
        query += fmt.Sprintf(" AND a.status = $%d", argIndex)
        args = append(args, status)
        argIndex++
    }
    if date != "" {
        query += fmt.Sprintf(" AND DATE(a.appointment_time) = $%d", argIndex)
        args = append(args, date)
        argIndex++
    }

    query += " ORDER BY a.appointment_time DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
    args = append(args, limit, offset)

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var appointments []models.AppointmentWithDetails
    for rows.Next() {
        var appointment models.AppointmentWithDetails
        var patientInfo models.PatientInfo
        var doctorInfo models.DoctorInfo
        var clinicInfo models.ClinicInfo

        err := rows.Scan(
            &appointment.ID, &appointment.PatientID, &appointment.ClinicID, &appointment.DoctorID,
            &appointment.DepartmentID, &appointment.BookingNumber, &appointment.AppointmentDate,
            &appointment.AppointmentTime, &appointment.DurationMinutes, &appointment.ConsultationType,
            &appointment.Reason, &appointment.Notes, &appointment.Status, &appointment.FeeAmount,
            &appointment.PaymentStatus, &appointment.PaymentMode, &appointment.IsPriority,
            &appointment.CreatedAt,
            &patientInfo.UserID, &patientInfo.MOID, &patientInfo.FirstName, &patientInfo.LastName, &patientInfo.Phone,
            &patientInfo.Email, &patientInfo.MedicalHistory, &patientInfo.Allergies, &patientInfo.BloodGroup,
            &doctorInfo.DoctorCode, &doctorInfo.Specialization, &doctorInfo.ConsultationFee,
            &doctorInfo.FollowUpFee, &doctorInfo.FirstName, &doctorInfo.LastName,
            &clinicInfo.ClinicCode, &clinicInfo.Name, &clinicInfo.Phone, &clinicInfo.Address,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        patientInfo.ID = appointment.PatientID
        doctorInfo.ID = appointment.DoctorID
        clinicInfo.ID = appointment.ClinicID

        appointment.Patient = patientInfo
        appointment.Doctor = doctorInfo
        appointment.Clinic = clinicInfo

        appointments = append(appointments, appointment)
    }

    c.JSON(http.StatusOK, appointments)
}

func GetAppointment(c *gin.Context) {
    appointmentID := c.Param("id")

    var appointment models.AppointmentWithDetails
    var patientInfo models.PatientInfo
    var doctorInfo models.DoctorInfo
    var clinicInfo models.ClinicInfo

    err := config.DB.QueryRow(`
        SELECT a.id, a.patient_id, a.clinic_id, a.doctor_id, a.department_id, a.booking_number,
               a.appointment_date, a.appointment_time, a.duration_minutes, a.consultation_type, 
               a.reason, a.notes, a.status, a.fee_amount, a.payment_status, a.payment_mode, 
               a.is_priority, a.created_at,
               p.user_id, p.mo_id, u.first_name, u.last_name, u.phone, u.email,
               p.medical_history, p.allergies, p.blood_group,
               d.doctor_code, d.specialization, d.consultation_fee, d.follow_up_fee,
               du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               c.clinic_code, c.name as clinic_name, c.phone as clinic_phone, c.address
        FROM appointments a
        JOIN patients p ON p.id = a.patient_id
        JOIN users u ON u.id = p.user_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
        WHERE a.id = $1
    `, appointmentID).Scan(
        &appointment.ID, &appointment.PatientID, &appointment.ClinicID, &appointment.DoctorID,
        &appointment.DepartmentID, &appointment.BookingNumber, &appointment.AppointmentDate,
        &appointment.AppointmentTime, &appointment.DurationMinutes, &appointment.ConsultationType,
        &appointment.Reason, &appointment.Notes, &appointment.Status, &appointment.FeeAmount,
        &appointment.PaymentStatus, &appointment.PaymentMode, &appointment.IsPriority,
        &appointment.CreatedAt,
        &patientInfo.UserID, &patientInfo.MOID, &patientInfo.FirstName, &patientInfo.LastName, &patientInfo.Phone,
        &patientInfo.Email, &patientInfo.MedicalHistory, &patientInfo.Allergies, &patientInfo.BloodGroup,
        &doctorInfo.DoctorCode, &doctorInfo.Specialization, &doctorInfo.ConsultationFee,
        &doctorInfo.FollowUpFee, &doctorInfo.FirstName, &doctorInfo.LastName,
        &clinicInfo.ClinicCode, &clinicInfo.Name, &clinicInfo.Phone, &clinicInfo.Address,
    )
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
        return
    }

    patientInfo.ID = appointment.PatientID
    doctorInfo.ID = appointment.DoctorID
    clinicInfo.ID = appointment.ClinicID

    appointment.Patient = patientInfo
    appointment.Doctor = doctorInfo
    appointment.Clinic = clinicInfo

    c.JSON(http.StatusOK, appointment)
}

func UpdateAppointment(c *gin.Context) {
    appointmentID := c.Param("id")
    var input UpdateAppointmentInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment time format"})
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

    result, err := config.DB.Exec(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update appointment"})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Appointment updated successfully"})
}

func RescheduleAppointment(c *gin.Context) {
    appointmentID := c.Param("id")
    var input RescheduleAppointmentInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Parse new appointment time
    newAppointmentTime, err := time.Parse("2006-01-02 15:04:05", input.NewAppointmentTime)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment time format"})
        return
    }

    // Get current appointment details
    var currentAppointment models.Appointment
    err = config.DB.QueryRow(`
        SELECT id, doctor_id, duration_minutes FROM appointments WHERE id = $1
    `, appointmentID).Scan(&currentAppointment.ID, &currentAppointment.DoctorID, &currentAppointment.DurationMinutes)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
        return
    }

    // Check if doctor is available at the new time
    isAvailable, err := utils.CheckDoctorAvailability(currentAppointment.DoctorID, newAppointmentTime, &currentAppointment.DurationMinutes)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check doctor availability"})
        return
    }
    if !isAvailable {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor is not available at the requested time"})
        return
    }

    // Update appointment time
    result, err := config.DB.Exec(`
        UPDATE appointments SET appointment_time = $1 WHERE id = $2
    `, newAppointmentTime, appointmentID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reschedule appointment"})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Appointment rescheduled successfully"})
}

func CancelAppointment(c *gin.Context) {
    appointmentID := c.Param("id")
    var input CancelAppointmentInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := config.DB.Exec(`
        UPDATE appointments SET status = 'cancelled' WHERE id = $1
    `, appointmentID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel appointment"})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Appointment cancelled successfully"})
}

func GetAvailableTimeSlots(c *gin.Context) {
    doctorID := c.Query("doctor_id")
    date := c.Query("date")

    if doctorID == "" || date == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id and date are required"})
        return
    }

    // Parse date
    targetDate, err := time.Parse("2006-01-02", date)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
        return
    }

    // Get doctor's schedule for the day
    dayOfWeek := int(targetDate.Weekday())
    if dayOfWeek == 0 {
        dayOfWeek = 7 // Convert Sunday from 0 to 7
    }

    rows, err := config.DB.Query(`
        SELECT start_time, end_time, slot_duration_minutes
        FROM doctor_schedules
        WHERE doctor_id = $1 AND day_of_week = $2 AND is_active = true
    `, doctorID, dayOfWeek)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var timeSlots []models.TimeSlot
    for rows.Next() {
        var startTimeStr, endTimeStr string
        var slotDuration int
        err := rows.Scan(&startTimeStr, &endTimeStr, &slotDuration)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        // Parse time strings
        startTime, err := time.Parse("15:04:05", startTimeStr)
        if err != nil {
            continue
        }
        endTime, err := time.Parse("15:04:05", endTimeStr)
        if err != nil {
            continue
        }

        // Generate slots
        slots := utils.GenerateTimeSlots(targetDate, startTime, endTime, slotDuration, doctorID)
        timeSlots = append(timeSlots, slots...)
    }

    c.JSON(http.StatusOK, timeSlots)
}

func CreatePatientWithAppointment(c *gin.Context) {
    var input CreatePatientWithAppointmentInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Parse appointment date
    appointmentDate, err := time.Parse("2006-01-02", input.AppointmentDate)
    if err != nil {
        security.SendValidationError(c, "Invalid appointment date format", "Use YYYY-MM-DD format")
        return
    }

    // Parse appointment time
    appointmentTime, err := time.Parse("2006-01-02 15:04:05", input.AppointmentTime)
    if err != nil {
        security.SendValidationError(c, "Invalid appointment time format", "Use YYYY-MM-DD HH:MM:SS format")
        return
    }

    // Parse date of birth if provided
    var dateOfBirth *time.Time
    if input.DateOfBirth != nil && *input.DateOfBirth != "" {
        parsed, err := time.Parse("2006-01-02", *input.DateOfBirth)
        if err != nil {
            security.SendValidationError(c, "Invalid date of birth format", "Use YYYY-MM-DD format")
            return
        }
        dateOfBirth = &parsed
    }

    // Start transaction
    tx, err := config.DB.Begin()
    if err != nil {
        security.SendDatabaseError(c, "Failed to start transaction")
        return
    }
    defer tx.Rollback()

    // Generate username from phone number
    username := "patient_" + input.Phone

    // Check if user already exists with this phone number
    var existingUserID string
    err = tx.QueryRow(`SELECT id FROM users WHERE phone = $1`, input.Phone).Scan(&existingUserID)
    if err == nil {
        // User exists, check if patient record exists
        var existingPatientID string
        err = tx.QueryRow(`SELECT id FROM patients WHERE user_id = $1`, existingUserID).Scan(&existingPatientID)
        if err == nil {
            security.SendValidationError(c, "Patient already exists", "A patient with this phone number already exists")
            return
        }
    }

    var userID string
    if existingUserID != "" {
        userID = existingUserID
    } else {
        // Create new user
        err = tx.QueryRow(`
            INSERT INTO users (username, first_name, last_name, phone, email, date_of_birth, gender)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
            RETURNING id
        `, username, input.FirstName, input.LastName, input.Phone, input.Email, dateOfBirth, input.Gender).Scan(&userID)
        if err != nil {
            security.SendDatabaseError(c, "Failed to create user")
            return
        }
    }

    // Create patient record
    // Note: User already has "patient" role assigned during registration
    var patientID string
    err = tx.QueryRow(`
        INSERT INTO patients (user_id, mo_id, medical_history, allergies, blood_group)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `, userID, input.MOID, input.MedicalHistory, input.Allergies, input.BloodGroup).Scan(&patientID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to create patient")
        return
    }

    // Assign patient to clinic
    _, err = tx.Exec(`
        INSERT INTO patient_clinics (patient_id, clinic_id, is_primary)
        VALUES ($1, $2, true)
        ON CONFLICT (patient_id, clinic_id) DO NOTHING
    `, patientID, input.ClinicID)
    if err != nil {
        security.SendDatabaseError(c, "Failed to assign patient to clinic")
        return
    }

    // Verify doctor exists and is active
    var doctorExists bool
    err = tx.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM doctors d
            JOIN users u ON u.id = d.user_id
            WHERE d.id = $1 AND d.is_active = true AND u.is_active = true
        )
    `, input.DoctorID).Scan(&doctorExists)
    if err != nil {
        security.SendDatabaseError(c, "Database error while checking doctor")
        return
    }
    if !doctorExists {
        security.SendNotFoundError(c, "doctor")
        return
    }

    // Verify clinic exists and is active
    var clinicExists bool
    err = tx.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM clinics WHERE id = $1 AND is_active = true)
    `, input.ClinicID).Scan(&clinicExists)
    if err != nil {
        security.SendDatabaseError(c, "Database error while checking clinic")
        return
    }
    if !clinicExists {
        security.SendNotFoundError(c, "clinic")
        return
    }

    // Check if doctor is available at the requested time
    isAvailable, err := utils.CheckDoctorAvailability(input.DoctorID, appointmentTime, input.DurationMinutes)
    if err != nil {
        security.SendDatabaseError(c, "Failed to check doctor availability")
        return
    }
    if !isAvailable {
        security.SendValidationError(c, "Doctor not available", "Doctor is not available at the requested time")
        return
    }

    // Get doctor details for fee calculation
    var doctor models.DoctorInfo
    err = tx.QueryRow(`
        SELECT d.id, d.user_id, d.doctor_code, d.specialization, d.consultation_fee, 
               d.follow_up_fee, d.follow_up_days, u.first_name, u.last_name
        FROM doctors d
        JOIN users u ON u.id = d.user_id
        WHERE d.id = $1
    `, input.DoctorID).Scan(
        &doctor.ID, &doctor.UserID, &doctor.DoctorCode, &doctor.Specialization,
        &doctor.ConsultationFee, &doctor.FollowUpFee, &doctor.FollowUpDays,
        &doctor.FirstName, &doctor.LastName,
    )
    if err != nil {
        security.SendDatabaseError(c, "Failed to get doctor details")
        return
    }

    // Calculate fee based on consultation type
    feeAmount := utils.CalculateAppointmentFee(doctor, input.ConsultationType, patientID)

    // Generate booking number
    bookingNumber, err := utils.GenerateBookingNumber(doctor.DoctorCode, appointmentTime)
    if err != nil {
        security.SendDatabaseError(c, "Failed to generate booking number")
        return
    }

    // Set default duration if not provided
    durationMinutes := 12
    if input.DurationMinutes != nil {
        durationMinutes = *input.DurationMinutes
    }

    // Set default priority
    isPriority := false
    if input.IsPriority != nil {
        isPriority = *input.IsPriority
    }

    // Format appointment date as string
    appointmentDateStr := appointmentDate.Format("2006-01-02")

    // Create appointment
    var appointment models.Appointment
    err = tx.QueryRow(`
        INSERT INTO appointments (
            patient_id, clinic_id, doctor_id, department_id, booking_number, 
            appointment_date, appointment_time, duration_minutes, consultation_type, 
            reason, notes, fee_amount, payment_mode, is_priority
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
        RETURNING id, patient_id, clinic_id, doctor_id, department_id, booking_number, 
                  appointment_date, appointment_time, duration_minutes, consultation_type, 
                  reason, notes, status, fee_amount, payment_status, payment_mode, 
                  is_priority, created_at
    `, patientID, input.ClinicID, input.DoctorID, input.DepartmentID, bookingNumber,
        appointmentDateStr, appointmentTime, durationMinutes, input.ConsultationType,
        input.Reason, input.Notes, feeAmount, input.PaymentMode, isPriority).Scan(
        &appointment.ID, &appointment.PatientID, &appointment.ClinicID, &appointment.DoctorID,
        &appointment.DepartmentID, &appointment.BookingNumber, &appointment.AppointmentDate,
        &appointment.AppointmentTime, &appointment.DurationMinutes, &appointment.ConsultationType,
        &appointment.Reason, &appointment.Notes, &appointment.Status, &appointment.FeeAmount,
        &appointment.PaymentStatus, &appointment.PaymentMode, &appointment.IsPriority,
        &appointment.CreatedAt,
    )
    if err != nil {
        security.SendDatabaseError(c, "Failed to create appointment")
        return
    }

    // If payment is made immediately, mark as paid and create check-in
    if input.PaymentMode != nil && *input.PaymentMode != "" {
        _, err = tx.Exec(`
            UPDATE appointments SET payment_status = 'paid' WHERE id = $1
        `, appointment.ID)
        if err != nil {
            security.SendDatabaseError(c, "Failed to update payment status")
            return
        }
        appointment.PaymentStatus = "paid"

        // Auto check-in if payment is completed
        _, err = tx.Exec(`
            INSERT INTO patient_checkins (appointment_id, payment_collected)
            VALUES ($1, true)
        `, appointment.ID)
        if err != nil {
            // Log error but don't fail the appointment creation
            fmt.Printf("Warning: Failed to create auto check-in: %v\n", err)
        }
    }

    // Commit transaction
    if err = tx.Commit(); err != nil {
        security.SendDatabaseError(c, "Failed to commit transaction")
        return
    }

    // Return appointment with patient details
    response := gin.H{
        "appointment": appointment,
        "patient": gin.H{
            "id":              patientID,
            "user_id":         userID,
            "mo_id":           input.MOID,
            "first_name":      input.FirstName,
            "last_name":       input.LastName,
            "phone":           input.Phone,
            "email":           input.Email,
            "medical_history": input.MedicalHistory,
            "allergies":       input.Allergies,
            "blood_group":     input.BloodGroup,
        },
        "message": "Patient created and appointment booked successfully",
    }

    c.JSON(http.StatusCreated, response)
}
