package controllers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"organization-service/config"
	"organization-service/models"
	"organization-service/services"

	"github.com/gin-gonic/gin"
)

// CreateClinicPharmacyLink initiates a link request
func CreateClinicPharmacyLink(c *gin.Context) {
	var input models.CreateClinicPharmacyLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	linkSvc := services.NewClinicPharmacyLinkService(config.DB)
	linkID, err := linkSvc.CreateLink(c.Request.Context(), input, userID)
	if err != nil {
		if err.Error() == "clinic not found" || err.Error() == "pharmacy not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "already") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "not authorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create link: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        linkID,
		"is_active": true,
		"message":   "Clinic-Pharmacy link created successfully.",
	})
}

// GetClinicPharmacyLinks list all clinic-pharmacy links with filters
func GetClinicPharmacyLinks(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	clinicID := c.Query("clinic_id")
	pharmacyID := c.Query("pharmacy_id")

	query := `
		SELECT cpl.id, cpl.clinic_id, c.name as clinic_name, c.clinic_code,
		       cpl.pharmacy_id, p.name as pharmacy_name, p.pharmacy_code,
		       cpl.is_active, cpl.created_at, cpl.updated_at
		FROM clinic_pharmacy_links cpl
		JOIN clinics c ON c.id = cpl.clinic_id
		JOIN pharmacies p ON p.id = cpl.pharmacy_id
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if clinicID != "" {
		query += " AND cpl.clinic_id = $" + strconv.Itoa(argIndex)
		args = append(args, clinicID)
		argIndex++
	}
	if pharmacyID != "" {
		query += " AND cpl.pharmacy_id = $" + strconv.Itoa(argIndex)
		args = append(args, pharmacyID)
		argIndex++
	}

	query += " ORDER BY cpl.created_at DESC"

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links: " + err.Error()})
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var id, cID, cName, cCode, pID, pName, pCode string
		var isActive bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &cID, &cName, &cCode, &pID, &pName, &pCode, &isActive, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		results = append(results, gin.H{
			"id":         id,
			"is_active":  isActive,
			"created_at": createdAt,
			"updated_at": updatedAt,
			"clinic": gin.H{
				"id":   cID,
				"name": cName,
				"code": cCode,
			},
			"pharmacy": gin.H{
				"id":   pID,
				"name": pName,
				"code": pCode,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"links": results, "count": len(results)})
}

// DeleteClinicPharmacyLink removes/deletes a link request or breaks the relation
func DeleteClinicPharmacyLink(c *gin.Context) {
	linkID := c.Param("id")
	userID := c.GetString("user_id")

	linkSvc := services.NewClinicPharmacyLinkService(config.DB)
	err := linkSvc.DeleteLink(c.Request.Context(), linkID, userID)
	if err != nil {
		if err.Error() == "link not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "not authorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete link: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Clinic-Pharmacy link deleted successfully"})
}
