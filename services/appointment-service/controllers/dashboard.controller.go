package controllers

import (
	"appointment-service/config"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DoctorStats struct {
	DoctorID          string  `json:"doctor_id"`
	DoctorName        string  `json:"doctor_name"`
	DoctorImage       string  `json:"doctor_image"`
	TotalAppointments int     `json:"total_appointments"`
	Completed         int     `json:"completed"`
	Revenue           float64 `json:"revenue"`
}

func calculatePercentageChange(current, previous float64) string {
	if previous == 0 {
		if current > 0 {
			return "+100%"
		}
		return "0%"
	}
	change := ((current - previous) / previous) * 100
	if change > 0 {
		return fmt.Sprintf("+%.1f%%", change)
	}
	return fmt.Sprintf("%.1f%%", change)
}

// GetDashboardStats - Get clinic dashboard statistics (Appointments, Revenue, Completions, Trends, and Doctor List)
// GET /appointments/dashboard?clinic_id=...&start_date=...&end_date=...
func GetDashboardStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	clinicID := c.Query("clinic_id")
	if clinicID == "" {
		if c.GetString("clinic_id") != "" {
			clinicID = c.GetString("clinic_id")
		} else if len(c.GetStringSlice("clinic_ids")) > 0 {
			clinicID = c.GetStringSlice("clinic_ids")[0]
		}
		if clinicID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "clinic_id is required"})
			return
		}
	}

	date := c.Query("date")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if date != "" {
		startDateStr = date
		endDateStr = date
	}

	var startDate, endDate time.Time
	var err error

	// Default to today if no dates
	if startDateStr == "" && endDateStr == "" {
		startDate = time.Now().Truncate(24 * time.Hour)
		endDate = startDate
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			startDate = time.Now().Truncate(24 * time.Hour)
		}
		if endDateStr != "" {
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				endDate = startDate
			}
		} else {
			endDate = startDate
		}
	}

	// Calculate previous period dates
	duration := endDate.Sub(startDate) + (24 * time.Hour) // Include both start and end
	prevStartDate := startDate.Add(-duration)
	prevEndDate := endDate.Add(-duration)

	startDateStr = startDate.Format("2006-01-02")
	endDateStr = endDate.Format("2006-01-02")
	prevStartDateStr := prevStartDate.Format("2006-01-02")
	prevEndDateStr := prevEndDate.Format("2006-01-02")

	// Helper for executing queries
	getStats := func(sDate, eDate string) (int, float64, int, int, int) {
		whereClause := "WHERE clinic_id = $1 AND appointment_date >= $2 AND appointment_date <= $3"
		args := []interface{}{clinicID, sDate, eDate}

		var totalAppts int
		config.DB.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM appointments %s", whereClause), args...).Scan(&totalAppts)

		var totalRev float64
		config.DB.QueryRowContext(ctx, fmt.Sprintf(`SELECT COALESCE(SUM(fee_amount), 0) FROM appointments %s AND payment_status IN ('paid', 'completed', 'success')`, whereClause), args...).Scan(&totalRev)

		queryCompleted := fmt.Sprintf(`SELECT consultation_type, COUNT(*) FROM appointments %s AND status = 'completed' GROUP BY consultation_type`, whereClause)
		rows, _ := config.DB.QueryContext(ctx, queryCompleted, args...)
		
		var clinicVisits, onlineVisits, otherVisits int
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var consultType string
				var count int
				if err := rows.Scan(&consultType, &count); err == nil {
					switch consultType {
					case "clinic_visit", "in_person", "offline", "walk_in", "follow_up":
						clinicVisits += count
					case "video_consultation", "video", "online":
						onlineVisits += count
					default:
						otherVisits += count
					}
				}
			}
		}
		return totalAppts, totalRev, clinicVisits, onlineVisits, otherVisits
	}

	// Current Period Stats
	curTotalAppts, curTotalRev, curClinicVisits, curOnlineVisits, curOtherVisits := getStats(startDateStr, endDateStr)
	curCompletedTotal := curClinicVisits + curOnlineVisits + curOtherVisits

	// Previous Period Stats
	prevTotalAppts, prevTotalRev, prevClinicVisits, prevOnlineVisits, prevOtherVisits := getStats(prevStartDateStr, prevEndDateStr)
	prevCompletedTotal := prevClinicVisits + prevOnlineVisits + prevOtherVisits

	// Trends
	apptsTrend := calculatePercentageChange(float64(curTotalAppts), float64(prevTotalAppts))
	revTrend := calculatePercentageChange(curTotalRev, prevTotalRev)
	completedTrend := calculatePercentageChange(float64(curCompletedTotal), float64(prevCompletedTotal))

	// Doctor List
	queryDoctors := `
		SELECT 
			d.id, du.first_name, du.last_name, COALESCE(d.profile_image, ''),
			COUNT(a.id) as total_appts,
			SUM(CASE WHEN a.status = 'completed' THEN 1 ELSE 0 END) as completed_appts,
			SUM(CASE WHEN a.payment_status IN ('paid', 'completed', 'success') THEN COALESCE(a.fee_amount, 0) ELSE 0 END) as rev
		FROM appointments a
		JOIN doctors d ON d.id = a.doctor_id
		JOIN users du ON du.id = d.user_id
		WHERE a.clinic_id = $1 AND a.appointment_date >= $2 AND a.appointment_date <= $3
		GROUP BY d.id, du.first_name, du.last_name, d.profile_image
		ORDER BY rev DESC, total_appts DESC
	`
	rows, err := config.DB.QueryContext(ctx, queryDoctors, clinicID, startDateStr, endDateStr)
	doctors := []DoctorStats{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var doc DoctorStats
			var fName, lName string
			if err := rows.Scan(&doc.DoctorID, &fName, &lName, &doc.DoctorImage, &doc.TotalAppointments, &doc.Completed, &doc.Revenue); err == nil {
				doc.DoctorName = "Dr. " + fName + " " + lName
				doctors = append(doctors, doc)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_appointments":       curTotalAppts,
			"total_appointments_trend": apptsTrend,
			"total_revenue":            curTotalRev,
			"total_revenue_trend":      revTrend,
			"completed_appointments": gin.H{
				"total":         curCompletedTotal,
				"trend":         completedTrend,
				"clinic_visits": curClinicVisits,
				"online_visits": curOnlineVisits,
				"other":         curOtherVisits,
			},
			"doctors_list": doctors,
		},
	})
}
