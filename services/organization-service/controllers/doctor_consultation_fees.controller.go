package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"organization-service/config"

	"github.com/gin-gonic/gin"
)

// GetDoctorConsultationFees returns the consultation fees for a specific clinic doctor link.
func GetDoctorConsultationFees(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	clinicID := c.Query("clinic_id")
	clinicDoctorID := c.Query("clinic_doctor_id")

	if clinicID == "" || clinicDoctorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clinic_id and clinic_doctor_id are required"})
		return
	}

	query := `
		SELECT 
			cdl.id as clinic_doctor_id, cdl.is_active, cdl.created_at, cdl.updated_at,
			cdl.consultation_fee_offline, cdl.consultation_fee_online,
			cdl.follow_up_fee, cdl.follow_up_days, cdl.notes,
			c.id as clinic_id, c.name as clinic_name,
			d.id as doctor_id, d.doctor_code, d.specialization,
			u.first_name, u.last_name
		FROM clinic_doctor_links cdl
		JOIN clinics c ON c.id = cdl.clinic_id
		JOIN doctors d ON d.id = cdl.doctor_id
		JOIN users u ON u.id = d.user_id
		WHERE cdl.clinic_id = $1 AND cdl.id = $2
	`

	// Try with cdl.id = clinicDoctorID first, then fallback to cdl.doctor_id = clinicDoctorID just in case frontend mixed them up
	row := config.DB.QueryRowContext(ctx, query, clinicID, clinicDoctorID)

	var linkID, fetchedClinicID, doctorID, clinicName string
	var doctorCode, specialization, firstName, lastName *string
	var consultationFeeOffline, consultationFeeOnline, followUpFee *float64
	var followUpDays *int
	var notes *string
	var isActive bool
	var createdAt string
	var updatedAt *string

	err := row.Scan(&linkID, &isActive, &createdAt, &updatedAt,
		&consultationFeeOffline, &consultationFeeOnline,
		&followUpFee, &followUpDays, &notes,
		&fetchedClinicID, &clinicName,
		&doctorID, &doctorCode, &specialization,
		&firstName, &lastName)

	if err != nil {
		if err == sql.ErrNoRows {
			// Fallback: assume clinicDoctorId sent by frontend is actually doctor_id
			queryFallback := `
				SELECT 
					cdl.id as clinic_doctor_id, cdl.is_active, cdl.created_at, cdl.updated_at,
					cdl.consultation_fee_offline, cdl.consultation_fee_online,
					cdl.follow_up_fee, cdl.follow_up_days, cdl.notes,
					c.id as clinic_id, c.name as clinic_name,
					d.id as doctor_id, d.doctor_code, d.specialization,
					u.first_name, u.last_name
				FROM clinic_doctor_links cdl
				JOIN clinics c ON c.id = cdl.clinic_id
				JOIN doctors d ON d.id = cdl.doctor_id
				JOIN users u ON u.id = d.user_id
				WHERE cdl.clinic_id = $1 AND cdl.doctor_id = $2
			`
			rowFallback := config.DB.QueryRowContext(ctx, queryFallback, clinicID, clinicDoctorID)
			err = rowFallback.Scan(&linkID, &isActive, &createdAt, &updatedAt,
				&consultationFeeOffline, &consultationFeeOnline,
				&followUpFee, &followUpDays, &notes,
				&fetchedClinicID, &clinicName,
				&doctorID, &doctorCode, &specialization,
				&firstName, &lastName)

			if err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{"error": "Consultation fees not found for this doctor in this clinic", "code": "404"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch consultation fees", "details": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch consultation fees", "details": err.Error()})
			return
		}
	}

	doctorName := ""
	if firstName != nil && lastName != nil {
		doctorName = fmt.Sprintf("Dr. %s %s", *firstName, *lastName)
	}

	c.JSON(http.StatusOK, gin.H{
		"clinic_doctor_id":         linkID,
		"clinic_id":                fetchedClinicID,
		"clinic_name":              clinicName,
		"doctor_id":                doctorID,
		"doctor_name":              doctorName,
		"doctor_code":              doctorCode,
		"specialization":           specialization,
		"consultation_fee_offline": consultationFeeOffline,
		"consultation_fee_online":  consultationFeeOnline,
		"follow_up_fee":            followUpFee,
		"follow_up_days":           followUpDays,
		"notes":                    notes,
		"is_active":                isActive,
		"created_at":               createdAt,
		"updated_at":               updatedAt,
	})
}

type ConsultationFeesInput struct {
	ClinicID               string   `json:"clinic_id" binding:"required"`
	ClinicDoctorID         string   `json:"clinic_doctor_id" binding:"required"`
	ConsultationFeeOffline *float64 `json:"consultation_fee_offline"`
	ConsultationFeeOnline  *float64 `json:"consultation_fee_online"`
	FollowUpFee            *float64 `json:"follow_up_fee"`
	FollowUpDays           *int     `json:"follow_up_days"`
	Notes                  *string  `json:"notes"`
}

// AddConsultationFees will map to POST /api/doctor-consultation-fees
func AddConsultationFees(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var input ConsultationFeesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// We assume ClinicDoctorID is the ID of the clinic_doctor_links table
	// Or maybe it's the doctor_id. Let's try to update by link ID first, or if not found, by clinic_id + doctor_id.

	// Check if this link exists
	var linkID string
	err := config.DB.QueryRowContext(ctx, "SELECT id FROM clinic_doctor_links WHERE clinic_id = $1 AND (id = $2 OR doctor_id = $2)", input.ClinicID, input.ClinicDoctorID).Scan(&linkID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Clinic doctor link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	query := `
		UPDATE clinic_doctor_links 
		SET updated_at = CURRENT_TIMESTAMP,
			consultation_fee_offline = COALESCE($1, consultation_fee_offline),
			consultation_fee_online = COALESCE($2, consultation_fee_online),
			follow_up_fee = COALESCE($3, follow_up_fee),
			follow_up_days = COALESCE($4, follow_up_days)
		WHERE id = $5
		RETURNING id, clinic_id, doctor_id, consultation_fee_offline, consultation_fee_online, follow_up_fee, follow_up_days
	`

	var retLinkID, retClinicID, retDoctorID string
	var retFeeOffline, retFeeOnline, retFollowUpFee *float64
	var retFollowUpDays *int

	err = config.DB.QueryRowContext(ctx, query,
		input.ConsultationFeeOffline,
		input.ConsultationFeeOnline,
		input.FollowUpFee,
		input.FollowUpDays,
		linkID,
	).Scan(&retLinkID, &retClinicID, &retDoctorID, &retFeeOffline, &retFeeOnline, &retFollowUpFee, &retFollowUpDays)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update fees", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"clinic_doctor_id":         retLinkID,
			"clinic_id":                retClinicID,
			"doctor_id":                retDoctorID,
			"consultation_fee_offline": retFeeOffline,
			"consultation_fee_online":  retFeeOnline,
			"follow_up_fee":            retFollowUpFee,
			"follow_up_days":           retFollowUpDays,
		},
		"message": "Consultation fees added/updated successfully",
	})
}

// UpdateConsultationFees will map to PUT /api/doctor-consultation-fees
func UpdateConsultationFees(c *gin.Context) {
	// Call the exact same logic as AddConsultationFees
	AddConsultationFees(c)
}
