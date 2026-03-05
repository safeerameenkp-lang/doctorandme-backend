# 📋 CreateSimpleAppointment API - Complete Documentation

## 🎯 **API Overview**
**Endpoint:** `POST /appointments/simple`  
**Purpose:** Create appointments for clinic-specific patients with comprehensive follow-up management  
**Authentication:** Required (JWT Token)

---

## 📥 **Request Structure**

### **JSON Input Format**
```json
{
  "clinic_patient_id": "uuid",           // ✅ REQUIRED - Patient ID in clinic
  "doctor_id": "uuid",                  // ✅ REQUIRED - Doctor ID
  "clinic_id": "uuid",                  // ✅ REQUIRED - Clinic ID
  "department_id": "uuid",              // ⚪ OPTIONAL - Department ID (can be null)
  "individual_slot_id": "uuid",         // ✅ REQUIRED - Time slot ID
  "appointment_date": "2025-01-15",     // ✅ REQUIRED - Date (YYYY-MM-DD)
  "appointment_time": "2025-01-15 10:30:00", // ✅ REQUIRED - Time (YYYY-MM-DD HH:MM:SS)
  "consultation_type": "clinic_visit",   // ✅ REQUIRED - Type of consultation
  "reason": "Regular checkup",          // ⚪ OPTIONAL - Appointment reason
  "notes": "Patient notes",             // ⚪ OPTIONAL - Additional notes
  "payment_method": "pay_now",           // ⚪ CONDITIONAL - Payment method
  "payment_type": "cash"                // ⚪ CONDITIONAL - Payment type
}
```

### **Consultation Types**
- `clinic_visit` - Regular in-person appointment
- `video_consultation` - Video call appointment  
- `follow-up-via-clinic` - Follow-up in-person appointment
- `follow-up-via-video` - Follow-up video appointment

### **Payment Methods**
- `pay_now` - Payment required immediately
- `pay_later` - Payment can be done later
- `way_off` - Payment waived/off

### **Payment Types** (when `payment_method = "pay_now"`)
- `cash` - Cash payment
- `card` - Card payment
- `upi` - UPI payment

---

## 🔄 **Complete Process Flow**

### **Step 1: Input Validation & Auto-Detection**
```go
// ✅ Auto-detect follow-up based on consultation_type
if input.ConsultationType == "follow-up-via-clinic" || input.ConsultationType == "follow-up-via-video" {
    input.IsFollowUp = true
}
```

### **Step 2: Patient Validation**
```sql
-- Check if patient exists and belongs to clinic
SELECT clinic_id FROM clinic_patients 
WHERE id = $1 AND is_active = true
```
**Validates:**
- ✅ Patient exists in database
- ✅ Patient is active
- ✅ Patient belongs to the specified clinic

### **Step 3: Follow-Up Eligibility Check** (Only for Follow-ups)
```go
// ✅ Use follow-up manager for validation
isFree, isEligible, message, err := followUpManager.CheckFollowUpEligibility(
    input.ClinicPatientID, 
    input.ClinicID, 
    input.DoctorID, 
    input.DepartmentID,
)
```

**Checks:**
- ✅ Patient has previous appointment with same doctor+department
- ✅ Follow-up is within 5-day window
- ✅ Free follow-up hasn't been used already
- ✅ Follow-up record exists in `follow_ups` table

### **Step 4: Fraud Prevention** (For Free Follow-ups)
```go
// ✅ Verify free follow-up is still available
activeFollowUp, err := followUpManager.GetActiveFollowUp(...)
if activeFollowUp == nil || !activeFollowUp.IsFree {
    // Prevent fraud - free follow-up no longer available
}
```

### **Step 5: Payment Validation**
```go
// ✅ Payment logic based on appointment type
if !input.IsFollowUp || (input.IsFollowUp && !isFreeFollowUp) {
    // Regular appointments OR Paid follow-ups require payment_method
    if input.PaymentMethod == nil {
        // Return error - payment required
    }
}
```

**Payment Rules:**
- 🟢 **Free Follow-ups:** No payment required
- 🟠 **Paid Follow-ups:** Payment required
- 🔵 **Regular Appointments:** Payment required

### **Step 6: Slot Availability Check**
```sql
-- Check slot capacity and availability
SELECT clinic_id, slot_start, slot_end, is_booked, status, max_patients, available_count
FROM doctor_individual_slots
WHERE id = $1
```

**Validates:**
- ✅ Slot exists
- ✅ Slot belongs to correct clinic
- ✅ Slot has available capacity (`available_count > 0`)
- ✅ Slot status is `available`

### **Step 7: Date/Time Parsing**
```go
appointmentDate, err := time.Parse("2006-01-02", input.AppointmentDate)
appointmentTime, err := time.Parse("2006-01-02 15:04:05", input.AppointmentTime)
```

### **Step 8: Doctor & Fee Calculation**
```sql
-- Get doctor info and fees
SELECT d.id, d.doctor_code, u.first_name, u.last_name,
       COALESCE(cdl.consultation_fee_offline, d.consultation_fee) as consultation_fee,
       COALESCE(cdl.follow_up_fee, d.follow_up_fee) as follow_up_fee
FROM doctors d
JOIN users u ON u.id = d.user_id
LEFT JOIN clinic_doctor_links cdl ON cdl.doctor_id = d.id AND cdl.clinic_id = $1
WHERE d.id = $2 AND d.is_active = true
```

**Fee Calculation Logic:**
```go
feeAmount := 0.0
if (input.ConsultationType == "follow-up-via-clinic" || input.ConsultationType == "follow-up-via-video") && followUpFee != nil {
    feeAmount = *followUpFee  // Use follow-up fee
} else if consultationFee != nil {
    feeAmount = *consultationFee  // Use regular consultation fee
}
```

### **Step 9: Generate Booking Number & Token**
```go
bookingNumber, err := utils.GenerateBookingNumber(doctor.DoctorCode, appointmentTime)
tokenNumber, err := utils.GenerateTokenNumber(input.DoctorID, input.ClinicID, appointmentDate)
```

### **Step 10: Payment Status Determination**
```go
// ✅ FREE follow-up: No payment required
if input.IsFollowUp && isFreeFollowUp {
    paymentStatus = "waived"
    paymentMode = nil
    feeAmount = 0.0 // No fee for free follow-ups
} else if input.PaymentMethod != nil {
    // ✅ Regular appointments OR Paid follow-ups
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
}
```

### **Step 11: Create Appointment Record**
```sql
INSERT INTO appointments (
    clinic_patient_id, clinic_id, doctor_id, department_id, booking_number, token_number,
    appointment_date, appointment_time, duration_minutes, consultation_type,
    reason, notes, fee_amount, payment_mode, payment_status, status, individual_slot_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING id, clinic_patient_id, clinic_id, doctor_id, booking_number, token_number,
          appointment_date, appointment_time, duration_minutes, consultation_type,
          reason, notes, status, fee_amount, payment_status, payment_mode, created_at
```

### **Step 12: Update Slot Availability**
```sql
-- Decrease slot capacity and mark as booked if full
UPDATE doctor_individual_slots
SET available_count = available_count - 1,
    is_booked = CASE WHEN available_count - 1 <= 0 THEN true ELSE is_booked END,
    status = CASE WHEN available_count - 1 <= 0 THEN 'booked' ELSE status END,
    booked_appointment_id = CASE WHEN available_count - 1 <= 0 THEN $1 ELSE booked_appointment_id END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $2
AND available_count > 0
AND status = 'available'
```

### **Step 13: Follow-Up Management**

#### **For Regular Appointments:**
```go
// ✅ Create new follow-up eligibility record
if input.ConsultationType == "clinic_visit" || input.ConsultationType == "video_consultation" {
    err = followUpManager.CreateFollowUp(
        input.ClinicPatientID,
        input.ClinicID,
        input.DoctorID,
        input.DepartmentID,
        appointment.ID,
        appointmentDate,
    )
}
```

#### **For Free Follow-ups:**
```go
// ✅ Mark free follow-up as used (Fraud Prevention)
if input.IsFollowUp && isFreeFollowUp {
    err = followUpManager.MarkFollowUpAsUsed(
        input.ClinicPatientID,
        input.ClinicID,
        input.DoctorID,
        input.DepartmentID,
        appointment.ID,
    )
    
    // ⚠️ CRITICAL: If marking fails, rollback everything
    if err != nil {
        // Delete appointment
        // Re-enable slot
        // Return error
    }
}
```

---

## 📤 **Response Structure**

### **Success Response (201 Created)**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "uuid",
    "clinic_patient_id": "uuid",
    "clinic_id": "uuid",
    "doctor_id": "uuid",
    "department_id": "uuid",
    "booking_number": "DOC-20250115-0001",
    "token_number": 1,
    "appointment_date": "2025-01-15",
    "appointment_time": "2025-01-15T10:30:00Z",
    "duration_minutes": 5,
    "consultation_type": "clinic_visit",
    "reason": "Regular checkup",
    "notes": "Patient notes",
    "status": "confirmed",
    "fee_amount": 500.0,
    "payment_status": "paid",
    "payment_mode": "cash",
    "created_at": "2025-01-15T10:30:00Z"
  },
  
  // ✅ Follow-up Information (for Follow-ups)
  "is_free_followup": true,
  "followup_type": "free",
  "followup_message": "This is a FREE follow-up (renewed after regular appointment)",
  
  // ✅ OR Regular Appointment Information
  "is_regular_appointment": true,
  "followup_granted": true,
  "followup_message": "Free follow-up eligibility granted (valid for 5 days)",
  "followup_valid_until": "2025-01-20"
}
```

### **Error Responses**

#### **400 Bad Request - Validation Errors**
```json
{
  "error": "Invalid input",
  "details": "Key: 'SimpleAppointmentInput.ClinicPatientID' Error:Field validation for 'ClinicPatientID' failed on the 'required' tag"
}
```

#### **400 Bad Request - Patient Not Found**
```json
{
  "error": "Patient not found"
}
```

#### **400 Bad Request - Follow-up Not Eligible**
```json
{
  "error": "Not eligible for follow-up",
  "message": "No previous appointment with this doctor and department"
}
```

#### **400 Bad Request - Payment Required**
```json
{
  "error": "Payment method required",
  "message": "This follow-up requires payment (free follow-up period expired or already used)"
}
```

#### **409 Conflict - Slot Not Available**
```json
{
  "error": "Slot not available",
  "message": "This slot is fully booked. Please select another slot.",
  "details": {
    "max_patients": 5,
    "available_count": 0,
    "booked_count": 5
  }
}
```

#### **409 Conflict - Free Follow-up Already Used**
```json
{
  "error": "Free follow-up already used",
  "message": "This free follow-up has already been used. Please book a paid follow-up or book a new regular appointment."
}
```

---

## 🔍 **Follow-Up Logic Deep Dive**

### **Follow-Up Creation Process**
1. **Regular Appointment Booked** → `CreateFollowUp()` called
2. **Check Existing Follow-ups** → `RenewExistingFollowUps()` marks old ones as 'renewed'
3. **Create New Follow-up** → Insert into `follow_ups` table with:
   - `status = 'active'`
   - `is_free = true`
   - `valid_from = appointment_date`
   - `valid_until = appointment_date + 5 days`

### **Follow-Up Usage Process**
1. **Follow-up Appointment Booked** → `CheckFollowUpEligibility()` called
2. **Verify Active Follow-up** → Check `follow_ups` table
3. **If Free Available** → `MarkFollowUpAsUsed()` called
4. **Update Follow-up Record** → Set `status = 'used'`

### **Follow-Up Renewal Process**
1. **New Regular Appointment** → `RenewExistingFollowUps()` called
2. **Mark Old Follow-ups** → Set `status = 'renewed'`
3. **Create New Follow-up** → Fresh 5-day eligibility window

---

## 🛡️ **Security & Fraud Prevention**

### **Race Condition Prevention**
- ✅ **Slot Booking:** Uses `WHERE available_count > 0` to prevent double booking
- ✅ **Free Follow-up:** Uses `SELECT FOR UPDATE` to prevent concurrent usage
- ✅ **Transaction Rollback:** If follow-up marking fails, entire appointment is rolled back

### **Data Validation**
- ✅ **UUID Validation:** All IDs validated as proper UUIDs
- ✅ **Date Format:** Strict date/time format validation
- ✅ **Enum Validation:** Consultation types and payment methods validated
- ✅ **Clinic Isolation:** Patients can only book in their own clinic

### **Business Logic Validation**
- ✅ **Follow-up Eligibility:** Only valid follow-ups can be booked
- ✅ **Payment Requirements:** Proper payment validation based on appointment type
- ✅ **Slot Availability:** Real-time slot capacity checking

---

## 📊 **Database Tables Affected**

### **Primary Tables**
1. **`appointments`** - Main appointment record
2. **`doctor_individual_slots`** - Slot availability updates
3. **`follow_ups`** - Follow-up eligibility management

### **Lookup Tables**
1. **`clinic_patients`** - Patient validation
2. **`doctors`** - Doctor info and fees
3. **`users`** - Doctor name lookup
4. **`departments`** - Department info
5. **`clinic_doctor_links`** - Clinic-specific doctor fees

---

## 🎯 **Use Cases & Examples**

### **Example 1: Regular Appointment**
```json
{
  "clinic_patient_id": "123e4567-e89b-12d3-a456-426614174000",
  "doctor_id": "123e4567-e89b-12d3-a456-426614174001",
  "clinic_id": "123e4567-e89b-12d3-a456-426614174002",
  "department_id": "123e4567-e89b-12d3-a456-426614174003",
  "individual_slot_id": "123e4567-e89b-12d3-a456-426614174004",
  "appointment_date": "2025-01-15",
  "appointment_time": "2025-01-15 10:30:00",
  "consultation_type": "clinic_visit",
  "reason": "Regular checkup",
  "payment_method": "pay_now",
  "payment_type": "cash"
}
```

**Result:** Creates appointment + grants 5-day free follow-up eligibility

### **Example 2: Free Follow-up**
```json
{
  "clinic_patient_id": "123e4567-e89b-12d3-a456-426614174000",
  "doctor_id": "123e4567-e89b-12d3-a456-426614174001",
  "clinic_id": "123e4567-e89b-12d3-a456-426614174002",
  "department_id": "123e4567-e89b-12d3-a456-426614174003",
  "individual_slot_id": "123e4567-e89b-12d3-a456-426614174005",
  "appointment_date": "2025-01-17",
  "appointment_time": "2025-01-17 14:00:00",
  "consultation_type": "follow-up-via-clinic"
}
```

**Result:** Creates free follow-up appointment (no payment required)

### **Example 3: Paid Follow-up**
```json
{
  "clinic_patient_id": "123e4567-e89b-12d3-a456-426614174000",
  "doctor_id": "123e4567-e89b-12d3-a456-426614174001",
  "clinic_id": "123e4567-e89b-12d3-a456-426614174002",
  "department_id": "123e4567-e89b-12d3-a456-426614174003",
  "individual_slot_id": "123e4567-e89b-12d3-a456-426614174006",
  "appointment_date": "2025-01-25",
  "appointment_time": "2025-01-25 16:00:00",
  "consultation_type": "follow-up-via-clinic",
  "payment_method": "pay_now",
  "payment_type": "card"
}
```

**Result:** Creates paid follow-up appointment (free period expired)

---

## 🔧 **Testing Scenarios**

### **Happy Path Tests**
1. ✅ Regular appointment creation
2. ✅ Free follow-up within 5 days
3. ✅ Paid follow-up after 5 days
4. ✅ Follow-up renewal after new regular appointment

### **Error Path Tests**
1. ❌ Invalid patient ID
2. ❌ Patient from different clinic
3. ❌ Follow-up without previous appointment
4. ❌ Free follow-up already used
5. ❌ Slot fully booked
6. ❌ Invalid payment method for paid appointment

### **Edge Cases**
1. 🔄 Race condition: Multiple users booking same slot
2. 🔄 Concurrent free follow-up usage
3. 🔄 Follow-up renewal timing
4. 🔄 Department-specific follow-ups

---

## 📈 **Performance Considerations**

### **Database Optimizations**
- ✅ **Indexed Queries:** All lookups use indexed fields
- ✅ **Atomic Updates:** Slot updates use single SQL statement
- ✅ **Transaction Management:** Critical operations wrapped in transactions

### **Caching Opportunities**
- 🔄 **Doctor Fees:** Could cache doctor fee information
- 🔄 **Slot Availability:** Could cache slot status
- 🔄 **Follow-up Status:** Could cache follow-up eligibility

### **Monitoring Points**
- 📊 **Slot Booking Conflicts:** Track race condition frequency
- 📊 **Follow-up Usage:** Monitor free vs paid follow-up ratios
- 📊 **Payment Failures:** Track payment validation errors

---

This documentation provides a complete understanding of the `CreateSimpleAppointment` API, including all validation steps, business logic, error handling, and follow-up management. The system is designed to be robust, secure, and user-friendly while maintaining data integrity and preventing fraud.
