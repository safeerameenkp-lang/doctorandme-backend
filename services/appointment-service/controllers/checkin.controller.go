package controllers

import (
    "appointment-service/config"
    "appointment-service/models"
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
    "strings"
    "time"
    "shared-security"
)

// Patient Check-in Controllers
type CreateCheckinInput struct {
    AppointmentID     string `json:"appointment_id" binding:"required,uuid"`
    CheckedInBy       string `json:"checked_in_by" binding:"required,uuid"`
    PaymentCollected  *bool  `json:"payment_collected"`
}

type UpdateCheckinInput struct {
    VitalsRecorded    *bool `json:"vitals_recorded"`
    PaymentCollected   *bool `json:"payment_collected"`
}

func CreateCheckin(c *gin.Context) {
    var input CreateCheckinInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Verify appointment exists and is not already checked in
    var appointmentExists bool
    var appointmentStatus string
    err := config.DB.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM appointments WHERE id = $1
        ), (
            SELECT status FROM appointments WHERE id = $1
        )
    `, input.AppointmentID).Scan(&appointmentExists, &appointmentStatus)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if !appointmentExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Appointment not found"})
        return
    }

    // Check if already checked in
    var checkinExists bool
    err = config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM patient_checkins WHERE appointment_id = $1)
    `, input.AppointmentID).Scan(&checkinExists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if checkinExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Patient already checked in"})
        return
    }

    // Verify user exists
    var userExists bool
    err = config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND is_active = true)
    `, input.CheckedInBy).Scan(&userExists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if !userExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
        return
    }

    // Set default payment collected
    paymentCollected := false
    if input.PaymentCollected != nil {
        paymentCollected = *input.PaymentCollected
    }

    // Create check-in
    var checkin models.PatientCheckin
    err = config.DB.QueryRow(`
        INSERT INTO patient_checkins (appointment_id, checked_in_by, payment_collected)
        VALUES ($1, $2, $3)
        RETURNING id, appointment_id, checkin_time, vitals_recorded, payment_collected, checked_in_by, created_at
    `, input.AppointmentID, input.CheckedInBy, paymentCollected).Scan(
        &checkin.ID, &checkin.AppointmentID, &checkin.CheckinTime,
        &checkin.VitalsRecorded, &checkin.PaymentCollected, &checkin.CheckedInBy, &checkin.CreatedAt,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create check-in"})
        return
    }

    // Update appointment status to "arrived"
    _, err = config.DB.Exec(`
        UPDATE appointments SET status = 'arrived' WHERE id = $1
    `, input.AppointmentID)
    if err != nil {
        // Log error but don't fail the check-in
        fmt.Printf("Warning: Failed to update appointment status: %v\n", err)
    }

    // If payment is collected, update payment status
    if paymentCollected {
        _, err = config.DB.Exec(`
            UPDATE appointments SET payment_status = 'paid' WHERE id = $1
        `, input.AppointmentID)
        if err != nil {
            fmt.Printf("Warning: Failed to update payment status: %v\n", err)
        }
    }

    c.JSON(http.StatusCreated, checkin)
}

func GetCheckins(c *gin.Context) {
    // Get query parameters
    appointmentID := c.Query("appointment_id")
    clinicID := c.Query("clinic_id")
    doctorID := c.Query("doctor_id")
    date := c.Query("date")

    query := `
        SELECT pc.id, pc.appointment_id, pc.checkin_time, pc.vitals_recorded,
               pc.payment_collected, pc.checked_in_by, pc.created_at,
               a.patient_id, a.clinic_id, a.doctor_id, a.booking_number,
               a.appointment_time, a.status, a.fee_amount, a.payment_status,
               p.user_id, u.first_name, u.last_name, u.phone,
               d.doctor_code, du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               c.clinic_code, c.name as clinic_name
        FROM patient_checkins pc
        JOIN appointments a ON a.id = pc.appointment_id
        JOIN patients p ON p.id = a.patient_id
        JOIN users u ON u.id = p.user_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
        WHERE 1=1
    `
    args := []interface{}{}
    argIndex := 1

    if appointmentID != "" {
        query += fmt.Sprintf(" AND pc.appointment_id = $%d", argIndex)
        args = append(args, appointmentID)
        argIndex++
    }
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
    if date != "" {
        query += fmt.Sprintf(" AND DATE(pc.checkin_time) = $%d", argIndex)
        args = append(args, date)
        argIndex++
    }

    query += " ORDER BY pc.checkin_time DESC"

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var checkins []gin.H
    for rows.Next() {
        var checkin models.PatientCheckin
        var appointment models.Appointment
        var patientInfo models.PatientInfo
        var doctorInfo models.DoctorInfo
        var clinicInfo models.ClinicInfo

        err := rows.Scan(
            &checkin.ID, &checkin.AppointmentID, &checkin.CheckinTime,
            &checkin.VitalsRecorded, &checkin.PaymentCollected, &checkin.CheckedInBy, &checkin.CreatedAt,
            &appointment.PatientID, &appointment.ClinicID, &appointment.DoctorID, &appointment.BookingNumber,
            &appointment.AppointmentTime, &appointment.Status, &appointment.FeeAmount, &appointment.PaymentStatus,
            &patientInfo.UserID, &patientInfo.FirstName, &patientInfo.LastName, &patientInfo.Phone,
            &doctorInfo.DoctorCode, &doctorInfo.FirstName, &doctorInfo.LastName,
            &clinicInfo.ClinicCode, &clinicInfo.Name,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        checkins = append(checkins, gin.H{
            "id":                checkin.ID,
            "appointment_id":    checkin.AppointmentID,
            "checkin_time":      checkin.CheckinTime,
            "vitals_recorded":   checkin.VitalsRecorded,
            "payment_collected": checkin.PaymentCollected,
            "checked_in_by":     checkin.CheckedInBy,
            "created_at":        checkin.CreatedAt,
            "appointment": gin.H{
                "patient_id":     appointment.PatientID,
                "clinic_id":      appointment.ClinicID,
                "doctor_id":      appointment.DoctorID,
                "booking_number": appointment.BookingNumber,
                "appointment_time": appointment.AppointmentTime,
                "status":         appointment.Status,
                "fee_amount":     appointment.FeeAmount,
                "payment_status": appointment.PaymentStatus,
            },
            "patient": gin.H{
                "user_id":    patientInfo.UserID,
                "first_name": patientInfo.FirstName,
                "last_name":  patientInfo.LastName,
                "phone":      patientInfo.Phone,
            },
            "doctor": gin.H{
                "doctor_code": doctorInfo.DoctorCode,
                "first_name":  doctorInfo.FirstName,
                "last_name":   doctorInfo.LastName,
            },
            "clinic": gin.H{
                "clinic_code": clinicInfo.ClinicCode,
                "name":        clinicInfo.Name,
            },
        })
    }

    c.JSON(http.StatusOK, checkins)
}

func GetCheckin(c *gin.Context) {
    checkinID := c.Param("id")

    var checkin models.PatientCheckin
    var appointment models.Appointment
    var patientInfo models.PatientInfo
    var doctorInfo models.DoctorInfo
    var clinicInfo models.ClinicInfo

    err := config.DB.QueryRow(`
        SELECT pc.id, pc.appointment_id, pc.checkin_time, pc.vitals_recorded,
               pc.payment_collected, pc.checked_in_by, pc.created_at,
               a.patient_id, a.clinic_id, a.doctor_id, a.booking_number,
               a.appointment_time, a.status, a.fee_amount, a.payment_status,
               p.user_id, u.first_name, u.last_name, u.phone,
               d.doctor_code, du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               c.clinic_code, c.name as clinic_name
        FROM patient_checkins pc
        JOIN appointments a ON a.id = pc.appointment_id
        JOIN patients p ON p.id = a.patient_id
        JOIN users u ON u.id = p.user_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
        WHERE pc.id = $1
    `, checkinID).Scan(
        &checkin.ID, &checkin.AppointmentID, &checkin.CheckinTime,
        &checkin.VitalsRecorded, &checkin.PaymentCollected, &checkin.CheckedInBy, &checkin.CreatedAt,
        &appointment.PatientID, &appointment.ClinicID, &appointment.DoctorID, &appointment.BookingNumber,
        &appointment.AppointmentTime, &appointment.Status, &appointment.FeeAmount, &appointment.PaymentStatus,
        &patientInfo.UserID, &patientInfo.FirstName, &patientInfo.LastName, &patientInfo.Phone,
        &doctorInfo.DoctorCode, &doctorInfo.FirstName, &doctorInfo.LastName,
        &clinicInfo.ClinicCode, &clinicInfo.Name,
    )
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Check-in not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":                checkin.ID,
        "appointment_id":    checkin.AppointmentID,
        "checkin_time":      checkin.CheckinTime,
        "vitals_recorded":   checkin.VitalsRecorded,
        "payment_collected": checkin.PaymentCollected,
        "checked_in_by":     checkin.CheckedInBy,
        "created_at":        checkin.CreatedAt,
        "appointment": gin.H{
            "patient_id":     appointment.PatientID,
            "clinic_id":      appointment.ClinicID,
            "doctor_id":      appointment.DoctorID,
            "booking_number": appointment.BookingNumber,
            "appointment_time": appointment.AppointmentTime,
            "status":         appointment.Status,
            "fee_amount":     appointment.FeeAmount,
            "payment_status": appointment.PaymentStatus,
        },
        "patient": gin.H{
            "user_id":    patientInfo.UserID,
            "first_name": patientInfo.FirstName,
            "last_name":  patientInfo.LastName,
            "phone":      patientInfo.Phone,
        },
        "doctor": gin.H{
            "doctor_code": doctorInfo.DoctorCode,
            "first_name":  doctorInfo.FirstName,
            "last_name":   doctorInfo.LastName,
        },
        "clinic": gin.H{
            "clinic_code": clinicInfo.ClinicCode,
            "name":        clinicInfo.Name,
        },
    })
}

func UpdateCheckin(c *gin.Context) {
    checkinID := c.Param("id")
    var input UpdateCheckinInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Build dynamic update query
    query := "UPDATE patient_checkins SET"
    args := []interface{}{}
    argIndex := 1
    updates := []string{}

    if input.VitalsRecorded != nil {
        updates = append(updates, fmt.Sprintf(" vitals_recorded = $%d", argIndex))
        args = append(args, *input.VitalsRecorded)
        argIndex++
    }
    if input.PaymentCollected != nil {
        updates = append(updates, fmt.Sprintf(" payment_collected = $%d", argIndex))
        args = append(args, *input.PaymentCollected)
        argIndex++
    }

    if len(updates) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
        return
    }

    query += strings.Join(updates, ",")
    query += fmt.Sprintf(" WHERE id = $%d", argIndex)
    args = append(args, checkinID)

    result, err := config.DB.Exec(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update check-in"})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Check-in not found"})
        return
    }

    // If payment is collected, update appointment payment status
    if input.PaymentCollected != nil && *input.PaymentCollected {
        _, err = config.DB.Exec(`
            UPDATE appointments SET payment_status = 'paid' 
            WHERE id = (SELECT appointment_id FROM patient_checkins WHERE id = $1)
        `, checkinID)
        if err != nil {
            fmt.Printf("Warning: Failed to update payment status: %v\n", err)
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "Check-in updated successfully"})
}

func GetDoctorQueue(c *gin.Context) {
    doctorID := c.Param("doctor_id")
    date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

    query := `
        SELECT pc.id, pc.appointment_id, pc.checkin_time, pc.vitals_recorded,
               pc.payment_collected, pc.created_at,
               a.patient_id, a.booking_number, a.appointment_time, a.status,
               a.fee_amount, a.payment_status, a.is_priority,
               p.user_id, u.first_name, u.last_name, u.phone,
               p.medical_history, p.allergies, p.blood_group
        FROM patient_checkins pc
        JOIN appointments a ON a.id = pc.appointment_id
        JOIN patients p ON p.id = a.patient_id
        JOIN users u ON u.id = p.user_id
        WHERE a.doctor_id = $1 
        AND DATE(a.appointment_time) = $2
        AND a.status IN ('arrived', 'in_consultation')
        ORDER BY a.is_priority DESC, pc.checkin_time ASC
    `

    rows, err := config.DB.Query(query, doctorID, date)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var queue []gin.H
    for rows.Next() {
        var checkin models.PatientCheckin
        var appointment models.Appointment
        var patientInfo models.PatientInfo

        err := rows.Scan(
            &checkin.ID, &checkin.AppointmentID, &checkin.CheckinTime,
            &checkin.VitalsRecorded, &checkin.PaymentCollected, &checkin.CreatedAt,
            &appointment.PatientID, &appointment.BookingNumber, &appointment.AppointmentTime,
            &appointment.Status, &appointment.FeeAmount, &appointment.PaymentStatus, &appointment.IsPriority,
            &patientInfo.UserID, &patientInfo.FirstName, &patientInfo.LastName, &patientInfo.Phone,
            &patientInfo.MedicalHistory, &patientInfo.Allergies, &patientInfo.BloodGroup,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        queue = append(queue, gin.H{
            "checkin_id":        checkin.ID,
            "appointment_id":    checkin.AppointmentID,
            "checkin_time":      checkin.CheckinTime,
            "vitals_recorded":   checkin.VitalsRecorded,
            "payment_collected": checkin.PaymentCollected,
            "booking_number":    appointment.BookingNumber,
            "appointment_time":  appointment.AppointmentTime,
            "status":            appointment.Status,
            "fee_amount":        appointment.FeeAmount,
            "payment_status":    appointment.PaymentStatus,
            "is_priority":       appointment.IsPriority,
            "patient": gin.H{
                "user_id":         patientInfo.UserID,
                "first_name":      patientInfo.FirstName,
                "last_name":       patientInfo.LastName,
                "phone":           patientInfo.Phone,
                "medical_history": patientInfo.MedicalHistory,
                "allergies":       patientInfo.Allergies,
                "blood_group":     patientInfo.BloodGroup,
            },
        })
    }

    c.JSON(http.StatusOK, gin.H{
        "doctor_id": doctorID,
        "date":      date,
        "queue":     queue,
        "count":     len(queue),
    })
}
