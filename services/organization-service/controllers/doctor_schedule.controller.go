package controllers

import (
    "organization-service/config"
    "organization-service/models"
    "github.com/gin-gonic/gin"
    "net/http"
    "strconv"
)

// Doctor Schedule Controllers
type CreateDoctorScheduleInput struct {
    DoctorID            string `json:"doctor_id" binding:"required,uuid"`
    DayOfWeek           int    `json:"day_of_week" binding:"required,min=0,max=6"`
    StartTime           string `json:"start_time" binding:"required"`
    EndTime             string `json:"end_time" binding:"required"`
    SlotDurationMinutes int    `json:"slot_duration_minutes" binding:"omitempty,min=5,max=120"`
}

func CreateDoctorSchedule(c *gin.Context) {
    var input CreateDoctorScheduleInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Verify doctor exists
    var doctorExists bool
    err := config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM doctors WHERE id = $1)`, input.DoctorID).Scan(&doctorExists)
    if err != nil || !doctorExists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found"})
        return
    }

    // Check for overlapping schedules
    var overlappingExists bool
    err = config.DB.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM doctor_schedules 
            WHERE doctor_id = $1 AND day_of_week = $2 AND is_active = true
            AND (
                (start_time <= $3 AND end_time > $3) OR
                (start_time < $4 AND end_time >= $4) OR
                (start_time >= $3 AND end_time <= $4)
            )
        )
    `, input.DoctorID, input.DayOfWeek, input.StartTime, input.EndTime).Scan(&overlappingExists)
    if err == nil && overlappingExists {
        c.JSON(http.StatusConflict, gin.H{"error": "Schedule overlaps with existing schedule"})
        return
    }

    // Set default slot duration if not provided
    if input.SlotDurationMinutes == 0 {
        input.SlotDurationMinutes = 12
    }

    var scheduleID string
    err = config.DB.QueryRow(`
        INSERT INTO doctor_schedules (doctor_id, day_of_week, start_time, end_time, slot_duration_minutes)
        VALUES ($1, $2, $3, $4, $5) RETURNING id
    `, input.DoctorID, input.DayOfWeek, input.StartTime, input.EndTime, input.SlotDurationMinutes).Scan(&scheduleID)
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create doctor schedule"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": scheduleID, "message": "Doctor schedule created successfully"})
}

func GetDoctorSchedules(c *gin.Context) {
    doctorID := c.Query("doctor_id")
    dayOfWeek := c.Query("day_of_week")
    
    var query string
    var args []interface{}
    
    if doctorID != "" && dayOfWeek != "" {
        dayInt, err := strconv.Atoi(dayOfWeek)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day_of_week format"})
            return
        }
        query = `
            SELECT id, doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, is_active, created_at
            FROM doctor_schedules WHERE doctor_id = $1 AND day_of_week = $2 ORDER BY start_time
        `
        args = []interface{}{doctorID, dayInt}
    } else if doctorID != "" {
        query = `
            SELECT id, doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, is_active, created_at
            FROM doctor_schedules WHERE doctor_id = $1 ORDER BY day_of_week, start_time
        `
        args = []interface{}{doctorID}
    } else {
        query = `
            SELECT id, doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, is_active, created_at
            FROM doctor_schedules ORDER BY doctor_id, day_of_week, start_time
        `
    }

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctor schedules"})
        return
    }
    defer rows.Close()

    var schedules []models.DoctorSchedule
    for rows.Next() {
        var schedule models.DoctorSchedule
        err := rows.Scan(&schedule.ID, &schedule.DoctorID, &schedule.DayOfWeek, &schedule.StartTime, 
                        &schedule.EndTime, &schedule.SlotDurationMinutes, &schedule.IsActive, &schedule.CreatedAt)
        if err != nil {
            continue
        }
        schedules = append(schedules, schedule)
    }

    c.JSON(http.StatusOK, schedules)
}

func GetDoctorSchedule(c *gin.Context) {
    scheduleID := c.Param("id")
    
    var schedule models.DoctorSchedule
    err := config.DB.QueryRow(`
        SELECT id, doctor_id, day_of_week, start_time, end_time, slot_duration_minutes, is_active, created_at
        FROM doctor_schedules WHERE id = $1
    `, scheduleID).Scan(&schedule.ID, &schedule.DoctorID, &schedule.DayOfWeek, &schedule.StartTime, 
                        &schedule.EndTime, &schedule.SlotDurationMinutes, &schedule.IsActive, &schedule.CreatedAt)
    
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Doctor schedule not found"})
        return
    }

    c.JSON(http.StatusOK, schedule)
}

type UpdateDoctorScheduleInput struct {
    DayOfWeek           *int    `json:"day_of_week" binding:"omitempty,min=0,max=6"`
    StartTime           *string `json:"start_time"`
    EndTime             *string `json:"end_time"`
    SlotDurationMinutes *int    `json:"slot_duration_minutes" binding:"omitempty,min=5,max=120"`
    IsActive            *bool   `json:"is_active"`
}

func UpdateDoctorSchedule(c *gin.Context) {
    scheduleID := c.Param("id")
    var input UpdateDoctorScheduleInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Build dynamic update query
    query := "UPDATE doctor_schedules SET "
    args := []interface{}{}
    argIndex := 1

    if input.DayOfWeek != nil {
        query += "day_of_week = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.DayOfWeek)
        argIndex++
    }
    if input.StartTime != nil {
        query += "start_time = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.StartTime)
        argIndex++
    }
    if input.EndTime != nil {
        query += "end_time = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.EndTime)
        argIndex++
    }
    if input.SlotDurationMinutes != nil {
        query += "slot_duration_minutes = $" + strconv.Itoa(argIndex) + ", "
        args = append(args, *input.SlotDurationMinutes)
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
    args = append(args, scheduleID)

    result, err := config.DB.Exec(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update doctor schedule"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Doctor schedule not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Doctor schedule updated successfully"})
}

func DeleteDoctorSchedule(c *gin.Context) {
    scheduleID := c.Param("id")
    
    result, err := config.DB.Exec(`DELETE FROM doctor_schedules WHERE id = $1`, scheduleID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete doctor schedule"})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Doctor schedule not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Doctor schedule deleted successfully"})
}
