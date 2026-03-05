package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"appointment-service/config"

	"github.com/gin-gonic/gin"
)

// AppointmentListItem represents the structure for appointment list UI
type AppointmentListItem struct {
	ID                  string   `json:"id"`
	TokenNumber         *string  `json:"token_number"`
	MoID                *string  `json:"mo_id"`
	PatientNumber       *string  `json:"patient_number"`
	ClinicPatientID     *string  `json:"clinic_patient_id"`
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
	CreatedAt           string   `json:"created_at"`
	DoctorImage         *string  `json:"doctor_image"`
	BookingMode         string   `json:"booking_mode"`
}

// GetAppointmentList - Get appointments list formatted for UI
func GetAppointmentList(c *gin.Context) {
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

	// Build simplified query for appointment list
	query := `
        SELECT a.id, a.booking_number, a.token_number, a.appointment_time, a.consultation_type, 
               a.status, a.fee_amount, a.payment_status, a.created_at, a.booking_mode,
               p.mo_id, u.first_name as patient_first_name, u.last_name as patient_last_name,
               du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               dept.name as department_name, d.profile_image
        FROM appointments a
        JOIN patients p ON p.id = a.patient_id
        JOIN users u ON u.id = p.user_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
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
		query += fmt.Sprintf(" AND a.patient_id = $%d", argIndex)
		args = append(args, patientID)
		argIndex++
	}
	if status != "" {
		query += fmt.Sprintf(" AND a.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if date != "" {
		query += fmt.Sprintf(" AND a.appointment_date = $%d", argIndex)
		args = append(args, date)
		argIndex++
	}

	query += " ORDER BY a.appointment_time DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var appointments []AppointmentListItem

	for rows.Next() {
		var appointment AppointmentListItem
		var moID, departmentName *string
		var appointmentTime time.Time
		var createdAt time.Time
		var patientFirstName, patientLastName, doctorFirstName, doctorLastName string
		var doctorImage *string

		err := rows.Scan(
			&appointment.ID, &appointment.BookingNumber, &appointment.TokenNumber, &appointmentTime, &appointment.ConsultationType,
			&appointment.Status, &appointment.FeeAmount, &appointment.PaymentStatus, &createdAt, &appointment.BookingMode,
			&moID, &patientFirstName, &patientLastName,
			&doctorFirstName, &doctorLastName, &departmentName, &doctorImage,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Set Mo ID
		if moID != nil {
			appointment.MoID = moID
		}

		// Set patient name
		appointment.PatientName = patientFirstName + " " + patientLastName + " (Patient)"

		// Set doctor name
		appointment.DoctorName = "Dr. " + doctorFirstName + " " + doctorLastName

		// Set department
		if departmentName != nil {
			appointment.Department = departmentName
		}
		appointment.DoctorImage = doctorImage

		// Format appointment date and time
		appointment.AppointmentDateTime = appointmentTime.Format("02-01-2006 03:04 PM")

		// Set created at
		appointment.CreatedAt = createdAt.Format("2006-01-02T15:04:05Z")

		// Format consultation type to match UI
		switch appointment.ConsultationType {
		case "follow_up":
			appointment.ConsultationType = "Follow Up"
		case "online", "video":
			appointment.ConsultationType = "Online Consultation"
		case "offline", "in_person", "clinic_visit":
			appointment.ConsultationType = "Clinic Visit"
		default:
			// Keep original value if no match
		}

		// Determine fee status based on payment status and fee amount
		if appointment.PaymentStatus == "paid" && appointment.FeeAmount != nil {
			appointment.FeeStatus = fmt.Sprintf("₹%.2f", *appointment.FeeAmount)
		} else {
			appointment.FeeStatus = "Pay Now"
		}

		appointments = append(appointments, appointment)
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"total_count":  len(appointments),
	})
}
