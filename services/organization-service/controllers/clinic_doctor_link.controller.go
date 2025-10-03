package controllers

import (
    "fmt"
    "net/http"

    "organization-service/config"

    "github.com/gin-gonic/gin"
)


// Clinic Doctor Link Controllers
type CreateClinicDoctorLinkInput struct {
    ClinicID  string `json:"clinic_id" binding:"required,uuid"`
    DoctorID  string `json:"doctor_id" binding:"required,uuid"`
}

// CreateClinicDoctorLink - Links any doctor to a clinic
// A doctor can be linked to multiple clinics
func CreateClinicDoctorLink(c *gin.Context) {
    var input CreateClinicDoctorLinkInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Verify clinic exists and is active
    var clinicExists bool
    err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM clinics WHERE id = $1 AND is_active = true)`, input.ClinicID).Scan(&clinicExists)
    if err != nil || !clinicExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Clinic not found or inactive"})
        return
    }

    // Verify doctor exists and is active (any doctor, not just main doctors)
    var doctorExists bool
    err = config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM doctors WHERE id = $1 AND is_active = true)`, input.DoctorID).Scan(&doctorExists)
    if err != nil || !doctorExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found or inactive"})
        return
    }

    // Check if link already exists
    var linkExists bool
    err = config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM clinic_doctor_links WHERE clinic_id = $1 AND doctor_id = $2)`, input.ClinicID, input.DoctorID).Scan(&linkExists)
    if err == nil && linkExists {
        c.JSON(http.StatusConflict, gin.H{"error": "Doctor is already linked to this clinic"})
        return
    }

    var linkID string
    err = config.DB.QueryRow(`
        INSERT INTO clinic_doctor_links (clinic_id, doctor_id)
        VALUES ($1, $2) RETURNING id
    `, input.ClinicID, input.DoctorID).Scan(&linkID)
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create clinic doctor link"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": linkID, "message": "Doctor linked to clinic successfully"})
}

// GetClinicDoctorLinks - List all clinic-doctor links
func GetClinicDoctorLinks(c *gin.Context) {
    type GetClinicDoctorLinksInput struct {
        ClinicID *string `form:"clinic_id" binding:"omitempty,uuid"`
        DoctorID *string `form:"doctor_id" binding:"omitempty,uuid"`
    }

    var input GetClinicDoctorLinksInput
    if err := c.ShouldBindQuery(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    query := `
        SELECT cdl.id,
               c.id as clinic_id, c.name as clinic_name, c.clinic_code,
               d.id as doctor_id, d.doctor_code, d.specialization,
               u.first_name, u.last_name, u.email, u.username
        FROM clinic_doctor_links cdl
        JOIN clinics c ON c.id = cdl.clinic_id
        JOIN doctors d ON d.id = cdl.doctor_id
        JOIN users u ON u.id = d.user_id
        WHERE 1=1
    `

    args := []interface{}{}
    argIndex := 1
    if input.ClinicID != nil {
        query += ` AND c.id = $` + fmt.Sprint(argIndex)
        args = append(args, *input.ClinicID)
        argIndex++
    }
    if input.DoctorID != nil {
        query += ` AND d.id = $` + fmt.Sprint(argIndex)
        args = append(args, *input.DoctorID)
        argIndex++
    }

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links"})
        return
    }
    defer rows.Close()

    var results []gin.H
    for rows.Next() {
        var linkID, clinicID, clinicName, clinicCode string
        var doctorID, doctorCode, specialization string
        var firstName, lastName, email, username string

        if err := rows.Scan(&linkID, &clinicID, &clinicName, &clinicCode,
            &doctorID, &doctorCode, &specialization,
            &firstName, &lastName, &email, &username); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
            return
        }

        results = append(results, gin.H{
            "link": linkID,
            "clinic": gin.H{
                "name": clinicName,
                "clinic_code": clinicCode,
            },
            "doctor": gin.H{
                "doctor_code": doctorCode,
                "specialization": specialization,
                "first_name": firstName,
                "last_name": lastName,
                "email": email,
                "username": username,
            },
        })
    }

    c.JSON(http.StatusOK, gin.H{"links": results})
}


func DeleteClinicDoctorLink(c *gin.Context) {
    linkID := c.Param("id")
    
    result, err := config.DB.Exec(`DELETE FROM clinic_doctor_links WHERE id = $1`, linkID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete clinic doctor link"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Clinic doctor link not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Doctor unlinked from clinic successfully"})
}
