package controllers

import (
	"fmt"
	"net/http"
	"organization-service/config"
	"strings"

	"organization-service/middleware"

	"github.com/gin-gonic/gin"
)

// =====================================================
// DEPARTMENT MANAGEMENT APIs
// =====================================================

type DepartmentResponse struct {
	ID          string `json:"id"`
	ClinicID    string `json:"clinic_id"`
	ClinicName  string `json:"clinic_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateDepartmentInput struct {
	ClinicID    string  `json:"clinic_id" binding:"required,uuid"`
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Description *string `json:"description"`
}

type UpdateDepartmentInput struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

// CreateDepartment - Create a new department for a clinic
func CreateDepartment(c *gin.Context) {
	var input CreateDepartmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Verify clinic exists and is active
	var clinicExists bool
	err := config.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM clinics 
			WHERE id = $1 AND is_active = true
		)
	`, input.ClinicID).Scan(&clinicExists)

	if err != nil || !clinicExists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Clinic not found",
			"message": "Clinic not found or is inactive",
		})
		return
	}

	// Trim name for consistency
	input.Name = strings.TrimSpace(input.Name)

	// Check if department name already exists for this clinic (case-insensitive)
	var nameExists bool
	err = config.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM departments 
			WHERE clinic_id = $1 AND LOWER(name) = LOWER($2)
		)
	`, input.ClinicID, input.Name).Scan(&nameExists)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check department name")
		return
	}

	if nameExists {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Department name exists",
			"message": "A department with this name already exists in this clinic",
		})
		return
	}

	// Insert new department
	var departmentID string
	err = config.DB.QueryRow(`
		INSERT INTO departments (clinic_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`, input.ClinicID, input.Name, input.Description).Scan(&departmentID)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to create department")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Department created successfully",
		"department_id": departmentID,
		"department": gin.H{
			"id":          departmentID,
			"clinic_id":   input.ClinicID,
			"name":        input.Name,
			"description": input.Description,
			"is_active":   true,
		},
	})
}

// ListDepartments - List departments for a clinic
func ListDepartments(c *gin.Context) {
	clinicID := c.Query("clinic_id")
	onlyActive := c.DefaultQuery("only_active", "false") // Changed default to false to show all entered departments

	// If clinicID is not provided, use the clinic_id set by the middleware (if any)
	if clinicID == "" {
		clinicID = c.GetString("clinic_id")
	}

	// Build WHERE clause
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if clinicID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.clinic_id = $%d", argIndex))
		args = append(args, clinicID)
		argIndex++
	}

	if onlyActive == "true" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.is_active = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Query departments with clinic names - Use COALESCE to handle NULL descriptions/names
	query := fmt.Sprintf(`
		SELECT d.id, d.clinic_id, d.name, COALESCE(d.description, ''), d.is_active, 
		       d.created_at, d.updated_at, COALESCE(c.name, 'Unknown Clinic') as clinic_name
		FROM departments d
		LEFT JOIN clinics c ON c.id = d.clinic_id
		%s
		ORDER BY d.name ASC
	`, whereClause)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch departments")
		return
	}
	defer rows.Close()

	departments := []DepartmentResponse{}
	for rows.Next() {
		var dept DepartmentResponse
		var createdAt, updatedAt string

		err := rows.Scan(
			&dept.ID, &dept.ClinicID, &dept.Name, &dept.Description,
			&dept.IsActive, &createdAt, &updatedAt, &dept.ClinicName,
		)
		if err != nil {
			continue
		}

		dept.CreatedAt = createdAt
		dept.UpdatedAt = updatedAt
		departments = append(departments, dept)
	}

	c.JSON(http.StatusOK, gin.H{
		"departments": departments,
		"total_count": len(departments),
	})
}

// GetDepartment - Get single department details
func GetDepartment(c *gin.Context) {
	departmentID := c.Param("id")

	var dept DepartmentResponse
	var createdAt, updatedAt string

	err := config.DB.QueryRow(`
		SELECT d.id, d.clinic_id, d.name, COALESCE(d.description, ''), d.is_active,
		       d.created_at, d.updated_at, COALESCE(c.name, 'Unknown Clinic') as clinic_name
		FROM departments d
		LEFT JOIN clinics c ON c.id = d.clinic_id
		WHERE d.id = $1
	`, departmentID).Scan(
		&dept.ID, &dept.ClinicID, &dept.Name, &dept.Description,
		&dept.IsActive, &createdAt, &updatedAt, &dept.ClinicName,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Department not found",
				"message": "The specified department does not exist",
			})
			return
		}
		middleware.SendDatabaseError(c, "Failed to fetch department")
		return
	}

	dept.CreatedAt = createdAt
	dept.UpdatedAt = updatedAt

	c.JSON(http.StatusOK, dept)
}

// UpdateDepartment - Update department details
func UpdateDepartment(c *gin.Context) {
	departmentID := c.Param("id")
	clinicIDContext := c.GetString("clinic_id")

	var input UpdateDepartmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// 1. Get current department info for validation
	var currentClinicID, existingName string
	err := config.DB.QueryRow(`
		SELECT clinic_id, name FROM departments 
		WHERE id = $1 AND (clinic_id = $2 OR $2 = '')
	`, departmentID, clinicIDContext).Scan(&currentClinicID, &existingName)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Department not found",
			"message": "The specified department does not exist or you don't have permission to update it",
		})
		return
	}

	// 2. Build dynamic update query
	query := `UPDATE departments SET updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{}
	argIndex := 1

	if input.Name != nil {
		// Trim and check for name conflict only if name is changing (case-insensitive)
		*input.Name = strings.TrimSpace(*input.Name)
		if !strings.EqualFold(*input.Name, existingName) {
			var nameExists bool
			err = config.DB.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM departments 
					WHERE clinic_id = $1 AND LOWER(name) = LOWER($2) AND id != $3
				)
			`, currentClinicID, *input.Name, departmentID).Scan(&nameExists)

			if err != nil {
				middleware.SendDatabaseError(c, "Failed to check department name")
				return
			}

			if nameExists {
				c.JSON(http.StatusConflict, gin.H{
					"error":   "Department name exists",
					"message": "A department with this name already exists in this clinic",
				})
				return
			}
		}

		query += fmt.Sprintf(`, name = $%d`, argIndex)
		args = append(args, *input.Name)
		argIndex++
	}

	if input.Description != nil {
		query += fmt.Sprintf(`, description = $%d`, argIndex)
		args = append(args, *input.Description)
		argIndex++
	}

	if input.IsActive != nil {
		query += fmt.Sprintf(`, is_active = $%d`, argIndex)
		args = append(args, *input.IsActive)
		argIndex++
	}

	// Finalize query with ID and Clinic security check
	query += fmt.Sprintf(` WHERE id = $%d AND (clinic_id = $%d OR $%d = '')`, argIndex, argIndex+1, argIndex+1)
	args = append(args, departmentID, clinicIDContext)

	result, err := config.DB.Exec(query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to update department")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Department not found",
			"message": "The specified department does not exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Department updated successfully",
	})
}

// DeleteDepartment - Delete a department
func DeleteDepartment(c *gin.Context) {
	departmentID := c.Param("id")
	clinicIDContext := c.GetString("clinic_id")

	// 1. Verify the department simply exists (no scope here — DELETE handles scope)
	var exists bool
	err := config.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM departments WHERE id = $1)
	`, departmentID).Scan(&exists)

	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Department not found",
			"message": "The specified department does not exist",
		})
		return
	}

	// 2. Use a transaction to safely clear FK references before deleting
	tx, err := config.DB.Begin()
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Clear department_id from doctors (prevents FK violation)
	_, err = tx.Exec(`UPDATE doctors SET department_id = NULL WHERE department_id = $1`, departmentID)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to unassign doctors from department")
		return
	}

	// Clear department_id from clinic_doctor_links (prevents FK violation)
	_, err = tx.Exec(`UPDATE clinic_doctor_links SET department_id = NULL WHERE department_id = $1`, departmentID)
	if err != nil {
		// Ignore: column may not exist yet before migration 039 runs
		_ = err
	}

	// 3. Delete with clinic scoping:
	//    - super_admin: clinicIDContext is "" → ($2 = '') is true → deletes any
	//    - clinic_admin: clinicIDContext is set → must match clinic_id on the row
	result, err := tx.Exec(`
		DELETE FROM departments 
		WHERE id = $1 AND (clinic_id = $2 OR $2 = '')
	`, departmentID, clinicIDContext)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to delete department")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You do not have permission to delete this department",
		})
		return
	}

	if err := tx.Commit(); err != nil {
		middleware.SendDatabaseError(c, "Failed to commit deletion")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Department deleted successfully",
	})
}

// GetDoctorsByDepartment - Get all doctors belonging to a specific department
// Supports both:
//  1. Doctors directly assigned via doctors.department_id
//  2. Doctors linked via clinic_doctor_links.department_id
func GetDoctorsByDepartment(c *gin.Context) {
	departmentID := c.Param("id")
	onlyActive := c.DefaultQuery("only_active", "true")

	// Verify department exists
	var departmentName, clinicName string
	err := config.DB.QueryRow(`
		SELECT dept.name, COALESCE(c.name, 'Unknown Clinic')
		FROM departments dept
		LEFT JOIN clinics c ON c.id = dept.clinic_id
		WHERE dept.id = $1 AND dept.is_active = true
	`, departmentID).Scan(&departmentName, &clinicName)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Department not found",
			"message": "Department not found or is inactive",
		})
		return
	}

	activeFilter := ""
	if onlyActive == "true" {
		activeFilter = "AND d.is_active = true AND u.is_active = true"
	}

	// Query doctors in department via UNION:
	// 1. Doctors with department_id directly on the doctors row
	// 2. Doctors linked to a clinic in this department via clinic_doctor_links
	query := fmt.Sprintf(`
		SELECT DISTINCT d.id, d.user_id, COALESCE(d.doctor_code, ''), 
		       COALESCE(d.specialization, ''), COALESCE(d.license_number, ''),
		       d.is_main_doctor, d.is_active, d.created_at::text,
		       u.first_name, u.last_name, COALESCE(u.email::text, ''), COALESCE(u.phone, ''),
		       COALESCE(d.profile_image, '')
		FROM doctors d
		JOIN users u ON u.id = d.user_id
		WHERE d.department_id = $1
		%s

		UNION

		SELECT DISTINCT d.id, d.user_id, COALESCE(d.doctor_code, ''),
		       COALESCE(d.specialization, ''), COALESCE(d.license_number, ''),
		       d.is_main_doctor, d.is_active, d.created_at::text,
		       u.first_name, u.last_name, COALESCE(u.email::text, ''), COALESCE(u.phone, ''),
		       COALESCE(d.profile_image, '')
		FROM doctors d
		JOIN users u ON u.id = d.user_id
		JOIN clinic_doctor_links cdl ON cdl.doctor_id = d.id
		WHERE cdl.department_id = $1 AND cdl.is_active = true
		%s

		ORDER BY first_name, last_name
	`, activeFilter, activeFilter)

	rows, err := config.DB.Query(query, departmentID)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch doctors")
		return
	}
	defer rows.Close()

	doctors := []map[string]interface{}{}
	for rows.Next() {
		var doctorID, userID, doctorCode, specialization, licenseNumber string
		var firstName, lastName, email, phone, profileImage string
		var isMainDoctor, isActive bool
		var createdAt string

		err := rows.Scan(
			&doctorID, &userID, &doctorCode, &specialization, &licenseNumber,
			&isMainDoctor, &isActive, &createdAt,
			&firstName, &lastName, &email, &phone,
			&profileImage,
		)
		if err != nil {
			continue
		}

		doctor := map[string]interface{}{
			"id":             doctorID,
			"user_id":        userID,
			"doctor_code":    doctorCode,
			"specialization": specialization,
			"license_number": licenseNumber,
			"is_main_doctor": isMainDoctor,
			"is_active":      isActive,
			"created_at":     createdAt,
			"first_name":     firstName,
			"last_name":      lastName,
			"full_name":      firstName + " " + lastName,
			"email":          email,
			"phone":          phone,
			"profile_image":  profileImage,
		}

		doctors = append(doctors, doctor)
	}

	c.JSON(http.StatusOK, gin.H{
		"department_id":   departmentID,
		"department_name": departmentName,
		"clinic_name":     clinicName,
		"doctors":         doctors,
		"total_count":     len(doctors),
	})
}
