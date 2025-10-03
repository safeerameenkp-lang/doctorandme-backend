package models

import (
    "database/sql/driver"
    "encoding/json"
    "time"
)

type User struct {
    ID           string     `json:"id" db:"id"`
    Email        *string    `json:"email" db:"email"`
    Username     string     `json:"username" db:"username"`
    PasswordHash string     `json:"-" db:"password_hash"`
    FirstName    string     `json:"first_name" db:"first_name"`
    LastName     string     `json:"last_name" db:"last_name"`
    Phone        *string    `json:"phone" db:"phone"`
    DateOfBirth  *time.Time `json:"date_of_birth" db:"date_of_birth"`
    Gender       *string    `json:"gender" db:"gender"`
    IsActive     bool       `json:"is_active" db:"is_active"`
    LastLogin    *time.Time `json:"last_login" db:"last_login"`
    CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type Organization struct {
    ID            string    `json:"id" db:"id"`
    Name          string    `json:"name" db:"name"`
    Email         *string   `json:"email" db:"email"`
    Phone         *string   `json:"phone" db:"phone"`
    Address       *string   `json:"address" db:"address"`
    LicenseNumber *string   `json:"license_number" db:"license_number"`
    IsActive      bool      `json:"is_active" db:"is_active"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type Clinic struct {
    ID             string    `json:"id" db:"id"`
    OrganizationID string    `json:"organization_id" db:"organization_id"`
    ClinicCode     string    `json:"clinic_code" db:"clinic_code"`
    Name           string    `json:"name" db:"name"`
    Email          *string   `json:"email" db:"email"`
    Phone          *string   `json:"phone" db:"phone"`
    Address        *string   `json:"address" db:"address"`
    LicenseNumber  *string   `json:"license_number" db:"license_number"`
    IsActive       bool      `json:"is_active" db:"is_active"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type ExternalService struct {
    ID           string    `json:"id" db:"id"`
    ServiceCode  string    `json:"service_code" db:"service_code"`
    Name         string    `json:"name" db:"name"`
    ServiceType  string    `json:"service_type" db:"service_type"`
    Email        *string   `json:"email" db:"email"`
    Phone        *string   `json:"phone" db:"phone"`
    Address      *string   `json:"address" db:"address"`
    LicenseNumber *string  `json:"license_number" db:"license_number"`
    IsActive     bool      `json:"is_active" db:"is_active"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type ClinicServiceLink struct {
    ID        string    `json:"id" db:"id"`
    ClinicID  string    `json:"clinic_id" db:"clinic_id"`
    ServiceID string    `json:"service_id" db:"service_id"`
    IsDefault bool      `json:"is_default" db:"is_default"`
    IsActive  bool      `json:"is_active" db:"is_active"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
    return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
    if value == nil {
        *j = make(map[string]interface{})
        return nil
    }
    
    bytes, ok := value.([]byte)
    if !ok {
        return nil
    }
    
    return json.Unmarshal(bytes, j)
}

type Doctor struct {
    ID               string     `json:"id" db:"id"`
    UserID           string     `json:"user_id" db:"user_id"`
    ClinicID         string     `json:"clinic_id" db:"clinic_id"`
    DoctorCode       *string    `json:"doctor_code" db:"doctor_code"`
    Specialization   *string    `json:"specialization" db:"specialization"`
    LicenseNumber    *string    `json:"license_number" db:"license_number"`
    ConsultationFee  *float64   `json:"consultation_fee" db:"consultation_fee"`
    FollowUpFee      *float64   `json:"follow_up_fee" db:"follow_up_fee"`
    FollowUpDays     *int       `json:"follow_up_days" db:"follow_up_days"`
    IsMainDoctor     bool       `json:"is_main_doctor" db:"is_main_doctor"`
    IsActive         bool       `json:"is_active" db:"is_active"`
    CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

type ClinicDoctorLink struct {
    ID           string    `json:"id" db:"id"`
    ClinicID     string    `json:"clinic_id" db:"clinic_id"`
    DoctorID     string    `json:"doctor_id" db:"doctor_id"`
    IsActive     bool      `json:"is_active" db:"is_active"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type DoctorSchedule struct {
    ID                  string    `json:"id" db:"id"`
    DoctorID            string    `json:"doctor_id" db:"doctor_id"`
    DayOfWeek           int       `json:"day_of_week" db:"day_of_week"`
    StartTime           string    `json:"start_time" db:"start_time"`
    EndTime             string    `json:"end_time" db:"end_time"`
    SlotDurationMinutes int       `json:"slot_duration_minutes" db:"slot_duration_minutes"`
    IsActive            bool      `json:"is_active" db:"is_active"`
    CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

type Patient struct {
    ID             string     `json:"id" db:"id"`
    UserID         string     `json:"user_id" db:"user_id"`
    MOID           *string    `json:"mo_id" db:"mo_id"`
    MedicalHistory *string    `json:"medical_history" db:"medical_history"`
    Allergies      *string    `json:"allergies" db:"allergies"`
    BloodGroup     *string    `json:"blood_group" db:"blood_group"`
    IsActive       bool       `json:"is_active" db:"is_active"`
    CreatedAt      time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type PatientClinic struct {
    ID        string    `json:"id" db:"id"`
    PatientID string    `json:"patient_id" db:"patient_id"`
    ClinicID  string    `json:"clinic_id" db:"clinic_id"`
    IsPrimary bool      `json:"is_primary" db:"is_primary"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ==================== ADMIN MODELS ====================

type Staff struct {
    ID          string    `json:"id" db:"id"`
    UserID      string    `json:"user_id" db:"user_id"`
    ClinicID    string    `json:"clinic_id" db:"clinic_id"`
    StaffType   string    `json:"staff_type" db:"staff_type"`
    Permissions JSONB     `json:"permissions" db:"permissions"`
    IsActive    bool      `json:"is_active" db:"is_active"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Queue struct {
    ID           string     `json:"id" db:"id"`
    ClinicID     string     `json:"clinic_id" db:"clinic_id"`
    QueueType    string     `json:"queue_type" db:"queue_type"`
    DoctorID     *string    `json:"doctor_id" db:"doctor_id"`
    IsActive     bool       `json:"is_active" db:"is_active"`
    IsPaused     bool       `json:"is_paused" db:"is_paused"`
    CurrentToken int        `json:"current_token" db:"current_token"`
    CreatedAt    time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type QueueToken struct {
    ID            string     `json:"id" db:"id"`
    QueueID       string     `json:"queue_id" db:"queue_id"`
    PatientID     string     `json:"patient_id" db:"patient_id"`
    AppointmentID string     `json:"appointment_id" db:"appointment_id"`
    TokenNumber   int        `json:"token_number" db:"token_number"`
    Status        string     `json:"status" db:"status"`
    Priority      bool       `json:"priority" db:"priority"`
    AssignedAt    time.Time  `json:"assigned_at" db:"assigned_at"`
    CalledAt      *time.Time `json:"called_at" db:"called_at"`
    CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
    CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

type PharmacyInventory struct {
    ID            string     `json:"id" db:"id"`
    ClinicID      string     `json:"clinic_id" db:"clinic_id"`
    MedicineName  string     `json:"medicine_name" db:"medicine_name"`
    GenericName   *string    `json:"generic_name" db:"generic_name"`
    MedicineCode  string     `json:"medicine_code" db:"medicine_code"`
    Category      *string    `json:"category" db:"category"`
    Unit          string     `json:"unit" db:"unit"`
    CurrentStock  int        `json:"current_stock" db:"current_stock"`
    MinStockLevel int        `json:"min_stock_level" db:"min_stock_level"`
    MaxStockLevel int        `json:"max_stock_level" db:"max_stock_level"`
    UnitPrice     float64    `json:"unit_price" db:"unit_price"`
    ExpiryDate    *time.Time `json:"expiry_date" db:"expiry_date"`
    SupplierName  *string    `json:"supplier_name" db:"supplier_name"`
    BatchNumber   *string    `json:"batch_number" db:"batch_number"`
    IsActive      bool       `json:"is_active" db:"is_active"`
    CreatedAt     time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type PharmacySupplier struct {
    ID            string    `json:"id" db:"id"`
    ClinicID      string    `json:"clinic_id" db:"clinic_id"`
    SupplierName  string    `json:"supplier_name" db:"supplier_name"`
    ContactPerson *string   `json:"contact_person" db:"contact_person"`
    Email         *string   `json:"email" db:"email"`
    Phone         *string   `json:"phone" db:"phone"`
    Address       *string   `json:"address" db:"address"`
    IsActive      bool      `json:"is_active" db:"is_active"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type PharmacyDiscount struct {
    ID                string     `json:"id" db:"id"`
    ClinicID          string     `json:"clinic_id" db:"clinic_id"`
    DiscountName      string     `json:"discount_name" db:"discount_name"`
    DiscountType      string     `json:"discount_type" db:"discount_type"`
    DiscountValue     float64    `json:"discount_value" db:"discount_value"`
    MinPurchaseAmount float64    `json:"min_purchase_amount" db:"min_purchase_amount"`
    MaxDiscountAmount *float64   `json:"max_discount_amount" db:"max_discount_amount"`
    ValidFrom         time.Time  `json:"valid_from" db:"valid_from"`
    ValidTo           time.Time  `json:"valid_to" db:"valid_to"`
    IsActive          bool       `json:"is_active" db:"is_active"`
    CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

type LabTest struct {
    ID                      string     `json:"id" db:"id"`
    ClinicID                string     `json:"clinic_id" db:"clinic_id"`
    TestCode                string     `json:"test_code" db:"test_code"`
    TestName                string     `json:"test_name" db:"test_name"`
    TestCategory            *string    `json:"test_category" db:"test_category"`
    Description             *string    `json:"description" db:"description"`
    SampleType              *string    `json:"sample_type" db:"sample_type"`
    PreparationInstructions *string    `json:"preparation_instructions" db:"preparation_instructions"`
    NormalRange             *string    `json:"normal_range" db:"normal_range"`
    Unit                    *string    `json:"unit" db:"unit"`
    Price                   float64    `json:"price" db:"price"`
    TurnaroundTimeHours     int        `json:"turnaround_time_hours" db:"turnaround_time_hours"`
    IsActive                bool       `json:"is_active" db:"is_active"`
    CreatedAt               time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt               time.Time  `json:"updated_at" db:"updated_at"`
}

type LabSampleCollector struct {
    ID             string     `json:"id" db:"id"`
    UserID         string     `json:"user_id" db:"user_id"`
    ClinicID       string     `json:"clinic_id" db:"clinic_id"`
    CollectorCode  *string    `json:"collector_code" db:"collector_code"`
    Specialization *string    `json:"specialization" db:"specialization"`
    IsActive       bool       `json:"is_active" db:"is_active"`
    CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type LabOrder struct {
    ID            string     `json:"id" db:"id"`
    ClinicID      string     `json:"clinic_id" db:"clinic_id"`
    PatientID     string     `json:"patient_id" db:"patient_id"`
    AppointmentID string     `json:"appointment_id" db:"appointment_id"`
    DoctorID      string     `json:"doctor_id" db:"doctor_id"`
    OrderNumber   string     `json:"order_number" db:"order_number"`
    OrderDate     time.Time  `json:"order_date" db:"order_date"`
    Status        string     `json:"status" db:"status"`
    TotalAmount   float64    `json:"total_amount" db:"total_amount"`
    PaymentStatus string     `json:"payment_status" db:"payment_status"`
    CollectorID   *string    `json:"collector_id" db:"collector_id"`
    CollectionDate *time.Time `json:"collection_date" db:"collection_date"`
    CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

type LabResult struct {
    ID                  string     `json:"id" db:"id"`
    OrderID             string     `json:"order_id" db:"order_id"`
    TestID              string     `json:"test_id" db:"test_id"`
    ResultValue         *string    `json:"result_value" db:"result_value"`
    ResultUnit          *string    `json:"result_unit" db:"result_unit"`
    NormalRange         *string    `json:"normal_range" db:"normal_range"`
    Status              string     `json:"status" db:"status"`
    Notes               *string    `json:"notes" db:"notes"`
    UploadedBy          string     `json:"uploaded_by" db:"uploaded_by"`
    UploadedAt          time.Time  `json:"uploaded_at" db:"uploaded_at"`
    IsVisibleToPatient  bool       `json:"is_visible_to_patient" db:"is_visible_to_patient"`
    CreatedAt           time.Time  `json:"created_at" db:"created_at"`
}

type FeeStructure struct {
    ID            string     `json:"id" db:"id"`
    ClinicID      string     `json:"clinic_id" db:"clinic_id"`
    ServiceType   string     `json:"service_type" db:"service_type"`
    ServiceName   string     `json:"service_name" db:"service_name"`
    BaseFee       float64    `json:"base_fee" db:"base_fee"`
    FollowUpFee   *float64   `json:"follow_up_fee" db:"follow_up_fee"`
    FollowUpDays  *int       `json:"follow_up_days" db:"follow_up_days"`
    IsActive      bool       `json:"is_active" db:"is_active"`
    CreatedAt     time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type InsuranceProvider struct {
    ID                  string                 `json:"id" db:"id"`
    ClinicID            string                 `json:"clinic_id" db:"clinic_id"`
    ProviderName        string                 `json:"provider_name" db:"provider_name"`
    ProviderCode        *string                `json:"provider_code" db:"provider_code"`
    ContactDetails      JSONB                  `json:"contact_details" db:"contact_details"`
    ConsultationCovered bool                   `json:"consultation_covered" db:"consultation_covered"`
    MedicinesCovered    bool                   `json:"medicines_covered" db:"medicines_covered"`
    LabTestsCovered     bool                   `json:"lab_tests_covered" db:"lab_tests_covered"`
    CoveragePercentage  float64                `json:"coverage_percentage" db:"coverage_percentage"`
    MaxCoverageAmount   *float64               `json:"max_coverage_amount" db:"max_coverage_amount"`
    IsActive            bool                   `json:"is_active" db:"is_active"`
    CreatedAt           time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
}

type PatientInsurance struct {
    ID                     string     `json:"id" db:"id"`
    PatientID              string     `json:"patient_id" db:"patient_id"`
    ProviderID             string     `json:"provider_id" db:"provider_id"`
    PolicyNumber           string     `json:"policy_number" db:"policy_number"`
    PolicyHolderName       *string    `json:"policy_holder_name" db:"policy_holder_name"`
    RelationshipToPatient  *string    `json:"relationship_to_patient" db:"relationship_to_patient"`
    CoverageStartDate      *time.Time `json:"coverage_start_date" db:"coverage_start_date"`
    CoverageEndDate        *time.Time `json:"coverage_end_date" db:"coverage_end_date"`
    IsActive               bool       `json:"is_active" db:"is_active"`
    CreatedAt              time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

type InsuranceClaim struct {
    ID              string     `json:"id" db:"id"`
    PatientID       string     `json:"patient_id" db:"patient_id"`
    ProviderID      string     `json:"provider_id" db:"provider_id"`
    AppointmentID   string     `json:"appointment_id" db:"appointment_id"`
    ClaimNumber     string     `json:"claim_number" db:"claim_number"`
    ClaimAmount     float64    `json:"claim_amount" db:"claim_amount"`
    CoveredAmount   float64    `json:"covered_amount" db:"covered_amount"`
    PatientPayable  float64    `json:"patient_payable" db:"patient_payable"`
    Status          string     `json:"status" db:"status"`
    SubmissionDate  *time.Time `json:"submission_date" db:"submission_date"`
    ApprovalDate    *time.Time `json:"approval_date" db:"approval_date"`
    RejectionReason *string    `json:"rejection_reason" db:"rejection_reason"`
    CreatedAt       time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type DailyCollection struct {
    ID                string     `json:"id" db:"id"`
    ClinicID          string     `json:"clinic_id" db:"clinic_id"`
    CollectionDate    time.Time  `json:"collection_date" db:"collection_date"`
    ConsultationAmount float64   `json:"consultation_amount" db:"consultation_amount"`
    LabAmount         float64    `json:"lab_amount" db:"lab_amount"`
    PharmacyAmount    float64    `json:"pharmacy_amount" db:"pharmacy_amount"`
    ProcedureAmount   float64    `json:"procedure_amount" db:"procedure_amount"`
    TotalAmount       float64    `json:"total_amount" db:"total_amount"`
    CashAmount        float64    `json:"cash_amount" db:"cash_amount"`
    CardAmount        float64    `json:"card_amount" db:"card_amount"`
    InsuranceAmount   float64    `json:"insurance_amount" db:"insurance_amount"`
    OutstandingAmount float64    `json:"outstanding_amount" db:"outstanding_amount"`
    CreatedAt         time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

type AnalyticsDailyStats struct {
    ID                     string    `json:"id" db:"id"`
    ClinicID               string    `json:"clinic_id" db:"clinic_id"`
    StatDate               time.Time `json:"stat_date" db:"stat_date"`
    TotalPatients          int       `json:"total_patients" db:"total_patients"`
    NewPatients            int       `json:"new_patients" db:"new_patients"`
    TotalAppointments      int       `json:"total_appointments" db:"total_appointments"`
    CompletedAppointments  int       `json:"completed_appointments" db:"completed_appointments"`
    CancelledAppointments  int       `json:"cancelled_appointments" db:"cancelled_appointments"`
    TotalRevenue           float64   `json:"total_revenue" db:"total_revenue"`
    ConsultationRevenue    float64   `json:"consultation_revenue" db:"consultation_revenue"`
    LabRevenue             float64   `json:"lab_revenue" db:"lab_revenue"`
    PharmacyRevenue        float64   `json:"pharmacy_revenue" db:"pharmacy_revenue"`
    AvgWaitTimeMinutes     int       `json:"avg_wait_time_minutes" db:"avg_wait_time_minutes"`
    CreatedAt              time.Time `json:"created_at" db:"created_at"`
}

type AnalyticsDoctorStats struct {
    ID                         string     `json:"id" db:"id"`
    ClinicID                   string     `json:"clinic_id" db:"clinic_id"`
    DoctorID                   string     `json:"doctor_id" db:"doctor_id"`
    StatDate                   time.Time  `json:"stat_date" db:"stat_date"`
    TotalAppointments          int        `json:"total_appointments" db:"total_appointments"`
    CompletedAppointments      int        `json:"completed_appointments" db:"completed_appointments"`
    AvgConsultationTimeMinutes int        `json:"avg_consultation_time_minutes" db:"avg_consultation_time_minutes"`
    TotalRevenue               float64    `json:"total_revenue" db:"total_revenue"`
    PatientSatisfactionScore   *float64   `json:"patient_satisfaction_score" db:"patient_satisfaction_score"`
    CreatedAt                  time.Time  `json:"created_at" db:"created_at"`
}