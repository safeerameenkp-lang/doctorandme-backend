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

	// If clinicID is not provided, try to get it from the user's context
	if clinicID == "" {
		// Check if user is super_admin (roles are pre-populated by RequireRole middleware)
		isSuperAdmin := false
		roles := c.GetStringSlice("user_roles")
		for _, role := range roles {
			if role == "super_admin" {
				isSuperAdmin = true
				break
			}
		}

		if !isSuperAdmin {
			// If not super admin, restrict to their assigned clinic
			clinicID = c.GetString("clinic_id")
		}
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

	var input UpdateDepartmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Build dynamic update query
	query := `UPDATE departments SET updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{}
	argIndex := 1

	if input.Name != nil {
		// Check if new name conflicts with existing department in same clinic
		var clinicID string
		var existingName string
		err := config.DB.QueryRow(`
			SELECT clinic_id, name FROM departments WHERE id = $1
		`, departmentID).Scan(&clinicID, &existingName)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Department not found",
				"message": "The specified department does not exist",
			})
			return
		}

		// Trim and check for name conflict only if name is changing (case-insensitive)
		*input.Name = strings.TrimSpace(*input.Name)
		if !strings.EqualFold(*input.Name, existingName) {
			var nameExists bool
			err = config.DB.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM departments 
					WHERE clinic_id = $1 AND LOWER(name) = LOWER($2) AND id != $3
				)
			`, clinicID, *input.Name, departmentID).Scan(&nameExists)

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

	query += fmt.Sprintf(` WHERE id = $%d`, argIndex)
	args = append(args, departmentID)

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

	// Check if department has any doctors assigned
	var doctorCount int
	err := config.DB.QueryRow(`
		SELECT COUNT(*) FROM doctors WHERE department_id = $1
	`, departmentID).Scan(&doctorCount)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to check department usage")
		return
	}

	if doctorCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Cannot delete department",
			"message": fmt.Sprintf("Department has %d doctors assigned. Please reassign doctors before deleting.", doctorCount),
		})
		return
	}

	result, err := config.DB.Exec(`DELETE FROM departments WHERE id = $1`, departmentID)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to delete department")
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
		"message": "Department deleted successfully",
	})
}

// GetDoctorsByDepartment - Get doctors in a specific department
func GetDoctorsByDepartment(c *gin.Context) {
	departmentID := c.Param("department_id")
	onlyActive := c.DefaultQuery("only_active", "true")

	// Verify department exists
	var departmentExists bool
	err := config.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM departments 
			WHERE id = $1 AND is_active = true
		)
	`, departmentID).Scan(&departmentExists)

	if err != nil || !departmentExists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Department not found",
			"message": "Department not found or is inactive",
		})
		return
	}

	// Build WHERE clause
	whereConditions := []string{"d.department_id = $1"}
	args := []interface{}{departmentID}
	argIndex := 2

	if onlyActive == "true" {
		whereConditions = append(whereConditions, fmt.Sprintf("d.is_active = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(whereConditions, " AND ")

	// Query doctors in department
	query := fmt.Sprintf(`
		SELECT d.id, d.user_id, d.doctor_code, COALESCE(d.specialization, ''), COALESCE(d.license_number, ''),
		       d.is_main_doctor, d.is_active, d.created_at,
		       u.first_name, u.last_name, u.email, COALESCE(u.phone, ''),
		       dept.name as department_name, COALESCE(c.name, 'Unknown Clinic') as clinic_name
		FROM doctors d
		JOIN users u ON u.id = d.user_id
		JOIN departments dept ON dept.id = d.department_id
		LEFT JOIN clinics c ON c.id = dept.clinic_id
		%s
		ORDER BY d.created_at DESC
	`, whereClause)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch doctors")
		return
	}
	defer rows.Close()

	doctors := []map[string]interface{}{}
	for rows.Next() {
		var doctorID, userID, doctorCode, specialization, licenseNumber string
		var firstName, lastName, email, phone, departmentName, clinicName string
		var isMainDoctor, isActive bool
		var createdAt string

		err := rows.Scan(
			&doctorID, &userID, &doctorCode, &specialization, &licenseNumber,
			&isMainDoctor, &isActive, &createdAt,
			&firstName, &lastName, &email, &phone,
			&departmentName, &clinicName,
		)
		if err != nil {
			continue
		}

		doctor := map[string]interface{}{
			"id":              doctorID,
			"user_id":         userID,
			"doctor_code":     doctorCode,
			"specialization":  specialization,
			"license_number":  licenseNumber,
			"is_main_doctor":  isMainDoctor,
			"is_active":       isActive,
			"created_at":      createdAt,
			"first_name":      firstName,
			"last_name":       lastName,
			"email":           email,
			"phone":           phone,
			"department_name": departmentName,
			"clinic_name":     clinicName,
		}

		doctors = append(doctors, doctor)
	}

	c.JSON(http.StatusOK, gin.H{
		"department_id": departmentID,
		"doctors":       doctors,
		"total_count":   len(doctors),
	})
}
