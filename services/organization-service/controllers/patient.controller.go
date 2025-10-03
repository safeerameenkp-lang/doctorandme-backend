package controllers

import (
    "organization-service/config"
    "organization-service/models"
    "github.com/gin-gonic/gin"
    "net/http"
    "strconv"
    "time"
)

// Patient Controllers
type CreatePatientInput struct {
    UserID         string  `json:"user_id" binding:"required,uuid"`
    MOID           *string `json:"mo_id" binding:"omitempty,max=50"`
    MedicalHistory *string `json:"medical_history"`
    Allergies      *string `json:"allergies"`
    BloodGroup     *string `json:"blood_group" binding:"omitempty,max=10"`
}

type UpdatePatientInput struct {
    MOID           *string `json:"mo_id" binding:"omitempty,max=50"`
    MedicalHistory *string `json:"medical_history"`
    Allergies      *string `json:"allergies"`
    BloodGroup     *string `json:"blood_group" binding:"omitempty,max=10"`
    IsActive       *bool   `json:"is_active"`
}

type AssignPatientToClinicInput struct {
    PatientID string `json:"patient_id" binding:"required,uuid"`
    ClinicID  string `json:"clinic_id" binding:"required,uuid"`
    IsPrimary *bool  `json:"is_primary"`
}

func CreatePatient(c *gin.Context) {
    var input CreatePatientInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Verify user exists
    var userExists bool
    err := config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND is_active = true)
    `, input.UserID).Scan(&userExists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if !userExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
        return
    }

    // Check if patient already exists for this user
    var patientExists bool
    err = config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM patients WHERE user_id = $1)
    `, input.UserID).Scan(&patientExists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if patientExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Patient already exists for this user"})
        return
    }

    // Create patient record
    // Note: User already has "patient" role assigned during registration
    var patient models.Patient
    err = config.DB.QueryRow(`
        INSERT INTO patients (user_id, mo_id, medical_history, allergies, blood_group)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, user_id, mo_id, medical_history, allergies, blood_group, is_active, created_at, updated_at
    `, input.UserID, input.MOID, input.MedicalHistory, input.Allergies, input.BloodGroup).Scan(
        &patient.ID, &patient.UserID, &patient.MOID, &patient.MedicalHistory, &patient.Allergies,
        &patient.BloodGroup, &patient.IsActive, &patient.CreatedAt, &patient.UpdatedAt,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create patient"})
        return
    }

    c.JSON(http.StatusCreated, patient)
}

func GetPatients(c *gin.Context) {
    query := `
        SELECT p.id, p.user_id, p.mo_id, p.medical_history, p.allergies, p.blood_group, 
               p.is_active, p.created_at, p.updated_at,
               u.first_name, u.last_name, u.email, u.phone
        FROM patients p
        JOIN users u ON u.id = p.user_id
        WHERE p.is_active = true
        ORDER BY p.created_at DESC
    `

    rows, err := config.DB.Query(query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var patients []gin.H
    for rows.Next() {
        var patient models.Patient
        var firstName, lastName, email, phone string
        err := rows.Scan(
            &patient.ID, &patient.UserID, &patient.MOID, &patient.MedicalHistory, &patient.Allergies,
            &patient.BloodGroup, &patient.IsActive, &patient.CreatedAt, &patient.UpdatedAt,
            &firstName, &lastName, &email, &phone,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        patients = append(patients, gin.H{
            "id":             patient.ID,
            "user_id":        patient.UserID,
            "mo_id":          patient.MOID,
            "medical_history": patient.MedicalHistory,
            "allergies":      patient.Allergies,
            "blood_group":    patient.BloodGroup,
            "is_active":      patient.IsActive,
            "created_at":     patient.CreatedAt,
            "updated_at":     patient.UpdatedAt,
            "user": gin.H{
                "first_name": firstName,
                "last_name":  lastName,
                "email":      email,
                "phone":      phone,
            },
        })
    }

    c.JSON(http.StatusOK, patients)
}

func GetPatient(c *gin.Context) {
    patientID := c.Param("id")

    var patient models.Patient
    var firstName, lastName, email, phone string
    err := config.DB.QueryRow(`
        SELECT p.id, p.user_id, p.mo_id, p.medical_history, p.allergies, p.blood_group, 
               p.is_active, p.created_at, p.updated_at,
               u.first_name, u.last_name, u.email, u.phone
        FROM patients p
        JOIN users u ON u.id = p.user_id
        WHERE p.id = $1
    `, patientID).Scan(
        &patient.ID, &patient.UserID, &patient.MOID, &patient.MedicalHistory, &patient.Allergies,
        &patient.BloodGroup, &patient.IsActive, &patient.CreatedAt, &patient.UpdatedAt,
        &firstName, &lastName, &email, &phone,
    )
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":             patient.ID,
        "user_id":        patient.UserID,
        "mo_id":          patient.MOID,
        "medical_history": patient.MedicalHistory,
        "allergies":      patient.Allergies,
        "blood_group":    patient.BloodGroup,
        "is_active":      patient.IsActive,
        "created_at":     patient.CreatedAt,
        "updated_at":     patient.UpdatedAt,
        "user": gin.H{
            "first_name": firstName,
            "last_name":  lastName,
            "email":      email,
            "phone":      phone,
        },
    })
}

func UpdatePatient(c *gin.Context) {
    patientID := c.Param("id")
    var input UpdatePatientInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Build dynamic update query
    query := "UPDATE patients SET updated_at = $1"
    args := []interface{}{time.Now()}
    argIndex := 2

    if input.MOID != nil {
        query += ", mo_id = $" + strconv.Itoa(argIndex)
        args = append(args, *input.MOID)
        argIndex++
    }
    if input.MedicalHistory != nil {
        query += ", medical_history = $" + strconv.Itoa(argIndex)
        args = append(args, *input.MedicalHistory)
        argIndex++
    }
    if input.Allergies != nil {
        query += ", allergies = $" + strconv.Itoa(argIndex)
        args = append(args, *input.Allergies)
        argIndex++
    }
    if input.BloodGroup != nil {
        query += ", blood_group = $" + strconv.Itoa(argIndex)
        args = append(args, *input.BloodGroup)
        argIndex++
    }
    if input.IsActive != nil {
        query += ", is_active = $" + strconv.Itoa(argIndex)
        args = append(args, *input.IsActive)
        argIndex++
    }

    query += " WHERE id = $" + strconv.Itoa(argIndex)
    args = append(args, patientID)

    result, err := config.DB.Exec(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update patient"})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Patient updated successfully"})
}

func DeletePatient(c *gin.Context) {
    patientID := c.Param("id")

    result, err := config.DB.Exec(`
        UPDATE patients SET is_active = false, updated_at = $1 WHERE id = $2
    `, time.Now(), patientID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete patient"})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Patient deleted successfully"})
}

func GetPatientsByClinic(c *gin.Context) {
    clinicID := c.Param("clinic_id")

    query := `
        SELECT p.id, p.user_id, p.mo_id, p.medical_history, p.allergies, p.blood_group, 
               p.is_active, p.created_at, p.updated_at,
               u.first_name, u.last_name, u.email, u.phone,
               pc.is_primary
        FROM patients p
        JOIN users u ON u.id = p.user_id
        JOIN patient_clinics pc ON pc.patient_id = p.id
        WHERE pc.clinic_id = $1 AND p.is_active = true
        ORDER BY pc.is_primary DESC, p.created_at DESC
    `

    rows, err := config.DB.Query(query, clinicID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var patients []gin.H
    for rows.Next() {
        var patient models.Patient
        var firstName, lastName, email, phone string
        var isPrimary bool
        err := rows.Scan(
            &patient.ID, &patient.UserID, &patient.MOID, &patient.MedicalHistory, &patient.Allergies,
            &patient.BloodGroup, &patient.IsActive, &patient.CreatedAt, &patient.UpdatedAt,
            &firstName, &lastName, &email, &phone, &isPrimary,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        patients = append(patients, gin.H{
            "id":             patient.ID,
            "user_id":        patient.UserID,
            "mo_id":          patient.MOID,
            "medical_history": patient.MedicalHistory,
            "allergies":      patient.Allergies,
            "blood_group":    patient.BloodGroup,
            "is_active":      patient.IsActive,
            "created_at":     patient.CreatedAt,
            "updated_at":     patient.UpdatedAt,
            "is_primary":     isPrimary,
            "user": gin.H{
                "first_name": firstName,
                "last_name":  lastName,
                "email":      email,
                "phone":      phone,
            },
        })
    }

    c.JSON(http.StatusOK, patients)
}
