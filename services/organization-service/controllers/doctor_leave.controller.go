package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"organization-service/config"
	"strconv"
	"strings"
	"time"

	"organization-service/middleware"

	"github.com/gin-gonic/gin"
)

// =====================================================
// DOCTOR LEAVE MANAGEMENT APIs
// =====================================================

type DoctorLeaveResponse struct {
	ID             string     `json:"id"`
	DoctorID       string     `json:"doctor_id"`
	DoctorName     string     `json:"doctor_name"`
	ClinicID       string     `json:"clinic_id"`
	ClinicName     string     `json:"clinic_name"`
	LeaveType      string     `json:"leave_type"`
	LeaveDuration  string     `json:"leave_duration"` // morning, afternoon, full_day
	FromDate       string     `json:"from_date"`
	ToDate         string     `json:"to_date"`
	TotalDays      int        `json:"total_days"`
	Reason         string     `json:"reason"`
	Status         string     `json:"status"`
	AppliedAt      time.Time  `json:"applied_at"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy     *string    `json:"reviewed_by,omitempty"`
	ReviewedByName *string    `json:"reviewed_by_name,omitempty"`
	ReviewNotes    *string    `json:"review_notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ApplyLeaveInput represents the input for applying leave
type ApplyLeaveInput struct {
	DoctorID      string `json:"doctor_id" binding:"required"` // REQUIRED: Doctor ID for leave application
	ClinicID      string `json:"clinic_id" binding:"required"`
	LeaveType     string `json:"leave_type" binding:"required"`     // e.g., sick_leave, emergency, casual_leave, vacation
	LeaveDuration string `json:"leave_duration" binding:"required"` // full_day, morning, afternoon
	FromDate      string `json:"from_date" binding:"required"`      // YYYY-MM-DD
	ToDate        string `json:"to_date" binding:"required"`        // YYYY-MM-DD
	Reason        string `json:"reason" binding:"required,min=10,max=500"`
}

// ApplyLeave - Apply leave for a doctor (Doctor/Clinic Admin/Receptionist)
func ApplyLeave(c *gin.Context) {
	var input ApplyLeaveInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Validate leave type (reason)
	input.LeaveType = strings.ToLower(input.LeaveType)
	validLeaveTypes := map[string]bool{
		"sick_leave":   true,
		"vacation":     true,
		"emergency":    true,
		"other":        true,
		"casual_leave": true,
	}
	if !validLeaveTypes[input.LeaveType] {
		middleware.SendValidationError(c, "Invalid leave type",
			"Leave type (reason) must be: sick_leave, vacation, emergency, casual_leave, or other")
		return
	}

	// Validate leave duration
	input.LeaveDuration = strings.ToLower(input.LeaveDuration)
	validDurations := map[string]bool{
		"morning":   true,
		"afternoon": true,
		"full_day":  true,
	}
	if !validDurations[input.LeaveDuration] {
		middleware.SendValidationError(c, "Invalid leave duration",
			"Leave duration must be: morning, afternoon, or full_day")
		return
	}

	// Parse and validate dates
	fromDate, err := time.Parse("2006-01-02", input.FromDate)
	if err != nil {
		middleware.SendValidationError(c, "Invalid from_date", "Date must be in YYYY-MM-DD format")
		return
	}

	toDate, err := time.Parse("2006-01-02", input.ToDate)
	if err != nil {
		middleware.SendValidationError(c, "Invalid to_date", "Date must be in YYYY-MM-DD format")
		return
	}

	// Validate date range
	if toDate.Before(fromDate) {
		middleware.SendValidationError(c, "Invalid date range", "to_date must be after or equal to from_date")
		return
	}

	// Calculate total days
	totalDays := int(toDate.Sub(fromDate).Hours()/24) + 1

	// Verify doctor exists and is in this clinic
	var doctorExists bool
	err = config.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM doctors d
			WHERE d.id = $1 
			AND d.is_active = true
			AND (
				d.clinic_id = $2 
				OR d.id IN (
					SELECT doctor_id FROM clinic_doctor_links 
					WHERE clinic_id = $2 AND is_active = true
				)
			)
		)
	`, input.DoctorID, input.ClinicID).Scan(&doctorExists)

	if err != nil || !doctorExists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Doctor not found",
			"message": "Doctor not found in this clinic or is inactive",
		})
		return
	}

	// Check for overlapping leaves
	var overlappingCount int
	err = config.DB.QueryRow(`
		SELECT COUNT(*) FROM doctor_leaves
		WHERE doctor_id = $1
		AND clinic_id = $2
		AND status IN ('pending', 'approved')
		AND (
			(from_date <= $3 AND to_date >= $3) OR
			(from_date <= $4 AND to_date >= $4) OR
			(from_date >= $3 AND to_date <= $4)
		)
	`, input.DoctorID, input.ClinicID, input.FromDate, input.ToDate).Scan(&overlappingCount)

	if err == nil && overlappingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Overlapping leave exists",
			"message": "This doctor already has a leave application during this period",
		})
		return
	}

	// Insert leave application
	var leaveID string
	err = config.DB.QueryRow(`
		INSERT INTO doctor_leaves (doctor_id, clinic_id, leave_type, leave_duration, from_date, to_date, total_days, reason, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending')
		RETURNING id
	`, input.DoctorID, input.ClinicID, input.LeaveType, input.LeaveDuration, input.FromDate, input.ToDate, totalDays, input.Reason).Scan(&leaveID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to apply for leave")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Leave application submitted successfully",
		"leave_id":   leaveID,
		"status":     "pending",
		"total_days": totalDays,
	})
}

// ListDoctorLeaves - List leave applications with filters
// Middleware handles authentication
// Query params: clinic_id, doctor_id, status, leave_type, page, page_size
func ListDoctorLeaves(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	clinicID := c.Query("clinic_id")
	doctorID := c.Query("doctor_id")
	leaveType := c.Query("leave_type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build WHERE clause based on query parameters
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Filter by query parameters
	if status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("dl.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if clinicID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("dl.clinic_id = $%d", argIndex))
		args = append(args, clinicID)
		argIndex++
	}

	if doctorID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("dl.doctor_id = $%d", argIndex))
		args = append(args, doctorID)
		argIndex++
	}

	if leaveType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("dl.leave_type = $%d", argIndex))
		args = append(args, leaveType)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM doctor_leaves dl %s", whereClause)
	var totalCount int
	err := config.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to count leaves")
		return
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Main query
	query := fmt.Sprintf(`
		SELECT dl.id, dl.doctor_id, dl.clinic_id, dl.leave_type, dl.leave_duration, dl.from_date, dl.to_date,
		       dl.total_days, dl.reason, dl.status, dl.applied_at, dl.reviewed_at,
		       dl.reviewed_by, dl.review_notes, dl.created_at,
		       u.first_name || ' ' || u.last_name as doctor_name,
		       c.name as clinic_name,
		       COALESCE(reviewer.first_name || ' ' || reviewer.last_name, '') as reviewed_by_name
		FROM doctor_leaves dl
		JOIN doctors d ON d.id = dl.doctor_id
		JOIN users u ON u.id = d.user_id
		JOIN clinics c ON c.id = dl.clinic_id
		LEFT JOIN users reviewer ON reviewer.id = dl.reviewed_by
		%s
		ORDER BY dl.applied_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch leaves")
		return
	}
	defer rows.Close()

	leaves := []DoctorLeaveResponse{}
	for rows.Next() {
		var leave DoctorLeaveResponse
		var doctorName, clinicName, reviewedByName string

		err := rows.Scan(
			&leave.ID, &leave.DoctorID, &leave.ClinicID, &leave.LeaveType, &leave.LeaveDuration,
			&leave.FromDate, &leave.ToDate, &leave.TotalDays, &leave.Reason,
			&leave.Status, &leave.AppliedAt, &leave.ReviewedAt, &leave.ReviewedBy,
			&leave.ReviewNotes, &leave.CreatedAt, &doctorName, &clinicName, &reviewedByName,
		)
		if err != nil {
			continue
		}

		leave.DoctorName = doctorName
		leave.ClinicName = clinicName
		if reviewedByName != "" {
			leave.ReviewedByName = &reviewedByName
		}

		leaves = append(leaves, leave)
	}

	c.JSON(http.StatusOK, gin.H{
		"leaves": leaves,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + pageSize - 1) / pageSize,
		},
	})
}

// GetDoctorLeave - Get single leave application details
func GetDoctorLeave(c *gin.Context) {
	leaveID := c.Param("id")
	userID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")

	// Build access check
	accessCheck := ""
	args := []interface{}{leaveID}

	if !isSuperAdmin {
		// Check if user is the doctor who applied
		var userDoctorID string
		config.DB.QueryRow(`SELECT id FROM doctors WHERE user_id = $1`, userID).Scan(&userDoctorID)

		if userDoctorID != "" {
			accessCheck = " AND dl.doctor_id = $2"
			args = append(args, userDoctorID)
		} else {
			// Check if user has clinic access (admin/receptionist)
			accessCheck = ` AND dl.clinic_id IN (
				SELECT DISTINCT ur.clinic_id FROM user_roles ur
				WHERE ur.user_id = $2 AND ur.clinic_id IS NOT NULL AND ur.is_active = true
			)`
			args = append(args, userID)
		}
	}

	query := fmt.Sprintf(`
		SELECT dl.id, dl.doctor_id, dl.clinic_id, dl.leave_type, dl.leave_duration, dl.from_date, dl.to_date,
		       dl.total_days, dl.reason, dl.status, dl.applied_at, dl.reviewed_at,
		       dl.reviewed_by, dl.review_notes, dl.created_at,
		       u.first_name || ' ' || u.last_name as doctor_name,
		       c.name as clinic_name,
		       COALESCE(reviewer.first_name || ' ' || reviewer.last_name, '') as reviewed_by_name
		FROM doctor_leaves dl
		JOIN doctors d ON d.id = dl.doctor_id
		JOIN users u ON u.id = d.user_id
		JOIN clinics c ON c.id = dl.clinic_id
		LEFT JOIN users reviewer ON reviewer.id = dl.reviewed_by
		WHERE dl.id = $1%s
	`, accessCheck)

	var leave DoctorLeaveResponse
	var doctorName, clinicName, reviewedByName string

	err := config.DB.QueryRow(query, args...).Scan(
		&leave.ID, &leave.DoctorID, &leave.ClinicID, &leave.LeaveType, &leave.LeaveDuration,
		&leave.FromDate, &leave.ToDate, &leave.TotalDays, &leave.Reason,
		&leave.Status, &leave.AppliedAt, &leave.ReviewedAt, &leave.ReviewedBy,
		&leave.ReviewNotes, &leave.CreatedAt, &doctorName, &clinicName, &reviewedByName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Leave application")
			return
		}
		middleware.SendDatabaseError(c, "Failed to fetch leave details")
		return
	}

	leave.DoctorName = doctorName
	leave.ClinicName = clinicName
	if reviewedByName != "" {
		leave.ReviewedByName = &reviewedByName
	}

	c.JSON(http.StatusOK, leave)
}

// UpdateDoctorLeave - Update an existing leave application (Doctor/Admin)
func UpdateDoctorLeave(c *gin.Context) {
	leaveID := c.Param("id")
	userID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")

	var input ApplyLeaveInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Validate leave type (reason)
	input.LeaveType = strings.ToLower(input.LeaveType)
	validLeaveTypes := map[string]bool{
		"sick_leave":   true,
		"vacation":     true,
		"emergency":    true,
		"casual_leave": true,
		"other":        true,
	}
	if !validLeaveTypes[input.LeaveType] {
		middleware.SendValidationError(c, "Invalid leave type",
			"Leave type must be: sick_leave, vacation, emergency, casual_leave or other")
		return
	}

	// Validate leave duration
	input.LeaveDuration = strings.ToLower(input.LeaveDuration)
	validDurations := map[string]bool{
		"morning":   true,
		"afternoon": true,
		"full_day":  true,
	}
	if !validDurations[input.LeaveDuration] {
		middleware.SendValidationError(c, "Invalid leave duration",
			"Leave duration must be: morning, afternoon, or full_day")
		return
	}

	// Parse and validate dates
	fromDate, err := time.Parse("2006-01-02", input.FromDate)
	if err != nil {
		middleware.SendValidationError(c, "Invalid from_date", "Date must be in YYYY-MM-DD format")
		return
	}

	toDate, err := time.Parse("2006-01-02", input.ToDate)
	if err != nil {
		middleware.SendValidationError(c, "Invalid to_date", "Date must be in YYYY-MM-DD format")
		return
	}

	if toDate.Before(fromDate) {
		middleware.SendValidationError(c, "Invalid date range", "to_date must be after or equal to from_date")
		return
	}

	// Get existing leave details
	var doctorID, status string
	err = config.DB.QueryRow(`
		SELECT doctor_id, status FROM doctor_leaves WHERE id = $1
	`, leaveID).Scan(&doctorID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Leave application")
			return
		}
		middleware.SendDatabaseError(c, "Failed to fetch leave")
		return
	}

	// Can only update pending or approved leaves
	if status != "pending" && status != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Cannot update",
			"message": fmt.Sprintf("Only pending or approved leaves can be updated. Current status: %s", status),
		})
		return
	}

	// Check permissions
	if !isSuperAdmin {
		var userDoctorID string
		config.DB.QueryRow(`SELECT id FROM doctors WHERE user_id = $1`, userID).Scan(&userDoctorID)

		if userDoctorID != doctorID {
			var hasAccess bool
			config.DB.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM user_roles ur
					JOIN doctor_leaves dl ON dl.clinic_id = ur.clinic_id
					WHERE ur.user_id = $1 AND dl.id = $2 AND ur.is_active = true
				)
			`, userID, leaveID).Scan(&hasAccess)

			if !hasAccess {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Access denied",
					"message": "You don't have permission to update this leave",
				})
				return
			}
		}
	}

	// Check for overlapping leaves (excluding this one)
	var overlappingCount int
	err = config.DB.QueryRow(`
		SELECT COUNT(*) FROM doctor_leaves
		WHERE doctor_id = $1
		AND id != $2
		AND status IN ('pending', 'approved')
		AND (
			(from_date <= $3 AND to_date >= $3) OR
			(from_date <= $4 AND to_date >= $4) OR
			(from_date >= $3 AND to_date <= $4)
		)
	`, doctorID, leaveID, input.FromDate, input.ToDate).Scan(&overlappingCount)

	if err == nil && overlappingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Overlapping leave exists",
			"message": "This doctor already has an overlapping leave application during this period",
		})
		return
	}

	// Calculate total days
	totalDays := int(toDate.Sub(fromDate).Hours()/24) + 1

	// Update leave application
	_, err = config.DB.Exec(`
		UPDATE doctor_leaves 
		SET leave_type = $1, leave_duration = $2, from_date = $3, to_date = $4, total_days = $5, reason = $6, applied_at = CURRENT_TIMESTAMP
		WHERE id = $7
	`, input.LeaveType, input.LeaveDuration, input.FromDate, input.ToDate, totalDays, input.Reason, leaveID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to update leave application")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Leave application updated successfully",
		"status":  status,
	})
}

// ReviewLeaveInput represents the input for reviewing a leave
type ReviewLeaveInput struct {
	Status      string  `json:"status" binding:"required"` // approved or rejected
	ReviewNotes *string `json:"review_notes"`
}

// ReviewLeave - Clinic Admin/Receptionist approves or rejects leave
func ReviewLeave(c *gin.Context) {
	leaveID := c.Param("id")
	reviewerID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")

	var input ReviewLeaveInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Validate status
	if input.Status != "approved" && input.Status != "rejected" {
		middleware.SendValidationError(c, "Invalid status", "Status must be 'approved' or 'rejected'")
		return
	}

	// Get leave details
	var clinicID, currentStatus string
	err := config.DB.QueryRow(`
		SELECT clinic_id, status FROM doctor_leaves WHERE id = $1
	`, leaveID).Scan(&clinicID, &currentStatus)

	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Leave application")
			return
		}
		middleware.SendDatabaseError(c, "Failed to fetch leave")
		return
	}

	// Check if already reviewed
	if currentStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Leave already reviewed",
			"message": fmt.Sprintf("This leave has already been %s", currentStatus),
		})
		return
	}

	// Verify reviewer has access to this clinic (unless super admin)
	if !isSuperAdmin {
		var hasAccess bool
		err = config.DB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM user_roles ur
				JOIN roles r ON r.id = ur.role_id
				WHERE ur.user_id = $1
				AND ur.clinic_id = $2
				AND r.name IN ('clinic_admin', 'receptionist', 'organization_admin')
				AND ur.is_active = true
			)
		`, reviewerID, clinicID).Scan(&hasAccess)

		if err != nil || !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "You don't have permission to review leaves for this clinic",
			})
			return
		}
	}

	// Update leave status
	_, err = config.DB.Exec(`
		UPDATE doctor_leaves
		SET status = $1, reviewed_at = CURRENT_TIMESTAMP, reviewed_by = $2, review_notes = $3
		WHERE id = $4
	`, input.Status, reviewerID, input.ReviewNotes, leaveID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to review leave")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Leave %s successfully", input.Status),
		"status":  input.Status,
	})
}

// CancelLeave - Doctor cancels their own leave application
func CancelLeave(c *gin.Context) {
	leaveID := c.Param("id")
	userID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")

	// Get leave details
	var leaveDoctorID, clinicID, status string
	err := config.DB.QueryRow(`
		SELECT doctor_id, clinic_id, status FROM doctor_leaves WHERE id = $1
	`, leaveID).Scan(&leaveDoctorID, &clinicID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendNotFoundError(c, "Leave application")
			return
		}
		middleware.SendDatabaseError(c, "Failed to fetch leave")
		return
	}

	// Check permissions
	if !isSuperAdmin {
		var userDoctorID string
		config.DB.QueryRow(`SELECT id FROM doctors WHERE user_id = $1`, userID).Scan(&userDoctorID)

		// If user is a doctor, they can only cancel their own leaves
		if userDoctorID != "" {
			if userDoctorID != leaveDoctorID {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Access denied",
					"message": "You can only cancel your own leave applications",
				})
				return
			}
		} else {
			// If not a doctor, check if they are a clinic admin for this clinic
			var hasAccess bool
			config.DB.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM user_roles ur
					JOIN roles r ON r.id = ur.role_id
					WHERE ur.user_id = $1
					AND ur.clinic_id = $2
					AND r.name IN ('clinic_admin', 'organization_admin')
					AND ur.is_active = true
				)
			`, userID, clinicID).Scan(&hasAccess)

			if !hasAccess {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Access denied",
					"message": "You don't have permission to cancel leaves for this clinic",
				})
				return
			}
		}
	}

	// Can only cancel pending or approved leaves
	if status != "pending" && status != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Cannot cancel",
			"message": "Only pending or approved leaves can be cancelled",
		})
		return
	}

	// Update status to cancelled
	_, err = config.DB.Exec(`
		UPDATE doctor_leaves
		SET status = 'cancelled'
		WHERE id = $1
	`, leaveID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to cancel leave")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Leave cancelled successfully",
	})
}

// // GetDoctorsByClinic - Get all doctors in a clinic (simplified - middleware handles access)
// func GetDoctorsByClinic(c *gin.Context) {
// 	clinicID := c.Param("clinic_id")

// 	// Get clinic details first
// 	var clinicName, clinicCode string
// 	var clinicAddress sql.NullString
// 	err := config.DB.QueryRow(`
// 		SELECT name, clinic_code, address
// 		FROM clinics
// 		WHERE id = $1
// 	`, clinicID).Scan(&clinicName, &clinicCode, &clinicAddress)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			c.JSON(http.StatusNotFound, gin.H{
// 				"error":   "Clinic not found",
// 				"message": "The specified clinic does not exist",
// 			})
// 			return
// 		}
// 		middleware.SendDatabaseError(c, "Failed to fetch clinic details")
// 		return
// 	}

// 	// Fetch doctors - includes both direct clinic assignment AND clinic_doctor_links
// 	rows, err := config.DB.Query(`
// 		SELECT DISTINCT d.id, d.user_id, d.doctor_code, d.specialization, d.license_number,
// 		       d.consultation_fee, d.follow_up_fee, d.follow_up_days, d.is_main_doctor,
// 		       d.is_active, d.created_at,
// 		       u.first_name, u.last_name, u.email, u.phone
// 		FROM doctors d
// 		JOIN users u ON u.id = d.user_id
// 		WHERE d.is_active = true
// 		AND (
// 			d.clinic_id = $1  -- Direct clinic assignment
// 			OR
// 			d.id IN (         -- OR linked via clinic_doctor_links
// 				SELECT doctor_id FROM clinic_doctor_links
// 				WHERE clinic_id = $1 AND is_active = true
// 			)
// 		)
// 		ORDER BY d.created_at DESC
// 	`, clinicID)

// 	if err != nil {
// 		middleware.SendDatabaseError(c, "Failed to fetch doctors")
// 		return
// 	}
// 	defer rows.Close()

// 	doctors := []map[string]interface{}{}
// 	for rows.Next() {
// 		var doctor struct {
// 			ID              string
// 			UserID          string
// 			DoctorCode      sql.NullString
// 			Specialization  sql.NullString
// 			LicenseNumber   sql.NullString
// 			ConsultationFee sql.NullFloat64
// 			FollowUpFee     sql.NullFloat64
// 			FollowUpDays    sql.NullInt64
// 			IsMainDoctor    bool
// 			IsActive        bool
// 			CreatedAt       time.Time
// 			FirstName       string
// 			LastName        string
// 			Email           sql.NullString
// 			Phone           sql.NullString
// 		}

// 		err := rows.Scan(
// 			&doctor.ID, &doctor.UserID, &doctor.DoctorCode, &doctor.Specialization,
// 			&doctor.LicenseNumber, &doctor.ConsultationFee, &doctor.FollowUpFee,
// 			&doctor.FollowUpDays, &doctor.IsMainDoctor, &doctor.IsActive,
// 			&doctor.CreatedAt, &doctor.FirstName, &doctor.LastName, &doctor.Email, &doctor.Phone,
// 		)
// 		if err != nil {
// 			continue
// 		}

// 		doctorMap := map[string]interface{}{
// 			"id":             doctor.ID,
// 			"user_id":        doctor.UserID,
// 			"first_name":     doctor.FirstName,
// 			"last_name":      doctor.LastName,
// 			"is_main_doctor": doctor.IsMainDoctor,
// 			"is_active":      doctor.IsActive,
// 			"created_at":     doctor.CreatedAt,
// 		}

// 		if doctor.DoctorCode.Valid {
// 			doctorMap["doctor_code"] = doctor.DoctorCode.String
// 		}
// 		if doctor.Specialization.Valid {
// 			doctorMap["specialization"] = doctor.Specialization.String
// 		}
// 		if doctor.LicenseNumber.Valid {
// 			doctorMap["license_number"] = doctor.LicenseNumber.String
// 		}
// 		if doctor.ConsultationFee.Valid {
// 			doctorMap["consultation_fee"] = doctor.ConsultationFee.Float64
// 		}
// 		if doctor.FollowUpFee.Valid {
// 			doctorMap["follow_up_fee"] = doctor.FollowUpFee.Float64
// 		}
// 		if doctor.FollowUpDays.Valid {
// 			doctorMap["follow_up_days"] = doctor.FollowUpDays.Int64
// 		}
// 		if doctor.Email.Valid {
// 			doctorMap["email"] = doctor.Email.String
// 		}
// 		if doctor.Phone.Valid {
// 			doctorMap["phone"] = doctor.Phone.String
// 		}

// 		doctors = append(doctors, doctorMap)
// 	}

// 	// Build clinic info
// 	clinicInfo := gin.H{
// 		"id":          clinicID,
// 		"name":        clinicName,
// 		"clinic_code": clinicCode,
// 	}
// 	if clinicAddress.Valid {
// 		clinicInfo["address"] = clinicAddress.String
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"clinic":      clinicInfo,
// 		"doctors":     doctors,
// 		"total_count": len(doctors),
// 	})
// }

// GetDoctorLeaveStats - Get leave statistics for a doctor
func GetDoctorLeaveStats(c *gin.Context) {
	doctorID := c.Param("doctor_id")
	userID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")

	// Verify access
	if !isSuperAdmin {
		var userDoctorID string
		config.DB.QueryRow(`SELECT id FROM doctors WHERE user_id = $1`, userID).Scan(&userDoctorID)

		// Only the doctor themselves or clinic admin/receptionist can view stats
		if userDoctorID != doctorID {
			var hasAccess bool
			config.DB.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM user_roles ur
					JOIN roles r ON r.id = ur.role_id
					JOIN doctors d ON d.clinic_id = ur.clinic_id
					WHERE ur.user_id = $1
					AND d.id = $2
					AND r.name IN ('clinic_admin', 'receptionist', 'organization_admin')
					AND ur.is_active = true
				)
			`, userID, doctorID).Scan(&hasAccess)

			if !hasAccess {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Access denied",
					"message": "You don't have permission to view these statistics",
				})
				return
			}
		}
	}

	// Get statistics
	var stats struct {
		TotalLeaves       int
		PendingLeaves     int
		ApprovedLeaves    int
		RejectedLeaves    int
		CancelledLeaves   int
		TotalDaysThisYear int
	}

	config.DB.QueryRow(`
		SELECT 
			COUNT(*) as total_leaves,
			COUNT(*) FILTER (WHERE status = 'pending') as pending_leaves,
			COUNT(*) FILTER (WHERE status = 'approved') as approved_leaves,
			COUNT(*) FILTER (WHERE status = 'rejected') as rejected_leaves,
			COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled_leaves,
			COALESCE(SUM(total_days) FILTER (
				WHERE status = 'approved' 
				AND EXTRACT(YEAR FROM from_date) = EXTRACT(YEAR FROM CURRENT_DATE)
			), 0) as total_days_this_year
		FROM doctor_leaves
		WHERE doctor_id = $1
	`, doctorID).Scan(
		&stats.TotalLeaves, &stats.PendingLeaves, &stats.ApprovedLeaves,
		&stats.RejectedLeaves, &stats.CancelledLeaves, &stats.TotalDaysThisYear,
	)

	c.JSON(http.StatusOK, gin.H{
		"doctor_id":            doctorID,
		"total_leaves":         stats.TotalLeaves,
		"pending_leaves":       stats.PendingLeaves,
		"approved_leaves":      stats.ApprovedLeaves,
		"rejected_leaves":      stats.RejectedLeaves,
		"cancelled_leaves":     stats.CancelledLeaves,
		"total_days_this_year": stats.TotalDaysThisYear,
	})
}
