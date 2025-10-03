package controllers

import (
    "appointment-service/config"
    "appointment-service/models"
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
    "strconv"
    "strings"
    "shared-security"
)

// Patient Vitals Controllers
type CreateVitalsInput struct {
    AppointmentID string   `json:"appointment_id" binding:"required,uuid"`
    SystolicBP    *int     `json:"systolic_bp" binding:"omitempty,min=50,max=300"`
    DiastolicBP   *int     `json:"diastolic_bp" binding:"omitempty,min=30,max=200"`
    Temperature   *float64 `json:"temperature" binding:"omitempty,min=30.0,max=45.0"`
    PulseRate     *int     `json:"pulse_rate" binding:"omitempty,min=30,max=200"`
    HeightCm      *int     `json:"height_cm" binding:"omitempty,min=50,max=250"`
    WeightKg      *float64 `json:"weight_kg" binding:"omitempty,min=1.0,max=500.0"`
    RecordedBy    string   `json:"recorded_by" binding:"required,uuid"`
}

type UpdateVitalsInput struct {
    SystolicBP  *int     `json:"systolic_bp" binding:"omitempty,min=50,max=300"`
    DiastolicBP *int     `json:"diastolic_bp" binding:"omitempty,min=30,max=200"`
    Temperature *float64 `json:"temperature" binding:"omitempty,min=30.0,max=45.0"`
    PulseRate   *int     `json:"pulse_rate" binding:"omitempty,min=30,max=200"`
    HeightCm    *int     `json:"height_cm" binding:"omitempty,min=50,max=250"`
    WeightKg    *float64 `json:"weight_kg" binding:"omitempty,min=1.0,max=500.0"`
}

func CreateVitals(c *gin.Context) {
    var input CreateVitalsInput
    if err := c.ShouldBindJSON(&input); err != nil {
        security.SendValidationError(c, "Invalid input data", err.Error())
        return
    }

    // Verify appointment exists
    var appointmentExists bool
    err := config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM appointments WHERE id = $1)
    `, input.AppointmentID).Scan(&appointmentExists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if !appointmentExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Appointment not found"})
        return
    }

    // Verify user exists
    var userExists bool
    err = config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND is_active = true)
    `, input.RecordedBy).Scan(&userExists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if !userExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
        return
    }

    // Check if vitals already exist for this appointment
    var vitalsExists bool
    err = config.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM patient_vitals WHERE appointment_id = $1)
    `, input.AppointmentID).Scan(&vitalsExists)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if vitalsExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Vitals already recorded for this appointment"})
        return
    }

    // Validate BP values if both are provided
    if input.SystolicBP != nil && input.DiastolicBP != nil {
        if *input.SystolicBP <= *input.DiastolicBP {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Systolic BP must be higher than diastolic BP"})
            return
        }
    }

    // Create vitals record
    var vitals models.PatientVitals
    err = config.DB.QueryRow(`
        INSERT INTO patient_vitals (
            appointment_id, systolic_bp, diastolic_bp, temperature,
            pulse_rate, height_cm, weight_kg, recorded_by
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, appointment_id, systolic_bp, diastolic_bp, temperature,
                  pulse_rate, height_cm, weight_kg, recorded_by, recorded_at
    `, input.AppointmentID, input.SystolicBP, input.DiastolicBP, input.Temperature,
        input.PulseRate, input.HeightCm, input.WeightKg, input.RecordedBy).Scan(
        &vitals.ID, &vitals.AppointmentID, &vitals.SystolicBP, &vitals.DiastolicBP,
        &vitals.Temperature, &vitals.PulseRate, &vitals.HeightCm, &vitals.WeightKg,
        &vitals.RecordedBy, &vitals.RecordedAt,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vitals record"})
        return
    }

    // Update check-in to mark vitals as recorded
    _, err = config.DB.Exec(`
        UPDATE patient_checkins SET vitals_recorded = true WHERE appointment_id = $1
    `, input.AppointmentID)
    if err != nil {
        // Log error but don't fail the vitals creation
        fmt.Printf("Warning: Failed to update check-in vitals status: %v\n", err)
    }

    c.JSON(http.StatusCreated, vitals)
}

func GetVitals(c *gin.Context) {
    // Get query parameters
    appointmentID := c.Query("appointment_id")
    patientID := c.Query("patient_id")
    doctorID := c.Query("doctor_id")
    clinicID := c.Query("clinic_id")
    date := c.Query("date")
    limitStr := c.DefaultQuery("limit", "50")
    offsetStr := c.DefaultQuery("offset", "0")

    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 50
    }
    offset, err := strconv.Atoi(offsetStr)
    if err != nil || offset < 0 {
        offset = 0
    }

    query := `
        SELECT pv.id, pv.appointment_id, pv.systolic_bp, pv.diastolic_bp,
               pv.temperature, pv.pulse_rate, pv.height_cm, pv.weight_kg,
               pv.recorded_by, pv.recorded_at,
               a.patient_id, a.doctor_id, a.clinic_id, a.booking_number,
               a.appointment_time, a.status,
               p.user_id, u.first_name, u.last_name, u.phone,
               d.doctor_code, du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               c.clinic_code, c.name as clinic_name
        FROM patient_vitals pv
        JOIN appointments a ON a.id = pv.appointment_id
        JOIN patients p ON p.id = a.patient_id
        JOIN users u ON u.id = p.user_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
        WHERE 1=1
    `
    args := []interface{}{}
    argIndex := 1

    if appointmentID != "" {
        query += fmt.Sprintf(" AND pv.appointment_id = $%d", argIndex)
        args = append(args, appointmentID)
        argIndex++
    }
    if patientID != "" {
        query += fmt.Sprintf(" AND a.patient_id = $%d", argIndex)
        args = append(args, patientID)
        argIndex++
    }
    if doctorID != "" {
        query += fmt.Sprintf(" AND a.doctor_id = $%d", argIndex)
        args = append(args, doctorID)
        argIndex++
    }
    if clinicID != "" {
        query += fmt.Sprintf(" AND a.clinic_id = $%d", argIndex)
        args = append(args, clinicID)
        argIndex++
    }
    if date != "" {
        query += fmt.Sprintf(" AND DATE(pv.recorded_at) = $%d", argIndex)
        args = append(args, date)
        argIndex++
    }

    query += " ORDER BY pv.recorded_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
    args = append(args, limit, offset)

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var vitalsList []gin.H
    for rows.Next() {
        var vitals models.PatientVitals
        var appointment models.Appointment
        var patientInfo models.PatientInfo
        var doctorInfo models.DoctorInfo
        var clinicInfo models.ClinicInfo

        err := rows.Scan(
            &vitals.ID, &vitals.AppointmentID, &vitals.SystolicBP, &vitals.DiastolicBP,
            &vitals.Temperature, &vitals.PulseRate, &vitals.HeightCm, &vitals.WeightKg,
            &vitals.RecordedBy, &vitals.RecordedAt,
            &appointment.PatientID, &appointment.DoctorID, &appointment.ClinicID, &appointment.BookingNumber,
            &appointment.AppointmentTime, &appointment.Status,
            &patientInfo.UserID, &patientInfo.FirstName, &patientInfo.LastName, &patientInfo.Phone,
            &doctorInfo.DoctorCode, &doctorInfo.FirstName, &doctorInfo.LastName,
            &clinicInfo.ClinicCode, &clinicInfo.Name,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        // Calculate BMI if height and weight are available
        var bmi *float64
        if vitals.HeightCm != nil && vitals.WeightKg != nil {
            heightM := float64(*vitals.HeightCm) / 100.0
            bmiValue := *vitals.WeightKg / (heightM * heightM)
            bmi = &bmiValue
        }

        vitalsList = append(vitalsList, gin.H{
            "id":             vitals.ID,
            "appointment_id": vitals.AppointmentID,
            "systolic_bp":   vitals.SystolicBP,
            "diastolic_bp":  vitals.DiastolicBP,
            "temperature":   vitals.Temperature,
            "pulse_rate":    vitals.PulseRate,
            "height_cm":    vitals.HeightCm,
            "weight_kg":    vitals.WeightKg,
            "bmi":          bmi,
            "recorded_by":  vitals.RecordedBy,
            "recorded_at":  vitals.RecordedAt,
            "appointment": gin.H{
                "patient_id":     appointment.PatientID,
                "doctor_id":      appointment.DoctorID,
                "clinic_id":      appointment.ClinicID,
                "booking_number": appointment.BookingNumber,
                "appointment_time": appointment.AppointmentTime,
                "status":         appointment.Status,
            },
            "patient": gin.H{
                "user_id":    patientInfo.UserID,
                "first_name": patientInfo.FirstName,
                "last_name":  patientInfo.LastName,
                "phone":      patientInfo.Phone,
            },
            "doctor": gin.H{
                "doctor_code": doctorInfo.DoctorCode,
                "first_name":  doctorInfo.FirstName,
                "last_name":   doctorInfo.LastName,
            },
            "clinic": gin.H{
                "clinic_code": clinicInfo.ClinicCode,
                "name":        clinicInfo.Name,
            },
        })
    }

    c.JSON(http.StatusOK, vitalsList)
}

func GetVitalsByAppointment(c *gin.Context) {
    appointmentID := c.Param("appointment_id")

    var vitals models.PatientVitals
    var appointment models.Appointment
    var patientInfo models.PatientInfo
    var doctorInfo models.DoctorInfo
    var clinicInfo models.ClinicInfo

    err := config.DB.QueryRow(`
        SELECT pv.id, pv.appointment_id, pv.systolic_bp, pv.diastolic_bp,
               pv.temperature, pv.pulse_rate, pv.height_cm, pv.weight_kg,
               pv.recorded_by, pv.recorded_at,
               a.patient_id, a.doctor_id, a.clinic_id, a.booking_number,
               a.appointment_time, a.status,
               p.user_id, u.first_name, u.last_name, u.phone,
               d.doctor_code, du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               c.clinic_code, c.name as clinic_name
        FROM patient_vitals pv
        JOIN appointments a ON a.id = pv.appointment_id
        JOIN patients p ON p.id = a.patient_id
        JOIN users u ON u.id = p.user_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
        WHERE pv.appointment_id = $1
    `, appointmentID).Scan(
        &vitals.ID, &vitals.AppointmentID, &vitals.SystolicBP, &vitals.DiastolicBP,
        &vitals.Temperature, &vitals.PulseRate, &vitals.HeightCm, &vitals.WeightKg,
        &vitals.RecordedBy, &vitals.RecordedAt,
        &appointment.PatientID, &appointment.DoctorID, &appointment.ClinicID, &appointment.BookingNumber,
        &appointment.AppointmentTime, &appointment.Status,
        &patientInfo.UserID, &patientInfo.FirstName, &patientInfo.LastName, &patientInfo.Phone,
        &doctorInfo.DoctorCode, &doctorInfo.FirstName, &doctorInfo.LastName,
        &clinicInfo.ClinicCode, &clinicInfo.Name,
    )
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Vitals not found for this appointment"})
        return
    }

    // Calculate BMI if height and weight are available
    var bmi *float64
    if vitals.HeightCm != nil && vitals.WeightKg != nil {
        heightM := float64(*vitals.HeightCm) / 100.0
        bmiValue := *vitals.WeightKg / (heightM * heightM)
        bmi = &bmiValue
    }

    c.JSON(http.StatusOK, gin.H{
        "id":             vitals.ID,
        "appointment_id": vitals.AppointmentID,
        "systolic_bp":   vitals.SystolicBP,
        "diastolic_bp":  vitals.DiastolicBP,
        "temperature":   vitals.Temperature,
        "pulse_rate":    vitals.PulseRate,
        "height_cm":    vitals.HeightCm,
        "weight_kg":    vitals.WeightKg,
        "bmi":          bmi,
        "recorded_by":  vitals.RecordedBy,
        "recorded_at":  vitals.RecordedAt,
        "appointment": gin.H{
            "patient_id":     appointment.PatientID,
            "doctor_id":      appointment.DoctorID,
            "clinic_id":      appointment.ClinicID,
            "booking_number": appointment.BookingNumber,
            "appointment_time": appointment.AppointmentTime,
            "status":         appointment.Status,
        },
        "patient": gin.H{
            "user_id":    patientInfo.UserID,
            "first_name": patientInfo.FirstName,
            "last_name":  patientInfo.LastName,
            "phone":      patientInfo.Phone,
        },
        "doctor": gin.H{
            "doctor_code": doctorInfo.DoctorCode,
            "first_name":  doctorInfo.FirstName,
            "last_name":   doctorInfo.LastName,
        },
        "clinic": gin.H{
            "clinic_code": clinicInfo.ClinicCode,
            "name":        clinicInfo.Name,
        },
    })
}

func UpdateVitals(c *gin.Context) {
    vitalsID := c.Param("id")
    var input UpdateVitalsInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate BP values if both are provided
    if input.SystolicBP != nil && input.DiastolicBP != nil {
        if *input.SystolicBP <= *input.DiastolicBP {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Systolic BP must be higher than diastolic BP"})
            return
        }
    }

    // Build dynamic update query
    query := "UPDATE patient_vitals SET"
    args := []interface{}{}
    argIndex := 1
    updates := []string{}

    if input.SystolicBP != nil {
        updates = append(updates, fmt.Sprintf(" systolic_bp = $%d", argIndex))
        args = append(args, *input.SystolicBP)
        argIndex++
    }
    if input.DiastolicBP != nil {
        updates = append(updates, fmt.Sprintf(" diastolic_bp = $%d", argIndex))
        args = append(args, *input.DiastolicBP)
        argIndex++
    }
    if input.Temperature != nil {
        updates = append(updates, fmt.Sprintf(" temperature = $%d", argIndex))
        args = append(args, *input.Temperature)
        argIndex++
    }
    if input.PulseRate != nil {
        updates = append(updates, fmt.Sprintf(" pulse_rate = $%d", argIndex))
        args = append(args, *input.PulseRate)
        argIndex++
    }
    if input.HeightCm != nil {
        updates = append(updates, fmt.Sprintf(" height_cm = $%d", argIndex))
        args = append(args, *input.HeightCm)
        argIndex++
    }
    if input.WeightKg != nil {
        updates = append(updates, fmt.Sprintf(" weight_kg = $%d", argIndex))
        args = append(args, *input.WeightKg)
        argIndex++
    }

    if len(updates) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
        return
    }

    query += strings.Join(updates, ",")
    query += fmt.Sprintf(" WHERE id = $%d", argIndex)
    args = append(args, vitalsID)

    result, err := config.DB.Exec(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vitals"})
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Vitals not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Vitals updated successfully"})
}

func GetPatientVitalsHistory(c *gin.Context) {
    patientID := c.Param("patient_id")
    limitStr := c.DefaultQuery("limit", "20")
    offsetStr := c.DefaultQuery("offset", "0")

    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 20
    }
    offset, err := strconv.Atoi(offsetStr)
    if err != nil || offset < 0 {
        offset = 0
    }

    query := `
        SELECT pv.id, pv.appointment_id, pv.systolic_bp, pv.diastolic_bp,
               pv.temperature, pv.pulse_rate, pv.height_cm, pv.weight_kg,
               pv.recorded_at,
               a.booking_number, a.appointment_time,
               d.doctor_code, du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               c.clinic_code, c.name as clinic_name
        FROM patient_vitals pv
        JOIN appointments a ON a.id = pv.appointment_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
        WHERE a.patient_id = $1
        ORDER BY pv.recorded_at DESC
        LIMIT $2 OFFSET $3
    `

    rows, err := config.DB.Query(query, patientID, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var vitalsHistory []gin.H
    for rows.Next() {
        var vitals models.PatientVitals
        var appointment models.Appointment
        var doctorInfo models.DoctorInfo
        var clinicInfo models.ClinicInfo

        err := rows.Scan(
            &vitals.ID, &vitals.AppointmentID, &vitals.SystolicBP, &vitals.DiastolicBP,
            &vitals.Temperature, &vitals.PulseRate, &vitals.HeightCm, &vitals.WeightKg,
            &vitals.RecordedAt,
            &appointment.BookingNumber, &appointment.AppointmentTime,
            &doctorInfo.DoctorCode, &doctorInfo.FirstName, &doctorInfo.LastName,
            &clinicInfo.ClinicCode, &clinicInfo.Name,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        // Calculate BMI if height and weight are available
        var bmi *float64
        if vitals.HeightCm != nil && vitals.WeightKg != nil {
            heightM := float64(*vitals.HeightCm) / 100.0
            bmiValue := *vitals.WeightKg / (heightM * heightM)
            bmi = &bmiValue
        }

        vitalsHistory = append(vitalsHistory, gin.H{
            "id":             vitals.ID,
            "appointment_id": vitals.AppointmentID,
            "systolic_bp":   vitals.SystolicBP,
            "diastolic_bp":  vitals.DiastolicBP,
            "temperature":   vitals.Temperature,
            "pulse_rate":    vitals.PulseRate,
            "height_cm":    vitals.HeightCm,
            "weight_kg":    vitals.WeightKg,
            "bmi":          bmi,
            "recorded_at":  vitals.RecordedAt,
            "appointment": gin.H{
                "booking_number": appointment.BookingNumber,
                "appointment_time": appointment.AppointmentTime,
            },
            "doctor": gin.H{
                "doctor_code": doctorInfo.DoctorCode,
                "first_name":  doctorInfo.FirstName,
                "last_name":   doctorInfo.LastName,
            },
            "clinic": gin.H{
                "clinic_code": clinicInfo.ClinicCode,
                "name":        clinicInfo.Name,
            },
        })
    }

    c.JSON(http.StatusOK, gin.H{
        "patient_id": patientID,
        "vitals_history": vitalsHistory,
        "count": len(vitalsHistory),
    })
}
