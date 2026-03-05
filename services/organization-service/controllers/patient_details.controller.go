package controllers

import (
	"context"
	"database/sql"
	"net/http"
	"organization-service/config"
	"organization-service/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Response Structures for Patient Full Details API

type PatientFullDetailsResponse struct {
	PatientInfo       *ClinicPatientResponse `json:"patient_info"`
	TotalSpent        float64                `json:"total_spent"`
	TotalAppointments int                    `json:"total_appointments"`
	DoctorVisits      []DoctorVisitSummary   `json:"doctor_visits"`
	RecentVitals      []VitalSignSummary     `json:"recent_vitals"`
	Timeline          []TimelineEventSummary `json:"visit_timeline"`
}

type DoctorVisitSummary struct {
	DoctorID       string              `json:"doctor_id"`
	DoctorName     string              `json:"doctor_name"`
	DepartmentName string              `json:"department_name"`
	TotalVisits    int                 `json:"total_visits"`
	NormalVisits   int                 `json:"normal_visits"`
	WalkInVisits   int                 `json:"walkin_visits"`
	TotalPaid      float64             `json:"total_paid"`
	LastVisitDate  string              `json:"last_visit_date"`
	Appointments   []DetailAppointment `json:"appointments"`
}

type DetailAppointment struct {
	AppointmentID     string  `json:"appointment_id"`
	Date              string  `json:"date"`
	Time              string  `json:"time"`
	Type              string  `json:"type"` // e.g., clinic_visit, walk_in
	Status            string  `json:"status"`
	Fee               float64 `json:"fee_amount"`
	PaymentStatus     string  `json:"payment_status"`
	Diagnosis         string  `json:"diagnosis"` // Fetched from notes
	FollowUpStatus    string  `json:"followup_status"`
	FollowUpValidTill string  `json:"followup_valid_till"`
}

type VitalSignSummary struct {
	RecordedAt    string  `json:"recorded_at"`
	BloodPressure string  `json:"blood_pressure"`
	PulseRate     int     `json:"pulse_rate"`
	Temperature   float64 `json:"temperature"`
	WeightKg      float64 `json:"weight_kg"`
	SpO2          int     `json:"spo2"`
}

type TimelineEventSummary struct {
	EventDate   string `json:"date"`
	EventType   string `json:"type"` // e.g., Appointment, Vital Sign
	Description string `json:"description"`
}

// GetClinicPatientFullDetails - Get complete details for a specific patient
// GET /clinic-specific-patients/:id/details
func GetClinicPatientFullDetails(c *gin.Context) {
	patientIDStr := c.Param("id")
	clinicIDStr := extractClinicIDFromContext(c)

	if _, err := uuid.Parse(patientIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient_id format"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 1. Fetch Basic Patient Info
	var patient ClinicPatientResponse
	err := config.DB.QueryRowContext(ctx, `
		SELECT cp.id, cp.clinic_id, cp.first_name, cp.last_name, cp.phone, 
		       COALESCE(cp.email, ''), COALESCE(cp.date_of_birth::text, ''), COALESCE(cp.age, 0), COALESCE(cp.gender, ''),
		       COALESCE(cp.address1, ''), COALESCE(cp.address2, ''), COALESCE(cp.district, ''), COALESCE(cp.state, ''), 
		       COALESCE(cp.mo_id, ''), COALESCE(cp.medical_history, ''), COALESCE(cp.allergies, ''), 
		       COALESCE(cp.blood_group, ''), COALESCE(cp.smoking_status, ''), COALESCE(cp.alcohol_use, ''), 
		       COALESCE(cp.height_cm, 0), COALESCE(cp.weight_kg, 0), 
		       cp.is_active, COALESCE(cp.global_patient_id::text, ''),
		       cp.created_at, cp.updated_at
		FROM clinic_patients cp
		WHERE cp.id = $1 AND cp.clinic_id = $2
	`, patientIDStr, clinicIDStr).Scan(
		&patient.ID, &patient.ClinicID, &patient.FirstName, &patient.LastName,
		&patient.Phone, &patient.Email, &patient.DateOfBirth, &patient.Age, &patient.Gender,
		&patient.Address1, &patient.Address2, &patient.District, &patient.State,
		&patient.MOID, &patient.MedicalHistory, &patient.Allergies, &patient.BloodGroup,
		&patient.SmokingStatus, &patient.AlcoholUse, &patient.HeightCm, &patient.WeightKg,
		&patient.IsActive, &patient.GlobalPatientID,
		&patient.CreatedAt, &patient.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
			return
		}
		middleware.SendDatabaseError(c, "Failed to fetch patient details")
		return
	}

	// 2. Fetch all appointments and follow-up data
	rows, err := config.DB.QueryContext(ctx, `
		SELECT 
			a.id, a.doctor_id, a.appointment_date, a.appointment_time, a.consultation_type, 
			a.status, COALESCE(a.fee_amount, 0), a.payment_status, COALESCE(a.booking_mode, 'slot'),
			COALESCE(a.notes, '') as diagnosis,
			COALESCE(f.status, 'none') as followup_status,
			COALESCE(f.valid_until::text, '') as followup_valid_until,
			COALESCE(u.first_name || ' ' || u.last_name, u.first_name) as doctor_name,
			COALESCE(dept.name, '') as department_name
		FROM appointments a
		JOIN doctors d ON a.doctor_id = d.id
		JOIN users u ON d.user_id = u.id
		LEFT JOIN departments dept ON a.department_id = dept.id
		LEFT JOIN follow_ups f ON f.source_appointment_id = a.id
		WHERE a.clinic_patient_id = $1 AND a.clinic_id = $2
		ORDER BY a.appointment_date DESC, a.appointment_time DESC
	`, patientIDStr, clinicIDStr)

	if err != nil {
		middleware.SendDatabaseError(c, "Failed to fetch appointment history")
		return
	}
	defer rows.Close()

	doctorMap := make(map[string]*DoctorVisitSummary)
	var timeline []TimelineEventSummary
	var totalSpent float64
	var totalAppts int

	// Process appointments
	for rows.Next() {
		var id, doctorID, consultationType, status, paymentStatus, bookingMode, diagnosis string
		var followupStatus, followupValidUntil, doctorName, departmentName string
		var apptDate, apptTime time.Time
		var fee float64

		if err := rows.Scan(
			&id, &doctorID, &apptDate, &apptTime, &consultationType,
			&status, &fee, &paymentStatus, &bookingMode,
			&diagnosis, &followupStatus, &followupValidUntil,
			&doctorName, &departmentName,
		); err != nil {
			continue
		}

		totalAppts++
		if paymentStatus == "completed" || paymentStatus == "paid" || paymentStatus == "success" {
			totalSpent += fee
		}

		// Update Doctor Summary Map
		docSym, exists := doctorMap[doctorID]
		if !exists {
			docSym = &DoctorVisitSummary{
				DoctorID:       doctorID,
				DoctorName:     doctorName,
				DepartmentName: departmentName,
				TotalVisits:    0,
				NormalVisits:   0,
				WalkInVisits:   0,
				TotalPaid:      0,
				LastVisitDate:  apptDate.Format("2006-01-02"),
				Appointments:   make([]DetailAppointment, 0),
			}
			doctorMap[doctorID] = docSym
		}

		docSym.TotalVisits++
		if paymentStatus == "completed" || paymentStatus == "paid" || paymentStatus == "success" {
			docSym.TotalPaid += fee
		}
		if consultationType == "walk_in" || bookingMode == "walk_in" {
			docSym.WalkInVisits++
		} else {
			docSym.NormalVisits++
		}

		// Keep the latest visit date
		if apptDate.Format("2006-01-02") > docSym.LastVisitDate {
			docSym.LastVisitDate = apptDate.Format("2006-01-02")
		}

		// Append to Doctor's appointments list
		apptDetail := DetailAppointment{
			AppointmentID:     id,
			Date:              apptDate.Format("2006-01-02"),
			Time:              apptTime.Format("15:04:05"),
			Type:              consultationType,
			Status:            status,
			Fee:               fee,
			PaymentStatus:     paymentStatus,
			Diagnosis:         diagnosis,
			FollowUpStatus:    followupStatus,
			FollowUpValidTill: followupValidUntil,
		}
		docSym.Appointments = append(docSym.Appointments, apptDetail)

		// Add to timeline
		timeline = append(timeline, TimelineEventSummary{
			EventDate:   apptDate.Format("2006-01-02") + " " + apptTime.Format("15:04:05"),
			EventType:   "Appointment",
			Description: "Appointment (" + status + ") with " + doctorName + " for " + consultationType,
		})
	}

	// 3. Fetch Recent Vitals
	vitalRows, err := config.DB.QueryContext(ctx, `
		SELECT
			recorded_at, COALESCE(blood_pressure, ''), COALESCE(pulse_rate, 0),
			COALESCE(temperature, 0.0), COALESCE(weight_kg, 0.0), COALESCE(spo2_percent, 0)
		FROM patient_vitals
		WHERE clinic_patient_id = $1
		ORDER BY recorded_at DESC
		LIMIT 10
	`, patientIDStr)

	var vitals []VitalSignSummary
	if err == nil {
		defer vitalRows.Close()
		for vitalRows.Next() {
			var recordedAt time.Time
			var v VitalSignSummary
			if err := vitalRows.Scan(
				&recordedAt, &v.BloodPressure, &v.PulseRate,
				&v.Temperature, &v.WeightKg, &v.SpO2,
			); err == nil {
				v.RecordedAt = recordedAt.Format("2006-01-02 15:04:05")
				vitals = append(vitals, v)

				timeline = append(timeline, TimelineEventSummary{
					EventDate:   v.RecordedAt,
					EventType:   "Vitals Recorded",
					Description: "Vitals taken (BP, Pulse, Temp, Weight, SpO2)",
				})
			}
		}
	}

	// Convert DoctorMap to slice
	doctorVisits := make([]DoctorVisitSummary, 0, len(doctorMap))
	for _, v := range doctorMap {
		doctorVisits = append(doctorVisits, *v)
	}

	// Prepare final response
	response := PatientFullDetailsResponse{
		PatientInfo:       &patient,
		TotalSpent:        totalSpent,
		TotalAppointments: totalAppts,
		DoctorVisits:      doctorVisits,
		RecentVitals:      vitals,
		Timeline:          timeline, // In production, we should sort this by EventDate descending
	}

	// If no vitals or timeline elements exist, ensure they're returned as empty arrays not nil
	if response.RecentVitals == nil {
		response.RecentVitals = make([]VitalSignSummary, 0)
	}
	if response.Timeline == nil {
		response.Timeline = make([]TimelineEventSummary, 0)
	}

	c.JSON(http.StatusOK, response)
}
