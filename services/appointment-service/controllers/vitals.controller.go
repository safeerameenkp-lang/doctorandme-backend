package controllers

import (
	"appointment-service/config"
	"appointment-service/middleware"
	"appointment-service/models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Patient Vitals Controllers
type CreateVitalsInput struct {
	AppointmentID   string   `json:"appointment_id" binding:"required,uuid"`
	ClinicPatientID *string  `json:"clinic_patient_id" binding:"omitempty,uuid"`
	SystolicBP      *int     `json:"systolic_bp" binding:"omitempty"`
	DiastolicBP     *int     `json:"diastolic_bp" binding:"omitempty"`
	BloodPressure   *string  `json:"blood_pressure" binding:"omitempty"`
	Temperature     *float64 `json:"temperature" binding:"omitempty"`
	PulseRate       *int     `json:"pulse_rate" binding:"omitempty"`
	RespBPM         *int     `json:"resp_bpm" binding:"omitempty"`
	Spo2Percent     *int     `json:"spo2_percent" binding:"omitempty"`
	SugarMgdl       *float64 `json:"sugar_mgdl" binding:"omitempty"`
	HeightCm        *int     `json:"height_cm" binding:"omitempty"`
	WeightKg        *float64 `json:"weight_kg" binding:"omitempty"`
	BMI             *float64 `json:"bmi" binding:"omitempty"`
	SmokingStatus   *string  `json:"smoking_status" binding:"omitempty"`
	AlcoholUse      *string  `json:"alcohol_use" binding:"omitempty"`
	Notes           *string  `json:"notes" binding:"omitempty"`
	RecordedBy      string   `json:"recorded_by" binding:"required,uuid"`
}

type UpdateVitalsInput struct {
	AppointmentID   *string  `json:"appointment_id" binding:"omitempty,uuid"`
	ClinicPatientID *string  `json:"clinic_patient_id" binding:"omitempty,uuid"`
	RecordedBy      *string  `json:"recorded_by" binding:"omitempty,uuid"`
	SystolicBP      *int     `json:"systolic_bp" binding:"omitempty"`
	DiastolicBP     *int     `json:"diastolic_bp" binding:"omitempty"`
	BloodPressure   *string  `json:"blood_pressure" binding:"omitempty"`
	Temperature     *float64 `json:"temperature" binding:"omitempty"`
	PulseRate       *int     `json:"pulse_rate" binding:"omitempty"`
	RespBPM         *int     `json:"resp_bpm" binding:"omitempty"`
	Spo2Percent     *int     `json:"spo2_percent" binding:"omitempty"`
	SugarMgdl       *float64 `json:"sugar_mgdl" binding:"omitempty"`
	HeightCm        *int     `json:"height_cm" binding:"omitempty"`
	WeightKg        *float64 `json:"weight_kg" binding:"omitempty"`
	BMI             *float64 `json:"bmi" binding:"omitempty"`
	SmokingStatus   *string  `json:"smoking_status" binding:"omitempty"`
	AlcoholUse      *string  `json:"alcohol_use" binding:"omitempty"`
	Notes           *string  `json:"notes" binding:"omitempty"`
}

func CreateVitals(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var input CreateVitalsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Step 1: Validation & Data Fetching
	var (
		appointmentExists, userExists bool
		clinicPatientIDFound          sql.NullString
	)

	err := config.DB.QueryRowContext(ctx, `
		SELECT 
			EXISTS(SELECT 1 FROM appointments WHERE id = $1) as app_exists,
			EXISTS(SELECT 1 FROM users WHERE id = $2 AND is_active = true) as user_exists,
			(SELECT clinic_patient_id FROM appointments WHERE id = $1) as cp_id
	`, input.AppointmentID, input.RecordedBy).Scan(&appointmentExists, &userExists, &clinicPatientIDFound)

	if err != nil {
		log.Printf("ERROR: CreateVitals validation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database validation failed"})
		return
	}

	if !appointmentExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}
	if !userExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recording user not found or inactive"})
		return
	}

	// Use clinic_patient_id from input if provided, otherwise from appointment record
	finalClinicPatientID := ""
	if input.ClinicPatientID != nil && *input.ClinicPatientID != "" {
		finalClinicPatientID = *input.ClinicPatientID
	} else if clinicPatientIDFound.Valid {
		finalClinicPatientID = clinicPatientIDFound.String
	}

	if finalClinicPatientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing clinic patient reference"})
		return
	}

	// Step 2: Atomic UPSERT & Update Status
	tx, err := config.DB.BeginTx(ctx, nil)
	if err != nil {
		middleware.SendDatabaseError(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var vitals models.PatientVitals
	// Using atomic UPSERT (INSERT ... ON CONFLICT)
	err = tx.QueryRowContext(ctx, `
        INSERT INTO patient_vitals (
            appointment_id, clinic_patient_id, systolic_bp, diastolic_bp, blood_pressure, 
            temperature, pulse_rate, resp_bpm, spo2_percent, sugar_mgdl, 
            height_cm, weight_kg, bmi, smoking_status, alcohol_use, notes, recorded_by
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
        ON CONFLICT (appointment_id) DO UPDATE SET
            clinic_patient_id = EXCLUDED.clinic_patient_id,
            systolic_bp = EXCLUDED.systolic_bp,
            diastolic_bp = EXCLUDED.diastolic_bp,
            blood_pressure = EXCLUDED.blood_pressure,
            temperature = EXCLUDED.temperature,
            pulse_rate = EXCLUDED.pulse_rate,
            resp_bpm = EXCLUDED.resp_bpm,
            spo2_percent = EXCLUDED.spo2_percent,
            sugar_mgdl = EXCLUDED.sugar_mgdl,
            height_cm = EXCLUDED.height_cm,
            weight_kg = EXCLUDED.weight_kg,
            bmi = EXCLUDED.bmi,
            smoking_status = EXCLUDED.smoking_status,
            alcohol_use = EXCLUDED.alcohol_use,
            notes = EXCLUDED.notes,
            recorded_by = EXCLUDED.recorded_by,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id, appointment_id, clinic_patient_id, systolic_bp, diastolic_bp, blood_pressure,
                  temperature, pulse_rate, resp_bpm, spo2_percent, sugar_mgdl,
                  height_cm, weight_kg, bmi, smoking_status, alcohol_use, notes, recorded_by, recorded_at, updated_at
    `,
		input.AppointmentID, finalClinicPatientID, input.SystolicBP, input.DiastolicBP, input.BloodPressure,
		input.Temperature, input.PulseRate, input.RespBPM, input.Spo2Percent, input.SugarMgdl,
		input.HeightCm, input.WeightKg, input.BMI, input.SmokingStatus, input.AlcoholUse, input.Notes, input.RecordedBy,
	).Scan(
		&vitals.ID, &vitals.AppointmentID, &vitals.ClinicPatientID, &vitals.SystolicBP, &vitals.DiastolicBP, &vitals.BloodPressure,
		&vitals.Temperature, &vitals.PulseRate, &vitals.RespBPM, &vitals.Spo2Percent, &vitals.SugarMgdl,
		&vitals.HeightCm, &vitals.WeightKg, &vitals.BMI, &vitals.SmokingStatus, &vitals.AlcoholUse, &vitals.Notes,
		&vitals.RecordedBy, &vitals.RecordedAt, &vitals.UpdatedAt,
	)

	if err != nil {
		log.Printf("ERROR: CreateVitals Upsert failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save vitals record", "details": err.Error()})
		return
	}

	// Update check-in status
	_, err = tx.ExecContext(ctx, `UPDATE patient_checkins SET vitals_recorded = true WHERE appointment_id = $1`, input.AppointmentID)
	if err != nil {
		log.Printf("Warning: Failed to mark vitals as recorded in check-ins: %v", err)
	}

	if err = tx.Commit(); err != nil {
		middleware.SendDatabaseError(c, "Failed to commit vitals record")
		return
	}

	c.JSON(http.StatusOK, vitals)
}

func GetVitals(c *gin.Context) {
	appointmentID := c.Query("appointment_id")
	patientID := c.Query("patient_id")
	doctorID := c.Query("doctor_id")
	clinicID := c.Query("clinic_id")
	date := c.Query("date")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 {
		limit = 50
	}

	query := `
        SELECT pv.id, pv.appointment_id, pv.clinic_patient_id, pv.systolic_bp, pv.diastolic_bp, pv.blood_pressure,
               pv.temperature, pv.pulse_rate, pv.resp_bpm, pv.spo2_percent, pv.sugar_mgdl, 
               pv.height_cm, pv.weight_kg, pv.bmi, pv.smoking_status, pv.alcohol_use, pv.notes,
               pv.recorded_by, pv.recorded_at, pv.updated_at,
               a.booking_number, a.appointment_time, a.status,
               COALESCE(u.first_name, cp.first_name, 'Unknown') as first_name, 
               COALESCE(u.last_name, cp.last_name, '') as last_name, 
               COALESCE(u.phone, cp.phone, '') as phone,
               du.first_name as doctor_first_name, du.last_name as doctor_last_name
        FROM patient_vitals pv
        LEFT JOIN appointments a ON a.id = pv.appointment_id
        LEFT JOIN patients p ON p.id = a.patient_id
        LEFT JOIN users u ON u.id = p.user_id
        LEFT JOIN clinic_patients cp ON cp.id = pv.clinic_patient_id
        LEFT JOIN doctors d ON d.id = a.doctor_id
        LEFT JOIN users du ON du.id = d.user_id
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
		query += fmt.Sprintf(" AND (a.patient_id = $%d OR pv.clinic_patient_id = $%d)", argIndex, argIndex)
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
		startTime, _ := time.Parse("2006-01-02", date)
		endTime := startTime.AddDate(0, 0, 1)
		query += fmt.Sprintf(" AND pv.recorded_at >= $%d AND pv.recorded_at < $%d", argIndex, argIndex+1)
		args = append(args, startTime, endTime)
		argIndex += 2
	}

	query += " ORDER BY pv.recorded_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("ERROR: GetVitals query failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	vitalsList := make([]gin.H, 0, limit)
	for rows.Next() {
		var vitals models.PatientVitals
		var bookingNo, status, pFN, pLN, pPhone, dFN, dLN sql.NullString
		var appTime sql.NullTime

		err := rows.Scan(
			&vitals.ID, &vitals.AppointmentID, &vitals.ClinicPatientID, &vitals.SystolicBP, &vitals.DiastolicBP, &vitals.BloodPressure,
			&vitals.Temperature, &vitals.PulseRate, &vitals.RespBPM, &vitals.Spo2Percent, &vitals.SugarMgdl,
			&vitals.HeightCm, &vitals.WeightKg, &vitals.BMI, &vitals.SmokingStatus, &vitals.AlcoholUse, &vitals.Notes,
			&vitals.RecordedBy, &vitals.RecordedAt, &vitals.UpdatedAt,
			&bookingNo, &appTime, &status, &pFN, &pLN, &pPhone, &dFN, &dLN,
		)
		if err != nil {
			continue
		}

		vitalsList = append(vitalsList, gin.H{
			"id": vitals.ID, "appointment_id": vitals.AppointmentID, "clinic_patient_id": vitals.ClinicPatientID,
			"systolic_bp": vitals.SystolicBP, "diastolic_bp": vitals.DiastolicBP, "blood_pressure": vitals.BloodPressure,
			"temperature": vitals.Temperature, "pulse_rate": vitals.PulseRate, "resp_bpm": vitals.RespBPM,
			"spo2_percent": vitals.Spo2Percent, "sugar_mgdl": vitals.SugarMgdl, "height_cm": vitals.HeightCm,
			"weight_kg": vitals.WeightKg, "bmi": vitals.BMI, "smoking_status": vitals.SmokingStatus,
			"alcohol_use": vitals.AlcoholUse, "notes": vitals.Notes, "recorded_by": vitals.RecordedBy,
			"recorded_at": vitals.RecordedAt, "updated_at": vitals.UpdatedAt,
			"appointment": gin.H{
				"booking_number": bookingNo.String, "appointment_time": appTime.Time, "status": status.String,
			},
			"patient": gin.H{"first_name": pFN.String, "last_name": pLN.String, "phone": pPhone.String},
			"doctor":  gin.H{"first_name": dFN.String, "last_name": dLN.String},
		})
	}
	c.JSON(http.StatusOK, vitalsList)
}

func GetVitalsByAppointment(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	appointmentID := c.Param("appointment_id")
	var vitals models.PatientVitals
	var bookingNo, status, pFN, pLN, pPhone, dFN, dLN sql.NullString
	var appTime sql.NullTime

	err := config.DB.QueryRowContext(ctx, `
        SELECT pv.id, pv.appointment_id, pv.clinic_patient_id, pv.systolic_bp, pv.diastolic_bp, pv.blood_pressure,
               pv.temperature, pv.pulse_rate, pv.resp_bpm, pv.spo2_percent, pv.sugar_mgdl, 
               pv.height_cm, pv.weight_kg, pv.bmi, pv.smoking_status, pv.alcohol_use, pv.notes,
               pv.recorded_by, pv.recorded_at, pv.updated_at,
               a.booking_number, a.appointment_time, a.status,
               COALESCE(u.first_name, cp.first_name, 'Unknown') as first_name, 
               COALESCE(u.last_name, cp.last_name, '') as last_name, 
               COALESCE(u.phone, cp.phone, '') as phone,
               du.first_name as doctor_first_name, du.last_name as doctor_last_name
        FROM patient_vitals pv
        LEFT JOIN appointments a ON a.id = pv.appointment_id
        LEFT JOIN patients p ON p.id = a.patient_id
        LEFT JOIN users u ON u.id = p.user_id
        LEFT JOIN clinic_patients cp ON cp.id = pv.clinic_patient_id
        LEFT JOIN doctors d ON d.id = a.doctor_id
        LEFT JOIN users du ON du.id = d.user_id
        WHERE pv.appointment_id = $1
    `, appointmentID).Scan(
		&vitals.ID, &vitals.AppointmentID, &vitals.ClinicPatientID, &vitals.SystolicBP, &vitals.DiastolicBP, &vitals.BloodPressure,
		&vitals.Temperature, &vitals.PulseRate, &vitals.RespBPM, &vitals.Spo2Percent, &vitals.SugarMgdl,
		&vitals.HeightCm, &vitals.WeightKg, &vitals.BMI, &vitals.SmokingStatus, &vitals.AlcoholUse, &vitals.Notes,
		&vitals.RecordedBy, &vitals.RecordedAt, &vitals.UpdatedAt,
		&bookingNo, &appTime, &status, &pFN, &pLN, &pPhone, &dFN, &dLN,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vitals not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": vitals.ID, "appointment_id": vitals.AppointmentID, "clinic_patient_id": vitals.ClinicPatientID,
		"systolic_bp": vitals.SystolicBP, "diastolic_bp": vitals.DiastolicBP, "blood_pressure": vitals.BloodPressure,
		"temperature": vitals.Temperature, "pulse_rate": vitals.PulseRate, "resp_bpm": vitals.RespBPM,
		"spo2_percent": vitals.Spo2Percent, "sugar_mgdl": vitals.SugarMgdl, "height_cm": vitals.HeightCm,
		"weight_kg": vitals.WeightKg, "bmi": vitals.BMI, "smoking_status": vitals.SmokingStatus,
		"alcohol_use": vitals.AlcoholUse, "notes": vitals.Notes, "recorded_by": vitals.RecordedBy,
		"recorded_at": vitals.RecordedAt, "updated_at": vitals.UpdatedAt,
		"appointment": gin.H{"booking_number": bookingNo.String, "appointment_time": appTime.Time, "status": status.String},
		"patient":     gin.H{"first_name": pFN.String, "last_name": pLN.String, "phone": pPhone.String},
		"doctor":      gin.H{"first_name": dFN.String, "last_name": dLN.String},
	})
}

func UpdateVitals(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	vitalsID := c.Param("id")
	var input UpdateVitalsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE patient_vitals SET"
	args := []interface{}{}
	argIndex := 1
	updates := []string{}

	if input.AppointmentID != nil {
		updates = append(updates, fmt.Sprintf(" appointment_id = $%d", argIndex))
		args = append(args, *input.AppointmentID)
		argIndex++
	}
	if input.ClinicPatientID != nil {
		updates = append(updates, fmt.Sprintf(" clinic_patient_id = $%d", argIndex))
		args = append(args, *input.ClinicPatientID)
		argIndex++
	}
	if input.RecordedBy != nil {
		updates = append(updates, fmt.Sprintf(" recorded_by = $%d", argIndex))
		args = append(args, *input.RecordedBy)
		argIndex++
	}
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
	if input.BloodPressure != nil {
		updates = append(updates, fmt.Sprintf(" blood_pressure = $%d", argIndex))
		args = append(args, *input.BloodPressure)
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
	if input.RespBPM != nil {
		updates = append(updates, fmt.Sprintf(" resp_bpm = $%d", argIndex))
		args = append(args, *input.RespBPM)
		argIndex++
	}
	if input.Spo2Percent != nil {
		updates = append(updates, fmt.Sprintf(" spo2_percent = $%d", argIndex))
		args = append(args, *input.Spo2Percent)
		argIndex++
	}
	if input.SugarMgdl != nil {
		updates = append(updates, fmt.Sprintf(" sugar_mgdl = $%d", argIndex))
		args = append(args, *input.SugarMgdl)
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
	if input.BMI != nil {
		updates = append(updates, fmt.Sprintf(" bmi = $%d", argIndex))
		args = append(args, *input.BMI)
		argIndex++
	}
	if input.SmokingStatus != nil {
		updates = append(updates, fmt.Sprintf(" smoking_status = $%d", argIndex))
		args = append(args, *input.SmokingStatus)
		argIndex++
	}
	if input.AlcoholUse != nil {
		updates = append(updates, fmt.Sprintf(" alcohol_use = $%d", argIndex))
		args = append(args, *input.AlcoholUse)
		argIndex++
	}
	if input.Notes != nil {
		updates = append(updates, fmt.Sprintf(" notes = $%d", argIndex))
		args = append(args, *input.Notes)
		argIndex++
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	query += strings.Join(updates, ",") + ", updated_at = CURRENT_TIMESTAMP WHERE id = $" + strconv.Itoa(argIndex)
	args = append(args, vitalsID)

	result, err := config.DB.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed", "details": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
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
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 {
		limit = 20
	}

	query := `
        SELECT pv.id, pv.appointment_id, pv.clinic_patient_id, pv.systolic_bp, pv.diastolic_bp, pv.blood_pressure,
               pv.temperature, pv.pulse_rate, pv.resp_bpm, pv.spo2_percent, pv.sugar_mgdl, 
               pv.height_cm, pv.weight_kg, pv.bmi, pv.smoking_status, pv.alcohol_use, pv.notes,
               pv.recorded_at, pv.updated_at,
               a.booking_number, a.appointment_time, a.status,
               COALESCE(u.first_name, cp.first_name, 'Unknown') as first_name, 
               COALESCE(u.last_name, cp.last_name, '') as last_name, 
               du.first_name as doctor_first_name, du.last_name as doctor_last_name
        FROM patient_vitals pv
        LEFT JOIN appointments a ON a.id = pv.appointment_id
        LEFT JOIN patients p ON p.id = a.patient_id
        LEFT JOIN users u ON u.id = p.user_id
        LEFT JOIN clinic_patients cp ON cp.id = COALESCE(pv.clinic_patient_id, a.clinic_patient_id)
        LEFT JOIN doctors d ON d.id = a.doctor_id
        LEFT JOIN users du ON du.id = d.user_id
        WHERE (pv.clinic_patient_id = $1 OR a.clinic_patient_id = $1 OR a.patient_id = $1)
        ORDER BY pv.recorded_at DESC LIMIT $2 OFFSET $3
    `
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := config.DB.QueryContext(ctx, query, patientID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	vitalsHistory := make([]gin.H, 0, limit)
	for rows.Next() {
		var vitals models.PatientVitals
		var bookingNo, status, pFN, pLN, dFN, dLN sql.NullString
		var appTime sql.NullTime

		err := rows.Scan(
			&vitals.ID, &vitals.AppointmentID, &vitals.ClinicPatientID, &vitals.SystolicBP, &vitals.DiastolicBP, &vitals.BloodPressure,
			&vitals.Temperature, &vitals.PulseRate, &vitals.RespBPM, &vitals.Spo2Percent, &vitals.SugarMgdl,
			&vitals.HeightCm, &vitals.WeightKg, &vitals.BMI, &vitals.SmokingStatus, &vitals.AlcoholUse, &vitals.Notes,
			&vitals.RecordedAt, &vitals.UpdatedAt,
			&bookingNo, &appTime, &status, &pFN, &pLN, &dFN, &dLN,
		)
		if err != nil {
			continue
		}

		vitalsHistory = append(vitalsHistory, gin.H{
			"id": vitals.ID, "appointment_id": vitals.AppointmentID, "clinic_patient_id": vitals.ClinicPatientID,
			"systolic_bp": vitals.SystolicBP, "diastolic_bp": vitals.DiastolicBP, "blood_pressure": vitals.BloodPressure,
			"temperature": vitals.Temperature, "pulse_rate": vitals.PulseRate, "resp_bpm": vitals.RespBPM,
			"spo2_percent": vitals.Spo2Percent, "sugar_mgdl": vitals.SugarMgdl, "height_cm": vitals.HeightCm,
			"weight_kg": vitals.WeightKg, "bmi": vitals.BMI, "smoking_status": vitals.SmokingStatus,
			"alcohol_use": vitals.AlcoholUse, "notes": vitals.Notes, "recorded_at": vitals.RecordedAt, "updated_at": vitals.UpdatedAt,
			"appointment": gin.H{"booking_number": bookingNo.String, "appointment_time": appTime.Time, "status": status.String},
			"patient":     gin.H{"first_name": pFN.String, "last_name": pLN.String},
			"doctor":      gin.H{"first_name": dFN.String, "last_name": dLN.String},
		})
	}
	c.JSON(http.StatusOK, gin.H{"patient_id": patientID, "vitals_history": vitalsHistory, "count": len(vitalsHistory)})
}

func GetVitalsHistoryByAppointment(c *gin.Context) {
	appointmentID := c.Param("appointment_id")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 {
		limit = 20
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 1. First, find the patient associated with this appointment
	var patientID, clinicPatientID sql.NullString
	err := config.DB.QueryRowContext(ctx, `
		SELECT patient_id, clinic_patient_id FROM appointments WHERE id = $1
	`, appointmentID).Scan(&patientID, &clinicPatientID)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// 2. Now fetch the history using either the global patient_id or clinic_patient_id
	query := `
        SELECT pv.id, pv.appointment_id, pv.clinic_patient_id, pv.systolic_bp, pv.diastolic_bp, pv.blood_pressure,
               pv.temperature, pv.pulse_rate, pv.resp_bpm, pv.spo2_percent, pv.sugar_mgdl, 
               pv.height_cm, pv.weight_kg, pv.bmi, pv.smoking_status, pv.alcohol_use, pv.notes,
               pv.recorded_at, pv.updated_at,
               a.booking_number, a.appointment_time, a.status,
               du.first_name as doctor_first_name, du.last_name as doctor_last_name
        FROM patient_vitals pv
        LEFT JOIN appointments a ON a.id = pv.appointment_id
        LEFT JOIN doctors d ON d.id = a.doctor_id
        LEFT JOIN users du ON du.id = d.user_id
        WHERE (a.patient_id = $1 OR pv.clinic_patient_id = $2)
        ORDER BY pv.recorded_at DESC LIMIT $3 OFFSET $4
    `
	rows, err := config.DB.QueryContext(ctx, query, patientID.String, clinicPatientID.String, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}
	defer rows.Close()

	vitalsHistory := make([]gin.H, 0)
	for rows.Next() {
		var vitals models.PatientVitals
		var bookingNo, status, dFN, dLN sql.NullString
		var appTime sql.NullTime

		rows.Scan(
			&vitals.ID, &vitals.AppointmentID, &vitals.ClinicPatientID, &vitals.SystolicBP, &vitals.DiastolicBP, &vitals.BloodPressure,
			&vitals.Temperature, &vitals.PulseRate, &vitals.RespBPM, &vitals.Spo2Percent, &vitals.SugarMgdl,
			&vitals.HeightCm, &vitals.WeightKg, &vitals.BMI, &vitals.SmokingStatus, &vitals.AlcoholUse, &vitals.Notes,
			&vitals.RecordedAt, &vitals.UpdatedAt,
			&bookingNo, &appTime, &status, &dFN, &dLN,
		)

		vitalsHistory = append(vitalsHistory, gin.H{
			"id": vitals.ID, "appointment_id": vitals.AppointmentID,
			"blood_pressure": vitals.BloodPressure, "temperature": vitals.Temperature,
			"pulse_rate": vitals.PulseRate, "spo2_percent": vitals.Spo2Percent,
			"recorded_at": vitals.RecordedAt,
			"appointment": gin.H{"booking_number": bookingNo.String, "appointment_time": appTime.Time, "status": status.String},
			"doctor":      gin.H{"first_name": dFN.String, "last_name": dLN.String},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"appointment_id": appointmentID,
		"patient_id":     patientID.String,
		"vitals_history": vitalsHistory,
		"count":          len(vitalsHistory),
	})
}
