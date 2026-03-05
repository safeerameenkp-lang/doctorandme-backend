package controllers

import (
	"fmt"
	"log"
	"net/http"
	"organization-service/config"
	"strconv"
	"strings"
	"time"

	"organization-service/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// =====================================================
// SESSION-BASED DOCTOR TIME SLOTS MANAGEMENT APIs
// =====================================================

// SessionRequest represents a session within a day
type SessionRequest struct {
	SessionName         string  `json:"session_name" binding:"required"`
	StartTime           string  `json:"start_time" binding:"required"`
	EndTime             string  `json:"end_time" binding:"required"`
	MaxPatients         int     `json:"max_patients" binding:"required"`
	SlotIntervalMinutes int     `json:"slot_interval_minutes" binding:"required"`
	Notes               *string `json:"notes,omitempty"`
}

// CreateDoctorSessionSlotsInput represents input for creating session-based slots
type CreateDoctorSessionSlotsInput struct {
	DoctorID     string           `json:"doctor_id" binding:"required,uuid"`
	ClinicID     string           `json:"clinic_id" binding:"required,uuid"`
	SlotType     string           `json:"slot_type" binding:"required,oneof=clinic_visit video_consultation"`
	SlotDuration int              `json:"slot_duration" binding:"required,min=1"`
	Date         *string          `json:"date,omitempty"`        // Optional: for single date slots
	Weekdays     []int            `json:"weekdays,omitempty"`    // Optional: for recurring weekly slots (0=Sunday, 1=Monday, ..., 6=Saturday) - creates slots that appear every week automatically
	DayOfWeek    *int             `json:"day_of_week,omitempty"` // Deprecated: use weekdays instead
	IsAvailable  *bool            `json:"is_available,omitempty"`
	Notes        *string          `json:"notes,omitempty"`
	Sessions     []SessionRequest `json:"sessions" binding:"required,min=1,dive"`
}

// UpdateSessionTimeRequest represents a lightweight session update
type UpdateSessionTimeRequest struct {
	SessionID string `json:"session_id" binding:"required,uuid"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

// UpdateSessionTimesInput captures session time adjustments and optional additions
type UpdateSessionTimesInput struct {
	Sessions    []UpdateSessionTimeRequest `json:"sessions"`
	NewSessions []SessionRequest           `json:"new_sessions"`
}

// IndividualSlotResponse represents a single bookable slot
type IndividualSlotResponse struct {
	ID                  string  `json:"id"`
	SlotStart           string  `json:"slot_start"`
	SlotEnd             string  `json:"slot_end"`
	IsBooked            bool    `json:"is_booked"`
	IsBookable          bool    `json:"is_bookable"`     // false if no capacity
	MaxPatients         int     `json:"max_patients"`    // Total capacity
	AvailableCount      int     `json:"available_count"` // Available spots
	BookedCount         int     `json:"booked_count"`    // Booked spots
	BookedPatientID     *string `json:"booked_patient_id,omitempty"`
	BookedAppointmentID *string `json:"booked_appointment_id,omitempty"`
	Status              string  `json:"status"`                   // available, booked, blocked
	DisplayMessage      string  `json:"display_message"`          // "Available" or "Booked" or "X/Y Available"
	StartDateTime       string  `json:"start_datetime,omitempty"` // ISO 8601 with timezone
	EndDateTime         string  `json:"end_datetime,omitempty"`   // ISO 8601 with timezone
}

// SessionResponse represents a session with its individual slots
type SessionResponse struct {
	ID                  string                   `json:"id"`
	SessionName         string                   `json:"session_name"`
	StartTime           string                   `json:"start_time"`
	EndTime             string                   `json:"end_time"`
	MaxPatients         int                      `json:"max_patients"`
	SlotIntervalMinutes int                      `json:"slot_interval_minutes"`
	GeneratedSlots      int                      `json:"generated_slots"`
	AvailableSlots      int                      `json:"available_slots"`
	BookedSlots         int                      `json:"booked_slots"`
	Notes               *string                  `json:"notes,omitempty"`
	Slots               []IndividualSlotResponse `json:"slots,omitempty"`
}

// TimeSlotWithSessionsResponse represents the complete response
type TimeSlotWithSessionsResponse struct {
	ID           string            `json:"id"`
	DoctorID     string            `json:"doctor_id"`
	ClinicID     string            `json:"clinic_id"`
	Date         string            `json:"date"`
	DayOfWeek    int               `json:"day_of_week"`
	SlotType     string            `json:"slot_type"`
	SlotDuration int               `json:"slot_duration"`
	IsAvailable  bool              `json:"is_available"`
	Notes        *string           `json:"notes,omitempty"`
	Sessions     []SessionResponse `json:"sessions"`
}

// CreateDoctorSessionSlots - Create session-based time slots with auto-generated individual slots
// POST /doctor-session-slots
// Supports both single date and recurring weekly slots
func CreateDoctorSessionSlots(c *gin.Context) {
	var input CreateDoctorSessionSlotsInput
	var err error

	if err = c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Validate that either date or weekdays is provided
	if input.Date == nil && len(input.Weekdays) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Either 'date' or 'weekdays' must be provided",
		})
		return
	}

	// Determine if we're creating recurring slots (weekdays) or single date slots
	isRecurringMode := len(input.Weekdays) > 0
	var weekdaysToCreate []int
	var singleDate *time.Time

	if isRecurringMode {
		// Validate weekday values (0-6)
		weekdayMap := make(map[int]bool)
		for _, weekday := range input.Weekdays {
			if weekday < 0 || weekday > 6 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Invalid weekday: %d. Weekdays must be between 0 (Sunday) and 6 (Saturday)", weekday),
				})
				return
			}
			if weekdayMap[weekday] {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Duplicate weekday: %d. Each weekday can only be specified once", weekday),
				})
				return
			}
			weekdayMap[weekday] = true
			weekdaysToCreate = append(weekdaysToCreate, weekday)
		}
	} else {
		// Single date mode - validate and parse date (using IST)
		loc, _ := time.LoadLocation("Asia/Kolkata")
		if loc == nil {
			loc = time.FixedZone("IST", 5*3600+30*60)
		}
		parsedDate, err := time.ParseInLocation("2006-01-02", *input.Date, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid date format. Use YYYY-MM-DD format",
			})
			return
		}
		singleDate = &parsedDate
	}

	// Validate doctor exists and is active
	var doctorExists bool
	err = config.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM doctors d
			WHERE d.id = $1 AND d.is_active = true
		)
	`, input.DoctorID).Scan(&doctorExists)

	if err != nil || !doctorExists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Doctor not found",
			"message": "Doctor not found or is inactive",
		})
		return
	}

	// Validate clinic exists and is active
	var clinicExists bool
	err = config.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM clinics c
			WHERE c.id = $1 AND c.is_active = true
		)
	`, input.ClinicID).Scan(&clinicExists)

	if err != nil || !clinicExists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Clinic not found",
			"message": "Clinic not found or is inactive",
		})
		return
	}

	// Validate clinic-doctor link exists
	var linkExists bool
	err = config.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM clinic_doctor_links cdl
			WHERE cdl.clinic_id = $1 
			AND cdl.doctor_id = $2 
			AND cdl.is_active = true
		)
	`, input.ClinicID, input.DoctorID).Scan(&linkExists)

	if err != nil || !linkExists {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Doctor is not linked to this clinic",
			"message": "The specified doctor is not associated with this clinic",
		})
		return
	}

	// Check for duplicate slots
	if isRecurringMode {
		// Check for existing recurring slots for these weekdays
		var existingWeekdays []int
		for _, weekday := range weekdaysToCreate {
			var existingSlot bool
			err = config.DB.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM doctor_time_slots dts
					WHERE dts.doctor_id = $1 
					AND dts.clinic_id = $2
					AND dts.day_of_week = $3
					AND dts.slot_type = $4
					AND dts.specific_date IS NULL
					AND dts.is_active = true
				)
			`, input.DoctorID, input.ClinicID, weekday, input.SlotType).Scan(&existingSlot)

			if err == nil && existingSlot {
				existingWeekdays = append(existingWeekdays, weekday)
			}
		}

		if len(existingWeekdays) > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error":             "Duplicate slots",
				"message":           fmt.Sprintf("Recurring slots already exist for weekdays: %v", existingWeekdays),
				"existing_weekdays": existingWeekdays,
			})
			return
		}
	} else {
		// Check for existing slot on this specific date
		dateStr := singleDate.Format("2006-01-02")
		var existingSlot bool
		err = config.DB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM doctor_time_slots dts
				WHERE dts.doctor_id = $1 
				AND dts.clinic_id = $2
				AND dts.specific_date = $3 
				AND dts.slot_type = $4
				AND dts.is_active = true
			)
		`, input.DoctorID, input.ClinicID, dateStr, input.SlotType).Scan(&existingSlot)

		if err == nil && existingSlot {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Duplicate slots",
				"message": fmt.Sprintf("Time slots already exist for this doctor on %s", dateStr),
			})
			return
		}
	}

	// Validate sessions for overlaps
	for i, session := range input.Sessions {
		h, m, s, err := parseTimeComponents(session.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   fmt.Sprintf("Invalid start_time format in session %d", i),
				"details": err.Error(),
			})
			return
		}
		// Build comparison time on a dummy date
		startTime := time.Date(2000, 1, 1, h, m, s, 0, time.UTC)

		h2, m2, s2, err := parseTimeComponents(session.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   fmt.Sprintf("Invalid end_time format in session %d", i),
				"details": err.Error(),
			})
			return
		}
		endTime := time.Date(2000, 1, 1, h2, m2, s2, 0, time.UTC)

		if !endTime.After(startTime) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Session %d: end_time must be after start_time", i),
			})
			return
		}

		// Check for overlaps with other sessions
		for j, otherSession := range input.Sessions {
			if i != j {
				oh, om, os, _ := parseTimeComponents(otherSession.StartTime)
				otherStart := time.Date(2000, 1, 1, oh, om, os, 0, time.UTC)
				oh2, om2, os2, _ := parseTimeComponents(otherSession.EndTime)
				otherEnd := time.Date(2000, 1, 1, oh2, om2, os2, 0, time.UTC)

				if startTime.Before(otherEnd) && endTime.After(otherStart) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Overlapping sessions",
						"message": fmt.Sprintf("Session '%s' overlaps with session '%s'", session.SessionName, otherSession.SessionName),
					})
					return
				}
			}
		}
	}

	// Set default isAvailable
	isAvailable := true
	if input.IsAvailable != nil {
		isAvailable = *input.IsAvailable
	}

	// Store all created slots for response
	var allCreatedSlots []TimeSlotWithSessionsResponse

	// Create slots - either recurring (one per weekday) or single date
	slotsToCreate := []struct {
		dayOfWeek    *int
		specificDate *string
	}{}

	if isRecurringMode {
		// Create one recurring slot per weekday
		for _, weekday := range weekdaysToCreate {
			slotsToCreate = append(slotsToCreate, struct {
				dayOfWeek    *int
				specificDate *string
			}{dayOfWeek: &weekday, specificDate: nil})
		}
	} else {
		// Create one slot for the specific date
		dateStr := singleDate.Format("2006-01-02")
		dayOfWeek := int(singleDate.Weekday())
		slotsToCreate = append(slotsToCreate, struct {
			dayOfWeek    *int
			specificDate *string
		}{dayOfWeek: &dayOfWeek, specificDate: &dateStr})
	}

	// Create slots
	for _, slotInfo := range slotsToCreate {
		// Start transaction
		tx, err := config.DB.Begin()
		if err != nil {
			log.Printf("Failed to start transaction: %v", err)
			continue
		}

		// Create main time slot record
		var timeSlotID string
		if slotInfo.specificDate != nil {
			// Single date slot
			err = tx.QueryRow(`
				INSERT INTO doctor_time_slots (
					doctor_id, clinic_id, slot_type, specific_date, day_of_week,
					slot_duration, is_active, notes
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id
			`, input.DoctorID, input.ClinicID, input.SlotType, *slotInfo.specificDate, *slotInfo.dayOfWeek,
				input.SlotDuration, isAvailable, input.Notes).Scan(&timeSlotID)
		} else {
			// Recurring slot (day_of_week set, specific_date NULL)
			err = tx.QueryRow(`
				INSERT INTO doctor_time_slots (
					doctor_id, clinic_id, slot_type, specific_date, day_of_week,
					slot_duration, is_active, notes
				)
				VALUES ($1, $2, $3, NULL, $4, $5, $6, $7)
				RETURNING id
			`, input.DoctorID, input.ClinicID, input.SlotType, *slotInfo.dayOfWeek,
				input.SlotDuration, isAvailable, input.Notes).Scan(&timeSlotID)
		}

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create time slot record",
				"message": err.Error(),
			})
			return
		}
		var sessionsResponse []SessionResponse

		// Create sessions and auto-generate individual slots
		for _, session := range input.Sessions {
			h, m, s, err := parseTimeComponents(session.StartTime)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid start time",
					"message": fmt.Sprintf("Session '%s': %v", session.SessionName, err),
				})
				return
			}
			startTime := time.Date(2000, 1, 1, h, m, s, 0, time.UTC)

			h2, m2, s2, err := parseTimeComponents(session.EndTime)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid end time",
					"message": fmt.Sprintf("Session '%s': %v", session.SessionName, err),
				})
				return
			}
			endTime := time.Date(2000, 1, 1, h2, m2, s2, 0, time.UTC)

			// Normalize session times to 24h format for storage
			normalizedStartTime := fmt.Sprintf("%02d:%02d", h, m)
			normalizedEndTime := fmt.Sprintf("%02d:%02d", h2, m2)

			var sessionID string
			err = tx.QueryRow(`
				INSERT INTO doctor_slot_sessions (
					time_slot_id, clinic_id, session_name, start_time, end_time,
					max_patients, slot_interval_minutes, notes
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id
			`, timeSlotID, input.ClinicID, session.SessionName, normalizedStartTime, normalizedEndTime,
				session.MaxPatients, session.SlotIntervalMinutes, session.Notes).Scan(&sessionID)

			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to create doctor session",
					"message": err.Error(),
				})
				return
			}

			// Auto-generate individual slots
			interval := time.Duration(session.SlotIntervalMinutes) * time.Minute

			var individualSlots []IndividualSlotResponse
			slotCount := 0
			currentTime := startTime

			// Calculate max_patients per slot (inherited from session capacity)
			maxPatientsPerSlot := session.MaxPatients

			for currentTime.Before(endTime) {
				slotEnd := currentTime.Add(interval)
				if slotEnd.After(endTime) {
					slotEnd = endTime
				}

				var individualSlotID string
				err = tx.QueryRow(`
					INSERT INTO doctor_individual_slots (
						session_id, clinic_id, slot_start, slot_end, 
						max_patients, available_count, is_booked, status
					)
					VALUES ($1, $2, $3, $4, $5, $6, false, 'available')
					RETURNING id
				`, sessionID, input.ClinicID, currentTime.Format("15:04"), slotEnd.Format("15:04"),
					maxPatientsPerSlot, maxPatientsPerSlot).Scan(&individualSlotID)

				if err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to create individual bookable slot",
						"message": err.Error(),
					})
					return
				}

				slotResp := IndividualSlotResponse{
					ID:             individualSlotID,
					SlotStart:      currentTime.Format("15:04"),
					SlotEnd:        slotEnd.Format("15:04"),
					IsBooked:       false,
					IsBookable:     true,
					MaxPatients:    maxPatientsPerSlot,
					AvailableCount: maxPatientsPerSlot,
					BookedCount:    0,
					Status:         "available",
					DisplayMessage: "Available",
				}

				// ✅ POPULATE EXPLICIT DATETIME for specific date slots
				if slotInfo.specificDate != nil {
					slotResp.StartDateTime = fmt.Sprintf("%sT%s:00+05:30", *slotInfo.specificDate, slotResp.SlotStart)
					slotResp.EndDateTime = fmt.Sprintf("%sT%s:00+05:30", *slotInfo.specificDate, slotResp.SlotEnd)
				}

				individualSlots = append(individualSlots, slotResp)

				slotCount++
				currentTime = slotEnd
			}

			sessionsResponse = append(sessionsResponse, SessionResponse{
				ID:                  sessionID,
				SessionName:         session.SessionName,
				StartTime:           session.StartTime,
				EndTime:             session.EndTime,
				MaxPatients:         session.MaxPatients,
				SlotIntervalMinutes: session.SlotIntervalMinutes,
				GeneratedSlots:      slotCount,
				AvailableSlots:      slotCount,
				BookedSlots:         0,
				Notes:               session.Notes,
				Slots:               individualSlots,
			})
		}

		// Commit transaction
		if err = tx.Commit(); err != nil {
			if slotInfo.specificDate != nil {
				log.Printf("Failed to commit transaction for date %s: %v", *slotInfo.specificDate, err)
			} else {
				log.Printf("Failed to commit transaction for weekday %d: %v", *slotInfo.dayOfWeek, err)
			}
			continue
		}

		// Add to response
		responseSlot := TimeSlotWithSessionsResponse{
			ID:           timeSlotID,
			DoctorID:     input.DoctorID,
			ClinicID:     input.ClinicID,
			DayOfWeek:    *slotInfo.dayOfWeek,
			SlotType:     input.SlotType,
			SlotDuration: input.SlotDuration,
			IsAvailable:  isAvailable,
			Sessions:     sessionsResponse,
		}
		if slotInfo.specificDate != nil {
			responseSlot.Date = *slotInfo.specificDate
		}
		if input.Notes != nil && *input.Notes != "" {
			responseSlot.Notes = input.Notes
		}
		allCreatedSlots = append(allCreatedSlots, responseSlot)
	}

	// Build response
	if len(allCreatedSlots) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create any time slots",
		})
		return
	}

	var message string
	if isRecurringMode {
		message = fmt.Sprintf("Recurring slots created successfully for %d weekday(s). These slots will appear every week automatically until removed.", len(allCreatedSlots))
	} else {
		message = fmt.Sprintf("Doctor time slots created successfully for %d date(s)", len(allCreatedSlots))
	}

	response := gin.H{
		"success":       true,
		"message":       message,
		"total_created": len(allCreatedSlots),
		"data":          allCreatedSlots,
	}

	// For backward compatibility, if only one slot was created, also include it at the top level
	if len(allCreatedSlots) == 1 {
		response["data"] = allCreatedSlots[0]
	}

	c.JSON(http.StatusCreated, response)
}

// ListDoctorSessionSlots - List session-based time slots
// GET /doctor-session-slots?doctor_id=xxx&date=xxx&clinic_id=xxx&slot_type=xxx
func ListDoctorSessionSlots(c *gin.Context) {
	doctorID := c.Query("doctor_id")
	date := c.Query("date")
	clinicID := c.Query("clinic_id")
	slotType := c.Query("slot_type")
	appointmentID := c.Query("appointment_id") // ✅ NEW: For reschedule - exclude current appointment

	// Load Location once
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}

	// ✅ AUTO-DETECT: If appointment_id is provided, get its individual_slot_id
	var currentAppointmentSlotID *string
	if appointmentID != "" {
		err := config.DB.QueryRow(`
			SELECT individual_slot_id 
			FROM appointments 
			WHERE id = $1 AND status IN ('scheduled', 'confirmed', 'pending')
		`, appointmentID).Scan(&currentAppointmentSlotID)

		if err != nil {
			// If appointment not found or has no slot, continue without exclusion
			currentAppointmentSlotID = nil
		}
	}

	if doctorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "doctor_id is required",
		})
		return
	}

	if _, err := uuid.Parse(doctorID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid doctor_id format",
		})
		return
	}

	// ✅ Map slot_type: Support follow-up prefixed types
	var actualSlotType string
	if slotType != "" {
		switch slotType {
		case "clinic_visit":
			actualSlotType = "clinic_visit"
		case "video_consultation":
			actualSlotType = "video_consultation"
		case "follow-up-via-clinic":
			actualSlotType = "clinic_visit"
		case "follow-up-via-video":
			actualSlotType = "video_consultation"
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid slot_type. Must be 'clinic_visit', 'video_consultation', 'follow-up-via-clinic', or 'follow-up-via-video'",
			})
			return
		}
	}

	// ✅ Validate date is not in the past (using IST timezone)
	var requestedDateDayOfWeek *int
	if date != "" {
		requestedDate, err := time.ParseInLocation("2006-01-02", date, loc)
		if err == nil { // Only validate if date is valid format
			now := time.Now().In(loc)
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

			if requestedDate.Before(today) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Cannot fetch slots for past dates",
					"message": "Please select a date from today onwards",
				})
				return
			}
			// Calculate day of week for the requested date (0=Sunday, 1=Monday, ..., 6=Saturday)
			dayOfWeek := int(requestedDate.Weekday())
			requestedDateDayOfWeek = &dayOfWeek
		}
	}

	// ✅ Check for Active Leaves on this date
	var blockMorning, blockAfternoon bool
	if date != "" {
		rows, err := config.DB.Query(`
			SELECT leave_duration FROM doctor_leaves 
			WHERE doctor_id = $1 
			AND status IN ('approved', 'pending')
			AND from_date <= $2 AND to_date >= $2
		`, doctorID, date)

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var lDuration string
				if err := rows.Scan(&lDuration); err == nil {
					switch lDuration {
					case "morning":
						blockMorning = true
					case "afternoon":
						blockAfternoon = true
					default: // full_day
						blockMorning = true
						blockAfternoon = true
					}
				}
			}
		}
	}

	query := `
		SELECT id, doctor_id, clinic_id, specific_date, day_of_week, slot_type, is_active
		FROM doctor_time_slots
		WHERE doctor_id = $1 AND is_active = true
	`
	args := []interface{}{doctorID}
	argIndex := 2

	if clinicID != "" {
		query += fmt.Sprintf(" AND clinic_id = $%d", argIndex)
		args = append(args, clinicID)
		argIndex++
	}

	// ✅ IMPORTANT: When date is provided, match BOTH:
	// 1. Specific date slots (specific_date = date)
	// 2. Recurring slots (day_of_week matches the requested date's day of week)
	// This ensures recurring slots appear for all matching weekdays automatically
	if date != "" && requestedDateDayOfWeek != nil {
		query += fmt.Sprintf(" AND (specific_date = $%d OR (day_of_week = $%d AND specific_date IS NULL))", argIndex, argIndex+1)
		args = append(args, date, *requestedDateDayOfWeek)
		argIndex += 2
	}

	if actualSlotType != "" {
		query += fmt.Sprintf(" AND slot_type = $%d", argIndex)
		args = append(args, actualSlotType)
		argIndex++
	}

	// Order by specific_date (NULLs last for recurring slots), then by day_of_week
	query += " ORDER BY COALESCE(specific_date, '9999-12-31'::date), day_of_week"

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch time slots",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var results []TimeSlotWithSessionsResponse

	for rows.Next() {
		var timeSlot TimeSlotWithSessionsResponse
		var specificDate *string
		var dayOfWeek *int

		err := rows.Scan(
			&timeSlot.ID, &timeSlot.DoctorID, &timeSlot.ClinicID,
			&specificDate, &dayOfWeek, &timeSlot.SlotType, &timeSlot.IsAvailable,
		)
		if err != nil {
			continue
		}

		if specificDate != nil {
			timeSlot.Date = *specificDate
		} else if dayOfWeek != nil {
			// For recurring slots, set Date to the requested date if provided
			// This allows the frontend to know which date these recurring slots apply to
			if date != "" {
				timeSlot.Date = date
			} else {
				timeSlot.Date = "" // Empty for recurring slots when no date requested
			}
		}
		if dayOfWeek != nil {
			timeSlot.DayOfWeek = *dayOfWeek
		}

		// Get sessions for this time slot
		sessionRows, err := config.DB.Query(`
			SELECT id, session_name, start_time, end_time, max_patients, slot_interval_minutes, notes
			FROM doctor_slot_sessions
			WHERE time_slot_id = $1
			ORDER BY start_time
		`, timeSlot.ID)

		if err != nil {
			continue
		}

		var sessions []SessionResponse
		for sessionRows.Next() {
			var session SessionResponse
			err := sessionRows.Scan(
				&session.ID, &session.SessionName, &session.StartTime, &session.EndTime,
				&session.MaxPatients, &session.SlotIntervalMinutes, &session.Notes,
			)
			if err != nil {
				continue
			}

			// Get individual slots for this session
			slotRows, err := config.DB.Query(`
				SELECT id, slot_start, slot_end, is_booked, max_patients, available_count,
				       booked_patient_id, booked_appointment_id, status
				FROM doctor_individual_slots
				WHERE session_id = $1
				ORDER BY slot_start
			`, session.ID)

			if err != nil {
				continue
			}

			var individualSlots []IndividualSlotResponse

			// ⚡ Optimized: 1. Collect Slots, 2. Batch Query Counts, 3. Process
			var tempSlots []IndividualSlotResponse
			var slotIDs []string

			for slotRows.Next() {
				var slot IndividualSlotResponse
				// Just scan data first
				err := slotRows.Scan(
					&slot.ID, &slot.SlotStart, &slot.SlotEnd, &slot.IsBooked,
					&slot.MaxPatients, &slot.AvailableCount,
					&slot.BookedPatientID, &slot.BookedAppointmentID, &slot.Status,
				)
				if err != nil {
					continue
				}
				tempSlots = append(tempSlots, slot)
				slotIDs = append(slotIDs, slot.ID)
			}
			slotRows.Close()

			// 🔥 Step 2: Batch Query for Booking Counts
			slotCountMap := make(map[string]int)
			if len(slotIDs) > 0 {
				query := `
					SELECT individual_slot_id, COUNT(*) 
					FROM appointments 
					WHERE individual_slot_id = ANY($1)
					AND status NOT IN ('cancelled', 'no_show')
				`
				args := []interface{}{pq.Array(slotIDs)}

				// If reschedule mode → exclude current appointment
				if appointmentID != "" {
					query += " AND id != $2"
					args = append(args, appointmentID)
				}

				query += " GROUP BY individual_slot_id"

				rows, err := config.DB.Query(query, args...)
				if err == nil {
					for rows.Next() {
						var slotID string
						var count int
						rows.Scan(&slotID, &count)
						slotCountMap[slotID] = count
					}
					rows.Close()
				}
			}

			// 🔥 Step 3: Process Slots with In-Memory Map Data
			now := time.Now().In(loc).Truncate(time.Minute)
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
			fmt.Printf("DEBUG: ListSlots | Now=%v | Today=%v\n", now, today)
			for _, slot := range tempSlots {
				var startDT time.Time
				var endDT time.Time

				// Get count from map (O(1) lookup)
				actualBookedCount := slotCountMap[slot.ID]

				// ✅ ALWAYS use actualBookedCount for reschedule mode, or when counts don't match
				isCurrentAppointmentSlot := currentAppointmentSlotID != nil && *currentAppointmentSlotID == slot.ID
				if isCurrentAppointmentSlot || actualBookedCount != (slot.MaxPatients-slot.AvailableCount) {
					newAvailableCount := slot.MaxPatients - actualBookedCount
					if newAvailableCount < 0 {
						newAvailableCount = 0
					}

					// Update slot object with correct values for this response
					slot.AvailableCount = newAvailableCount
					slot.BookedCount = actualBookedCount
					slot.IsBooked = (newAvailableCount <= 0)
					if newAvailableCount <= 0 {
						slot.Status = "booked"
					} else {
						slot.Status = "available"
					}
				} else {
					// Use original values when counts match
					slot.BookedCount = slot.MaxPatients - slot.AvailableCount
				}

				// Set is_bookable and display_message based on capacity
				// ⚠️ Deriving availability purely from capacity to avoid "haunted" DB status values
				if slot.AvailableCount <= 0 {
					slot.IsBookable = false
					slot.Status = "booked"
					slot.DisplayMessage = "Fully Booked"
				} else {
					slot.IsBookable = true
					if slot.MaxPatients == 1 {
						slot.DisplayMessage = "Available"
					} else {
						slot.DisplayMessage = fmt.Sprintf("%d/%d Available", slot.AvailableCount, slot.MaxPatients)
					}
				}

				// ✅ FORCE-ANCHOR TO REQUESTED DATE (Fix for Recurring Slots Bug)
				// Recurring slots have no specific date in DB, so we MUST use the requested date query param
				// to generate a correct timestamp. Otherwise, they default to 0001-01-01 or today.
				anchorDateStr := timeSlot.Date
				if anchorDateStr == "" && date != "" {
					anchorDateStr = date
				}

				if anchorDateStr != "" {
					startDT, _ = buildDateTimeIST(anchorDateStr, slot.SlotStart)
					endDT, _ = buildDateTimeIST(anchorDateStr, slot.SlotEnd)

					if !startDT.IsZero() {
						slot.StartDateTime = startDT.Format(time.RFC3339)
						// Format strictly as HH:MM AM/PM for frontend clarity
						slot.SlotStart = startDT.Format("03:04 PM")
						fmt.Printf("DEBUG: Slot %s | TimeStr='%s' | ParsedDT=%v\n", slot.ID, slot.SlotStart, startDT)
					}
					if !endDT.IsZero() {
						slot.EndDateTime = endDT.Format(time.RFC3339)
						slot.SlotEnd = endDT.Format("03:04 PM")
					}
				}
				// 🔒 FINAL OVERRIDE — Standard Booking Logic
				// If current_time > slot_time → DISABLE (Time Passed)
				// If current_time <= slot_time → ENABLE (Future/Current is bookable)
				if !startDT.IsZero() {
					// 🔒 ENTERPRISE DATE-AWARE LOGIC
					// extract slot date (midnight)
					slotDate := time.Date(startDT.Year(), startDT.Month(), startDT.Day(), 0, 0, 0, 0, loc)

					// ⚡ Check Leave Status First (Overrides everything)
					// Check if slot falls in a blocked period
					hour := startDT.Hour()
					isLeaveBlocked := false

					if (hour < 12 && blockMorning) || (hour >= 12 && blockAfternoon) {
						isLeaveBlocked = true
					}

					if isLeaveBlocked {
						slot.IsBookable = false
						slot.Status = "blocked"
						slot.DisplayMessage = "Doctor on Leave"
						goto SlotFinalize
					}

					if slotDate.Before(today) {
						// 1. Past Date -> ALWAYS BLOCK
						slot.IsBookable = false
						slot.Status = "blocked"
						slot.DisplayMessage = "Time Passed"
					} else if slotDate.Equal(today) {
						// 2. Today -> CHECK TIME
						slotTrunc := startDT.Truncate(time.Minute)

						// Strict Rule: Block if slot <= now
						if !slotTrunc.After(now) {
							fmt.Printf("DEBUG: Slot %s BLOCKED (TimePassed) | SlotTrunc=%v <= Now=%v\n", slot.ID, slotTrunc, now)
							slot.IsBookable = false
							slot.Status = "blocked"
							slot.DisplayMessage = "Time Passed"
						} else {
							fmt.Printf("DEBUG: Slot %s ALLOWED (FutureTime) | SlotTrunc=%v > Now=%v\n", slot.ID, slotTrunc, now)
						}
					}
					// 3. Future Date (> Today) -> ALWAYS ENABLE (Implicit)
				}

			SlotFinalize: // Label for skipping checks if leave applies

				individualSlots = append(individualSlots, slot)
			}

			// ✅ DEFINITIVE COUNTER RECALCULATION
			// Recalculate counters after all overrides (capacity + time-passed)
			// to guarantee the session summary matches the visible slots.
			finalAvailable := 0
			finalBooked := 0
			for _, s := range individualSlots {
				if s.IsBookable {
					finalAvailable++
				} else {
					finalBooked++
				}
			}

			session.Slots = individualSlots
			session.GeneratedSlots = len(individualSlots)
			session.AvailableSlots = finalAvailable
			session.BookedSlots = finalBooked

			sessions = append(sessions, session)
		}
		sessionRows.Close()

		timeSlot.Sessions = sessions
		results = append(results, timeSlot)
	}

	// ✅ CALCULATE DAY-WIDE WALK-IN ELIGIBILITY
	// Logic: Walk-in is now ALWAYS available if there are sessions for the day,
	// regardless of whether normal slots are fully booked or not.
	hasAnySessions := len(results) > 0

	walkinAvailable := false
	walkinReason := ""

	if !hasAnySessions {
		walkinAvailable = false
		walkinReason = "No sessions defined for this date"
	} else {
		// Always enable walk-in if sessions exist (as per user request)
		walkinAvailable = true
		walkinReason = "Walk-in booking available"
	}

	response := gin.H{
		"doctor_id":        doctorID,
		"clinic_id":        clinicID,
		"date":             date,
		"slot_type":        slotType,
		"appointment_id":   appointmentID, // ✅ Include for debugging reschedule mode
		"walkin_available": walkinAvailable,
		"walkin_reason":    walkinReason,
		"slots":            results,
		"total":            len(results),
	}

	// ✅ Add reschedule info for debugging
	if appointmentID != "" {
		response["reschedule_info"] = gin.H{
			"appointment_id":   appointmentID,
			"current_slot_id":  currentAppointmentSlotID,
			"has_current_slot": currentAppointmentSlotID != nil,
		}
	}

	c.JSON(http.StatusOK, response)
}

// SyncSlotBookingStatus - Sync all slot booking statuses with appointments table
// POST /doctor-session-slots/sync-booking-status
func SyncSlotBookingStatus(c *gin.Context) {
	clinicID := c.Query("clinic_id")

	// Sync slots that have appointments but aren't marked as booked
	query := `
		UPDATE doctor_individual_slots dis
		SET is_booked = true, 
		    status = 'booked',
		    updated_at = CURRENT_TIMESTAMP
		FROM appointments a
		WHERE dis.id = a.individual_slot_id
		AND a.status NOT IN ('cancelled', 'no_show')
		AND dis.is_booked = false
	`

	args := []interface{}{}
	argIndex := 1

	if clinicID != "" {
		query += fmt.Sprintf(" AND dis.clinic_id = $%d", argIndex)
		args = append(args, clinicID)
		argIndex++
	}

	result, err := config.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to sync booking status",
			"details": err.Error(),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	// Also sync slots marked as booked but with no active appointment
	query2 := `
		UPDATE doctor_individual_slots dis
		SET is_booked = false, 
		    status = 'available',
		    booked_patient_id = NULL,
		    booked_appointment_id = NULL,
		    updated_at = CURRENT_TIMESTAMP
		WHERE dis.is_booked = true
		AND NOT EXISTS (
			SELECT 1 FROM appointments a 
			WHERE a.individual_slot_id = dis.id 
			AND a.status NOT IN ('cancelled', 'no_show')
		)
	`

	args2 := []interface{}{}
	argIndex2 := 1

	if clinicID != "" {
		query2 += fmt.Sprintf(" AND dis.clinic_id = $%d", argIndex2)
		args2 = append(args2, clinicID)
	}

	result2, err := config.DB.Exec(query2, args2...)
	if err != nil {
		log.Printf("Failed to sync freed slots: %v", err)
	}

	rowsFreed, _ := result2.RowsAffected()

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "Slot booking status synced successfully",
		"slots_booked": rowsAffected,
		"slots_freed":  rowsFreed,
		"total_synced": rowsAffected + rowsFreed,
	})
}

// UpdateSessionSlotSessions - Update existing session times and/or add new ones
// PUT /doctor-session-slots/:id
func UpdateSessionSlotSessions(c *gin.Context) {
	timeSlotID := c.Param("id")
	var input UpdateSessionTimesInput

	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Begin transaction
	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Get clinic_id from time slot
	var clinicID string
	err = tx.QueryRow(`
		SELECT clinic_id
		FROM doctor_time_slots 
		WHERE id = $1 AND is_active = true
	`, timeSlotID).Scan(&clinicID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Time slot not found"})
		return
	}

	updatedCount := 0
	createdCount := 0

	// 1. Process Updates
	for _, req := range input.Sessions {
		h, m, _, errStart := parseTimeComponents(req.StartTime)
		h2, m2, _, errEnd := parseTimeComponents(req.EndTime)
		if errStart != nil || errEnd != nil {
			continue // skip invalid instead of failing everything
		}

		normalizedStart := fmt.Sprintf("%02d:%02d", h, m)
		normalizedEnd := fmt.Sprintf("%02d:%02d", h2, m2)

		// Get session info
		var maxPatients, slotInterval int
		err = tx.QueryRow(`
			SELECT max_patients, slot_interval_minutes 
			FROM doctor_slot_sessions 
			WHERE id = $1 AND time_slot_id = $2
		`, req.SessionID, timeSlotID).Scan(&maxPatients, &slotInterval)

		if err != nil {
			continue
		}

		// Update session times
		_, err = tx.Exec(`
			UPDATE doctor_slot_sessions 
			SET start_time = $1, end_time = $2 
			WHERE id = $3
		`, normalizedStart, normalizedEnd, req.SessionID)

		if err != nil {
			continue
		}

		// Delete unbooked slots that fall outside the new time range
		_, err = tx.Exec(`
			DELETE FROM doctor_individual_slots
			WHERE session_id = $1 
			AND is_booked = false 
			AND (slot_start < $2 OR slot_end > $3)
		`, req.SessionID, normalizedStart, normalizedEnd)

		// Add new slots if needed for the new time range
		startTime := time.Date(2000, 1, 1, h, m, 0, 0, time.UTC)
		endTime := time.Date(2000, 1, 1, h2, m2, 0, 0, time.UTC)
		interval := time.Duration(slotInterval) * time.Minute

		currentTime := startTime
		for currentTime.Before(endTime) {
			slotEnd := currentTime.Add(interval)
			if slotEnd.After(endTime) {
				slotEnd = endTime
			}
			sStart := currentTime.Format("15:04")
			sEnd := slotEnd.Format("15:04")

			// Check if slot exists (we only check times, to avoid overlapping and recreating)
			var exists bool
			err = tx.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM doctor_individual_slots 
					WHERE session_id = $1 AND slot_start = $2 AND slot_end = $3
				)
			`, req.SessionID, sStart, sEnd).Scan(&exists)

			if err == nil && !exists {
				_, err = tx.Exec(`
					INSERT INTO doctor_individual_slots (
						session_id, clinic_id, slot_start, slot_end, 
						max_patients, available_count, is_booked, status
					)
					VALUES ($1, $2, $3, $4, $5, $6, false, 'available')
				`, req.SessionID, clinicID, sStart, sEnd, maxPatients, maxPatients)
			}
			currentTime = slotEnd
		}
		updatedCount++
	}

	// 2. Process New Sessions
	for _, session := range input.NewSessions {
		h, m, _, err := parseTimeComponents(session.StartTime)
		h2, m2, _, err2 := parseTimeComponents(session.EndTime)
		if err != nil || err2 != nil {
			continue
		}

		normalizedStart := fmt.Sprintf("%02d:%02d", h, m)
		normalizedEnd := fmt.Sprintf("%02d:%02d", h2, m2)

		var sessionID string
		err = tx.QueryRow(`
			INSERT INTO doctor_slot_sessions (
				time_slot_id, clinic_id, session_name, start_time, end_time,
				max_patients, slot_interval_minutes, notes
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id
		`, timeSlotID, clinicID, session.SessionName, normalizedStart, normalizedEnd,
			session.MaxPatients, session.SlotIntervalMinutes, session.Notes).Scan(&sessionID)

		if err != nil {
			continue
		}

		startTime := time.Date(2000, 1, 1, h, m, 0, 0, time.UTC)
		endTime := time.Date(2000, 1, 1, h2, m2, 0, 0, time.UTC)
		interval := time.Duration(session.SlotIntervalMinutes) * time.Minute
		currentTime := startTime

		for currentTime.Before(endTime) {
			slotEnd := currentTime.Add(interval)
			if slotEnd.After(endTime) {
				slotEnd = endTime
			}
			sStart := currentTime.Format("15:04")
			sEnd := slotEnd.Format("15:04")

			_, err = tx.Exec(`
				INSERT INTO doctor_individual_slots (
					session_id, clinic_id, slot_start, slot_end, 
					max_patients, available_count, is_booked, status
				)
				VALUES ($1, $2, $3, $4, $5, $6, false, 'available')
			`, sessionID, clinicID, sStart, sEnd, session.MaxPatients, session.MaxPatients)

			currentTime = slotEnd
		}
		createdCount++
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"message":          "Session slot updated successfully",
		"updated_sessions": updatedCount,
		"created_sessions": createdCount,
	})
}

// parseTimeComponents extracts hour, minute, second from various time string formats
func parseTimeComponents(tStr string) (h, m, s int, err error) {
	// 1. Try RFC3339 (e.g., "0000-01-01T13:28:00Z")
	// This handles the user's specific case where DB returns full timestamp string
	if t, err := time.Parse(time.RFC3339, tStr); err == nil {
		h, m, s := t.Clock()
		return h, m, s, nil
	}

	// 2. Try Time Only (HH:MM:SS)
	if t, err := time.Parse("15:04:05", tStr); err == nil {
		h, m, s := t.Clock()
		return h, m, s, nil
	}

	// 3. Try Time Only (HH:MM)
	if t, err := time.Parse("15:04", tStr); err == nil {
		h, m, s := t.Clock()
		return h, m, s, nil
	}

	// 4. Try 12-hour format with AM/PM (Manual fallback for legacy strings like "01:00 PM")
	upperT := strings.TrimSpace(strings.ToUpper(tStr))
	if strings.Contains(upperT, "AM") || strings.Contains(upperT, "PM") {
		// Clean string
		cleanT := strings.ReplaceAll(upperT, "AM", "")
		cleanT = strings.ReplaceAll(cleanT, "PM", "")
		cleanT = strings.TrimSpace(cleanT)

		parts := strings.Split(cleanT, ":")
		if len(parts) >= 1 {
			h, _ = strconv.Atoi(parts[0])
		}
		if len(parts) >= 2 {
			m, _ = strconv.Atoi(parts[1])
		}

		if strings.Contains(upperT, "PM") && h < 12 {
			h += 12
		}
		if strings.Contains(upperT, "AM") && h == 12 {
			h = 0
		}
		return h, m, 0, nil
	}

	// Fallback manual parsing if everything fails (robustness)
	cleanT2 := strings.ReplaceAll(upperT, "T", " ") // Handle 'T' separator if missed by Parse
	parts := strings.FieldsFunc(cleanT2, func(r rune) bool {
		return r == ':' || r == ' ' || r == '.' || r == '-'
	})

	// If we have parts, try to grab the last few as time?
	// This covers edge cases like "0000-01-01 13:28:00" if RFC3339 fail due to Z
	if len(parts) >= 2 {
		// Assume last parts are H:M? risky.
		// Let's stick to returning error if formal parsing failed to ensure data quality
		return 0, 0, 0, fmt.Errorf("unable to parse time string: %s", tStr)
	}

	return 0, 0, 0, fmt.Errorf("unable to parse time string: %s", tStr)
}

// buildDateTimeIST creates a time.Time object in Asia/Kolkata for a given date and time string
func buildDateTimeIST(dateStr, timeStr string) (time.Time, error) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}

	// ✅ CRITICAL FIX — Parse date IN IST, not UTC
	d, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return time.Time{}, err
	}

	h, m, s, err := parseTimeComponents(timeStr)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(
		d.Year(), d.Month(), d.Day(),
		h, m, s, 0,
		loc,
	), nil
}
