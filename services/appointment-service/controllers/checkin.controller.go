package controllers

import (
	"appointment-service/config"
	"appointment-service/middleware"
	"appointment-service/models"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Patient Check-in Controllers
type CreateCheckinInput struct {
	AppointmentID    string `json:"appointment_id" binding:"required,uuid"`
	CheckedInBy      string `json:"checked_in_by" binding:"required,uuid"`
	PaymentCollected *bool  `json:"payment_collected"`
}

type UpdateCheckinInput struct {
	VitalsRecorded   *bool `json:"vitals_recorded"`
	PaymentCollected *bool `json:"payment_collected"`
}

func CreateCheckin(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var input CreateCheckinInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Step 1: Merged Validation
	var (
		appExists, userExists, checkinExists bool
		currentStatus                        string
	)

	err := config.DB.QueryRowContext(ctx, `
		SELECT 
			EXISTS(SELECT 1 FROM appointments WHERE id = $1) as app_exists,
			EXISTS(SELECT 1 FROM users WHERE id = $2 AND is_active = true) as user_exists,
			EXISTS(SELECT 1 FROM patient_checkins WHERE appointment_id = $1) as checkin_exists,
			COALESCE((SELECT status FROM appointments WHERE id = $1), 'unknown')
	`, input.AppointmentID, input.CheckedInBy).Scan(&appExists, &userExists, &checkinExists, &currentStatus)

	if err != nil || !appExists {
		middleware.SendNotFoundError(c, "Appointment")
		return
	}
	if !userExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found or inactive"})
		return
	}
	if checkinExists {
		c.JSON(http.StatusConflict, gin.H{"error": "Already checked in"})
		return
	}

	// Step 2: Atomic Transactional Create
	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		middleware.SendDatabaseError(c, "Transaction failed")
		return
	}
	defer tx.Rollback()

	paymentCollected := false
	if input.PaymentCollected != nil {
		paymentCollected = *input.PaymentCollected
	}

	var checkin models.PatientCheckin
	err = tx.QueryRowContext(ctx, `
        INSERT INTO patient_checkins (appointment_id, checked_in_by, payment_collected)
        VALUES ($1, $2, $3)
        RETURNING id, appointment_id, checkin_time, vitals_recorded, payment_collected, checked_in_by, created_at
    `, input.AppointmentID, input.CheckedInBy, paymentCollected).Scan(
		&checkin.ID, &checkin.AppointmentID, &checkin.CheckinTime,
		&checkin.VitalsRecorded, &checkin.PaymentCollected, &checkin.CheckedInBy, &checkin.CreatedAt,
	)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to create check-in")
		return
	}

	// Update appointment status to "arrived"
	queryUpdate := "UPDATE appointments SET status = 'arrived'"
	if paymentCollected {
		queryUpdate += ", payment_status = 'paid'"
	}
	queryUpdate += " WHERE id = $1"

	_, err = tx.ExecContext(ctx, queryUpdate, input.AppointmentID)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to update appointment status")
		return
	}

	if err = tx.Commit(); err != nil {
		middleware.SendDatabaseError(c, "Commit failed")
		return
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
		startTime, _ := time.Parse("2006-01-02", date)
		endTime := startTime.AddDate(0, 0, 1)
		query += fmt.Sprintf(" AND pc.checkin_time >= $%d AND pc.checkin_time < $%d", argIndex, argIndex+1)
		args = append(args, startTime, endTime)
		argIndex += 2
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
				"patient_id":       appointment.PatientID,
				"clinic_id":        appointment.ClinicID,
				"doctor_id":        appointment.DoctorID,
				"booking_number":   appointment.BookingNumber,
				"appointment_time": appointment.AppointmentTime,
				"status":           appointment.Status,
				"fee_amount":       appointment.FeeAmount,
				"payment_status":   appointment.PaymentStatus,
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
			"patient_id":       appointment.PatientID,
			"clinic_id":        appointment.ClinicID,
			"doctor_id":        appointment.DoctorID,
			"booking_number":   appointment.BookingNumber,
			"appointment_time": appointment.AppointmentTime,
			"status":           appointment.Status,
			"fee_amount":       appointment.FeeAmount,
			"payment_status":   appointment.PaymentStatus,
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

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
        AND a.appointment_date = $2
        AND a.status IN ('arrived', 'in_consultation')
        ORDER BY a.is_priority DESC, pc.checkin_time ASC
    `

	rows, err := config.DB.QueryContext(ctx, query, doctorID, date)
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
	if err = rows.Err(); err != nil {
		middleware.SendDatabaseError(c, "Error during queue iteration")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"doctor_id": doctorID,
		"date":      date,
		"queue":     queue,
		"count":     len(queue),
	})
}
