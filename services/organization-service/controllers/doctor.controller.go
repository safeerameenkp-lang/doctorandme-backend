package controllers

import (
	"context"
	"database/sql"
	"net/http"
	"organization-service/config"
	"organization-service/models"
	"organization-service/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Doctor Controllers
// CreateDoctorInput - Creates a doctor profile without clinic assignment
// Use clinic-doctor-links API to link doctor to multiple clinics
type CreateDoctorInput struct {
	// Option 1: Use existing doctor user
	UserID string `json:"user_id" binding:"omitempty,uuid" form:"user_id"`

	// Option 2: Create new user
	FirstName string  `json:"first_name" binding:"max=50" form:"first_name"`
	LastName  string  `json:"last_name" binding:"max=50" form:"last_name"`
	Email     string  `json:"email" binding:"required_without=UserID,email" form:"email"`
	Username  string  `json:"username" binding:"required_without=UserID,min=3,max=30" form:"username"`
	Phone     *string `json:"phone" binding:"omitempty" form:"phone"`
	Password  string  `json:"password" binding:"required_without=UserID,min=8" form:"password"`

	// Doctor profile
	DoctorCode      *string  `json:"doctor_code" binding:"omitempty,max=20" form:"doctor_code"`
	Specialization  *string  `json:"specialization" binding:"omitempty,max=100" form:"specialization"`
	LicenseNumber   *string  `json:"license_number" binding:"omitempty,max=100" form:"license_number"`
	ConsultationFee *float64 `json:"consultation_fee" binding:"omitempty,min=0" form:"consultation_fee"`
	FollowUpFee     *float64 `json:"follow_up_fee" binding:"omitempty,min=0" form:"follow_up_fee"`
	FollowUpDays    *int     `json:"follow_up_days" binding:"omitempty,min=1,max=365" form:"follow_up_days"`
}

func CreateDoctor(c *gin.Context) {
	var input CreateDoctorInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.UserID == "" {
		if input.FirstName == "" {
			input.FirstName = input.Username
		}
		if input.LastName == "" {
			input.LastName = input.Username
		}
	}

	// Handle Profile Image Upload
	var profileImagePath *string
	fileHeader, err := c.FormFile("profile_image")
	if err == nil { // No error means a file was provided
		if err := utils.ValidateImage(fileHeader); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Save image
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
			return
		}
		defer file.Close()

		savedPath, err := utils.SaveOptimizedImage(file, fileHeader.Filename, "doctors")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}
		profileImagePath = &savedPath
	} else if err != http.ErrMissingFile { // If error is not just missing file, it's a real error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process profile image: " + err.Error()})
		return
	}

	var userID string
	var firstName, lastName string

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Begin strict transaction bounds for guaranteed unique lock generation
	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stat transaction limits"})
		return
	}
	defer tx.Rollback()

	// Case 1: Create new doctor user
	if input.UserID == "" {
		// Check duplicate username/email
		var exists bool
		err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1 OR email=$2)`, input.Username, input.Email).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate user"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
			return
		}

		// Hash password
		passHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Insert new user
		err = tx.QueryRowContext(ctx, `
            INSERT INTO users (first_name, last_name, email, username, phone, password_hash)
            VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
        `, input.FirstName, input.LastName, input.Email, input.Username, input.Phone, string(passHash)).Scan(&userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create doctor user"})
			return
		}

		firstName = input.FirstName
		lastName = input.LastName

		// Assign doctor role
		var roleID string
		err = tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name='doctor' LIMIT 1`).Scan(&roleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Doctor role not found"})
			return
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id, is_active) VALUES ($1, $2, true)`, userID, roleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign doctor role"})
			return
		}

	} else {
		// Case 2: Validate existing user
		err := tx.QueryRowContext(ctx, `
            SELECT u.first_name, u.last_name FROM users u
            JOIN user_roles ur ON ur.user_id = u.id
            JOIN roles r ON r.id = ur.role_id
            WHERE u.id=$1 AND r.name='doctor' AND ur.is_active=true
        `, input.UserID).Scan(&firstName, &lastName)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User not found or not a doctor"})
			return
		}

		userID = input.UserID
	}

	// Generate or Verify Doctor Code Concurrency Safely
	var finalDoctorCode string
	if input.DoctorCode != nil && *input.DoctorCode != "" {
		finalDoctorCode = *input.DoctorCode
		var exists bool
		err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM doctors WHERE doctor_code=$1)`, finalDoctorCode).Scan(&exists)
		if err == nil && exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Doctor code already exists"})
			return
		}
	} else {
		// Autogenerate Unique
		finalDoctorCode, err = utils.GenerateDoctorCode(ctx, tx, config.DB, firstName, lastName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign unique doctor code mapping"})
			return
		}
	}

	// Create doctor profile
	var doctorID string
	err = tx.QueryRowContext(ctx, `
        INSERT INTO doctors (user_id, clinic_id, doctor_code, specialization, license_number, consultation_fee, follow_up_fee, follow_up_days, profile_image)
        VALUES ($1, NULL, $2, $3, $4, $5, $6, $7, $8) RETURNING id
    `, userID, finalDoctorCode, input.Specialization, input.LicenseNumber, input.ConsultationFee, input.FollowUpFee, input.FollowUpDays, profileImagePath).Scan(&doctorID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create doctor: " + err.Error()})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit doctor creation constraints"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"doctor_id":     doctorID,
		"user_id":       userID,
		"role":          "doctor",
		"doctor_code":   finalDoctorCode,
		"profile_image": profileImagePath,
		"message":       "Doctor created successfully. Use clinic-doctor-links API to assign to clinics.",
	})
}

func GetAllDoctors(c *gin.Context) {
	rows, err := config.DB.Query(`
        SELECT d.id, d.doctor_code, d.specialization, d.license_number, d.consultation_fee,
               d.follow_up_fee, d.follow_up_days, d.profile_image, u.id, u.first_name, u.last_name, u.email, u.username, u.phone
        FROM doctors d
        JOIN users u ON d.user_id = u.id
    `)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctors"})
		return
	}
	defer rows.Close()

	var doctors []map[string]interface{}

	for rows.Next() {
		var dID, uID, firstName, lastName, email, username, phone, doctorCode, specialization, license string
		var consultationFee, followUpFee sql.NullFloat64
		var followUpDays sql.NullInt64
		var profileImage sql.NullString

		err := rows.Scan(&dID, &doctorCode, &specialization, &license, &consultationFee,
			&followUpFee, &followUpDays, &profileImage, &uID, &firstName, &lastName, &email, &username, &phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse doctor data"})
			return
		}

		doctor := map[string]interface{}{
			"doctor_id":      dID,
			"doctor_code":    doctorCode,
			"specialization": specialization,
			"license_number": license,
			"follow_up_days": followUpDays.Int64,
			"profile_image":  profileImage.String,
			"user": map[string]interface{}{
				"user_id":    uID,
				"first_name": firstName,
				"last_name":  lastName,
				"email":      email,
				"username":   username,
				"phone":      phone,
			},
		}

		// Handle nullable fee fields
		if consultationFee.Valid {
			doctor["consultation_fee"] = consultationFee.Float64
		}
		if followUpFee.Valid {
			doctor["follow_up_fee"] = followUpFee.Float64
		}
		doctors = append(doctors, doctor)
	}

	c.JSON(http.StatusOK, gin.H{"doctors": doctors})
}

func GetDoctors(c *gin.Context) {
	clinicID := c.Query("clinic_id")

	var query string
	var args []interface{}

	if clinicID != "" {
		// If clinic_id is provided, list only doctors linked to that clinic
		query = `
            SELECT DISTINCT d.id, d.user_id, d.clinic_id, d.doctor_code, d.specialization, d.license_number, 
                   d.consultation_fee, d.follow_up_fee, d.follow_up_days, d.is_main_doctor, d.profile_image, d.is_active, d.created_at,
                   u.first_name, u.last_name, u.email, u.username, u.phone
            FROM doctors d
            JOIN users u ON u.id = d.user_id
            JOIN clinic_doctor_links cdl ON cdl.doctor_id = d.id
            WHERE cdl.clinic_id = $1 AND cdl.is_active = true AND d.is_active = true AND u.is_active = true
            ORDER BY d.created_at DESC
        `
		args = []interface{}{clinicID}
	} else {
		// If no clinic_id, list all doctors
		query = `
            SELECT d.id, d.user_id, d.clinic_id, d.doctor_code, d.specialization, d.license_number, 
                   d.consultation_fee, d.follow_up_fee, d.follow_up_days, d.is_main_doctor, d.profile_image, d.is_active, d.created_at,
                   u.first_name, u.last_name, u.email, u.username, u.phone
            FROM doctors d
            JOIN users u ON u.id = d.user_id
            WHERE d.is_active = true AND u.is_active = true
            ORDER BY d.created_at DESC
        `
	}

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctors"})
		return
	}
	defer rows.Close()

	var doctors []gin.H
	for rows.Next() {
		var doctor models.Doctor
		var firstName, lastName, email, username string
		var phone *string
		err := rows.Scan(&doctor.ID, &doctor.UserID, &doctor.ClinicID, &doctor.DoctorCode, &doctor.Specialization,
			&doctor.LicenseNumber, &doctor.ConsultationFee, &doctor.FollowUpFee, &doctor.FollowUpDays,
			&doctor.IsMainDoctor, &doctor.ProfileImage, &doctor.IsActive, &doctor.CreatedAt, &firstName, &lastName, &email, &username, &phone)
		if err != nil {
			continue
		}
		doctors = append(doctors, gin.H{
			"doctor": doctor,
			"user": gin.H{
				"first_name": firstName,
				"last_name":  lastName,
				"email":      email,
				"username":   username,
				"phone":      phone,
			},
		})
	}

	c.JSON(http.StatusOK, doctors)
}

func GetDoctor(c *gin.Context) {
	doctorID := c.Param("id")

	var doctor models.Doctor
	var firstName, lastName, email, username string
	var phone *string
	err := config.DB.QueryRow(`
        SELECT d.id, d.user_id, d.clinic_id, d.doctor_code, d.specialization, d.license_number, 
               d.consultation_fee, d.follow_up_fee, d.follow_up_days, d.is_main_doctor, d.profile_image, d.is_active, d.created_at,
               u.first_name, u.last_name, u.email, u.username, u.phone
        FROM doctors d
        JOIN users u ON u.id = d.user_id
        WHERE d.id = $1
    `, doctorID).Scan(&doctor.ID, &doctor.UserID, &doctor.ClinicID, &doctor.DoctorCode, &doctor.Specialization,
		&doctor.LicenseNumber, &doctor.ConsultationFee, &doctor.FollowUpFee, &doctor.FollowUpDays,
		&doctor.IsMainDoctor, &doctor.ProfileImage, &doctor.IsActive, &doctor.CreatedAt, &firstName, &lastName, &email, &username, &phone)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"doctor": doctor,
		"user": gin.H{
			"first_name": firstName,
			"last_name":  lastName,
			"email":      email,
			"username":   username,
			"phone":      phone,
		},
	})
}

type UpdateDoctorInput struct {
	DoctorCode      *string  `json:"doctor_code" binding:"omitempty,max=20"`
	Specialization  *string  `json:"specialization" binding:"omitempty,max=100"`
	LicenseNumber   *string  `json:"license_number" binding:"omitempty,max=100"`
	ConsultationFee *float64 `json:"consultation_fee" binding:"omitempty,min=0"`
	FollowUpFee     *float64 `json:"follow_up_fee" binding:"omitempty,min=0"`
	FollowUpDays    *int     `json:"follow_up_days" binding:"omitempty,min=1,max=365"`
	IsMainDoctor    *bool    `json:"is_main_doctor"`
	IsActive        *bool    `json:"is_active"`
}

func UpdateDoctor(c *gin.Context) {
	doctorID := c.Param("id")
	var input UpdateDoctorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build dynamic update query
	query := "UPDATE doctors SET "
	args := []interface{}{}
	argIndex := 1

	if input.DoctorCode != nil {
		query += "doctor_code = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.DoctorCode)
		argIndex++
	}
	if input.Specialization != nil {
		query += "specialization = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.Specialization)
		argIndex++
	}
	if input.LicenseNumber != nil {
		query += "license_number = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.LicenseNumber)
		argIndex++
	}
	if input.ConsultationFee != nil {
		query += "consultation_fee = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.ConsultationFee)
		argIndex++
	}
	if input.FollowUpFee != nil {
		query += "follow_up_fee = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.FollowUpFee)
		argIndex++
	}
	if input.FollowUpDays != nil {
		query += "follow_up_days = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.FollowUpDays)
		argIndex++
	}
	if input.IsMainDoctor != nil {
		query += "is_main_doctor = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.IsMainDoctor)
		argIndex++
	}
	if input.IsActive != nil {
		query += "is_active = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.IsActive)
		argIndex++
	}

	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Remove trailing comma and add WHERE clause
	query = query[:len(query)-2] + " WHERE id = $" + strconv.Itoa(argIndex)
	args = append(args, doctorID)

	result, err := config.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update doctor"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Doctor updated successfully"})
}

func DeleteDoctor(c *gin.Context) {
	doctorID := c.Param("id")

	// Soft delete by setting is_active to false
	result, err := config.DB.Exec(`UPDATE doctors SET is_active = false WHERE id = $1`, doctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate doctor"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Doctor deactivated successfully"})
}

// GetDoctorsByClinic - Get all doctors linked to a specific clinic with clinic-specific fees
func GetDoctorsByClinic(c *gin.Context) {
	clinicID := c.Param("clinic_id")

	// Verify clinic exists
	var clinicExists bool
	err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM clinics WHERE id = $1 AND is_active = true)`, clinicID).Scan(&clinicExists)
	if err != nil || !clinicExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clinic not found or inactive"})
		return
	}

	query := `
        SELECT 
            cdl.id as link_id,
            cdl.consultation_fee_offline, 
            cdl.consultation_fee_online,
            cdl.follow_up_fee, 
            cdl.follow_up_days, 
            cdl.notes,
            cdl.is_active as link_active,
            d.id as doctor_id, 
            d.doctor_code, 
            d.specialization, 
            d.license_number,
            d.consultation_fee as default_consultation_fee,
            d.follow_up_fee as default_follow_up_fee,
            d.follow_up_days as default_follow_up_days,
            d.profile_image,
            d.is_active as doctor_active,
            u.id as user_id,
            u.first_name, 
            u.last_name, 
            u.email, 
            u.username, 
            u.phone,
            u.is_active as user_active
        FROM clinic_doctor_links cdl
        JOIN doctors d ON d.id = cdl.doctor_id
        JOIN users u ON u.id = d.user_id
        WHERE cdl.clinic_id = $1 
            AND cdl.is_active = true 
            AND d.is_active = true 
            AND u.is_active = true
        ORDER BY u.first_name, u.last_name
    `

	rows, err := config.DB.Query(query, clinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctors for clinic", "details": err.Error()})
		return
	}
	defer rows.Close()

	var doctors []gin.H
	for rows.Next() {
		var linkID, doctorID, userID string
		var doctorCode, specialization, licenseNumber sql.NullString
		var firstName, lastName, email, username string
		var phone, notes *string
		var consultationFeeOffline, consultationFeeOnline, followUpFee *float64
		var defaultConsultationFee, defaultFollowUpFee *float64
		var followUpDays, defaultFollowUpDays *int
		var linkActive, doctorActive, userActive bool
		var profileImage sql.NullString

		err := rows.Scan(
			&linkID,
			&consultationFeeOffline, &consultationFeeOnline,
			&followUpFee, &followUpDays, &notes,
			&linkActive,
			&doctorID, &doctorCode, &specialization, &licenseNumber,
			&defaultConsultationFee, &defaultFollowUpFee, &defaultFollowUpDays, &profileImage,
			&doctorActive,
			&userID, &firstName, &lastName, &email, &username, &phone,
			&userActive,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan doctor data", "details": err.Error()})
			return
		}

		doctor := gin.H{
			"link_id":        linkID,
			"doctor_id":      doctorID,
			"user_id":        userID,
			"doctor_code":    doctorCode.String,
			"specialization": specialization.String,
			"license_number": licenseNumber.String,
			"profile_image":  profileImage.String,
			"first_name":     firstName,
			"last_name":      lastName,
			"full_name":      firstName + " " + lastName,
			"email":          email,
			"username":       username,
			"phone":          phone,
			"is_active":      linkActive && doctorActive && userActive,
			"clinic_specific_fees": gin.H{
				"consultation_fee_offline": consultationFeeOffline,
				"consultation_fee_online":  consultationFeeOnline,
				"follow_up_fee":            followUpFee,
				"follow_up_days":           followUpDays,
				"notes":                    notes,
			},
			"default_fees": gin.H{
				"consultation_fee": defaultConsultationFee,
				"follow_up_fee":    defaultFollowUpFee,
				"follow_up_days":   defaultFollowUpDays,
			},
		}

		doctors = append(doctors, doctor)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading doctor data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clinic_id":     clinicID,
		"doctors":       doctors,
		"total_doctors": len(doctors),
	})
}
