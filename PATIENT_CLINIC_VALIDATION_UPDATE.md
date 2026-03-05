# Patient-Clinic Validation Update

## 🎯 Overview

Updated the appointment creation logic to **validate** that patients are registered at the clinic before booking appointments.

---

## 🔄 What Changed

### Before (Old Behavior) ❌
```
CreateAppointment:
  1. Get patient
  2. Create appointment
  3. No validation if patient belongs to clinic
```

**Problem:** Patients could book appointments at clinics they're not registered with.

---

### After (New Behavior) ✅
```
CreateAppointment:
  1. Get patient
  2. ✅ Validate patient is registered at this clinic
  3. If not registered → Error
  4. If registered → Create appointment
```

**Benefit:** Ensures proper patient-clinic relationships

---

## 📋 Two Different APIs

### API 1: Create Appointment (Existing Patient)

**Endpoint:** `POST /appointments`

**Behavior:** 
- ✅ **Validates** patient is already registered at the clinic
- ❌ **Does NOT** auto-register patient to clinic
- ✅ Returns error if patient not registered

**Request:**
```json
POST /api/appointments
{
  "patient_id": "patient-uuid-123",
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline"
}
```

**Success Response (201):**
```json
{
  "appointment": {
    "id": "appointment-uuid",
    "patient_id": "patient-uuid-123",
    ...
  }
}
```

**Error Response if Not Registered (400):**
```json
{
  "error": "Patient not registered at this clinic",
  "message": "This patient must be registered at the clinic before booking appointments"
}
```

---

### API 2: Create Patient + Appointment

**Endpoint:** `POST /appointments/patient-appointment`

**Behavior:**
- ✅ Creates new patient
- ✅ **Automatically registers** patient at the clinic
- ✅ Creates appointment

**Request:**
```json
POST /api/appointments/patient-appointment
{
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "email": "john.doe@example.com",
  "clinic_id": "clinic-uuid",
  "doctor_id": "doctor-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:35:00",
  "consultation_type": "new"
}
```

**What Happens:**
1. ✅ Creates user account
2. ✅ Creates patient record
3. ✅ **Assigns patient to clinic** (patient_clinics table)
4. ✅ Creates appointment

**Success Response (201):**
```json
{
  "user": {
    "id": "user-uuid-new",
    "username": "johndoe",
    ...
  },
  "patient": {
    "id": "patient-uuid-new",
    ...
  },
  "appointment": {
    "id": "appointment-uuid",
    ...
  }
}
```

---

## 🔍 Validation Logic

### CreateAppointment Validation Flow

```
1. Get patient_id (from input)
   ↓
2. Check patient exists and is active ✅
   ↓
3. Check doctor is linked to clinic ✅
   ↓
4. ✅ NEW: Check patient is registered at clinic
   ↓
   Query:
   SELECT EXISTS(
       SELECT 1 FROM patient_clinics
       WHERE patient_id = $1 AND clinic_id = $2
   )
   ↓
5. If NOT registered → Return 400 error
   ↓
6. If registered → Continue with appointment creation
```

---

## 📊 Database Query

### New Validation Query
```sql
SELECT EXISTS(
    SELECT 1 FROM patient_clinics 
    WHERE patient_id = 'patient-uuid' 
    AND clinic_id = 'clinic-uuid'
)
```

**Returns:**
- `true` - Patient is registered at this clinic ✅
- `false` - Patient not registered at this clinic ❌

---

## ❌ Error Scenarios

### Scenario 1: Patient Not Registered at Clinic

**Request:**
```json
POST /appointments
{
  "patient_id": "patient-from-clinic-a",
  "clinic_id": "clinic-b-uuid",  // Different clinic!
  ...
}
```

**Response (400 Bad Request):**
```json
{
  "error": "Patient not registered at this clinic",
  "message": "This patient must be registered at the clinic before booking appointments"
}
```

**Solution:** 
1. First register patient at Clinic B
2. Then create appointment

---

### Scenario 2: Patient Registered at Wrong Clinic

**Setup:**
- Patient registered at: Clinic A ✅
- Trying to book at: Clinic B ❌

**Request:**
```json
POST /appointments
{
  "patient_id": "patient-uuid",
  "clinic_id": "clinic-b-uuid",
  ...
}
```

**Response (400):**
```json
{
  "error": "Patient not registered at this clinic",
  "message": "This patient must be registered at the clinic before booking appointments"
}
```

---

## ✅ Correct Flow

### Step 1: Register Patient at Clinic (if not already)

This should be done through patient management APIs:
```json
POST /patients/{patient_id}/clinics
{
  "clinic_id": "clinic-uuid",
  "is_primary": true
}
```

Or patient is auto-registered when created via CreatePatientWithAppointment.

---

### Step 2: Book Appointment

**Request:**
```json
POST /appointments
{
  "patient_id": "patient-uuid",  // Already registered at clinic
  "clinic_id": "clinic-uuid",    // Same clinic
  "doctor_id": "doctor-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline"
}
```

**Validation:**
1. ✅ Patient exists
2. ✅ Doctor linked to clinic
3. ✅ **Patient registered at clinic** (NEW!)
4. ✅ Slot available
5. ✅ Create appointment

---

## 🎯 Benefits

### 1. Data Integrity ✅
- Ensures patients are properly registered before booking
- Prevents orphan appointments

### 2. Multi-Clinic Support ✅
- Patients can be registered at multiple clinics
- Each clinic has its own patient list
- Prevents cross-clinic booking errors

### 3. Better Error Messages ✅
```json
{
  "error": "Patient not registered at this clinic",
  "message": "This patient must be registered at the clinic before booking appointments"
}
```

Clear guidance on what's wrong and how to fix it.

### 4. Proper Workflow ✅
```
Register Patient → Book Appointment
(Not: Book → Auto-register)
```

---

## 📝 Summary

| API | Patient-Clinic Handling |
|-----|------------------------|
| **POST /appointments** | ✅ Validates patient is registered |
| **POST /appointments/patient-appointment** | ✅ Auto-registers new patient |

---

## 🔧 Technical Details

### Code Changes

**File:** `services/appointment-service/controllers/appointment.controller.go`

**Added:**
```go
// Validate patient is registered at this clinic
var patientLinked bool
err = config.DB.QueryRow(`
    SELECT EXISTS(
        SELECT 1 FROM patient_clinics 
        WHERE patient_id = $1 AND clinic_id = $2
    )
`, patientID, input.ClinicID).Scan(&patientLinked)

if !patientLinked {
    return error: "Patient not registered at this clinic"
}
```

**Location:** After doctor-clinic validation, before creating appointment

---

## ✅ Status

| Feature | Status |
|---------|--------|
| Patient-clinic validation | ✅ Added |
| Error message | ✅ Clear |
| CreateAppointment | ✅ Validates |
| CreatePatientWithAppointment | ✅ Auto-registers |
| Multi-clinic support | ✅ Working |

---

**Status:** ✅ Complete  
**Last Updated:** October 15, 2025  
**Version:** 1.0

Patients must now be registered at the clinic before booking appointments! ✅

