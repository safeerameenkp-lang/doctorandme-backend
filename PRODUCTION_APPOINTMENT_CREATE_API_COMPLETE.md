# Production-Level Appointment Create API - Complete Documentation 📘

## 🎯 **Overview**

Complete documentation for the Appointment Create API with full follow-up checking, status tracking, and production-level validation.

---

## 📋 **API Endpoint**

**Endpoint:** `POST /api/appointments/simple`

**Authentication:** Bearer Token Required

**Content-Type:** `application/json`

---

## 📥 **Request**

### **Request Headers**
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

### **Request Body**

```json
{
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
  "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
  "department_id": "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
  "individual_slot_id": "slot-uuid-123",
  "appointment_date": "2025-10-26",
  "appointment_time": "10:30:00",
  "consultation_type": "clinic_visit",
  "payment_method": "pay_now",
  "payment_type": "upi",
  "reason": "Regular checkup",
  "notes": "Patient complaint"
}
```

### **Field Validation Rules**

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `clinic_id` | UUID | ✅ Yes | Valid UUID |
| `clinic_patient_id` | UUID | ✅ Yes | Valid UUID, must belong to same clinic |
| `doctor_id` | UUID | ✅ Yes | Valid UUID, active doctor |
| `department_id` | UUID | ❌ Optional | Valid UUID |
| `individual_slot_id` | UUID | ✅ Yes | Available slot |
| `appointment_date` | Date | ✅ Yes | YYYY-MM-DD format |
| `appointment_time` | Time | ✅ Yes | YYYY-MM-DD HH:MM:SS format |
| `consultation_type` | String | ✅ Yes | One of: clinic_visit, video_consultation, follow-up-via-clinic, follow-up-via-video |
| `payment_method` | String | ⚠️ Conditional | Required for paid appointments |
| `payment_type` | String | ⚠️ Conditional | Required when payment_method=pay_now |

---

## 📤 **Response**

### **Success Response (Regular Appointment)**

```json
{
  "message": "Appointment created successfully",
  
  "appointment": {
    "id": "a6b77b4c-71f9-4ff2-bd09-ff2cb7e92c99",
    "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
    "department_id": "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
    "booking_number": "AP-2025-CL001-0101",
    "token_number": 5,
    "appointment_date": "2025-10-26",
    "appointment_time": "2025-10-26T10:30:00Z",
    "consultation_type": "clinic_visit",
    "status": "confirmed",
    "fee_amount": 250.00,
    "payment_status": "paid",
    "payment_mode": "upi",
    "is_priority": false,
    "created_at": "2025-10-25T10:00:00Z"
  },
  
  "follow_up": {
    "id": "fup-89b4d-9123",
    "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
    "department_id": "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
    "source_appointment_id": "a6b77b4c-71f9-4ff2-bd09-ff2cb7e92c99",
    "status": "active",
    "is_free": true,
    "valid_from": "2025-10-26",
    "valid_until": "2025-10-31",
    "days_remaining": 5,
    "used_appointment_id": null,
    "renewed_by_appointment_id": null,
    "created_at": "2025-10-26T10:00:00Z",
    "updated_at": "2025-10-26T10:00:00Z"
  },
  
  "clinic_patient_update": {
    "current_followup_status": "active",
    "last_appointment_id": "a6b77b4c-71f9-4ff2-bd09-ff2cb7e92c99",
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

### **Success Response (Free Follow-Up Appointment)**

```json
{
  "message": "Appointment created successfully",
  
  "appointment": {
    "id": "b8f45d3c-2a1b-4f3e-9c6d-5e8a7b9c0d1e",
    "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
    "department_id": "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
    "consultation_type": "follow-up-via-clinic",
    "status": "confirmed",
    "fee_amount": 0.00,
    "payment_status": "waived"
  },
  
  "is_free_followup": true,
  
  "follow_up_info": {
    "is_followup": true,
    "is_free": true,
    "follow_up_status": "used",
    "message": "This is a FREE follow-up (renewed after regular appointment)"
  },
  
  "clinic_patient_update": {
    "current_followup_status": "used",
    "last_appointment_id": "b8f45d3c-2a1b-4f3e-9c6d-5e8a7b9c0d1e"
  }
}
```

### **Success Response (Renewal)**

```json
{
  "message": "Appointment created successfully",
  
  "appointment": {
    "id": "a7b82e6f-5a3b-4237-a09e-df8a82366c32",
    "consultation_type": "clinic_visit",
    "status": "confirmed"
  },
  
  "follow_up": {
    "id": "new-fup-9921",
    "source_appointment_id": "a7b82e6f-5a3b-4237-a09e-df8a82366c32",
    "status": "active",
    "is_free": true,
    "valid_until": "2025-11-05",
    "days_remaining": 10
  },
  
  "clinic_patient_update": {
    "current_followup_status": "renewed",
    "last_appointment_id": "a7b82e6f-5a3b-4237-a09e-df8a82366c32",
    "last_followup_id": "new-fup-9921"
  },
  
  "follow_up_action": {
    "previous_followup_id": "fup-89b4d-9123",
    "previous_status": "expired",
    "new_followup_created": true,
    "renewed_by_appointment_id": "a7b82e6f-5a3b-4237-a09e-df8a82366c32"
  }
}
```

---

## ✅ **Complete Validation & Follow-Up Checks**

### **Step 1: Patient Validation**
```go
// ✅ Check patient exists and belongs to this clinic
SELECT clinic_id FROM clinic_patients 
WHERE id = clinic_patient_id AND is_active = true

// If mismatch → Error: "Patient belongs to different clinic"
```

### **Step 2: Follow-Up Eligibility Check**
```go
// ✅ If booking follow-up, check eligibility
if isFollowUp {
    followUpManager.CheckFollowUpEligibility(
        clinic_patient_id, clinic_id, doctor_id, department_id
    )
    
    // Returns: isFree, isEligible, message, daysRemaining
    // If not eligible → Error: "Not eligible for follow-up"
}
```

### **Step 3: Payment Validation**
```go
// ✅ Payment required for:
if !isFollowUp || (isFollowUp && !isFree) {
    // Payment required
    if paymentMethod == null {
        Error: "Payment method required"
    }
}

// ✅ Free follow-ups don't require payment
if isFollowUp && isFree {
    paymentStatus = "waived"
    feeAmount = 0.0
}
```

### **Step 4: Slot Validation**
```go
// ✅ Check slot is available
SELECT available_count, status 
FROM doctor_individual_slots 
WHERE id = individual_slot_id

// If available_count <= 0 → Error: "Slot not available"
// If status != "available" → Error: "Slot fully booked"
```

### **Step 5: Follow-Up Creation**
```go
// ✅ For regular appointments, create follow-up
if consultation_type == "clinic_visit" || "video_consultation" {
    // Check if renewal or new
    existingFollowUp = GetActiveFollowUp(...)
    
    if existingFollowUp != nil {
        newStatus = "renewed"
        // Mark old follow-up as renewed
    } else {
        newStatus = "active"
        // Create new follow-up
    }
    
    // Create follow-up record
    CreateFollowUp(
        clinic_patient_id, clinic_id, doctor_id, 
        department_id, appointment_id, appointment_date
    )
    
    // Update clinic_patient status
    Update clinic_patients SET 
        current_followup_status = newStatus,
        last_appointment_id = appointment_id,
        last_followup_id = followup_id
}
```

### **Step 6: Follow-Up Usage**
```go
// ✅ For free follow-ups, mark as used
if isFollowUp && isFree {
    MarkFollowUpAsUsed(...)
    
    // Update clinic_patient status
    Update clinic_patients SET 
        current_followup_status = "used",
        last_appointment_id = appointment_id
}
```

---

## 🔄 **Complete Follow-Up Status Flow**

### **Flow 1: Book Regular Appointment**
```
1. Create appointment
   ↓
2. Create follow-up record
   → Status: "active"
   → is_free: true
   → valid_until: appointment_date + 5 days
   ↓
3. Update clinic_patient
   → current_followup_status = "active"
   → last_appointment_id = appointment_id
   → last_followup_id = followup_id
```

### **Flow 2: Use Free Follow-Up**
```
1. Book follow-up (within 5 days, first follow-up)
   ↓
2. Mark follow-up as used
   → Status: "used"
   → used_appointment_id = appointment_id
   ↓
3. Update clinic_patient
   → current_followup_status = "used"
   → last_appointment_id = appointment_id
```

### **Flow 3: Follow-Up Expires**
```
1. Wait 5+ days
   ↓
2. Auto-update follow-up status
   → Status: "expired"
   ↓
3. Update clinic_patient
   → current_followup_status = "expired"
```

### **Flow 4: Renewal After Expiry**
```
1. Book new regular appointment (same doctor+dept)
   ↓
2. Detect renewal
   → Renew old follow-up (status = "renewed")
   → Create new follow-up (status = "active")
   ↓
3. Update clinic_patient
   → current_followup_status = "renewed" then "active"
   → last_appointment_id = new_appointment_id
   → last_followup_id = new_followup_id
```

---

## 📊 **Response Fields Explained**

### **Appointment Object**
| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Appointment ID |
| `clinic_patient_id` | UUID | Patient reference |
| `clinic_id` | UUID | Clinic reference |
| `doctor_id` | UUID | Doctor reference |
| `department_id` | UUID | Department reference |
| `booking_number` | String | Unique booking number |
| `token_number` | Int | Token number |
| `appointment_date` | Date | YYYY-MM-DD |
| `appointment_time` | DateTime | ISO 8601 |
| `consultation_type` | String | clinic_visit, video_consultation, follow-up |
| `status` | String | booked, confirmed, completed |
| `fee_amount` | Decimal | Appointment fee |
| `payment_status` | String | paid, waived, pending |
| `payment_mode` | String | cash, card, upi, online |

### **Follow-Up Object**
| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Follow-up ID |
| `status` | String | active, used, expired, renewed |
| `is_free` | Boolean | Free follow-up or paid |
| `valid_from` | Date | When follow-up starts |
| `valid_until` | Date | When follow-up expires |
| `days_remaining` | Int | Days left for free follow-up |
| `used_appointment_id` | UUID | Which appointment used it (if used) |
| `source_appointment_id` | UUID | Original appointment that granted follow-up |

### **Clinic Patient Update Object**
| Field | Type | Description |
|-------|------|-------------|
| `current_followup_status` | String | none, active, used, expired, renewed |
| `last_appointment_id` | UUID | Last appointment reference |
| `last_followup_id` | UUID | Last follow-up reference |

---

## 🎉 **Complete Production Features**

### ✅ **Implemented**
1. Patient validation (exists, belongs to clinic)
2. Slot validation (available capacity check)
3. Follow-up eligibility checking
4. Free follow-up detection
5. Renewal detection
6. Status lifecycle management
7. Complete follow-up details in response
8. Clinic patient status updates
9. Race condition prevention (slot booking)
10. Payment validation logic

### ✅ **Follow-Up Checks**
- ✅ First appointment → creates free follow-up
- ✅ Within 5 days → follow-up is free
- ✅ After 5 days → follow-up is paid
- ✅ Different doctor/dept → new paid appointment
- ✅ Renewal detection → auto-renew follow-up
- ✅ Status tracking → updates clinic_patient

---

## 🚀 **Ready for Production!**

Your appointment creation API now includes:
- ✅ Complete follow-up checking
- ✅ Production-level validation
- ✅ Status lifecycle management
- ✅ Comprehensive response format
- ✅ All requirements from Production-Level Checklist

**API is production-ready! 🎉**

