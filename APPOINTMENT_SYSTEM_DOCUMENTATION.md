# 📋 COMPLETE APPOINTMENT SYSTEM DOCUMENTATION

## 🎯 **Overview**
This document provides complete API documentation for the appointment system, including follow-up management, with JSON examples for frontend integration.

---

## 🔗 **API Endpoints**

### 1. **Create Simple Appointment**
**Endpoint:** `POST /api/v1/appointments/simple`  
**Description:** Creates appointments for clinic-specific patients with automatic follow-up management

#### **Request Body:**
```json
{
  "clinic_patient_id": "d27a8fa7-b8bc-43e3-837b-87db5dfd4bed",
  "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
  "individual_slot_id": "0d1ed772-114d-41d6-b780-96ab0cd2d6d2",
  "appointment_date": "2025-10-27",
  "appointment_time": "2025-10-27 14:37:00",
  "consultation_type": "clinic_visit",
  "reason": "Regular checkup",
  "notes": "Patient notes",
  "payment_method": "pay_now",
  "payment_type": "cash"
}
```

#### **Response - Regular Appointment (Success):**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "clinic_patient_id": "d27a8fa7-b8bc-43e3-837b-87db5dfd4bed",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
    "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
    "booking_number": "DOC001-20251027-0001",
    "token_number": 1,
    "appointment_date": "2025-10-27",
    "appointment_time": "2025-10-27T14:37:00Z",
    "duration_minutes": 5,
    "consultation_type": "clinic_visit",
    "reason": "Regular checkup",
    "notes": "Patient notes",
    "status": "confirmed",
    "fee_amount": 100.00,
    "payment_status": "paid",
    "payment_mode": "cash",
    "created_at": "2025-10-25T10:30:00Z"
  },
  "is_regular_appointment": true,
  "followup_granted": true,
  "followup_message": "Free follow-up eligibility granted (valid for 5 days)",
  "followup_valid_until": "2025-11-01"
}
```

#### **Response - Free Follow-up (Success):**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "b2c3d4e5-f6g7-8901-bcde-f23456789012",
    "clinic_patient_id": "d27a8fa7-b8bc-43e3-837b-87db5dfd4bed",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
    "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
    "booking_number": "DOC001-20251027-0002",
    "token_number": 2,
    "appointment_date": "2025-10-28",
    "appointment_time": "2025-10-28T10:00:00Z",
    "duration_minutes": 5,
    "consultation_type": "follow-up-via-clinic",
    "reason": "Follow-up visit",
    "notes": "Follow-up notes",
    "status": "confirmed",
    "fee_amount": 0.00,
    "payment_status": "waived",
    "payment_mode": null,
    "created_at": "2025-10-25T10:35:00Z"
  },
  "is_free_followup": true,
  "followup_type": "free",
  "followup_message": "This is a FREE follow-up (renewed after regular appointment)"
}
```

#### **Response - Paid Follow-up (Success):**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "c3d4e5f6-g7h8-9012-cdef-345678901234",
    "clinic_patient_id": "d27a8fa7-b8bc-43e3-837b-87db5dfd4bed",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
    "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
    "booking_number": "DOC001-20251029-0001",
    "token_number": 1,
    "appointment_date": "2025-10-29",
    "appointment_time": "2025-10-29T11:00:00Z",
    "duration_minutes": 5,
    "consultation_type": "follow-up-via-clinic",
    "reason": "Follow-up visit",
    "notes": "Follow-up notes",
    "status": "confirmed",
    "fee_amount": 50.00,
    "payment_status": "paid",
    "payment_mode": "card",
    "created_at": "2025-10-25T10:40:00Z"
  },
  "is_free_followup": false,
  "followup_type": "paid",
  "followup_message": "This is a PAID follow-up (free follow-up already used or expired)"
}
```

#### **Error Responses:**

**Patient Not Found:**
```json
{
  "error": "Patient not found"
}
```

**Not Eligible for Follow-up:**
```json
{
  "error": "Not eligible for follow-up",
  "message": "No previous appointment found with this doctor"
}
```

**Payment Method Required:**
```json
{
  "error": "Payment method required",
  "message": "Please specify payment_method for appointments"
}
```

**Slot Not Available:**
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

---

### 2. **Check Follow-up Eligibility**
**Endpoint:** `GET /api/v1/appointments/followup-eligibility`  
**Description:** Check if a patient is eligible for follow-up with a specific doctor+department

#### **Query Parameters:**
- `clinic_patient_id` (required): Patient ID
- `clinic_id` (required): Clinic ID
- `doctor_id` (required): Doctor ID
- `department_id` (optional): Department ID

#### **Request Example:**
```
GET /api/v1/appointments/followup-eligibility?clinic_patient_id=d27a8fa7-b8bc-43e3-837b-87db5dfd4bed&clinic_id=f7658c53-72ae-4bd3-9960-741225ebc0a2&doctor_id=ef378478-1091-472e-af40-1655e77985b3&department_id=ad958b90-d383-4478-bfe3-08b53b8eeef7
```

#### **Response - Free Follow-up Available:**
```json
{
  "is_free": true,
  "is_eligible": true,
  "message": "Free follow-up available (3 days remaining)",
  "followup_details": {
    "followup_id": "5ffaed5c-ee2b-492a-8039-dfd7d2a69c25",
    "valid_until": "2025-10-30",
    "days_remaining": 3,
    "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
    "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7"
  }
}
```

#### **Response - Paid Follow-up Available:**
```json
{
  "is_free": false,
  "is_eligible": true,
  "message": "Follow-up available (payment required)",
  "followup_details": {
    "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
    "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
    "previous_appointment_date": "2025-10-20"
  }
}
```

#### **Response - Not Eligible:**
```json
{
  "is_free": false,
  "is_eligible": false,
  "message": "No previous appointment found with this doctor"
}
```

---

### 3. **List Active Follow-ups**
**Endpoint:** `GET /api/v1/appointments/followup-eligibility/active`  
**Description:** Get all active follow-ups for a patient

#### **Query Parameters:**
- `clinic_patient_id` (required): Patient ID
- `clinic_id` (required): Clinic ID

#### **Response:**
```json
{
  "active_followups": [
    {
      "followup_id": "5ffaed5c-ee2b-492a-8039-dfd7d2a69c25",
      "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
      "doctor_name": "Dr. Smith",
      "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
      "department_name": "Cardiology",
      "appointment_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "appointment_date": "2025-10-27",
      "valid_until": "2025-10-30",
      "days_remaining": 3,
      "is_free": true
    }
  ],
  "total_count": 1
}
```

---

### 4. **Get Clinic Patient List**
**Endpoint:** `GET /api/v1/clinic-specific-patients`  
**Description:** Get all patients for a clinic with follow-up status

#### **Query Parameters:**
- `clinic_id` (required): Clinic ID
- `search` (optional): Search term
- `only_active` (optional): Show only active patients (default: true)
- `doctor_id` (optional): Filter by doctor
- `department_id` (optional): Filter by department

#### **Response:**
```json
{
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "total": 1,
  "patients": [
    {
      "id": "d27a8fa7-b8bc-43e3-837b-87db5dfd4bed",
      "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
      "first_name": "ashiq",
      "last_name": "m",
      "phone": "+1234567890",
      "email": "ashiq@example.com",
      "date_of_birth": "1990-01-01",
      "age": 34,
      "gender": "male",
      "address1": "123 Main St",
      "address2": "Apt 4B",
      "district": "Downtown",
      "state": "CA",
      "mo_id": "MO123456",
      "medical_history": "Diabetes",
      "allergies": "None",
      "blood_group": "O+",
      "smoking_status": "non_smoker",
      "alcohol_use": "none",
      "height_cm": 175,
      "weight_kg": 70,
      "is_active": true,
      "created_at": "2025-10-20T08:00:00Z",
      "updated_at": "2025-10-25T10:30:00Z",
      "eligible_follow_ups": [
        {
          "appointment_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
          "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
          "doctor_name": "Dr. Smith",
          "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
          "department": "Cardiology",
          "appointment_date": "2025-10-27",
          "remaining_days": 3,
          "next_follow_up_expiry": "2025-10-30",
          "note": "Free follow-up available"
        }
      ],
      "expired_followups": [],
      "appointment_history": [
        {
          "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
          "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
          "doctor_name": "Dr. Smith",
          "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
          "department": "Cardiology",
          "consultation_type": "clinic_visit",
          "appointment_date": "2025-10-27"
        }
      ]
    }
  ]
}
```

---

## 🔄 **Follow-up System Logic**

### **How Follow-ups Work:**

1. **Regular Appointment Creation:**
   - When `consultation_type` is `clinic_visit` or `video_consultation`
   - Automatically creates a new active follow-up record
   - Valid for 5 days from appointment date
   - Marks any existing follow-ups as `renewed`

2. **Follow-up Appointment Creation:**
   - When `consultation_type` is `follow-up-via-clinic` or `follow-up-via-video`
   - Checks for active free follow-ups
   - If free follow-up exists: Creates appointment with `payment_status: "waived"` and `fee_amount: 0`
   - If no free follow-up: Requires payment and uses follow-up fee

3. **Renewal System:**
   - New regular appointment with same doctor+department
   - Marks old follow-ups as `renewed`
   - Creates new active follow-up for 5 days

---

## 🎨 **Frontend Integration Guide**

### **1. Appointment Creation Flow:**

```javascript
// Create regular appointment
const createRegularAppointment = async (appointmentData) => {
  const response = await fetch('/api/v1/appointments/simple', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      ...appointmentData,
      consultation_type: 'clinic_visit',
      payment_method: 'pay_now',
      payment_type: 'cash'
    })
  });
  
  const result = await response.json();
  
  if (response.ok) {
    // Show success message
    if (result.followup_granted) {
      showMessage(`✅ Appointment created! Free follow-up available until ${result.followup_valid_until}`);
    }
  } else {
    showError(result.error);
  }
};
```

### **2. Follow-up Creation Flow:**

```javascript
// Create follow-up appointment
const createFollowUpAppointment = async (appointmentData) => {
  // First check eligibility
  const eligibilityResponse = await fetch(
    `/api/v1/appointments/followup-eligibility?clinic_patient_id=${appointmentData.clinic_patient_id}&clinic_id=${appointmentData.clinic_id}&doctor_id=${appointmentData.doctor_id}&department_id=${appointmentData.department_id}`,
    {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }
  );
  
  const eligibility = await eligibilityResponse.json();
  
  if (!eligibility.is_eligible) {
    showError(eligibility.message);
    return;
  }
  
  // Create follow-up appointment
  const response = await fetch('/api/v1/appointments/simple', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      ...appointmentData,
      consultation_type: 'follow-up-via-clinic',
      // Payment method only required if not free
      ...(eligibility.is_free ? {} : {
        payment_method: 'pay_now',
        payment_type: 'cash'
      })
    })
  });
  
  const result = await response.json();
  
  if (response.ok) {
    if (result.is_free_followup) {
      showMessage('✅ Free follow-up appointment created!');
    } else {
      showMessage('✅ Paid follow-up appointment created!');
    }
  } else {
    showError(result.error);
  }
};
```

### **3. Patient List Display:**

```javascript
// Display patient list with follow-up status
const displayPatientList = (patients) => {
  patients.forEach(patient => {
    const followUpStatus = getFollowUpStatus(patient);
    console.log(`Patient: ${patient.first_name} ${patient.last_name}`);
    console.log(`Follow-up Status: ${followUpStatus}`);
  });
};

const getFollowUpStatus = (patient) => {
  if (patient.eligible_follow_ups && patient.eligible_follow_ups.length > 0) {
    const followUp = patient.eligible_follow_ups[0];
    return `✅ Free follow-up available (${followUp.remaining_days} days remaining)`;
  } else if (patient.expired_followups && patient.expired_followups.length > 0) {
    return '⚠️ Follow-up expired - create new regular appointment to renew';
  } else {
    return '❌ No follow-up available';
  }
};
```

### **4. Follow-up Status Check:**

```javascript
// Check follow-up status for specific doctor+department
const checkFollowUpStatus = async (patientId, clinicId, doctorId, departmentId) => {
  const response = await fetch(
    `/api/v1/appointments/followup-eligibility?clinic_patient_id=${patientId}&clinic_id=${clinicId}&doctor_id=${doctorId}&department_id=${departmentId}`,
    {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }
  );
  
  const result = await response.json();
  
  if (result.is_eligible) {
    if (result.is_free) {
      return {
        status: 'free',
        message: result.message,
        daysRemaining: result.followup_details.days_remaining
      };
    } else {
      return {
        status: 'paid',
        message: result.message
      };
    }
  } else {
    return {
      status: 'not_eligible',
      message: result.message
    };
  }
};
```

---

## 📊 **Database Schema**

### **Follow-ups Table:**
```sql
CREATE TABLE follow_ups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clinic_patient_id UUID NOT NULL REFERENCES clinic_patients(id) ON DELETE CASCADE,
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    source_appointment_id UUID NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, used, expired, renewed
    is_free BOOLEAN NOT NULL DEFAULT TRUE,
    valid_from DATE NOT NULL,
    valid_until DATE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    used_appointment_id UUID REFERENCES appointments(id) ON DELETE SET NULL,
    renewed_at TIMESTAMP WITH TIME ZONE,
    renewed_by_appointment_id UUID REFERENCES appointments(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

---

## 🚨 **Error Handling**

### **Common Error Scenarios:**

1. **Authentication Required:**
```json
{
  "error": "Authentication required",
  "message": "Please provide a valid authorization token in the request header",
  "code": "MISSING_TOKEN"
}
```

2. **Invalid Input:**
```json
{
  "error": "Invalid input",
  "details": "Key: 'SimpleAppointmentInput.ClinicPatientID' Error:Field validation for 'ClinicPatientID' failed on the 'required' tag"
}
```

3. **Slot Conflict:**
```json
{
  "error": "Slot just got booked",
  "message": "This slot was just booked by another patient. Please select another slot."
}
```

---

## 🔧 **Testing Examples**

### **Test Regular Appointment Creation:**
```bash
curl -X POST "http://localhost:8082/api/v1/appointments/simple" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "clinic_patient_id": "d27a8fa7-b8bc-43e3-837b-87db5dfd4bed",
    "doctor_id": "ef378478-1091-472e-af40-1655e77985b3",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "department_id": "ad958b90-d383-4478-bfe3-08b53b8eeef7",
    "individual_slot_id": "0d1ed772-114d-41d6-b780-96ab0cd2d6d2",
    "appointment_date": "2025-10-27",
    "appointment_time": "2025-10-27 14:37:00",
    "consultation_type": "clinic_visit",
    "payment_method": "pay_now",
    "payment_type": "cash"
  }'
```

### **Test Follow-up Eligibility Check:**
```bash
curl -X GET "http://localhost:8082/api/v1/appointments/followup-eligibility?clinic_patient_id=d27a8fa7-b8bc-43e3-837b-87db5dfd4bed&clinic_id=f7658c53-72ae-4bd3-9960-741225ebc0a2&doctor_id=ef378478-1091-472e-af40-1655e77985b3&department_id=ad958b90-d383-4478-bfe3-08b53b8eeef7" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## ✅ **Summary**

This documentation provides complete API reference for:
- ✅ **Appointment Creation** (Regular & Follow-up)
- ✅ **Follow-up Eligibility Checking**
- ✅ **Patient List with Follow-up Status**
- ✅ **Renewal System**
- ✅ **Error Handling**
- ✅ **Frontend Integration Examples**

The system automatically handles follow-up creation, renewal, and eligibility checking, making it easy for the frontend to provide a seamless user experience.
