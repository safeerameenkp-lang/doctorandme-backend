# Clinic Patient Creation - Complete API Guide

## 🎯 Overview

There are **3 ways** to create patients in the system. For clinic-specific patient creation, use the **Clinic Patient API**.

---

## 📋 Three Patient Creation Methods

### Method 1: Clinic Creates Patient (✅ RECOMMENDED FOR YOU)

**Endpoint:**
```
POST /api/organizations/clinic-patients
```

**Purpose:** Clinic staff creates a new patient directly for their clinic

**Authorization:** Clinic Admin, Receptionist

**Request:**
```json
POST /api/organizations/clinic-patients
Content-Type: application/json
Authorization: Bearer {token}

{
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "email": "john.doe@example.com",
  "date_of_birth": "1990-05-15",
  "gender": "male",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "mo_id": "MO123456",
  "medical_history": "Diabetes, Hypertension",
  "allergies": "Penicillin",
  "blood_group": "O+"
}
```

**What Happens:**
1. ✅ Creates user account
2. ✅ Assigns 'patient' role
3. ✅ Creates patient record
4. ✅ **Automatically links patient to the clinic**
5. ✅ Returns complete patient info

**Response (201 Created):**
```json
{
  "user": {
    "id": "user-uuid-abc123",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "email": "john.doe@example.com",
    "date_of_birth": "1990-05-15",
    "gender": "male",
    "is_active": true,
    "created_at": "2024-10-15T10:30:00Z"
  },
  "patient": {
    "id": "patient-uuid-xyz789",
    "user_id": "user-uuid-abc123",
    "mo_id": "MO123456",
    "medical_history": "Diabetes, Hypertension",
    "allergies": "Penicillin",
    "blood_group": "O+",
    "is_active": true,
    "created_at": "2024-10-15T10:30:00Z",
    "updated_at": "2024-10-15T10:30:00Z"
  },
  "clinic_assignment": {
    "id": "assignment-uuid",
    "patient_id": "patient-uuid-xyz789",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "is_primary": true,
    "created_at": "2024-10-15T10:30:00Z"
  }
}
```

---

### Method 2: Create Patient + Appointment (One Step)

**Endpoint:**
```
POST /api/appointments/patient-appointment
```

**Purpose:** Create patient and book appointment in one call

**Request:**
```json
POST /api/appointments/patient-appointment
{
  "first_name": "Jane",
  "last_name": "Smith",
  "phone": "+9876543210",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "new"
}
```

**What Happens:**
1. ✅ Creates user
2. ✅ Creates patient
3. ✅ Links to clinic
4. ✅ **Creates appointment immediately**

---

### Method 3: Super Admin Creates Patient (Global)

**Endpoint:**
```
POST /api/organizations/patients
```

**Purpose:** Super admin creates patient globally (not clinic-specific)

**Authorization:** Super Admin only

---

## ✅ Recommended Flow for Clinics

### Step 1: Create Patient for Your Clinic

**Request:**
```json
POST /api/organizations/clinic-patients
{
  "first_name": "Ahmed",
  "last_name": "Khan",
  "phone": "+971501234567",
  "email": "ahmed.khan@email.com",
  "date_of_birth": "1985-03-20",
  "gender": "male",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "mo_id": "MO789012",
  "medical_history": "None",
  "blood_group": "B+"
}
```

**Response:**
```json
{
  "user": {
    "id": "user-uuid-new"
  },
  "patient": {
    "id": "patient-uuid-new",
    "mo_id": "MO789012"
  },
  "clinic_assignment": {
    "patient_id": "patient-uuid-new",
    "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
    "is_primary": true
  }
}
```

---

### Step 2: Book Appointment with Created Patient

**Request:**
```json
POST /api/appointments
{
  "patient_id": "patient-uuid-new",
  "doctor_id": "85394ce8-94f7-4dca-a536-34305c46a98e",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-09-30-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline"
}
```

**Validation Passes:**
- ✅ Patient exists
- ✅ Patient is registered at this clinic (from Step 1)
- ✅ Appointment created successfully

---

## 📊 List All Patients in a Clinic

**Endpoint:**
```
GET /api/organizations/clinic-patients?clinic_id={clinic-uuid}
```

**Query Parameters:**
- `clinic_id` (optional) - Filter by specific clinic
- `only_active` (optional) - Default: "true"
- `search` (optional) - Search by name, phone, or mo_id

**Example:**
```
GET /api/organizations/clinic-patients?clinic_id=7a6c1211-c029-4923-a1a6-fe3dfe48bdf2&only_active=true
```

**Response:**
```json
[
  {
    "id": "patient-uuid-1",
    "user_id": "user-uuid-1",
    "first_name": "Ahmed",
    "last_name": "Khan",
    "phone": "+971501234567",
    "email": "ahmed.khan@email.com",
    "date_of_birth": "1985-03-20",
    "gender": "male",
    "mo_id": "MO789012",
    "medical_history": "None",
    "allergies": null,
    "blood_group": "B+",
    "is_active": true,
    "created_at": "2024-10-15T10:30:00Z",
    "clinic_name": "Main Hospital"
  },
  {
    "id": "patient-uuid-2",
    "user_id": "user-uuid-2",
    "first_name": "Sara",
    "last_name": "Ali",
    "phone": "+971507654321",
    "email": "sara.ali@email.com",
    "mo_id": "MO123456",
    "is_active": true,
    "created_at": "2024-10-14T14:20:00Z",
    "clinic_name": "Main Hospital"
  }
  // ... more patients
]
```

---

## 🔍 Search Patients

**Search by Name:**
```
GET /clinic-patients?clinic_id=xxx&search=Ahmed
```

**Search by Phone:**
```
GET /clinic-patients?clinic_id=xxx&search=+971501234567
```

**Search by Mo ID:**
```
GET /clinic-patients?clinic_id=xxx&search=MO123456
```

---

## 📝 Complete Workflow Example

### Scenario: Clinic Receptionist Registers New Patient

**Step 1: Create Patient**
```json
POST /clinic-patients
{
  "first_name": "Mohammed",
  "last_name": "Hassan",
  "phone": "+971509876543",
  "email": "mohammed.hassan@email.com",
  "date_of_birth": "1992-08-10",
  "gender": "male",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "mo_id": "MO345678",
  "blood_group": "A+"
}
```

**Response:**
```json
{
  "patient": {
    "id": "patient-new-uuid"
  }
}
```

---

**Step 2: Book Appointment**
```json
POST /appointments
{
  "patient_id": "patient-new-uuid",
  "doctor_id": "doctor-uuid",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "individual_slot_id": "slot-uuid",
  "appointment_date": "2025-10-18",
  "appointment_time": "2025-10-18 09:30:00",
  "consultation_type": "offline"
}
```

**Response:**
```json
{
  "appointment": {
    "id": "appointment-uuid",
    "patient_id": "patient-new-uuid",
    "booking_number": "BN202510180001",
    "status": "confirmed"
  }
}
```

---

## 🎯 API Comparison

| API | Purpose | Requires patient_id? | Creates Appointment? |
|-----|---------|---------------------|---------------------|
| **POST /clinic-patients** | ✅ Create patient for clinic | ❌ No | ❌ No |
| **POST /appointments** | Book appointment | ✅ Yes | ✅ Yes |
| **POST /appointments/patient-appointment** | Create patient + book | ❌ No | ✅ Yes |

---

## 📊 Required Fields for Clinic Patient Creation

### Minimum Required:
```json
{
  "first_name": "string (required)",
  "last_name": "string (required)",
  "phone": "string (required, unique)",
  "clinic_id": "UUID (required)"
}
```

### Optional Fields:
```json
{
  "email": "string (optional)",
  "date_of_birth": "YYYY-MM-DD (optional)",
  "gender": "string (optional)",
  "mo_id": "string (optional, unique)",
  "medical_history": "string (optional)",
  "allergies": "string (optional)",
  "blood_group": "string (optional)"
}
```

---

## ✅ Your Use Case Solution

**For clinic to create patients directly:**

```
POST /api/organizations/clinic-patients
```

**Full Example:**
```json
{
  "first_name": "Ali",
  "last_name": "Mohammed",
  "phone": "+971501112233",
  "email": "ali.m@email.com",
  "date_of_birth": "1988-11-25",
  "gender": "male",
  "clinic_id": "7a6c1211-c029-4923-a1a6-fe3dfe48bdf2",
  "mo_id": "MO999888",
  "medical_history": "Asthma",
  "allergies": "Pollen",
  "blood_group": "AB+"
}
```

**Result:**
- ✅ Creates patient
- ✅ Links to YOUR clinic
- ✅ Ready to book appointments
- ✅ No patient_id needed in request!

---

## 📖 Summary

**To list clinic patients:**
```
GET /clinic-patients?clinic_id=your-clinic-uuid
```

**To create clinic patient:**
```
POST /clinic-patients
{
  "first_name": "...",
  "last_name": "...",
  "phone": "...",
  "clinic_id": "your-clinic-uuid"
}
```

**To book appointment:**
```
POST /appointments
{
  "patient_id": "patient-uuid-from-step-above",
  "clinic_id": "your-clinic-uuid",
  ...
}
```

---

**Status:** ✅ API Already Exists  
**Endpoint:** `POST /api/organizations/clinic-patients`

