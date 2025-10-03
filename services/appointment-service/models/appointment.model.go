package models

import (
    "time"
)

type Appointment struct {
    ID               string     `json:"id" db:"id"`
    PatientID        string     `json:"patient_id" db:"patient_id"`
    ClinicID         string     `json:"clinic_id" db:"clinic_id"`
    DoctorID         string     `json:"doctor_id" db:"doctor_id"`
    DepartmentID     *string    `json:"department_id" db:"department_id"`
    BookingNumber    string     `json:"booking_number" db:"booking_number"`
    AppointmentDate  *string    `json:"appointment_date" db:"appointment_date"`
    AppointmentTime  time.Time  `json:"appointment_time" db:"appointment_time"`
    DurationMinutes  int        `json:"duration_minutes" db:"duration_minutes"`
    ConsultationType string     `json:"consultation_type" db:"consultation_type"`
    Reason           *string    `json:"reason" db:"reason"`
    Notes            *string    `json:"notes" db:"notes"`
    Status           string     `json:"status" db:"status"`
    FeeAmount        *float64   `json:"fee_amount" db:"fee_amount"`
    PaymentStatus    string     `json:"payment_status" db:"payment_status"`
    PaymentMode      *string    `json:"payment_mode" db:"payment_mode"`
    IsPriority       bool       `json:"is_priority" db:"is_priority"`
    CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

type PatientCheckin struct {
    ID                string     `json:"id" db:"id"`
    AppointmentID     string     `json:"appointment_id" db:"appointment_id"`
    CheckinTime       time.Time  `json:"checkin_time" db:"checkin_time"`
    VitalsRecorded    bool       `json:"vitals_recorded" db:"vitals_recorded"`
    PaymentCollected   bool       `json:"payment_collected" db:"payment_collected"`
    CheckedInBy       *string    `json:"checked_in_by" db:"checked_in_by"`
    CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

type PatientVitals struct {
    ID            string     `json:"id" db:"id"`
    AppointmentID string     `json:"appointment_id" db:"appointment_id"`
    SystolicBP    *int       `json:"systolic_bp" db:"systolic_bp"`
    DiastolicBP    *int       `json:"diastolic_bp" db:"diastolic_bp"`
    Temperature   *float64   `json:"temperature" db:"temperature"`
    PulseRate     *int       `json:"pulse_rate" db:"pulse_rate"`
    HeightCm      *int       `json:"height_cm" db:"height_cm"`
    WeightKg      *float64   `json:"weight_kg" db:"weight_kg"`
    RecordedBy    *string    `json:"recorded_by" db:"recorded_by"`
    RecordedAt    time.Time  `json:"recorded_at" db:"recorded_at"`
}

// Extended models with related data
type AppointmentWithDetails struct {
    Appointment
    Patient PatientInfo `json:"patient"`
    Doctor  DoctorInfo  `json:"doctor"`
    Clinic  ClinicInfo  `json:"clinic"`
}

type PatientInfo struct {
    ID             string  `json:"id"`
    UserID         string  `json:"user_id"`
    MOID           *string `json:"mo_id"`
    FirstName      string  `json:"first_name"`
    LastName       string  `json:"last_name"`
    Phone          *string `json:"phone"`
    Email          *string `json:"email"`
    MedicalHistory *string `json:"medical_history"`
    Allergies      *string `json:"allergies"`
    BloodGroup     *string `json:"blood_group"`
}

type DoctorInfo struct {
    ID             string  `json:"id"`
    UserID         string  `json:"user_id"`
    DoctorCode     *string `json:"doctor_code"`
    Specialization *string `json:"specialization"`
    FirstName      string  `json:"first_name"`
    LastName       string  `json:"last_name"`
    ConsultationFee *float64 `json:"consultation_fee"`
    FollowUpFee    *float64 `json:"follow_up_fee"`
    FollowUpDays   *int    `json:"follow_up_days"`
}

type ClinicInfo struct {
    ID          string  `json:"id"`
    ClinicCode  string  `json:"clinic_code"`
    Name        string  `json:"name"`
    Phone       *string `json:"phone"`
    Address     *string `json:"address"`
}

type TimeSlot struct {
    StartTime time.Time `json:"start_time"`
    EndTime   time.Time `json:"end_time"`
    IsAvailable bool    `json:"is_available"`
    IsBooked   bool     `json:"is_booked"`
    AppointmentID *string `json:"appointment_id,omitempty"`
}

type BookingNumber struct {
    DoctorCode string `json:"doctor_code"`
    Date       string `json:"date"`
    SerialNo   int    `json:"serial_no"`
}

type AppointmentReport struct {
    Date           string  `json:"date"`
    DoctorID       string  `json:"doctor_id"`
    DoctorName     string  `json:"doctor_name"`
    TotalBookings  int     `json:"total_bookings"`
    Completed      int     `json:"completed"`
    NoShow         int     `json:"no_show"`
    Cancelled      int     `json:"cancelled"`
    TotalRevenue   float64 `json:"total_revenue"`
    CashRevenue    float64 `json:"cash_revenue"`
    CardRevenue    float64 `json:"card_revenue"`
    UPIRevenue     float64 `json:"upi_revenue"`
    PendingRevenue float64 `json:"pending_revenue"`
}
