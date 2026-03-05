# 🚀 Simple Follow-Up System Implementation

## 🎯 **Follow-Up Rules Summary**

### **Regular Appointment**
- Patient books a Clinic Visit or Video Consultation
- This is considered a regular appointment

### **Follow-Up Appointment** 
- After a regular appointment, the patient can book a follow-up
- Follow-up is linked to the same doctor and department
- Two follow-up types:
  - Clinic Visit Follow-Up
  - Video Consultation Follow-Up

### **Free Follow-Up**
- First follow-up after a regular appointment is free if booked within 5 days
- Once the free follow-up is used or 5 days pass, subsequent follow-ups become paid

### **Renewal Free Follow-Up**
- If the patient books a new regular appointment with the same doctor + department
- They can again get a free follow-up (renewal follow-up)
- System must track this per patient, per doctor, per department

---

## 🔧 **Simple Implementation**

### **1. Simple Follow-Up Check Function**

```go
// SimpleFollowUpStatus represents follow-up status
type SimpleFollowUpStatus struct {
    IsFree      bool   `json:"is_free"`
    Message     string `json:"message"`
    ColorCode   string `json:"color_code"`
    DaysLeft    int    `json:"days_left,omitempty"`
}

// CheckSimpleFollowUp checks follow-up eligibility for a patient with specific doctor+department
func CheckSimpleFollowUp(patientID, doctorID, departmentID string, db *sql.DB) SimpleFollowUpStatus {
    // Query to find the most recent regular appointment and check follow-up usage
    query := `
        WITH last_regular AS (
            SELECT 
                appointment_date,
                id as appointment_id
            FROM appointments
            WHERE clinic_patient_id = $1
              AND doctor_id = $2
              AND (department_id = $3 OR (department_id IS NULL AND $3 = ''))
              AND consultation_type IN ('clinic_visit', 'video_consultation')
              AND status IN ('completed', 'confirmed')
            ORDER BY appointment_date DESC, appointment_time DESC
            LIMIT 1
        ),
        follow_up_used AS (
            SELECT COUNT(*) > 0 as has_used_followup
            FROM appointments a
            JOIN last_regular lr ON a.clinic_patient_id = $1
            WHERE a.clinic_patient_id = $1
              AND a.doctor_id = $2
              AND (a.department_id = $3 OR (a.department_id IS NULL AND $3 = ''))
              AND a.consultation_type LIKE 'follow-up%'
              AND a.appointment_date >= lr.appointment_date
              AND a.status IN ('completed', 'confirmed')
        )
        SELECT 
            lr.appointment_date,
            lr.appointment_id,
            fu.has_used_followup
        FROM last_regular lr
        CROSS JOIN follow_up_used fu
    `
    
    var lastAppointmentDate time.Time
    var lastAppointmentID string
    var hasUsedFollowUp bool
    
    err := db.QueryRow(query, patientID, doctorID, departmentID).Scan(
        &lastAppointmentDate, &lastAppointmentID, &hasUsedFollowUp,
    )
    
    if err != nil {
        // No previous appointment found
        return SimpleFollowUpStatus{
            IsFree:    false,
            Message:   "No previous appointment with this doctor and department",
            ColorCode: "gray",
        }
    }
    
    // Calculate days since last regular appointment
    daysSince := int(time.Since(lastAppointmentDate).Hours() / 24)
    
    // Simple follow-up logic
    if daysSince <= 5 && !hasUsedFollowUp {
        // Free follow-up available
        daysLeft := 5 - daysSince
        return SimpleFollowUpStatus{
            IsFree:    true,
            Message:   fmt.Sprintf("Free follow-up available (%d days left)", daysLeft),
            ColorCode: "green",
            DaysLeft:  daysLeft,
        }
    } else if daysSince <= 5 && hasUsedFollowUp {
        // Free follow-up already used
        return SimpleFollowUpStatus{
            IsFree:    false,
            Message:   "Free follow-up already used (payment required)",
            ColorCode: "orange",
        }
    } else {
        // Follow-up period expired
        return SimpleFollowUpStatus{
            IsFree:    false,
            Message:   "Follow-up period expired (payment required)",
            ColorCode: "orange",
        }
    }
}
```

### **2. Updated CreateSimpleAppointment Function**

```go
// CreateSimpleAppointment - Simplified appointment creation with simple follow-up logic
func CreateSimpleAppointment(c *gin.Context) {
    var input SimpleAppointmentInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid input",
            "details": err.Error(),
        })
        return
    }

    // Auto-detect follow-up based on consultation_type
    if input.ConsultationType == "follow-up-via-clinic" || input.ConsultationType == "follow-up-via-video" {
        input.IsFollowUp = true
    }

    // Step 1: Validate clinic patient exists and belongs to this clinic
    var clinicPatientClinicID string
    err := config.DB.QueryRow(`
        SELECT clinic_id FROM clinic_patients 
        WHERE id = $1 AND is_active = true
    `, input.ClinicPatientID).Scan(&clinicPatientClinicID)

    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "Patient not found",
        })
        return
    }

    if clinicPatientClinicID != input.ClinicID {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Patient belongs to different clinic",
        })
        return
    }

    // Step 2: Simple follow-up validation
    var isFreeFollowUp bool = false
    
    if input.IsFollowUp {
        // Use simple follow-up check
        followUpStatus := CheckSimpleFollowUp(input.ClinicPatientID, input.DoctorID, input.DepartmentID, config.DB)
        
        if !followUpStatus.IsFree && followUpStatus.ColorCode == "gray" {
            // No previous appointment
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Not eligible for follow-up",
                "message": followUpStatus.Message,
            })
            return
        }
        
        isFreeFollowUp = followUpStatus.IsFree
        log.Printf("✅ Simple follow-up check: Free=%v, Message=%s", isFreeFollowUp, followUpStatus.Message)
    }

    // Step 3: Validate payment logic
    if !input.IsFollowUp || (input.IsFollowUp && !isFreeFollowUp) {
        // Regular appointments OR Paid follow-ups require payment_method
        if input.PaymentMethod == nil {
            var message string
            if input.IsFollowUp {
                message = "This follow-up requires payment (free follow-up period expired or already used)"
            } else {
                message = "Please specify payment_method for appointments"
            }
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Payment method required",
                "message": message,
            })
            return
        }
        
        if *input.PaymentMethod == "pay_now" {
            if input.PaymentType == nil || *input.PaymentType == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "Payment type required",
                    "message": "When payment_method is 'pay_now', you must provide payment_type (cash, card, or upi)",
                })
                return
            }
        }
    }

    // Step 4: Validate individual slot is available
    var slotClinicID string
    var slotStart, slotEnd string
    var isBooked bool
    var slotStatus string
    var maxPatients, availableCount int

    err = config.DB.QueryRow(`
        SELECT clinic_id, slot_start, slot_end, is_booked, status, max_patients, available_count
        FROM doctor_individual_slots
        WHERE id = $1
    `, input.IndividualSlotID).Scan(&slotClinicID, &slotStart, &slotEnd, &isBooked, &slotStatus, &maxPatients, &availableCount)

    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "Slot not found",
        })
        return
    }

    if slotClinicID != input.ClinicID {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Slot belongs to different clinic",
        })
        return
    }

    if availableCount <= 0 || slotStatus != "available" {
        c.JSON(http.StatusConflict, gin.H{
            "error":   "Slot not available",
            "message": "This slot is fully booked. Please select another slot.",
        })
        return
    }

    // Step 5: Parse dates
    appointmentDate, err := time.Parse("2006-01-02", input.AppointmentDate)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid date format. Use YYYY-MM-DD",
        })
        return
    }

    appointmentTime, err := time.Parse("2006-01-02 15:04:05", input.AppointmentTime)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid time format. Use YYYY-MM-DD HH:MM:SS",
        })
        return
    }

    // Step 6: Get doctor and calculate fee
    var doctor models.DoctorInfo
    var consultationFee, followUpFee *float64

    err = config.DB.QueryRow(`
        SELECT d.id, d.doctor_code, u.first_name, u.last_name,
               COALESCE(cdl.consultation_fee_offline, d.consultation_fee) as consultation_fee,
               COALESCE(cdl.follow_up_fee, d.follow_up_fee) as follow_up_fee
        FROM doctors d
        JOIN users u ON u.id = d.user_id
        LEFT JOIN clinic_doctor_links cdl ON cdl.doctor_id = d.id AND cdl.clinic_id = $1
        WHERE d.id = $2 AND d.is_active = true
    `, input.ClinicID, input.DoctorID).Scan(
        &doctor.ID, &doctor.DoctorCode, &doctor.FirstName, &doctor.LastName,
        &consultationFee, &followUpFee,
    )

    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "Doctor not found",
        })
        return
    }

    // Calculate fee
    feeAmount := 0.0
    if (input.ConsultationType == "follow-up-via-clinic" || input.ConsultationType == "follow-up-via-video") && followUpFee != nil {
        feeAmount = *followUpFee
    } else if consultationFee != nil {
        feeAmount = *consultationFee
    }

    // Step 7: Generate booking number and token
    bookingNumber, err := utils.GenerateBookingNumber(doctor.DoctorCode, appointmentTime)
    if err != nil {
        bookingNumber = "BN" + time.Now().Format("20060102150405")
    }

    tokenNumber, err := utils.GenerateTokenNumber(input.DoctorID, input.ClinicID, appointmentDate)
    if err != nil {
        tokenNumber = 1
    }

    // Step 8: Create appointment
    var appointment models.Appointment
    appointmentDateStr := appointmentDate.Format("2006-01-02")

    // Set payment status based on follow-up type
    var paymentStatus string
    var paymentMode *string

    // Simple payment logic
    if input.IsFollowUp && isFreeFollowUp {
        paymentStatus = "waived"
        paymentMode = nil
        feeAmount = 0.0 // No fee for free follow-ups
    } else if input.PaymentMethod != nil {
        switch *input.PaymentMethod {
        case "pay_now":
            paymentStatus = "paid"
            paymentMode = input.PaymentType
        case "pay_later":
            paymentStatus = "pending"
            paymentMode = nil
        case "way_off":
            paymentStatus = "waived"
            paymentMode = nil
        }
    } else {
        paymentStatus = "pending"
        paymentMode = nil
    }

    err = config.DB.QueryRow(`
        INSERT INTO appointments (
            clinic_patient_id, clinic_id, doctor_id, department_id, booking_number, token_number,
            appointment_date, appointment_time, duration_minutes, consultation_type,
            reason, notes, fee_amount, payment_mode, payment_status, status, individual_slot_id
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
        RETURNING id, clinic_patient_id, clinic_id, doctor_id, booking_number, token_number,
                  appointment_date, appointment_time, duration_minutes, consultation_type,
                  reason, notes, status, fee_amount, payment_status, payment_mode, created_at
    `, input.ClinicPatientID, input.ClinicID, input.DoctorID, input.DepartmentID, bookingNumber, tokenNumber,
        appointmentDateStr, appointmentTime, 5, input.ConsultationType,
        input.Reason, input.Notes, feeAmount, paymentMode, 
        paymentStatus, "confirmed", input.IndividualSlotID).Scan(
        &appointment.ID, &appointment.ClinicPatientID, &appointment.ClinicID, &appointment.DoctorID,
        &appointment.BookingNumber, &appointment.TokenNumber, &appointment.AppointmentDate,
        &appointment.AppointmentTime, &appointment.DurationMinutes, &appointment.ConsultationType,
        &appointment.Reason, &appointment.Notes, &appointment.Status, &appointment.FeeAmount,
        &appointment.PaymentStatus, &appointment.PaymentMode, &appointment.CreatedAt,
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Failed to create appointment",
            "details": err.Error(),
        })
        return
    }

    // Step 9: Update slot availability
    result, err := config.DB.Exec(`
        UPDATE doctor_individual_slots
        SET available_count = available_count - 1,
            is_booked = CASE WHEN available_count - 1 <= 0 THEN true ELSE is_booked END,
            status = CASE WHEN available_count - 1 <= 0 THEN 'booked' ELSE status END,
            booked_appointment_id = CASE WHEN available_count - 1 <= 0 THEN $1 ELSE booked_appointment_id END,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        AND available_count > 0
        AND status = 'available'
    `, appointment.ID, input.IndividualSlotID)

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "Failed to update slot availability",
            "details": err.Error(),
        })
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusConflict, gin.H{
            "error":   "Slot just got booked",
            "message": "This slot was just booked by another patient. Please select another slot.",
        })
        return
    }

    // Step 10: Success response with simple follow-up info
    response := gin.H{
        "message":     "Appointment created successfully",
        "appointment": appointment,
    }
    
    if input.IsFollowUp {
        response["is_free_followup"] = isFreeFollowUp
        if isFreeFollowUp {
            response["followup_type"] = "free"
            response["followup_message"] = "This is a FREE follow-up"
        } else {
            response["followup_type"] = "paid"
            response["followup_message"] = "This is a PAID follow-up"
        }
    } else {
        // Regular appointment - inform about follow-up eligibility
        response["is_regular_appointment"] = true
        response["followup_granted"] = true
        response["followup_message"] = "Free follow-up eligibility granted (valid for 5 days)"
        
        expiryDate := appointmentDate.AddDate(0, 0, 5)
        response["followup_valid_until"] = expiryDate.Format("2006-01-02")
    }
    
    c.JSON(http.StatusCreated, response)
}
```

### **3. Updated Patient List Function**

```go
// SimplePatientResponse represents patient with simple follow-up status
type SimplePatientResponse struct {
    ID                  string                `json:"id"`
    ClinicID            string                `json:"clinic_id"`
    FirstName           string                `json:"first_name"`
    LastName            string                `json:"last_name"`
    Phone               string                `json:"phone"`
    Email               *string               `json:"email,omitempty"`
    IsActive            bool                  `json:"is_active"`
    CreatedAt           time.Time             `json:"created_at"`
    UpdatedAt           time.Time             `json:"updated_at"`
    FollowUpStatus      *SimpleFollowUpStatus `json:"follow_up_status,omitempty"`
}

// ListSimplePatients - List patients with simple follow-up status
func ListSimplePatients(c *gin.Context) {
    clinicID := c.Query("clinic_id")
    search := c.Query("search")
    onlyActive := c.DefaultQuery("only_active", "true")
    
    // Optional parameters for follow-up status check
    doctorID := c.Query("doctor_id")
    departmentID := c.Query("department_id")

    if clinicID == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "clinic_id is required",
        })
        return
    }

    // Build query
    query := `
        SELECT id, clinic_id, first_name, last_name, phone, email, 
               is_active, created_at, updated_at
        FROM clinic_patients
        WHERE clinic_id = $1
    `
    args := []interface{}{clinicID}
    argIndex := 2

    if onlyActive == "true" {
        query += fmt.Sprintf(" AND is_active = $%d", argIndex)
        args = append(args, true)
        argIndex++
    }

    if search != "" {
        query += fmt.Sprintf(` AND (
            LOWER(first_name) LIKE LOWER($%d) OR 
            LOWER(last_name) LIKE LOWER($%d) OR 
            LOWER(phone) LIKE LOWER($%d) OR 
            LOWER(mo_id) LIKE LOWER($%d)
        )`, argIndex, argIndex, argIndex, argIndex)
        args = append(args, "%"+search+"%")
    }

    query += " ORDER BY created_at DESC"

    rows, err := config.DB.Query(query, args...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch patients",
        })
        return
    }
    defer rows.Close()

    var patients []SimplePatientResponse
    for rows.Next() {
        var patient SimplePatientResponse
        err := rows.Scan(
            &patient.ID, &patient.ClinicID, &patient.FirstName, &patient.LastName,
            &patient.Phone, &patient.Email, &patient.IsActive, 
            &patient.CreatedAt, &patient.UpdatedAt,
        )
        if err != nil {
            continue
        }
        
        // Add simple follow-up status if doctor and department provided
        if doctorID != "" {
            followUpStatus := CheckSimpleFollowUp(patient.ID, doctorID, departmentID, config.DB)
            patient.FollowUpStatus = &followUpStatus
        } else {
            // No doctor selected - show neutral status
            patient.FollowUpStatus = &SimpleFollowUpStatus{
                IsFree:    false,
                Message:   "Please select a doctor to check follow-up status",
                ColorCode: "gray",
            }
        }
        
        patients = append(patients, patient)
    }

    c.JSON(http.StatusOK, gin.H{
        "clinic_id": clinicID,
        "total":     len(patients),
        "patients":  patients,
    })
}
```

### **4. Updated Routes**

```go
// Add simple routes
func SetupSimpleRoutes(r *gin.Engine) {
    api := r.Group("/api")
    
    // Simple patient list with follow-up status
    api.GET("/simple-patients", ListSimplePatients)
    
    // Simple appointment creation
    api.POST("/appointments/simple", CreateSimpleAppointment)
}
```

---

## 🎯 **Simple API Usage**

### **1. Get Patients with Follow-Up Status**
```
GET /api/simple-patients?clinic_id=xxx&doctor_id=xxx&department_id=xxx&search=xxx
```

**Response:**
```json
{
  "clinic_id": "clinic-123",
  "total": 2,
  "patients": [
    {
      "id": "patient-123",
      "first_name": "John",
      "last_name": "Doe",
      "phone": "1234567890",
      "follow_up_status": {
        "is_free": true,
        "message": "Free follow-up available (3 days left)",
        "color_code": "green",
        "days_left": 3
      }
    },
    {
      "id": "patient-456",
      "first_name": "Jane",
      "last_name": "Smith",
      "phone": "0987654321",
      "follow_up_status": {
        "is_free": false,
        "message": "Free follow-up already used (payment required)",
        "color_code": "orange"
      }
    }
  ]
}
```

### **2. Create Appointment**
```
POST /api/appointments/simple
```

**Request:**
```json
{
  "clinic_patient_id": "patient-123",
  "doctor_id": "doctor-456",
  "clinic_id": "clinic-789",
  "department_id": "dept-101",
  "individual_slot_id": "slot-001",
  "appointment_date": "2025-01-15",
  "appointment_time": "2025-01-15 10:00:00",
  "consultation_type": "follow-up-via-clinic"
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "appt-123",
    "consultation_type": "follow-up-via-clinic",
    "fee_amount": 0.0,
    "payment_status": "waived"
  },
  "is_free_followup": true,
  "followup_type": "free",
  "followup_message": "This is a FREE follow-up"
}
```

---

## ✅ **Benefits of Simple Implementation**

### **Performance:**
- ✅ **Fast queries** - Single SQL query for follow-up check
- ✅ **No complex joins** - Simple appointment history check
- ✅ **No loading issues** - Lightweight database operations
- ✅ **Easy caching** - Simple data structure

### **Maintenance:**
- ✅ **Easy to understand** - Clear follow-up rules
- ✅ **Easy to debug** - Simple logic flow
- ✅ **Easy to modify** - Minimal code changes
- ✅ **No complex tables** - Uses existing appointments table

### **Functionality:**
- ✅ **Same follow-up rules** - 5 days, free/paid logic
- ✅ **Same user experience** - Green/orange/gray colors
- ✅ **Same business logic** - Doctor+department specific
- ✅ **Renewal support** - New regular appointment resets follow-up

This simple implementation gives you exactly the follow-up rules you specified with minimal complexity and no loading issues!
