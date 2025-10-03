package controllers

import (
    "appointment-service/config"
    "appointment-service/models"
    "appointment-service/utils"
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
    "strconv"
    "time"
)

// Report Controllers
func GetDailyCollectionReport(c *gin.Context) {
    // Get query parameters
    startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
    endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
    doctorID := c.Query("doctor_id")

    // Parse dates
    startDate, err := time.Parse("2006-01-02", startDateStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
        return
    }
    endDate, err := time.Parse("2006-01-02", endDateStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
        return
    }

    var doctorIDPtr *string
    if doctorID != "" {
        doctorIDPtr = &doctorID
    }

    reports, err := utils.GetAppointmentReports("daily_collection", startDate, endDate, doctorIDPtr)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "report_type": "daily_collection",
        "start_date":  startDateStr,
        "end_date":    endDateStr,
        "doctor_id":   doctorID,
        "reports":     reports,
        "count":       len(reports),
    })
}

func GetPendingPaymentsReport(c *gin.Context) {
    // Get query parameters
    clinicID := c.Query("clinic_id")
    doctorID := c.Query("doctor_id")
    limitStr := c.DefaultQuery("limit", "100")
    offsetStr := c.DefaultQuery("offset", "0")

    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 100
    }
    offset, err := strconv.Atoi(offsetStr)
    if err != nil || offset < 0 {
        offset = 0
    }

    query := `
        SELECT a.id, a.booking_number, a.appointment_time, a.fee_amount,
               a.payment_status, a.payment_mode, a.status,
               p.user_id, u.first_name, u.last_name, u.phone,
               d.doctor_code, du.first_name as doctor_first_name, du.last_name as doctor_last_name,
               c.clinic_code, c.name as clinic_name
        FROM appointments a
        JOIN patients p ON p.id = a.patient_id
        JOIN users u ON u.id = p.user_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
        WHERE a.payment_status = 'pending' AND a.fee_amount > 0
    `
    args := []interface{}{}
    argIndex := 1

    if clinicID != "" {
        query += fmt.Sprintf(" AND a.clinic_id = $%d", argIndex)
        args = append(args, clinicID)
        argIndex++
    }
    if doctorID != "" {
        query += fmt.Sprintf(" AND a.doctor_id = $%d", argIndex)
        args = append(args, doctorID)
        argIndex++
    }

    query += " ORDER BY a.appointment_time DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
    args = append(args, limit, offset)

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var pendingPayments []gin.H
    totalPendingAmount := 0.0

    for rows.Next() {
        var appointment models.Appointment
        var patientInfo models.PatientInfo
        var doctorInfo models.DoctorInfo
        var clinicInfo models.ClinicInfo

        err := rows.Scan(
            &appointment.ID, &appointment.BookingNumber, &appointment.AppointmentTime,
            &appointment.FeeAmount, &appointment.PaymentStatus, &appointment.PaymentMode, &appointment.Status,
            &patientInfo.UserID, &patientInfo.FirstName, &patientInfo.LastName, &patientInfo.Phone,
            &doctorInfo.DoctorCode, &doctorInfo.FirstName, &doctorInfo.LastName,
            &clinicInfo.ClinicCode, &clinicInfo.Name,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        if appointment.FeeAmount != nil {
            totalPendingAmount += *appointment.FeeAmount
        }

        pendingPayments = append(pendingPayments, gin.H{
            "appointment_id":   appointment.ID,
            "booking_number":   appointment.BookingNumber,
            "appointment_time": appointment.AppointmentTime,
            "fee_amount":       appointment.FeeAmount,
            "payment_status":   appointment.PaymentStatus,
            "payment_mode":     appointment.PaymentMode,
            "status":          appointment.Status,
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

    c.JSON(http.StatusOK, gin.H{
        "report_type":         "pending_payments",
        "clinic_id":           clinicID,
        "doctor_id":           doctorID,
        "pending_payments":    pendingPayments,
        "count":               len(pendingPayments),
        "total_pending_amount": totalPendingAmount,
    })
}

func GetUtilizationReport(c *gin.Context) {
    // Get query parameters
    startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
    endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
    doctorID := c.Query("doctor_id")

    // Parse dates
    startDate, err := time.Parse("2006-01-02", startDateStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
        return
    }
    endDate, err := time.Parse("2006-01-02", endDateStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
        return
    }

    query := `
        SELECT 
            d.id as doctor_id,
            CONCAT(du.first_name, ' ', du.last_name) as doctor_name,
            d.doctor_code,
            COUNT(DISTINCT ds.id) as total_slots_per_day,
            COUNT(DISTINCT a.id) as booked_slots,
            COUNT(CASE WHEN a.status = 'completed' THEN 1 END) as completed_slots,
            COUNT(CASE WHEN a.status = 'no_show' THEN 1 END) as no_show_slots,
            COUNT(CASE WHEN a.status = 'cancelled' THEN 1 END) as cancelled_slots,
            ROUND(
                CASE 
                    WHEN COUNT(DISTINCT ds.id) > 0 
                    THEN (COUNT(DISTINCT a.id)::float / COUNT(DISTINCT ds.id)::float) * 100 
                    ELSE 0 
                END, 2
            ) as utilization_percentage
        FROM doctors d
        JOIN users du ON du.id = d.user_id
        LEFT JOIN doctor_schedules ds ON ds.doctor_id = d.id AND ds.is_active = true
        LEFT JOIN appointments a ON a.doctor_id = d.id 
            AND DATE(a.appointment_time) BETWEEN $1 AND $2
            AND a.status NOT IN ('cancelled')
        WHERE d.is_active = true
    `
    args := []interface{}{startDate, endDate}
    argIndex := 3

    if doctorID != "" {
        query += fmt.Sprintf(" AND d.id = $%d", argIndex)
        args = append(args, doctorID)
        argIndex++
    }

    query += " GROUP BY d.id, du.first_name, du.last_name, d.doctor_code ORDER BY utilization_percentage DESC"

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var utilizationReports []gin.H
    for rows.Next() {
        var doctorID, doctorName, doctorCode string
        var totalSlots, bookedSlots, completedSlots, noShowSlots, cancelledSlots int
        var utilizationPercentage float64

        err := rows.Scan(
            &doctorID, &doctorName, &doctorCode, &totalSlots, &bookedSlots,
            &completedSlots, &noShowSlots, &cancelledSlots, &utilizationPercentage,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        utilizationReports = append(utilizationReports, gin.H{
            "doctor_id":              doctorID,
            "doctor_name":            doctorName,
            "doctor_code":            doctorCode,
            "total_slots_per_day":    totalSlots,
            "booked_slots":           bookedSlots,
            "completed_slots":        completedSlots,
            "no_show_slots":          noShowSlots,
            "cancelled_slots":        cancelledSlots,
            "utilization_percentage": utilizationPercentage,
        })
    }

    c.JSON(http.StatusOK, gin.H{
        "report_type": "utilization",
        "start_date":  startDateStr,
        "end_date":    endDateStr,
        "doctor_id":   doctorID,
        "reports":     utilizationReports,
        "count":       len(utilizationReports),
    })
}

func GetNoShowReport(c *gin.Context) {
    // Get query parameters
    startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
    endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
    doctorID := c.Query("doctor_id")

    // Parse dates
    startDate, err := time.Parse("2006-01-02", startDateStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
        return
    }
    endDate, err := time.Parse("2006-01-02", endDateStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
        return
    }

    query := `
        SELECT 
            a.id, a.booking_number, a.appointment_time, a.fee_amount,
            a.payment_status, a.status,
            p.user_id, u.first_name, u.last_name, u.phone,
            d.doctor_code, du.first_name as doctor_first_name, du.last_name as doctor_last_name,
            c.clinic_code, c.name as clinic_name
        FROM appointments a
        JOIN patients p ON p.id = a.patient_id
        JOIN users u ON u.id = p.user_id
        JOIN doctors d ON d.id = a.doctor_id
        JOIN users du ON du.id = d.user_id
        JOIN clinics c ON c.id = a.clinic_id
        WHERE a.status = 'no_show' 
        AND DATE(a.appointment_time) BETWEEN $1 AND $2
    `
    args := []interface{}{startDate, endDate}
    argIndex := 3

    if doctorID != "" {
        query += fmt.Sprintf(" AND a.doctor_id = $%d", argIndex)
        args = append(args, doctorID)
        argIndex++
    }

    query += " ORDER BY a.appointment_time DESC"

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    defer rows.Close()

    var noShowReports []gin.H
    totalLostRevenue := 0.0

    for rows.Next() {
        var appointment models.Appointment
        var patientInfo models.PatientInfo
        var doctorInfo models.DoctorInfo
        var clinicInfo models.ClinicInfo

        err := rows.Scan(
            &appointment.ID, &appointment.BookingNumber, &appointment.AppointmentTime,
            &appointment.FeeAmount, &appointment.PaymentStatus, &appointment.Status,
            &patientInfo.UserID, &patientInfo.FirstName, &patientInfo.LastName, &patientInfo.Phone,
            &doctorInfo.DoctorCode, &doctorInfo.FirstName, &doctorInfo.LastName,
            &clinicInfo.ClinicCode, &clinicInfo.Name,
        )
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        if appointment.FeeAmount != nil {
            totalLostRevenue += *appointment.FeeAmount
        }

        noShowReports = append(noShowReports, gin.H{
            "appointment_id":   appointment.ID,
            "booking_number":   appointment.BookingNumber,
            "appointment_time": appointment.AppointmentTime,
            "fee_amount":       appointment.FeeAmount,
            "payment_status":   appointment.PaymentStatus,
            "status":          appointment.Status,
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

    c.JSON(http.StatusOK, gin.H{
        "report_type":         "no_show",
        "start_date":          startDateStr,
        "end_date":            endDateStr,
        "doctor_id":           doctorID,
        "no_show_appointments": noShowReports,
        "count":               len(noShowReports),
        "total_lost_revenue":  totalLostRevenue,
    })
}

func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":  "healthy",
        "service": "appointment-service",
        "time":    time.Now(),
    })
}
