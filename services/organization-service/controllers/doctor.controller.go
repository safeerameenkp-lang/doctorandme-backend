package controllers

import (
    "organization-service/config"
    "organization-service/models"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "net/http"
    "strconv"
)

// Doctor Controllers
// CreateDoctorInput - Creates a doctor profile without clinic assignment
// Use clinic-doctor-links API to link doctor to multiple clinics
type CreateDoctorInput struct {
    // Option 1: Use existing doctor user
    UserID string `json:"user_id" binding:"omitempty,uuid"`

    // Option 2: Create new user
    FirstName string  `json:"first_name" binding:"required_without=UserID,min=2,max=50"`
    LastName  string  `json:"last_name" binding:"required_without=UserID,min=2,max=50"`
    Email     string  `json:"email" binding:"required_without=UserID,email"`
    Username  string  `json:"username" binding:"required_without=UserID,min=3,max=30"`
    Phone     *string `json:"phone" binding:"omitempty"`
    Password  string  `json:"password" binding:"required_without=UserID,min=8"`

    // Doctor profile
    DoctorCode      *string  `json:"doctor_code" binding:"omitempty,max=20"`
    Specialization  *string  `json:"specialization" binding:"omitempty,max=100"`
    LicenseNumber   *string  `json:"license_number" binding:"omitempty,max=100"`
    ConsultationFee *float64 `json:"consultation_fee" binding:"omitempty,min=0"`
    FollowUpFee     *float64 `json:"follow_up_fee" binding:"omitempty,min=0"`
    FollowUpDays    *int     `json:"follow_up_days" binding:"omitempty,min=1,max=365"`
}


func CreateDoctor(c *gin.Context) {
    var input CreateDoctorInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var userID string

    // Case 1: Create new doctor user
    if input.UserID == "" {
        // Check duplicate username/email
        var exists bool
        err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username=$1 OR email=$2)`, input.Username, input.Email).Scan(&exists)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate user"})
            return
        }
        if exists {
            c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
            return
        }

        // Hash password
        passHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
            return
        }

        // Insert new user
        err = config.DB.QueryRow(`
            INSERT INTO users (first_name, last_name, email, username, phone, password_hash)
            VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
        `, input.FirstName, input.LastName, input.Email, input.Username, input.Phone, string(passHash)).Scan(&userID)

        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create doctor user"})
            return
        }

        // Assign doctor role
        var roleID string
        err = config.DB.QueryRow(`SELECT id FROM roles WHERE name='doctor' LIMIT 1`).Scan(&roleID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Doctor role not found"})
            return
        }

        _, err = config.DB.Exec(`INSERT INTO user_roles (user_id, role_id, is_active) VALUES ($1, $2, true)`, userID, roleID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign doctor role"})
            return
        }

    } else {
        // Case 2: Validate existing user
        var validUser bool
        err := config.DB.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM users u
                JOIN user_roles ur ON ur.user_id = u.id
                JOIN roles r ON r.id = ur.role_id
                WHERE u.id=$1 AND r.name='doctor' AND ur.is_active=true
            )
        `, input.UserID).Scan(&validUser)

        if err != nil || !validUser {
            c.JSON(http.StatusBadRequest, gin.H{"error": "User not found or not a doctor"})
            return
        }

        userID = input.UserID
    }

    // Ensure doctor_code unique
    if input.DoctorCode != nil && *input.DoctorCode != "" {
        var exists bool
        err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM doctors WHERE doctor_code=$1)`, *input.DoctorCode).Scan(&exists)
        if err == nil && exists {
            c.JSON(http.StatusConflict, gin.H{"error": "Doctor code already exists"})
            return
        }
    }

    // Create doctor profile
    var doctorID string
    err := config.DB.QueryRow(`
        INSERT INTO doctors (user_id, clinic_id, doctor_code, specialization, license_number, consultation_fee, follow_up_fee, follow_up_days)
        VALUES ($1, NULL, $2, $3, $4, $5, $6, $7) RETURNING id
    `, userID, input.DoctorCode, input.Specialization, input.LicenseNumber, input.ConsultationFee, input.FollowUpFee, input.FollowUpDays).Scan(&doctorID)

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create doctor"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "doctor_id": doctorID,
        "user_id":   userID,
        "role":      "doctor",
        "message":   "Doctor created successfully. Use clinic-doctor-links API to assign to clinics.",
    })
}

func GetAllDoctors(c *gin.Context) {
    rows, err := config.DB.Query(`
        SELECT d.id, d.doctor_code, d.specialization, d.license_number, d.consultation_fee,
               d.follow_up_fee, d.follow_up_days, u.id, u.first_name, u.last_name, u.email, u.username, u.phone
        FROM doctors d
        JOIN users u ON d.user_id = u.id
    `)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctors"})
        return
    }
    defer rows.Close()

    var doctors []map[string]interface{}

    for rows.Next() {
        var dID, uID, firstName, lastName, email, username, phone, doctorCode, specialization, license string
        var consultationFee, followUpFee float64
        var followUpDays int

        err := rows.Scan(&dID, &doctorCode, &specialization, &license, &consultationFee,
            &followUpFee, &followUpDays, &uID, &firstName, &lastName, &email, &username, &phone)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse doctor data"})
            return
        }

        doctor := map[string]interface{}{
            "doctor_id":        dID,
            "doctor_code":      doctorCode,
            "specialization":   specialization,
            "license_number":   license,
            "consultation_fee": consultationFee,
            "follow_up_fee":    followUpFee,
            "follow_up_days":   followUpDays,
            "user": map[string]interface{}{
                "user_id":    uID,
                "first_name": firstName,
                "last_name":  lastName,
                "email":      email,
                "username":   username,
                "phone":      phone,
            },
        }
        doctors = append(doctors, doctor)
    }

    c.JSON(http.StatusOK, gin.H{"doctors": doctors})
}


func GetDoctors(c *gin.Context) {
    clinicID := c.Query("clinic_id")
    
    var query string
    var args []interface{}
    
    if clinicID != "" {
        query = `
            SELECT d.id, d.user_id, d.clinic_id, d.doctor_code, d.specialization, d.license_number, 
                   d.consultation_fee, d.follow_up_fee, d.follow_up_days, d.is_main_doctor, d.is_active, d.created_at,
                   u.first_name, u.last_name, u.email, u.username, u.phone
            FROM doctors d
            JOIN users u ON u.id = d.user_id
            WHERE d.clinic_id = $1 ORDER BY d.created_at DESC
        `
        args = []interface{}{clinicID}
    } else {
        query = `
            SELECT d.id, d.user_id, d.clinic_id, d.doctor_code, d.specialization, d.license_number, 
                   d.consultation_fee, d.follow_up_fee, d.follow_up_days, d.is_main_doctor, d.is_active, d.created_at,
                   u.first_name, u.last_name, u.email, u.username, u.phone
            FROM doctors d
            JOIN users u ON u.id = d.user_id
            ORDER BY d.created_at DESC
        `
    }

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctors"})
        return
    }
    defer rows.Close()

    var doctors []gin.H
    for rows.Next() {
        var doctor models.Doctor
        var firstName, lastName, email, username string
        var phone *string
        err := rows.Scan(&doctor.ID, &doctor.UserID, &doctor.ClinicID, &doctor.DoctorCode, &doctor.Specialization, 
                        &doctor.LicenseNumber, &doctor.ConsultationFee, &doctor.FollowUpFee, &doctor.FollowUpDays, 
                        &doctor.IsMainDoctor, &doctor.IsActive, &doctor.CreatedAt, &firstName, &lastName, &email, &username, &phone)
        if err != nil {
            continue
        }
        doctors = append(doctors, gin.H{
            "doctor": doctor,
            "user": gin.H{
                "first_name": firstName,
                "last_name": lastName,
                "email": email,
                "username": username,
                "phone": phone,
            },
        })
    }

    c.JSON(http.StatusOK, doctors)
}

func GetDoctor(c *gin.Context) {
    doctorID := c.Param("id")
    
    var doctor models.Doctor
    var firstName, lastName, email, username string
    var phone *string
    err := config.DB.QueryRow(`
        SELECT d.id, d.user_id, d.clinic_id, d.doctor_code, d.specialization, d.license_number, 
               d.consultation_fee, d.follow_up_fee, d.follow_up_days, d.is_main_doctor, d.is_active, d.created_at,
               u.first_name, u.last_name, u.email, u.username, u.phone
        FROM doctors d
        JOIN users u ON u.id = d.user_id
        WHERE d.id = $1
    `, doctorID).Scan(&doctor.ID, &doctor.UserID, &doctor.ClinicID, &doctor.DoctorCode, &doctor.Specialization, 
                     &doctor.LicenseNumber, &doctor.ConsultationFee, &doctor.FollowUpFee, &doctor.FollowUpDays, 
                     &doctor.IsMainDoctor, &doctor.IsActive, &doctor.CreatedAt, &firstName, &lastName, &email, &username, &phone)
    
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "doctor": doctor,
        "user": gin.H{
            "first_name": firstName,
            "last_name": lastName,
            "email": email,
            "username": username,
            "phone": phone,
        },
    })
}

type UpdateDoctorInput struct {
    DoctorCode         *string  `json:"doctor_code" binding:"omitempty,max=20"`
    Specialization     *string  `json:"specialization" binding:"omitempty,max=100"`
    LicenseNumber      *string  `json:"license_number" binding:"omitempty,max=100"`
    ConsultationFee    *float64 `json:"consultation_fee" binding:"omitempty,min=0"`
    FollowUpFee        *float64 `json:"follow_up_fee" binding:"omitempty,min=0"`
    FollowUpDays       *int     `json:"follow_up_days" binding:"omitempty,min=1,max=365"`
    IsMainDoctor       *bool    `json:"is_main_doctor"`
    IsActive           *bool    `json:"is_active"`
}

func UpdateDoctor(c *gin.Context) {
    doctorID := c.Param("id")
    var input UpdateDoctorInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Build dynamic update query
    query := "UPDATE doctors SET "
    args := []interface{}{}
    argIndex := 1

    if input.DoctorCode != nil {
        query += "doctor_code = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.DoctorCode)
        argIndex++
    }
    if input.Specialization != nil {
        query += "specialization = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.Specialization)
        argIndex++
    }
    if input.LicenseNumber != nil {
        query += "license_number = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.LicenseNumber)
        argIndex++
    }
    if input.ConsultationFee != nil {
        query += "consultation_fee = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.ConsultationFee)
        argIndex++
    }
    if input.FollowUpFee != nil {
        query += "follow_up_fee = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.FollowUpFee)
        argIndex++
    }
    if input.FollowUpDays != nil {
        query += "follow_up_days = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.FollowUpDays)
        argIndex++
    }
    if input.IsMainDoctor != nil {
        query += "is_main_doctor = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.IsMainDoctor)
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
    args = append(args, doctorID)

    result, err := config.DB.Exec(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update doctor"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Doctor updated successfully"})
}

func DeleteDoctor(c *gin.Context) {
    doctorID := c.Param("id")
    
    // Soft delete by setting is_active to false
    result, err := config.DB.Exec(`UPDATE doctors SET is_active = false WHERE id = $1`, doctorID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate doctor"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Doctor deactivated successfully"})
}
