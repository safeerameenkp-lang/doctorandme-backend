# Follow-Up JSON Response - Complete Format 📋

## 🎯 **Complete Follow-Up Response After Booking Appointment**

Your appointment create API now returns the complete follow-up information with all required fields!

---

## 📤 **Complete Response Format**

### **Example: Regular Appointment (Creates Follow-Up)**

```json
{
  "message": "Appointment created successfully",
  
  "appointment": {
    "id": "a9876543-21ab-cdef-9876-543210abcdef",
    "clinic_patient_id": "f1234567-89ab-cdef-0123-456789abcdef",
    "clinic_id": "c7658c53-72ae-4bd3-9960-741225ebc0a2",
    "doctor_id": "d3456789-abcd-ef01-2345-6789abcdef01",
    "department_id": "dep98765-4321-0987-6543-210abcdef987",
    "booking_number": "AP-2025-CL001-0101",
    "token_number": 5,
    "appointment_date": "2025-10-26",
    "appointment_time": "2025-10-26T10:00:00Z",
    "consultation_type": "clinic_visit",
    "status": "confirmed",
    "fee_amount": 250.00,
    "payment_status": "paid",
    "payment_mode": "upi"
  },
  
  "follow_up": {
    "id": "fup-89b4d-9123",
    "clinic_patient_id": "f1234567-89ab-cdef-0123-456789abcdef",
    "clinic_id": "c7658c53-72ae-4bd3-9960-741225ebc0a2",
    "patient_name": "John Doe",
    "doctor_id": "d3456789-abcd-ef01-2345-6789abcdef01",
    "doctor_name": "Dr. Smith",
    "department_id": "dep98765-4321-0987-6543-210abcdef987",
    "department_name": "Cardiology",
    "source_appointment_id": "a9876543-21ab-cdef-9876-543210abcdef",
    "follow_up_status": "active",
    "is_free": true,
    "valid_from": "2025-10-26T10:00:00Z",
    "valid_until": "2025-10-31T10:00:00Z",
    "used_appointment_id": null,
    "used_at": null,
    "renewed_at": null,
    "renewed_by_appointment_id": null,
    "appointment_slot_type": "clinic_visit",
    "follow_up_type": "",
    "days_remaining": 5,
    "created_at": "2025-10-26T10:00:00Z",
    "updated_at": "2025-10-26T10:00:00Z"
  },
  
  "clinic_patient_update": {
    "current_followup_status": "active",
    "last_appointment_id": "a9876543-21ab-cdef-9876-543210abcdef",
    "last_followup_id": "fup-89b4d-9123"
  },
  
  "is_regular_appointment": true,
  "followup_granted": true,
  "followup_message": "Free follow-up eligibility granted (valid for 5 days)",
  "followup_valid_until": "2025-10-31",
  
  "renewal_options": {
    "can_renew": false,
    "message": "Patient has active follow-up. Cannot renew until used or expired."
  }
}
```

### **Example: Free Follow-Up Appointment**

```json
{
  "message": "Appointment created successfully",
  
  "appointment": {
    "id": "b8f45d3c-2a1b-4f3e-9c6d-5e8a7b9c0d1e",
    "consultation_type": "follow-up-via-clinic",
    "status": "confirmed",
    "fee_amount": 0.00,
    "payment_status": "waived"
  },
  
  "follow_up": {
    "id": "fup-89b4d-9123",
    "clinic_patient_id": "f1234567-89ab-cdef-0123-456789abcdef",
    "clinic_id": "c7658c53-72ae-4bd3-9960-741225ebc0a2",
    "patient_name": "John Doe",
    "doctor_id": "d3456789-abcd-ef01-2345-6789abcdef01",
    "doctor_name": "Dr. Smith",
    "department_id": "dep98765-4321-0987-6543-210abcdef987",
    "department_name": "Cardiology",
    "source_appointment_id": "a9876543-21ab-cdef-9876-543210abcdef",
    "follow_up_status": "used",
    "is_free": true,
    "valid_from": "2025-10-26T10:00:00Z",
    "valid_until": "2025-10-31T10:00:00Z",
    "used_appointment_id": "b8f45d3c-2a1b-4f3e-9c6d-5e8a7b9c0d1e",
    "used_at": "2025-10-27T14:00:00Z",
    "appointment_slot_type": "follow-up-via-clinic",
    "follow_up_type": "clinic_followup",
    "days_remaining": 4,
    "created_at": "2025-10-26T10:00:00Z",
    "updated_at": "2025-10-27T14:00:00Z"
  },
  
  "clinic_patient_update": {
    "current_followup_status": "used",
    "last_appointment_id": "b8f45d3c-2a1b-4f3e-9c6d-5e8a7b9c0d1e"
  },
  
  "is_free_followup": true,
  "followup_type": "free",
  "followup_message": "This is a FREE follow-up"
}
```

---

## 📋 **Follow-Up Response Fields Explained**

### **Identity Fields**
| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Follow-up ID |
| `clinic_patient_id` | UUID | Patient linked to this clinic |
| `clinic_id` | UUID | Clinic where appointment/follow-up is booked |

### **Patient Info**
| Field | Type | Description |
|-------|------|-------------|
| `patient_name` | String | Full patient name (First + Last) |

### **Doctor & Department Info**
| Field | Type | Description |
|-------|------|-------------|
| `doctor_id` | UUID | Doctor of the follow-up |
| `doctor_name` | String | Doctor name with "Dr." prefix |
| `department_id` | UUID | Department ID |
| `department_name` | String | Department name |

### **Follow-Up Status**
| Field | Type | Description |
|-------|------|-------------|
| `follow_up_status` | String | `active` (available) \| `used` (already booked) \| `expired` (past validity) \| `renewed` (replaced by newer) |
| `is_free` | Boolean | `true` if first follow-up eligible for free |
| `source_appointment_id` | UUID | Appointment that generated this follow-up |

### **Validity Period**
| Field | Type | Description |
|-------|------|-------------|
| `valid_from` | DateTime | Source appointment datetime (ISO 8601) |
| `valid_until` | DateTime | 5 days validity window (ISO 8601) |
| `days_remaining` | Int | Days left for free follow-up eligibility |

### **Usage Tracking**
| Field | Type | Description |
|-------|------|-------------|
| `used_appointment_id` | UUID | If already used, appointment ID here |
| `used_at` | DateTime | Timestamp when follow-up was used (ISO 8601) |
| `renewed_at` | DateTime | Timestamp if renewed (ISO 8601) |
| `renewed_by_appointment_id` | UUID | New appointment if renewal happened |

### **Type Information**
| Field | Type | Description |
|-------|------|-------------|
| `appointment_slot_type` | String | `clinic_visit` \| `video_consultation` \| `follow-up-via-clinic` \| `follow-up-via-video` |
| `follow_up_type` | String | `clinic_followup` \| `video_followup` (if follow-up appointment) |

### **Timestamps**
| Field | Type | Description |
|-------|------|-------------|
| `created_at` | DateTime | Timestamp when follow-up was created (ISO 8601) |
| `updated_at` | DateTime | Last update timestamp (ISO 8601) |

---

## 🔄 **Status Flow Examples**

### **Status: active**
```json
{
  "follow_up_status": "active",
  "is_free": true,
  "days_remaining": 5,
  "used_appointment_id": null,
  "used_at": null,
  "renewed_at": null
}
```
**Meaning:** Follow-up is available for booking (free within 5 days)

### **Status: used**
```json
{
  "follow_up_status": "used",
  "is_free": true,
  "days_remaining": 3,
  "used_appointment_id": "appointment-id-that-used-it",
  "used_at": "2025-10-27T14:00:00Z",
  "follow_up_type": "clinic_followup"
}
```
**Meaning:** Free follow-up has been consumed

### **Status: expired**
```json
{
  "follow_up_status": "expired",
  "is_free": true,
  "days_remaining": 0,
  "valid_until": "2025-10-26T10:00:00Z",
  "used_appointment_id": null,
  "used_at": null
}
```
**Meaning:** Follow-up expired (past 5 days)

### **Status: renewed**
```json
{
  "follow_up_status": "renewed",
  "is_free": true,
  "renewed_at": "2025-11-01T11:00:00Z",
  "renewed_by_appointment_id": "new-appointment-id",
  "days_remaining": 0
}
```
**Meaning:** This follow-up was replaced by a newer one

---

## ✅ **All Fields Implemented**

✅ `clinic_patient_id` - Patient linked to this clinic  
✅ `clinic_id` - Clinic where appointment/follow-up is booked  
✅ `patient_name` - Patient full name  
✅ `doctor_id` - Doctor of the follow-up  
✅ `doctor_name` - Doctor name with "Dr." prefix  
✅ `department_id` - Department ID  
✅ `department_name` - Department name  
✅ `source_appointment_id` - Appointment that generated this follow-up  
✅ `follow_up_status` - active, used, expired, renewed  
✅ `is_free` - Free follow-up eligibility  
✅ `valid_from` - Source appointment datetime (ISO 8601)  
✅ `valid_until` - 5 days validity window (ISO 8601)  
✅ `used_appointment_id` - If already used  
✅ `used_at` - Timestamp when follow-up was used  
✅ `renewed_at` - Timestamp if renewed  
✅ `renewed_by_appointment_id` - New appointment if renewal happened  
✅ `appointment_slot_type` - clinic_visit, video_consultation, follow-up types  
✅ `follow_up_type` - clinic_followup, video_followup  
✅ `days_remaining` - Days left for free follow-up  
✅ `created_at` - Creation timestamp (ISO 8601)  
✅ `updated_at` - Last update timestamp (ISO 8601)  

---

## 🚀 **Your follow-up response is complete!**

All fields implemented as per your requirements. The API now returns the full production-ready follow-up JSON with all the details you specified! 🎉

