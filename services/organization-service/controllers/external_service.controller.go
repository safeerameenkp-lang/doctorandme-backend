package controllers

import (
    "organization-service/config"
    "organization-service/models"
    "github.com/gin-gonic/gin"
    "net/http"
    "strconv"
)

// External Service Controllers
type CreateExternalServiceInput struct {
    ServiceCode   string  `json:"service_code" binding:"required,min=2,max=20"`
    Name          string  `json:"name" binding:"required,min=2,max=255"`
    ServiceType   string  `json:"service_type" binding:"required,oneof=lab pharmacy"`
    Email         *string `json:"email" binding:"omitempty,email"`
    Phone         *string `json:"phone" binding:"omitempty,len=10"`
    Address       *string `json:"address" binding:"omitempty,max=500"`
    LicenseNumber *string `json:"license_number" binding:"omitempty,max=100"`
}

func CreateExternalService(c *gin.Context) {
    var input CreateExternalServiceInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var serviceID string
    err := config.DB.QueryRow(`
        INSERT INTO external_services (service_code, name, service_type, email, phone, address, license_number)
        VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
    `, input.ServiceCode, input.Name, input.ServiceType, input.Email, input.Phone, input.Address, input.LicenseNumber).Scan(&serviceID)
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create external service"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": serviceID, "message": "External service created successfully"})
}

func GetExternalServices(c *gin.Context) {
    serviceType := c.Query("service_type")
    
    var query string
    var args []interface{}
    
    if serviceType != "" {
        query = `
            SELECT id, service_code, name, service_type, email, phone, address, license_number, is_active, created_at
            FROM external_services WHERE service_type = $1 ORDER BY created_at DESC
        `
        args = []interface{}{serviceType}
    } else {
        query = `
            SELECT id, service_code, name, service_type, email, phone, address, license_number, is_active, created_at
            FROM external_services ORDER BY created_at DESC
        `
    }

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch external services"})
        return
    }
    defer rows.Close()

    var services []models.ExternalService
    for rows.Next() {
        var service models.ExternalService
        err := rows.Scan(&service.ID, &service.ServiceCode, &service.Name, &service.ServiceType, &service.Email, &service.Phone, &service.Address, &service.LicenseNumber, &service.IsActive, &service.CreatedAt)
        if err != nil {
            continue
        }
        services = append(services, service)
    }

    c.JSON(http.StatusOK, services)
}

func GetExternalService(c *gin.Context) {
    serviceID := c.Param("id")
    
    var service models.ExternalService
    err := config.DB.QueryRow(`
        SELECT id, service_code, name, service_type, email, phone, address, license_number, is_active, created_at
        FROM external_services WHERE id = $1
    `, serviceID).Scan(&service.ID, &service.ServiceCode, &service.Name, &service.ServiceType, &service.Email, &service.Phone, &service.Address, &service.LicenseNumber, &service.IsActive, &service.CreatedAt)
    
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "External service not found"})
        return
    }

    c.JSON(http.StatusOK, service)
}

type UpdateExternalServiceInput struct {
    ServiceCode   *string `json:"service_code" binding:"omitempty,min=2,max=20"`
    Name          *string `json:"name" binding:"omitempty,min=2,max=255"`
    ServiceType   *string `json:"service_type" binding:"omitempty,oneof=lab pharmacy"`
    Email         *string `json:"email" binding:"omitempty,email"`
    Phone         *string `json:"phone" binding:"omitempty,len=10"`
    Address       *string `json:"address" binding:"omitempty,max=500"`
    LicenseNumber *string `json:"license_number" binding:"omitempty,max=100"`
    IsActive      *bool   `json:"is_active"`
}

func UpdateExternalService(c *gin.Context) {
    serviceID := c.Param("id")
    var input UpdateExternalServiceInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Build dynamic update query
    query := "UPDATE external_services SET "
    args := []interface{}{}
    argIndex := 1

    if input.ServiceCode != nil {
        query += "service_code = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.ServiceCode)
        argIndex++
    }
    if input.Name != nil {
        query += "name = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.Name)
        argIndex++
    }
    if input.ServiceType != nil {
        query += "service_type = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.ServiceType)
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
    args = append(args, serviceID)

    result, err := config.DB.Exec(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update external service"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "External service not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "External service updated successfully"})
}

func DeleteExternalService(c *gin.Context) {
    serviceID := c.Param("id")
    
    result, err := config.DB.Exec(`DELETE FROM external_services WHERE id = $1`, serviceID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete external service"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "External service not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "External service deleted successfully"})
}
