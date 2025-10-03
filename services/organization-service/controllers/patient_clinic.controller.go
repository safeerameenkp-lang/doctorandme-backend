package controllers

import (
    "organization-service/config"
    "organization-service/models"
    "github.com/gin-gonic/gin"
    "net/http"
    "time"
)

// Patient-Clinic Assignment Controllers
type CreatePatientClinicInput struct {
    PatientID string `json:"patient_id" binding:"required,uuid"`
    ClinicID  string `json:"clinic_id" binding:"required,uuid"`
    IsPrimary *bool  `json:"is_primary"`
}

type UpdatePatientClinicInput struct {
    IsPrimary *bool `json:"is_primary"`
}

func AssignPatientToClinic(c *gin.Context) {
    var input CreatePatientClinicInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Verify patient exists
    var patientExists bool
    err := config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM patients WHERE id = $1 AND is_active = true)
    `, input.PatientID).Scan(&patientExists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if !patientExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Patient not found"})
        return
    }

    // Verify clinic exists
    var clinicExists bool
    err = config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM clinics WHERE id = $1 AND is_active = true)
    `, input.ClinicID).Scan(&clinicExists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if !clinicExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Clinic not found"})
        return
    }

    // Check if assignment already exists
    var assignmentExists bool
    err = config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM patient_clinics WHERE patient_id = $1 AND clinic_id = $2)
    `, input.PatientID, input.ClinicID).Scan(&assignmentExists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if assignmentExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Patient already assigned to this clinic"})
        return
    }

    // Set default is_primary to false if not provided
    isPrimary := false
    if input.IsPrimary != nil {
        isPrimary = *input.IsPrimary
    }

    // If this is being set as primary, unset other primary assignments for this patient
    if isPrimary {
        _, err = config.DB.Exec(`
            UPDATE patient_clinics SET is_primary = false WHERE patient_id = $1
        `, input.PatientID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update primary assignments"})
            return
        }
    }

    // Create patient-clinic assignment
    var patientClinic models.PatientClinic
    err = config.DB.QueryRow(`
        INSERT INTO patient_clinics (patient_id, clinic_id, is_primary)
        VALUES ($1, $2, $3)
        RETURNING id, patient_id, clinic_id, is_primary, created_at
    `, input.PatientID, input.ClinicID, isPrimary).Scan(
        &patientClinic.ID, &patientClinic.PatientID, &patientClinic.ClinicID,
        &patientClinic.IsPrimary, &patientClinic.CreatedAt,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign patient to clinic"})
        return
    }

    c.JSON(http.StatusCreated, patientClinic)
}

func GetPatientClinicAssignments(c *gin.Context) {
    query := `
        SELECT pc.id, pc.patient_id, pc.clinic_id, pc.is_primary, pc.created_at,
               p.user_id, p.medical_history, p.allergies, p.blood_group,
               u.first_name, u.last_name, u.email,
               c.name as clinic_name, c.clinic_code
        FROM patient_clinics pc
        JOIN patients p ON p.id = pc.patient_id
        JOIN users u ON u.id = p.user_id
        JOIN clinics c ON c.id = pc.clinic_id
        WHERE p.is_active = true AND c.is_active = true
        ORDER BY pc.created_at DESC
    `

    rows, err := config.DB.Query(query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var assignments []gin.H
    for rows.Next() {
        var patientClinic models.PatientClinic
        var userID, firstName, lastName, email, clinicName, clinicCode string
        var medicalHistory, allergies, bloodGroup *string
        err := rows.Scan(
            &patientClinic.ID, &patientClinic.PatientID, &patientClinic.ClinicID,
            &patientClinic.IsPrimary, &patientClinic.CreatedAt,
            &userID, &medicalHistory, &allergies, &bloodGroup,
            &firstName, &lastName, &email, &clinicName, &clinicCode,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        assignments = append(assignments, gin.H{
            "id":         patientClinic.ID,
            "patient_id": patientClinic.PatientID,
            "clinic_id":  patientClinic.ClinicID,
            "is_primary": patientClinic.IsPrimary,
            "created_at": patientClinic.CreatedAt,
            "patient": gin.H{
                "user_id":         userID,
                "medical_history": medicalHistory,
                "allergies":       allergies,
                "blood_group":     bloodGroup,
                "user": gin.H{
                    "first_name": firstName,
                    "last_name":  lastName,
                    "email":      email,
                },
            },
            "clinic": gin.H{
                "name":        clinicName,
                "clinic_code": clinicCode,
            },
        })
    }

    c.JSON(http.StatusOK, assignments)
}

func GetPatientClinicAssignment(c *gin.Context) {
    assignmentID := c.Param("id")

    var patientClinic models.PatientClinic
    var userID, firstName, lastName, email, clinicName, clinicCode string
    var medicalHistory, allergies, bloodGroup *string
    err := config.DB.QueryRow(`
        SELECT pc.id, pc.patient_id, pc.clinic_id, pc.is_primary, pc.created_at,
               p.user_id, p.medical_history, p.allergies, p.blood_group,
               u.first_name, u.last_name, u.email,
               c.name as clinic_name, c.clinic_code
        FROM patient_clinics pc
        JOIN patients p ON p.id = pc.patient_id
        JOIN users u ON u.id = p.user_id
        JOIN clinics c ON c.id = pc.clinic_id
        WHERE pc.id = $1
    `, assignmentID).Scan(
        &patientClinic.ID, &patientClinic.PatientID, &patientClinic.ClinicID,
        &patientClinic.IsPrimary, &patientClinic.CreatedAt,
        &userID, &medicalHistory, &allergies, &bloodGroup,
        &firstName, &lastName, &email, &clinicName, &clinicCode,
    )
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":         patientClinic.ID,
        "patient_id": patientClinic.PatientID,
        "clinic_id":  patientClinic.ClinicID,
        "is_primary": patientClinic.IsPrimary,
        "created_at": patientClinic.CreatedAt,
        "patient": gin.H{
            "user_id":         userID,
            "medical_history": medicalHistory,
            "allergies":       allergies,
            "blood_group":     bloodGroup,
            "user": gin.H{
                "first_name": firstName,
                "last_name":  lastName,
                "email":      email,
            },
        },
        "clinic": gin.H{
            "name":        clinicName,
            "clinic_code": clinicCode,
        },
    })
}

func UpdatePatientClinicAssignment(c *gin.Context) {
    assignmentID := c.Param("id")
    var input UpdatePatientClinicInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if input.IsPrimary == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
        return
    }

    // If setting as primary, first unset other primary assignments for this patient
    if *input.IsPrimary {
        // Get patient_id first
        var patientID string
        err := config.DB.QueryRow(`
            SELECT patient_id FROM patient_clinics WHERE id = $1
        `, assignmentID).Scan(&patientID)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
            return
        }

        // Unset other primary assignments
        _, err = config.DB.Exec(`
            UPDATE patient_clinics SET is_primary = false WHERE patient_id = $1 AND id != $2
        `, patientID, assignmentID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update primary assignments"})
            return
        }
    }

    // Update the assignment
    result, err := config.DB.Exec(`
        UPDATE patient_clinics SET is_primary = $1 WHERE id = $2
    `, *input.IsPrimary, assignmentID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignment"})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Assignment updated successfully"})
}

func RemovePatientFromClinic(c *gin.Context) {
    assignmentID := c.Param("id")

    result, err := config.DB.Exec(`
        DELETE FROM patient_clinics WHERE id = $1
    `, assignmentID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove patient from clinic"})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Patient removed from clinic successfully"})
}

func GetClinicsByPatient(c *gin.Context) {
    patientID := c.Param("patient_id")

    query := `
        SELECT c.id, c.organization_id, c.clinic_code, c.name, c.email, c.phone, 
               c.address, c.license_number, c.is_active, c.created_at,
               pc.is_primary, pc.created_at as assigned_at
        FROM clinics c
        JOIN patient_clinics pc ON pc.clinic_id = c.id
        WHERE pc.patient_id = $1 AND c.is_active = true
        ORDER BY pc.is_primary DESC, pc.created_at DESC
    `

    rows, err := config.DB.Query(query, patientID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var clinics []gin.H
    for rows.Next() {
        var clinic models.Clinic
        var isPrimary bool
        var assignedAt time.Time
        err := rows.Scan(
            &clinic.ID, &clinic.OrganizationID, &clinic.ClinicCode, &clinic.Name,
            &clinic.Email, &clinic.Phone, &clinic.Address, &clinic.LicenseNumber,
            &clinic.IsActive, &clinic.CreatedAt, &isPrimary, &assignedAt,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        clinics = append(clinics, gin.H{
            "id":             clinic.ID,
            "organization_id": clinic.OrganizationID,
            "clinic_code":    clinic.ClinicCode,
            "name":           clinic.Name,
            "email":          clinic.Email,
            "phone":          clinic.Phone,
            "address":        clinic.Address,
            "license_number": clinic.LicenseNumber,
            "is_active":      clinic.IsActive,
            "created_at":     clinic.CreatedAt,
            "is_primary":     isPrimary,
            "assigned_at":    assignedAt,
        })
    }

    c.JSON(http.StatusOK, clinics)
}
