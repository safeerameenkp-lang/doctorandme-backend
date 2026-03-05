# 🔄 Follow-Up Renewal Logic Test & Verification

## 🎯 **Test Scenario: Expired Follow-Up Renewal**

**Scenario:** Patient has a previous follow-up that expired, then books a new regular appointment with the same doctor and department. The system should automatically renew the follow-up eligibility.

---

## 📋 **Test Case: Expired Follow-Up → New Regular Appointment → Free Follow-Up Renewed**

### **Step 1: Initial State (Patient has expired follow-up)**
```sql
-- Patient had a regular appointment 10 days ago
INSERT INTO appointments (
    id, clinic_patient_id, clinic_id, doctor_id, department_id,
    appointment_date, appointment_time, consultation_type, status
) VALUES (
    'appt-001', 'patient-123', 'clinic-456', 'doctor-789', 'dept-101',
    '2025-01-05', '2025-01-05 10:00:00', 'clinic_visit', 'completed'
);

-- Follow-up was created but expired (5 days ago)
INSERT INTO follow_ups (
    id, clinic_patient_id, clinic_id, doctor_id, department_id,
    source_appointment_id, status, is_free, valid_from, valid_until
) VALUES (
    'followup-001', 'patient-123', 'clinic-456', 'doctor-789', 'dept-101',
    'appt-001', 'expired', true, '2025-01-05', '2025-01-10'
);
```

### **Step 2: Patient Books New Regular Appointment**
```json
POST /appointments/simple
{
  "clinic_patient_id": "patient-123",
  "doctor_id": "doctor-789",
  "clinic_id": "clinic-456",
  "department_id": "dept-101",
  "individual_slot_id": "slot-001",
  "appointment_date": "2025-01-15",
  "appointment_time": "2025-01-15 14:00:00",
  "consultation_type": "clinic_visit",
  "reason": "Regular checkup",
  "payment_method": "pay_now",
  "payment_type": "cash"
}
```

### **Step 3: Expected System Behavior**

#### **3.1: Follow-Up Manager CreateFollowUp() Called**
```go
// In CreateFollowUp() function:
err := fm.RenewExistingFollowUps(clinicPatientID, clinicID, doctorID, departmentID, appointmentID)
```

#### **3.2: RenewExistingFollowUps() Updates Expired Follow-Up**
```sql
-- This query should find and update the expired follow-up
UPDATE follow_ups
SET status = 'renewed',
    renewed_at = CURRENT_TIMESTAMP,
    renewed_by_appointment_id = 'appt-002',  -- New appointment ID
    updated_at = CURRENT_TIMESTAMP
WHERE clinic_patient_id = 'patient-123'
  AND clinic_id = 'clinic-456'
  AND doctor_id = 'doctor-789'
  AND department_id = 'dept-101'
  AND status IN ('active', 'expired')  -- ✅ This includes 'expired'
```

#### **3.3: New Follow-Up Record Created**
```sql
-- New active follow-up created
INSERT INTO follow_ups (
    clinic_patient_id, clinic_id, doctor_id, department_id,
    source_appointment_id, status, is_free, valid_from, valid_until
) VALUES (
    'patient-123', 'clinic-456', 'doctor-789', 'dept-101',
    'appt-002', 'active', true, '2025-01-15', '2025-01-20'
);
```

### **Step 4: Verification Queries**

#### **4.1: Check Follow-Up Status After Renewal**
```sql
-- Should show the old follow-up as 'renewed' and new one as 'active'
SELECT 
    id,
    source_appointment_id,
    status,
    is_free,
    valid_from,
    valid_until,
    renewed_at,
    renewed_by_appointment_id
FROM follow_ups
WHERE clinic_patient_id = 'patient-123'
  AND doctor_id = 'doctor-789'
  AND department_id = 'dept-101'
ORDER BY created_at DESC;
```

**Expected Result:**
```
id            | source_appointment_id | status  | is_free | valid_from | valid_until | renewed_at           | renewed_by_appointment_id
followup-002  | appt-002             | active  | true    | 2025-01-15 | 2025-01-20  | NULL                 | NULL
followup-001  | appt-001             | renewed | true    | 2025-01-05 | 2025-01-10  | 2025-01-15 14:00:00  | appt-002
```

#### **4.2: Test Follow-Up Eligibility Check**
```go
// This should now return: (true, true, "Free follow-up available (5 days remaining)")
isFree, isEligible, message, err := followUpManager.CheckFollowUpEligibility(
    "patient-123", "clinic-456", "doctor-789", "dept-101"
)
```

### **Step 5: Test Follow-Up Booking**

#### **5.1: Book Free Follow-Up**
```json
POST /appointments/simple
{
  "clinic_patient_id": "patient-123",
  "doctor_id": "doctor-789",
  "clinic_id": "clinic-456",
  "department_id": "dept-101",
  "individual_slot_id": "slot-002",
  "appointment_date": "2025-01-17",
  "appointment_time": "2025-01-17 16:00:00",
  "consultation_type": "follow-up-via-clinic"
  // No payment_method required for free follow-up
}
```

#### **5.2: Expected Response**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "consultation_type": "follow-up-via-clinic",
    "fee_amount": 0.0,
    "payment_status": "waived",
    "payment_mode": null
  },
  "is_free_followup": true,
  "followup_type": "free",
  "followup_message": "This is a FREE follow-up (renewed after regular appointment)"
}
```

---

## 🔍 **Code Analysis: Is Renewal Logic Working?**

### **✅ What Should Happen:**

1. **Regular Appointment Created** → `CreateFollowUp()` called
2. **RenewExistingFollowUps()** → Finds expired follow-up and marks as 'renewed'
3. **New Follow-Up Created** → Fresh 5-day eligibility window
4. **Follow-Up Booking** → Should be free within 5 days

### **🔍 Current Implementation Check:**

#### **RenewExistingFollowUps() Logic:**
```go
// ✅ CORRECT: Updates both 'active' and 'expired' follow-ups
query := `
    UPDATE follow_ups
    SET status = 'renewed',
        renewed_at = CURRENT_TIMESTAMP,
        renewed_by_appointment_id = $1,
        updated_at = CURRENT_TIMESTAMP
    WHERE clinic_patient_id = $2
      AND clinic_id = $3
      AND doctor_id = $4
      AND status IN ('active', 'expired')  // ✅ Includes 'expired'
`
```

#### **Department Handling:**
```go
// ✅ CORRECT: Handles both NULL and empty string departments
if departmentID != nil && *departmentID != "" {
    query += ` AND department_id = $5`
    args = append(args, *departmentID)
} else {
    query += ` AND (department_id IS NULL OR department_id = '')`
}
```

#### **CreateFollowUp() Flow:**
```go
// ✅ CORRECT: Calls renewal first, then creates new follow-up
err := fm.RenewExistingFollowUps(clinicPatientID, clinicID, doctorID, departmentID, appointmentID)
// ... then creates new follow-up
```

---

## 🧪 **Manual Test Script**

### **Test 1: Create Test Data**
```sql
-- Clean up any existing test data
DELETE FROM follow_ups WHERE clinic_patient_id = 'test-patient-123';
DELETE FROM appointments WHERE clinic_patient_id = 'test-patient-123';

-- Create test patient
INSERT INTO clinic_patients (id, clinic_id, first_name, last_name, phone) 
VALUES ('test-patient-123', 'test-clinic-456', 'Test', 'Patient', '1234567890');

-- Create test doctor
INSERT INTO doctors (id, clinic_id, user_id, doctor_code) 
VALUES ('test-doctor-789', 'test-clinic-456', 'test-user-789', 'DOC001');

-- Create test department
INSERT INTO departments (id, name) 
VALUES ('test-dept-101', 'Test Department');

-- Create test slot
INSERT INTO doctor_individual_slots (id, clinic_id, doctor_id, slot_date, slot_start, slot_end, max_patients, available_count, status)
VALUES ('test-slot-001', 'test-clinic-456', 'test-doctor-789', '2025-01-15', '10:00:00', '10:30:00', 1, 1, 'available');
```

### **Test 2: Create Expired Follow-Up**
```sql
-- Create old appointment (10 days ago)
INSERT INTO appointments (
    id, clinic_patient_id, clinic_id, doctor_id, department_id,
    appointment_date, appointment_time, consultation_type, status, fee_amount, payment_status
) VALUES (
    'test-appt-001', 'test-patient-123', 'test-clinic-456', 'test-doctor-789', 'test-dept-101',
    '2025-01-05', '2025-01-05 10:00:00', 'clinic_visit', 'completed', 500.0, 'paid'
);

-- Create expired follow-up
INSERT INTO follow_ups (
    id, clinic_patient_id, clinic_id, doctor_id, department_id,
    source_appointment_id, status, is_free, valid_from, valid_until,
    created_at, updated_at
) VALUES (
    'test-followup-001', 'test-patient-123', 'test-clinic-456', 'test-doctor-789', 'test-dept-101',
    'test-appt-001', 'expired', true, '2025-01-05', '2025-01-10',
    '2025-01-05 10:00:00', '2025-01-05 10:00:00'
);
```

### **Test 3: Book New Regular Appointment**
```bash
curl -X POST http://localhost:8080/appointments/simple \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "clinic_patient_id": "test-patient-123",
    "doctor_id": "test-doctor-789",
    "clinic_id": "test-clinic-456",
    "department_id": "test-dept-101",
    "individual_slot_id": "test-slot-001",
    "appointment_date": "2025-01-15",
    "appointment_time": "2025-01-15 10:00:00",
    "consultation_type": "clinic_visit",
    "reason": "Test renewal",
    "payment_method": "pay_now",
    "payment_type": "cash"
  }'
```

### **Test 4: Verify Follow-Up Status**
```sql
-- Check follow-up records
SELECT 
    id,
    source_appointment_id,
    status,
    is_free,
    valid_from,
    valid_until,
    renewed_at,
    renewed_by_appointment_id
FROM follow_ups
WHERE clinic_patient_id = 'test-patient-123'
ORDER BY created_at DESC;
```

### **Test 5: Test Follow-Up Eligibility**
```bash
curl -X GET "http://localhost:8080/api/organizations/clinic-specific-patients/test-patient-123?doctor_id=test-doctor-789&department_id=test-dept-101" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Expected Response:**
```json
{
  "patient": {
    "follow_up_eligibility": {
      "eligible": true,
      "is_free": true,
      "status_label": "free",
      "color_code": "green",
      "message": "Free follow-up available (5 days remaining)"
    }
  }
}
```

---

## ✅ **Conclusion: Renewal Logic Analysis**

### **✅ What's Working Correctly:**

1. **RenewExistingFollowUps()** properly updates both 'active' and 'expired' follow-ups
2. **Department handling** correctly manages NULL and empty string departments
3. **CreateFollowUp()** calls renewal before creating new follow-up
4. **Follow-up eligibility** checks the follow_ups table correctly
5. **Fraud prevention** uses SELECT FOR UPDATE to prevent race conditions

### **🎯 Expected Behavior:**

When a patient books a new regular appointment with the same doctor and department after their follow-up expired:

1. ✅ **Old follow-up** → Status changed to 'renewed'
2. ✅ **New follow-up** → Created with 'active' status
3. ✅ **Fresh eligibility** → 5-day free follow-up window
4. ✅ **Follow-up booking** → Should be free within 5 days

### **🔧 If Issues Found:**

If the renewal isn't working, check:

1. **Database constraints** - Ensure follow_ups table allows status updates
2. **Transaction isolation** - Check if concurrent requests interfere
3. **Department matching** - Verify NULL vs empty string handling
4. **Date comparisons** - Ensure timezone handling is correct

The current implementation looks correct and should handle expired follow-up renewal properly. The key is that `RenewExistingFollowUps()` includes both 'active' and 'expired' statuses in its WHERE clause, which means it will find and renew expired follow-ups when a new regular appointment is booked.
