package controllers

import (
	"auth-service/config"
	"auth-service/middleware"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// =====================================================
// CLINIC MANAGEMENT APIs (Role-Based Scoped)
// =====================================================

type ClinicResponse struct {
	ID               string    `json:"id"`
	OrganizationID   *string   `json:"organization_id"`
	OrganizationName *string   `json:"organization_name"`
	UserID           *string   `json:"user_id"`
	ClinicCode       string    `json:"clinic_code"`
	Name             string    `json:"name"`
	Email            *string   `json:"email"`
	Phone            *string   `json:"phone"`
	Address          *string   `json:"address"`
	LicenseNumber    *string   `json:"license_number"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	DoctorCount      int       `json:"doctor_count"`
	PatientCount     int       `json:"patient_count"`
	StaffCount       int       `json:"staff_count"`
}

// ListClinics - Role-based clinic listing
// Super Admin: ALL clinics
// Org Admin: Clinics in their organization
// Clinic Admin: Only their clinic
func ListClinics(c *gin.Context) {
	userID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	isActive := c.Query("is_active")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build WHERE clause based on role
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Scope filtering
	if !isSuperAdmin {
		if isOrgAdmin {
			orgIDs, _ := c.Get("organization_ids")
			if orgIDList, ok := orgIDs.([]string); ok && len(orgIDList) > 0 {
				placeholders := []string{}
				for _, orgID := range orgIDList {
					placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
					args = append(args, orgID)
					argIndex++
				}
				whereConditions = append(whereConditions,
					fmt.Sprintf("c.organization_id IN (%s)", strings.Join(placeholders, ",")))
			}
		} else if isClinicAdmin {
			clinicIDs, _ := c.Get("clinic_ids")
			if clinicIDList, ok := clinicIDs.([]string); ok && len(clinicIDList) > 0 {
				placeholders := []string{}
				for _, clinicID := range clinicIDList {
					placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
					args = append(args, clinicID)
					argIndex++
				}
				whereConditions = append(whereConditions,
					fmt.Sprintf("c.id IN (%s)", strings.Join(placeholders, ",")))
			}
		} else {
			// Regular staff - get their clinics from user_roles
			whereConditions = append(whereConditions,
				fmt.Sprintf(`c.id IN (
					SELECT DISTINCT ur.clinic_id FROM user_roles ur 
					WHERE ur.user_id = $%d AND ur.clinic_id IS NOT NULL AND ur.is_active = true
				)`, argIndex))
			args = append(args, userID)
			argIndex++
		}
	}

	// Search filter
	if search != "" {
		searchPattern := "%" + search + "%"
		whereConditions = append(whereConditions,
			fmt.Sprintf("(c.name ILIKE $%d OR c.clinic_code ILIKE $%d OR c.email ILIKE $%d)",
				argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	// Active filter
	if isActive != "" {
		if isActive == "true" {
			whereConditions = append(whereConditions, "c.is_active = true")
		} else if isActive == "false" {
			whereConditions = append(whereConditions, "c.is_active = false")
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM clinics c %s", whereClause)
	var totalCount int
	err := config.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to count clinics")
		return
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Main query
	query := fmt.Sprintf(`
		SELECT c.id, c.organization_id, c.user_id, c.clinic_code, c.name, c.email, 
		       c.phone, c.address, c.license_number, c.is_active, c.created_at,
		       o.name as organization_name,
		       (SELECT COUNT(*) FROM doctors d WHERE d.clinic_id = c.id AND d.is_active = true) as doctor_count,
		       (SELECT COUNT(DISTINCT pc.patient_id) FROM patient_clinics pc WHERE pc.clinic_id = c.id) as patient_count,
		       (SELECT COUNT(DISTINCT ur.user_id) FROM user_roles ur 
		        JOIN roles r ON r.id = ur.role_id 
		        WHERE ur.clinic_id = c.id AND ur.is_active = true 
		        AND r.name IN ('receptionist', 'pharmacist', 'lab_technician', 'billing_staff')) as staff_count
		FROM clinics c
		LEFT JOIN organizations o ON o.id = c.organization_id
		%s
		ORDER BY c.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch clinics")
		return
	}
	defer rows.Close()

	clinics := make([]ClinicResponse, 0, pageSize)
	for rows.Next() {
		var clinic ClinicResponse
		var orgName sql.NullString

		err := rows.Scan(
			&clinic.ID, &clinic.OrganizationID, &clinic.UserID, &clinic.ClinicCode,
			&clinic.Name, &clinic.Email, &clinic.Phone, &clinic.Address,
			&clinic.LicenseNumber, &clinic.IsActive, &clinic.CreatedAt,
			&orgName, &clinic.DoctorCount, &clinic.PatientCount, &clinic.StaffCount,
		)
		if err != nil {
			continue
		}

		if orgName.Valid {
			clinic.OrganizationName = &orgName.String
		}

		clinics = append(clinics, clinic)
	}

	if err = rows.Err(); err != nil {
		middleware.SendDatabaseError(c, "Failure processing clinics list")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clinics": clinics,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + pageSize - 1) / pageSize,
		},
		"scope": gin.H{
			"is_super_admin":        isSuperAdmin,
			"is_organization_admin": isOrgAdmin,
			"is_clinic_admin":       isClinicAdmin,
		},
	})
}

// =====================================================
// PATIENT MANAGEMENT APIs (Role-Based Scoped)
// =====================================================

type PatientResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Email          *string    `json:"email"`
	Phone          *string    `json:"phone"`
	DateOfBirth    *time.Time `json:"date_of_birth"`
	Gender         *string    `json:"gender"`
	BloodGroup     *string    `json:"blood_group"`
	MedicalHistory *string    `json:"medical_history"`
	Allergies      *string    `json:"allergies"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	Clinics        []string   `json:"clinics"`
	ClinicNames    []string   `json:"clinic_names"`
}

// ListPatients - Role-based patient listing
// Super Admin: ALL patients
// Org Admin: Patients in their organization's clinics
// Clinic Admin: Patients in their clinic
// Doctor/Staff: Patients in their clinic
func ListPatients(c *gin.Context) {
	userID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	clinicID := c.Query("clinic_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build WHERE clause based on role
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Scope filtering
	if !isSuperAdmin {
		if isOrgAdmin {
			orgIDs, _ := c.Get("organization_ids")
			if orgIDList, ok := orgIDs.([]string); ok && len(orgIDList) > 0 {
				placeholders := []string{}
				for _, orgID := range orgIDList {
					placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
					args = append(args, orgID)
					argIndex++
				}
				whereConditions = append(whereConditions,
					fmt.Sprintf(`p.id IN (
						SELECT DISTINCT pc.patient_id FROM patient_clinics pc
						JOIN clinics c ON c.id = pc.clinic_id
						WHERE c.organization_id IN (%s)
					)`, strings.Join(placeholders, ",")))
			}
		} else {
			// Clinic Admin or Staff - filter by their clinics
			clinicIDs, _ := c.Get("clinic_ids")
			var userClinicIDs []string

			if clinicIDList, ok := clinicIDs.([]string); ok && len(clinicIDList) > 0 {
				userClinicIDs = clinicIDList
			} else {
				// Regular staff - get their clinics
				rows, _ := config.DB.Query(`
					SELECT DISTINCT ur.clinic_id FROM user_roles ur 
					WHERE ur.user_id = $1 AND ur.clinic_id IS NOT NULL AND ur.is_active = true
				`, userID)
				for rows.Next() {
					var clinicID string
					rows.Scan(&clinicID)
					userClinicIDs = append(userClinicIDs, clinicID)
				}
				rows.Close()
			}

			if len(userClinicIDs) > 0 {
				placeholders := []string{}
				for _, cID := range userClinicIDs {
					placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
					args = append(args, cID)
					argIndex++
				}
				whereConditions = append(whereConditions,
					fmt.Sprintf(`p.id IN (
						SELECT DISTINCT pc.patient_id FROM patient_clinics pc
						WHERE pc.clinic_id IN (%s)
					)`, strings.Join(placeholders, ",")))
			} else {
				// No clinic access
				whereConditions = append(whereConditions, "1=0")
			}
		}
	}

	// Specific clinic filter
	if clinicID != "" {
		whereConditions = append(whereConditions,
			fmt.Sprintf(`p.id IN (
				SELECT pc.patient_id FROM patient_clinics pc WHERE pc.clinic_id = $%d
			)`, argIndex))
		args = append(args, clinicID)
		argIndex++
	}

	// Search filter
	if search != "" {
		searchPattern := "%" + search + "%"
		whereConditions = append(whereConditions,
			fmt.Sprintf("(u.first_name ILIKE $%d OR u.last_name ILIKE $%d OR u.email ILIKE $%d OR u.phone ILIKE $%d)",
				argIndex, argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM patients p
		JOIN users u ON u.id = p.user_id
		%s
	`, whereClause)
	var totalCount int
	err := config.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to count patients")
		return
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Main query
	query := fmt.Sprintf(`
		SELECT p.id, p.user_id, u.first_name, u.last_name, u.email, u.phone,
		       u.date_of_birth, u.gender, p.blood_group, p.medical_history, 
		       p.allergies, p.is_active, p.created_at
		FROM patients p
		JOIN users u ON u.id = p.user_id
		%s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch patients")
		return
	}
	defer rows.Close()

	patients := make([]PatientResponse, 0, pageSize)
	var patientIDs []string

	for rows.Next() {
		var patient PatientResponse

		err := rows.Scan(
			&patient.ID, &patient.UserID, &patient.FirstName, &patient.LastName,
			&patient.Email, &patient.Phone, &patient.DateOfBirth, &patient.Gender,
			&patient.BloodGroup, &patient.MedicalHistory, &patient.Allergies,
			&patient.IsActive, &patient.CreatedAt,
		)
		if err != nil {
			continue
		}

		patient.Clinics = make([]string, 0)
		patient.ClinicNames = make([]string, 0)
		patients = append(patients, patient)
		patientIDs = append(patientIDs, patient.ID)
	}

	if err = rows.Err(); err != nil {
		middleware.SendDatabaseError(c, "Error processing patients")
		return
	}

	// Optimize N+1 issue for patient clinics mapping
	if len(patientIDs) > 0 {
		placeholders := make([]string, len(patientIDs))
		pcArgs := make([]interface{}, len(patientIDs))
		for i, id := range patientIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			pcArgs[i] = id
		}

		pcQuery := fmt.Sprintf(`
			SELECT pc.patient_id, c.id, c.name 
			FROM clinics c
			JOIN patient_clinics pc ON pc.clinic_id = c.id
			WHERE pc.patient_id IN (%s)
		`, strings.Join(placeholders, ","))

		pcRows, err := config.DB.QueryContext(ctx, pcQuery, pcArgs...)
		if err == nil {
			defer pcRows.Close()

			clinicMap := make(map[string][]string)
			clinicNameMap := make(map[string][]string)

			for pcRows.Next() {
				var pID, cID, cName string
				if err := pcRows.Scan(&pID, &cID, &cName); err == nil {
					clinicMap[pID] = append(clinicMap[pID], cID)
					clinicNameMap[pID] = append(clinicNameMap[pID], cName)
				}
			}

			for i := range patients {
				if ids, exists := clinicMap[patients[i].ID]; exists {
					patients[i].Clinics = ids
					patients[i].ClinicNames = clinicNameMap[patients[i].ID]
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"patients": patients,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + pageSize - 1) / pageSize,
		},
		"scope": gin.H{
			"is_super_admin":        isSuperAdmin,
			"is_organization_admin": isOrgAdmin,
			"is_clinic_admin":       isClinicAdmin,
		},
	})
}

// =====================================================
// DOCTOR MANAGEMENT APIs (Role-Based Scoped)
// =====================================================

type DoctorResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	ClinicID        *string   `json:"clinic_id"`
	ClinicName      *string   `json:"clinic_name"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Email           *string   `json:"email"`
	Phone           *string   `json:"phone"`
	DoctorCode      *string   `json:"doctor_code"`
	Specialization  *string   `json:"specialization"`
	LicenseNumber   *string   `json:"license_number"`
	ConsultationFee *float64  `json:"consultation_fee"`
	FollowUpFee     *float64  `json:"follow_up_fee"`
	FollowUpDays    *int      `json:"follow_up_days"`
	IsMainDoctor    bool      `json:"is_main_doctor"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

// ListDoctors - Role-based doctor listing
// Super Admin: ALL doctors
// Org Admin: Doctors in their organization's clinics
// Clinic Admin: Doctors in their clinic
// Staff: Doctors in their clinic
func ListDoctors(c *gin.Context) {
	userID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	clinicID := c.Query("clinic_id")
	specialization := c.Query("specialization")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build WHERE clause based on role
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Scope filtering
	if !isSuperAdmin {
		if isOrgAdmin {
			orgIDs, _ := c.Get("organization_ids")
			if orgIDList, ok := orgIDs.([]string); ok && len(orgIDList) > 0 {
				placeholders := []string{}
				for _, orgID := range orgIDList {
					placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
					args = append(args, orgID)
					argIndex++
				}
				whereConditions = append(whereConditions,
					fmt.Sprintf(`d.clinic_id IN (
						SELECT id FROM clinics WHERE organization_id IN (%s)
					)`, strings.Join(placeholders, ",")))
			}
		} else {
			// Clinic Admin or Staff
			clinicIDs, _ := c.Get("clinic_ids")
			var userClinicIDs []string

			if clinicIDList, ok := clinicIDs.([]string); ok && len(clinicIDList) > 0 {
				userClinicIDs = clinicIDList
			} else {
				// Regular staff - get their clinics
				rows, _ := config.DB.Query(`
					SELECT DISTINCT ur.clinic_id FROM user_roles ur 
					WHERE ur.user_id = $1 AND ur.clinic_id IS NOT NULL AND ur.is_active = true
				`, userID)
				for rows.Next() {
					var cID string
					rows.Scan(&cID)
					userClinicIDs = append(userClinicIDs, cID)
				}
				rows.Close()
			}

			if len(userClinicIDs) > 0 {
				placeholders := []string{}
				for _, cID := range userClinicIDs {
					placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
					args = append(args, cID)
					argIndex++
				}
				whereConditions = append(whereConditions,
					fmt.Sprintf("d.clinic_id IN (%s)", strings.Join(placeholders, ",")))
			} else {
				whereConditions = append(whereConditions, "1=0")
			}
		}
	}

	// Specific clinic filter
	if clinicID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.clinic_id = $%d", argIndex))
		args = append(args, clinicID)
		argIndex++
	}

	// Specialization filter
	if specialization != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.specialization ILIKE $%d", argIndex))
		args = append(args, "%"+specialization+"%")
		argIndex++
	}

	// Search filter
	if search != "" {
		searchPattern := "%" + search + "%"
		whereConditions = append(whereConditions,
			fmt.Sprintf("(u.first_name ILIKE $%d OR u.last_name ILIKE $%d OR d.doctor_code ILIKE $%d OR d.specialization ILIKE $%d)",
				argIndex, argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM doctors d
		JOIN users u ON u.id = d.user_id
		%s
	`, whereClause)
	var totalCount int
	err := config.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to count doctors")
		return
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Main query
	query := fmt.Sprintf(`
		SELECT d.id, d.user_id, d.clinic_id, d.doctor_code, d.specialization,
		       d.license_number, d.consultation_fee, d.follow_up_fee, d.follow_up_days,
		       d.is_main_doctor, d.is_active, d.created_at,
		       u.first_name, u.last_name, u.email, u.phone,
		       c.name as clinic_name
		FROM doctors d
		JOIN users u ON u.id = d.user_id
		LEFT JOIN clinics c ON c.id = d.clinic_id
		%s
		ORDER BY d.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch doctors")
		return
	}
	defer rows.Close()

	doctors := make([]DoctorResponse, 0, pageSize)
	for rows.Next() {
		var doctor DoctorResponse
		var clinicName sql.NullString

		err := rows.Scan(
			&doctor.ID, &doctor.UserID, &doctor.ClinicID, &doctor.DoctorCode,
			&doctor.Specialization, &doctor.LicenseNumber, &doctor.ConsultationFee,
			&doctor.FollowUpFee, &doctor.FollowUpDays, &doctor.IsMainDoctor,
			&doctor.IsActive, &doctor.CreatedAt, &doctor.FirstName, &doctor.LastName,
			&doctor.Email, &doctor.Phone, &clinicName,
		)
		if err != nil {
			continue
		}

		if clinicName.Valid {
			doctor.ClinicName = &clinicName.String
		}

		doctors = append(doctors, doctor)
	}

	if err = rows.Err(); err != nil {
		middleware.SendDatabaseError(c, "Failure processing doctors listing")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"doctors": doctors,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + pageSize - 1) / pageSize,
		},
		"scope": gin.H{
			"is_super_admin":        isSuperAdmin,
			"is_organization_admin": isOrgAdmin,
			"is_clinic_admin":       isClinicAdmin,
		},
	})
}

// =====================================================
// STAFF MANAGEMENT APIs (Role-Based Scoped)
// =====================================================

type StaffResponse struct {
	ID         string    `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      *string   `json:"email"`
	Phone      *string   `json:"phone"`
	Role       string    `json:"role"`
	ClinicID   *string   `json:"clinic_id"`
	ClinicName *string   `json:"clinic_name"`
	IsActive   bool      `json:"is_active"`
	AssignedAt time.Time `json:"assigned_at"`
}

// ListStaff - Role-based staff listing (receptionists, pharmacists, lab technicians, billing staff)
// Super Admin: ALL staff
// Org Admin: Staff in their organization's clinics
// Clinic Admin: Staff in their clinic
// Staff: Staff in their clinic
func ListStaff(c *gin.Context) {
	userID := c.GetString("user_id")
	isSuperAdmin := c.GetBool("is_super_admin")
	isOrgAdmin := c.GetBool("is_organization_admin")
	isClinicAdmin := c.GetBool("is_clinic_admin")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	clinicID := c.Query("clinic_id")
	roleFilter := c.Query("role") // receptionist, pharmacist, lab_technician, billing_staff

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build WHERE clause
	whereConditions := []string{
		"r.name IN ('receptionist', 'pharmacist', 'lab_technician', 'billing_staff')",
	}
	args := []interface{}{}
	argIndex := 1

	// Role filter
	if roleFilter != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("r.name = $%d", argIndex))
		args = append(args, roleFilter)
		argIndex++
	}

	// Scope filtering
	if !isSuperAdmin {
		if isOrgAdmin {
			orgIDs, _ := c.Get("organization_ids")
			if orgIDList, ok := orgIDs.([]string); ok && len(orgIDList) > 0 {
				placeholders := []string{}
				for _, orgID := range orgIDList {
					placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
					args = append(args, orgID)
					argIndex++
				}
				whereConditions = append(whereConditions,
					fmt.Sprintf(`ur.clinic_id IN (
						SELECT id FROM clinics WHERE organization_id IN (%s)
					)`, strings.Join(placeholders, ",")))
			}
		} else {
			// Clinic Admin or Staff
			clinicIDs, _ := c.Get("clinic_ids")
			var userClinicIDs []string

			if clinicIDList, ok := clinicIDs.([]string); ok && len(clinicIDList) > 0 {
				userClinicIDs = clinicIDList
			} else {
				// Regular staff - get their clinics
				rows, _ := config.DB.Query(`
					SELECT DISTINCT ur.clinic_id FROM user_roles ur 
					WHERE ur.user_id = $1 AND ur.clinic_id IS NOT NULL AND ur.is_active = true
				`, userID)
				for rows.Next() {
					var cID string
					rows.Scan(&cID)
					userClinicIDs = append(userClinicIDs, cID)
				}
				rows.Close()
			}

			if len(userClinicIDs) > 0 {
				placeholders := []string{}
				for _, cID := range userClinicIDs {
					placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
					args = append(args, cID)
					argIndex++
				}
				whereConditions = append(whereConditions,
					fmt.Sprintf("ur.clinic_id IN (%s)", strings.Join(placeholders, ",")))
			} else {
				whereConditions = append(whereConditions, "1=0")
			}
		}
	}

	// Specific clinic filter
	if clinicID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("ur.clinic_id = $%d", argIndex))
		args = append(args, clinicID)
		argIndex++
	}

	// Search filter
	if search != "" {
		searchPattern := "%" + search + "%"
		whereConditions = append(whereConditions,
			fmt.Sprintf("(u.first_name ILIKE $%d OR u.last_name ILIKE $%d OR u.email ILIKE $%d)",
				argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(whereConditions, " AND ")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM user_roles ur
		JOIN users u ON u.id = ur.user_id
		JOIN roles r ON r.id = ur.role_id
		%s AND ur.is_active = true
	`, whereClause)
	var totalCount int
	err := config.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to count staff")
		return
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Main query
	query := fmt.Sprintf(`
		SELECT u.id, u.first_name, u.last_name, u.email, u.phone,
		       r.name as role, ur.clinic_id, ur.assigned_at, ur.is_active,
		       c.name as clinic_name
		FROM user_roles ur
		JOIN users u ON u.id = ur.user_id
		JOIN roles r ON r.id = ur.role_id
		LEFT JOIN clinics c ON c.id = ur.clinic_id
		%s AND ur.is_active = true
		ORDER BY ur.assigned_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch staff")
		return
	}
	defer rows.Close()

	staff := make([]StaffResponse, 0, pageSize)
	for rows.Next() {
		var s StaffResponse
		var clinicName sql.NullString

		err := rows.Scan(
			&s.ID, &s.FirstName, &s.LastName, &s.Email, &s.Phone,
			&s.Role, &s.ClinicID, &s.AssignedAt, &s.IsActive, &clinicName,
		)
		if err != nil {
			continue
		}

		if clinicName.Valid {
			s.ClinicName = &clinicName.String
		}

		staff = append(staff, s)
	}

	if err = rows.Err(); err != nil {
		middleware.SendDatabaseError(c, "Failure processing staff listing")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"staff": staff,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total_count": totalCount,
			"total_pages": (totalCount + pageSize - 1) / pageSize,
		},
		"scope": gin.H{
			"is_super_admin":        isSuperAdmin,
			"is_organization_admin": isOrgAdmin,
			"is_clinic_admin":       isClinicAdmin,
		},
	})
}
