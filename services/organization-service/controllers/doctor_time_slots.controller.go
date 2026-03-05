package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"organization-service/config"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"organization-service/middleware"
)

// =====================================================
// DOCTOR TIME SLOTS MANAGEMENT APIs
// =====================================================

// DoctorTimeSlotResponse represents a complete time slot response with availability info
type DoctorTimeSlotResponse struct {
	ID             string     `json:"id"`
	DoctorID       string     `json:"doctor_id"`
	ClinicID       string     `json:"clinic_id"`
	Date           string     `json:"date,omitempty"`           // For specific date slots or display text for recurring
	DayOfWeek      *int       `json:"day_of_week,omitempty"`    // For recurring weekly slots (0=Sunday to 6=Saturday)
	SlotType       string     `json:"slot_type"`
	StartTime      string     `json:"start_time"`
	EndTime        string     `json:"end_time"`
	MaxPatients    int        `json:"max_patients"`
	BookedPatients int        `json:"booked_patients"`
	AvailableSpots int        `json:"available_spots"`
	IsAvailable    bool       `json:"is_available"`
	Status         string     `json:"status"`
	Notes          *string    `json:"notes,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateDoctorTimeSlotsInput represents input for creating multiple time slots
type CreateDoctorTimeSlotsInput struct {
	DoctorID    string               `json:"doctor_id" binding:"required,uuid"`
	ClinicID    string               `json:"clinic_id" binding:"required,uuid"`
	SlotType    string               `json:"slot_type" binding:"required"`
	Date        *string              `json:"date,omitempty"`        // For specific date slots (YYYY-MM-DD)
	DayOfWeek   *int                 `json:"day_of_week,omitempty"` // For recurring weekly slots (0=Sunday to 6=Saturday)
	Slots       []TimeSlotDefinition `json:"slots" binding:"required,min=1,dive"`
}

// TimeSlotDefinition represents a single time slot definition
type TimeSlotDefinition struct {
	StartTime   string  `json:"start_time" binding:"required"`
	EndTime     string  `json:"end_time" binding:"required"`
	MaxPatients *int    `json:"max_patients"`
	Notes       *string `json:"notes"`
	DayOfWeek   *int    `json:"day_of_week,omitempty"` // Optional: for UI validation (1=Monday to 7=Sunday)
}

// UpdateDoctorTimeSlotInput represents input for updating a time slot
type UpdateDoctorTimeSlotInput struct {
	SlotType    *string `json:"slot_type,omitempty"`
	StartTime   *string `json:"start_time,omitempty"`
	EndTime     *string `json:"end_time,omitempty"`
	MaxPatients *int    `json:"max_patients,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// CreateDoctorTimeSlots - Create multiple time slots for a doctor (Bulk Create)
// POST /doctor-time-slots
func CreateDoctorTimeSlots(c *gin.Context) {
	var input CreateDoctorTimeSlotsInput
	var err error
	
	if err = c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Validate slot type
	validSlotTypes := map[string]bool{
		"clinic_visit":        true,
		"video_consultation": true,
	}
	if !validSlotTypes[input.SlotType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid slot_type. Must be one of: clinic_visit, video_consultation",
		})
		return
	}

	// Validate that either date OR day_of_week is provided, but not both
	if (input.Date == nil && input.DayOfWeek == nil) || (input.Date != nil && input.DayOfWeek != nil) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provide either 'date' for specific date slots OR 'day_of_week' for recurring weekly slots, but not both",
		})
		return
	}

	// Validate date format if provided
	if input.Date != nil {
		_, err = time.Parse("2006-01-02", *input.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid date format. Use YYYY-MM-DD format",
			})
			return
		}
	}

	// Validate day_of_week range if provided
	if input.DayOfWeek != nil {
		if *input.DayOfWeek < 0 || *input.DayOfWeek > 6 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid day_of_week. Must be between 0 (Sunday) and 6 (Saturday)",
			})
			return
		}
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

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate clinic-doctor link",
		})
		return
	}

	if !linkExists {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Doctor is not linked to this clinic",
			"message": "The specified doctor is not associated with this clinic",
		})
		return
	}

	var createdSlots []DoctorTimeSlotResponse
	var failedSlots []gin.H

	// Process each slot
	for i, slot := range input.Slots {
		// Validate time format
		_, err = time.Parse("15:04", slot.StartTime)
		if err != nil {
			failedSlots = append(failedSlots, gin.H{
				"index": i,
				"error": "Invalid start_time format. Use HH:MM format",
			})
			continue
		}

		_, err = time.Parse("15:04", slot.EndTime)
		if err != nil {
			failedSlots = append(failedSlots, gin.H{
				"index": i,
				"error": "Invalid end_time format. Use HH:MM format",
			})
			continue
		}

		// Validate day_of_week matches the date if both are provided
		if input.Date != nil && slot.DayOfWeek != nil {
			parsedDate, _ := time.Parse("2006-01-02", *input.Date)
			// Convert Go weekday (0=Sunday) to ISO 8601 (1=Monday, 7=Sunday)
			goWeekday := int(parsedDate.Weekday())
			isoWeekday := goWeekday
			if goWeekday == 0 {
				isoWeekday = 7 // Sunday is 7 in ISO 8601
			}
			
			if *slot.DayOfWeek < 1 || *slot.DayOfWeek > 7 {
				failedSlots = append(failedSlots, gin.H{
					"index": i,
					"error": "Invalid day_of_week in slot. Must be between 1 (Monday) and 7 (Sunday)",
				})
				continue
			}
			
			if isoWeekday != *slot.DayOfWeek {
				dayNames := []string{"", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
				failedSlots = append(failedSlots, gin.H{
					"index": i,
					"error": fmt.Sprintf("Date %s is a %s, but day_of_week is set to %d (%s)", 
						*input.Date, dayNames[isoWeekday], *slot.DayOfWeek, dayNames[*slot.DayOfWeek]),
				})
				continue
			}
		}

		// Set default max_patients
		maxPatients := 1
		if slot.MaxPatients != nil && *slot.MaxPatients > 0 {
			maxPatients = *slot.MaxPatients
		}

		// Insert slot
		var slotID string
		var createdAt, updatedAt time.Time

		// Build INSERT query based on whether it's a date-specific or recurring slot
		if input.Date != nil {
			// Insert date-specific slot
			err = config.DB.QueryRow(`
				INSERT INTO doctor_time_slots (
					doctor_id, clinic_id, slot_type, specific_date,
					start_time, end_time, max_patients, notes
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id, created_at, updated_at
			`, input.DoctorID, input.ClinicID, input.SlotType, *input.Date,
				slot.StartTime, slot.EndTime, maxPatients, slot.Notes).Scan(&slotID, &createdAt, &updatedAt)
		} else {
			// Insert recurring weekly slot
			err = config.DB.QueryRow(`
				INSERT INTO doctor_time_slots (
					doctor_id, clinic_id, slot_type, day_of_week,
					start_time, end_time, max_patients, notes
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id, created_at, updated_at
			`, input.DoctorID, input.ClinicID, input.SlotType, *input.DayOfWeek,
				slot.StartTime, slot.EndTime, maxPatients, slot.Notes).Scan(&slotID, &createdAt, &updatedAt)
		}

		if err != nil {
			failedSlots = append(failedSlots, gin.H{
				"index": i,
				"error": "Failed to create slot: " + err.Error(),
			})
			continue
		}

		// Build response
		responseSlot := DoctorTimeSlotResponse{
			ID:             slotID,
			DoctorID:       input.DoctorID,
			ClinicID:       input.ClinicID,
			SlotType:       input.SlotType,
			StartTime:      slot.StartTime,
			EndTime:        slot.EndTime,
			MaxPatients:    maxPatients,
			BookedPatients: 0,
			AvailableSpots: maxPatients,
			IsAvailable:    true,
			Status:         "available",
			Notes:          slot.Notes,
			IsActive:       true,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		}

		// Set date or day_of_week depending on slot type
		if input.Date != nil {
			responseSlot.Date = *input.Date
		} else if input.DayOfWeek != nil {
			responseSlot.DayOfWeek = input.DayOfWeek
			// For recurring slots, add day name in date field for clarity
			dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
			responseSlot.Date = fmt.Sprintf("Every %s", dayNames[*input.DayOfWeek])
		}

		createdSlots = append(createdSlots, responseSlot)
	}

	// Prepare response
	response := gin.H{
		"message":       fmt.Sprintf("Slot creation completed. %d created, %d failed", len(createdSlots), len(failedSlots)),
		"created_slots": createdSlots,
		"failed_slots":  failedSlots,
		"total_created": len(createdSlots),
		"total_failed":  len(failedSlots),
	}

	// Determine HTTP status code
	if len(failedSlots) == 0 {
		c.JSON(http.StatusCreated, response)
	} else if len(createdSlots) == 0 {
		c.JSON(http.StatusBadRequest, response)
	} else {
		c.JSON(http.StatusPartialContent, response)
	}
}

// ListDoctorTimeSlots - List time slots with optional filtering
// GET /doctor-time-slots?doctor_id=xxx&clinic_id=xxx&slot_type=xxx&date=2024-10-15
func ListDoctorTimeSlots(c *gin.Context) {
	// Get query parameters
	doctorID := c.Query("doctor_id")
	clinicID := c.Query("clinic_id")
	slotType := c.Query("slot_type")
	date := c.Query("date")

	// Validate required parameter
	if doctorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "doctor_id is required",
		})
		return
	}

	// Validate doctor_id UUID format
	if _, err := uuid.Parse(doctorID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid doctor_id format. Must be a valid UUID",
		})
		return
	}

	// Validate clinic_id UUID format if provided
	if clinicID != "" {
		if _, err := uuid.Parse(clinicID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid clinic_id format. Must be a valid UUID",
			})
			return
		}
	}

	// Validate slot_type if provided
	if slotType != "" && slotType != "clinic_visit" && slotType != "video_consultation" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid slot_type. Must be one of: clinic_visit, video_consultation",
		})
		return
	}

	// Validate date format if provided
	if date != "" {
		_, err := time.Parse("2006-01-02", date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid date format. Use YYYY-MM-DD",
			})
			return
		}
	}

	// Build query with appointment count
	query := `
		SELECT 
			dts.id, dts.doctor_id, dts.clinic_id, dts.specific_date, dts.day_of_week,
			dts.slot_type, dts.start_time, dts.end_time, dts.max_patients, dts.notes,
			dts.is_active, dts.created_at, dts.updated_at,
			COALESCE(appointment_count.booked_count, 0) as booked_patients
		FROM doctor_time_slots dts
		LEFT JOIN (
			SELECT 
				slot_id,
				COUNT(*) as booked_count
			FROM appointments 
			WHERE status IN ('confirmed', 'completed')
			GROUP BY slot_id
		) appointment_count ON dts.id = appointment_count.slot_id
		WHERE dts.doctor_id = $1 AND dts.is_active = true
	`

	args := []interface{}{doctorID}
	argIndex := 2

	if clinicID != "" {
		query += fmt.Sprintf(" AND dts.clinic_id = $%d", argIndex)
		args = append(args, clinicID)
		argIndex++
	}

	if slotType != "" {
		query += fmt.Sprintf(" AND dts.slot_type = $%d", argIndex)
		args = append(args, slotType)
		argIndex++
	}

	if date != "" {
		query += fmt.Sprintf(" AND dts.specific_date = $%d", argIndex)
		args = append(args, date)
		argIndex++
	}

	query += " ORDER BY dts.specific_date, dts.start_time"

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		log.Printf("Error fetching time slots: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch time slots",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var slots []DoctorTimeSlotResponse

	for rows.Next() {
		var slot DoctorTimeSlotResponse
		var bookedPatients int
		var slotDate *string
		var dayOfWeek *int

		err := rows.Scan(
			&slot.ID, &slot.DoctorID, &slot.ClinicID, &slotDate, &dayOfWeek,
			&slot.SlotType, &slot.StartTime, &slot.EndTime, &slot.MaxPatients, &slot.Notes,
			&slot.IsActive, &slot.CreatedAt, &slot.UpdatedAt, &bookedPatients,
		)
		if err != nil {
			continue
		}

		// Set date or day_of_week
		if slotDate != nil {
			slot.Date = *slotDate
		} else if dayOfWeek != nil {
			slot.DayOfWeek = dayOfWeek
			dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
			slot.Date = fmt.Sprintf("Every %s", dayNames[*dayOfWeek])
		}

		// Calculate availability
		slot.BookedPatients = bookedPatients
		slot.AvailableSpots = slot.MaxPatients - bookedPatients

		if slot.AvailableSpots > 0 {
			slot.IsAvailable = true
			slot.Status = "available"
		} else {
			slot.IsAvailable = false
			slot.Status = "booking_full"
		}

		slots = append(slots, slot)
	}

	c.JSON(http.StatusOK, gin.H{
		"slots":     slots,
		"total":     len(slots),
		"doctor_id": doctorID,
		"clinic_id": clinicID,
		"slot_type": slotType,
		"date":      date,
	})
}

// GetDoctorTimeSlot - Get a single time slot by ID
// GET /doctor-time-slots/:id
func GetDoctorTimeSlot(c *gin.Context) {
	slotID := c.Param("id")

	// Validate slot_id UUID format
	if _, err := uuid.Parse(slotID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid slot_id format. Must be a valid UUID",
		})
		return
	}

	// Query slot with booked count
	var slot DoctorTimeSlotResponse
	var slotDate *string
	var dayOfWeek *int
	var bookedPatients int

	err := config.DB.QueryRow(`
		SELECT 
			dts.id, dts.doctor_id, dts.clinic_id, dts.specific_date, dts.day_of_week,
			dts.slot_type, dts.start_time, dts.end_time, dts.max_patients, dts.notes,
			dts.is_active, dts.created_at, dts.updated_at,
			COALESCE(
				(SELECT COUNT(*) FROM appointments 
				 WHERE slot_id = dts.id AND status IN ('confirmed', 'completed')),
				0
			) as booked_patients
		FROM doctor_time_slots dts
		WHERE dts.id = $1
	`, slotID).Scan(
		&slot.ID, &slot.DoctorID, &slot.ClinicID, &slotDate, &dayOfWeek,
		&slot.SlotType, &slot.StartTime, &slot.EndTime, &slot.MaxPatients, &slot.Notes,
		&slot.IsActive, &slot.CreatedAt, &slot.UpdatedAt, &bookedPatients,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Time slot not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch time slot",
		})
		return
	}

	// Set date or day_of_week
	if slotDate != nil {
		slot.Date = *slotDate
	} else if dayOfWeek != nil {
		slot.DayOfWeek = dayOfWeek
		dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		slot.Date = fmt.Sprintf("Every %s", dayNames[*dayOfWeek])
	}

	// Calculate availability
	slot.BookedPatients = bookedPatients
	slot.AvailableSpots = slot.MaxPatients - bookedPatients

	if slot.AvailableSpots > 0 {
		slot.IsAvailable = true
		slot.Status = "available"
	} else {
		slot.IsAvailable = false
		slot.Status = "booking_full"
	}

	c.JSON(http.StatusOK, gin.H{
		"slot": slot,
	})
}

// UpdateDoctorTimeSlot - Update a time slot
// PUT /doctor-time-slots/:id
func UpdateDoctorTimeSlot(c *gin.Context) {
	slotID := c.Param("id")

	// Validate slot_id UUID format
	if _, err := uuid.Parse(slotID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid slot_id format. Must be a valid UUID",
		})
		return
	}

	var input UpdateDoctorTimeSlotInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.SendValidationError(c, "Invalid input data", err.Error())
		return
	}

	// Validate slot_type if provided
	if input.SlotType != nil && *input.SlotType != "clinic_visit" && *input.SlotType != "video_consultation" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid slot_type. Must be one of: clinic_visit, video_consultation",
		})
		return
	}

	// Validate time format if provided
	if input.StartTime != nil {
		if _, err := time.Parse("15:04", *input.StartTime); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid start_time format. Use HH:MM format",
			})
			return
		}
	}

	if input.EndTime != nil {
		if _, err := time.Parse("15:04", *input.EndTime); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid end_time format. Use HH:MM format",
			})
			return
		}
	}

	// Build update query dynamically
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if input.SlotType != nil {
		setParts = append(setParts, fmt.Sprintf("slot_type = $%d", argIndex))
		args = append(args, *input.SlotType)
		argIndex++
	}

	if input.StartTime != nil {
		setParts = append(setParts, fmt.Sprintf("start_time = $%d", argIndex))
		args = append(args, *input.StartTime)
		argIndex++
	}

	if input.EndTime != nil {
		setParts = append(setParts, fmt.Sprintf("end_time = $%d", argIndex))
		args = append(args, *input.EndTime)
		argIndex++
	}

	if input.MaxPatients != nil {
		setParts = append(setParts, fmt.Sprintf("max_patients = $%d", argIndex))
		args = append(args, *input.MaxPatients)
		argIndex++
	}

	if input.Notes != nil {
		setParts = append(setParts, fmt.Sprintf("notes = $%d", argIndex))
		args = append(args, *input.Notes)
		argIndex++
	}

	if input.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *input.IsActive)
		argIndex++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No fields to update",
		})
		return
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add slot_id to args
	args = append(args, slotID)

	// Build the UPDATE query
	query := fmt.Sprintf(`
		UPDATE doctor_time_slots 
		SET %s
		WHERE id = $%d
		RETURNING id, doctor_id, clinic_id, specific_date, day_of_week,
		          slot_type, start_time, end_time, max_patients, notes,
		          is_active, created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex)

	var slot DoctorTimeSlotResponse
	var slotDate *string
	var dayOfWeek *int

	err := config.DB.QueryRow(query, args...).Scan(
		&slot.ID, &slot.DoctorID, &slot.ClinicID, &slotDate, &dayOfWeek,
		&slot.SlotType, &slot.StartTime, &slot.EndTime, &slot.MaxPatients, &slot.Notes,
		&slot.IsActive, &slot.CreatedAt, &slot.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Time slot not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update time slot",
		})
		return
	}

	// Set date or day_of_week
	if slotDate != nil {
		slot.Date = *slotDate
	} else if dayOfWeek != nil {
		slot.DayOfWeek = dayOfWeek
		dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		slot.Date = fmt.Sprintf("Every %s", dayNames[*dayOfWeek])
	}

	// Get booked count
	var bookedPatients int
	config.DB.QueryRow(`
		SELECT COALESCE(COUNT(*), 0)
		FROM appointments 
		WHERE slot_id = $1 AND status IN ('confirmed', 'completed')
	`, slot.ID).Scan(&bookedPatients)

	slot.BookedPatients = bookedPatients
	slot.AvailableSpots = slot.MaxPatients - bookedPatients

	if slot.AvailableSpots > 0 {
		slot.IsAvailable = true
		slot.Status = "available"
	} else {
		slot.IsAvailable = false
		slot.Status = "booking_full"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Time slot updated successfully",
		"slot":    slot,
	})
}

// DeleteDoctorTimeSlot - Soft delete a time slot
// DELETE /doctor-time-slots/:id
func DeleteDoctorTimeSlot(c *gin.Context) {
	slotID := c.Param("id")

	// Validate slot_id UUID format
	if _, err := uuid.Parse(slotID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid slot_id format. Must be a valid UUID",
		})
		return
	}

	// Soft delete by setting is_active to false
	result, err := config.DB.Exec(`
		UPDATE doctor_time_slots 
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`, time.Now(), slotID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete time slot",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check deletion result",
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Time slot not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Time slot deleted successfully",
		"slot_id": slotID,
	})
}

