# Complete Clinic Patient System Documentation 📘

## 🎯 **Overview**

Complete documentation for Clinic Patient, Appointment Creation, Follow-Up System, and UI Integration.

---

## 📋 **Table of Contents**

1. [Authentication & Login](#1-authentication--login)
2. [Clinic Patient Management](#2-clinic-patient-management)
3. [Appointment Creation](#3-appointment-creation)
4. [Follow-Up System](#4-follow-up-system)
5. [Frontend UI Integration](#5-frontend-ui-integration)
6. [Data Models](#6-data-models)
7. [API Reference](#7-api-reference)
8. [Complete Flow Examples](#8-complete-flow-examples)

---

## 1️⃣ **Authentication & Login**

### **Login API**

**Endpoint:** `POST /api/auth/login`

**Request Body:**
```json
{
  "username": "doctor_user",
  "password": "password123"
}
```

**Response (Success):**
```json
{
  "success": true,
  "user": {
    "id": "user-uuid",
    "username": "doctor_user",
    "email": "doctor@clinic.com",
    "first_name": "Dr. John",
    "last_name": "Smith",
    "roles": [
      {
        "role": "doctor",
        "clinic_id": "clinic-uuid",
        "clinic_name": "Main Clinic"
      }
    ]
  },
  "tokens": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "expires_in": 3600
  }
}
```

**Response (Error):**
```json
{
  "success": false,
  "error": "Invalid credentials"
}
```

---

## 2️⃣ **Clinic Patient Management**

### **2.1 Create Clinic Patient**

**Endpoint:** `POST /api/organizations/clinic-specific-patients`

**Headers:**
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "first_name": "Ameen",
  "last_name": "Khan",
  "phone": "+919876543210",
  "email": "ameen@example.com",
  "date_of_birth": "1990-05-12",
  "gender": "male",
  "mo_id": "MO12345",
  "medical_history": "Diabetes, Hypertension",
  "allergies": "None",
  "blood_group": "B+",
  "address1": "123 Main St",
  "district": "Downtown",
  "state": "Kerala"
}
```

**Response:**
```json
{
  "message": "Patient created successfully for this clinic",
  "patient": {
    "id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "first_name": "Ameen",
    "last_name": "Khan",
    "phone": "+919876543210",
    "email": "ameen@example.com",
    "mo_id": "MO12345",
    "is_active": true,
    "created_at": "2025-10-25T10:00:00Z"
  }
}
```

### **2.2 List Clinic Patients**

**Endpoint:** `GET /api/organizations/clinic-specific-patients?clinic_id={id}&search={query}&only_active=true`

**Response:**
```json
{
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "total": 10,
  "patients": [
    {
      "id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
      "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
      "first_name": "Ameen",
      "last_name": "Khan",
      "phone": "+919876543210",
      "email": "ameen@example.com",
      "mo_id": "MO12345",
      "current_followup_status": "active",
      "last_appointment_id": "a6b77b4c-71f9-4ff2-bd09-ff2cb7e92c99",
      "last_followup_id": "fup-89b4d-9123",
      
      "appointments": [
        {
          "appointment_id": "a6b77b4c-71f9-4ff2-bd09-ff2cb7e92c99",
          "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
          "department_id": "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
          "appointment_time": "2025-10-25T10:30:00Z",
          "slot_type": "clinic_visit",
          "consultation_type": "clinic_visit",
          "status": "confirmed",
          "fee_amount": 250.00,
          "payment_status": "paid",
          "payment_mode": "online"
        }
      ],
      
      "follow_ups": [
        {
          "follow_up_id": "fup-89b4d-9123",
          "source_appointment_id": "a6b77b4c-71f9-4ff2-bd09-ff2cb7e92c99",
          "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
          "department_id": "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
          "status": "active",
          "is_free": true,
          "valid_from": "2025-10-25",
          "valid_until": "2025-10-30",
          "used_appointment_id": null,
          "created_at": "2025-10-25T12:00:00Z"
        }
      ]
    }
  ]
}
```

---

## 3️⃣ **Appointment Creation**

### **3.1 Create Simple Appointment**

**Endpoint:** `POST /api/appointments/simple`

**Headers:**
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**Request Body:**
```json
{
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
  "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
  "department_id": "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
  "individual_slot_id": "slot-123",
  "appointment_date": "2025-10-25",
  "appointment_time": "10:30:00",
  "consultation_type": "clinic_visit",
  "payment_method": "pay_now",
  "payment_type": "upi"
}
```

**Response:**
```json
{
  "message": "Appointment created successfully",
  "appointment": {
    "id": "a6b77b4c-71f9-4ff2-bd09-ff2cb7e92c99",
    "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
    "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
    "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
    "booking_number": "AP-2025-CL001-0101",
    "token_number": 5,
    "appointment_date": "2025-10-25",
    "appointment_time": "2025-10-25T10:30:00Z",
    "consultation_type": "clinic_visit",
    "status": "confirmed",
    "fee_amount": 250.00,
    "payment_status": "paid",
    "payment_mode": "upi",
    "created_at": "2025-10-25T10:00:00Z"
  },
  "is_regular_appointment": true,
  "followup_granted": true,
  "followup_message": "Free follow-up eligibility granted (valid for 5 days)",
  "followup_valid_until": "2025-10-30"
}
```

### **3.2 Create Follow-Up Appointment**

**Request Body:**
```json
{
  "clinic_id": "f7658c53-72ae-4bd3-9960-741225ebc0a2",
  "clinic_patient_id": "b7e83e77-1272-4c73-9d12-68f6c9f91555",
  "doctor_id": "d932bfa4-82ab-4b93-a08a-142c1e259a44",
  "department_id": "a13b9f33-72c7-46b7-92b2-d6b54ef11c1e",
  "individual_slot_id": "slot-456",
  "appointment_date": "2025-10-27",
  "appointment_time": "14:00:00",
  "consultation_type": "follow-up-via-clinic",
  "payment_method": "free"
}
```

**Response:**
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
  "is_free_followup": true,
  "followup_type": "free",
  "followup_message": "This is a FREE follow-up"
}
```

---

## 4️⃣ **Follow-Up System**

### **4.1 Follow-Up Status Lifecycle**

```
Book Regular Appointment
    ↓
status = "active" (has valid follow-up for 5 days)
    ↓
Book Free Follow-Up
    ↓
status = "used" (follow-up consumed)
    ↓
[5+ days pass with no new appointment]
    ↓
status = "expired" (follow-up expired)
    ↓
Book New Regular Appointment (same doctor+dept)
    ↓
status = "renewed" (new free follow-up created)
    ↓
status = "active" (cycle repeats)
```

### **4.2 Follow-Up Eligibility Rules**

| Condition | Is Free Follow-Up? | Reason |
|-----------|-------------------|---------|
| First appointment with doctor+dept | ✅ Yes | Creates new free follow-up |
| Within 5 days, first follow-up | ✅ Yes | Free follow-up available |
| Within 5 days, already used | ❌ No | Free follow-up already used |
| After 5 days | ❌ No | Follow-up expired |
| Different doctor | ❌ No | Different doctor = new appointment |
| Different department | ❌ No | Different department = new appointment |

### **4.3 Check Follow-Up Eligibility**

**Endpoint:** `GET /api/appointments/check-follow-up-eligibility?clinic_patient_id={id}&clinic_id={id}&doctor_id={id}&department_id={id}`

**Response:**
```json
{
  "is_free": true,
  "is_eligible": true,
  "days_remaining": 3,
  "message": "Free follow-up available (3 days remaining)",
  "valid_until": "2025-10-30"
}
```

---

## 5️⃣ **Frontend UI Integration**

### **5.1 Patient Creation UI Form**

**Fields:**
```javascript
const patientForm = {
  firstName: string (required),
  lastName: string (required),
  phone: string (required, format: +XXXXXXXXXXX),
  email: string (optional, email format),
  dateOfBirth: date (optional, format: YYYY-MM-DD),
  gender: enum ["male", "female", "other"] (optional),
  moId: string (optional, max 50 chars),
  medicalHistory: text (optional),
  allergies: text (optional),
  bloodGroup: string (optional),
  address1: string (optional),
  district: string (optional),
  state: string (optional),
  smokingStatus: string (optional),
  alcoholUse: string (optional),
  heightCm: number (optional),
  weightKg: number (optional)
}
```

**API Call:**
```javascript
async function createPatient(patientData) {
  const response = await fetch('/api/organizations/clinic-specific-patients', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(patientData)
  });
  return response.json();
}
```

### **5.2 Appointment Booking UI**

**Step 1: Select Patient**
```javascript
// Show patient search with follow-up status
const patients = await fetch('/api/organizations/clinic-specific-patients?clinic_id=...');
// Display: name, phone, mo_id, current_followup_status
```

**Step 2: Select Doctor & Department**
```javascript
// Show available doctors for this clinic
const doctors = await fetch('/api/organizations/doctors?clinic_id=...');
```

**Step 3: Select Time Slot**
```javascript
// Show available slots for selected doctor+date
const slots = await fetch('/api/doctors/{doctor_id}/time-slots?date=...');
```

**Step 4: Check Follow-Up Eligibility**
```javascript
// If patient has active follow-up for same doctor+dept
const followUpEligibility = await fetch(
  '/api/appointments/check-follow-up-eligibility?' +
  `clinic_patient_id=${patientId}&doctor_id=${doctorId}&department_id=${deptId}`
);

if (followUpEligibility.is_free && followUpEligibility.is_eligible) {
  // Show: "You have FREE follow-up available!"
  // Show: Days remaining
}
```

**Step 5: Create Appointment**
```javascript
async function createAppointment(appointmentData) {
  const response = await fetch('/api/appointments/simple', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      clinic_id: currentClinicId,
      clinic_patient_id: selectedPatient.id,
      doctor_id: selectedDoctor.id,
      department_id: selectedDepartment.id,
      individual_slot_id: selectedSlot.id,
      appointment_date: selectedDate,
      appointment_time: selectedTime,
      consultation_type: followUpEligibility.is_free ? 'follow-up-via-clinic' : 'clinic_visit',
      payment_method: 'pay_now',
      payment_type: 'upi'
    })
  });
  return response.json();
}
```

### **5.3 Display Follow-Up Information**

```javascript
// Patient List View
function displayPatientWithFollowUp(patient) {
  return {
    id: patient.id,
    name: `${patient.first_name} ${patient.last_name}`,
    phone: patient.phone,
    moId: patient.mo_id,
    followUpStatus: patient.current_followup_status, // 'active', 'used', 'expired', 'renewed'
    hasActiveFollowUp: patient.current_followup_status === 'active',
    daysRemaining: calculateDaysRemaining(patient.follow_ups),
    appointments: patient.appointments,
    followUps: patient.follow_ups
  };
}

// Display Follow-Up Status Badge
function FollowUpStatusBadge({ status }) {
  const statusColors = {
    'none': 'gray',
    'active': 'green',
    'used': 'blue',
    'expired': 'red',
    'renewed': 'purple'
  };
  
  return (
    <Badge color={statusColors[status]}>
      {status.toUpperCase()}
    </Badge>
  );
}
```

---

## 6️⃣ **Data Models**

### **6.1 Clinic Patient Model**

```typescript
interface ClinicPatient {
  id: string;                          // UUID
  clinic_id: string;                   // UUID
  first_name: string;                  // Required
  last_name: string;                   // Required
  phone: string;                       // Required, unique per clinic
  email?: string;
  mo_id?: string;                      // Clinic-specific patient ID
  date_of_birth?: string;              // YYYY-MM-DD
  age?: number;
  gender?: 'male' | 'female' | 'other';
  medical_history?: string;
  allergies?: string;
  blood_group?: string;
  address1?: string;
  address2?: string;
  district?: string;
  state?: string;
  smoking_status?: string;
  alcohol_use?: string;
  height_cm?: number;
  weight_kg?: number;
  is_active: boolean;                  // Default: true
  
  // Status fields
  current_followup_status: 'none' | 'active' | 'used' | 'expired' | 'renewed';
  last_appointment_id?: string;        // UUID
  last_followup_id?: string;          // UUID
  
  global_patient_id?: string;         // Optional link to global patient
  created_at: string;                 // ISO 8601
  updated_at: string;                  // ISO 8601
}
```

### **6.2 Appointment Model**

```typescript
interface Appointment {
  id: string;                          // UUID
  clinic_patient_id?: string;         // UUID
  clinic_id: string;                  // UUID
  doctor_id: string;                  // UUID
  department_id?: string;             // UUID
  booking_number: string;             // Unique
  token_number?: number;
  appointment_date?: string;          // YYYY-MM-DD
  appointment_time: string;           // ISO 8601
  consultation_type: string;         // 'clinic_visit', 'video_consultation', 'follow-up-via-clinic', 'follow-up-via-video'
  status: string;                     // 'booked', 'confirmed', 'completed', 'cancelled'
  fee_amount?: number;
  payment_status: string;             // 'paid', 'waived', 'pending'
  payment_mode?: string;             // 'cash', 'card', 'upi', 'online'
  is_priority: boolean;
  individual_slot_id?: string;
  created_at: string;                 // ISO 8601
}
```

### **6.3 Follow-Up Model**

```typescript
interface FollowUp {
  id: string;                         // UUID
  clinic_patient_id: string;         // UUID
  clinic_id: string;                  // UUID
  doctor_id: string;                  // UUID
  department_id?: string;             // UUID
  source_appointment_id: string;      // UUID
  
  status: 'active' | 'used' | 'expired' | 'renewed';
  is_free: boolean;                   // Default: true
  
  valid_from: string;                 // YYYY-MM-DD
  valid_until: string;               // YYYY-MM-DD
  
  used_appointment_id?: string;       // UUID
  used_at?: string;                   // ISO 8601
  
  renewed_by_appointment_id?: string; // UUID
  renewed_at?: string;               // ISO 8601
  
  created_at: string;                 // ISO 8601
  updated_at: string;                 // ISO 8601
}
```

---

## 7️⃣ **Complete Flow Examples**

### **Flow 1: New Patient with Appointment & Follow-Up**

```json
// Step 1: Create Patient
POST /api/organizations/clinic-specific-patients
{
  "clinic_id": "clinic-1",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+911234567890",
  "mo_id": "MO001"
}
// Response: patient_id = "patient-1"

// Step 2: Create Appointment
POST /api/appointments/simple
{
  "clinic_id": "clinic-1",
  "clinic_patient_id": "patient-1",
  "doctor_id": "doctor-1",
  "department_id": "dept-1",
  "appointment_date": "2025-10-25",
  "consultation_type": "clinic_visit"
}
// Response: appointment_id = "appt-1"
// Follow-up created automatically

// Step 3: Check Status
GET /api/organizations/clinic-specific-patients?clinic_id=clinic-1&search=John
// Response:
{
  "current_followup_status": "active",
  "last_appointment_id": "appt-1",
  "last_followup_id": "fup-1",
  "follow_ups": [
    {
      "status": "active",
      "is_free": true,
      "valid_until": "2025-10-30"
    }
  ]
}
```

---

## 🎉 **Complete System Ready!**

All documentation provided for:
- ✅ Login flow
- ✅ Patient creation
- ✅ Appointment booking
- ✅ Follow-up system
- ✅ UI integration
- ✅ Data models
- ✅ Complete flow examples

System is ready for frontend integration! 🚀

