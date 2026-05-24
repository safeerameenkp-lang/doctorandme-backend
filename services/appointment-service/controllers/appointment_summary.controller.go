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

	// Get payment breakdown (collections) for today using payment_method and paid_amount
	var cashRevenue, cardRevenue, upiRevenue, totalRevenue float64
	paymentQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN LOWER(payment_method) = 'cash' THEN paid_amount ELSE 0 END), 0) as cash_rev,
			COALESCE(SUM(CASE WHEN LOWER(payment_method) = 'card' THEN paid_amount ELSE 0 END), 0) as card_rev,
			COALESCE(SUM(CASE WHEN LOWER(payment_method) = 'upi' THEN paid_amount ELSE 0 END), 0) as upi_rev,
			COALESCE(SUM(paid_amount), 0) as total_rev
		FROM appointments 
		WHERE clinic_id = $1 AND appointment_date = $2 AND payment_status = 'paid' AND status != 'cancelled'
	`
	paymentArgs := []interface{}{clinicID, date}
	paymentArgIndex := 3

	if doctorID != "" && doctorID != "all" {
		paymentQuery += fmt.Sprintf(" AND doctor_id = $%d", paymentArgIndex)
		paymentArgs = append(paymentArgs, doctorID)
		paymentArgIndex++
	}

	_ = config.DB.QueryRowContext(ctx, paymentQuery, paymentArgs...).Scan(&cashRevenue, &cardRevenue, &upiRevenue, &totalRevenue)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"clinic_id": clinicID,
		"date":      date,
		"doctor_id": doctorID,
		"status":    statusFilter,
		"summary":   summary,
		"payments": gin.H{
			"cash":  cashRevenue,
			"card":  cardRevenue,
			"upi":   upiRevenue,
			"total": totalRevenue,
		},
	})
}

// GetCollections - Get collection breakdown by payment methods
// GET /appointments/collections?clinic_id=...&date=...&doctor_id=...
func GetCollections(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	clinicID := c.Query("clinic_id")
	date := c.Query("date")
	doctorID := c.Query("doctor_id")

	if clinicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clinic_id is required"})
		return
	}

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN LOWER(payment_method) = 'cash' THEN paid_amount ELSE 0 END), 0) as cash_total,
			COALESCE(SUM(CASE WHEN LOWER(payment_method) = 'upi' THEN paid_amount ELSE 0 END), 0) as upi_total,
			COALESCE(SUM(CASE WHEN LOWER(payment_method) = 'card' THEN paid_amount ELSE 0 END), 0) as card_total,
			COALESCE(SUM(paid_amount), 0) as total_collection
		FROM appointments 
		WHERE clinic_id = $1 AND payment_status = 'paid' AND status != 'cancelled'
	`
	args := []interface{}{clinicID}
	argIndex := 2

	if date != "" {
		query += fmt.Sprintf(" AND appointment_date = $%d", argIndex)
		args = append(args, date)
		argIndex++
	}

	if doctorID != "" && doctorID != "all" {
		query += fmt.Sprintf(" AND doctor_id = $%d", argIndex)
		args = append(args, doctorID)
		argIndex++
	}

	var cashTotal, upiTotal, cardTotal, totalCollection float64
	err := config.DB.QueryRowContext(ctx, query, args...).Scan(&cashTotal, &upiTotal, &cardTotal, &totalCollection)
	if err != nil {
		log.Printf("ERROR: GetCollections failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collections"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_collection": totalCollection,
			"cash_total":       cashTotal,
			"upi_total":        upiTotal,
			"card_total":       cardTotal,
		},
	})
}

