package controllers

import (
	"appointment-service/config"
	"appointment-service/utils"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var locISTFollowUp *time.Location

func init() {
	var err error
	locISTFollowUp, err = time.LoadLocation("Asia/Kolkata")
	if err != nil {
		locISTFollowUp = time.FixedZone("IST", 5*3600+30*60)
	}
}

// =====================================================
// FOLLOW-UP ELIGIBILITY API
// Dedicated endpoints for checking follow-up eligibility
// =====================================================

// FollowUpEligibilityResponse represents the response structure
type FollowUpEligibilityResponse struct {
	Eligible       bool    `json:"eligible"`
	IsFree         bool    `json:"is_free"`
	Message        string  `json:"message"`
	ValidUntil     *string `json:"valid_until,omitempty"`
	DaysRemaining  *int    `json:"days_remaining,omitempty"`
	DoctorName     *string `json:"doctor_name,omitempty"`
	DepartmentName *string `json:"department_name,omitempty"`
}

// ActiveFollowUpItem represents a single active follow-up
type ActiveFollowUpItem struct {
	FollowUpID     string  `json:"followup_id"`
	DoctorID       string  `json:"doctor_id"`
	DoctorName     string  `json:"doctor_name"`
	DepartmentID   *string `json:"department_id,omitempty"`
	DepartmentName *string `json:"department_name,omitempty"`
	IsFree         bool    `json:"is_free"`
	ValidFrom      string  `json:"valid_from"`
	ValidUntil     string  `json:"valid_until"`
	DaysRemaining  int     `json:"days_remaining"`
	Message        string  `json:"message"`
}

// CheckFollowUpEligibility - Check if patient is eligible for follow-up with specific doctor+department
// GET /appointments/followup-eligibility?clinic_patient_id=xxx&clinic_id=xxx&doctor_id=xxx&department_id=xxx
func CheckFollowUpEligibility(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	clinicPatientID := c.Query("clinic_patient_id")
	clinicID := c.Query("clinic_id")
	doctorID := c.Query("doctor_id")
	departmentID := c.Query("department_id")

	if clinicPatientID == "" || clinicID == "" || doctorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clinic_patient_id, clinic_id, and doctor_id are required"})
		return
	}

	// 1. Follow-up Eligibility Logic
	// Expiration is handled by a background worker/cron job to keep the API fast.
	followUpMgr := &utils.FollowUpManager{DB: config.DB}
	_ = followUpMgr // keep for future use in transactional context if needed

	var response FollowUpEligibilityResponse
	response.Eligible = false
	response.IsFree = false
	response.Message = "No previous appointment found with this doctor"

	// 2. Check for active free follow-up WITH doctor/department details in one query
	var (
		fid, status string
		isFree      bool
		validUntil  time.Time
		docName     sql.NullString
		deptName    sql.NullString
	)

	query := `
		SELECT 
			f.id, f.status, f.is_free, f.valid_until,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
			dept.name as department_name
		FROM follow_ups f
		JOIN doctors d ON d.id = f.doctor_id
		JOIN users u ON u.id = d.user_id
		LEFT JOIN departments dept ON dept.id = f.department_id
		WHERE f.clinic_patient_id = $1 AND f.clinic_id = $2 AND f.doctor_id = $3
	`
	args := []interface{}{clinicPatientID, clinicID, doctorID}
	argIdx := 4
	if departmentID != "" {
		query += fmt.Sprintf(" AND f.department_id = $%d", argIdx)
		args = append(args, departmentID)
		argIdx++
	} else {
		query += ` AND f.department_id IS NULL`
	}
	query += ` ORDER BY f.created_at DESC LIMIT 1`

	now := time.Now().In(locISTFollowUp)
	err := config.DB.QueryRowContext(ctx, query, args...).Scan(&fid, &status, &isFree, &validUntil, &docName, &deptName)

	if err == nil {
		// Found a follow-up record
		if status == "active" && validUntil.After(now) {
			response.Eligible = true
			response.IsFree = isFree
			if isFree {
				daysRemaining := int(validUntil.Sub(now).Hours() / 24)
				if daysRemaining < 0 {
					daysRemaining = 0
				}
				response.Message = fmt.Sprintf("Free follow-up available (%d days remaining)", daysRemaining)
				vUntil := validUntil.Format("2006-01-02")
				response.ValidUntil = &vUntil
				response.DaysRemaining = &daysRemaining
			} else {
				response.Message = "Follow-up available (payment required)"
			}
		} else if status == "expired" || (status == "active" && validUntil.Before(now)) {
			response.Eligible = true
			response.Message = "Free follow-up expired. This follow-up requires payment."
		} else if status == "used" {
			response.Eligible = true
			response.Message = "Free follow-up already used. This follow-up requires payment."
		}

		if docName.Valid {
			response.DoctorName = &docName.String
		}
		if deptName.Valid {
			response.DepartmentName = &deptName.String
		}

		c.JSON(http.StatusOK, gin.H{"eligibility": response})
		return
	}

	// 3. If no follow-up record found, check for ANY completed appointment to allow paid follow-up
	var exists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM appointments
			WHERE clinic_patient_id = $1 AND clinic_id = $2 AND doctor_id = $3
			AND consultation_type IN ('clinic_visit', 'video_consultation')
			AND status IN ('completed', 'confirmed')
	`
	checkArgs := []interface{}{clinicPatientID, clinicID, doctorID}
	if departmentID != "" {
		checkQuery += ` AND department_id = $4`
		checkArgs = append(checkArgs, departmentID)
	}
	checkQuery += `)`

	err = config.DB.QueryRowContext(ctx, checkQuery, checkArgs...).Scan(&exists)
	if err == nil && exists {
		response.Eligible = true
		response.Message = "Follow-up available (payment required)"
	}

	c.JSON(http.StatusOK, gin.H{"eligibility": response})
}

// ListActiveFollowUps - Get all active follow-ups for a patient
// GET /appointments/followup-eligibility/active?clinic_patient_id=xxx&clinic_id=xxx
func ListActiveFollowUps(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	clinicPatientID := c.Query("clinic_patient_id")
	clinicID := c.Query("clinic_id")

	if clinicPatientID == "" || clinicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clinic_patient_id and clinic_id are required"})
		return
	}

	// Optimized query with JOINs to avoid N+1 problem
	query := `
		SELECT 
			f.id, f.doctor_id, 
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
			f.department_id, dept.name as department_name,
			f.is_free, f.valid_from, f.valid_until
		FROM follow_ups f
		JOIN doctors d ON d.id = f.doctor_id
		JOIN users u ON u.id = d.user_id
		LEFT JOIN departments dept ON dept.id = f.department_id
		WHERE f.clinic_patient_id = $1 AND f.clinic_id = $2 
		  AND f.status = 'active' AND f.valid_until >= CURRENT_DATE
		ORDER BY f.valid_until ASC
	`

	rows, err := config.DB.QueryContext(ctx, query, clinicPatientID, clinicID)
	if err != nil {
		log.Printf("ERROR: ListActiveFollowUps failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch active follow-ups"})
		return
	}
	defer rows.Close()

	activeFollowUps := make([]ActiveFollowUpItem, 0)
	now := time.Now().In(locISTFollowUp)
	for rows.Next() {
		var item ActiveFollowUpItem
		var validFrom, validUntil time.Time
		if err := rows.Scan(
			&item.FollowUpID, &item.DoctorID, &item.DoctorName,
			&item.DepartmentID, &item.DepartmentName,
			&item.IsFree, &validFrom, &validUntil,
		); err != nil {
			continue
		}

		daysRemaining := int(validUntil.Sub(now).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}

		item.ValidFrom = validFrom.Format("2006-01-02")
		item.ValidUntil = validUntil.Format("2006-01-02")
		item.DaysRemaining = daysRemaining
		item.Message = "Free follow-up available"
		if !item.IsFree {
			item.Message = "Follow-up available (payment required)"
		}

		activeFollowUps = append(activeFollowUps, item)
	}

	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Cursor error in ListActiveFollowUps: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":            len(activeFollowUps),
		"active_followups": activeFollowUps,
	})
}

// ExpireOldFollowUps - Manually trigger expiration of old follow-ups
func ExpireOldFollowUps(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	followUpMgr := &utils.FollowUpManager{DB: config.DB}
	count, err := followUpMgr.ExpireOldFollowUpsContext(ctx)
	if err != nil {
		log.Printf("ERROR: ExpireOldFollowUps failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to expire old follow-ups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Successfully expired old follow-ups",
		"expired_count": count,
	})
	_ = ctx // avoid unused
}
