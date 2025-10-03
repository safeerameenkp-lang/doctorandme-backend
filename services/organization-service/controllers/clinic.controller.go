package controllers

import (
    "organization-service/config"
    "organization-service/models"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "net/http"
    "regexp"
    "strconv"
    "shared-security"
)

// Clinic Controllers
type CreateClinicInput struct {
    OrganizationID string  `json:"organization_id" binding:"required,uuid"`
    UserID         string  `json:"user_id" binding:"required,uuid"`
    ClinicCode     string  `json:"clinic_code" binding:"required,min=2,max=20"`
    Name           string  `json:"name" binding:"required,min=2,max=255"`
    Email          *string `json:"email" binding:"omitempty,email"`
    Phone          *string `json:"phone" binding:"omitempty,len=10"`
    Address        *string `json:"address" binding:"omitempty,max=500"`
    LicenseNumber  *string `json:"license_number" binding:"omitempty,max=100"`
}

func CreateClinic(c *gin.Context) {
    var input CreateClinicInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Verify organization exists
    var orgExists bool
    err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)`, input.OrganizationID).Scan(&orgExists)
    if err != nil || !orgExists {
        security.SendNotFoundError(c, "organization")
        return
    }

    // Verify user exists
    var userExists bool
    err = config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, input.UserID).Scan(&userExists)
    if err != nil || !userExists {
        security.SendNotFoundError(c, "user")
        return
    }

    var clinicID string
    err = config.DB.QueryRow(`
        INSERT INTO clinics (organization_id, user_id, clinic_code, name, email, phone, address, license_number)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id
    `, input.OrganizationID, input.UserID, input.ClinicCode, input.Name, input.Email, input.Phone, input.Address, input.LicenseNumber).Scan(&clinicID)
    
    if err != nil {
        security.SendDatabaseError(c, "Failed to create clinic")
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": clinicID, "message": "Clinic created successfully"})
}

func GetClinics(c *gin.Context) {
    orgID := c.Query("organization_id")
    
    var query string
    var args []interface{}
    
    if orgID != "" {
        query = `
            SELECT id, organization_id, clinic_code, name, email, phone, address, license_number, is_active, created_at
            FROM clinics WHERE organization_id = $1 ORDER BY created_at DESC
        `
        args = []interface{}{orgID}
    } else {
        query = `
            SELECT id, organization_id, clinic_code, name, email, phone, address, license_number, is_active, created_at
            FROM clinics ORDER BY created_at DESC
        `
    }

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clinics"})
        return
    }
    defer rows.Close()

    var clinics []models.Clinic
    for rows.Next() {
        var clinic models.Clinic
        err := rows.Scan(&clinic.ID, &clinic.OrganizationID, &clinic.ClinicCode, &clinic.Name, &clinic.Email, &clinic.Phone, &clinic.Address, &clinic.LicenseNumber, &clinic.IsActive, &clinic.CreatedAt)
        if err != nil {
            continue
        }
        clinics = append(clinics, clinic)
    }

    c.JSON(http.StatusOK, clinics)
}

func GetClinic(c *gin.Context) {
    clinicID := c.Param("id")
    
    var clinic models.Clinic
    err := config.DB.QueryRow(`
        SELECT id, organization_id, clinic_code, name, email, phone, address, license_number, is_active, created_at
        FROM clinics WHERE id = $1
    `, clinicID).Scan(&clinic.ID, &clinic.OrganizationID, &clinic.ClinicCode, &clinic.Name, &clinic.Email, &clinic.Phone, &clinic.Address, &clinic.LicenseNumber, &clinic.IsActive, &clinic.CreatedAt)
    
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Clinic not found"})
        return
    }

    c.JSON(http.StatusOK, clinic)
}

type UpdateClinicInput struct {
    ClinicCode    *string `json:"clinic_code" binding:"omitempty,min=2,max=20"`
    Name          *string `json:"name" binding:"omitempty,min=2,max=255"`
    Email         *string `json:"email" binding:"omitempty,email"`
    Phone         *string `json:"phone" binding:"omitempty,len=10"`
    Address       *string `json:"address" binding:"omitempty,max=500"`
    LicenseNumber *string `json:"license_number" binding:"omitempty,max=100"`
    IsActive      *bool   `json:"is_active"`
}

func UpdateClinic(c *gin.Context) {
    clinicID := c.Param("id")
    var input UpdateClinicInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Build dynamic update query
    query := "UPDATE clinics SET "
    args := []interface{}{}
    argIndex := 1

    if input.ClinicCode != nil {
        query += "clinic_code = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.ClinicCode)
        argIndex++
    }
    if input.Name != nil {
        query += "name = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.Name)
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
    query = query[:len(query)-2] + " WHERE id = $" + strconv.Itoa(argIndex)
    args = append(args, clinicID)

    result, err := config.DB.Exec(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update clinic"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Clinic not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Clinic updated successfully"})
}

func DeleteClinic(c *gin.Context) {
    clinicID := c.Param("id")
    
    result, err := config.DB.Exec(`DELETE FROM clinics WHERE id = $1`, clinicID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete clinic"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Clinic not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Clinic deleted successfully"})
}

// Create clinic admin when creating clinic
type CreateClinicWithAdminInput struct {
    OrganizationID string  `json:"organization_id" binding:"required,uuid"`
    ClinicCode     string  `json:"clinic_code" binding:"required,min=2,max=20"`
    Name           string  `json:"name" binding:"required,min=2,max=255"`
    Email          *string `json:"email" binding:"omitempty,email"`
    Phone          *string `json:"phone" binding:"omitempty,len=10"`
    Address        *string `json:"address" binding:"omitempty,max=500"`
    LicenseNumber  *string `json:"license_number" binding:"omitempty,max=100"`
    // Admin details
    AdminFirstName string `json:"admin_first_name" binding:"required,min=2,max=50"`
    AdminLastName  string `json:"admin_last_name" binding:"required,min=2,max=50"`
    AdminEmail     string `json:"admin_email" binding:"required,email"`
    AdminUsername  string `json:"admin_username" binding:"required,min=3,max=30"`
    AdminPhone     string `json:"admin_phone" binding:"omitempty,len=10"`
    AdminPassword  string `json:"admin_password" binding:"required,min=8"`
}

func CreateClinicWithAdmin(c *gin.Context) {
    var input CreateClinicWithAdminInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate admin email format
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(input.AdminEmail) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin email format"})
        return
    }

    // Validate admin phone format if provided
    if input.AdminPhone != "" {
        phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
        if !phoneRegex.MatchString(input.AdminPhone) {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin phone format"})
            return
        }
    }

    // Verify organization exists
    var orgExists bool
    err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)`, input.OrganizationID).Scan(&orgExists)
    if err != nil || !orgExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Organization not found"})
        return
    }

    // Check if admin username already exists
    var existingUserID string
    err = config.DB.QueryRow(`SELECT id FROM users WHERE username = $1`, input.AdminUsername).Scan(&existingUserID)
    if err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Admin username already exists"})
        return
    }

    // Check if admin email already exists
    err = config.DB.QueryRow(`SELECT id FROM users WHERE email = $1`, input.AdminEmail).Scan(&existingUserID)
    if err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Admin email already exists"})
        return
    }

    // Start transaction
    tx, err := config.DB.Begin()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
        return
    }
    defer tx.Rollback()

    // Hash admin password
    passHash, err := bcrypt.GenerateFromPassword([]byte(input.AdminPassword), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash admin password"})
        return
    }

    // Create admin user
    var adminID string
    err = tx.QueryRow(`
        INSERT INTO users (first_name, last_name, email, username, phone, password_hash)
        VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
    `, input.AdminFirstName, input.AdminLastName, input.AdminEmail, input.AdminUsername, input.AdminPhone, string(passHash)).Scan(&adminID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user"})
        return
    }

    // Create clinic with admin as user_id
    var clinicID string
    err = tx.QueryRow(`
        INSERT INTO clinics (organization_id, user_id, clinic_code, name, email, phone, address, license_number)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id
    `, input.OrganizationID, adminID, input.ClinicCode, input.Name, input.Email, input.Phone, input.Address, input.LicenseNumber).Scan(&clinicID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create clinic"})
        return
    }

    // Assign clinic_admin role
    var roleID string
    err = tx.QueryRow(`SELECT id FROM roles WHERE name='clinic_admin' LIMIT 1`).Scan(&roleID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find clinic_admin role"})
        return
    }

    _, err = tx.Exec(`INSERT INTO user_roles (user_id, role_id, clinic_id) VALUES ($1,$2,$3)`, adminID, roleID, clinicID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign clinic admin role"})
        return
    }

    // Commit transaction
    if err = tx.Commit(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "clinic": gin.H{
            "id": clinicID,
            "organization_id": input.OrganizationID,
            "clinic_code": input.ClinicCode,
            "name": input.Name,
            "email": input.Email,
            "phone": input.Phone,
            "address": input.Address,
            "license_number": input.LicenseNumber,
        },
        "admin": gin.H{
            "id": adminID,
            "first_name": input.AdminFirstName,
            "last_name": input.AdminLastName,
            "email": input.AdminEmail,
            "username": input.AdminUsername,
            "phone": input.AdminPhone,
            "role": "clinic_admin",
        },
        "message": "Clinic and admin created successfully",
    })
}
