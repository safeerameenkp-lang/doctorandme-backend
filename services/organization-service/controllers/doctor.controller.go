package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"organization-service/config"
	"organization-service/middleware"
	"organization-service/models"
	"organization-service/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
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
	clinicID := c.Query("clinic_id")
	onlyActive := c.DefaultQuery("only_active", "true")
	isMain := c.Query("is_main")

	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if clinicID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(d.clinic_id = $%d OR EXISTS(SELECT 1 FROM clinic_doctor_links cdl WHERE cdl.doctor_id = d.id AND cdl.clinic_id = $%d))", argIndex, argIndex))
		args = append(args, clinicID)
		argIndex++
	}

	if onlyActive == "true" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.is_active = $%d AND u.is_active = $%d", argIndex, argIndex+1))
		args = append(args, true, true)
		argIndex += 2
	}

	if isMain == "true" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.is_main_doctor = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	query := fmt.Sprintf(`
        SELECT d.id, d.doctor_code, d.specialization, d.license_number, d.consultation_fee,
               d.follow_up_fee, d.follow_up_days, d.profile_image, u.id, u.first_name, u.last_name, u.email, u.username, u.phone
        FROM doctors d
        JOIN users u ON d.user_id = u.id
        %s
        ORDER BY u.first_name, u.last_name
    `, whereClause)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctors", "details": err.Error()})
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
			continue
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

		if consultationFee.Valid {
			doctor["consultation_fee"] = consultationFee.Float64
		}
		if followUpFee.Valid {
			doctor["follow_up_fee"] = followUpFee.Float64
		}

		doctors = append(doctors, doctor)
	}

	c.JSON(http.StatusOK, gin.H{
		"doctors":     doctors,
		"total_count": len(doctors),
	})
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
	// User fields
	FirstName *string `json:"first_name" form:"first_name"`
	LastName  *string `json:"last_name" form:"last_name"`
	Email     *string `json:"email" form:"email"`
	Phone     *string `json:"phone" form:"phone"`
	Username  *string `json:"username" form:"username"`
	Password  *string `json:"password" form:"password"`

	// Doctor fields
	DoctorCode      *string  `json:"doctor_code" form:"doctor_code"`
	Specialization  *string  `json:"specialization" form:"specialization"`
	LicenseNumber   *string  `json:"license_number" form:"license_number"`
	ConsultationFee *float64 `json:"consultation_fee" form:"consultation_fee"`
	FollowUpFee     *float64 `json:"follow_up_fee" form:"follow_up_fee"`
	FollowUpDays    *int     `json:"follow_up_days" form:"follow_up_days"`
	IsMainDoctor    *bool    `json:"is_main_doctor" form:"is_main_doctor"`
	IsActive        *bool    `json:"is_active" form:"is_active"`
}

func UpdateDoctor(c *gin.Context) {
	doctorID := c.Param("id")
	clinicIDContext := c.GetString("clinic_id")

	// 1. Check if doctor exists and get their user_id
	var userID string
	var currentClinicID sql.NullString
	err := config.DB.QueryRow(`SELECT user_id, clinic_id FROM doctors WHERE id = $1`, doctorID).Scan(&userID, &currentClinicID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}

	// 2. Permission check: if clinic_id is provided in context, ensure doctor belongs to it
	if clinicIDContext != "" {
		var linked bool
		err := config.DB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM doctors WHERE id = $1 AND (clinic_id = $2 OR $2 = '')
				UNION
				SELECT 1 FROM clinic_doctor_links WHERE doctor_id = $1 AND clinic_id = $2
			)
		`, doctorID, clinicIDContext).Scan(&linked)

		if err != nil || !linked {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "You do not have permission to update this doctor",
			})
			return
		}
	}

	var input UpdateDoctorInput
	// Support both JSON and Form (for image upload)
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle Profile Image Upload
	var profileImagePath *string
	fileHeader, err := c.FormFile("profile_image")
	if err == nil {
		if err := utils.ValidateImage(fileHeader); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image: " + err.Error()})
			return
		}

		file, err := fileHeader.Open()
		if err == nil {
			defer file.Close()
			savedPath, err := utils.SaveOptimizedImage(file, fileHeader.Filename, "doctors")
			if err == nil {
				profileImagePath = &savedPath
			}
		}
	}

	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// 3. Update User Table
	userUpdates := []string{}
	userArgs := []interface{}{}
	uIdx := 1

	if input.FirstName != nil {
		userUpdates = append(userUpdates, fmt.Sprintf("first_name = $%d", uIdx))
		userArgs = append(userArgs, *input.FirstName)
		uIdx++
	}
	if input.LastName != nil {
		userUpdates = append(userUpdates, fmt.Sprintf("last_name = $%d", uIdx))
		userArgs = append(userArgs, *input.LastName)
		uIdx++
	}
	if input.Email != nil {
		userUpdates = append(userUpdates, fmt.Sprintf("email = $%d", uIdx))
		userArgs = append(userArgs, *input.Email)
		uIdx++
	}
	if input.Phone != nil {
		userUpdates = append(userUpdates, fmt.Sprintf("phone = $%d", uIdx))
		userArgs = append(userArgs, *input.Phone)
		uIdx++
	}
	if input.Username != nil {
		userUpdates = append(userUpdates, fmt.Sprintf("username = $%d", uIdx))
		userArgs = append(userArgs, *input.Username)
		uIdx++
	}
	if input.Password != nil {
		passHash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		userUpdates = append(userUpdates, fmt.Sprintf("password_hash = $%d", uIdx))
		userArgs = append(userArgs, string(passHash))
		uIdx++
	}

	if len(userUpdates) > 0 {
		userQuery := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(userUpdates, ", "), uIdx)
		userArgs = append(userArgs, userID)
		if _, err := tx.Exec(userQuery, userArgs...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile", "details": err.Error()})
			return
		}
	}

	// 4. Update Doctor Table
	docUpdates := []string{}
	docArgs := []interface{}{}
	dIdx := 1

	if input.DoctorCode != nil {
		docUpdates = append(docUpdates, fmt.Sprintf("doctor_code = $%d", dIdx))
		docArgs = append(docArgs, *input.DoctorCode)
		dIdx++
	}
	if input.Specialization != nil {
		docUpdates = append(docUpdates, fmt.Sprintf("specialization = $%d", dIdx))
		docArgs = append(docArgs, *input.Specialization)
		dIdx++
	}
	if input.LicenseNumber != nil {
		docUpdates = append(docUpdates, fmt.Sprintf("license_number = $%d", dIdx))
		docArgs = append(docArgs, *input.LicenseNumber)
		dIdx++
	}
	if input.ConsultationFee != nil {
		docUpdates = append(docUpdates, fmt.Sprintf("consultation_fee = $%d", dIdx))
		docArgs = append(docArgs, *input.ConsultationFee)
		dIdx++
	}
	if input.FollowUpFee != nil {
		docUpdates = append(docUpdates, fmt.Sprintf("follow_up_fee = $%d", dIdx))
		docArgs = append(docArgs, *input.FollowUpFee)
		dIdx++
	}
	if input.FollowUpDays != nil {
		docUpdates = append(docUpdates, fmt.Sprintf("follow_up_days = $%d", dIdx))
		docArgs = append(docArgs, *input.FollowUpDays)
		dIdx++
	}
	if input.IsMainDoctor != nil {
		docUpdates = append(docUpdates, fmt.Sprintf("is_main_doctor = $%d", dIdx))
		docArgs = append(docArgs, *input.IsMainDoctor)
		dIdx++
	}
	if input.IsActive != nil {
		docUpdates = append(docUpdates, fmt.Sprintf("is_active = $%d", dIdx))
		docArgs = append(docArgs, *input.IsActive)
		dIdx++
	}
	if profileImagePath != nil {
		docUpdates = append(docUpdates, fmt.Sprintf("profile_image = $%d", dIdx))
		docArgs = append(docArgs, *profileImagePath)
		dIdx++
	}

	if len(docUpdates) > 0 {
		docQuery := fmt.Sprintf("UPDATE doctors SET %s WHERE id = $%d", strings.Join(docUpdates, ", "), dIdx)
		docArgs = append(docArgs, doctorID)
		if _, err := tx.Exec(docQuery, docArgs...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update doctor profile", "details": err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Doctor and user profiles updated successfully"})
}

func DeleteDoctor(c *gin.Context) {
	doctorID := c.Param("id")
	clinicIDContext := c.GetString("clinic_id")
	clinicIDs := c.GetStringSlice("clinic_ids")
	userRoles := c.GetStringSlice("user_roles")

	// Check if user is super_admin
	isSuperAdmin := false
	for _, role := range userRoles {
		if role == "super_admin" {
			isSuperAdmin = true
			break
		}
	}

	// 1. Get doctor and user information and verify existence/permission
	var userID string
	var linked bool

	// Query to check if doctor exists and if user has permission
	// Permission:
	// - Super Admin: Always allowed
	// - Clinic Admin: Must be linked to one of their clinic_ids
	query := `
		SELECT user_id, EXISTS(
			SELECT 1 FROM doctors d
			LEFT JOIN clinic_doctor_links cdl ON cdl.doctor_id = d.id
			WHERE d.id = $1 AND (
				$2 = true OR -- Super admin
				d.clinic_id = ANY($3::uuid[]) OR -- Direct ownership via clinic_id
				cdl.clinic_id = ANY($3::uuid[]) -- Linked via clinic_doctor_links
			)
		)
		FROM doctors WHERE id = $1
	`

	// Ensure we have at least one ID in the slice for ANY($3) to work if not super admin
	if len(clinicIDs) == 0 && clinicIDContext != "" {
		clinicIDs = []string{clinicIDContext}
	}

	err := config.DB.QueryRow(query, doctorID, isSuperAdmin, pq.Array(clinicIDs)).Scan(&userID, &linked)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
			return
		}
		middleware.SendDatabaseError(c, "Failed to verify doctor existence")
		return
	}

	if !linked {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You do not have permission to delete this doctor record",
		})
		return
	}

	// 2. Perform hard delete in a transaction
	tx, err := config.DB.Begin()
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Delete from user_roles first to avoid FK issues
	_, err = tx.Exec(`DELETE FROM user_roles WHERE user_id = $1`, userID)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to clean up user roles")
		return
	}

	// Delete from patients table if this user was also a patient
	_, err = tx.Exec(`DELETE FROM patients WHERE user_id = $1`, userID)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to clean up patient profile")
		return
	}

	// Delete doctor profile (cascades to clinic_links, schedules, etc.)
	_, err = tx.Exec(`DELETE FROM doctors WHERE id = $1`, doctorID)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to delete doctor profile")
		return
	}

	// Delete user account (resolving the issue where user record remained)
	_, err = tx.Exec(`DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to delete associated user account")
		return
	}

	if err := tx.Commit(); err != nil {
		middleware.SendDatabaseError(c, "Failed to commit deletion")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Doctor and associated user account deleted successfully",
	})
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
