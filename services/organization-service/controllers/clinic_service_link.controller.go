package controllers

import (
    "organization-service/config"
    "organization-service/models"
    "github.com/gin-gonic/gin"
    "net/http"
)

// Clinic Service Link Controllers
type CreateClinicServiceLinkInput struct {
    ClinicID  string `json:"clinic_id" binding:"required,uuid"`
    ServiceID string `json:"service_id" binding:"required,uuid"`
    IsDefault bool   `json:"is_default"`
}

func CreateClinicServiceLink(c *gin.Context) {
    var input CreateClinicServiceLinkInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Verify clinic and service exist
    var clinicExists, serviceExists bool
    err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM clinics WHERE id = $1)`, input.ClinicID).Scan(&clinicExists)
    if err != nil || !clinicExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Clinic not found"})
        return
    }

    err = config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM external_services WHERE id = $1)`, input.ServiceID).Scan(&serviceExists)
    if err != nil || !serviceExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "External service not found"})
        return
    }

    var linkID string
    err = config.DB.QueryRow(`
        INSERT INTO clinic_service_links (clinic_id, service_id, is_default)
        VALUES ($1, $2, $3) RETURNING id
    `, input.ClinicID, input.ServiceID, input.IsDefault).Scan(&linkID)
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create clinic service link"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": linkID, "message": "Clinic service link created successfully"})
}

func GetClinicServiceLinks(c *gin.Context) {
    clinicID := c.Query("clinic_id")
    serviceID := c.Query("service_id")
    
    var query string
    var args []interface{}
    
    if clinicID != "" && serviceID != "" {
        query = `
            SELECT id, clinic_id, service_id, is_default, is_active, created_at
            FROM clinic_service_links WHERE clinic_id = $1 AND service_id = $2
        `
        args = []interface{}{clinicID, serviceID}
    } else if clinicID != "" {
        query = `
            SELECT id, clinic_id, service_id, is_default, is_active, created_at
            FROM clinic_service_links WHERE clinic_id = $1 ORDER BY created_at DESC
        `
        args = []interface{}{clinicID}
    } else if serviceID != "" {
        query = `
            SELECT id, clinic_id, service_id, is_default, is_active, created_at
            FROM clinic_service_links WHERE service_id = $1 ORDER BY created_at DESC
        `
        args = []interface{}{serviceID}
    } else {
        query = `
            SELECT id, clinic_id, service_id, is_default, is_active, created_at
            FROM clinic_service_links ORDER BY created_at DESC
        `
    }

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clinic service links"})
        return
    }
    defer rows.Close()

    var links []models.ClinicServiceLink
    for rows.Next() {
        var link models.ClinicServiceLink
        err := rows.Scan(&link.ID, &link.ClinicID, &link.ServiceID, &link.IsDefault, &link.IsActive, &link.CreatedAt)
        if err != nil {
            continue
        }
        links = append(links, link)
    }

    c.JSON(http.StatusOK, links)
}

func GetClinicServiceLink(c *gin.Context) {
    linkID := c.Param("id")
    
    var link models.ClinicServiceLink
    err := config.DB.QueryRow(`
        SELECT id, clinic_id, service_id, is_default, is_active, created_at
        FROM clinic_service_links WHERE id = $1
    `, linkID).Scan(&link.ID, &link.ClinicID, &link.ServiceID, &link.IsDefault, &link.IsActive, &link.CreatedAt)
    
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Clinic service link not found"})
        return
    }

    c.JSON(http.StatusOK, link)
}

func DeleteClinicServiceLink(c *gin.Context) {
    linkID := c.Param("id")
    
    result, err := config.DB.Exec(`DELETE FROM clinic_service_links WHERE id = $1`, linkID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete clinic service link"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Clinic service link not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Clinic service link deleted successfully"})
}
