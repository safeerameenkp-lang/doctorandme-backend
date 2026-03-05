# Appointment Binding Update ✅

## 🔄 What Was Updated

Updated `SimpleAppointmentInput` binding validation to match the new slot type naming convention.

---

## 📝 Changes Made

### ❌ Old Binding (Incorrect):
```go
ConsultationType string `json:"consultation_type" binding:"required,oneof=offline online video in_person follow_up"`
```

### ✅ New Binding (Correct):
```go
ConsultationType string `json:"consultation_type" binding:"required,oneof=clinic_visit video_consultation in_person follow_up"`
```

---

## 📊 Complete Binding Reference

### SimpleAppointmentInput Struct:

```go
type SimpleAppointmentInput struct {
    // Required fields
    ClinicPatientID  string  `json:"clinic_patient_id" binding:"required,uuid"`
    DoctorID         string  `json:"doctor_id" binding:"required,uuid"`
    ClinicID         string  `json:"clinic_id" binding:"required,uuid"`
    IndividualSlotID string  `json:"individual_slot_id" binding:"required,uuid"`
    AppointmentDate  string  `json:"appointment_date" binding:"required"`
    AppointmentTime  string  `json:"appointment_time" binding:"required"`
    ConsultationType string  `json:"consultation_type" binding:"required,oneof=clinic_visit video_consultation in_person follow_up"`
    
    // Optional fields
    DepartmentID     *string `json:"department_id" binding:"omitempty,uuid"`
    IsFollowUp       bool    `json:"is_follow_up"`
    Reason           *string `json:"reason"`
    Notes            *string `json:"notes"`
    
    // Payment fields (optional for follow-ups)
    PaymentMethod    *string `json:"payment_method" binding:"omitempty,oneof=pay_now pay_later way_off"`
    PaymentType      *string `json:"payment_type" binding:"omitempty,oneof=cash card upi"`
}
```

---

## 📋 Field Validation Rules

| Field | Type | Required | Validation | Notes |
|-------|------|----------|------------|-------|
| `clinic_patient_id` | string | ✅ Yes | UUID | Must exist in clinic_patients |
| `doctor_id` | string | ✅ Yes | UUID | Must be active doctor |
| `clinic_id` | string | ✅ Yes | UUID | Must be active clinic |
| `department_id` | *string | ❌ No | UUID (if provided) | Optional, can be null |
| `individual_slot_id` | string | ✅ Yes | UUID | Must be available slot |
| `appointment_date` | string | ✅ Yes | YYYY-MM-DD | Future date |
| `appointment_time` | string | ✅ Yes | YYYY-MM-DD HH:MM:SS | Must match slot time |
| `consultation_type` | string | ✅ Yes | See below | Type of consultation |
| `is_follow_up` | bool | ❌ No | true/false | Triggers follow-up validation |
| `reason` | *string | ❌ No | Any text | Optional visit reason |
| `notes` | *string | ❌ No | Any text | Optional notes |
| `payment_method` | *string | ⚠️ Conditional | See below | Required for regular appointments |
| `payment_type` | *string | ⚠️ Conditional | See below | Required if payment_method = pay_now |

---

## 🔤 Valid Values

### consultation_type:
| Value | Description | UI Label |
|-------|-------------|----------|
| `clinic_visit` | In-person clinic visit | 🏥 Clinic Visit |
| `video_consultation` | Remote video call | 💻 Video Consultation |
| `in_person` | In-person visit | 👤 In Person |
| `follow_up` | Follow-up appointment | 🔄 Follow-Up |

### payment_method:
| Value | Description |
|-------|-------------|
| `pay_now` | Immediate payment |
| `pay_later` | Deferred payment |
| `way_off` | Payment waived |

### payment_type (when payment_method = "pay_now"):
| Value | Description |
|-------|-------------|
| `cash` | Cash payment |
| `card` | Card payment |
| `upi` | UPI payment |

---

## 📝 Request Examples

### 1. Regular Clinic Visit ✅
```json
{
  "clinic_patient_id": "patient-uuid",
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "department_id": "dept-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-20",
  "appointment_time": "2025-10-20 10:00:00",
  "consultation_type": "clinic_visit",    // ✅ Updated
  "is_follow_up": false,
  "payment_method": "pay_now",
  "payment_type": "cash"
}
```

---

### 2. Video Consultation ✅
```json
{
  "clinic_patient_id": "patient-uuid",
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-20",
  "appointment_time": "2025-10-20 14:00:00",
  "consultation_type": "video_consultation",  // ✅ Updated
  "is_follow_up": false,
  "payment_method": "pay_later"
}
```

---

### 3. Follow-Up Appointment (No Payment) ✅
```json
{
  "clinic_patient_id": "patient-uuid",
  "doctor_id": "doctor-uuid",
  "clinic_id": "clinic-uuid",
  "department_id": "dept-uuid",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-20",
  "appointment_time": "2025-10-20 15:00:00",
  "consultation_type": "follow_up",
  "is_follow_up": true                    // ✅ Flag set
  // ✅ NO payment_method needed
}
```

---

## ❌ Invalid Requests

### 1. Old Value (offline) ❌
```json
{
  "consultation_type": "offline"  // ❌ ERROR: Not in valid list
}
```

**Error:**
```json
{
  "error": "Invalid input",
  "details": "Key: 'SimpleAppointmentInput.ConsultationType' Error:Field validation for 'ConsultationType' failed on the 'oneof' tag"
}
```

---

### 2. Old Value (online) ❌
```json
{
  "consultation_type": "online"  // ❌ ERROR: Not in valid list
}
```

**Error:**
```json
{
  "error": "Invalid input",
  "details": "Key: 'SimpleAppointmentInput.ConsultationType' Error:Field validation for 'ConsultationType' failed on the 'oneof' tag"
}
```

---

### 3. Missing payment_method (Regular Appointment) ❌
```json
{
  "consultation_type": "clinic_visit",
  "is_follow_up": false
  // ❌ No payment_method
}
```

**Error:**
```json
{
  "error": "Payment method required",
  "message": "Please specify payment_method for regular appointments"
}
```

---

## 🔄 Migration Guide

### For Existing API Calls:

**Update:**
- ❌ `"consultation_type": "offline"` 
- ✅ `"consultation_type": "clinic_visit"`

**Update:**
- ❌ `"consultation_type": "online"` 
- ✅ `"consultation_type": "video_consultation"`

**Keep:**
- ✅ `"consultation_type": "in_person"` (unchanged)
- ✅ `"consultation_type": "follow_up"` (unchanged)

---

## 💻 Flutter Model Update

```dart
class SimpleAppointmentInput {
  final String clinicPatientId;
  final String doctorId;
  final String clinicId;
  final String? departmentId;
  final String individualSlotId;
  final String appointmentDate;
  final String appointmentTime;
  final String consultationType;  // ✅ Update enum
  final bool isFollowUp;
  final String? reason;
  final String? notes;
  final String? paymentMethod;
  final String? paymentType;
  
  // ✅ Updated consultation type enum
  static const List<String> validConsultationTypes = [
    'clinic_visit',         // ✅ Was 'offline'
    'video_consultation',   // ✅ Was 'online'
    'in_person',
    'follow_up',
  ];
  
  // ✅ Flutter dropdown
  static const Map<String, String> consultationTypeLabels = {
    'clinic_visit': '🏥 Clinic Visit',
    'video_consultation': '💻 Video Consultation',
    'in_person': '👤 In Person',
    'follow_up': '🔄 Follow-Up',
  };
}
```

---

## ✅ Validation Summary

| Aspect | Status |
|--------|--------|
| Binding updated | ✅ Done |
| Old values removed | ✅ Yes (`offline`, `online`) |
| New values added | ✅ Yes (`clinic_visit`, `video_consultation`) |
| Linter errors | ✅ None |
| Backward compatible | ❌ No (breaking change) |
| Documentation | ✅ Complete |

---

## 🚀 Deployment Notes

**Breaking Change:** ⚠️
- API now rejects `offline` and `online` values
- Clients must update to use `clinic_visit` and `video_consultation`

**Update Checklist:**
1. ✅ Backend binding updated
2. ⏳ Update Flutter app constants
3. ⏳ Update API documentation
4. ⏳ Notify frontend team

---

## 🧪 Testing

```bash
# ✅ Valid - New values
curl -X POST /api/appointments/simple \
  -d '{"consultation_type": "clinic_visit", ...}'

curl -X POST /api/appointments/simple \
  -d '{"consultation_type": "video_consultation", ...}'

# ❌ Invalid - Old values
curl -X POST /api/appointments/simple \
  -d '{"consultation_type": "offline", ...}'
# Returns: 400 Bad Request

curl -X POST /api/appointments/simple \
  -d '{"consultation_type": "online", ...}'
# Returns: 400 Bad Request
```

---

**Status:** ✅ **Binding updated and validated!** 🎉

