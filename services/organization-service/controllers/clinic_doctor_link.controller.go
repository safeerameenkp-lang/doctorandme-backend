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

// Clinic Doctor Link Controllers
type CreateClinicDoctorLinkInput struct {
	ClinicID               string   `json:"clinic_id" binding:"required,uuid"`
	DoctorID               string   `json:"doctor_id" binding:"required,uuid"`
	DepartmentID           *string  `json:"department_id" binding:"omitempty,uuid"` // Optional: link doctor to a department within this clinic
	ConsultationFeeOffline *float64 `json:"consultation_fee_offline"`               // Optional: Fee for offline consultation
	ConsultationFeeOnline  *float64 `json:"consultation_fee_online"`                // Optional: Fee for online consultation
	FollowUpFee            *float64 `json:"follow_up_fee"`                          // Optional: Follow-up fee
	FollowUpDays           *int     `json:"follow_up_days"`                         // Optional: Follow-up days validity
	Notes                  *string  `json:"notes"`                                  // Optional: Clinic-specific notes
}

// CreateClinicDoctorLink - Links any doctor to a clinic with clinic-specific fees
// A doctor can be linked to multiple clinics with different fees for each clinic
func CreateClinicDoctorLink(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var input CreateClinicDoctorLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate fee amounts (DECIMAL(10,2) allows max 99,999,999.99)
	if input.ConsultationFeeOffline != nil {
		if *input.ConsultationFeeOffline < 0 || *input.ConsultationFeeOffline > 99999999.99 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consultation_fee_offline", "message": "Fee must be between 0 and 99,999,999.99"})
			return
		}
	}
	if input.ConsultationFeeOnline != nil {
		if *input.ConsultationFeeOnline < 0 || *input.ConsultationFeeOnline > 99999999.99 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consultation_fee_online", "message": "Fee must be between 0 and 99,999,999.99"})
			return
		}
	}
	if input.FollowUpFee != nil {
		if *input.FollowUpFee < 0 || *input.FollowUpFee > 99999999.99 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid follow_up_fee", "message": "Fee must be between 0 and 99,999,999.99"})
			return
		}
	}
	if input.FollowUpDays != nil {
		if *input.FollowUpDays < 1 || *input.FollowUpDays > 365 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid follow_up_days", "message": "Follow-up days must be between 1 and 365"})
			return
		}
	}

	// Optimize: Verify clinic, doctor and check existing link in a single query
	var clinicExists, doctorExists, linkExists bool
	err := config.DB.QueryRowContext(ctx, `
		SELECT 
			EXISTS(SELECT 1 FROM clinics WHERE id = $1 AND is_active = true) as clinic_exists,
			EXISTS(SELECT 1 FROM doctors WHERE id = $2 AND is_active = true) as doctor_exists,
			EXISTS(SELECT 1 FROM clinic_doctor_links WHERE clinic_id = $1 AND doctor_id = $2) as link_exists
	`, input.ClinicID, input.DoctorID).Scan(&clinicExists, &doctorExists, &linkExists)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error validating entities"})
		return
	}

	if !clinicExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Clinic not found or inactive"})
		return
	}
	if !doctorExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found or inactive"})
		return
	}
	if linkExists {
		c.JSON(http.StatusConflict, gin.H{"error": "Doctor is already linked to this clinic"})
		return
	}

	// If department_id provided, validate it belongs to the same clinic
	if input.DepartmentID != nil {
		var deptBelongsToClinic bool
		err := config.DB.QueryRowContext(ctx, `
			SELECT EXISTS(SELECT 1 FROM departments WHERE id = $1 AND clinic_id = $2 AND is_active = true)
		`, *input.DepartmentID, input.ClinicID).Scan(&deptBelongsToClinic)
		if err != nil || !deptBelongsToClinic {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid department",
				"message": "The specified department does not exist or does not belong to this clinic",
			})
			return
		}
	}

	// Insert with clinic-specific fees and department
	var linkID string
	err = config.DB.QueryRowContext(ctx, `
        INSERT INTO clinic_doctor_links (
            clinic_id, doctor_id, department_id,
            consultation_fee_offline, consultation_fee_online,
            follow_up_fee, follow_up_days, notes
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
        RETURNING id
    `, input.ClinicID, input.DoctorID, input.DepartmentID,
		input.ConsultationFeeOffline, input.ConsultationFeeOnline,
		input.FollowUpFee, input.FollowUpDays, input.Notes).Scan(&linkID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create clinic doctor link", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      linkID,
		"message": "Doctor linked to clinic successfully with clinic-specific fees",
		"fees": gin.H{
			"consultation_fee_offline": input.ConsultationFeeOffline,
			"consultation_fee_online":  input.ConsultationFeeOnline,
			"follow_up_fee":            input.FollowUpFee,
			"follow_up_days":           input.FollowUpDays,
		},
	})
}

// GetClinicDoctorLinks - List all clinic-doctor links
func GetClinicDoctorLinks(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	type GetClinicDoctorLinksInput struct {
		ClinicID *string `form:"clinic_id" binding:"omitempty,uuid"`
		DoctorID *string `form:"doctor_id" binding:"omitempty,uuid"`
	}

	var input GetClinicDoctorLinksInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
        SELECT cdl.id, cdl.is_active, cdl.created_at,
               cdl.consultation_fee_offline, cdl.consultation_fee_online,
               cdl.follow_up_fee, cdl.follow_up_days, cdl.notes,
               c.id as clinic_id, c.name as clinic_name, c.clinic_code,
               d.id as doctor_id, d.doctor_code, d.specialization, d.license_number, d.profile_image,
               COALESCE(d.experience_years, 0), COALESCE(d.qualification, ''), COALESCE(d.bio, ''),
               u.first_name, u.last_name, u.email, u.username, u.phone
        FROM clinic_doctor_links cdl
        JOIN clinics c ON c.id = cdl.clinic_id
        JOIN doctors d ON d.id = cdl.doctor_id
        JOIN users u ON u.id = d.user_id
        WHERE 1=1
    `

	args := []interface{}{}
	argIndex := 1
	if input.ClinicID != nil {
		query += ` AND c.id = $` + fmt.Sprint(argIndex)
		args = append(args, *input.ClinicID)
		argIndex++
	}
	if input.DoctorID != nil {
		query += ` AND d.id = $` + fmt.Sprint(argIndex)
		args = append(args, *input.DoctorID)
		argIndex++
	}

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links"})
		return
	}
	defer rows.Close()

	// Pre-allocate slice for performance
	results := make([]gin.H, 0, 20)
	for rows.Next() {
		var linkID, clinicID, clinicName, clinicCode string
		var doctorID, doctorCode string
		var specialization, licenseNumber, firstName, lastName, email, username string
		var phone, notes, profileImage *string
		var consultationFeeOffline, consultationFeeOnline, followUpFee *float64
		var followUpDays *int
		var isActive bool
		var createdAt string
		var experienceYears int
		var qualification, bio string

		if err := rows.Scan(&linkID, &isActive, &createdAt,
			&consultationFeeOffline, &consultationFeeOnline,
			&followUpFee, &followUpDays, &notes,
			&clinicID, &clinicName, &clinicCode,
			&doctorID, &doctorCode, &specialization, &licenseNumber, &profileImage,
			&experienceYears, &qualification, &bio,
			&firstName, &lastName, &email, &username, &phone); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
			return
		}

		results = append(results, gin.H{
			"link_id":    linkID,
			"is_active":  isActive,
			"join_date":  createdAt,
			"created_at": createdAt,
			"clinic": gin.H{
				"clinic_id":   clinicID,
				"name":        clinicName,
				"clinic_code": clinicCode,
			},
			"doctor": gin.H{
				"doctor_id":        doctorID,
				"doctor_code":      doctorCode,
				"specialization":   specialization,
				"license_number":   licenseNumber,
				"first_name":       firstName,
				"last_name":        lastName,
				"email":            email,
				"username":         username,
				"phone":            phone,
				"profile_image":    profileImage,
				"experience_years": experienceYears,
				"qualification":    qualification,
				"bio":              bio,
			},
			"fees": gin.H{
				"consultation_fee_offline": consultationFeeOffline,
				"consultation_fee_online":  consultationFeeOnline,
				"follow_up_fee":            followUpFee,
				"follow_up_days":           followUpDays,
			},
			"notes": notes,
		})
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"links": results, "count": len(results)})
}

// GetClinicDoctorLinksByDoctor - Get all clinic links for a specific doctor
func GetClinicDoctorLinksByDoctor(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	doctorID := c.Param("doctor_id")

	if doctorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id is required"})
		return
	}

	// Get doctor basic info (also verifies existence intrinsically)
	var doctorInfo struct {
		DoctorCode      *string
		Specialization  *string
		LicenseNumber   *string
		FirstName       string
		LastName        string
		Email           *string
		Phone           *string
		ProfileImage    *string
		ExperienceYears int
		Qualification   string
		Bio             string
	}

	err := config.DB.QueryRowContext(ctx, `
        SELECT CAST(d.doctor_code AS VARCHAR), CAST(d.specialization AS VARCHAR), CAST(d.license_number AS VARCHAR), CAST(d.profile_image AS VARCHAR),
               COALESCE(d.experience_years, 0), COALESCE(d.qualification, ''), COALESCE(d.bio, ''),
               u.first_name, u.last_name, CAST(u.email AS VARCHAR), CAST(u.phone AS VARCHAR)
        FROM doctors d
        JOIN users u ON u.id = d.user_id
        WHERE d.id = $1 AND d.is_active = true
    `, doctorID).Scan(
		&doctorInfo.DoctorCode, &doctorInfo.Specialization, &doctorInfo.LicenseNumber, &doctorInfo.ProfileImage,
		&doctorInfo.ExperienceYears, &doctorInfo.Qualification, &doctorInfo.Bio,
		&doctorInfo.FirstName, &doctorInfo.LastName, &doctorInfo.Email, &doctorInfo.Phone,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found or inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctor info"})
		return
	}

	query := `
        SELECT cdl.id, cdl.is_active, cdl.created_at, cdl.updated_at,
               cdl.consultation_fee_offline, cdl.consultation_fee_online,
               cdl.follow_up_fee, cdl.follow_up_days, cdl.notes,
               c.id as clinic_id, c.name as clinic_name, c.clinic_code,
               c.phone as clinic_phone, c.address as clinic_address, c.email as clinic_email
        FROM clinic_doctor_links cdl
        JOIN clinics c ON c.id = cdl.clinic_id
        WHERE cdl.doctor_id = $1
        ORDER BY cdl.created_at DESC
    `

	rows, err := config.DB.QueryContext(ctx, query, doctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clinic links", "details": err.Error()})
		return
	}
	defer rows.Close()

	clinics := make([]gin.H, 0, 10) // Pre-allocate
	for rows.Next() {
		var linkID, clinicID, clinicName, clinicCode string
		var clinicPhone, clinicAddress, clinicEmail, notes *string
		var consultationFeeOffline, consultationFeeOnline, followUpFee *float64
		var followUpDays *int
		var isActive bool
		var createdAt, updatedAt string

		if err := rows.Scan(
			&linkID, &isActive, &createdAt, &updatedAt,
			&consultationFeeOffline, &consultationFeeOnline,
			&followUpFee, &followUpDays, &notes,
			&clinicID, &clinicName, &clinicCode,
			&clinicPhone, &clinicAddress, &clinicEmail,
		); err != nil {
			continue
		}

		clinics = append(clinics, gin.H{
			"link_id":    linkID,
			"is_active":  isActive,
			"join_date":  createdAt,
			"created_at": createdAt,
			"updated_at": updatedAt,
			"clinic": gin.H{
				"clinic_id":   clinicID,
				"name":        clinicName,
				"clinic_code": clinicCode,
				"phone":       clinicPhone,
				"address":     clinicAddress,
				"email":       clinicEmail,
			},
			"fees": gin.H{
				"consultation_fee_offline": consultationFeeOffline,
				"consultation_fee_online":  consultationFeeOnline,
				"follow_up_fee":            followUpFee,
				"follow_up_days":           followUpDays,
			},
			"notes": notes,
		})
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"doctor": gin.H{
			"doctor_id":        doctorID,
			"doctor_code":      doctorInfo.DoctorCode,
			"specialization":   doctorInfo.Specialization,
			"license_number":   doctorInfo.LicenseNumber,
			"first_name":       doctorInfo.FirstName,
			"last_name":        doctorInfo.LastName,
			"full_name":        doctorInfo.FirstName + " " + doctorInfo.LastName,
			"email":            doctorInfo.Email,
			"phone":            doctorInfo.Phone,
			"profile_image":    doctorInfo.ProfileImage,
			"experience_years": doctorInfo.ExperienceYears,
			"qualification":    doctorInfo.Qualification,
			"bio":              doctorInfo.Bio,
		},
		"clinics":       clinics,
		"total_clinics": len(clinics),
	})
}

// UpdateClinicDoctorLinkFees - Update clinic-specific fees for a doctor
type UpdateClinicDoctorLinkInput struct {
	DepartmentID           *string  `json:"department_id" binding:"omitempty,uuid"`
	ConsultationFeeOffline *float64 `json:"consultation_fee_offline"`
	ConsultationFeeOnline  *float64 `json:"consultation_fee_online"`
	FollowUpFee            *float64 `json:"follow_up_fee"`
	FollowUpDays           *int     `json:"follow_up_days"`
	Notes                  *string  `json:"notes"`
}

func UpdateClinicDoctorLink(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	linkID := c.Param("id")

	var input UpdateClinicDoctorLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.ConsultationFeeOffline != nil {
		if *input.ConsultationFeeOffline < 0 || *input.ConsultationFeeOffline > 99999999.99 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consultation_fee_offline", "message": "Fee must be between 0 and 99,999,999.99"})
			return
		}
	}
	if input.ConsultationFeeOnline != nil {
		if *input.ConsultationFeeOnline < 0 || *input.ConsultationFeeOnline > 99999999.99 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consultation_fee_online", "message": "Fee must be between 0 and 99,999,999.99"})
			return
		}
	}
	if input.FollowUpFee != nil {
		if *input.FollowUpFee < 0 || *input.FollowUpFee > 99999999.99 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid follow_up_fee", "message": "Fee must be between 0 and 99,999,999.99"})
			return
		}
	}
	if input.FollowUpDays != nil {
		if *input.FollowUpDays < 1 || *input.FollowUpDays > 365 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid follow_up_days", "message": "Follow-up days must be between 1 and 365"})
			return
		}
	}

	query := `UPDATE clinic_doctor_links SET updated_at = CURRENT_TIMESTAMP`
	args := make([]interface{}, 0, 7)
	argIndex := 1

	if input.DepartmentID != nil {
		query += fmt.Sprintf(`, department_id = $%d`, argIndex)
		args = append(args, *input.DepartmentID)
		argIndex++
	}
	if input.ConsultationFeeOffline != nil {
		query += fmt.Sprintf(`, consultation_fee_offline = $%d`, argIndex)
		args = append(args, *input.ConsultationFeeOffline)
		argIndex++
	}
	if input.ConsultationFeeOnline != nil {
		query += fmt.Sprintf(`, consultation_fee_online = $%d`, argIndex)
		args = append(args, *input.ConsultationFeeOnline)
		argIndex++
	}
	if input.FollowUpFee != nil {
		query += fmt.Sprintf(`, follow_up_fee = $%d`, argIndex)
		args = append(args, *input.FollowUpFee)
		argIndex++
	}
	if input.FollowUpDays != nil {
		query += fmt.Sprintf(`, follow_up_days = $%d`, argIndex)
		args = append(args, *input.FollowUpDays)
		argIndex++
	}
	if input.Notes != nil {
		query += fmt.Sprintf(`, notes = $%d`, argIndex)
		args = append(args, *input.Notes)
		argIndex++
	}

	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	query += fmt.Sprintf(` WHERE id = $%d`, argIndex)
	args = append(args, linkID)

	result, err := config.DB.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update clinic doctor link fees"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clinic doctor link not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Clinic-specific fees updated successfully"})
}

func DeleteClinicDoctorLink(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	linkID := c.Param("id")

	result, err := config.DB.ExecContext(ctx, `DELETE FROM clinic_doctor_links WHERE id = $1`, linkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete clinic doctor link"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clinic doctor link not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Doctor unlinked from clinic successfully"})
}
