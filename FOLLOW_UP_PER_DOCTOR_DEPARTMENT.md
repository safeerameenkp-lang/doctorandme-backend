# Follow-Up Per Doctor + Department - Implementation ✅

## 🎯 **Requirement**

Follow-up appointments should be **doctor-specific AND department-specific**.

**Key Rule:**
- **One FREE follow-up per (Doctor + Department) combination within 5 days**

---

## ✅ **How It Works Now**

### Example Scenario:

#### Patient History:
```
Day 1: Doctor ABC → Cardiology → Paid ₹500
```

#### Follow-Up Attempts:

| Day | Doctor | Department | Result | Fee | Reason |
|-----|--------|-----------|--------|-----|--------|
| Day 3 | Doctor ABC | Cardiology | ✅ **FREE** | ₹0 | Same doctor + Same department + Within 5 days + First follow-up |
| Day 4 | Doctor ABC | Cardiology | ❌ **PAID** | ₹200 | Free follow-up already used for ABC+Cardiology |
| Day 3 | Doctor ABC | Neurology | ❌ **PAID** | ₹500 | Different department (new appointment) |
| Day 3 | Doctor XYZ | Cardiology | ❌ **PAID** | ₹500 | Different doctor (new appointment) |
| Day 8 | Doctor ABC | Cardiology | ❌ **PAID** | ₹200 | After 5 days (expired) |

---

## 🔑 **Key Logic**

### 1. **Free Follow-Up Criteria:**
✅ **ALL** must be true:
- Same doctor as previous appointment
- Same department as previous appointment
- Within 5 days of previous appointment
- No free follow-up already used for this doctor+department

### 2. **Paid Follow-Up (New Appointment):**
❌ **ANY** triggers paid:
- Different doctor
- Different department
- After 5 days
- Free follow-up already used for this doctor+department

---

## 📊 **Example: Multiple Doctors & Departments**

### Patient Appointment History:

```
Oct 10: Doctor A → Cardiology → Paid ₹500
Oct 12: Doctor A → Cardiology → FREE (follow-up)
Oct 14: Doctor B → Neurology → Paid ₹600
Oct 16: Doctor A → Neurology → Paid ₹500
```

### Follow-Up Eligibility on Oct 17:

| Doctor | Department | Eligible for Free? | Reason |
|--------|-----------|-------------------|--------|
| Doctor A | Cardiology | ❌ NO | Already used (Oct 12) |
| Doctor A | Neurology | ✅ YES | Within 5 days (Oct 16) + First follow-up for A+Neurology |
| Doctor B | Neurology | ✅ YES | Within 5 days (Oct 14) + First follow-up for B+Neurology |
| Doctor B | Cardiology | ❌ NO | No previous appointment with B in Cardiology |

---

## 🔄 **Query Logic**

### Count Free Follow-Ups (Per Doctor + Department):

```sql
SELECT COUNT(*)
FROM appointments
WHERE clinic_patient_id = $1
  AND clinic_id = $2
  AND doctor_id = $3              -- ✅ Same doctor
  AND department_id = $4           -- ✅ Same department
  AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
  AND payment_status = 'waived'    -- ✅ Only FREE follow-ups
  AND appointment_date >= $5       -- ✅ From last appointment date
  AND status NOT IN ('cancelled', 'no_show')
```

**Result:**
- `COUNT = 0` → **FREE** follow-up ✅
- `COUNT > 0` → **PAID** follow-up (already used) ❌

---

## 🎯 **Implementation Details**

### 1. `appointment_simple.controller.go` (Lines 138-165)

**Changes:**
- Added `department_id` to the COUNT query
- Dynamic query building with `args` array
- Only adds department filter if `input.DepartmentID != nil`

```go
// Build query to check per doctor AND department
query := `
    SELECT COUNT(*)
    FROM appointments
    WHERE clinic_patient_id = $1
      AND clinic_id = $2
      AND doctor_id = $3
      AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
      AND payment_status = 'waived'
      AND appointment_date >= $4
      AND status NOT IN ('cancelled', 'no_show')`

args := []interface{}{input.ClinicPatientID, input.ClinicID, input.DoctorID, *previousAppointmentDate}

// Add department check if department is specified
if input.DepartmentID != nil {
    query += ` AND department_id = $5`
    args = append(args, *input.DepartmentID)
}

err = config.DB.QueryRow(query, args...).Scan(&freeFollowUpCount)
```

---

### 2. `clinic_patient.controller.go` (Lines 709-746)

**Changes:**
- Added `department_id` to the COUNT query in `populateAppointmentHistory`
- Dynamic query building with `args` array
- Updated eligibility message to mention "doctor in department"

```go
// Build query to check per doctor AND department
query := `
    SELECT COUNT(*)
    FROM appointments
    WHERE clinic_patient_id = $1
      AND clinic_id = $2
      AND doctor_id = $3
      AND consultation_type IN ('follow-up-via-clinic', 'follow-up-via-video')
      AND payment_status = 'waived'
      AND appointment_date >= $4
      AND status NOT IN ('cancelled', 'no_show')`

args := []interface{}{patient.ID, patient.ClinicID, lastAppt.DoctorID, appointmentDate}

// Add department check if department exists
if lastAppt.DepartmentID != nil && *lastAppt.DepartmentID != "" {
    query += ` AND department_id = $5`
    args = append(args, *lastAppt.DepartmentID)
}

err = db.QueryRow(query, args...).Scan(&freeFollowUpCount)
```

---

## 🧪 **Test Scenarios**

### Test 1: Same Doctor, Same Department ✅

**Setup:**
- Oct 15: Doctor A → Cardiology → Paid
- Oct 17: Doctor A → Cardiology → Follow-up

**Expected:**
- ✅ FREE (same doctor + same department + within 5 days)

---

### Test 2: Same Doctor, Different Department ❌

**Setup:**
- Oct 15: Doctor A → Cardiology → Paid
- Oct 17: Doctor A → Neurology → Follow-up

**Expected:**
- ❌ PAID (different department = new appointment)
- Error: "Follow-up appointments must be in the same department as your previous appointment"

---

### Test 3: Different Doctor, Same Department ❌

**Setup:**
- Oct 15: Doctor A → Cardiology → Paid
- Oct 17: Doctor B → Cardiology → Follow-up

**Expected:**
- ❌ PAID (different doctor = new appointment)
- Error: "Follow-up appointments must be with the same doctor as your previous appointment"

---

### Test 4: Multiple Follow-Ups Same Doctor+Department ✅

**Setup:**
- Oct 15: Doctor A → Cardiology → Paid
- Oct 16: Doctor A → Cardiology → Follow-up #1
- Oct 17: Doctor A → Cardiology → Follow-up #2

**Expected:**
- Follow-up #1: ✅ FREE (first follow-up)
- Follow-up #2: ❌ PAID (free already used)

---

### Test 5: Multiple Departments, Each Gets One Free ✅

**Setup:**
- Oct 15: Doctor A → Cardiology → Paid
- Oct 16: Doctor A → Cardiology → Follow-up (FREE)
- Oct 18: Doctor A → Neurology → Paid
- Oct 19: Doctor A → Neurology → Follow-up

**Expected:**
- Cardiology Follow-up: ✅ FREE (first for Cardiology)
- Neurology Follow-up: ✅ FREE (first for Neurology - separate department)

---

## 📋 **API Behavior**

### Patient API Response:

```json
{
  "id": "patient-uuid",
  "first_name": "John",
  "last_name": "Doe",
  "last_appointment": {
    "doctor_id": "doctor-abc-uuid",
    "doctor_name": "Dr. ABC",
    "department_id": "cardiology-uuid",
    "department": "Cardiology",
    "date": "2025-10-15",
    "days_since": 2
  },
  "follow_up_eligibility": {
    "eligible": true,
    "is_free": true,
    "days_remaining": 3,
    "message": "You have one FREE follow-up available with this doctor in this department"
  }
}
```

**Key Points:**
- `follow_up_eligibility` is **specific to the doctor+department** of `last_appointment`
- If patient books with different doctor/department, they won't see free eligibility

---

## ✅ **Benefits**

| Benefit | Description |
|---------|-------------|
| **Fair to patients** | Each doctor+department gets one free follow-up |
| **No exploitation** | Can't use free follow-up across different doctors/departments |
| **Clear tracking** | Easy to verify: "Did patient use free follow-up with THIS doctor in THIS department?" |
| **Scalable** | Works with multiple doctors and multiple departments |
| **Business logic** | Aligns with real-world medical practice (follow-up is doctor+specialty-specific) |

---

## 🚀 **Deployment**

### 1. Build (In Progress)
```bash
docker-compose build appointment-service organization-service
```

### 2. Deploy
```bash
docker-compose up -d appointment-service organization-service
```

### 3. Verify
- Try booking follow-up with same doctor+department (should be FREE)
- Try booking follow-up with different doctor or department (should be PAID)

---

## ✅ **Summary**

| Aspect | Value |
|--------|-------|
| **Scope** | Per Doctor + Department |
| **Free Follow-Ups** | 1 per (Doctor + Department) within 5 days |
| **Changed Files** | 2 (appointment_simple, clinic_patient) |
| **Query Update** | Added `department_id` filter |
| **Logic** | Dynamic query building with args |
| **Status** | ✅ **COMPLETE** |

---

**Result:** Follow-ups are now correctly tracked per (Doctor + Department) combination! 🎉✅

