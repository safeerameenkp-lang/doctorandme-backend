package utils

import (
    "appointment-service/config"
    "appointment-service/models"
    "fmt"
    "time"
)

// CheckDoctorAvailability checks if a doctor is available at a specific time
func CheckDoctorAvailability(doctorID string, appointmentTime time.Time, durationMinutes *int) (bool, error) {
    // Set default duration
    duration := 12
    if durationMinutes != nil {
        duration = *durationMinutes
    }

    endTime := appointmentTime.Add(time.Duration(duration) * time.Minute)

    // Check for overlapping appointments
    var count int
    err := config.DB.QueryRow(`
        SELECT COUNT(*) FROM appointments
        WHERE doctor_id = $1 
        AND status NOT IN ('cancelled', 'no_show')
        AND (
            (appointment_time <= $2 AND appointment_time + INTERVAL '1 minute' * duration_minutes > $3)
            OR (appointment_time < $4 AND appointment_time + INTERVAL '1 minute' * duration_minutes >= $5)
        )
    `, doctorID, appointmentTime, appointmentTime, endTime, endTime).Scan(&count)

    if err != nil {
        return false, err
    }

    return count == 0, nil
}

// CalculateAppointmentFee calculates the fee based on consultation type and doctor rules
func CalculateAppointmentFee(doctor models.DoctorInfo, consultationType string, patientID string) *float64 {
    switch consultationType {
    case "new":
        return doctor.ConsultationFee
    case "followup":
        // Check if this is within the free follow-up period
        var lastAppointment time.Time
        err := config.DB.QueryRow(`
            SELECT MAX(appointment_time) FROM appointments
            WHERE patient_id = $1 AND doctor_id = $2 AND status = 'completed'
        `, patientID, doctor.ID).Scan(&lastAppointment)
        
        if err == nil && !lastAppointment.IsZero() {
            daysSinceLastVisit := int(time.Since(lastAppointment).Hours() / 24)
            if doctor.FollowUpDays != nil && daysSinceLastVisit <= *doctor.FollowUpDays {
                // Free follow-up
                return nil
            }
        }
        return doctor.FollowUpFee
    case "walkin", "emergency":
        // Emergency and walk-in appointments might have different pricing
        if doctor.ConsultationFee != nil {
            // Add 20% premium for emergency/walk-in
            fee := *doctor.ConsultationFee * 1.2
            return &fee
        }
        return doctor.ConsultationFee
    default:
        return doctor.ConsultationFee
    }
}

// GenerateBookingNumber generates a unique booking number
func GenerateBookingNumber(doctorCode *string, appointmentTime time.Time) (string, error) {
    // Get clinic code from doctor
    var clinicCode string
    err := config.DB.QueryRow(`
        SELECT c.clinic_code FROM clinics c
        JOIN doctors d ON d.clinic_id = c.id
        WHERE d.id = (SELECT id FROM doctors WHERE doctor_code = $1 LIMIT 1)
    `, doctorCode).Scan(&clinicCode)
    if err != nil {
        return "", err
    }

    // Use clinic code if doctor code is not available
    code := clinicCode
    if doctorCode != nil && *doctorCode != "" {
        code = *doctorCode
    }

    dateStr := appointmentTime.Format("20060102")
    
    // Get next serial number for the day
    var serialNo int
    err = config.DB.QueryRow(`
        SELECT COALESCE(MAX(CAST(SUBSTRING(booking_number FROM '.*-.*-(.*)$') AS INTEGER)), 0) + 1
        FROM appointments
        WHERE booking_number LIKE $1
    `, fmt.Sprintf("%s-%s-%%", code, dateStr)).Scan(&serialNo)
    if err != nil {
        return "", err
    }

    bookingNumber := fmt.Sprintf("%s-%s-%04d", code, dateStr, serialNo)
    return bookingNumber, nil
}

// GenerateTimeSlots generates available time slots for a doctor on a specific date
func GenerateTimeSlots(targetDate time.Time, startTime, endTime time.Time, slotDuration int, doctorID string) []models.TimeSlot {
    var slots []models.TimeSlot
    
    // Combine date with time
    startDateTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
        startTime.Hour(), startTime.Minute(), startTime.Second(), 0, targetDate.Location())
    endDateTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
        endTime.Hour(), endTime.Minute(), endTime.Second(), 0, targetDate.Location())

    current := startDateTime
    for current.Before(endDateTime) {
        slotEnd := current.Add(time.Duration(slotDuration) * time.Minute)
        
        // Check if this slot is booked
        var appointmentID *string
        err := config.DB.QueryRow(`
            SELECT id FROM appointments
            WHERE doctor_id = $1 
            AND appointment_time = $2
            AND status NOT IN ('cancelled', 'no_show')
        `, doctorID, current).Scan(&appointmentID)

        isBooked := err == nil && appointmentID != nil
        
        slot := models.TimeSlot{
            StartTime:     current,
            EndTime:       slotEnd,
            IsAvailable:   !isBooked,
            IsBooked:      isBooked,
            AppointmentID: appointmentID,
        }
        
        slots = append(slots, slot)
        current = slotEnd
    }
    
    return slots
}

// GetAppointmentReports generates various appointment reports
func GetAppointmentReports(reportType string, startDate, endDate time.Time, doctorID *string) ([]models.AppointmentReport, error) {
    var query string
    var args []interface{}
    
    switch reportType {
    case "daily_collection":
        query = `
            SELECT 
                DATE(a.appointment_time) as date,
                a.doctor_id,
                CONCAT(du.first_name, ' ', du.last_name) as doctor_name,
                COUNT(*) as total_bookings,
                COUNT(CASE WHEN a.status = 'completed' THEN 1 END) as completed,
                COUNT(CASE WHEN a.status = 'no_show' THEN 1 END) as no_show,
                COUNT(CASE WHEN a.status = 'cancelled' THEN 1 END) as cancelled,
                COALESCE(SUM(CASE WHEN a.payment_status = 'paid' THEN a.fee_amount ELSE 0 END), 0) as total_revenue,
                COALESCE(SUM(CASE WHEN a.payment_status = 'paid' AND a.payment_mode = 'cash' THEN a.fee_amount ELSE 0 END), 0) as cash_revenue,
                COALESCE(SUM(CASE WHEN a.payment_status = 'paid' AND a.payment_mode = 'card' THEN a.fee_amount ELSE 0 END), 0) as card_revenue,
                COALESCE(SUM(CASE WHEN a.payment_status = 'paid' AND a.payment_mode = 'upi' THEN a.fee_amount ELSE 0 END), 0) as upi_revenue,
                COALESCE(SUM(CASE WHEN a.payment_status = 'pending' THEN a.fee_amount ELSE 0 END), 0) as pending_revenue
            FROM appointments a
            JOIN doctors d ON d.id = a.doctor_id
            JOIN users du ON du.id = d.user_id
            WHERE DATE(a.appointment_time) BETWEEN $1 AND $2
        `
        args = []interface{}{startDate, endDate}
        
        if doctorID != nil {
            query += " AND a.doctor_id = $3"
            args = append(args, *doctorID)
        }
        
        query += " GROUP BY DATE(a.appointment_time), a.doctor_id, du.first_name, du.last_name ORDER BY date DESC"
    }
    
    rows, err := config.DB.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var reports []models.AppointmentReport
    for rows.Next() {
        var report models.AppointmentReport
        err := rows.Scan(
            &report.Date, &report.DoctorID, &report.DoctorName,
            &report.TotalBookings, &report.Completed, &report.NoShow,
            &report.Cancelled, &report.TotalRevenue, &report.CashRevenue,
            &report.CardRevenue, &report.UPIRevenue, &report.PendingRevenue,
        )
        if err != nil {
            return nil, err
        }
        reports = append(reports, report)
    }
    
    return reports, nil
}

