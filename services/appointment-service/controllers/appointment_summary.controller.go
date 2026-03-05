package controllers

import (
	"appointment-service/config"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetAppointmentSummary - Get appointment counts by status
// GET /appointments/summary?clinic_id=...&date=...&doctor_id=...&status=...
func GetAppointmentSummary(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	clinicID := c.Query("clinic_id")
	date := c.Query("date")
	doctorID := c.Query("doctor_id")
	statusFilter := c.Query("status")

	if clinicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clinic_id is required"})
		return
	}

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// Use a single optimized query to get counts GROUPED BY status
	query := "SELECT status, COUNT(*) FROM appointments WHERE clinic_id = $1 AND appointment_date = $2"
	args := []interface{}{clinicID, date}
	argIndex := 3

	if doctorID != "" && doctorID != "all" {
		query += fmt.Sprintf(" AND doctor_id = $%d", argIndex)
		args = append(args, doctorID)
		argIndex++
	}

	if statusFilter != "" && statusFilter != "all" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, statusFilter)
		argIndex++
	}

	query += " GROUP BY status"

	rows, err := config.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("ERROR: GetAppointmentSummary failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch summary"})
		return
	}
	defer rows.Close()

	// Initial Summary Map with default zeros
	summary := map[string]int{
		"total":     0,
		"confirmed": 0,
		"arrived":   0,
		"completed": 0,
		"cancelled": 0,
		"no_show":   0,
		"pending":   0,
	}

	total := 0
	counts := make(map[string]int)

	for rows.Next() {
		var s string
		var count int
		if err := rows.Scan(&s, &count); err == nil {
			counts[s] = count
			total += count
		}
	}

	// Apply business logic for summary counts
	summary["total"] = total
	summary["confirmed"] = counts["confirmed"]
	summary["completed"] = counts["completed"]
	summary["cancelled"] = counts["cancelled"]
	summary["no_show"] = counts["no_show"]
	summary["pending"] = counts["pending"]

	// ✅ REFINED ARRIVED LOGIC:
	// "Arrived" count should reflect all patients who reached the clinic.
	// This includes those currently 'arrived', 'in_consultation', and 'completed'.
	summary["arrived"] = counts["arrived"] + counts["in_consultation"] + counts["completed"]

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"clinic_id": clinicID,
		"date":      date,
		"doctor_id": doctorID,
		"status":    statusFilter,
		"summary":   summary,
	})
}
