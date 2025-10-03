package controllers

import (
    "organization-service/config"
    "organization-service/models"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "net/http"
    "regexp"
    "strings"
    "time"
    "fmt"
    "shared-security"
)

// HealthCheck endpoint
func HealthCheck(c *gin.Context) {
    // Test database connection
    err := config.DB.Ping()
    if err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "unhealthy",
            "error": "Database connection failed",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "service": "organization-service",
        "timestamp": time.Now().Unix(),
    })
}

// Organization Controllers
type CreateOrganizationInput struct {
    Name          string  `json:"name" binding:"required,min=2,max=255"`
    Email         *string `json:"email" binding:"omitempty,email"`
    Phone         *string `json:"phone" binding:"omitempty,len=10"`
    Address       *string `json:"address" binding:"omitempty,max=500"`
    LicenseNumber *string `json:"license_number" binding:"omitempty,max=100"`
}

func CreateOrganization(c *gin.Context) {
    var input CreateOrganizationInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    var orgID string
    err := config.DB.QueryRow(`
        INSERT INTO organizations (name, email, phone, address, license_number)
        VALUES ($1, $2, $3, $4, $5) RETURNING id
    `, input.Name, input.Email, input.Phone, input.Address, input.LicenseNumber).Scan(&orgID)
    
    if err != nil {
        security.SendDatabaseError(c, "Failed to create organization")
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": orgID, "message": "Organization created successfully"})
}

func GetOrganizations(c *gin.Context) {
    rows, err := config.DB.Query(`
        SELECT id, name, email, phone, address, license_number, is_active, created_at
        FROM organizations ORDER BY created_at DESC
    `)
    if err != nil {
        security.SendDatabaseError(c, "Failed to fetch organizations")
        return
    }
    defer rows.Close()

    var organizations []models.Organization
    for rows.Next() {
        var org models.Organization
        err := rows.Scan(&org.ID, &org.Name, &org.Email, &org.Phone, &org.Address, &org.LicenseNumber, &org.IsActive, &org.CreatedAt)
        if err != nil {
            continue
        }
        organizations = append(organizations, org)
    }

    c.JSON(http.StatusOK, organizations)
}

func GetOrganization(c *gin.Context) {
    orgID := c.Param("id")
    
    var org models.Organization
    err := config.DB.QueryRow(`
        SELECT id, name, email, phone, address, license_number, is_active, created_at
        FROM organizations WHERE id = $1
    `, orgID).Scan(&org.ID, &org.Name, &org.Email, &org.Phone, &org.Address, &org.LicenseNumber, &org.IsActive, &org.CreatedAt)
    
    if err != nil {
        security.SendNotFoundError(c, "organization")
        return
    }

    c.JSON(http.StatusOK, org)
}

type UpdateOrganizationInput struct {
    Name          *string `json:"name" binding:"omitempty,min=2,max=255"`
    Email         *string `json:"email" binding:"omitempty,email"`
    Phone         *string `json:"phone" binding:"omitempty,len=10"`
    Address       *string `json:"address" binding:"omitempty,max=500"`
    LicenseNumber *string `json:"license_number" binding:"omitempty,max=100"`
    IsActive      *bool   `json:"is_active"`
}

func UpdateOrganization(c *gin.Context) {
    orgID := c.Param("id")
    var input UpdateOrganizationInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Build dynamic update query safely
    query := "UPDATE organizations SET "
    args := []interface{}{}
    argIndex := 1
    updates := []string{}

    if input.Name != nil {
        updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
        args = append(args, *input.Name)
        argIndex++
    }
    if input.Email != nil {
        updates = append(updates, fmt.Sprintf("email = $%d", argIndex))
        args = append(args, *input.Email)
        argIndex++
    }
    if input.Phone != nil {
        updates = append(updates, fmt.Sprintf("phone = $%d", argIndex))
        args = append(args, *input.Phone)
        argIndex++
    }
    if input.Address != nil {
        updates = append(updates, fmt.Sprintf("address = $%d", argIndex))
        args = append(args, *input.Address)
        argIndex++
    }
    if input.LicenseNumber != nil {
        updates = append(updates, fmt.Sprintf("license_number = $%d", argIndex))
        args = append(args, *input.LicenseNumber)
        argIndex++
    }
    if input.IsActive != nil {
        updates = append(updates, fmt.Sprintf("is_active = $%d", argIndex))
        args = append(args, *input.IsActive)
        argIndex++
    }

    if len(updates) == 0 {
        security.SendValidationError(c, "No fields to update", "At least one field must be provided for update")
        return
    }

    query += strings.Join(updates, ", ")
    query += fmt.Sprintf(" WHERE id = $%d", argIndex)
    args = append(args, orgID)

    result, err := config.DB.Exec(query, args...)
    if err != nil {
        security.SendDatabaseError(c, "Failed to update organization")
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        security.SendNotFoundError(c, "organization")
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Organization updated successfully"})
}

func DeleteOrganization(c *gin.Context) {
    orgID := c.Param("id")
    
    result, err := config.DB.Exec(`DELETE FROM organizations WHERE id = $1`, orgID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete organization"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Organization deleted successfully"})
}

// Create organization admin when creating organization
type CreateOrganizationWithAdminInput struct {
    Name          string `json:"name" binding:"required,min=2,max=255"`
    Email         *string `json:"email" binding:"omitempty,email"`
    Phone         *string `json:"phone" binding:"omitempty,len=10"`
    Address       *string `json:"address" binding:"omitempty,max=500"`
    LicenseNumber *string `json:"license_number" binding:"omitempty,max=100"`
    // Admin details
    AdminFirstName string `json:"admin_first_name" binding:"required,min=2,max=50"`
    AdminLastName  string `json:"admin_last_name" binding:"required,min=2,max=50"`
    AdminEmail     string `json:"admin_email" binding:"required,email"`
    AdminUsername  string `json:"admin_username" binding:"required,min=3,max=30"`
    AdminPhone     string `json:"admin_phone" binding:"omitempty,len=10"`
    AdminPassword  string `json:"admin_password" binding:"required,min=8"`
}

func CreateOrganizationWithAdmin(c *gin.Context) {
    var input CreateOrganizationWithAdminInput
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

    // Check if admin username already exists
    var existingUserID string
    err := config.DB.QueryRow(`SELECT id FROM users WHERE username = $1`, input.AdminUsername).Scan(&existingUserID)
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

    // Create organization
    var orgID string
    err = tx.QueryRow(`
        INSERT INTO organizations (name, email, phone, address, license_number)
        VALUES ($1, $2, $3, $4, $5) RETURNING id
    `, input.Name, input.Email, input.Phone, input.Address, input.LicenseNumber).Scan(&orgID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization"})
        return
    }

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

    // Assign organization_admin role
    var roleID string
    err = tx.QueryRow(`SELECT id FROM roles WHERE name='organization_admin' LIMIT 1`).Scan(&roleID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find organization_admin role"})
        return
    }

    _, err = tx.Exec(`INSERT INTO user_roles (user_id, role_id, organization_id) VALUES ($1,$2,$3)`, adminID, roleID, orgID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign organization admin role"})
        return
    }

    // Commit transaction
    if err = tx.Commit(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "organization": gin.H{
            "id": orgID,
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
            "role": "organization_admin",
        },
        "message": "Organization and admin created successfully",
    })
}
