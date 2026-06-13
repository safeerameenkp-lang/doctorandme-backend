package controllers

import (
	"mime/multipart"
	"net/http"
	"organization-service/config"
	"organization-service/middleware"
	"organization-service/models"
	"organization-service/services"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CreatePharmacy handles creating a new pharmacy with logo support
func CreatePharmacy(c *gin.Context) {
	// 1. Bind form fields
	var input models.CreatePharmacyInput
	if err := c.ShouldBind(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// 2. Handle file upload logic
	fileHeader, err := c.FormFile("logo")
	var fileHeaderPtr *multipart.FileHeader
	if err == nil {
		fileHeaderPtr = fileHeader
	}

	// 3. Call Service
	pharmacySvc := services.NewPharmacyService(config.DB)
	pharmacyID, logoPath, err := pharmacySvc.CreatePharmacy(c.Request.Context(), input, fileHeaderPtr)
	if err != nil {
		// Differentiate errors (not found, validation, server)
		if err.Error() == "organization not found" || err.Error() == "user not found" || err.Error() == "clinic not found" {
			middleware.SendNotFoundError(c, err.Error())
		} else if err.Error() == "file size exceeds 5MB limit" || err.Error() == "unsupported file format. Allowed formats: JPG, JPEG, PNG" {
			middleware.SendValidationError(c, "Image validation failed", err.Error())
		} else {
			middleware.SendDatabaseError(c, "Failed to create pharmacy: "+err.Error())
		}
		return
	}

	// Return response with logo path
	c.JSON(http.StatusCreated, gin.H{
		"id":      pharmacyID,
		"logo":    logoPath,
		"message": "Pharmacy created successfully",
	})
}

// CreatePharmacyWithAdmin handles creating pharmacy + admin with logo
func CreatePharmacyWithAdmin(c *gin.Context) {
	var input models.CreatePharmacyWithAdminInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.AdminFirstName == "" {
		input.AdminFirstName = input.AdminUsername
	}
	if input.AdminLastName == "" {
		input.AdminLastName = input.AdminUsername
	}

	fileHeader, err := c.FormFile("logo")
	var fileHeaderPtr *multipart.FileHeader
	if err == nil {
		fileHeaderPtr = fileHeader
	}

	pharmacySvc := services.NewPharmacyService(config.DB)
	pharmacyID, adminID, logoPath, err := pharmacySvc.CreatePharmacyWithAdmin(c.Request.Context(), input, fileHeaderPtr)
	if err != nil {
		if err.Error() == "organization not found" || err.Error() == "clinic not found" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if err.Error() == "admin username already exists" || err.Error() == "admin email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"pharmacy": gin.H{
			"id":              pharmacyID,
			"organization_id": input.OrganizationID,
			"clinic_id":       input.ClinicID,
			"pharmacy_code":   input.PharmacyCode,
			"name":            input.Name,
			"pharmacy_type":   input.PharmacyType,
			"email":           input.Email,
			"phone":           input.Phone,
			"address":         input.Address,
			"license_number":  input.LicenseNumber,
			"logo":            logoPath,
		},
		"admin": gin.H{
			"id":         adminID,
			"first_name": input.AdminFirstName,
			"last_name":  input.AdminLastName,
			"email":      input.AdminEmail,
			"username":   input.AdminUsername,
			"phone":      input.AdminPhone,
			"role":       "pharmacy_admin",
		},
		"message": "Pharmacy and admin created successfully",
	})
}

// GetPharmacies lists all pharmacies, optionally filtered by organization_id or clinic_id
func GetPharmacies(c *gin.Context) {
	orgID := c.Query("organization_id")
	clinicID := c.Query("clinic_id")

	var query string
	var args []interface{}
	argIndex := 1

	query = `
        SELECT id, organization_id, clinic_id, pharmacy_code, name, pharmacy_type, email, phone, address, license_number, logo, is_active, created_at, updated_at
        FROM pharmacies
    `

	whereClauses := []string{}
	if orgID != "" {
		whereClauses = append(whereClauses, "organization_id = $"+strconv.Itoa(argIndex))
		args = append(args, orgID)
		argIndex++
	}
	if clinicID != "" {
		whereClauses = append(whereClauses, "clinic_id = $"+strconv.Itoa(argIndex))
		args = append(args, clinicID)
		argIndex++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY created_at DESC"

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pharmacies"})
		return
	}
	defer rows.Close()

	var pharmacies []models.Pharmacy = []models.Pharmacy{}
	for rows.Next() {
		var p models.Pharmacy
		err := rows.Scan(&p.ID, &p.OrganizationID, &p.ClinicID, &p.PharmacyCode, &p.Name, &p.PharmacyType, &p.Email, &p.Phone, &p.Address, &p.LicenseNumber, &p.Logo, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		pharmacies = append(pharmacies, p)
	}

	c.JSON(http.StatusOK, pharmacies)
}

// GetPharmacy fetches a single pharmacy by ID
func GetPharmacy(c *gin.Context) {
	pharmacyID := c.Param("id")

	var p models.Pharmacy
	err := config.DB.QueryRow(`
        SELECT id, organization_id, clinic_id, pharmacy_code, name, pharmacy_type, email, phone, address, license_number, logo, is_active, created_at, updated_at
        FROM pharmacies WHERE id = $1
    `, pharmacyID).Scan(&p.ID, &p.OrganizationID, &p.ClinicID, &p.PharmacyCode, &p.Name, &p.PharmacyType, &p.Email, &p.Phone, &p.Address, &p.LicenseNumber, &p.Logo, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pharmacy not found"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// UpdatePharmacy updates pharmacy fields dynamically
func UpdatePharmacy(c *gin.Context) {
	pharmacyID := c.Param("id")
	var input models.UpdatePharmacyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build dynamic update query
	query := "UPDATE pharmacies SET "
	args := []interface{}{}
	argIndex := 1

	if input.ClinicID != nil {
		query += "clinic_id = $" + strconv.Itoa(argIndex) + ", "
		if *input.ClinicID == "" {
			args = append(args, nil)
		} else {
			args = append(args, *input.ClinicID)
		}
		argIndex++
	}
	if input.PharmacyCode != nil {
		query += "pharmacy_code = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.PharmacyCode)
		argIndex++
	}
	if input.Name != nil {
		query += "name = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.Name)
		argIndex++
	}
	if input.PharmacyType != nil {
		query += "pharmacy_type = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.PharmacyType)
		argIndex++
	}
	if input.Email != nil {
		query += "email = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.Email)
		argIndex++
	}
	if input.Phone != nil {
		query += "phone = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.Phone)
		argIndex++
	}
	if input.Address != nil {
		query += "address = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.Address)
		argIndex++
	}
	if input.LicenseNumber != nil {
		query += "license_number = $" + strconv.Itoa(argIndex) + ", "
		args = append(args, *input.LicenseNumber)
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
	query = query[:len(query)-2] + ", updated_at = CURRENT_TIMESTAMP WHERE id = $" + strconv.Itoa(argIndex)
	args = append(args, pharmacyID)

	result, err := config.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pharmacy: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pharmacy not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pharmacy updated successfully"})
}

// DeletePharmacy deletes pharmacy and associated admin user if they are not linked elsewhere
func DeletePharmacy(c *gin.Context) {
	pharmacyID := c.Param("id")

	pharmacySvc := services.NewPharmacyService(config.DB)
	err := pharmacySvc.DeletePharmacy(c.Request.Context(), pharmacyID)

	if err != nil {
		if err.Error() == "pharmacy not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pharmacy not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pharmacy and its admin: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pharmacy and its associated admin deleted successfully"})
}
